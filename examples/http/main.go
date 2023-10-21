package main

import (
	"log"
	"net/http"
	"time"

	"github.com/ArthurHlt/gohc"
)

func main() {
	conf := &gohc.HttpOpt{
		Host: "",
		Path: "/",
		Send: &gohc.Payload{
			Text:   "test",
			Binary: nil,
		},
		Receive: &gohc.Payload{
			Text:   "",
			Binary: []byte("test"),
		},
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		ExpectedStatuses: &gohc.IntRange{
			Start: 200,
			End:   300, // end is exclusive
		},
		CodecClientType: gohc.CodecClientType_HTTP1,
		Method:          http.MethodGet,
		Timeout:         5 * time.Second,
		TlsEnabled:      false,
		TlsConfig:       nil,
		// you can set an alternative port for host which will override any port set in the host during check
		AltPort: 0,
	}
	hc := gohc.NewHttpHealthCheck(conf)

	err := hc.Check("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
