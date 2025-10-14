package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new logger instance
func New(level, format string) *zap.Logger {
	var config zap.Config

	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Set log level
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Set output
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		// Fallback to a basic logger
		logger = zap.NewNop()
	}

	return logger
}

// NewFileLogger creates a logger that writes to a file
func NewFileLogger(level, format, filepath string) *zap.Logger {
	var config zap.Config

	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Set log level
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Set output to file
	config.OutputPaths = []string{filepath}
	config.ErrorOutputPaths = []string{filepath}

	logger, err := config.Build()
	if err != nil {
		// Fallback to stdout
		return New(level, format)
	}

	return logger
}

// NewTestLogger creates a logger suitable for testing
func NewTestLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	config.OutputPaths = []string{os.DevNull}
	config.ErrorOutputPaths = []string{os.DevNull}

	logger, err := config.Build()
	if err != nil {
		return zap.NewNop()
	}

	return logger
}
