package client

import (
	"log"
	"net"
)

type Client struct {
}

func New() *Client {
	return &Client{}
}

func (c *Client) Connect() {
	serverConn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer serverConn.Close()
	log.Println("Connected to server")
	serverConn.Write([]byte(":10800"))
}
