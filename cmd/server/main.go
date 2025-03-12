package main

import (
	"CBA-Backproxy/internal/server"
	"CBA-Backproxy/pkg/logger"
	"context"
	"go.uber.org/zap"
	"time"
)

func main() {
	ctx := context.Background()
	ctx, err := logger.New(ctx)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "Failed to create logger", zap.Error(err))
	}
	server := server.NewServer(ctx)
	server.Run()
	for {
		if server.Process() {
			server.RunSocks5()
		}
		time.Sleep(time.Second)
	}
}
