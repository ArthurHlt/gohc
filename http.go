package gohc

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go/http3"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
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

// HttpOpt Describes the health check policy for a given endpoint.
type HttpOpt struct {
	// The value of the host header in the HTTP health check request
	Host string
	// Specifies the HTTP path that will be requested during health checking. For example
	// */healthcheck*.
	Path string
	// HTTP specific payload.
	Send *Payload
	// HTTP specific response.
	Receive *Payload
	// Specifies a list of HTTP headers that should be added to each request that is sent to the
	// health checked cluster.
	Headers http.Header
	// Specifies a list of HTTP response statuses considered healthy. If provided, replaces default
	// 200-only policy - 200 must be included explicitly as needed. Ranges follow half-open
	// semantics of Int64Range. The start and end of each
	ExpectedStatuses *IntRange
	// Use specified application protocol for health checks.
	CodecClientType CodecClientType
	// HTTP Method that will be used for health checking, default is "GET".
	// If a non-200 response is expected by the method, it needs to be set in expected_statuses.
	Method string
	// Timeout for the http health check request. If left empty (default to 5s)
	Timeout time.Duration
	// TlsEnabled set to true if the gRPC health check request should be sent over TLS.
	TlsEnabled bool
	// TlsConfig specifies the TLS configuration to use for TLS enabled gRPC health check requests.
	TlsConfig *tls.Config
	// AltPort specifies the port to use for gRPC health check requests.
	// If left empty it taks the port from host during check.
	AltPort uint32
}

type HttpHealthCheck struct {
	httpClient *http.Client
	opt        *HttpOpt
}

func NewHttpHealthCheck(opt *HttpOpt) *HttpHealthCheck {
	return &HttpHealthCheck{
		httpClient: makeHttpClient(opt.CodecClientType, opt.TlsConfig, opt.Timeout),
		opt:        opt,
	}
}

func (h *HttpHealthCheck) Check(host string) error {
	var err error
	host, err = FormatHost(host, h.opt.AltPort)
	if err != nil {
		return err
	}
	protocol := "http"
	if h.opt.TlsEnabled {
		protocol = "https"
	}

	path := h.opt.Path
	if path == "" {
		path = "/"
	}

	url := fmt.Sprintf("%s://%s%s", protocol, host, path)

	method := http.MethodGet
	if h.opt.Method != "" {
		method = strings.ToUpper(h.opt.Method)
	}
	var body io.Reader
	if h.opt.Send != nil {
		body = bytes.NewReader(h.opt.Send.GetData())
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if h.opt.Host != "" {
		req.Host = h.opt.Host
	}
	if h.opt.Headers != nil {
		req.Header = h.opt.Headers
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	start := int64(200)
	end := int64(201)
	if h.opt.ExpectedStatuses != nil {
		start = h.opt.ExpectedStatuses.Start
		end = h.opt.ExpectedStatuses.End
	}
	statusCode := int64(resp.StatusCode)
	if statusCode < start || statusCode >= end {
		return fmt.Errorf("unexpected status code, got %d not in range [%d, %d)", statusCode, start, end)
	}
	if h.opt.Receive != nil {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}
		if !bytes.Contains(b, h.opt.Receive.GetData()) {
			return fmt.Errorf("response body does not contains expected data")
		}
	}
	return nil
}

func makeHttpClient(codec CodecClientType, tlsConf *tls.Config, timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	var roundTripper http.RoundTripper
	switch codec {
	case CodecClientType_HTTP3:
		roundTripper = &http3.RoundTripper{
			TLSClientConfig: tlsConf,
		}
	default:
		roundTripper = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       tlsConf,
		}
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	httpClient := &http.Client{
		Transport: roundTripper,
		Timeout:   timeout,
	}
	return httpClient
}

func (h *HttpHealthCheck) String() string {
	return "HttpHealthCheck"
}
