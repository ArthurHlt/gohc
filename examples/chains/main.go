package main

import (
	"log"
	"time"

	"github.com/ArthurHlt/gohc"
)

func main() {
	hcIcmp := gohc.NewIcmpHealthCheck(&gohc.IcmpOpt{
		Timeout: 5 * time.Second,
		Delay:   1 * time.Second,
	})

	hcTcp := gohc.NewTcpHealthCheck(&gohc.TcpOpt{
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
	})

	// you can add multiple health check
	// set inParallel to false to check them in serial
	// set requireAll to false to make only one succeed to consider as success
	hc := gohc.NewChains(true, true, hcIcmp, hcTcp)
	err := hc.Check("localhost")
	if err != nil {
		log.Fatal(err)
	}
}
