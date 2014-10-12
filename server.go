package gotcp

import (
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	AcceptTimeout  time.Duration // connection acception timeout
	ReadTimeout    time.Duration // connection read timeout
	WriteTimeout   time.Duration // connection write timeout
	MaxPacLen      uint32        // the maximum length of packet
	RecPacBufLimit uint32        // the limit of receive packet buffer, block but not discarded
}

type Callbacks struct {
	OnConnect       func(*net.TCPConn) error
	OnDisconnect    func(*net.TCPConn)
	OnReceivePacket func(*net.TCPConn, *Packet) error
}

type Server struct {
	exitCh    chan struct{}   // notify all goroutine terminate
	waitGroup *sync.WaitGroup // wait all goroutine
	config    *Config         // configure info
	callbacks *Callbacks      // message callbacks
}

func NewServer(config *Config, callbacks *Callbacks) *Server {
	return &Server{
		exitCh:    make(chan struct{}),
		waitGroup: &sync.WaitGroup{},
		config:    config,
		callbacks: callbacks,
	}
}

func (s *Server) Start(listener *net.TCPListener) {
	log.Printf("Start listen on %v\r\n", listener.Addr())
	s.waitGroup.Add(1)
	defer func() {
		listener.Close()
		s.waitGroup.Done()
	}()

	for {
		select {
		case <-s.exitCh:
			log.Printf("Stop listen on %v\r\n", listener.Addr())
			return

		default:
		}

		listener.SetDeadline(time.Now().Add(s.config.AcceptTimeout))

		conn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}

		if s.callbacks.OnConnect(conn) != nil {
			if conn != nil {
				conn.Close()
			}
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Server) Stop() {
	close(s.exitCh)
	s.waitGroup.Wait()
}

func (s *Server) handleConn(conn *net.TCPConn) {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()

	recPackets := make(chan *Packet, s.config.RecPacBufLimit)
	var bStop int32 = 0 // notify the handleConn() and handlePacket() terminate each other

	defer func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("Panic: %v\r\n", e)
			}
		}()

		conn.Close()
		atomic.StoreInt32(&bStop, 1)
		s.callbacks.OnDisconnect(conn)
	}()

	// handle the received packets
	go s.handlePacket(conn, recPackets, &bStop)

	var (
		pacBLen  []byte = make([]byte, 4)
		pacBType []byte = make([]byte, 4)
		pacLen   uint32
	)

	for {
		select {
		case <-s.exitCh:
			return

		default:
			if atomic.LoadInt32(&bStop) == 1 {
				return
			}
		}

		conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))

		// read pacLen
		if n, err := io.ReadFull(conn, pacBLen); err != nil && n != 4 {
			return
		}
		if pacLen = BytesToUint32(pacBLen); pacLen > s.config.MaxPacLen {
			return
		}

		// read pacType
		if n, err := io.ReadFull(conn, pacBType); err != nil && n != 4 {
			return
		}

		// read pacData
		pacData := make([]byte, pacLen-8)
		if n, err := io.ReadFull(conn, pacData); err != nil && n != int(pacLen) {
			return
		}

		recPackets <- NewPacket(BytesToUint32(pacBType), pacData)
	}
}

func (s *Server) handlePacket(conn *net.TCPConn, recPackets <-chan *Packet, bStop *int32) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("Panic: %v\r\n", e)
		}
		atomic.StoreInt32(bStop, 1)
	}()

	for {
		select {
		case p := <-recPackets:
			if s.callbacks.OnReceivePacket(conn, p) != nil {
				if conn != nil {
					conn.Close()
				}
				return
			}

		default:
			if atomic.LoadInt32(bStop) == 1 {
				return
			}
		}
	}
}
