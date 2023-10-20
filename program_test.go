package gohc_test

import (
	"crypto/tls"
	"crypto/x509"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"

	. "github.com/ArthurHlt/gohc"
)

var _ = Describe("Program", func() {
	Context("Check", func() {
		It("should execute program on the most basic test", func() {
			hc := NewProgramHealthCheck(&ProgramConfig{
				Path: "bash",
				Args: []string{"-c", "cat - && exit 1"},
			}, 5*time.Second, false, nil)

			err := hc.Check("127.0.0.1:8080")
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("127.0.0.1:8080"))
			Expect(err.Error()).To(ContainSubstring(`"timeout_seconds":5,`))
		})
		When("set an alternative port", func() {
			It("should receive host with alternative port", func() {
				hc := NewProgramHealthCheck(&ProgramConfig{
					Path: "bash",
					Args: []string{"-c", "cat - && exit 1"},
				}, 5*time.Second, false, nil)

				hc.SetAltPort(8090)
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

				hc := NewProgramHealthCheck(&ProgramConfig{
					Path: "bash",
					Args: []string{"-c", "cat - && exit 1"},
				}, 5*time.Second, false, &tls.Config{
					InsecureSkipVerify: true,
					ServerName:         "localhost",
					RootCAs:            cp,
				})

				err := hc.Check("127.0.0.1:8080")
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring(`"insecure_skip_verify":true,`))
				Expect(err.Error()).To(ContainSubstring(`"server_name":"localhost"`))
			})
		})
	})
})
