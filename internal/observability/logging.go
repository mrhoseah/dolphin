package observability

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel represents log levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// LogConfig represents logging configuration
type LogConfig struct {
	Level       LogLevel `yaml:"level" json:"level"`
	Format      string   `yaml:"format" json:"format"` // json, console
	Output      string   `yaml:"output" json:"output"` // stdout, stderr, file
	FilePath    string   `yaml:"file_path" json:"file_path"`
	MaxSize     int      `yaml:"max_size" json:"max_size"`     // MB
	MaxBackups  int      `yaml:"max_backups" json:"max_backups"`
	MaxAge      int      `yaml:"max_age" json:"max_age"`       // days
	Compress    bool     `yaml:"compress" json:"compress"`
	Development bool     `yaml:"development" json:"development"`
	Caller      bool     `yaml:"caller" json:"caller"`
	Stacktrace  bool     `yaml:"stacktrace" json:"stacktrace"`
}

// DefaultLogConfig returns default logging configuration
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:       InfoLevel,
		Format:      "json",
		Output:      "stdout",
		FilePath:    "logs/app.log",
		MaxSize:     100, // 100MB
		MaxBackups:  3,
		MaxAge:      28, // 28 days
		Compress:    true,
		Development: false,
		Caller:      true,
		Stacktrace:  false,
	}
}

// LoggerManager manages application logging
type LoggerManager struct {
	logger *zap.Logger
	config *LogConfig
}

// NewLoggerManager creates a new logger manager
func NewLoggerManager(config *LogConfig) (*LoggerManager, error) {
	if config == nil {
		config = DefaultLogConfig()
	}

	// Create zap config
	zapConfig := zap.NewProductionConfig()
	
	// Set log level
	switch config.Level {
	case DebugLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case InfoLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case WarnLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case ErrorLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case FatalLevel:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Set format
	if config.Format == "console" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	// Set development mode
	if config.Development {
		zapConfig.Development = true
		zapConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		zapConfig.EncoderConfig = zap.NewProductionEncoderConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Set caller and stacktrace
	zapConfig.DisableCaller = !config.Caller
	zapConfig.DisableStacktrace = !config.Stacktrace

	// Set output
	if config.Output == "file" && config.FilePath != "" {
		zapConfig.OutputPaths = []string{config.FilePath}
		zapConfig.ErrorOutputPaths = []string{config.FilePath}
	}

	// Build logger
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &LoggerManager{
		logger: logger,
		config: config,
	}, nil
}

// GetLogger returns the underlying zap logger
func (lm *LoggerManager) GetLogger() *zap.Logger {
	return lm.logger
}

// GetSugarLogger returns a sugar logger
func (lm *LoggerManager) GetSugarLogger() *zap.SugaredLogger {
	return lm.logger.Sugar()
}

// WithContext creates a logger with context fields
func (lm *LoggerManager) WithContext(ctx context.Context) *zap.Logger {
	logger := lm.logger

	// Add request ID if present
	if requestID := ctx.Value("request_id"); requestID != nil {
		logger = logger.With(zap.String("request_id", fmt.Sprintf("%v", requestID)))
	}

	// Add user ID if present
	if userID := ctx.Value("user_id"); userID != nil {
		logger = logger.With(zap.String("user_id", fmt.Sprintf("%v", userID)))
	}

	// Add trace ID if present
	if traceID := ctx.Value("trace_id"); traceID != nil {
		logger = logger.With(zap.String("trace_id", fmt.Sprintf("%v", traceID)))
	}

	// Add span ID if present
	if spanID := ctx.Value("span_id"); spanID != nil {
		logger = logger.With(zap.String("span_id", fmt.Sprintf("%v", spanID)))
	}

	return logger
}

// WithFields creates a logger with additional fields
func (lm *LoggerManager) WithFields(fields map[string]interface{}) *zap.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return lm.logger.With(zapFields...)
}

// LogHTTPRequest logs an HTTP request
func (lm *LoggerManager) LogHTTPRequest(r *http.Request, statusCode int, duration time.Duration, size int64) {
	fields := []zap.Field{
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("query", r.URL.RawQuery),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.Int64("size", size),
	}

	// Add request ID if present
	if requestID := r.Context().Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", fmt.Sprintf("%v", requestID)))
	}

	// Add user ID if present
	if userID := r.Context().Value("user_id"); userID != nil {
		fields = append(fields, zap.String("user_id", fmt.Sprintf("%v", userID)))
	}

	// Log based on status code
	if statusCode >= 500 {
		lm.logger.Error("HTTP request", fields...)
	} else if statusCode >= 400 {
		lm.logger.Warn("HTTP request", fields...)
	} else {
		lm.logger.Info("HTTP request", fields...)
	}
}

// LogDatabaseQuery logs a database query
func (lm *LoggerManager) LogDatabaseQuery(operation, table string, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		lm.logger.Error("Database query", fields...)
	} else {
		lm.logger.Debug("Database query", fields...)
	}
}

// LogCacheOperation logs a cache operation
func (lm *LoggerManager) LogCacheOperation(operation, cacheName, key string, hit bool, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("cache_name", cacheName),
		zap.String("key", key),
		zap.Bool("hit", hit),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		lm.logger.Error("Cache operation", fields...)
	} else {
		lm.logger.Debug("Cache operation", fields...)
	}
}

// LogBusinessEvent logs a business event
func (lm *LoggerManager) LogBusinessEvent(eventType, status string, data map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("status", status),
	}

	for key, value := range data {
		fields = append(fields, zap.Any(key, value))
	}

	lm.logger.Info("Business event", fields...)
}

// LogError logs an error with context
func (lm *LoggerManager) LogError(err error, message string, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	lm.logger.Error(message, allFields...)
}

// LogPanic logs a panic and recovers
func (lm *LoggerManager) LogPanic(r interface{}) {
	lm.logger.Panic("Panic recovered", zap.Any("panic", r))
}

// LogSecurityEvent logs a security-related event
func (lm *LoggerManager) LogSecurityEvent(eventType, severity string, data map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("severity", severity),
		zap.String("category", "security"),
	}

	for key, value := range data {
		fields = append(fields, zap.Any(key, value))
	}

	// Log security events at warn level or higher
	switch severity {
	case "critical":
		lm.logger.Fatal("Security event", fields...)
	case "high":
		lm.logger.Error("Security event", fields...)
	case "medium":
		lm.logger.Warn("Security event", fields...)
	default:
		lm.logger.Info("Security event", fields...)
	}
}

// LogPerformance logs performance metrics
func (lm *LoggerManager) LogPerformance(operation string, duration time.Duration, metrics map[string]interface{}) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.String("category", "performance"),
	}

	for key, value := range metrics {
		fields = append(fields, zap.Any(key, value))
	}

	lm.logger.Info("Performance metric", fields...)
}

// LogAudit logs an audit event
func (lm *LoggerManager) LogAudit(action, resource, userID string, success bool, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("user_id", userID),
		zap.Bool("success", success),
		zap.String("category", "audit"),
	}

	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	lm.logger.Info("Audit event", fields...)
}

// LoggingMiddleware creates HTTP logging middleware
func LoggingMiddleware(logger *LoggerManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap response writer to capture status code and size
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			// Process request
			next.ServeHTTP(wrapped, r)
			
			// Log request
			duration := time.Since(start)
			logger.LogHTTPRequest(r, wrapped.statusCode, duration, wrapped.size)
		})
	}
}

// StructuredLogger provides structured logging methods
type StructuredLogger struct {
	logger *zap.Logger
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(logger *zap.Logger) *StructuredLogger {
	return &StructuredLogger{logger: logger}
}

// Debug logs a debug message
func (sl *StructuredLogger) Debug(msg string, fields ...zap.Field) {
	sl.logger.Debug(msg, fields...)
}

// Info logs an info message
func (sl *StructuredLogger) Info(msg string, fields ...zap.Field) {
	sl.logger.Info(msg, fields...)
}

// Warn logs a warning message
func (sl *StructuredLogger) Warn(msg string, fields ...zap.Field) {
	sl.logger.Warn(msg, fields...)
}

// Error logs an error message
func (sl *StructuredLogger) Error(msg string, fields ...zap.Field) {
	sl.logger.Error(msg, fields...)
}

// Fatal logs a fatal message
func (sl *StructuredLogger) Fatal(msg string, fields ...zap.Field) {
	sl.logger.Fatal(msg, fields...)
}

// With creates a child logger with fields
func (sl *StructuredLogger) With(fields ...zap.Field) *StructuredLogger {
	return &StructuredLogger{logger: sl.logger.With(fields...)}
}

// WithContext creates a child logger with context
func (sl *StructuredLogger) WithContext(ctx context.Context) *StructuredLogger {
	logger := sl.logger

	// Add context fields
	if requestID := ctx.Value("request_id"); requestID != nil {
		logger = logger.With(zap.String("request_id", fmt.Sprintf("%v", requestID)))
	}
	if userID := ctx.Value("user_id"); userID != nil {
		logger = logger.With(zap.String("user_id", fmt.Sprintf("%v", userID)))
	}
	if traceID := ctx.Value("trace_id"); traceID != nil {
		logger = logger.With(zap.String("trace_id", fmt.Sprintf("%v", traceID)))
	}

	return &StructuredLogger{logger: logger}
}

// LogLevelFromString converts string to LogLevel
func LogLevelFromString(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// LogConfigFromEnv creates log config from environment variables
func LogConfigFromEnv() *LogConfig {
	config := DefaultLogConfig()

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = LogLevelFromString(level)
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Format = format
	}
	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		config.Output = output
	}
	if filePath := os.Getenv("LOG_FILE_PATH"); filePath != "" {
		config.FilePath = filePath
	}
	if development := os.Getenv("LOG_DEVELOPMENT"); development == "true" {
		config.Development = true
	}

	return config
}
