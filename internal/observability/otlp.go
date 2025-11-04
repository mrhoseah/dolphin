package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
)

// OTLPConfig represents OTLP exporter configuration
type OTLPConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Endpoint    string `yaml:"endpoint" json:"endpoint"`
	Protocol    string `yaml:"protocol" json:"protocol"` // "grpc" or "http"
	Insecure    bool   `yaml:"insecure" json:"insecure"`
	Headers     map[string]string `yaml:"headers" json:"headers"`
}

// DefaultOTLPConfig returns default OTLP configuration
func DefaultOTLPConfig() *OTLPConfig {
	return &OTLPConfig{
		Enabled:  false,
		Endpoint: "localhost:4317",
		Protocol: "grpc",
		Insecure: true,
		Headers:  make(map[string]string),
	}
}

// SetupOTLPTracing sets up OTLP tracing exporter
func SetupOTLPTracing(ctx context.Context, config *OTLPConfig, serviceName, version, environment string, logger *zap.Logger) (*sdktrace.TracerProvider, error) {
	if config == nil || !config.Enabled {
		return nil, fmt.Errorf("OTLP tracing not enabled")
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
			semconv.DeploymentEnvironmentKey.String(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP exporter
	var exporter sdktrace.SpanExporter
	switch config.Protocol {
	case "grpc":
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(config.Endpoint),
		}
		if config.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		if len(config.Headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(config.Headers))
		}

		client := otlptracegrpc.NewClient(opts...)
		exporter, err = otlptrace.New(ctx, client)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP gRPC exporter: %w", err)
		}
	case "http":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(config.Endpoint),
		}
		if config.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		if len(config.Headers) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(config.Headers))
		}

		exporter, err = otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP HTTP exporter: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported OTLP protocol: %s", config.Protocol)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("OTLP tracing configured",
		zap.String("endpoint", config.Endpoint),
		zap.String("protocol", config.Protocol))

	return tp, nil
}

// SetupPrometheusMetrics sets up Prometheus metrics exporter
func SetupPrometheusMetrics(serviceName, version string, logger *zap.Logger) (*metric.MeterProvider, error) {
	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create meter provider
	mp := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)

	// Set global meter provider
	otel.SetMeterProvider(mp)

	logger.Info("Prometheus metrics configured",
		zap.String("service", serviceName))

	return mp, nil
}

// GetPrometheusExporter returns the Prometheus exporter for HTTP handler
func GetPrometheusExporter() (*prometheus.Exporter, error) {
	// This would need to be stored globally or in a manager
	// For now, this is a placeholder
	return prometheus.New()
}

