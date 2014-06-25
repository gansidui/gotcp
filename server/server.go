package server

import (
	"github.com/gansidui/gotcp/packet"
	"github.com/gansidui/gotcp/utils"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	exitCh        chan bool       // 结束信号
	waitGroup     *sync.WaitGroup // 等待goroutine
	connMap       *utils.SafeMap  // addr到conn的映射(string --> *net.TCPConn)
	funcMap       *utils.FuncMap  // 映射消息处理函数(uint16 --> func)
	acceptTimeout time.Duration   // 连接超时时间
	readTimeout   time.Duration   // 读超时时间,其实也就是心跳维持时间
}

func NewServer() *Server {
	return &Server{
		exitCh:        make(chan bool),
		waitGroup:     &sync.WaitGroup{},
		connMap:       utils.NewSafeMap(),
		funcMap:       utils.NewFuncMap(),
		acceptTimeout: 10,
		readTimeout:   60,
	}
}

func (this *Server) SetAcceptTimeout(acceptTimeout time.Duration) {
	this.acceptTimeout = acceptTimeout
}

func (this *Server) SetReadTimeout(readTimeout time.Duration) {
	this.readTimeout = readTimeout
}

func (this *Server) Start(listener *net.TCPListener) {
	log.Printf("Start listen on %v", listener.Addr())
	this.waitGroup.Add(1)
	defer func() {
		listener.Close()
		this.waitGroup.Done()
	}()

	for {
		select {
		case <-this.exitCh:
			log.Printf("Stop listen on %v", listener.Addr())
			return
		default:
		}

		listener.SetDeadline(time.Now().Add(this.acceptTimeout))
		conn, err := listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				// log.Printf("Accept timeout: %v", opErr)
				continue
			}
			log.Printf("Accept error: %v", err)
			continue
		}

		log.Printf("Accept: %v", conn.RemoteAddr())
		go this.handleClientConn(conn)
	}
}

func (this *Server) Stop() {
	// 有多个goroutine使用exitCh，必须选择close, close后goroutine那边再读的话会一直返回false
	close(this.exitCh)
	this.waitGroup.Wait()
}

func (this *Server) BindMsgHandler(pacType uint16, fn interface{}) error {
	return this.funcMap.Bind(pacType, fn)
}

func (this *Server) handleClientConn(conn *net.TCPConn) {
	this.waitGroup.Add(1)
	defer this.waitGroup.Done()

	// 建立addr-->conn映射
	addr := conn.RemoteAddr().String()
	this.connMap.Set(addr, conn)

	receivePackets := make(chan *packet.Packet, 20) // 接收到的包
	chStop := make(chan bool)                       // 通知停止消息处理

	defer func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("Panic: %v", e)
			}
		}()
		conn.Close()
		this.connMap.Delete(addr)
		log.Printf("Disconnect: %v", addr)
		chStop <- true
	}()

	// 处理接收到的包
	go this.handlePackets(conn, receivePackets, chStop)

	// 接收数据
	log.Printf("HandleClient: %v", addr)
	request := make([]byte, 1024)
	buf := make([]byte, 0)
	var bufLen uint32 = 0

	for {
		select {
		case <-this.exitCh:
			log.Printf("Stop handleClientConn")
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(this.readTimeout))
		readSize, err := conn.Read(request)
		if err != nil {
			log.Printf("Read failed: %v", err)
			return
		}

		if readSize > 0 {
			buf = append(buf, request[:readSize]...)
			bufLen += uint32(readSize)

			for {
				if bufLen >= 6 {
					pacLen := utils.BytesToUint32(buf[0:4])
					if bufLen >= pacLen {
						receivePackets <- &packet.Packet{
							Len:  pacLen,
							Type: utils.BytesToUint16(buf[4:6]),
							Data: buf[6:pacLen],
						}
						buf = buf[pacLen:]
						bufLen -= pacLen
					} else {
						break
					}
				} else {
					break
				}
			}

		}

	}
}

func (this *Server) handlePackets(conn *net.TCPConn, receivePackets <-chan *packet.Packet, chStop <-chan bool) {
	for {
		select {
		case <-chStop:
			log.Printf("Stop handle receivePackets.")
			return

		// 消息包处理
		case p := <-receivePackets:
			if this.funcMap.Exist(p.Type) {
				this.funcMap.Call(p.Type, conn, p)
			} else {
				log.Printf("Unknown packet type")
			}
		}
	}
}
