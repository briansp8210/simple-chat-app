package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/briansp8210/simple-chat-app/internal/server"
)

var (
	host = flag.String("host", "0.0.0.0", "Binding IP address")
	port = flag.Int("port", 61234, "Binding port number")
)

func main() {
	flag.Parse()

	chatServer := server.NewChatServer()
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatal(err)
	}
	chatServer.Serve(&lis)
}
