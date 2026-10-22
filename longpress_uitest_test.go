package ui_test

import (
	"testing"
	"time"

	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// longPressNeeds is a little more than the half second Go UI waits for a long press.
const longPressNeeds = 750 * time.Millisecond

// countingButton is one button, counting its clicks and its long presses.
func countingButton(clicks, longPresses *int) func(*ui.Context) {
	return func(context *ui.Context) {
		if context.Button("North light") {
			*clicks++
		}
		if context.LongPressed() {
			*longPresses++
		}
	}
}

// openWindow makes a test window and runs its first frame, so the controls have a place.
func openWindow(t *testing.T, build func(*ui.Context)) *uitest.Window {
	t.Helper()
	window, err := uitest.New(build)
	if err != nil {
		t.Fatal(err)
	}
	window.Frame()
	return window
}

// middleOf is the middle of an element of the last frame.
func middleOf(t *testing.T, window *uitest.Window, label string) (x, y float32) {
	t.Helper()
	element, err := window.Element(label)
	if err != nil {
		t.Fatal(err)
	}
	return element.Rect.X + element.Rect.Width/2, element.Rect.Y + element.Rect.Height/2
}

func TestAFingerHeldStillIsALongPressAndNoClick(t *testing.T) {
	clicks, longPresses := 0, 0
	window := openWindow(t, countingButton(&clicks, &longPresses))
	x, y := middleOf(t, window, "North light")
	window.TouchDown(1, x, y)
	if wait, asked := window.State().Feed().TakeFrameRequest(); !asked || wait > longPressNeeds {
		t.Fatalf("frame asked for %v (%v); want one by the time the press becomes long", wait, asked)
	}
	window.Wait(300 * time.Millisecond)
	if longPresses != 0 {
		t.Fatal("a long press before half a second")
	}
	window.Wait(longPressNeeds - 300*time.Millisecond)
	if longPresses != 1 {
		t.Fatalf("%d long presses after holding for 750 ms; want one", longPresses)
	}
	window.TouchUp(1)
	if clicks != 0 {
		t.Fatal("lifting the finger after a long press clicked the button")
	}
}

func TestAShortTouchIsAClickAndNoLongPress(t *testing.T) {
	clicks, longPresses := 0, 0
	window := openWindow(t, countingButton(&clicks, &longPresses))
	x, y := middleOf(t, window, "North light")
	window.TouchDown(1, x, y)
	window.Wait(100 * time.Millisecond)
	window.TouchUp(1)
	if clicks != 1 || longPresses != 0 {
		t.Fatalf("%d clicks and %d long presses; want a click only", clicks, longPresses)
	}
}

func TestAFingerThatMovesIsNoLongPress(t *testing.T) {
	clicks, longPresses := 0, 0
	window := openWindow(t, countingButton(&clicks, &longPresses))
	x, y := middleOf(t, window, "North light")
	window.TouchDown(1, x, y)
	window.TouchMove(1, x, y+20)
	window.Wait(time.Second)
	if longPresses != 0 {
		t.Fatal("a finger that moved became a long press")
	}
}
