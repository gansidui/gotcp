package gotcp

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Conn exposes a set of callbacks for the
// various events that occur on a connection
type Conn struct {
	conn        *net.TCPConn     // the raw TCPConn
	config      *Config          // configure information
	delegate    ConnDelegate     // callbacks in Conn
	deliverData *deliverConnData // server delivery deliverConnData to the connection to control

	extraData interface{} // save the extra data with conn

	closeOnce sync.Once // close the conn, once, per instance.
	closeFlag int32
	closeChan chan struct{}

	sendPacketChan    chan *Packet // send packet queue
	receivePacketChan chan *Packet // receive packet queue
}

// ConnDelegate is an interface of methods
// that are used as callbacks in Conn
type ConnDelegate interface {
	// OnConnect is called when the connection was accepted,
	// If the return value of false is closed
	OnConnect(*Conn) bool

	// OnMessage is called when the connection receives a packet,
	// If the return value of false is closed
	OnMessage(*Conn, *Packet) bool

	// OnClose is called when the connection closed
	OnClose(*Conn)

	// OnIOError is called when the connection experiences
	// a low-level TCP transport error
	OnIOError(*Conn, error)
}

// The configure of connection
type Config struct {
	AcceptTimeout          time.Duration // connection accepted timeout
	ReadTimeout            time.Duration // connection read timeout
	WriteTimeout           time.Duration // connection write timeout
	MaxPacketLength        uint32        // the maximum length of packet
	SendPacketChanLimit    uint32        // the limit of packet send channel
	ReceivePacketChanLimit uint32        // the limit of packet receive channel
}

func newConn(conn *net.TCPConn, config *Config, delegate ConnDelegate, deliverData *deliverConnData) *Conn {
	return &Conn{
		conn:              conn,
		config:            config,
		delegate:          delegate,
		deliverData:       deliverData,
		closeChan:         make(chan struct{}),
		sendPacketChan:    make(chan *Packet, config.SendPacketChanLimit),
		receivePacketChan: make(chan *Packet, config.ReceivePacketChanLimit),
	}
}

// Get extra data
func (c *Conn) GetExtraData() interface{} {
	return c.extraData
}

// Put extra data
func (c *Conn) PutExtraData(data interface{}) {
	c.extraData = data
}

// Get the raw connection to use more features
func (c *Conn) GetRawConn() *net.TCPConn {
	return c.conn
}

// Close the Conn
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		close(c.closeChan)
		c.conn.Close()
		c.delegate.OnClose(c)
	})
}

func (c *Conn) SetReadDeadline(timeout time.Duration) {
	c.conn.SetReadDeadline(time.Now().Add(timeout))
}

func (c *Conn) SetWriteDeadline(timeout time.Duration) {
	c.conn.SetWriteDeadline(time.Now().Add(timeout))
}

// Indicates whether or not the connection is closed
func (c *Conn) IsClosed() bool {
	return atomic.LoadInt32(&c.closeFlag) == 1
}

// Sync read a packet, this method will block on IO
func (c *Conn) ReadPacket() (*Packet, error) {
	return ReadPacket(c.conn, c.config.MaxPacketLength)
}

// Async read a packet, this method will never block
func (c *Conn) AsyncReadPacket(timeout time.Duration) (*Packet, error) {
	if c.IsClosed() {
		return nil, ConnClosedError
	}

	if timeout == 0 {
		select {
		case p := <-c.receivePacketChan:
			return p, nil

		case <-c.closeChan:
			return nil, ConnClosedError

		default:
			return nil, ReadBlockedError
		}

	} else {
		select {
		case p := <-c.receivePacketChan:
			return p, nil

		case <-c.closeChan:
			return nil, ConnClosedError

		case <-time.After(timeout):
			return nil, ReadBlockedError
		}
	}
}

// Sync write a packet, this method will block on IO
func (c *Conn) WritePacket(p *Packet) error {
	if c.IsClosed() {
		return ConnClosedError
	}

	if n, err := c.conn.Write(p.Serialize()); err != nil || n != int(p.GetLen()) {
		return errors.New(fmt.Sprintf("Write error: [%v], n[%v], p.pacLen[%v]", err, n, p.pacLen))
	}

	return nil
}

// Async write a packet, this method will never block
func (c *Conn) AsyncWritePacket(p *Packet, timeout time.Duration) error {
	if c.IsClosed() {
		return ConnClosedError
	}

	if timeout == 0 {
		select {
		case c.sendPacketChan <- p:
			return nil

		case <-c.closeChan:
			return ConnClosedError

		default:
			return WriteBlockedError
		}

	} else {
		select {
		case c.sendPacketChan <- p:
			return nil

		case <-c.closeChan:
			return ConnClosedError

		case <-time.After(timeout):
			return WriteBlockedError
		}
	}
}

func (c *Conn) Do() {
	c.deliverData.waitGroup.Add(3)
	go c.handleLoop()
	go c.readLoop()
	go c.writeLoop()
}

func (c *Conn) readLoop() {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("readLoop panic: %v\r\n", e)
		}
		c.Close()
		c.deliverData.waitGroup.Done()
	}()

	for {
		select {
		case <-c.deliverData.exitChan:
			return

		case <-c.closeChan:
			return

		default:
		}

		c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))

		recPacket, err := ReadPacket(c.conn, c.config.MaxPacketLength)
		if err != nil {
			c.delegate.OnIOError(c, err)
			return
		}

		c.receivePacketChan <- recPacket
	}
}

func (c *Conn) writeLoop() {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("writeLoop panic: %v\r\n", e)
		}
		c.Close()
		c.deliverData.waitGroup.Done()
	}()

	for {
		select {
		case <-c.deliverData.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.sendPacketChan:
			err := c.WritePacket(p)
			if err != nil {
				c.delegate.OnIOError(c, err)
				return
			}
		}
	}
}

func (c *Conn) handleLoop() {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("handleLoop panic: %v\r\n", e)
		}
		c.Close()
		c.deliverData.waitGroup.Done()
	}()

	if !c.delegate.OnConnect(c) {
		return
	}

	for {
		select {
		case <-c.deliverData.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.receivePacketChan:
			if !c.delegate.OnMessage(c, p) {
				return
			}
		}
	}
}
