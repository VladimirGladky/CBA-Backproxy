package server

import (
	"CBA-Backproxy/internal/client"
	"CBA-Backproxy/pkg/logger"
	"context"
	"fmt"
	"github.com/google/uuid"

	//"github.com/google/uuid"
	"github.com/things-go/go-socks5"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

const (
	defaultTimeout = time.Second
)

type Server struct {
	listener   net.Listener
	socks5port string
	ctx        context.Context
	clientConn net.Conn
	Clients    map[string]client.Client
	mu         sync.Mutex
}

func NewServer(ctx context.Context, socks5port string) *Server {
	return &Server{
		ctx:        ctx,
		socks5port: socks5port,
		Clients:    make(map[string]client.Client),
	}
}

func (s *Server) Run(host, port string) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	var err error
	s.listener, err = net.Listen("tcp", host+":"+port)
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Failed to listen on port 18000", zap.Error(err))
	}
	logger.GetLoggerFromCtx(s.ctx).Info(s.ctx, "Server started on port 18000")
	for {
		s.clientConn, err = s.listener.Accept()
		if err != nil {
			logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Failed to accept connection", zap.Error(err))
			continue
		}
		go s.handleConnection(s.clientConn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Failed to read from client", zap.Error(err))
		return
	}
	if string(buffer[:n]) != ":10800" {
		logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Invalid data received from client", zap.String("data", string(buffer[:n])))
		return
	}
	id := uuid.New().String()
	logger.GetLoggerFromCtx(s.ctx).Info(s.ctx, "Accepted connection from client", zap.String("id", id))
	s.mu.Lock()
	s.Clients[id] = client.Client{Ip: conn.RemoteAddr().String(), Ctx: s.ctx, Conn: conn}
	s.mu.Unlock()
	buffer = make([]byte, 1024)
	for {
		if conn != nil {
			n, err = conn.Read(buffer)
			if err != nil {
				logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Failed to read from client", zap.Error(err))
				return
			}
			logger.GetLoggerFromCtx(s.ctx).Info(s.ctx, "Received data from client", zap.String("data", string(buffer[:n])))
		}
		time.Sleep(defaultTimeout)
	}
}

func (s *Server) CloseAllConnections() {
	for _, client := range s.Clients {
		client.Conn.Close()
	}
}

func (s *Server) SendReqToClient(addr string) {
	keys := make([]string, 0, len(s.Clients))
	for k := range s.Clients {
		keys = append(keys, k)
	}
	randomKey := keys[rand.Intn(len(keys))]
	s.clientConn = s.Clients[randomKey].Conn
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
	if err := server.ListenAndServe("tcp", ":"+s.socks5port); err != nil {
		logger.GetLoggerFromCtx(s.ctx).Fatal(s.ctx, "Failed to start socks5 server", zap.Error(err))
	}
}

type CustomDialer struct {
	s *Server
}

func (d *CustomDialer) Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	d.s.SendReqToClient(addr)
	fmt.Println(addr)
	fmt.Println(network)
	return net.Dial(network, addr)
}
