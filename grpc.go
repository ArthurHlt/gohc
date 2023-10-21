package gohc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"time"
)

// GrpcOpt Describes the gRPC health check specific options.
type GrpcOpt struct {
	// An optional service name parameter which will be sent to gRPC service in
	// `grpc.health.v1.HealthCheckRequest
	// <https://github.com/grpc/grpc/blob/master/src/proto/grpc/health/v1/health.proto#L20>`_.
	// message. See `gRPC health-checking overview
	// <https://github.com/grpc/grpc/blob/master/doc/health-checking.md>`_ for more information.
	ServiceName string
	// The value of the :authority header in the gRPC health check request. If
	// left empty (default value) this will be host in check.
	// The authority header can be customized for a specific endpoint by setting
	// the HealthCheckConfig.hostname field.
	Authority string
	// Timeout for the gRPC health check request. If left empty (default to 5s)
	Timeout time.Duration
	// TlsEnabled set to true if the gRPC health check request should be sent over TLS.
	TlsEnabled bool
	// TlsConfig specifies the TLS configuration to use for TLS enabled gRPC health check requests.
	TlsConfig *tls.Config
	// AltPort specifies the port to use for gRPC health check requests.
	// If left empty it taks the port from host during check.
	AltPort uint32
}

type GrpcHealthCheck struct {
	opt *GrpcOpt
}

func NewGrpcHealthCheck(opt *GrpcOpt) *GrpcHealthCheck {
	return &GrpcHealthCheck{
		opt: opt,
	}
}

func (h *GrpcHealthCheck) Check(host string) error {
	conn, err := h.makeGrpcConn(host)
	if err != nil {
		return err
	}
	defer conn.Close()

	timeout := h.opt.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client := healthpb.NewHealthClient(conn)
	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{
		Service: h.opt.ServiceName,
	})
	if err != nil {
		if stat, ok := status.FromError(err); ok {
			switch stat.Code() {
			case codes.Unimplemented:
				return fmt.Errorf("gRPC server does not implement the health protocol: %w", err)
			case codes.DeadlineExceeded:
				return fmt.Errorf("gRPC health check timeout: %w", err)
			}
		}

		return fmt.Errorf("gRPC health check failed: %w", err)
	}

	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("received gRPC status code: %v", resp.Status)
	}
	return nil
}

func (h *GrpcHealthCheck) makeGrpcConn(host string) (*grpc.ClientConn, error) {
	var err error
	host, err = FormatHost(host, h.opt.AltPort)
	if err != nil {
		return nil, err
	}
	var opts []grpc.DialOption
	if !h.opt.TlsEnabled {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(h.opt.TlsConfig)))
	}

	timeout := h.opt.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, host, opts...)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("fail to connect to %s within %s: %w", host, timeout, err)
		}
		return nil, fmt.Errorf("fail to connect to %s: %w", host, err)
	}
	return conn, nil
}
