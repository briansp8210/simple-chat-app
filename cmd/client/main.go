package main

import (
	"flag"

	"github.com/briansp8210/simple-chat-app/internal/client"
)

func main() {
	flag.Parse()
	host := flag.String("host", "127.0.0.1", "Binding IP address")
	port := flag.Int("port", 61234, "Binding port number")

	client := client.NewChatClient(*host, *port)
	client.Run()
}
