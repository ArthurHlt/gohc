package main

import (
	"log"
	"time"

	"github.com/ArthurHlt/gohc"
)

func main() {
	conf := &gohc.GrpcConfig{
		ServiceName: "my-service-can-be-empty",
		Authority:   "",
	}
	hc := gohc.NewGrpcHealthCheck(conf, 5*time.Second, false, nil)

	// you can set an alternative port for host which will override any port set in the host during check
	//hc.SetAltPort(8090)

	err := hc.Check("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
