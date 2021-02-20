package ui

import (
	"fmt"

	"github.com/pascalhuerst/recorder-booth/io"
)

// LedGPIOMapping type to map gpios to leds
type LedGPIOMapping struct {
	Controller io.GPIOController
	GPIOIndex  int
	Invert     bool
}

// Led type
type Led struct {
	mapping LedGPIOMapping
}

// NewLed factory
func NewLed(mapping LedGPIOMapping) *Led {
	return &Led{
		mapping: mapping,
	}
}

// Set turn a led on or off
func (l *Led) Set(on bool) error {

	if l.mapping.Controller == nil {
		return fmt.Errorf("Cannot conrol led. No controller set")
	}

	if l.mapping.Invert {
		on = !on
	}

	l.mapping.Controller.Set(l.mapping.GPIOIndex, on)
	return nil
}
