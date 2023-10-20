package main

import (
	"log"
	"time"

	"github.com/ArthurHlt/gohc"
)

func main() {
	conf := &gohc.TcpConfig{
		Send: &gohc.PayloadConfig{
			Text:   "test",
			Binary: nil,
		},
		Receive: []*gohc.PayloadConfig{
			{
				Text:   "",
				Binary: []byte("test"),
			},
		},
	}
	hc := gohc.NewTcpHealthCheck(conf, 5*time.Second, false, nil)

	// you can set an alternative port for host which will override any port set in the host during check
	//hc.SetAltPort(8090)

	err := hc.Check("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
