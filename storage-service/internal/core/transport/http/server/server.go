package core_http_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	core_logger "storage-service/internal/core/logger"
)

type HTTPServer struct {
	server *http.Server
	config Config
	log    *core_logger.Logger
}

func NewHTTPServer(config Config, log *core_logger.Logger, handler http.Handler) *HTTPServer {
	return &HTTPServer{
		config: config,
		log:    log,
		server: &http.Server{
			Addr:    config.Addr,
			Handler: handler,
		},
	}
}

func (s *HTTPServer) Run(ctx context.Context) error {
	serverErr := make(chan error, 1)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("listen and serve: %w", err)
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		s.log.Debug("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		s.log.Debug("HTTP server stopped gracefully")
		return nil

	case err := <-serverErr:
		return err
	}
}
