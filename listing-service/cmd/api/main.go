package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "listing-service/internal/core/config"
	core_logger "listing-service/internal/core/logger"
	core_pgx_pool "listing-service/internal/core/repository/postgres/pool/pgx"
	core_goredis_cache "listing-service/internal/core/repository/redis/goredis"
	core_grpc_server "listing-service/internal/core/transport/grpc/server"
	core_http_server "listing-service/internal/core/transport/http/server"
	core_kafka "listing-service/internal/core/transport/kafka"
	listings_grpc "listing-service/internal/grpc"
	"listing-service/internal/grpc/listingpb"
	listings_repository "listing-service/internal/features/listings/repository"
	listings_service "listing-service/internal/features/listings/service"
	listings_transport_http "listing-service/internal/features/listings/transport/http"
	listings_transport_kafka "listing-service/internal/features/listings/transport/kafka"

	"go.uber.org/zap"
)

// @title			Listing Service API
// @version		1.0
// @description	Сервис объявлений о продаже автомобилей auto-platform.
// @BasePath		/api/listings
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

	// Init Logger
	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Init PostgreSQL
	logger.Debug("initializing postgres connection pool")
	postgresPool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to init postgres pool", zap.Error(err))
	}
	defer postgresPool.Close()

	// Init Kafka Producer
	kafkaCfg := core_kafka.NewProducerConfigMust()
	kafkaProducer, err := listings_transport_kafka.NewProducer(kafkaCfg)
	if err != nil {
		logger.Fatal("failed to init kafka producer", zap.Error(err))
	}
	defer kafkaProducer.Close()


	//Init Redis cache
	logger.Debug("initializing redis client")
	redisClient, err := core_goredis_cache.NewClient(ctx, core_goredis_cache.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to init redis client", zap.Error(err))
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("failed to close redis client:", zap.Error(err))
		}
	}()

	// Init Repository
	repo := listings_repository.NewRepository(postgresPool)
	cachedRepo := listings_repository.NewCachedRepository(repo, redisClient, logger)
	
	// Init Service
	svc := listings_service.NewService(cachedRepo, kafkaProducer, logger)

	// Init Handler & Router
	handler := listings_transport_http.NewListingsHandler(svc)
	httpCfg := core_http_server.NewConfigMust()
	router := handler.InitRoutes(logger, httpCfg.AllowedOrigins)

	// Init HTTP Server
	httpServer := core_http_server.NewHTTPServer(
		httpCfg,
		logger,
		router,
	)

	// Init gRPC Server
	grpcSrv := core_grpc_server.NewGRPCServer(core_grpc_server.NewConfigMust(), logger)
	listingpb.RegisterListingServiceServer(grpcSrv.Server(), listings_grpc.NewServer(svc))

	go func() {
		if err := grpcSrv.Run(ctx); err != nil {
			logger.Fatal("gRPC server error", zap.Error(err))
		}
	}()

	logger.Info("starting listing-service", zap.String("addr", os.Getenv("HTTP_ADDR")))

	if err := httpServer.Run(ctx); err != nil {
		logger.Fatal("HTTP server error", zap.Error(err))
	}
}
