package ui

import (
	"time"

	"github.com/akrck02/go-renderer/input"
)

// Insets are how far in from each edge of the window the system draws over it, in layout pixels:
// a phone's status bar at the top and its gesture bar at the bottom. A window that covers the
// whole screen paints its own bars' backgrounds under them, and keeps its controls clear of them.
// On a desktop and in a browser they are zero.
type Insets struct {
	Top, Bottom, Left, Right float32
}

// SetInsets tells the interface what the system draws over the window; the window calls it.
func (ui *Context) SetInsets(insets Insets) { ui.insets = insets }

// Insets returns what the system draws over the window's edges.
func (ui *Context) Insets() Insets { return ui.insets }

// Size returns the window's size in this frame, in layout pixels: a program chooses a layout by it
// (a list alone on a phone, a list and its details side by side on a tablet).
func (ui *Context) Size() (width, height float32) { return ui.width, ui.height }

// RequestFrame asks the window for another frame within a delay, even if no input arrives: for a
// program that animates something of its own (a clock, a replay). A window that draws on demand
// otherwise sleeps until the next touch or key.
func (ui *Context) RequestFrame(delay time.Duration) {
	if ui.state != nil {
		ui.state.RequestFrame(delay)
	}
}

// Touches returns the fingers on a touch screen in this frame, in layout pixels, all of them:
// those on controls too. A control of a program's own that follows fingers (an on-screen stick, a
// gesture) reads them here; the first finger is also the pointer that clicks. None on a desktop.
func (ui *Context) Touches() []input.Touch {
	if ui.state == nil {
		return nil
	}
	return ui.state.Touches()
}
