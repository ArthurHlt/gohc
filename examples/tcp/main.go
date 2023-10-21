package main

import (
	"log"
	"time"

	"github.com/ArthurHlt/gohc"
)

func main() {
	conf := &gohc.TcpOpt{
		Send: &gohc.Payload{
			Text:   "test",
			Binary: nil,
		},
		Receive: []*gohc.Payload{
			{
				Text:   "",
				Binary: []byte("test"),
			},
		},
		Timeout:    5 * time.Second,
		TlsEnabled: false,
		TlsConfig:  nil,
		// you can set an alternative port for host which will override any port set in the host during check
		AltPort: 0,
	}
	hc := gohc.NewTcpHealthCheck(conf)

	err := hc.Check("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
