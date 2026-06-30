package core_grpc_server

import (
	"context"
	"fmt"
	"net"

	core_logger "listing-service/internal/core/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	server *grpc.Server
	config Config
	log    *core_logger.Logger
}

func NewGRPCServer(config Config, log *core_logger.Logger, opts ...grpc.ServerOption) *GRPCServer {
	return &GRPCServer{
		server: grpc.NewServer(opts...),
		config: config,
		log:    log,
	}
}

// Server возвращает *grpc.Server для регистрации сервисов (RegisterXxxServer).
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

func (s *GRPCServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("listen gRPC %s: %w", s.config.Addr, err)
	}

	serverErr := make(chan error, 1)
	go func() {
		s.log.Info("gRPC server listening", zap.String("addr", s.config.Addr))
		if err := s.server.Serve(lis); err != nil {
			serverErr <- fmt.Errorf("gRPC serve: %w", err)
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		s.server.GracefulStop()
		s.log.Debug("gRPC server stopped gracefully")
		return nil
	case err := <-serverErr:
		return err
	}
}
