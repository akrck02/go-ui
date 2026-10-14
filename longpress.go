package ui

import (
	"time"

	"github.com/akrck02/go-renderer/input"
)

// A long press is how a finger asks a control for more than a touch: to choose a row, to open
// what can be done to it. The finger stays still on the control for longPressTime; with a mouse,
// which has a button of its own for that, the right button does the same. A touch that became a
// long press is no click when it lifts.

// longPressTime is how long a finger stays still before a touch becomes a long press, as on Android.
const longPressTime = 500 * time.Millisecond

// press is the pointer held down, followed for a long press.
type press struct {
	since          time.Time
	startX, startY float32
	finger         bool // a finger on a touch screen, not the mouse
	moved          bool // went past dragSlop: a drag, not a press
	long           bool // became a long press, so its release is no click
	longNow        bool // became a long press in this frame
}

// followPress follows the pointer held down, and makes a finger that stays still long enough a
// long press. A window that draws on demand is asked for the frame in which it becomes one.
func (ui *Context) followPress() {
	state, press := ui.state, &ui.press
	press.longNow = false
	ui.longPressed = 0
	x, y := float32(state.CursorX), float32(state.CursorY)
	if state.ButtonPressed(input.MouseLeft) {
		*press = pressAt(ui.now, x, y, len(state.Touches()) == 1) // and on, to ask for its frame
	}
	if !state.Button(input.MouseLeft) || !press.finger || press.moved || press.long {
		return
	}
	if max(abs32(x-press.startX), abs32(y-press.startY)) >= dragSlop {
		press.moved = true
		return
	}
	if waited := ui.now.Sub(press.since); waited < longPressTime {
		state.RequestFrame(longPressTime - waited)
		return
	}
	press.long, press.longNow = true, true
}

func pressAt(now time.Time, x, y float32, finger bool) press {
	return press{since: now, startX: x, startY: y, finger: finger}
}

// noteLongPress gives this frame's long press to the control that has the pointer: the one held
// when the finger's time comes, or the one under the mouse when its right button goes down.
func (ui *Context) noteLongPress(id ID, result interaction) {
	if ui.active == id && result.held && ui.press.longNow {
		ui.longPressed = id
	}
	if result.hovered && ui.state.ButtonPressed(input.MouseRight) {
		ui.longPressed = id
	}
}

// LongPressed tells whether the control added just before it was long-pressed in this frame: a
// finger held still on it for half a second, or the right mouse button pressed on it. It is true
// in one frame only; a program opens what it shows then, and keeps it open itself.
func (ui *Context) LongPressed() bool {
	return ui.longPressed != 0 && ui.longPressed == ui.interacted
}
