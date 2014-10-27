package gotcp

import (
	"log"
	"net"
	"sync"
	"time"
)

// Server struct
type Server struct {
	config      *Config          // configure infomation
	delegate    ConnDelegate     // conn delegate(message callbacks)
	deliverData *deliverConnData // deliver to conn
}

// Server delivery deliverConnData to the connection to control
type deliverConnData struct {
	exitChan  chan struct{}   // server notify all goroutines to shutdown
	waitGroup *sync.WaitGroup // wait for all goroutines
}

func NewServer(config *Config, delegate ConnDelegate) *Server {
	return &Server{
		config:   config,
		delegate: delegate,
		deliverData: &deliverConnData{
			exitChan:  make(chan struct{}),
			waitGroup: &sync.WaitGroup{},
		},
	}
}

// Start server
func (s *Server) Start(listener *net.TCPListener) {
	log.Printf("Start listen on %v\r\n", listener.Addr())
	s.deliverData.waitGroup.Add(1)
	defer func() {
		log.Printf("Stop listen on %v\r\n", listener.Addr())
		listener.Close()
		s.deliverData.waitGroup.Done()
	}()

	for {
		select {
		case <-s.deliverData.exitChan:
			return

		default:
		}

		listener.SetDeadline(time.Now().Add(s.config.AcceptTimeout))

		conn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}

		go NewConn(conn, s.config, s.delegate, s.deliverData).do()
	}
}

// Stop server
func (s *Server) Stop() {
	close(s.deliverData.exitChan)
	s.deliverData.waitGroup.Wait()
}
