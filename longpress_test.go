package ui

import (
	"testing"
	"time"

	"github.com/akrck02/go-renderer/input"
)

// lighthouseButton is one button, counting its clicks and its long presses.
func lighthouseButton(clicks, longPresses *int) func(ui *Context) {
	return func(ui *Context) {
		if ui.Button("North light") {
			*clicks++
		}
		if ui.LongPressed() {
			*longPresses++
		}
	}
}

// putFingerOn puts one finger down in the middle of a control laid out in the last frame.
func (h *harness) putFingerOn(id ID, build func(ui *Context)) (x, y float64) {
	area := h.rect(id)
	x, y = float64(area.X+area.Width/2), float64(area.Y+area.Height/2)
	h.state.Feed().SetTouch(1, x, y)
	h.moveTo(x, y)
	h.press()
	h.frame(build)
	return x, y
}

func (h *harness) liftFinger(build func(ui *Context)) {
	h.release()
	h.state.Feed().LiftTouch(1)
	h.frame(build)
}

func TestTheMouseLongPressesWithItsRightButtonOnly(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	clicks, longPresses := 0, 0
	build := lighthouseButton(&clicks, &longPresses)
	h.frame(build)
	area := h.rect(idOf("North light"))
	h.moveTo(float64(area.X+area.Width/2), float64(area.Y+area.Height/2))
	h.press()
	h.frame(build)
	*now = now.Add(time.Second)
	h.frame(build)
	h.release()
	h.frame(build)
	if clicks != 1 || longPresses != 0 {
		t.Fatalf("%d clicks and %d long presses after holding the left button; want a slow click", clicks, longPresses)
	}
	h.state.Feed().SetButton(input.MouseRight, true)
	h.frame(build)
	h.state.Feed().SetButton(input.MouseRight, false)
	h.frame(build)
	if longPresses != 1 {
		t.Fatalf("%d long presses after a right click; want one", longPresses)
	}
}
