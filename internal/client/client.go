package client

import (
	"CBA-Backproxy/pkg/logger"
	"context"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strings"
)

type Client struct {
	Ctx  context.Context
	Ip   string
	Conn net.Conn
}

func NewClient(ctx context.Context) *Client {
	return &Client{
		Ctx: ctx,
	}
}

func reverseDNSLookup(ip string) (string, error) {
	names, err := net.LookupAddr(ip)
	if err != nil {
		return "", err
	}
	if len(names) > 0 {
		return strings.TrimSuffix(names[0], "."), nil
	}
	return "", fmt.Errorf("no domain found for IP: %s", ip)
}

func (c *Client) Connect() {
	serverConn, err := net.Dial("tcp", "localhost:18000")
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
