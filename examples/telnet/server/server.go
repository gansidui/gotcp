package main

import (
	"github.com/gansidui/gotcp"
	"github.com/gansidui/gotcp/examples/telnet"
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

	// listen
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:23")
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	// set config and delegate
	config := &gotcp.Config{
		AcceptTimeout:          10 * time.Second,
		ReadTimeout:            60 * 10 * time.Second,
		WriteTimeout:           60 * 10 * time.Second,
		MaxPacketLength:        2048,
		SendPacketChanLimit:    10,
		ReceivePacketChanLimit: 10,
	}
	delegate := &telnet.TelnetConnDelegate{}
	protocol := &telnet.TelnetProtocol{}

	// start server
	svr := gotcp.NewServer(config, delegate, protocol)
	go svr.Start(listener)

	// catch signal
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal: %v\r\n", <-ch)

	// stop server
	svr.Stop()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
