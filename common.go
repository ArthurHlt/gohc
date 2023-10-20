package gohc

import (
	"fmt"
	"net"
)

func FormatHost(host string, altPort uint32) (string, error) {
	splitHost, port, err := net.SplitHostPort(host)
	if err != nil {
		return "", fmt.Errorf("fail to split host and port: %w", err)
	}
	if altPort > 0 {
		port = fmt.Sprintf("%d", altPort)
	}
	return net.JoinHostPort(splitHost, port), nil
}
