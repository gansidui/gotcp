package main

import (
	"github.com/gansidui/gotcp/handlers"
	"github.com/gansidui/gotcp/packet"
	"github.com/gansidui/gotcp/server"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var sv *server.Server

func init() {
	sv = server.NewServer()
	sv.SetAcceptTimeout(10 * time.Second)
	sv.SetReadTimeout(60 * time.Second)
	sv.BindMsgHandler(packet.TYPE_MESSAGE, handlers.HandleChatMsg)
}

func main() {
	service := "127.0.0.1:8989"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	// Start server
	go sv.Start(listener)

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal: %v", <-ch)

	// Stop the server gracefully.
	sv.Stop()
}

func checkError(err error) {
	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
