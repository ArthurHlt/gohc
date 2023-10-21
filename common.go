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

// IntRange Specifies the int64 start and end of the range using half-open interval semantics [start,
// end).
type IntRange struct {
	// start of the range (inclusive)
	Start int64
	// end of the range (exclusive)
	End int64
}

// Payload Describes the encoding of the payload bytes in the payload.
// It can be either text or binary.
type Payload struct {
	// Text payload
	Text string
	// Binary payload.
	Binary []byte
}

func (pc *Payload) GetData() []byte {
	if pc == nil {
		return nil
	}
	if len(pc.Binary) > 0 {
		return pc.Binary
	}
	return []byte(pc.Text)
}
