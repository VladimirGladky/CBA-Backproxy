package server

import (
	"CBA-Backproxy/pkg/logger"
	"context"
	"github.com/things-go/go-socks5"
	"go.uber.org/zap"
	"log"
	"net"
	"os"
)

type Server struct {
	listener   net.Listener
	socks5port string
	ctx        context.Context
	clientConn net.Conn
}

func NewServer(ctx context.Context) *Server {
	return &Server{
		ctx: ctx,
	}
}

func (s *Server) Run() {
	var err error
	s.listener, err = net.Listen("tcp", ":18000")
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Failed to listen on port 8000", zap.Error(err))
	}
	logger.GetLoggerFromCtx(s.ctx).Info(s.ctx, "Server started on port 8000")
}

func (s *Server) Process() bool {
	var err error
	s.clientConn, err = s.listener.Accept()
	if err != nil {
		log.Fatalf("Failed to accept connection: %v", err)
		return false
	}
	logger.GetLoggerFromCtx(s.ctx).Info(s.ctx, "Accepted connection from client")
	buffer := make([]byte, 1024)
	n, err := s.clientConn.Read(buffer)
	if err != nil {
		log.Fatalf("Failed to read from client: %v", err)
		return false
	}
	logger.GetLoggerFromCtx(s.ctx).Info(s.ctx, "Received data from client", zap.String("data", string(buffer[:n])))
	s.socks5port = string(buffer[:n])
	return true
}

func (s *Server) SendReqToClient(addr string) {
	_, err := s.clientConn.Write([]byte(addr))
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Failed to send request to client", zap.Error(err))
	}
}

func (s *Server) RunSocks5() {
	customDialer := &CustomDialer{
		s: s,
	}

	server := socks5.NewServer(
		socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
		socks5.WithDial(customDialer.Dial),
	)

	logger.GetLoggerFromCtx(s.ctx).Info(s.ctx, "Socks5 Server started on port"+s.socks5port)
	if err := server.ListenAndServe("tcp", s.socks5port); err != nil {
		panic(err)
	}
}

type CustomDialer struct {
	s *Server
}

func (d *CustomDialer) Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	d.s.SendReqToClient(addr)
	return net.Dial(network, addr)
}
