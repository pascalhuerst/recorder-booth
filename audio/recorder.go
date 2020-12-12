package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sync/atomic"

	"github.com/yobert/alsa"
)

// Config encapsulates recording specifics
type Config struct {
	Samplerate int
	Channels   int
	Format     alsa.FormatType
	BufferSize int
}

// Frame represents an audio frame
type Frame struct {
	Left  int16
	Right int16
}

// Recorder can record audio
type Recorder struct {
	device *alsa.Device
	buffer alsa.Buffer
	xRuns  int

	config Config

	isRunning uint32
	ctx       context.Context
	shutdown  context.CancelFunc

	rawStream   chan []byte
	frameStream chan []Frame
}

// NewRecorder recorder factory
func NewRecorder(device *alsa.Device, config Config, rawStream chan []byte, frameStream chan []Frame) *Recorder {
	return &Recorder{
		device:      device,
		buffer:      alsa.Buffer{},
		config:      config,
		isRunning:   0,
		rawStream:   rawStream,
		frameStream: frameStream,
	}
}

// IsRunning returns true if recorder is running
func (a *Recorder) IsRunning() bool {
	return atomic.LoadUint32(&a.isRunning) != 0
}

// Start starts the recorder
func (a *Recorder) Start() error {

	if a.IsRunning() {
		return fmt.Errorf("Cannot start recorder: Already running")
	}

	atomic.StoreUint32(&a.isRunning, 1)
	go a.run()
	return nil
}

// Stop stops the recorder
func (a *Recorder) Stop() error {

	if !a.IsRunning() {
		return fmt.Errorf("Cannot stop recorder: Not running")
	}

	a.shutdown()
	return nil
}

func (a *Recorder) setup() error {

	if err := a.device.Open(); err != nil {
		return fmt.Errorf("Cannot open alsa device: %v", err)
	}

	confirmedChannels, err := a.device.NegotiateChannels(a.config.Channels)
	if err != nil || confirmedChannels != a.config.Channels {
		return fmt.Errorf("Cannot negotiate channels: %v", err)
	}

	confirmedRate, err := a.device.NegotiateRate(a.config.Samplerate)
	if err != nil || confirmedRate != a.config.Samplerate {
		return fmt.Errorf("Cannot negotiate sample rate: %v", err)
	}

	confirmedFormat, err := a.device.NegotiateFormat(a.config.Format)
	if err != nil || confirmedFormat != a.config.Format {
		return fmt.Errorf("Cannot negotiate sample format: %v", err)
	}

	confirmedBufferSize, err := a.device.NegotiateBufferSize(a.config.BufferSize)
	if err != nil || confirmedBufferSize != a.config.BufferSize {
		return fmt.Errorf("Cannot negotiate buffer size: %v", err)
	}

	a.buffer = a.device.NewBufferFrames(a.config.BufferSize)

	if err = a.device.Prepare(); err != nil {
		return fmt.Errorf("Cannot prepare recording: %v", err)
	}

	return nil
}

func (a *Recorder) run() {

	defer atomic.StoreUint32(&a.isRunning, 0)
	ctx, shutdown := context.WithCancel(context.Background())
	a.ctx = ctx
	a.shutdown = shutdown

setup:
	err := a.setup()
	if err != nil {
		fmt.Printf("ERROR: %v", err)
		return
	}

	for {
		select {
		case <-a.ctx.Done():
			fmt.Printf("Recorder received shutdown request....\n")
			a.device.Close()
			return
		default:
			err := a.device.Read(a.buffer.Data)
			if err != nil {
				a.xRuns++
				fmt.Printf("ERROR: %v, xruns: %d\n", err, a.xRuns)
				a.device.Close()
				goto setup
			}

			if a.rawStream != nil {
				a.rawStream <- a.buffer.Data
			}

			if a.frameStream != nil {

				buf := bytes.NewBuffer(a.buffer.Data)
				frames := []Frame{}

				for len(buf.Bytes()) > 0 {
					frame := Frame{}
					// Left Channel
					binary.Read(buf, binary.LittleEndian, &frame.Left)
					// Right Channel
					binary.Read(buf, binary.LittleEndian, &frame.Right)
					frames = append(frames, frame)
				}
				a.frameStream <- frames
			}
		}
	}
}
