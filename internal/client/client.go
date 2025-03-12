package client

import (
	"CBA-Backproxy/pkg/logger"
	"context"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	defaultTimeout = time.Second
)

type Client struct {
	ctx context.Context
}

func NewClient(ctx context.Context) *Client {
	return &Client{
		ctx: ctx,
	}
}

func (c *Client) Connect() {
	serverConn, err := net.Dial("tcp", "localhost:18000")
	if err != nil {
		logger.GetLoggerFromCtx(c.ctx).Error(c.ctx, "Failed to connect to server", zap.Error(err))
		return
	}
	logger.GetLoggerFromCtx(c.ctx).Info(c.ctx, "Connected to server")
	_, err = serverConn.Write([]byte(":10800"))
	if err != nil {
		logger.GetLoggerFromCtx(c.ctx).Error(c.ctx, "Failed to send data to server", zap.Error(err))
		return
	}

	for {
		buffer := make([]byte, 1024)
		n, err := serverConn.Read(buffer)
		if err != nil {
			logger.GetLoggerFromCtx(c.ctx).Error(c.ctx, "Failed to read from server", zap.Error(err))
			continue
		}

		addr := string(buffer[:n])
		logger.GetLoggerFromCtx(c.ctx).Info(c.ctx, "Received request from server", zap.String("address", addr))

		protocol := "http://"
		if strings.Contains(addr, ":443") {
			protocol = "https://"
		}

		fullURL := protocol + addr
		resp, err := http.Get(fullURL)
		if err != nil {
			logger.GetLoggerFromCtx(c.ctx).Error(c.ctx, "Failed to perform HTTP request", zap.Error(err))
			continue
		}
		defer resp.Body.Close()

		responseData := make([]byte, 1024)
		n, err = resp.Body.Read(responseData)
		if err != nil {
			logger.GetLoggerFromCtx(c.ctx).Error(c.ctx, "Failed to read response body", zap.Error(err))
			continue
		}
		logger.GetLoggerFromCtx(c.ctx).Info(c.ctx, "Received response from server", zap.String("data", string(responseData[:n])))
		_, err = serverConn.Write(responseData[:n])
		if err != nil {
			logger.GetLoggerFromCtx(c.ctx).Error(c.ctx, "Failed to send response to server", zap.Error(err))
		}
	}
}
