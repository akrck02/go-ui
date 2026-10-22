package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

// stripes is a backdrop that draws one band across the top of its area.
type stripes struct{ painted []draw2d.Rectangle }

func (backdrop *stripes) Paint(area draw2d.Rectangle, drawing *draw2d.List) {
	backdrop.painted = append(backdrop.painted, area)
	drawing.Rectangle(draw2d.Rectangle{X: area.X, Y: area.Y, Width: area.Width, Height: 2}, 0, draw2d.MustHex("#ff0000"), draw2d.Color{}, 0)
}

func TestABackdropPaintsInItsBoxArea(t *testing.T) {
	h := newHarness(t)
	backdrop := &stripes{}
	build := func(ui *Context) {
		ui.Frame(Layout{Width: Fixed(120), Height: Fixed(40)}, Look{Fill: draw2d.MustHex("#ffffff"), Backdrop: backdrop}, func() {
			ui.Label("Inside")
		})
	}
	h.frame(build)
	if len(backdrop.painted) != 1 || backdrop.painted[0].Width != 120 || backdrop.painted[0].Height != 40 {
		t.Fatalf("the backdrop painted %v, want once in the 120 × 40 box", backdrop.painted)
	}
}

func TestATextShadowGoesUnderTheTextAPixelAway(t *testing.T) {
	h := newHarness(t)
	shadow, ink := draw2d.MustHex("#203040"), draw2d.MustHex("#f0f0f0")
	h.frame(func(ui *Context) { ui.Text("Shaded", TextStyle{Color: ink, Shadow: shadow}) })
	var under, over *draw2d.Command
	for index := range h.ui.drawing.Commands {
		command := &h.ui.drawing.Commands[index]
		if command.Text != "Shaded" {
			continue
		}
		if command.Fill == shadow {
			under = command
		} else if command.Fill == ink && under != nil {
			over = command
		}
	}
	if under == nil || over == nil {
		t.Fatal("the shadow is not drawn before the text")
	}
	if under.Area.X != over.Area.X+1 || under.Baseline != over.Baseline+1 {
		t.Fatalf("the shadow is at %v,%v and the text at %v,%v: want a pixel right and down", under.Area.X, under.Baseline, over.Area.X, over.Baseline)
	}
}
