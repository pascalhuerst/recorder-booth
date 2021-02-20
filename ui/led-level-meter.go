package ui

import (
	"fmt"
)

// Mode sets the display mode dot/bar
type Mode byte

const (
	// ModeBar shows a bar
	ModeBar = Mode(iota)
	// ModeDot shows a dot
	ModeDot
)

// LedLevelMeter type
type LedLevelMeter struct {
	mappings     map[int]LedGPIOMapping
	segmentCount int
	mode         Mode
	dirty        bool
}

// NewLedLevelMeter factory
func NewLedLevelMeter(mappings map[int]LedGPIOMapping, mode Mode) *LedLevelMeter {

	ret := &LedLevelMeter{
		segmentCount: len(mappings),
		mappings:     mappings,
		mode:         mode,
	}

	return ret
}

// Set updates the display with a new value
func (l *LedLevelMeter) Set(v int) error {
	if v >= l.segmentCount {
		return fmt.Errorf("Cannot set level meter: Out of range %d/%d", v, l.segmentCount)
	}

	if l.mode == ModeBar {
		for i := 0; i < l.segmentCount; i++ {
			if m, ok := l.mappings[i]; ok {
				if m.Controller == nil {
					return fmt.Errorf("Cannot conrol led. No controller set for %d", i)
				}

				if i > v != m.Invert {
					m.Controller.Set(m.GPIOIndex, true)
					continue
				}
				m.Controller.Set(m.GPIOIndex, false)
			} else {
				return fmt.Errorf("No mapping found for value: %d which is needed to set %d", i, v)
			}
		}
	} else {
		return fmt.Errorf("Dot mode not implemented")
	}

	return nil
}
