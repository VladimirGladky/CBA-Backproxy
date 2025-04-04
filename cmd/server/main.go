package main

import (
	"CBA-Backproxy/internal/config"
	"CBA-Backproxy/internal/server"
	"CBA-Backproxy/pkg/logger"
	"context"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

const (
	defaultTimeout = time.Second
)

func main() {
	if _, err := os.Stat("./config/config.yaml"); os.IsNotExist(err) {
		log.Fatal("Конфиг не найден. Создайте config.yaml на основе config.example.yaml")
	}
	ctx := context.Background()
	ctx, err := logger.New(ctx)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "Failed to create logger", zap.Error(err))
	}
	cfg, err := config.NewConfig()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "Failed to create config", zap.Error(err))
	}
	srv := server.NewServer(ctx, cfg.Socks5Port)
	go func() {
		srv.Run(cfg.Host, cfg.Port)
	}()
	for {
		if len(srv.FreeClients) != 0 {
			go func() {
				srv.RunSocks5()
			}()
			break
		}
		time.Sleep(defaultTimeout)
	}
	select {}
}
