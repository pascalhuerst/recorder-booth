package main

import (
	"fmt"
	"os"

	"github.com/pascalhuerst/framebuffer"
	"github.com/pascalhuerst/recorder-booth/audio"
	"github.com/pascalhuerst/recorder-booth/storage"
	"github.com/pascalhuerst/recorder-booth/ui"
	"github.com/yobert/alsa"
)

func makeDisplay(index int) (*ui.Display, error) {

	dev := fmt.Sprintf("/dev/fb%d", index)
	fbCanvas, err := framebuffer.Open(nil, dev)
	if err != nil {
		return nil, fmt.Errorf("Cannot open framebuffer %s: %v", dev, err)
	}

	fb, err := fbCanvas.Image()
	if err != nil {
		return nil, fmt.Errorf("Cannot get framebuffer: %v", err)
	}

	display, err := ui.NewDisplay(fb)
	if err != nil {
		return nil, fmt.Errorf("Cannot create display: %v", err)
	}

	return display, nil
}

func main() {

	d1, err := makeDisplay(0)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	rss1 := ui.NewRecordStatusScreen(d1)

	d2, err := makeDisplay(1)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	rss2 := ui.NewRecordStatusScreen(d2)

	//////////////////////////////

	cfg := audio.Config{
		BufferSize: 1024,
		Channels:   2,
		Format:     alsa.S16_LE,
		Samplerate: 48000,
	}

	cards, err := alsa.OpenCards()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer alsa.CloseCards(cards)

	// use the first recording device we find
	var recordDevice *alsa.Device

	for _, card := range cards {
		devices, err := card.Devices()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, device := range devices {
			if device.Type != alsa.PCM {
				continue
			}
			if device.Record && recordDevice == nil {
				recordDevice = device
			}
		}
	}

	if recordDevice == nil {
		fmt.Println("No recording device found")
		os.Exit(1)
	}
	fmt.Printf("Recording device: %v\n", recordDevice)

	rmsCh := make(chan audio.RmsAnalyzerResult)
	go func() {
		counter := 0
		for {
			v := <-rmsCh
			d1V := float32(100+v.RmsDB.Left) / 100.0
			rss1.SetLevel(d1V)

			d2V := float32(100+v.RmsDB.Right) / 100.0
			rss2.SetLevel(d2V)

			counter++
			if counter >= 100 {
				counter = 0
				fmt.Printf("%s\n", v.String())
			}
		}
	}()

	headroomCh := make(chan audio.HeadroomAnalyzerResult)
	go func() {
		counter := 0
		for {
			v := <-headroomCh
			counter++
			if counter >= 100 {
				counter = 0
				fmt.Printf("%s\n", v.String())
			}
		}
	}()

	analyzer := audio.NewAnalyzer()
	headroomAnalyzer := audio.NewHeadroomAnalyzer(headroomCh)
	analyzer.Add(headroomAnalyzer)
	rmsAnalyzer := audio.NewRmsAnalyzer(rmsCh)
	analyzer.Add(rmsAnalyzer)

	manager := storage.NewManager()
	manager.Add(storage.NewSnapcastStorageHandler("/tmp/stream-pipe", "RecorderBooth"))
	manager.Add(storage.NewChunkStorageHandler("/tmp/chunks", "RecorderBooth", 1024*32))

	recorder := audio.NewRecorder(recordDevice, cfg, manager.InputChannel(), analyzer.InputChannel())
	err = recorder.Start()
	if err != nil {
		fmt.Printf("Error starting recorder: %v\n", err)
	}

	select {}

	return
}
