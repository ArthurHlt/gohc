package main

import (
	"log"
	"net/http"
	"time"

	"github.com/ArthurHlt/gohc"
)

func main() {
	conf := &gohc.HttpConfig{
		Host: "",
		Path: "/",
		Send: &gohc.PayloadConfig{
			Text:   "test",
			Binary: nil,
		},
		Receive: &gohc.PayloadConfig{
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
	}
	hc := gohc.NewHttpHealthCheck(conf, 5*time.Second, false, nil)

	// you can set an alternative port for host which will override any port set in the host during check
	//hc.SetAltPort(8090)

	err := hc.Check("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
