package main

import (
	"log"
	"time"

	"github.com/ArthurHlt/gohc"
)

func main() {
	conf := &gohc.ProgramConfig{
		Path: "bash",
		Args: []string{"-c", "cat - && exit 0"},
	}
	hc := gohc.NewProgramHealthCheck(conf, 5*time.Second, false, nil)

	// you can set an alternative port for host which will override any port set in the host during check
	//hc.SetAltPort(8090)

	// This given payload will be sent to the program in stdin:
	// {
	//  "host": localhost:8080",
	//  "timeout_seconds": 5,
	//  "tls_enabled": false,
	//  "tls_config": {
	//    "insecure_skip_verify": false,
	//    "server_name": ""
	//  }
	//}

	err := hc.Check("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

}
