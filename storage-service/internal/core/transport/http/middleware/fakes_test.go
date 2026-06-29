package core_middleware

import (
	core_logger "storage-service/internal/core/logger"

	"go.uber.org/zap"
)

func testLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}
