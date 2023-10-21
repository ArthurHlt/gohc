package gohc_test

import (
	"crypto/x509"
	. "github.com/ArthurHlt/gohc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Program", func() {
	Context("Check", func() {
		It("should execute program on the most basic test", func() {
			hc := NewProgramHealthCheck(&ProgramOpt{
				Path: "bash",
				Args: []string{"-c", "cat - && exit 1"},
				Options: map[string]any{
					"test": "test",
				},
			})

			err := hc.Check("127.0.0.1:8080")
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("127.0.0.1:8080"))
			Expect(err.Error()).To(ContainSubstring(`"timeout_seconds":5,`))
			Expect(err.Error()).To(ContainSubstring(`"test":"test"`))
		})
		When("set an alternative port", func() {
			It("should receive host with alternative port", func() {
				hc := NewProgramHealthCheck(&ProgramOpt{
					Path:    "bash",
					Args:    []string{"-c", "cat - && exit 1"},
					AltPort: 8090,
				})
				err := hc.Check("127.0.0.1:8080")
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("127.0.0.1:8090"))
				Expect(err.Error()).To(ContainSubstring(`"timeout_seconds":5,`))
			})
		})
		When("set tls config", func() {
			It("should pass tls config to program", func() {

				cp := x509.NewCertPool()
				cp.AppendCertsFromPEM(LocalhostCert)

				hc := NewProgramHealthCheck(&ProgramOpt{
					Path:       "bash",
					Args:       []string{"-c", "cat - && exit 1"},
					TlsEnabled: true,
					ProgramTlsConfig: &ProgramTlsOpt{
						InsecureSkipVerify: true,
						ServerName:         "localhost",
						RootCAs:            []string{string(LocalhostCert)},
					},
				})

				err := hc.Check("127.0.0.1:8080")
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring(`"insecure_skip_verify":true,`))
				Expect(err.Error()).To(ContainSubstring(`"server_name":"localhost"`))
				Expect(err.Error()).To(ContainSubstring(`"root_cas":["-----BEGIN CERTIFICATE-----\nMIIDOTCCAiGgAwIBAgIQSRJrEpBGFc7tNb1fb5pKFzANBgkqhkiG9w0BAQsFADAS`))
			})
		})
	})
})
