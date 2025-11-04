package grpc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"go.uber.org/zap"
)

// Gateway represents a gRPC-Gateway server
type Gateway struct {
	mux     *runtime.ServeMux
	config  *GatewayConfig
	logger  *zap.Logger
	handlers []HandlerRegistration
}

// GatewayConfig represents gateway configuration
type GatewayConfig struct {
	Port          int    `yaml:"port" json:"port"`
	GRPCEndpoint  string `yaml:"grpc_endpoint" json:"grpc_endpoint"` // e.g., "localhost:9090"
	EnableCORS    bool   `yaml:"enable_cors" json:"enable_cors"`
	EnableSwagger bool   `yaml:"enable_swagger" json:"enable_swagger"`
}

// DefaultGatewayConfig returns default gateway configuration
func DefaultGatewayConfig() *GatewayConfig {
	return &GatewayConfig{
		Port:         8080,
		GRPCEndpoint: "localhost:9090",
		EnableCORS:   true,
		EnableSwagger: false,
	}
}

// HandlerRegistration represents a handler registration function
type HandlerRegistration func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error

// NewGateway creates a new gRPC-Gateway
func NewGateway(config *GatewayConfig, logger *zap.Logger) (*Gateway, error) {
	if config == nil {
		config = DefaultGatewayConfig()
	}

	// Create ServeMux
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(runtime.DefaultHeaderMatcher),
	)

	return &Gateway{
		mux:      mux,
		config:   config,
		logger:   logger,
		handlers: []HandlerRegistration{},
	}, nil
}

// RegisterHandler registers a handler with the gateway
func (g *Gateway) RegisterHandler(handler HandlerRegistration) {
	g.handlers = append(g.handlers, handler)
}

// Start starts the gRPC-Gateway server
func (g *Gateway) Start(ctx context.Context) error {
	// Create gRPC connection
	conn, err := grpc.DialContext(
		ctx,
		g.config.GRPCEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}
	defer conn.Close()

	// Register all handlers
	for _, handler := range g.handlers {
		if err := handler(ctx, g.mux, conn); err != nil {
			return fmt.Errorf("failed to register handler: %w", err)
		}
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", g.config.Port),
		Handler: g.mux,
	}

	g.logger.Info("gRPC-Gateway starting",
		zap.Int("port", g.config.Port),
		zap.String("grpc_endpoint", g.config.GRPCEndpoint))

	// Start server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start gateway: %w", err)
	}

	return nil
}

// GetMux returns the underlying ServeMux
func (g *Gateway) GetMux() *runtime.ServeMux {
	return g.mux
}

// customHeaderMatcher matches custom headers
func customHeaderMatcher(key string) (string, bool) {
	switch key {
	case "X-Request-Id":
		return key, true
	case "X-User-Id":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

