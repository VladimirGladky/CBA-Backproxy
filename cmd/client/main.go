package main

import (
	"CBA-Backproxy/internal/client"
	"CBA-Backproxy/internal/config"
	"CBA-Backproxy/pkg/logger"
	"context"
	"go.uber.org/zap"
	"time"
)

const (
	defaultTimeout = time.Second
)

func main() {
	baseCtx := context.Background()

	ctx1, err := logger.New(baseCtx)
	if err != nil {
		logger.GetLoggerFromCtx(baseCtx).Error(baseCtx, "Failed to create logger", zap.Error(err))
	}
	cfg, err := config.NewConfig()
	if err != nil {
		logger.GetLoggerFromCtx(ctx1).Error(ctx1, "Failed to create config", zap.Error(err))
	}
	cl := client.NewClient(ctx1)
	go cl.Connect(cfg.Host, cfg.Port)

	time.Sleep(defaultTimeout)
	ctx2, err := logger.New(baseCtx)
	if err != nil {
		logger.GetLoggerFromCtx(baseCtx).Error(baseCtx, "Failed to create logger", zap.Error(err))
	}
	cl1 := client.NewClient(ctx2)
	go cl1.Connect(cfg.Host, cfg.Port)

	time.Sleep(defaultTimeout)
	ctx3, err := logger.New(baseCtx)
	if err != nil {
		logger.GetLoggerFromCtx(baseCtx).Error(baseCtx, "Failed to create logger", zap.Error(err))
	}
	cl2 := client.NewClient(ctx3)
	go cl2.Connect(cfg.Host, cfg.Port)

	select {}
}
