package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/CodeGoy/go-gqrx"
	"log"
)

var (
	version = "v0.0.2"
)

type Bookmarker struct {
	app    fyne.App
	client *gqrx.Client
}

type SdrData struct {
	freq      int64
	bandwidth int64
	mode      string
}

func (b *Bookmarker) setSdrData(data *SdrData) error {
	if err := b.client.SetDemod(data.mode, data.bandwidth); err != nil {
		return fmt.Errorf("SetDemod error: %v", err)
	}
	if err := b.client.SetFreq(data.freq); err != nil {
		return fmt.Errorf("SetFreq error: %v", err)
	}
	return nil
}

func (b *Bookmarker) getSdrData() (*SdrData, error) {
	mode, bandwidth, err := b.client.GetDemod()
	if err != nil {
		return nil, fmt.Errorf("get mod err: %v", err)
	}
	freq, err := b.client.GetFreq()
	if err != nil {
		return nil, fmt.Errorf("get freq err: %v", err)
	}
	fmt.Println(mode, bandwidth, freq)
	return &SdrData{freq: freq, bandwidth: bandwidth, mode: mode}, nil
}

func main() {
	b := &Bookmarker{}
	b.app = app.New()
	addr := "127.0.0.1:7356"
	b.client = gqrx.NewClient(addr)
	if err := b.client.Connect(); err != nil {
		log.Panicf("Error connecting: %v", err)
	}
	w := b.app.NewWindow(fmt.Sprintf("GQRX-Bookmarker v%s", version))
	ng := container.NewGridWithColumns(6)
	ng.Add(widget.NewButton("Add Bookmark", func() {
		sdrData, err := b.getSdrData()
		if err != nil {
			log.Printf("getSdrData error: %v", err)
			return
		}
		ng.Add(widget.NewButton(fmt.Sprintf("%.3f", float64(sdrData.freq)/1000000), func() {
			if err := b.setSdrData(sdrData); err != nil {
				log.Printf("setSdrData error: %v", err)
			}
		}))
	}))
	w.SetContent(ng)
	w.ShowAndRun()
}
