package gotcp

import (
	"net"
	"sync"
	"time"
)

type Server struct {
	basic *basicSrv
}

type Config struct {
	AcceptTimeout          time.Duration // connection accepted timeout
	ReadTimeout            time.Duration // connection read timeout
	WriteTimeout           time.Duration // connection write timeout
	PacketSizeLimit        uint32        // the limit of packet size
	PacketSendChanLimit    uint32        // the limit of packet send channel
	PacketReceiveChanLimit uint32        // the limit of packet receive channel
}

type basicSrv struct {
	config    *Config         // server configuration
	callback  ConnCallback    // message callbacks in connection
	protocol  Protocol        // customize packet protocol
	exitChan  chan struct{}   // notify all goroutines to shutdown
	waitGroup *sync.WaitGroup // wait for all goroutines
}

func newBasicSrv(config *Config, callback ConnCallback, protocol Protocol) *basicSrv {
	return &basicSrv{
		config:    config,
		callback:  callback,
		protocol:  protocol,
		exitChan:  make(chan struct{}),
		waitGroup: &sync.WaitGroup{},
	}
}

// NewServer creates a new Server
func NewServer(config *Config, callback ConnCallback, protocol Protocol) *Server {
	basic := newBasicSrv(config, callback, protocol)
	return &Server{basic}
}

// Start starts server
func (s *Server) Start(listener *net.TCPListener) {
	s.basic.waitGroup.Add(1)
	defer func() {
		listener.Close()
		s.basic.waitGroup.Done()
	}()

	for {
		select {
		case <-s.basic.exitChan:
			return

		default:
		}

		listener.SetDeadline(time.Now().Add(s.basic.config.AcceptTimeout))

		conn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}

		go newConn(conn, s.basic).Do()
	}
}

// Stop stops server
func (s *Server) Stop() {
	close(s.basic.exitChan)
	s.basic.waitGroup.Wait()
}

// Dial dials to the other server
func (s *Server) Dial(network, address string, config *Config, callback ConnCallback, protocol Protocol) (*Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	basic := newBasicSrv(config, callback, protocol)

	return newConn(conn, basic), nil
}
