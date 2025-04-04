package client

import (
	"CBA-Backproxy/pkg/logger"
	"context"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strings"
	"time"
)

type Program struct {
	LastActivity      time.Time
	AddrClientProgram string
}

type Client struct {
	Ctx  context.Context
	Conn net.Conn
	Cp   []Program
}

func NewClient(ctx context.Context) *Client {
	return &Client{
		Ctx: ctx,
	}
}

func reverseDNSLookup(ip string) (string, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
	names, err := resolver.LookupAddr(context.Background(), ip)
	if err != nil {
		return ip, nil
	}
	if len(names) > 0 {
		return strings.TrimSuffix(names[0], "."), nil
	}
	return ip, nil
}

func (c *Client) Connect(host, port string) {
	serverConn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		logger.GetLoggerFromCtx(c.Ctx).Error(c.Ctx, "Failed to connect to server", zap.Error(err))
		return
	}
	logger.GetLoggerFromCtx(c.Ctx).Info(c.Ctx, "Connected to server")
	_, err = serverConn.Write([]byte(":10800"))
	if err != nil {
		logger.GetLoggerFromCtx(c.Ctx).Error(c.Ctx, "Failed to send data to server", zap.Error(err))
		return
	}

	for {
		buffer := make([]byte, 1024)
		n, err := serverConn.Read(buffer)
		if err != nil {
			logger.GetLoggerFromCtx(c.Ctx).Error(c.Ctx, "Failed to read from server", zap.Error(err))
			return
		}

		addr := string(buffer[:n])
		logger.GetLoggerFromCtx(c.Ctx).Info(c.Ctx, "Received request from server", zap.String("address", addr))

		parts := strings.Split(addr, ":")
		if len(parts) != 2 {
			logger.GetLoggerFromCtx(c.Ctx).Error(c.Ctx, "Invalid address format", zap.String("address", addr))
			return
		}
		ip, port := parts[0], parts[1]

		domain, err := reverseDNSLookup(ip)
		fmt.Println(domain)
		if err != nil {
			logger.GetLoggerFromCtx(c.Ctx).Error(c.Ctx, "Failed to resolve IP to domain", zap.Error(err))
			return
		}

		protocol := "http://"
		if port == "443" {
			protocol = "https://"
		}

		fullURL := protocol + domain
		if port != "80" && port != "443" {
			fullURL += ":" + port
		}

		resp, err := http.Get(fullURL)
		if err != nil {
			logger.GetLoggerFromCtx(c.Ctx).Error(c.Ctx, "Failed to perform HTTP request", zap.Error(err))
			return
		}
		defer resp.Body.Close()

		responseData := make([]byte, 1024)
		n, err = resp.Body.Read(responseData)
		if err != nil {
			logger.GetLoggerFromCtx(c.Ctx).Error(c.Ctx, "Failed to read response body", zap.Error(err))
			return
		}
		_, err = serverConn.Write(responseData[:n])
		if err != nil {
			logger.GetLoggerFromCtx(c.Ctx).Error(c.Ctx, "Failed to send response to server", zap.Error(err))
			return
		}
	}
}
