package main

import (
	"log"
	"time"

	"github.com/ArthurHlt/gohc"
)

func main() {
	conf := &gohc.UdpOpt{
		// if nil will send "test-gohc" by default
		Send: nil,
		// we not set receive so it will use ping check and port unreachable icmp response
		// it will root privileges to run this example
		Receive:     nil,
		Timeout:     5 * time.Second,
		PingTimeout: 5 * time.Second,
		Delay:       1 * time.Second,
		// you can set an alternative port for host which will override any port set in the host during check
		AltPort: 0,
	}
	hc := gohc.NewUdpHealthCheck(conf)
	err := hc.Check("localhost:53")
	if err != nil {
		log.Fatal(err)
	}

	// more stronger version but it requires udp server to reply something
	conf = &gohc.UdpOpt{
		// if nil will send "test-gohc" by default
		Send: nil,
		Receive: []*gohc.Payload{
			{
				Text: "received",
			},
		},
		Timeout: 5 * time.Second,
		// you can set an alternative port for host which will override any port set in the host during check
		AltPort: 0,
	}
	hc = gohc.NewUdpHealthCheck(conf)
	err = hc.Check("localhost:53")
	if err != nil {
		log.Fatal(err)
	}

}
