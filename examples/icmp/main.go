package main

import (
	"log"
	"time"

	"github.com/ArthurHlt/gohc"
)

func main() {
	conf := &gohc.IcmpOpt{
		Timeout: 5 * time.Second,
		Delay:   1 * time.Second,
	}
	hc := gohc.NewIcmpHealthCheck(conf)

	// on icmp port is useless and are ignored if set.
	err := hc.Check("localhost")
	if err != nil {
		log.Fatal(err)
	}
}
