package server

import (
	"github.com/things-go/go-socks5"
	"log"
	"net"
	"os"
)

type Server struct {
	listener   net.Listener
	socks5port string
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Run() {
	var err error
	s.listener, err = net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("Failed to start client listener: %v", err)
	}
	log.Println("Server is listening for client connections on :8000")
}

func (s *Server) Process() {
	clientConn, err := s.listener.Accept()
	if err != nil {
		log.Fatalf("Failed to accept client connection: %v", err)
	}
	log.Println("Client connected:", clientConn.RemoteAddr())
	buffer := make([]byte, 1024)
	n, err := clientConn.Read(buffer)
	if err != nil {
		log.Fatalf("Failed to read from client: %v", err)
	}
	log.Printf("Received from client: %s", buffer[:n])
	s.socks5port = string(buffer[:n])
}

func (s *Server) RunSocks5() {
	server := socks5.NewServer(
		socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
	)

	if err := server.ListenAndServe("tcp", s.socks5port); err != nil {
		panic(err)
	}
}
