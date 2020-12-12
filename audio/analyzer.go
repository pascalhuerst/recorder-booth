package audio

import (
	"fmt"
	"sync"
)

// Analyzer can analyze frames for clipping, rms, ...
type Analyzer struct {
	frameStream chan []Frame
	mutex       sync.Mutex
	analyzers   []AnalyzerInterface
}

// AnalyzerInterface used to add analyzers
type AnalyzerInterface interface {
	process([]Frame)
}

// AnalyzerFrame is a frame representation in float
type AnalyzerFrame struct {
	Left  float64
	Right float64
}

// NewAnalyzer analyzer factory
func NewAnalyzer() *Analyzer {
	ret := &Analyzer{
		frameStream: make(chan []Frame),
	}
	go ret.run()

	return ret
}

// InputChannel returns the input channel for the analyzer
func (a *Analyzer) InputChannel() chan []Frame {
	return a.frameStream
}

// Add adds an analyzer
func (a *Analyzer) Add(ai AnalyzerInterface) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.analyzers = append(a.analyzers, ai)
}

func (a *Analyzer) run() {
	fmt.Printf("Starting analyzer\n")
	for {
		data := <-a.frameStream

		a.mutex.Lock()
		curAnalyzers := a.analyzers
		a.mutex.Unlock()

		for _, ca := range curAnalyzers {
			go ca.process(data)
		}
	}
}
