package main

import (
	"github.com/ArthurHlt/gohc"
	"log"
	"time"
)

func main() {
	conf := &gohc.GrpcOpt{
		ServiceName: "my-service-can-be-empty",
		Authority:   "",
		Timeout:     5 * time.Second,
		TlsEnabled:  false,
		TlsConfig:   nil,
		// you can set an alternative port for host which will override any port set in the host during check
		AltPort: 0,
	}
	hc := gohc.NewGrpcHealthCheck(conf)
	err := hc.Check("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
