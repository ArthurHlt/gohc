package gohc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"time"

	. "github.com/ArthurHlt/gohc"
)

var _ = Describe("Icmp", func() {
	Context("Check", func() {
		BeforeEach(func() {
			if os.Getenv("GITHUB_ACTION") != "" {
				Skip("Skip icmp test on github action")
			}
		})
		It("should return nil on the most basic test when ping succeed", func() {
			hc := NewIcmpHealthCheck(&IcmpOpt{})

			err := hc.Check("127.0.0.1")
			Expect(err).ToNot(HaveOccurred())
		})
		When("using hostname instead of ip", func() {
			It("should resolve hostname and ping first ip", func() {
				hc := NewIcmpHealthCheck(&IcmpOpt{})

				err := hc.Check("localhost")
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("ping reach timeout", func() {
			It("should raise an error because ping failed", func() {
				hc := NewIcmpHealthCheck(&IcmpOpt{
					Timeout: 200 * time.Millisecond,
				})

				err := hc.Check("127.0.0.10")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("timeout"))
			})
		})
	})
})
