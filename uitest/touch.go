package uitest

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	"time"

	"github.com/akrck02/go-renderer/input"
)

// longPressHold is how long LongPress keeps the finger down: more than Go UI's long press needs.
const longPressHold = time.Second

// noFinger means no finger is the pointer.
const noFinger = -1

// TouchDown puts a finger on the screen and runs a frame. The first finger down also moves and
// presses the pointer, as a phone's backend does, so controls answer it as a click.
func (window *Window) TouchDown(finger int64, x, y float32) {
	defer window.doing("Touch", draw2d.Rect{}, false)()
	if len(window.state.Touches()) == 0 {
		window.pointerFinger = finger
	}
	window.state.SetTouch(finger, float64(x), float64(y))
	if finger == window.pointerFinger {
		window.state.MoveCursor(float64(x), float64(y))
		window.pointerMoved = true
		window.state.SetButton(input.MouseLeft, true)
		window.pointerDown = true
	}
	window.Frame()
}

// TouchMove moves a finger that is down and runs a frame.
func (window *Window) TouchMove(finger int64, x, y float32) {
	defer window.doing("Move a finger", draw2d.Rect{}, false)()
	window.state.SetTouch(finger, float64(x), float64(y))
	if finger == window.pointerFinger {
		window.state.MoveCursor(float64(x), float64(y))
	}
	window.Frame()
}

// TouchUp lifts a finger and runs a frame.
func (window *Window) TouchUp(finger int64) {
	defer window.doing("Lift a finger", draw2d.Rect{}, false)()
	window.state.LiftTouch(finger)
	if finger == window.pointerFinger {
		window.state.SetButton(input.MouseLeft, false)
		window.pointerDown = false
		window.pointerFinger = noFinger
	}
	window.Frame()
}

// Tap touches the middle of an element that can be pressed and lifts the finger at once.
func (window *Window) Tap(label string) error {
	target, err := window.pressableElement(label)
	if err != nil {
		return err
	}
	middle := center(target.Rect)
	defer window.doing(fmt.Sprintf("Tap %q", label), target.Rect, true)()
	window.TouchDown(1, middle.x, middle.y)
	window.TouchUp(1)
	return nil
}

// LongPress holds a finger still on an element for a second, letting the time pass, then lifts it.
func (window *Window) LongPress(label string) error {
	target, err := window.pressableElement(label)
	if err != nil {
		return err
	}
	middle := center(target.Rect)
	defer window.doing(fmt.Sprintf("Long press %q", label), target.Rect, true)()
	window.TouchDown(1, middle.x, middle.y)
	window.Wait(longPressHold)
	window.TouchUp(1)
	return nil
}
