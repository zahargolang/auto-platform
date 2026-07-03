package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	core_config "user-service/internal/core/config"
	core_logger "user-service/internal/core/logger"
	core_pgx_pool "user-service/internal/core/repository/postgres/pool/pgx"
	core_http_server "user-service/internal/core/transport/http/server"
	core_kafka "user-service/internal/core/transport/kafka"
	"user-service/internal/feauture/users/repository"
	"user-service/internal/feauture/users/service"
	transport_http "user-service/internal/feauture/users/transport/http"
	transport_kafka "user-service/internal/feauture/users/transport/kafka"

	"go.uber.org/zap"
)


// @title			User Service API
// @version		1.0
// @description	Сервис профилей пользователей auto-platform.
// @BasePath		/api/user
func main() {
	cfg := core_config.NewConfigMust()
	kafkaCfg := core_kafka.NewConsumerConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("appplication time zone", zap.Any("zone", time.Local))

	logger.Debug("initializing postgres connection pool")
	postgresPool, err := core_pgx_pool.NewPool(
		ctx,
		core_pgx_pool.NewConfigMust(),
	)
	if err != nil {
		logger.Fatal("failed to init postgres connecion pool", zap.Error(err))
	}
	defer postgresPool.Close()

	logger.Debug("Initializing feature", zap.String("feature", "auth"))
	repo := repository.NewRepo(postgresPool, logger)

	service := service.NewService(repo, logger)

	HTTPHandler := transport_http.NewHTTPHandler(service)

	dlqProducer, err := transport_kafka.NewProducer(core_kafka.NewProducerConfigMust())
	if err != nil {
		logger.Fatal("Failed to init DLQ producer", zap.Error(err))
	}
	defer dlqProducer.Close()

	consumer, err := transport_kafka.NewConsumer(kafkaCfg, service, core_kafka.TopicUserRegistered, logger, dlqProducer)
	if err != nil {
		logger.Fatal("Failed to init Consumer", zap.Error(err))
	}

	go func() {
		err := consumer.Run(ctx)
		if err != nil {
			logger.Fatal("Run kafka consumer error", zap.Error(err))
		}
	}()

	httpCfg := core_http_server.NewConfigMust()
	router := HTTPHandler.InitRoutes(logger, httpCfg.AllowedOrigins)

	httpServer := core_http_server.NewHTTPServer(
		httpCfg,
		logger,
		router,
	)

	logger.Info("Staring HTTP server")

	if err := httpServer.Run(ctx); err != nil {
		logger.Fatal("HTTP server error", zap.Error(err))
	}
}