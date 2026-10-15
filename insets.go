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

// Copy puts a text on the system's clipboard (a link, a code, what a row holds), to paste in
// another application. Without a clipboard (tests) it does nothing.
func (ui *Context) Copy(text string) {
	if ui.state != nil {
		ui.state.SetClipboardText(text)
	}
}

// ClipboardText is the text on the system's clipboard, "" when there is none.
func (ui *Context) ClipboardText() string {
	if ui.state == nil {
		return ""
	}
	return ui.state.ClipboardText()
}

// SetTextScale sets how much larger than its size each text is drawn: the person's text size, as
// a phone's settings give it (1.3 for a larger text). Only text grows; icons, paddings and fixed
// boxes keep their size, as on Android. The window sets it in every frame.
func (ui *Context) SetTextScale(scale float32) { ui.textScale = scale }

// TextScale is the text size SetTextScale gave, 1 by default: a control of a program's own whose
// height holds text grows by it.
func (ui *Context) TextScale() float32 {
	if ui.textScale <= 0 {
		return 1
	}
	return ui.textScale
}

func (ui *Context) scaledText(size float32) float32 { return size * ui.TextScale() }

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
