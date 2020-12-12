package audio

import (
	"fmt"
	"math"
)

// HeadroomAnalyzer can analyze samples for it's rams value
type HeadroomAnalyzer struct {
	output chan HeadroomAnalyzerResult
	result HeadroomAnalyzerResult
}

// HeadroomAnalyzerResult is the output of this analyzer
type HeadroomAnalyzerResult struct {
	LastHeadroom  Frame
	WorstHeadroom Frame
	ClippingCount int
}

func (h *HeadroomAnalyzerResult) String() string {
	ret := "Headroom:\n"
	ret += fmt.Sprintf("  Last :\tl: %05d\tr: %05d\n", h.LastHeadroom.Left, h.LastHeadroom.Right)
	ret += fmt.Sprintf("  Worst:\tl: %05d\tr: %05d\n", h.WorstHeadroom.Left, h.WorstHeadroom.Right)
	ret += fmt.Sprintf("  Clipping Frames: %d\n", h.ClippingCount)
	return ret
}

// NewHeadroomAnalyzer factory
func NewHeadroomAnalyzer(output chan HeadroomAnalyzerResult) *HeadroomAnalyzer {
	return &HeadroomAnalyzer{
		output: output,
		result: HeadroomAnalyzerResult{
			LastHeadroom: Frame{
				Left:  math.MaxInt16,
				Right: math.MaxInt16,
			},
			WorstHeadroom: Frame{
				Left:  math.MaxInt16,
				Right: math.MaxInt16,
			},
			ClippingCount: 0,
		},
	}
}

func min(a, b int16) int16 {
	a = abs(a)
	b = abs(b)
	if a < b {
		return a
	}
	return b
}

func abs(a int16) int16 {
	if a < 0 {
		return -a
	}
	return a
}

func (h *HeadroomAnalyzer) process(frames []Frame) {

	h.result.LastHeadroom.Left = math.MaxInt16
	h.result.LastHeadroom.Right = math.MaxInt16

	for _, frame := range frames {
		headRoomLeft := math.MaxInt16 - abs(frame.Left)
		h.result.LastHeadroom.Left = min(h.result.LastHeadroom.Left, headRoomLeft)
		headRoomRight := math.MaxInt16 - abs(frame.Right)
		h.result.LastHeadroom.Right = min(h.result.LastHeadroom.Right, headRoomRight)
	}

	if h.result.LastHeadroom.Left <= 1 || h.result.LastHeadroom.Right <= 1 {
		h.result.ClippingCount++
	}

	h.result.WorstHeadroom.Left = min(h.result.WorstHeadroom.Left, h.result.LastHeadroom.Left)
	h.result.WorstHeadroom.Right = min(h.result.WorstHeadroom.Right, h.result.LastHeadroom.Right)

	if h.output != nil {
		h.output <- h.result
	}

}
