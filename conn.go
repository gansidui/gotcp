package gotcp

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Error type
var (
	ErrConnClosing   = errors.New("use of closed network connection")
	ErrWriteBlocking = errors.New("write packet was blocking")
	ErrReadBlocking  = errors.New("read packet was blocking")
)

// Conn exposes a set of callbacks for the various events that occur on a connection
type Conn struct {
	basic             *basicSrv
	conn              *net.TCPConn // the raw connection
	extraData         interface{}  // save the extra data
	closeOnce         sync.Once    // close the conn, once, per instance
	closeFlag         int32
	closeChan         chan struct{}
	packetSendChan    chan Packet // packet send queue
	packetReceiveChan chan Packet // packeet receive queue
}

// ConnCallback is an interface of methods that are used as callbacks on a connection
type ConnCallback interface {
	// OnConnect is called when the connection was accepted,
	// If the return value of false is closed
	OnConnect(*Conn) bool

	// OnMessage is called when the connection receives a packet,
	// If the return value of false is closed
	OnMessage(*Conn, Packet) bool

	// OnClose is called when the connection closed
	OnClose(*Conn)
}

func newConn(conn *net.TCPConn, basic *basicSrv) *Conn {
	return &Conn{
		basic:             basic,
		conn:              conn,
		closeChan:         make(chan struct{}),
		packetSendChan:    make(chan Packet, basic.config.PacketSendChanLimit),
		packetReceiveChan: make(chan Packet, basic.config.PacketReceiveChanLimit),
	}
}

// GetExtraData gets the extra data from the Conn
func (c *Conn) GetExtraData() interface{} {
	return c.extraData
}

// PutExtraData puts the extra data with the Conn
func (c *Conn) PutExtraData(data interface{}) {
	c.extraData = data
}

// GetRawConn returns the raw net.TCPConn from the Conn
func (c *Conn) GetRawConn() *net.TCPConn {
	return c.conn
}

// Close closes the connection
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		close(c.closeChan)
		c.conn.Close()
		c.basic.callback.OnClose(c)
	})
}

// IsClosed indicates whether or not the connection is closed
func (c *Conn) IsClosed() bool {
	return atomic.LoadInt32(&c.closeFlag) == 1
}

// AsyncReadPacket async reads a packet, this method will never block
func (c *Conn) AsyncReadPacket(timeout time.Duration) (Packet, error) {
	if c.IsClosed() {
		return nil, ErrConnClosing
	}

	if timeout == 0 {
		select {
		case p := <-c.packetReceiveChan:
			return p, nil

		default:
			return nil, ErrReadBlocking
		}

	} else {
		select {
		case p := <-c.packetReceiveChan:
			return p, nil

		case <-c.closeChan:
			return nil, ErrConnClosing

		case <-time.After(timeout):
			return nil, ErrReadBlocking
		}
	}
}

// AsyncWritePacket async writes a packet, this method will never block
func (c *Conn) AsyncWritePacket(p Packet, timeout time.Duration) error {
	if c.IsClosed() {
		return ErrConnClosing
	}

	if timeout == 0 {
		select {
		case c.packetSendChan <- p:
			return nil

		default:
			return ErrWriteBlocking
		}

	} else {
		select {
		case c.packetSendChan <- p:
			return nil

		case <-c.closeChan:
			return ErrConnClosing

		case <-time.After(timeout):
			return ErrWriteBlocking
		}
	}
}

// Do it
func (c *Conn) Do() {
	if !c.basic.callback.OnConnect(c) {
		return
	}

	c.basic.waitGroup.Add(3)
	go c.handleLoop()
	go c.readLoop()
	go c.writeLoop()
}

func (c *Conn) readLoop() {
	defer func() {
		recover()
		c.Close()
		c.basic.waitGroup.Done()
	}()

	for {
		select {
		case <-c.basic.exitChan:
			return

		case <-c.closeChan:
			return

		default:
		}

		c.conn.SetReadDeadline(time.Now().Add(c.basic.config.ReadTimeout))
		recPacket, err := c.basic.protocol.ReadPacket(c.conn, c.basic.config.PacketSizeLimit)
		if err != nil {
			return
		}

		c.packetReceiveChan <- recPacket
	}
}

func (c *Conn) writeLoop() {
	defer func() {
		recover()
		c.Close()
		c.basic.waitGroup.Done()
	}()

	for {
		select {
		case <-c.basic.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.packetSendChan:
			c.conn.SetWriteDeadline(time.Now().Add(c.basic.config.WriteTimeout))
			if _, err := c.conn.Write(p.Serialize()); err != nil {
				return
			}
		}
	}
}

func (c *Conn) handleLoop() {
	defer func() {
		recover()
		c.Close()
		c.basic.waitGroup.Done()
	}()

	for {
		select {
		case <-c.basic.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.packetReceiveChan:
			if !c.basic.callback.OnMessage(c, p) {
				return
			}
		}
	}
}
