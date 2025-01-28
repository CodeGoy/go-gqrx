package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/CodeGoy/go-gqrx"
	"log"
)

var (
	DefaultBandwidths = map[string]int64{
		"OFF":         0,
		"RAW":         10000,
		"AM":          10000,
		"AMS":         10000,
		"LSB":         2800,
		"USB":         2800,
		"CWL":         500,
		"CWR":         500,
		"CWU":         500,
		"CW":          500,
		"FM":          10000,
		"WFM":         160000,
		"WFM_ST":      160000,
		"WFM_ST_OIRT": 160000,
	}
)

type Remote struct {
	Gqrx *gqrx.Client
	App  fyne.App
}

func (r *Remote) gui() {
	w := r.App.NewWindow("set demod")
	ng := container.NewGridWithColumns(6)
	for _, demod := range gqrx.Demods {
		ng.Add(widget.NewButton(demod, func() {
			if err := r.Gqrx.SetDemod(demod, DefaultBandwidths[demod]); err != nil {
				log.Fatalf("Error setting demod: %v", err)
			}
		}))
	}
	w.SetContent(ng)
	w.ShowAndRun()
}

func main() {
	r := &Remote{}
	r.Gqrx = gqrx.NewClient("127.0.0.1:7356")
	if err := r.Gqrx.Connect(); err != nil {
		log.Panicf("Error connecting: %v", err)
	}
	r.App = app.New()
	r.gui()
}
