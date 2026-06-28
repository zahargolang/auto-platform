package core_middleware

import (
	"go.uber.org/zap"

	core_logger "github.com/zosinkin/social_network/internal/core/logger"
)

func testLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}
