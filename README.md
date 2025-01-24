# go-gqrx

A library that provides a client to connect to [GQRX](https://github.com/gqrx-sdr/gqrx)

## use

This example starts the DSP and tunes to 92.7FM

```go
package main

import (
	"fmt"
	"github.com/CodeGoy/go-gqrx"
	"log"
)

func main() {
	var station int64 = 93700000 // 93.7 FM
	client := gqrx.NewClient("127.0.0.1:7356")
	if err := client.Connect(); err != nil {
		log.Fatalf("Error connecting: %v", err)
		return
	}
	// get current mode and bandwidth
	currentMode, currentBandwidth, err := client.GetDemod()
	if err != nil {
		log.Fatalf("Error getting current mode: %v", err)
	}
	// get current freq
	currentFreq, err := client.GetFreq()
	if err != nil {
		log.Fatalf("Error getting frequency: %v", err)
	}
	fmt.Printf("Mode: %s\nBandwidth: %d\nFreq: %d\n", currentMode, currentBandwidth, currentFreq)
	// mute audio
	if err := client.SetMute(false); err != nil {
		log.Fatalf("Error setting mute: %v", err)
	}
	// True on DPS if not running
	dspStatus, err := client.GetDspStatus()
	if err != nil {
		log.Fatalf("Error getting DSP status: %v", err)
	}
	if !dspStatus {
		err := client.SetDspStatus(true)
		if err != nil {
			log.Fatalf("Error setting DSP status: %v", err)
		}
	}
	// set mode and bandwidth, FM RADIO
	if err := client.SetDemod("WFM_ST", 160000); err != nil {
		log.Fatalf("Error setting demod: %v", err)
	}
	// set freq
	if err := client.SetFreq(station); err != nil {
		log.Fatalf("Error setting Freq: %v", err)
	}
	// disconnect
	if err := client.Disconnect(); err != nil {
		log.Fatalf("Error disconnecting: %v", err)
	}
}
```