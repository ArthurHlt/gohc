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

type HttpHealthCheck struct {
	httpClient *http.Client
	httpHcConf *HttpConfig
	tlsEnabled bool
	altPort    uint32
}

func NewHttpHealthCheck(httpHcConf *HttpConfig, timeout time.Duration, tlsEnabled bool, tlsConf *tls.Config) *HttpHealthCheck {
	return &HttpHealthCheck{
		httpClient: makeHttpClient(httpHcConf.CodecClientType, tlsConf, timeout),
		httpHcConf: httpHcConf,
		tlsEnabled: tlsEnabled,
	}
}

func (h *HttpHealthCheck) SetAltPort(altPort uint32) {
	h.altPort = altPort
}

func (h *HttpHealthCheck) Check(host string) error {
	var err error
	host, err = FormatHost(host, h.altPort)
	if err != nil {
		return err
	}
	protocol := "http"
	if h.tlsEnabled {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s%s", protocol, host, h.httpHcConf.Path)

	method := http.MethodGet
	if h.httpHcConf.Method != "" {
		method = strings.ToUpper(h.httpHcConf.Method)
	}
	var body io.Reader
	if h.httpHcConf.Send != nil {
		body = bytes.NewReader(h.httpHcConf.Send.GetData())
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if h.httpHcConf.Host != "" {
		req.Host = h.httpHcConf.Host
	}
	if h.httpHcConf.Headers != nil {
		req.Header = h.httpHcConf.Headers
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	start := int64(200)
	end := int64(201)
	if h.httpHcConf.ExpectedStatuses != nil {
		start = h.httpHcConf.ExpectedStatuses.Start
		end = h.httpHcConf.ExpectedStatuses.End
	}
	statusCode := int64(resp.StatusCode)
	if statusCode < start || statusCode >= end {
		return fmt.Errorf("unexpected status code, got %d not in range [%d, %d)", statusCode, start, end)
	}
	if h.httpHcConf.Receive != nil {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}
		if !bytes.Contains(b, h.httpHcConf.Receive.GetData()) {
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
	httpClient := &http.Client{
		Transport: roundTripper,
		Timeout:   timeout,
	}
	return httpClient
}
