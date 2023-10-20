package gohc

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"
)

type TcpHealthCheck struct {
	hcConf     *TcpConfig
	tlsEnabled bool
	timeout    time.Duration
	tlsConf    *tls.Config
	altPort    uint32
}

func NewTcpHealthCheck(hcConf *TcpConfig, timeout time.Duration, tlsEnabled bool, tlsConf *tls.Config) *TcpHealthCheck {
	return &TcpHealthCheck{
		hcConf:     hcConf,
		tlsEnabled: tlsEnabled,
		timeout:    timeout,
		tlsConf:    tlsConf,
	}
}

func (h *TcpHealthCheck) SetAltPort(altPort uint32) {
	h.altPort = altPort
}

func (h *TcpHealthCheck) Check(host string) error {
	netConn, err := h.makeNetConn(host)
	if err != nil {
		return err
	}
	defer netConn.Close()

	if h.hcConf.Send != nil {
		_, err = netConn.Write(h.hcConf.Send.GetData())
		if err != nil {
			return err
		}
	}

	for _, toReceive := range h.hcConf.Receive {
		err := netConn.SetReadDeadline(time.Now().Add(h.timeout))
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
	host, err = FormatHost(host, h.altPort)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{
		Timeout: h.timeout,
	}
	if h.tlsEnabled {
		return tls.DialWithDialer(dialer, "tcp", host, h.tlsConf)
	}
	return dialer.Dial("tcp", host)
}
