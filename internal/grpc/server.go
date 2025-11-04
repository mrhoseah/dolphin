package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"go.uber.org/zap"
)

// Server represents a gRPC server
type Server struct {
	server   *grpc.Server
	config   *ServerConfig
	logger   *zap.Logger
	services []ServiceRegistration
}

// ServerConfig represents gRPC server configuration
type ServerConfig struct {
	Port            int           `yaml:"port" json:"port"`
	EnableReflection bool         `yaml:"enable_reflection" json:"enable_reflection"`
	EnableTLS       bool          `yaml:"enable_tls" json:"enable_tls"`
	CertFile        string        `yaml:"cert_file" json:"cert_file"`
	KeyFile         string        `yaml:"key_file" json:"key_file"`
	MaxRecvMsgSize  int           `yaml:"max_recv_msg_size" json:"max_recv_msg_size"` // bytes
	MaxSendMsgSize  int           `yaml:"max_send_msg_size" json:"max_send_msg_size"` // bytes
	Keepalive       *KeepaliveConfig `yaml:"keepalive" json:"keepalive"`
}

// KeepaliveConfig represents keepalive configuration
type KeepaliveConfig struct {
	Time                time.Duration `yaml:"time" json:"time"`
	Timeout             time.Duration `yaml:"timeout" json:"timeout"`
	PermitWithoutStream bool          `yaml:"permit_without_stream" json:"permit_without_stream"`
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:             9090,
		EnableReflection: true,
		EnableTLS:        false,
		MaxRecvMsgSize:   4 * 1024 * 1024, // 4MB
		MaxSendMsgSize:   4 * 1024 * 1024, // 4MB
		Keepalive: &KeepaliveConfig{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		},
	}
}

// ServiceRegistration represents a service registration function
type ServiceRegistration func(server *grpc.Server)

// NewServer creates a new gRPC server
func NewServer(config *ServerConfig, logger *zap.Logger) (*Server, error) {
	if config == nil {
		config = DefaultServerConfig()
	}

	// Build server options
	var opts []grpc.ServerOption

	// TLS
	if config.EnableTLS {
		creds, err := credentials.NewServerTLSFromFile(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Message size limits
	if config.MaxRecvMsgSize > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(config.MaxRecvMsgSize))
	}
	if config.MaxSendMsgSize > 0 {
		opts = append(opts, grpc.MaxSendMsgSize(config.MaxSendMsgSize))
	}

	// Keepalive
	if config.Keepalive != nil {
		kaep := keepalive.EnforcementPolicy{
			MinTime:             config.Keepalive.Time,
			PermitWithoutStream: config.Keepalive.PermitWithoutStream,
		}
		opts = append(opts, grpc.KeepaliveEnforcementPolicy(kaep))

		kasp := keepalive.ServerParameters{
			Time:    config.Keepalive.Time,
			Timeout: config.Keepalive.Timeout,
		}
		opts = append(opts, grpc.KeepaliveParams(kasp))
	}

	// Create gRPC server
	grpcServer := grpc.NewServer(opts...)

	// Enable reflection
	if config.EnableReflection {
		reflection.Register(grpcServer)
	}

	return &Server{
		server:   grpcServer,
		config:   config,
		logger:   logger,
		services: []ServiceRegistration{},
	}, nil
}

// RegisterService registers a service with the server
func (s *Server) RegisterService(registration ServiceRegistration) {
	s.services = append(s.services, registration)
	registration(s.server)
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// Register all services
	for _, registration := range s.services {
		registration(s.server)
	}

	// Listen on port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.logger.Info("gRPC server starting",
		zap.Int("port", s.config.Port),
		zap.Bool("reflection", s.config.EnableReflection),
		zap.Bool("tls", s.config.EnableTLS))

	// Start serving
	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("gRPC server stopping")

	// Graceful stop with timeout
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	}
}

// GetServer returns the underlying gRPC server
func (s *Server) GetServer() *grpc.Server {
	return s.server
}

