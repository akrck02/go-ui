package ui

import (
	"strconv"
	"testing"
	"time"
)

func TestAFingerDragsAListAndItGlidesOnWithoutTouchingWhatItStartedOn(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	clicked := 0
	build := func(ui *Context) {
		ui.ScrollArea("list", Layout{Width: Fixed(300), Height: Fixed(200)}, func() {
			for index := range 50 {
				if ui.Button("row " + strconv.Itoa(index)) {
					clicked++
				}
			}
		})
	}
	h.frame(build)
	h.frame(build)
	step := func(y float64) {
		*now = now.Add(16 * time.Millisecond)
		h.state.SetTouch(1, 20, y)
		h.moveTo(20, y)
		h.frame(build)
	}
	// on a row's button, which the drag must not press
	h.state.SetTouch(1, 20, 150)
	h.moveTo(20, 150)
	h.press()
	h.frame(build)
	for y := 140.0; y >= 50; y -= 10 {
		step(y)
	}
	h.release()
	h.state.LiftTouch(1)
	*now = now.Add(16 * time.Millisecond)
	h.frame(build)
	memory := remember(h.ui, idOf("list"), func() *scrollMemory { return nil })
	afterDrag := memory.offset
	if clicked != 0 {
		t.Fatalf("%d rows clicked: a drag touches nothing", clicked)
	}
	if afterDrag < 80 {
		t.Fatalf("offset %v after dragging 100 pixels up; want the content to follow", afterDrag)
	}
	for range 10 {
		*now = now.Add(16 * time.Millisecond)
		h.frame(build)
	}
	if memory.offset <= afterDrag {
		t.Fatalf("offset %v after letting go while moving; want it to glide on from %v", memory.offset, afterDrag)
	}
}

func TestATouchThatHardlyMovesStillPresses(t *testing.T) {
	h := newHarness(t)
	clicked := 0
	build := func(ui *Context) {
		ui.ScrollArea("list", Layout{Width: Fixed(300), Height: Fixed(200)}, func() {
			for index := range 50 {
				if ui.Button("row " + strconv.Itoa(index)) {
					clicked++
				}
			}
		})
	}
	h.frame(build)
	h.frame(build)
	row := h.rect(idOf("row 0"))
	x, y := float64(row.X+row.Width/2), float64(row.Y+row.Height/2)
	h.state.SetTouch(1, x, y)
	h.moveTo(x, y)
	h.press()
	h.frame(build)
	h.state.SetTouch(1, x, y+3)
	h.moveTo(x, y+3)
	h.frame(build)
	h.release()
	h.state.LiftTouch(1)
	h.frame(build)
	if clicked != 1 {
		t.Fatalf("%d clicks; a touch moving 3 pixels is a touch", clicked)
	}
}

func TestARowScrolledOutOfViewDoesNotTakeATouchMeantForWhatIsThere(t *testing.T) {
	h := newHarness(t)
	rowClicks, barClicks := 0, 0
	build := func(ui *Context) {
		ui.Box(Layout{Direction: Vertical, Width: Fixed(300), Height: Fixed(300)}, func() {
			ui.ScrollArea("list", Layout{Width: Grow(), Height: Grow()}, func() {
				for index := range 40 {
					if ui.Button("row " + strconv.Itoa(index)) {
						rowClicks++
					}
				}
			})
			if ui.Button("bar") {
				barClicks++
			}
		})
	}
	h.frame(build)
	h.frame(build)
	bar := h.rect(idOf("bar"))
	h.moveTo(float64(bar.X+4), float64(bar.Y+bar.Height/2)) // where rows run on under the list's edge
	h.press()
	h.frame(build)
	h.release()
	h.frame(build)
	if barClicks != 1 || rowClicks != 0 {
		t.Fatalf("bar clicked %d, rows %d: the touch is the bar's", barClicks, rowClicks)
	}
}
