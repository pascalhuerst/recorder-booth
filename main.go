package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/pascalhuerst/framebuffer"
	"github.com/pascalhuerst/recorder-booth/audio"
	"github.com/pascalhuerst/recorder-booth/io"
	"github.com/pascalhuerst/recorder-booth/storage"
	"github.com/pascalhuerst/recorder-booth/ui"
	"github.com/skip2/go-qrcode"
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

	qrdata, err := qrcode.Encode("http://domestic-affairs.de", qrcode.Medium, 64)
	if err != nil {
		fmt.Printf("Cannot create qr code: %v\n", err)
		return
	}

	bbb := bytes.NewBuffer(qrdata)

	img, err := png.Decode(bbb)
	if err != nil {
		log.Fatal(err)
	}

	d1, err := makeDisplay(0)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	d1.DrawImage(img)
	/*
		base := make([]byte, 1024)
		buf := bytes.NewBuffer(base)

		r := 0
		for i := len(base); i != 0; i -= r {
			forread := make([]byte, int(math.Min(float64(33), float64(i))))
			r, err = buf.Read(forread)
			if err != nil {
				fmt.Printf("ERR: %v\n", err)
			}
			fmt.Printf("r=%d buf.Len=%d len(forread)=%d\n", r, buf.Len(), len(forread))
		}

		fmt.Printf("buf.Len=%d\n", buf.Len())
	*/

	//rss1 := ui.NewRecordStatusScreen(d1)

	d2, err := makeDisplay(1)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	rss2 := ui.NewRecordStatusScreen(d2)
	rss2.SetTitle("recording")

	//////////////////////////////

	bus := io.NewI2C(1)
	gpioController1 := io.NewGPIOControllerPCF8574T(bus, 0x21)
	gpioController2 := io.NewGPIOControllerPCF8574T(bus, 0x20)

	mapping := map[int]ui.LedGPIOMapping{
		0: {Controller: gpioController1, GPIOIndex: 7, Invert: false},
		1: {Controller: gpioController1, GPIOIndex: 6, Invert: false},
		2: {Controller: gpioController1, GPIOIndex: 5, Invert: false},
		3: {Controller: gpioController1, GPIOIndex: 4, Invert: false},
		4: {Controller: gpioController1, GPIOIndex: 3, Invert: false},
		5: {Controller: gpioController1, GPIOIndex: 2, Invert: false},
		6: {Controller: gpioController1, GPIOIndex: 1, Invert: false},
		7: {Controller: gpioController1, GPIOIndex: 0, Invert: false},
		8: {Controller: gpioController2, GPIOIndex: 7, Invert: false},
		9: {Controller: gpioController2, GPIOIndex: 6, Invert: false},
	}
	leftLedBar := ui.NewLedLevelMeter(mapping, ui.ModeBar)

	gpioController3 := io.NewGPIOControllerPCF8574T(bus, 0x22)
	gpioController4 := io.NewGPIOControllerPCF8574T(bus, 0x23)

	mapping = map[int]ui.LedGPIOMapping{
		0: {Controller: gpioController3, GPIOIndex: 7, Invert: false},
		1: {Controller: gpioController3, GPIOIndex: 6, Invert: false},
		2: {Controller: gpioController3, GPIOIndex: 5, Invert: false},
		3: {Controller: gpioController3, GPIOIndex: 4, Invert: false},
		4: {Controller: gpioController3, GPIOIndex: 3, Invert: false},
		5: {Controller: gpioController3, GPIOIndex: 2, Invert: false},
		6: {Controller: gpioController3, GPIOIndex: 1, Invert: false},
		7: {Controller: gpioController3, GPIOIndex: 0, Invert: false},
		8: {Controller: gpioController4, GPIOIndex: 7, Invert: false},
		9: {Controller: gpioController4, GPIOIndex: 6, Invert: false},
	}
	rightLedBar := ui.NewLedLevelMeter(mapping, ui.ModeBar)

	///////////////////////////////

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
			//d1V := float32(80+v.RmsDB.Left) / 100.0
			//rss1.SetLevel(d1V)

			//pl := v.Rms.Left * 10.0
			pl := (80.0 - math.Abs(2.0*math.Max(v.RmsDB.Left, -40.0))) / 8.0 // 0..10
			nLedsL := int(math.Round(pl - 0.5))
			leftLedBar.Set(nLedsL)

			d2V := float32(80+v.RmsDB.Right) / 100.0
			rss2.SetLevel(d2V)

			//pr := v.Rms.Right * 10.0
			pr := (80.0 - math.Abs(2.0*math.Max(v.RmsDB.Right, -40.0))) / 8.0 // 0..10
			nLedsR := int(math.Round(pr - 0.5))
			rightLedBar.Set(nLedsR)

			counter++
			if counter >= 40 {
				counter = 0
				//	fmt.Printf("%s\n", v.String())
			}
		}
	}()

	clippingLed := ui.NewLed(ui.LedGPIOMapping{Controller: gpioController2, GPIOIndex: 0, Invert: true})
	headroomCh := make(chan audio.HeadroomAnalyzerResult)
	go func() {
		counter := 0
		lastClippingCount := 0

		for {
			v := <-headroomCh
			// Test: Turn on red led

			if v.ClippingCount > lastClippingCount {
				clippingLed.Set(true)
			} else {
				clippingLed.Set(false)
			}

			lastClippingCount = v.ClippingCount

			counter++
			if counter >= 100 {
				counter = 0
				//fmt.Printf("%s\n", v.String())
			}
		}
	}()

	metricsCh := make(chan audio.Metrics)
	go func() {
		for {
			v := <-metricsCh
			rss2.SetDuration(v.Duration)
		}
	}()

	analyzer := audio.NewAnalyzer()
	headroomAnalyzer := audio.NewHeadroomAnalyzer(headroomCh)
	analyzer.Add(headroomAnalyzer)
	rmsAnalyzer := audio.NewRmsAnalyzer(rmsCh)
	analyzer.Add(rmsAnalyzer)

	manager := storage.NewManager()
	manager.Add(storage.NewSnapcastStorageHandler("/tmp/stream-pipe", "RecorderBooth"))
	//manager.Add(storage.NewChunkStorageHandler("/tmp/chunks", "RecorderBooth", 1024*32))
	manager.Add(storage.NewHTTPStorageHandler("http://server.lan:8080/upload", "RecorderBooth", 1024*256))

	recorder := audio.NewRecorder(recordDevice, cfg, manager.InputChannel(), analyzer.InputChannel(), metricsCh)
	err = recorder.Start()
	if err != nil {
		fmt.Printf("Error starting recorder: %v\n", err)
	}

	select {}

	return
}
