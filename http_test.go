package gohc_test

import (
	"crypto/tls"
	. "github.com/ArthurHlt/gohc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

var _ = Describe("Http", func() {
	Context("Check", func() {
		Context("Basic", func() {
			var server *ghttp.Server
			BeforeEach(func() {
				server = ghttp.NewServer()
			})
			AfterEach(func() {
				//shut down the httpServer between tests
				server.Close()
			})
			It("should return nil on the most basic test on path / and 200 status code", func() {
				server.AppendHandlers(ghttp.RespondWith(200, "OK"))

				hc := NewHttpHealthCheck(&HttpOpt{})

				err := hc.Check(urlToHost(server.URL()))
				Expect(err).To(BeNil())
			})
			It("should return an error when status code is not expected", func() {
				server.AppendHandlers(ghttp.RespondWith(404, "NOT FOUND"))

				hc := NewHttpHealthCheck(&HttpOpt{})

				err := hc.Check(urlToHost(server.URL()))
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("404"))
			})
			It("should return an error when take more time than timeout", func() {
				server.AppendHandlers(func(w http.ResponseWriter, req *http.Request) {
					time.Sleep(10 * time.Millisecond)
				})

				hc := NewHttpHealthCheck(&HttpOpt{
					Timeout: 1 * time.Nanosecond,
				})

				err := hc.Check(urlToHost(server.URL()))
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("context deadline exceeded"))
			})
			It("should return nil when status code is in expected range", func() {
				server.AppendHandlers(ghttp.RespondWith(404, "NOT FOUND"))

				hc := NewHttpHealthCheck(&HttpOpt{
					ExpectedStatuses: &IntRange{
						Start: 400,
						End:   500,
					},
				})

				err := hc.Check(urlToHost(server.URL()))
				Expect(err).To(BeNil())
			})
			It("should append headers when user declare it", func() {
				server.AppendHandlers(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Header.Get("X-Test")).To(Equal("test"))
					Expect(req.Header.Values("X-Test-Append")).To(Equal([]string{"test1", "test2"}))
				})

				hc := NewHttpHealthCheck(&HttpOpt{
					Headers: map[string][]string{
						"X-Test":        {"test"},
						"X-Test-Append": {"test1", "test2"},
					},
				})

				err := hc.Check(urlToHost(server.URL()))
				Expect(err).To(BeNil())
				Expect(server.ReceivedRequests()).Should(HaveLen(1))
			})
			It("should set different method when user declare it", func() {
				server.AppendHandlers(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Method).To(Equal("POST"))
				})

				hc := NewHttpHealthCheck(&HttpOpt{
					Method: http.MethodPost,
				})

				err := hc.Check(urlToHost(server.URL()))
				Expect(err).To(BeNil())
				Expect(server.ReceivedRequests()).Should(HaveLen(1))
			})
			It("should set different host when user declare it", func() {
				server.AppendHandlers(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Host).To(Equal("myhost"))
				})

				hc := NewHttpHealthCheck(&HttpOpt{
					Host: "myhost",
				})

				err := hc.Check(urlToHost(server.URL()))
				Expect(err).To(BeNil())
				Expect(server.ReceivedRequests()).Should(HaveLen(1))
			})
			It("should use alt port when given", func() {
				server.AppendHandlers(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.Host).To(Equal("myhost"))
				})

				host := urlToHost(server.URL())
				simpleHost, port, err := net.SplitHostPort(host)
				Expect(err).To(BeNil())
				portInt, err := strconv.Atoi(port)
				Expect(err).To(BeNil())

				hc := NewHttpHealthCheck(&HttpOpt{
					Host:    "myhost",
					AltPort: uint32(portInt),
				})

				err = hc.Check(simpleHost + ":1")
				Expect(err).To(BeNil())
				Expect(server.ReceivedRequests()).Should(HaveLen(1))
			})
			When("User set send/receive payload", func() {
				It("should return nil if received payload contains what user wanted", func() {
					server.AppendHandlers(ghttp.RespondWith(200, "long text contains ok here"))

					hc := NewHttpHealthCheck(&HttpOpt{
						Receive: &Payload{
							Text: "ok",
						},
					})

					err := hc.Check(urlToHost(server.URL()))
					Expect(err).To(BeNil())
				})
				It("should return an error if received payload doesn't contains what user wanted", func() {
					server.AppendHandlers(ghttp.RespondWith(200, "well not here"))

					hc := NewHttpHealthCheck(&HttpOpt{
						Receive: &Payload{
							Text: "ok",
						},
					})

					err := hc.Check(urlToHost(server.URL()))
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("not contains"))
				})
				It("should send payload given by user", func() {
					server.AppendHandlers(func(w http.ResponseWriter, req *http.Request) {
						body, err := io.ReadAll(req.Body)
						Expect(err).To(BeNil())
						Expect(string(body)).To(Equal("ok"))
					})

					hc := NewHttpHealthCheck(&HttpOpt{
						Send: &Payload{
							Text: "ok",
						},
					})

					err := hc.Check(urlToHost(server.URL()))
					Expect(err).To(BeNil())
					Expect(server.ReceivedRequests()).Should(HaveLen(1))
				})
			})
		})
	})

	Context("Https Server", func() {
		var server *ghttp.Server
		BeforeEach(func() {
			server = ghttp.NewTLSServer()
		})
		AfterEach(func() {
			//shut down the httpServer between tests
			server.Close()
		})
		It("should return an error if tls not enabled", func() {
			server.AppendHandlers(ghttp.RespondWith(200, "OK"))

			hc := NewHttpHealthCheck(&HttpOpt{})

			err := hc.Check(urlToHost(server.URL()))
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("400"))
		})
		It("should return nil on the most basic test on path / and 200 status code", func() {
			server.AppendHandlers(ghttp.RespondWith(200, "OK"))

			hc := NewHttpHealthCheck(&HttpOpt{
				TlsEnabled: true,
				TlsConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			})

			err := hc.Check(urlToHost(server.URL()))
			Expect(err).To(BeNil())
		})
	})

})
