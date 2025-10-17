package observability

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TraceConfig represents tracing configuration
type TraceConfig struct {
	Enabled     bool    `yaml:"enabled" json:"enabled"`
	ServiceName string  `yaml:"service_name" json:"service_name"`
	Version     string  `yaml:"version" json:"version"`
	Environment string  `yaml:"environment" json:"environment"`
	Sampler     string  `yaml:"sampler" json:"sampler"` // always_on, always_off, traceid_ratio
	Ratio       float64 `yaml:"ratio" json:"ratio"`

	// Exporters
	JaegerEndpoint string `yaml:"jaeger_endpoint" json:"jaeger_endpoint"`
	ZipkinEndpoint string `yaml:"zipkin_endpoint" json:"zipkin_endpoint"`

	// Headers
	TraceHeader string `yaml:"trace_header" json:"trace_header"`
	SpanHeader  string `yaml:"span_header" json:"span_header"`
}

// DefaultTraceConfig returns default tracing configuration
func DefaultTraceConfig() *TraceConfig {
	return &TraceConfig{
		Enabled:        true,
		ServiceName:    "dolphin-app",
		Version:        "1.0.0",
		Environment:    "development",
		Sampler:        "traceid_ratio",
		Ratio:          1.0,
		JaegerEndpoint: "http://localhost:14268/api/traces",
		ZipkinEndpoint: "http://localhost:9411/api/v2/spans",
		TraceHeader:    "X-Trace-Id",
		SpanHeader:     "X-Span-Id",
	}
}

// TracerManager manages distributed tracing
type TracerManager struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
	config   *TraceConfig
	logger   *zap.Logger
}

// NewTracerManager creates a new tracer manager
func NewTracerManager(config *TraceConfig, logger *zap.Logger) (*TracerManager, error) {
	if config == nil {
		config = DefaultTraceConfig()
	}

	if !config.Enabled {
		// Return a no-op tracer
		return &TracerManager{
			tracer: trace.NewNoopTracerProvider().Tracer("noop"),
			config: config,
			logger: logger,
		}, nil
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.Version),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create sampler
	var sampler sdktrace.Sampler
	switch config.Sampler {
	case "always_on":
		sampler = sdktrace.AlwaysSample()
	case "always_off":
		sampler = sdktrace.NeverSample()
	case "traceid_ratio":
		sampler = sdktrace.TraceIDRatioBased(config.Ratio)
	default:
		sampler = sdktrace.TraceIDRatioBased(config.Ratio)
	}

	// Create exporters
	var exporters []sdktrace.SpanExporter

	// Jaeger exporter
	if config.JaegerEndpoint != "" {
		jaegerExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
		if err != nil {
			logger.Warn("Failed to create Jaeger exporter", zap.Error(err))
		} else {
			exporters = append(exporters, jaegerExporter)
		}
	}

	// Zipkin exporter
	if config.ZipkinEndpoint != "" {
		zipkinExporter, err := zipkin.New(config.ZipkinEndpoint)
		if err != nil {
			logger.Warn("Failed to create Zipkin exporter", zap.Error(err))
		} else {
			exporters = append(exporters, zipkinExporter)
		}
	}

	if len(exporters) == 0 {
		return nil, fmt.Errorf("no exporters configured")
	}

	// Create multi-exporter
	multiExporter := &MultiSpanExporter{exporters: exporters}

	// Create tracer provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(multiExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(provider)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := provider.Tracer(config.ServiceName)

	return &TracerManager{
		tracer:   tracer,
		provider: provider,
		config:   config,
		logger:   logger,
	}, nil
}

// GetTracer returns the tracer
func (tm *TracerManager) GetTracer() trace.Tracer {
	return tm.tracer
}

// StartSpan starts a new span
func (tm *TracerManager) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tm.tracer.Start(ctx, name, opts...)
}

// StartSpanWithAttributes starts a span with attributes
func (tm *TracerManager) StartSpanWithAttributes(ctx context.Context, name string, attrs map[string]interface{}, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	spanOpts := make([]trace.SpanStartOption, len(opts))
	copy(spanOpts, opts)

	// Add attributes
	attributes := make([]attribute.KeyValue, 0, len(attrs))
	for key, value := range attrs {
		switch v := value.(type) {
		case string:
			attributes = append(attributes, attribute.String(key, v))
		case int:
			attributes = append(attributes, attribute.Int(key, v))
		case int64:
			attributes = append(attributes, attribute.Int64(key, v))
		case float64:
			attributes = append(attributes, attribute.Float64(key, v))
		case bool:
			attributes = append(attributes, attribute.Bool(key, v))
		default:
			attributes = append(attributes, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}
	if len(attributes) > 0 {
		spanOpts = append(spanOpts, trace.WithAttributes(attributes...))
	}
	return tm.tracer.Start(ctx, name, spanOpts...)
}

// AddSpanEvent adds an event to the current span
func (tm *TracerManager) AddSpanEvent(ctx context.Context, name string, attrs map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		attributes := make([]attribute.KeyValue, 0, len(attrs))
		for key, value := range attrs {
			switch v := value.(type) {
			case string:
				attributes = append(attributes, attribute.String(key, v))
			case int:
				attributes = append(attributes, attribute.Int(key, v))
			case int64:
				attributes = append(attributes, attribute.Int64(key, v))
			case float64:
				attributes = append(attributes, attribute.Float64(key, v))
			case bool:
				attributes = append(attributes, attribute.Bool(key, v))
			default:
				attributes = append(attributes, attribute.String(key, fmt.Sprintf("%v", v)))
			}
		}
		span.AddEvent(name, trace.WithAttributes(attributes...))
	}
}

// SetSpanAttributes sets attributes on the current span
func (tm *TracerManager) SetSpanAttributes(ctx context.Context, attrs map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		for key, value := range attrs {
			switch v := value.(type) {
			case string:
				span.SetAttributes(attribute.String(key, v))
			case int:
				span.SetAttributes(attribute.Int(key, v))
			case int64:
				span.SetAttributes(attribute.Int64(key, v))
			case float64:
				span.SetAttributes(attribute.Float64(key, v))
			case bool:
				span.SetAttributes(attribute.Bool(key, v))
			default:
				span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
			}
		}
	}
}

// SetSpanStatus sets the status of the current span
func (tm *TracerManager) SetSpanStatus(ctx context.Context, code codes.Code, description string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetStatus(code, description)
	}
}

// FinishSpan finishes the current span
func (tm *TracerManager) FinishSpan(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.End()
	}
}

// ExtractTraceContext extracts trace context from HTTP headers
func (tm *TracerManager) ExtractTraceContext(ctx context.Context, headers http.Header) context.Context {
	// Extract trace context from headers
	ctx = otel.GetTextMapPropagator().Extract(ctx, &HTTPHeaderCarrier{headers: headers})
	return ctx
}

// InjectTraceContext injects trace context into HTTP headers
func (tm *TracerManager) InjectTraceContext(ctx context.Context, headers http.Header) {
	otel.GetTextMapPropagator().Inject(ctx, &HTTPHeaderCarrier{headers: headers})
}

// GetTraceID returns the trace ID from context
func (tm *TracerManager) GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from context
func (tm *TracerManager) GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasSpanID() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// TracingMiddleware creates HTTP tracing middleware
func TracingMiddleware(tracer *TracerManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract trace context from request
			ctx := tracer.ExtractTraceContext(r.Context(), r.Header)

			// Start span
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := tracer.StartSpan(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))

			// Add span attributes
			span.SetAttributes(
				semconv.HTTPMethodKey.String(r.Method),
				semconv.HTTPURLKey.String(r.URL.String()),
				semconv.HTTPUserAgentKey.String(r.UserAgent()),
				semconv.HTTPClientIPKey.String(r.RemoteAddr),
			)

			// Add trace headers to response
			tracer.InjectTraceContext(ctx, w.Header())

			// Wrap response writer
			wrapped := &tracingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Set span status based on response
			if wrapped.statusCode >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrapped.statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}

			// Add final attributes
			span.SetAttributes(
				semconv.HTTPStatusCodeKey.Int(wrapped.statusCode),
				attribute.Int64("http.response_size", int64(wrapped.size)),
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
			)

			// Finish span
			span.End()
		})
	}
}

// DatabaseTracingMiddleware creates database tracing middleware
func DatabaseTracingMiddleware(tracer *TracerManager) func(string, string, func() error) error {
	return func(operation, table string, fn func() error) error {
		ctx := context.Background()
		spanName := fmt.Sprintf("db.%s", operation)

		ctx, span := tracer.StartSpan(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))
		span.SetAttributes(
			attribute.String("db.operation", operation),
			attribute.String("db.table", table),
		)

		err := fn()

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
		return err
	}
}

// CacheTracingMiddleware creates cache tracing middleware
func CacheTracingMiddleware(tracer *TracerManager) func(string, string, func() error) error {
	return func(operation, key string, fn func() error) error {
		ctx := context.Background()
		spanName := fmt.Sprintf("cache.%s", operation)

		ctx, span := tracer.StartSpan(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))
		span.SetAttributes(
			attribute.String("cache.operation", operation),
			attribute.String("cache.key", key),
		)

		err := fn()

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
		return err
	}
}

// MultiSpanExporter implements multiple span exporters
type MultiSpanExporter struct {
	exporters []sdktrace.SpanExporter
}

// ExportSpans exports spans to all configured exporters
func (mse *MultiSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	var lastErr error
	for _, exporter := range mse.exporters {
		if err := exporter.ExportSpans(ctx, spans); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Shutdown shuts down all exporters
func (mse *MultiSpanExporter) Shutdown(ctx context.Context) error {
	var lastErr error
	for _, exporter := range mse.exporters {
		if err := exporter.Shutdown(ctx); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// HTTPHeaderCarrier implements the TextMapCarrier interface for HTTP headers
type HTTPHeaderCarrier struct {
	headers http.Header
}

// Get returns the value for a key
func (c *HTTPHeaderCarrier) Get(key string) string {
	return c.headers.Get(key)
}

// Set sets the value for a key
func (c *HTTPHeaderCarrier) Set(key, value string) {
	c.headers.Set(key, value)
}

// Keys returns all keys
func (c *HTTPHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c.headers))
	for k := range c.headers {
		keys = append(keys, k)
	}
	return keys
}

// tracingResponseWriter wraps http.ResponseWriter for tracing
type tracingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

func (trw *tracingResponseWriter) WriteHeader(code int) {
	trw.statusCode = code
	trw.ResponseWriter.WriteHeader(code)
}

func (trw *tracingResponseWriter) Write(b []byte) (int, error) {
	size, err := trw.ResponseWriter.Write(b)
	trw.size += int64(size)
	return size, err
}

// TraceConfigFromEnv creates trace config from environment variables
func TraceConfigFromEnv() *TraceConfig {
	config := DefaultTraceConfig()

	if enabled := os.Getenv("TRACE_ENABLED"); enabled == "false" {
		config.Enabled = false
	}
	if serviceName := os.Getenv("TRACE_SERVICE_NAME"); serviceName != "" {
		config.ServiceName = serviceName
	}
	if version := os.Getenv("TRACE_VERSION"); version != "" {
		config.Version = version
	}
	if environment := os.Getenv("TRACE_ENVIRONMENT"); environment != "" {
		config.Environment = environment
	}
	if sampler := os.Getenv("TRACE_SAMPLER"); sampler != "" {
		config.Sampler = sampler
	}
	if jaegerEndpoint := os.Getenv("TRACE_JAEGER_ENDPOINT"); jaegerEndpoint != "" {
		config.JaegerEndpoint = jaegerEndpoint
	}
	if zipkinEndpoint := os.Getenv("TRACE_ZIPKIN_ENDPOINT"); zipkinEndpoint != "" {
		config.ZipkinEndpoint = zipkinEndpoint
	}

	return config
}
