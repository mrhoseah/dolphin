package observability

import (
	"context"
	"dolphin/internal/observability"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// SetupOTLPTracing sets up OTLP tracing exporter
func SetupOTLPTracing(ctx context.Context, config *OTLPConfig, serviceName, version, environment string, logger *zap.Logger) (*trace.TracerProvider, error) {
	return observability.SetupOTLPTracing(ctx, config, serviceName, version, environment, logger)
}

// OTLPConfig represents OTLP exporter configuration
type OTLPConfig = observability.OTLPConfig

// DefaultOTLPConfig returns default OTLP configuration
func DefaultOTLPConfig() *OTLPConfig {
	return observability.DefaultOTLPConfig()
}

// SetupPrometheusMetrics sets up Prometheus metrics exporter
func SetupPrometheusMetrics(serviceName, version string, logger *zap.Logger) (*metric.MeterProvider, error) {
	return observability.SetupPrometheusMetrics(serviceName, version, logger)
}

