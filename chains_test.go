package gohc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net"

	"github.com/ArthurHlt/gohc"
)

var _ = FDescribe("Chains", func() {
	var lis net.Listener
	var tcpHc *gohc.TcpHealthCheck
	var icmpHc *gohc.IcmpHealthCheck
	Context("Check", func() {
		BeforeEach(func() {
			var err error
			lis, err = net.Listen("tcp4", "127.0.0.1:0")
			Expect(err).To(BeNil())
			tcpHc = gohc.NewTcpHealthCheck(&gohc.TcpOpt{})
			icmpHc = gohc.NewIcmpHealthCheck(&gohc.IcmpOpt{})
		})
		AfterEach(func() {
			lis.Close()
		})
		Context("Serial", func() {
			When("Without requiring all check passing", func() {
				It("should return nil when none fail", func() {
					hc := gohc.NewChains(false, false, icmpHc, tcpHc)

					err := hc.Check(lis.Addr().String())

					Expect(err).To(BeNil())
				})
				It("should return nil when one fail", func() {
					hc := gohc.NewChains(false, false, icmpHc, tcpHc)
					lis.Close()
					err := hc.Check(lis.Addr().String())

					Expect(err).To(BeNil())
				})
				It("should return errors when all fail", func() {
					hc := gohc.NewChains(false, false, icmpHc, tcpHc)
					lis.Close()
					err := hc.Check("172.0.0.0:60")

					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("TcpHealthCheck"))
					Expect(err.Error()).To(ContainSubstring("IcmpHealthCheck"))
				})
			})
			When("With requiring all check passing", func() {
				It("should return nil when none fail", func() {
					hc := gohc.NewChains(false, true, icmpHc, tcpHc)

					err := hc.Check(lis.Addr().String())

					Expect(err).To(BeNil())
				})
				It("should return an error when one fail", func() {
					hc := gohc.NewChains(false, true, icmpHc, tcpHc)
					lis.Close()
					err := hc.Check(lis.Addr().String())

					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("TcpHealthCheck"))
				})
				It("should return first error when all fail", func() {
					hc := gohc.NewChains(false, true, tcpHc, icmpHc)
					lis.Close()
					err := hc.Check("172.0.0.0:60")

					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("TcpHealthCheck"))
					Expect(err.Error()).ToNot(ContainSubstring("IcmpHealthCheck"))
				})
			})
		})
		Context("Parallel", func() {
			When("Without requiring all check passing", func() {
				It("should return nil when none fail", func() {
					hc := gohc.NewChains(true, false, icmpHc, tcpHc)

					err := hc.Check(lis.Addr().String())

					Expect(err).To(BeNil())
				})
				It("should return nil when one fail", func() {
					hc := gohc.NewChains(true, false, icmpHc, tcpHc)
					lis.Close()
					err := hc.Check(lis.Addr().String())

					Expect(err).To(BeNil())
				})
				It("should return errors when all fail", func() {
					hc := gohc.NewChains(true, false, icmpHc, tcpHc)
					lis.Close()
					err := hc.Check("172.0.0.0:60")

					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("TcpHealthCheck"))
					Expect(err.Error()).To(ContainSubstring("IcmpHealthCheck"))
				})
			})
			When("With requiring all check passing", func() {
				It("should return nil when none fail", func() {
					hc := gohc.NewChains(true, true, icmpHc, tcpHc)

					err := hc.Check(lis.Addr().String())

					Expect(err).To(BeNil())
				})
				It("should return an error when one fail", func() {
					hc := gohc.NewChains(true, true, icmpHc, tcpHc)
					lis.Close()
					err := hc.Check(lis.Addr().String())

					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("TcpHealthCheck"))
				})
				It("should return all errors when all fail", func() {
					hc := gohc.NewChains(true, true, tcpHc, icmpHc)
					lis.Close()
					err := hc.Check("172.0.0.0:60")

					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("TcpHealthCheck"))
					Expect(err.Error()).To(ContainSubstring("IcmpHealthCheck"))
				})
			})
		})
	})
})
