package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "messenger-service/internal/core/config"
	core_logger "messenger-service/internal/core/logger"
	core_pgx_pool "messenger-service/internal/core/repository/postgres/pool/pgx"
	core_http_server "messenger-service/internal/core/transport/http/server"
	core_kafka "messenger-service/internal/core/transport/kafka"
	listing_client "messenger-service/internal/clients/listing"
	"messenger-service/internal/grpc/listingpb"
	messenger_repository "messenger-service/internal/features/messenger/repository"
	messenger_service "messenger-service/internal/features/messenger/service"
	messenger_transport_http "messenger-service/internal/features/messenger/transport/http"
	messenger_transport_kafka "messenger-service/internal/features/messenger/transport/kafka"
	messenger_transport_ws "messenger-service/internal/features/messenger/transport/ws"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title			Messenger Service API
// @version		1.0
// @description	Сервис переписки покупателей и продавцов auto-platform.
// @BasePath		/api/messenger
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

	// Init listing-service gRPC client — единственный синхронный поход в другой
	// сервис, только чтобы узнать продавца при создании треда.
	grpCfg := listing_client.NewConfigMust()
	grpcConn, err := grpc.NewClient(
		grpCfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{
			"methodConfig": [{
				"name": [{"service": "listingpb.ListingService"}],
				"retryPolicy": {
					"maxAttempts": 3,
					"initialBackoff": "0.1s",
					"maxBackoff": "2s",
					"backoffMultiplier": 2.0,
					"retryableStatusCodes": ["UNAVAILABLE"]
				}
			}]
		}`),
	)
	if err != nil {
		logger.Fatal("failed to connect to listing-service gRPC", zap.Error(err))
	}
	defer func() {
		if err := grpcConn.Close(); err != nil {
			logger.Error("grpc conn close error", zap.Error(err))
		}
	}()
	listingClient := listing_client.NewGRPCClient(listingpb.NewListingServiceClient(grpcConn))

	// Init Kafka producer (публикация message.sent для фан-аута между репликами)
	kafkaProducer, err := messenger_transport_kafka.NewProducer(core_kafka.NewProducerConfigMust())
	if err != nil {
		logger.Fatal("failed to init kafka producer", zap.Error(err))
	}
	defer kafkaProducer.Close()

	// Init Repository + Service
	repo := messenger_repository.NewRepository(postgresPool)
	svc := messenger_service.NewService(repo, listingClient, kafkaProducer, logger)

	// Init WS Hub + handler — Hub передаётся и в HTTP-хендлер (регистрирует
	// соединения), и в Kafka-консьюмер (доставляет фан-аут локально подключённым)
	hub := messenger_transport_ws.NewHub()
	wsHandler := messenger_transport_ws.NewHandler(hub, svc)

	// Init Kafka fanout consumer
	consumer, err := messenger_transport_kafka.NewConsumer(
		core_kafka.NewConsumerConfigMust(),
		hub,
		core_kafka.TopicMessageSent,
		logger,
	)
	if err != nil {
		logger.Fatal("failed to init kafka fanout consumer", zap.Error(err))
	}

	go func() {
		if err := consumer.Run(ctx); err != nil {
			logger.Fatal("kafka fanout consumer run error", zap.Error(err))
		}
	}()

	httpCfg := core_http_server.NewConfigMust()

	// Init HTTP Handler & Router
	httpHandler := messenger_transport_http.NewHandler(svc, wsHandler)
	router := httpHandler.InitRoutes(logger, httpCfg.AllowedOrigins)

	// Init HTTP Server
	httpServer := core_http_server.NewHTTPServer(
		httpCfg,
		logger,
		router,
	)

	logger.Info("starting messenger-service", zap.String("addr", os.Getenv("HTTP_ADDR")))

	if err := httpServer.Run(ctx); err != nil {
		logger.Fatal("HTTP server error", zap.Error(err))
	}
}
