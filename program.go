package gohc

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type ProgramHealthCheck struct {
	hcConf     *ProgramConfig
	tlsEnabled bool
	timeout    time.Duration
	altPort    uint32
	dataTls    *programDataTls
}

type programData struct {
	Host           string          `json:"host"`
	TimeoutSeconds int64           `json:"timeout_seconds"`
	TlsEnabled     bool            `json:"tls_enabled"`
	TlsConfig      *programDataTls `json:"tls_config"`
}

type programDataTls struct {
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	ServerName         string `json:"server_name"`
}

func NewProgramHealthCheck(hcConf *ProgramConfig, timeout time.Duration, tlsEnabled bool, tlsConf *tls.Config) *ProgramHealthCheck {
	return &ProgramHealthCheck{
		hcConf:     hcConf,
		tlsEnabled: tlsEnabled,
		timeout:    timeout,
		dataTls:    tlsConfToData(tlsConf),
	}
}

func (h *ProgramHealthCheck) SetAltPort(altPort uint32) {
	h.altPort = altPort
}

func (h *ProgramHealthCheck) Check(host string) error {
	host, err := FormatHost(host, h.altPort)
	if err != nil {
		return err
	}

	content := &programData{
		Host:           host,
		TimeoutSeconds: int64(h.timeout.Seconds()),
		TlsEnabled:     h.tlsEnabled,
		TlsConfig:      h.dataTls,
	}
	dataJson, err := json.Marshal(content)
	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), h.timeout)
	defer cancelFunc()

	cmd := exec.CommandContext(ctx, h.hcConf.Path, h.hcConf.Args...)
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

func tlsConfToData(tlsConf *tls.Config) *programDataTls {
	if tlsConf == nil {
		return &programDataTls{}
	}
	return &programDataTls{
		InsecureSkipVerify: tlsConf.InsecureSkipVerify,
		ServerName:         tlsConf.ServerName,
	}
}
