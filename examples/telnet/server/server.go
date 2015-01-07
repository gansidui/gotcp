package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gansidui/gotcp"
	"github.com/gansidui/gotcp/examples/telnet"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// create a listener
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:23")
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	// initialize server params
	config := &gotcp.Config{
		AcceptTimeout:          10 * time.Second,
		ReadTimeout:            60 * 10 * time.Second,
		WriteTimeout:           60 * 10 * time.Second,
		PacketSizeLimit:        2048,
		PacketSendChanLimit:    10,
		PacketReceiveChanLimit: 10,
	}
	srv := gotcp.NewServer(config, &telnet.TelnetCallback{}, &telnet.TelnetProtocol{})

	// start server
	go srv.Start(listener)
	fmt.Println("listening:", listener.Addr())

	// catch system signal
	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Signal: ", <-chSig)

	// stop server
	srv.Stop()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
