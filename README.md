# go-gqrx

A library that provides a client to connect to [GQRX](https://github.com/gqrx-sdr/gqrx)

## use

This example starts the DSP and tunes to 93.7FM

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
	mode, bandwidth, err := client.GetDemod()
	if err != nil {
		log.Fatalf("Error getting current mode: %v", err)
	}
	fmt.Printf("Mode: %s\nBandwidth: %d\n", mode, bandwidth)
	// get current freq
	freq, err := client.GetFreq()
	if err != nil {
		log.Fatalf("Error getting frequency: %v", err)
	}
	fmt.Printf("Freq: %d\n", freq)
	// get mute status
	muted, err := client.GetMute()
	if err != nil {
		log.Fatalf("Error getting mute: %v", err)
	}
	// if muted
	if muted {
		// unmute audio
		if err := client.SetMute(false); err != nil {
			log.Fatalf("Error setting mute: %v", err)
		}
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
	// get SQL
	sql, err := client.GetSql()
	if err != nil {
		log.Fatalf("Error getting sql: %v", err)
	}
	fmt.Printf("SQL: %.2f\n", sql)
	// get signal strength
	strength, err := client.GetSigStrength()
	if err != nil {
		return
	}
	fmt.Printf("Strength: %.2f\n", strength)
	// set SQL
	if err := client.SetSql(-150); err != nil {
		log.Fatalf("Error setting sql: %v", err)
	}
	// disconnect
	if err := client.Disconnect(); err != nil {
		log.Fatalf("Error disconnecting: %v", err)
	}
}
```

## Resources

[GQRX remote API](https://github.com/gqrx-sdr/gqrx/blob/master/resources/remote-control.txt)