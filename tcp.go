package gohc

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"
)

// TcpOpt Describes the TCP health check specific options.
type TcpOpt struct {
	// TCP specific payload.
	// Empty payloads imply a connect-only health check.
	Send *Payload
	// When checking the response, “fuzzy” matching is performed such that each
	// binary block must be found, and in the order specified, but not
	// necessarily contiguous.
	Receive []*Payload
	// Timeout for connection and for each receive data. If left empty (default to 5s)
	Timeout time.Duration
	// TlsEnabled set to true if the gRPC health check request should be sent over TLS.
	TlsEnabled bool
	// TlsConfig specifies the TLS configuration to use for TLS enabled gRPC health check requests.
	TlsConfig *tls.Config
	// AltPort specifies the port to use for gRPC health check requests.
	// If left empty it taks the port from host during check.
	AltPort uint32
}

type TcpHealthCheck struct {
	opt *TcpOpt
}

func NewTcpHealthCheck(opt *TcpOpt) *TcpHealthCheck {
	return &TcpHealthCheck{
		opt: opt,
	}
}

func (h *TcpHealthCheck) Check(host string) error {
	netConn, err := h.makeNetConn(host)
	if err != nil {
		return err
	}
	defer netConn.Close()

	timeout := h.opt.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	if h.opt.Send != nil {
		_, err = netConn.Write(h.opt.Send.GetData())
		if err != nil {
			return err
		}
	}

	for _, toReceive := range h.opt.Receive {
		err := netConn.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			return err
		}
		buf := make([]byte, len(toReceive.GetData()))
		n, err := io.ReadFull(netConn, buf)
		if err != nil {
			return fmt.Errorf("failed to read %d bytes: %v", len(toReceive.GetData()), err)
		}
		got := buf[0:n]
		if string(got) != string(toReceive.GetData()) {
			return fmt.Errorf("expected %s, got %s", string(toReceive.GetData()), string(got))
		}
	}
	return nil
}

func (h *TcpHealthCheck) makeNetConn(host string) (net.Conn, error) {
	var err error

	timeout := h.opt.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	host, err = FormatHost(host, h.opt.AltPort)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	if h.opt.TlsEnabled {
		return tls.DialWithDialer(dialer, "tcp", host, h.opt.TlsConfig)
	}
	return dialer.Dial("tcp", host)
}

func (h *TcpHealthCheck) String() string {
	return "TcpHealthCheck"
}
