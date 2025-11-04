package grpc

import (
	dolphingrpc "dolphin/internal/grpc"

	"go.uber.org/zap"
)

// Server represents a gRPC server
type Server = dolphingrpc.Server

// NewServer creates a new gRPC server
func NewServer(config *ServerConfig, logger *zap.Logger) (*Server, error) {
	return dolphingrpc.NewServer(config, logger)
}

// ServerConfig represents gRPC server configuration
type ServerConfig = dolphingrpc.ServerConfig

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() *ServerConfig {
	return dolphingrpc.DefaultServerConfig()
}

// KeepaliveConfig represents keepalive configuration
type KeepaliveConfig = dolphingrpc.KeepaliveConfig

// ServiceRegistration represents a service registration function
type ServiceRegistration = dolphingrpc.ServiceRegistration

// Gateway represents a gRPC-Gateway
type Gateway = dolphingrpc.Gateway

// NewGateway creates a new gRPC-Gateway
func NewGateway(config *GatewayConfig, logger *zap.Logger) (*Gateway, error) {
	return dolphingrpc.NewGateway(config, logger)
}

// GatewayConfig represents gateway configuration
type GatewayConfig = dolphingrpc.GatewayConfig

// DefaultGatewayConfig returns default gateway configuration
func DefaultGatewayConfig() *GatewayConfig {
	return dolphingrpc.DefaultGatewayConfig()
}

// HandlerRegistration represents a handler registration function
type HandlerRegistration = dolphingrpc.HandlerRegistration
