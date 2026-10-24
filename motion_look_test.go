package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

func TestAFadedBoxDrawsItselfAndWhatItHoldsSeeThrough(t *testing.T) {
	h := newHarness(t)
	white := draw2d.MustHex("#ffffff")
	h.frame(func(ui *Context) {
		ui.Frame(Layout{Width: Fixed(120), Height: Fixed(40)}, Look{Fill: white, Fade: 0.75}, func() { ui.Label("Inside") })
	})
	for _, command := range h.ui.drawing.Commands {
		if command.Kind == draw2d.RectangleKind && command.Area.Width == 120 && command.Fill[3] != 0.25 {
			t.Errorf("the box's fill is %v seen, not a quarter", command.Fill[3])
		}
		if command.Text == "Inside" && command.Fill[3] > 0.25+1e-6 {
			t.Errorf("the text inside is %v seen, not faded with its box", command.Fill[3])
		}
	}
}

func TestAShiftedBoxIsDrawnMovedButStaysInPlaceInTheLayout(t *testing.T) {
	placesWith := func(shift float32) (box, inside, below float32) {
		h := newHarness(t)
		h.frame(func(ui *Context) {
			ui.Frame(Layout{Width: Fixed(120), Height: Fixed(40)}, Look{Fill: draw2d.MustHex("#ffffff"), ShiftY: shift}, func() { ui.Label("Inside") })
			ui.Label("Below")
		})
		for _, command := range h.ui.drawing.Commands {
			switch {
			case command.Kind == draw2d.RectangleKind && command.Area.Width == 120:
				box = command.Area.Y
			case command.Text == "Inside":
				inside = command.Area.Y
			case command.Text == "Below":
				below = command.Area.Y
			}
		}
		return box, inside, below
	}
	box, inside, below := placesWith(0)
	shiftedBox, shiftedInside, shiftedBelow := placesWith(30)
	if shiftedBox != box+30 || shiftedInside != inside+30 {
		t.Errorf("the box and its text are drawn at %v and %v, not 30 below %v and %v", shiftedBox, shiftedInside, box, inside)
	}
	if shiftedBelow != below {
		t.Errorf("what follows moved from %v to %v: the shift moved the layout too", below, shiftedBelow)
	}
}
