package ui

import (
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// Gestures of one finger on a touch screen. The interface follows the finger from where it comes
// down; once it has moved a few pixels (dragSlop), the way it went first, up and down or sideways,
// decides which part of the interface takes it: the innermost scroll area for up and down, a swipe
// box for sideways. The part that takes it takes the pointer from the control the finger came down
// on, which is then not touched. A finger that hardly moves stays a touch.

// dragSlop is how far a finger moves before a touch becomes a drag.
const dragSlop = 8

// finger is the one finger followed, from where it came down until it lifts.
type finger struct {
	down                 bool
	released             bool // lifted in this frame
	startX, startY       float32
	x, y                 float32
	previousX, previousY float32   // where it was in the last frame
	moved                bool      // past dragSlop: a drag, not a touch
	axis                 Direction // the way it went first, once moved
	owner                ID        // the part of the interface that took it
	velocityX, velocityY float32   // pixels a second, smoothed
	lastTime             time.Time
}

// trackFinger follows the finger in this frame's input.
func (ui *Context) trackFinger() {
	state, finger := ui.state, &ui.finger
	x, y := float32(state.CursorX), float32(state.CursorY)
	finger.released = false
	if state.ButtonPressed(input.MouseLeft) && len(state.Touches()) == 1 {
		*finger = ui.fingerAt(x, y)
		return
	}
	if !finger.down {
		return
	}
	if !state.Button(input.MouseLeft) || len(state.Touches()) > 1 {
		finger.down, finger.released = false, true
		finger.previousX, finger.previousY = finger.x, finger.y
		return
	}
	seconds := float32(ui.now.Sub(finger.lastTime).Seconds())
	finger.previousX, finger.previousY, finger.x, finger.y, finger.lastTime = finger.x, finger.y, x, y, ui.now
	if seconds > 0 {
		// smoothed: one step can be uneven
		finger.velocityX = 0.7*finger.velocityX + 0.3*(x-finger.previousX)/seconds
		finger.velocityY = 0.7*finger.velocityY + 0.3*(y-finger.previousY)/seconds
	}
	if !finger.moved && max(abs32(x-finger.startX), abs32(y-finger.startY)) >= dragSlop {
		finger.moved, finger.axis = true, Vertical
		if abs32(x-finger.startX) > abs32(y-finger.startY) {
			finger.axis = Horizontal
		}
	}
}

func (ui *Context) fingerAt(x, y float32) finger {
	return finger{down: true, startX: x, startY: y, x: x, y: y, previousX: x, previousY: y, lastTime: ui.now}
}

// takeFinger gives the finger to a part of the interface that follows it along an axis: when it
// went that way first, from inside the part's area of the last frame. It tells whether the part
// has it, in the frame it lets go too.
func (ui *Context) takeFinger(id ID, axis Direction) bool {
	finger := &ui.finger
	if finger.owner == id {
		if finger.down {
			ui.active = id // the control the finger came down on is not touched after all
		}
		return finger.down || finger.released
	}
	if finger.owner != 0 || !finger.down || !finger.moved || finger.axis != axis {
		return false
	}
	area, shown := ui.previous[id]
	if !shown || !area.Contains(finger.startX, finger.startY) || ui.hoveredRoot != ui.currentRoot() {
		return false
	}
	finger.owner, ui.active = id, id
	return true
}

// Swipe is what a finger does to a swipe box in a frame.
type Swipe struct {
	Offset   float32 // how far the finger has moved along the box's axis since it took the box
	Velocity float32 // pixels a second along the axis, as it last moved
	Held     bool    // a finger is dragging the box
	Released bool    // the finger let go in this frame; Offset and Velocity are its last
	// Area is where the box was laid out in the last frame (zero in its first): its width is a
	// page's, for content that moves a page at a time.
	Area draw2d.Rect
}

// SwipeBox lays content out in a box a finger drags along an axis (sideways, for pages, cards in a
// row, a row swiped away): content is given what the finger does in this frame, and places itself
// by it. The box clips what goes past its edges. A swipe box takes a finger before what is inside
// it, along its own axis only, so a list inside pages still scrolls up and down.
func (ui *Context) SwipeBox(name string, layout Layout, axis Direction, content func(swipe Swipe)) {
	id := ui.idFor(name)
	var swipe Swipe
	if ui.takeFinger(id, axis) {
		finger := &ui.finger
		along, velocity := finger.x-finger.startX, finger.velocityX
		if axis == Vertical {
			along, velocity = finger.y-finger.startY, finger.velocityY
		}
		// from where the drag began, past the slop, so the content does not jump
		if along > 0 {
			along = max(along-dragSlop, 0)
		} else {
			along = min(along+dragSlop, 0)
		}
		swipe = Swipe{Offset: along, Velocity: velocity, Held: finger.down, Released: finger.released}
	}
	swipe.Area = ui.previous[id]
	ui.frameWithID(id, layout, Look{Clip: true}, func() { content(swipe) })
}

// Pointer is where the pointer is, in the window's layout pixels: a finger on a touch screen. For a
// control of a program's own that takes a position (a clock's dial, a colour wheel), read while it
// is pressed.
func (ui *Context) Pointer() (x, y float32) {
	if ui.state == nil {
		return 0, 0
	}
	return float32(ui.state.CursorX), float32(ui.state.CursorY)
}

// Pressing tells whether the pointer is pressed now: a finger is on a touch screen, the mouse's
// left button is down. With Pointer, for a control of a program's own that follows a drag.
func (ui *Context) Pressing() bool {
	return ui.state != nil && ui.state.Button(input.MouseLeft)
}
