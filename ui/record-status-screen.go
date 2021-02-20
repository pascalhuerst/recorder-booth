package ui

import (
	"fmt"
	"sync"
	"time"
)

// RecordStatusScreen represents a screen which is shown during recording
type RecordStatusScreen struct {
	d        *Display
	mutex    sync.Mutex
	title    string
	duration time.Duration
	level    float32
}

// SetLevel is used to set the level
func (s *RecordStatusScreen) SetLevel(level float32) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if level > s.level {
		s.level = level
	}
}

// SetTitle is used to set the title
func (s *RecordStatusScreen) SetTitle(title string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.title = title
}

// SetDuration is used to set the duration
func (s *RecordStatusScreen) SetDuration(duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.duration = duration
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func (s *RecordStatusScreen) update() {

	for {
		s.mutex.Lock()
		s.d.drawProgressBar( /*y*/ 20, 120, 6, float32(s.level))
		s.level *= 0.98
		s.mutex.Unlock()
		time.Sleep(time.Millisecond * 20)
	}
}

func (s *RecordStatusScreen) refresh() {
	/*
		fontHeightBig := s.d.textFaceBig.Metrics().Height.Ceil()
		fontHeightSmall := s.d.textFaceSmall.Metrics().Height.Ceil()

		y := fontHeightBig

		s.d.clear()
		s.d.drawTextAt(1, y, s.title, true, alignLeft)
		y += 4

		s.d.drawHorizontalLine(1, 128, y)
		y += 20
	*/
	s.d.drawProgressBar( /*y*/ 20, 120, 6, float32(s.level))
	/*
		y += 12 + fontHeightSmall

		s.d.drawTextAt(4, y, fmt.Sprintf("%s", fmtDuration(s.duration)), false, alignLeft)
	*/
}

// NewRecordStatusScreen factory
func NewRecordStatusScreen(d *Display) *RecordStatusScreen {
	ret := &RecordStatusScreen{
		d:        d,
		title:    "",
		duration: 0,
		level:    0,
	}
	go ret.update()
	return ret
}
