package logger

import (
	"github.com/mrhoseah/dolphin/internal/logger"
	"go.uber.org/zap"
)

// Logger represents the application logger
type Logger = zap.Logger

// New creates a new logger instance
func New(level, output string) *Logger {
	return logger.New(level, output)
}

