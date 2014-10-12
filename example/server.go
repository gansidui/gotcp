package main

import (
	"fmt"
	"github.com/gansidui/gotcp"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	config := &gotcp.Config{
		AcceptTimeout:  30 * time.Second,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxPacLen:      uint32(1024),
		RecPacBufLimit: uint32(20),
	}

	callbacks := &gotcp.Callbacks{
		OnConnect:       onConnect,
		OnDisconnect:    onDisconnect,
		OnReceivePacket: onReceivePacket,
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	svr := gotcp.NewServer(config, callbacks)
	go svr.Start(listener)

	go func() {
		for {
			fmt.Println("=======num goroutine === ", runtime.NumGoroutine())
			fmt.Println(onConnectNum, onDisconnectNum, onReceivePacketNum)
			time.Sleep(1 * time.Second)
		}
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal: %v\r\n", <-ch)

	svr.Stop()
}

var onConnectNum, onDisconnectNum, onReceivePacketNum int

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func onConnect(conn *net.TCPConn) error {
	onConnectNum++
	fmt.Println("onConnect", conn.RemoteAddr())
	return nil
}

func onDisconnect(conn *net.TCPConn) {
	onDisconnectNum++
	fmt.Println("onDisconnect", conn.RemoteAddr())
}

func onReceivePacket(conn *net.TCPConn, pac *gotcp.Packet) error {
	onReceivePacketNum++
	fmt.Println("onReceivePacket", conn.RemoteAddr())
	fmt.Println(pac.GetLen(), pac.GetType(), string(pac.GetData()))
	return nil
}
