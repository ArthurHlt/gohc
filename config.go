package gohc

import (
	"errors"
	"net/http"
	"strings"
)

type CodecClientType int32

const (
	CodecClientType_HTTP1 CodecClientType = 0
	CodecClientType_HTTP2 CodecClientType = 1
	// [#not-implemented-hide:] QUIC implementation is not production ready yet. Use this enum with
	// caution to prevent accidental execution of QUIC code. I.e. `!= HTTP2` is no longer sufficient
	// to distinguish HTTP1 and HTTP2 traffic.
	CodecClientType_HTTP3 CodecClientType = 2
)

// Enum value maps for CodecClientType.
var (
	CodecClientType_name = map[int32]string{
		0: "HTTP1",
		1: "HTTP2",
		2: "HTTP3",
	}
	CodecClientType_value = map[string]int32{
		"HTTP1": 0,
		"HTTP2": 1,
		"HTTP3": 2,
	}
)

// IntRange Specifies the int64 start and end of the range using half-open interval semantics [start,
// end).
type IntRange struct {
	// start of the range (inclusive)
	Start int64
	// end of the range (exclusive)
	End int64
}

// HttpConfig Describes the health check policy for a given endpoint.
type HttpConfig struct {
	// The value of the host header in the HTTP health check request
	Host string
	// Specifies the HTTP path that will be requested during health checking. For example
	// */healthcheck*.
	Path string
	// HTTP specific payload.
	Send *PayloadConfig
	// HTTP specific response.
	Receive *PayloadConfig
	// Specifies a list of HTTP headers that should be added to each request that is sent to the
	// health checked cluster.
	Headers http.Header
	// Specifies a list of HTTP response statuses considered healthy. If provided, replaces default
	// 200-only policy - 200 must be included explicitly as needed. Ranges follow half-open
	// semantics of Int64Range. The start and end of each
	// range are required. Only statuses in the range [100, 600) are allowed.
	ExpectedStatuses *IntRange
	// Use specified application protocol for health checks.
	CodecClientType CodecClientType
	// HTTP Method that will be used for health checking, default is "GET".
	// GET, HEAD, POST, PUT, DELETE, OPTIONS, TRACE, PATCH methods are supported, but making request body is not supported.
	// CONNECT method is disallowed because it is not appropriate for health check request.
	// If a non-200 response is expected by the method, it needs to be set in expected_statuses.
	Method string
}

func (h *HttpConfig) Check() error {
	if h.Path == "" {
		return errors.New("path is empty")
	}
	switch strings.ToUpper(h.Method) {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodTrace:
	default:
		return errors.New("method is not supported")
	}
	return nil
}

// PayloadConfig Describes the encoding of the payload bytes in the payload.
// It can be either text or binary.
type PayloadConfig struct {
	// Text payload
	Text string
	// Binary payload.
	Binary []byte
}

func (pc *PayloadConfig) Check() error {
	if pc.Text == "" && pc.Binary == nil {
		return errors.New("payload config is empty")
	}
	return nil
}

func (pc *PayloadConfig) GetData() []byte {
	if pc == nil {
		return nil
	}
	if len(pc.Binary) > 0 {
		return pc.Binary
	}
	return []byte(pc.Text)
}

// TcpConfig Describes the TCP health check specific options.
type TcpConfig struct {
	// TCP specific payload.
	// Empty payloads imply a connect-only health check.
	Send *PayloadConfig
	// When checking the response, “fuzzy” matching is performed such that each
	// binary block must be found, and in the order specified, but not
	// necessarily contiguous.
	Receive []*PayloadConfig
}

// GrpcConfig Describes the gRPC health check specific options.
type GrpcConfig struct {
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
}

type ProgramConfig struct {
	// The path to the executable.
	Path string
	// The arguments to pass to the executable.
	Args []string
}
