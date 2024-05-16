package gohc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type ProgramTlsOpt struct {
	// Set to true to not verify the server's certificate. This is strongly discouraged.
	InsecureSkipVerify bool
	// ServerName is used to verify the hostname on the returned
	ServerName string
	// Trusted CA certificates for verifying the server certificate. It must be in pem format
	RootCAs []string
}
type ProgramOpt struct {
	// The path to the executable.
	Path string
	// The arguments to pass to the executable.
	Args []string
	// Options to pass to the executable.
	Options map[string]any
	// Timeout program finish. If left empty (default to 5s)
	Timeout time.Duration
	// TlsEnabled set to true if the gRPC health check request should be sent over TLS.
	TlsEnabled bool
	// Tls configuration for program
	ProgramTlsConfig *ProgramTlsOpt
	// AltPort specifies the port to use for gRPC health check requests.
	// If left empty it taks the port from host during check.
	AltPort uint32
}

type ProgramHealthCheck struct {
	opt     *ProgramOpt
	dataTls *programDataTls
}

type programData struct {
	Host           string          `json:"host"`
	TimeoutSeconds int64           `json:"timeout_seconds"`
	Options        map[string]any  `json:"options"`
	TlsEnabled     bool            `json:"tls_enabled"`
	TlsConfig      *programDataTls `json:"tls_config"`
}

type programDataTls struct {
	InsecureSkipVerify bool     `json:"insecure_skip_verify"`
	ServerName         string   `json:"server_name"`
	RootCAs            []string `json:"root_cas"`
}

func NewProgramHealthCheck(opt *ProgramOpt) *ProgramHealthCheck {
	return &ProgramHealthCheck{
		opt:     opt,
		dataTls: tlsConfToData(opt.ProgramTlsConfig),
	}
}

func (h *ProgramHealthCheck) Check(host string) error {
	host, err := FormatHost(host, h.opt.AltPort)
	if err != nil {
		return err
	}

	timeout := h.opt.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	content := &programData{
		Host:           host,
		TimeoutSeconds: int64(timeout.Seconds()),
		Options:        h.opt.Options,
		TlsEnabled:     h.opt.TlsEnabled,
		TlsConfig:      h.dataTls,
	}
	dataJson, err := json.Marshal(content)
	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()

	cmd := exec.CommandContext(ctx, h.opt.Path, h.opt.Args...)
	output := &bytes.Buffer{}
	input := bytes.NewBuffer(dataJson)

	cmd.Stderr = output
	cmd.Stdout = output
	cmd.Stdin = input

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("program health check failed: %s, output from program: %s",
			err.Error(),
			output.String(),
		)
	}

	return nil
}

func (h *ProgramHealthCheck) String() string {
	return "ProgramHealthCheck"
}

func tlsConfToData(tlsConf *ProgramTlsOpt) *programDataTls {
	if tlsConf == nil {
		return &programDataTls{}
	}
	return &programDataTls{
		InsecureSkipVerify: tlsConf.InsecureSkipVerify,
		ServerName:         tlsConf.ServerName,
		RootCAs:            tlsConf.RootCAs,
	}
}
