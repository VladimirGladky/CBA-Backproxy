package server

import (
	"CBA-Backproxy/internal/client"
	"CBA-Backproxy/pkg/logger"
	"context"
	"github.com/google/uuid"
	"github.com/things-go/go-socks5"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout = time.Second
	clientTTL      = 30 * time.Second
)

type Server struct {
	listener         net.Listener
	socks5port       string
	ctx              context.Context
	clientConn       net.Conn
	FreeClients      map[string]*client.Client
	BusyClients      map[string]*client.Client
	FreeClientsCount int
	mu               sync.Mutex
	remoteAddr       string
}

func NewServer(ctx context.Context, socks5port string) *Server {
	s := &Server{
		ctx:         ctx,
		socks5port:  socks5port,
		FreeClients: make(map[string]*client.Client),
		BusyClients: make(map[string]*client.Client),
	}
	go s.startClientTTLChecker()
	return s
}

func (s *Server) startClientTTLChecker() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkClientTTL()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Server) checkClientTTL() {
	now := time.Now()
	for id, cl := range s.BusyClients {
		activePrograms := make([]client.Program, 0, len(cl.Cp))
		for _, p := range cl.Cp {
			if now.Sub(p.LastActivity) <= clientTTL {
				activePrograms = append(activePrograms, p)
			}
		}
		if len(activePrograms) == 0 {
			s.FreeClients[id] = cl
			s.FreeClientsCount++
			delete(s.BusyClients, id)
			logger.GetLoggerFromCtx(s.ctx).Info(s.ctx, "Client moved back to free pool due to TTL", zap.String("id", id))
		} else if len(activePrograms) != len(cl.Cp) {
			cl.Cp = activePrograms
		}
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

	s.FreeClients[id] = &client.Client{
		Ctx:  s.ctx,
		Conn: conn,
		Cp:   []client.Program{},
	}
	s.FreeClientsCount++
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
	for _, cl := range s.FreeClients {
		err := cl.Conn.Close()
		if err != nil {
			logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Failed to close connection", zap.Error(err))
		}
	}
}

func (s *Server) SendReqToClient(addr string, remoteAddr string) {
	for _, cl := range s.BusyClients {
		for i := range cl.Cp {
			if cl.Cp[i].AddrClientProgram == remoteAddr {
				s.mu.Lock()
				cl.Cp[i].LastActivity = time.Now()
				s.mu.Unlock()
				s.clientConn = cl.Conn
				_, err := s.clientConn.Write([]byte(addr))
				if err != nil {
					logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Failed to send request to c", zap.Error(err))
				}
				return
			}
		}
	}
	if len(s.FreeClients) == 0 {
		c := 10000
		for _, cl := range s.BusyClients {
			for _, a := range cl.Cp {
				if len(a.AddrClientProgram) < c {
					c = len(a.AddrClientProgram)
				}
			}
		}
		for _, cl := range s.BusyClients {
			for _, a := range cl.Cp {
				if len(a.AddrClientProgram) == c {
					s.mu.Lock()
					a.LastActivity = time.Now()
					s.mu.Unlock()
					s.clientConn = cl.Conn
					cl.Cp = append(cl.Cp, a)
					_, err := s.clientConn.Write([]byte(addr))
					if err != nil {
						logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Failed to send request to c", zap.Error(err))
					}
					return
				}
			}
		}
	}
	put := false
	if s.FreeClientsCount > 0 {
		var c *client.Client
		keys := make([]string, 0, len(s.FreeClients))
		for k := range s.FreeClients {
			keys = append(keys, k)
		}
		randomKey := keys[rand.Intn(len(keys))]
		s.mu.Lock()
		defer s.mu.Unlock()
		for _, a := range s.FreeClients[randomKey].Cp {
			if a.AddrClientProgram == "" {
				put = true
				a.AddrClientProgram = remoteAddr
				a.LastActivity = time.Now()
			}
		}
		if !put {
			s.FreeClients[randomKey].Cp = append(s.FreeClients[randomKey].Cp, client.Program{
				AddrClientProgram: remoteAddr,
				LastActivity:      time.Now(),
			})
		}
		c = s.FreeClients[randomKey]
		s.BusyClients[randomKey] = c
		delete(s.FreeClients, randomKey)
		s.FreeClientsCount--
		s.clientConn = s.BusyClients[randomKey].Conn
		_, err := s.clientConn.Write([]byte(addr))
		if err != nil {
			logger.GetLoggerFromCtx(s.ctx).Error(s.ctx, "Failed to send request to c", zap.Error(err))
		}
	}
}

func (s *Server) RunSocks5() {
	logger.GetLoggerFromCtx(s.ctx).Info(s.ctx, "Starting socks5 server")
	customDialer := &CustomDialer{
		s: s,
	}
	server := socks5.NewServer(
		socks5.WithDialAndRequest(customDialer.Dial),
		socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
	)
	if err := server.ListenAndServe("tcp", ":10800"); err != nil {
		logger.GetLoggerFromCtx(s.ctx).Fatal(s.ctx, "Failed to start socks5 server", zap.Error(err))
	}
}

type CustomDialer struct {
	s *Server
}

func (d *CustomDialer) Dial(ctx context.Context, network, addr string, request *socks5.Request) (net.Conn, error) {
	ra := strings.Split(request.RemoteAddr.String(), ":")
	d.s.remoteAddr = ra[0]
	logger.GetLoggerFromCtx(d.s.ctx).Info(d.s.ctx, "Received request from client", zap.String("address", d.s.remoteAddr))
	d.s.SendReqToClient(addr, d.s.remoteAddr)
	return net.Dial(network, addr)
}
