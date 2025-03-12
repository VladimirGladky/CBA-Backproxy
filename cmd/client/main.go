package main

import (
	"CBA-Backproxy/internal/client"
	"CBA-Backproxy/pkg/logger"
	"context"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	ctx, err := logger.New(ctx)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "Failed to create logger", zap.Error(err))
	}
	cl := client.NewClient(ctx)
	cl.Connect()
}
