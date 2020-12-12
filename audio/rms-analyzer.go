package audio

import (
	"fmt"
	"math"
)

// RmsAnalyzerResult is the output of this analyzer
type RmsAnalyzerResult struct {
	Rms   AnalyzerFrame
	RmsDB AnalyzerFrame
}

func (r *RmsAnalyzerResult) String() string {
	ret := "Level:\n"
	ret += fmt.Sprintf("  Linear: l: %v\n", r.Rms.Left)
	ret += fmt.Sprintf("          r: %v\n", r.Rms.Right)
	ret += fmt.Sprintf("  Log   : l: %v dB\n", r.RmsDB.Left)
	ret += fmt.Sprintf("          r: %v dB\n", r.RmsDB.Right)
	return ret
}

// RmsAnalyzer can analyze samples for it's rams value
type RmsAnalyzer struct {
	output  chan RmsAnalyzerResult
	counter int
}

// NewRmsAnalyzer factory
func NewRmsAnalyzer(output chan RmsAnalyzerResult) *RmsAnalyzer {
	return &RmsAnalyzer{
		output: output,
	}
}

func (r *RmsAnalyzer) process(frames []Frame) {

	result := RmsAnalyzerResult{}
	nSamples := len(frames)

	for _, frame := range frames {
		floatL := float64(frame.Left) / float64(math.MaxInt16)
		result.Rms.Left += floatL * floatL
		floatR := float64(frame.Right) / float64(math.MaxInt16)
		result.Rms.Right += floatR * floatR
	}

	result.Rms.Left = math.Sqrt(result.Rms.Left / float64(nSamples))
	result.Rms.Right = math.Sqrt(result.Rms.Right / float64(nSamples))
	result.RmsDB.Left = 20 * math.Log10(result.Rms.Left)
	result.RmsDB.Right = 20 * math.Log10(result.Rms.Right)

	if r.output != nil {
		r.output <- result
	}
}
