package ui

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"path"

	"github.com/golang/freetype/truetype"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Display is the central object of the Crust UI
type Display struct {
	width           int
	height          int
	dpi             float64
	target          draw.Image
	fg, bg          *image.Uniform
	font            *truetype.Font
	textFaceSmall   font.Face
	textFaceBig     font.Face
	textDrawerSmall font.Drawer
	textDrawerBig   font.Drawer
}

type alignement int

const (
	alignLeft = alignement(iota)
	alignCenter
)

func (d *Display) clear() {
	draw.Draw(d.target, d.target.Bounds(), d.bg, image.ZP, draw.Src)
}

// DrawImage dras an image
func (d *Display) DrawImage(img image.Image) {

	d.clear()

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			if c.Y > 0 {
				draw.Draw(d.target, image.Rect(x, y, x+1, y+1), d.fg, image.ZP, draw.Src)
			} else {
				draw.Draw(d.target, image.Rect(x, y, x+1, y+1), d.bg, image.ZP, draw.Src)
			}
		}
	}

}

func (d *Display) drawTextAt(x, y int, text string, isBig bool, align alignement) {

	drawer := d.textDrawerSmall
	if isBig {
		drawer = d.textDrawerBig
	}

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	switch align {
	case alignLeft:
	case alignCenter:
		adv := drawer.MeasureString(text)
		drawer.Dot.X -= adv / 2
	}

	b, _ := drawer.BoundString(text)
	draw.Draw(d.target, image.Rect(b.Min.X.Ceil(), b.Min.Y.Ceil(), b.Max.X.Ceil(), b.Max.Y.Ceil()), d.bg, image.ZP, draw.Src)
	drawer.DrawString(text)
}

func (d *Display) drawHorizontalLine(x1, x2, y int) {
	draw.Draw(d.target, image.Rect(x1, y, x2, y+1), d.fg, image.ZP, draw.Src)
}

func (d *Display) drawProgressBar(y, w, h int, val float32) {

	x1 := (d.width - w) / 2
	x2 := x1 + w
	y1 := y
	y2 := y + h

	// Clean bar
	draw.Draw(d.target, image.Rect(x1+1, y1+1, x2-1, y2-1), d.bg, image.ZP, draw.Src)

	// Frame
	draw.Draw(d.target, image.Rect(x1, y1, x2, y1+1), d.fg, image.ZP, draw.Src)
	draw.Draw(d.target, image.Rect(x1, y2, x2, y2-1), d.fg, image.ZP, draw.Src)
	draw.Draw(d.target, image.Rect(x1, y1, x1+1, y2), d.fg, image.ZP, draw.Src)
	draw.Draw(d.target, image.Rect(x2, y1, x2-1, y2), d.fg, image.ZP, draw.Src)

	// Fill
	x2 = x1 + int(float32(w)*val)
	draw.Draw(d.target, image.Rect(x1, y1, x2, y2), d.fg, image.ZP, draw.Src)

}

func (d *Display) loadFont() error {

	dirs := []string{".", "../..", "/usr/share/fonts/truetype", "~"}

	var fontBytes []byte
	var err error

	for _, dir := range dirs {
		fname := path.Join(dir, "Recorder_8_Regular.ttf")
		fontBytes, err = ioutil.ReadFile(fname)
		if err == nil {
			break
		}
	}

	if len(fontBytes) == 0 {
		return fmt.Errorf("Couldn't read font file from any of %v", dirs)
	}

	d.font, err = truetype.Parse(fontBytes)
	if err != nil {
		return err
	}

	d.textFaceSmall = truetype.NewFace(d.font, &truetype.Options{
		Size:       8.0,
		SubPixelsX: 1,
		SubPixelsY: 1,
		DPI:        d.dpi,
		Hinting:    font.HintingNone,
	})

	d.textDrawerSmall = font.Drawer{
		Dst:  d.target,
		Src:  d.fg,
		Face: d.textFaceSmall,
	}

	d.textFaceBig = truetype.NewFace(d.font, &truetype.Options{
		Size:       12.0,
		SubPixelsX: 1,
		SubPixelsY: 1,
		DPI:        d.dpi,
		Hinting:    font.HintingNone,
	})

	d.textDrawerBig = font.Drawer{
		Dst:  d.target,
		Src:  d.fg,
		Face: d.textFaceBig,
	}

	return nil
}

// NewDisplay creates a new display with 'target' as backing buffer.
func NewDisplay(target draw.Image) (*Display, error) {

	ret := &Display{
		height: target.Bounds().Size().Y,
		width:  target.Bounds().Size().X,
		bg:     image.Black,
		fg:     image.White,
		dpi:    float64(72),
		target: target,
	}

	err := ret.loadFont()
	if err != nil {
		return nil, err
	}

	return ret, nil
}
