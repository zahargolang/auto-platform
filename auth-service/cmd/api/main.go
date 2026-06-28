package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/zosinkin/social_network/internal/core/config"
	core_logger "github.com/zosinkin/social_network/internal/core/logger"
	core_pgx_pool "github.com/zosinkin/social_network/internal/core/repository/postgres/pool/pgx"
	core_http_server "github.com/zosinkin/social_network/internal/core/transport/http/server"
	core_kafka "github.com/zosinkin/social_network/internal/core/transport/kafka"
	auth_postgres_repository "github.com/zosinkin/social_network/internal/features/auth/repository"
	auth_service "github.com/zosinkin/social_network/internal/features/auth/service"
	auth_transport_http "github.com/zosinkin/social_network/internal/features/auth/transport/http"
	transport_kafka "github.com/zosinkin/social_network/internal/features/auth/transport/kafka"
	"go.uber.org/zap"
)

// @title			Auth Service API
// @version		1.0
// @description	Сервис аутентификации и регистрации пользователей kolesa.
// @BasePath		/api/auth
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	cfg := core_config.NewConfigMust()
	kafkaCfg := core_kafka.NewProducerConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()


	//Init Logger
	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("appplication time zone", zap.Any("zone", time.Local))


	//Init PostgreSQL
	logger.Debug("initializing postgres connection pool")
	postgresPool, err := core_pgx_pool.NewPool(
		ctx,
		core_pgx_pool.NewConfigMust(),
	)
	if err != nil {
		logger.Fatal("failed to init postgres connecion pool", zap.Error(err))
	}
	defer postgresPool.Close()


	//Init Repo
	logger.Debug("Initializing feature", zap.String("feature", "auth"))
	userRepo := auth_postgres_repository.NewUsersRepo(postgresPool)
	refreshRepo := auth_postgres_repository.NewRefreshTokenRepo(postgresPool)

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if jwtSecret == nil {
		logger.Fatal("JWT_SECRET is not set")
	}

	accessTokenTTL, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_EXPIRY"))
	if err != nil {
		logger.Fatal("invalid TOKENTTL", zap.Error(err))
	}


	//Init kafka Producer
	KafkaProducer, err := transport_kafka.NewProducer(kafkaCfg)
	if err != nil {
		logger.Fatal("Failed to init kafka", zap.Error(err))
	}
	defer KafkaProducer.Close()


	//Init service
	authService := auth_service.NewAuthService(
		userRepo,
		refreshRepo,
		jwtSecret,
		accessTokenTTL,
		KafkaProducer,
		logger,
	)


	//Init Handler
	authHTTPHandler := auth_transport_http.NewAuthHTTPHandler(authService)


	
	//Init router
	router := authHTTPHandler.InitRoutes(authService)


	//Init Server
	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		router,
	)
	logger.Info("Staring HTTP server")

	if err := httpServer.Run(ctx); err != nil {
		logger.Fatal("HTTP server error", zap.Error(err))
	}
}
