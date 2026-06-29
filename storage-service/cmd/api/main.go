package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "storage-service/internal/core/config"
	core_logger "storage-service/internal/core/logger"
	core_s3 "storage-service/internal/core/s3"
	core_http_server "storage-service/internal/core/transport/http/server"
	storage_service "storage-service/internal/features/storage/service"
	transport_http "storage-service/internal/features/storage/transport/http"

	"go.uber.org/zap"
)

// @title			Storage Service API
// @version		1.0
// @description	Выдаёт presigned-ссылки на загрузку файлов в S3-совместимое хранилище kolesa.
// @BasePath		/api/storage
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	svc, err := storage_service.NewService(ctx, core_s3.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to init storage service", zap.Error(err))
	}

	handler := transport_http.NewHandler(svc)
	router := handler.InitRoutes(logger)

	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		router,
	)

	logger.Info("starting storage-service", zap.String("addr", os.Getenv("HTTP_ADDR")))

	if err := httpServer.Run(ctx); err != nil {
		logger.Fatal("HTTP server error", zap.Error(err))
	}
}
