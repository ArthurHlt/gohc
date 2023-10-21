package gohc_test

import (
	"github.com/ArthurHlt/gohc/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sync/atomic"
	"time"

	. "github.com/ArthurHlt/gohc"
)

var _ = Describe("Udp", func() {
	Context("Check", func() {
		var udpServer *testhelpers.UdpServer
		BeforeEach(func() {
			var err error
			udpServer, err = testhelpers.NewUdpServer()
			Expect(err).To(BeNil())
			go func() {
				defer GinkgoRecover()
				err := udpServer.Run()
				Expect(err).To(BeNil())
			}()
		})
		AfterEach(func() {
			udpServer.Close()
		})
		When("Receive payload", func() {
			It("should return nil when response match", func() {
				var ops int64
				udpServer.SetHandler(func(b []byte) {
					Expect(string(b)).To(Equal(DefaultUdpSend))
					atomic.AddInt64(&ops, 1)
				})
				udpServer.SetResponse([]byte("received"))

				hc := NewUdpHealthCheck(&UdpOpt{
					Receive: []*Payload{
						{
							Text: "received",
						},
					},
				})

				err := hc.Check(udpServer.Addr())
				Expect(err).To(BeNil())
				testhelpers.EventuallyAtomic(&ops).Should(Equal(1))
			})
			It("should return error when response does not match", func() {
				var ops int64
				udpServer.SetHandler(func(b []byte) {
					Expect(string(b)).To(Equal(DefaultUdpSend))
					atomic.AddInt64(&ops, 1)
				})
				udpServer.SetResponse([]byte("not-match"))

				hc := NewUdpHealthCheck(&UdpOpt{
					Receive: []*Payload{
						{
							Text: "received",
						},
					},
				})

				err := hc.Check(udpServer.Addr())
				Expect(err).To(HaveOccurred())
				testhelpers.EventuallyAtomic(&ops).Should(Equal(1))
			})
		})
		When("No receive payload", func() {
			When("ping succeed and no port unreachable received", func() {
				It("should return nil", func() {
					var ops int64
					udpServer.SetHandler(func(b []byte) {
						Expect(string(b)).To(Equal(DefaultUdpSend))
						atomic.AddInt64(&ops, 1)
					})

					hc := NewUdpHealthCheck(&UdpOpt{
						Timeout: 1 * time.Second,
					})

					err := hc.Check(udpServer.Addr())
					Expect(err).To(BeNil())
					testhelpers.EventuallyAtomic(&ops).Should(Equal(1))
				})
			})
			When("ping failed", func() {
				It("should return error", func() {
					hc := NewUdpHealthCheck(&UdpOpt{
						Timeout:     1 * time.Second,
						PingTimeout: 1 * time.Second,
					})

					err := hc.Check("127.0.0.10")
					Expect(err).To(HaveOccurred())
				})
			})
			When("port unreachable received", func() {
				It("should return error", func() {
					addr := udpServer.Addr()
					udpServer.Close()
					hc := NewUdpHealthCheck(&UdpOpt{
						Timeout: 1 * time.Second,
					})

					err := hc.Check(addr)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("port unreachable"))
				})
			})
		})
	})
})
