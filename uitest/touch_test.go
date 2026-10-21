package uitest

import (
	"testing"
	"time"

	ui "github.com/akrck02/go-ui"
)

// beaconButton counts the clicks and long presses of one button.
func beaconButton(clicks, longPresses *int) func(*ui.Context) {
	return func(context *ui.Context) {
		if context.Button("Beacon") {
			*clicks++
		}
		if context.LongPressed() {
			*longPresses++
		}
	}
}

func TestATapPressesAButton(t *testing.T) {
	clicks, longPresses := 0, 0
	window := newWindow(t, beaconButton(&clicks, &longPresses))
	if err := window.Tap("Beacon"); err != nil {
		t.Fatal(err)
	}
	if clicks != 1 || longPresses != 0 {
		t.Fatalf("%d clicks and %d long presses; want one click", clicks, longPresses)
	}
}

func TestALongPressOpensTheMenu(t *testing.T) {
	clicks, longPresses := 0, 0
	window := newWindow(t, beaconButton(&clicks, &longPresses))
	if err := window.LongPress("Beacon"); err != nil {
		t.Fatal(err)
	}
	if clicks != 0 || longPresses != 1 {
		t.Fatalf("%d clicks and %d long presses; want one long press", clicks, longPresses)
	}
}

func TestALongPressNeedsItsTime(t *testing.T) {
	clicks, longPresses := 0, 0
	window := newWindow(t, beaconButton(&clicks, &longPresses))
	beacon, err := window.Element("Beacon")
	if err != nil {
		t.Fatal(err)
	}
	middle := center(beacon.Rect)
	window.TouchDown(1, middle.x, middle.y)
	window.Frames(100) // many frames, but no time passes
	if longPresses != 0 {
		t.Fatal("a long press with no time passing")
	}
	window.Wait(time.Second)
	window.TouchUp(1)
	if longPresses != 1 {
		t.Fatalf("%d long presses after a second held", longPresses)
	}
}
