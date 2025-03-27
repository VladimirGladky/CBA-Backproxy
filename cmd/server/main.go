package main

import (
	"CBA-Backproxy/internal/config"
	"CBA-Backproxy/internal/server"
	"CBA-Backproxy/pkg/logger"
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"
)

const (
	defaultTimeout = time.Second
)

func main() {
	ctx := context.Background()
	ctx, err := logger.New(ctx)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "Failed to create logger", zap.Error(err))
	}
	cfg, err := config.NewConfig()
	fmt.Println(cfg)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "Failed to create config", zap.Error(err))
	}
	srv := server.NewServer(ctx, cfg.Socks5Port)
	go func() {
		srv.Run(cfg.Host, cfg.Port)
	}()
	for {
		if len(srv.Clients) != 0 {
			go func() {
				srv.RunSocks5()
			}()
			break
		}
		time.Sleep(defaultTimeout)
	}
	select {}
}
