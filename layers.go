package ui

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// Layers for controls of a program's own look over everything: Sheet and Toast draw their own card,
// these draw what the program gives them, where it says.

// LayerOptions place what a layer holds in the window.
type LayerOptions struct {
	AlignX, AlignY Align   // Center, Center for a dialog; Center, End for a sheet from the bottom
	Padding        Padding // how far it keeps from the window's edges
	Width          Sizing  // its width; the zero value is as wide as its content
	Height         Sizing  // its height; the zero value is as tall as its content, Grow() all of it
	// Hidden is how much of a modal's dimming is not shown yet, from 0 (all of it, the zero value)
	// to 1 (none): a modal fading in as it opens (see Appear).
	Hidden float32
	// Leave is how it goes when the program stops showing it (leave.go); at once by default.
	Leave Leave
}

// Modal shows content over everything with what is behind it dimmed, as Sheet does, but drawn
// entirely by the program: a dialog, a sheet from the bottom of a phone. While it is shown nothing
// behind it reacts to the mouse or the keyboard, and Tab goes round its own fields. It returns true
// on the step the person asks to close it (Escape, the system's back, or a touch on the dimmed
// backdrop); the program then stops showing it. name tells modals apart.
func (ui *Context) Modal(name string, options LayerOptions, content func()) (dismissed bool) {
	id := ui.idFor(name)
	ui.dropOutsideFocusWhenOpening(id)
	escaped := ui.state.Pressed(input.KeyEscape) && ui.focus == 0 // a field in it takes Escape first
	escaped = escaped || (ui.frontSheet(id) && ui.Back())
	outside := ui.open
	ui.open = nil
	backdrop := Layout{Direction: Vertical, Width: Grow(), Height: Grow(), Padding: options.Padding,
		AlignX: options.AlignX, AlignY: options.AlignY}
	scrim := ui.Theme.Scrim
	scrim[3] *= 1 - clamp(options.Hidden, 0, 1)
	index := ui.openElement(element{id: id, kind: boxElement, layout: backdrop, look: Look{Fill: scrim}})
	ui.roots = append(ui.roots, root{id: id, element: index, sheet: true, options: options, bounds: ui.within, bounded: ui.isWithin})
	ui.nextSheets = append(ui.nextSheets, id)
	ui.Push(name)
	contentID := childID(id, "##content")
	start := len(ui.focusables)
	ui.frameWithID(contentID, Layout{Direction: Vertical, Width: options.Width, Height: options.Height}, Look{}, content)
	ui.sheetFocusStart, ui.sheetFocusEnd = start, len(ui.focusables)
	ui.pointerOn(contentID) // a touch on the content is not one on the backdrop
	ui.Pop()
	touchedOutside := ui.pointerOn(id).clicked
	ui.open = outside
	return escaped || touchedOutside
}

// Overlay shows content over the screen and its panels without dimming or blocking what is behind
// it: a note at the bottom with an action in it, a floating toolbar. Its controls work; elsewhere
// the screen behind does. Modals and sheets cover it, as a dialog or a drawer covers a phone's
// floating buttons; menus and tooltips go over it. The program shows it for as long as it wants.
// name tells overlays apart.
func (ui *Context) Overlay(name string, options LayerOptions, content func()) {
	id := ui.idFor(name)
	outside := ui.open
	ui.open = nil
	layout := Layout{Direction: Vertical, Width: options.Width, Height: options.Height}
	index := ui.openElement(element{id: id, kind: boxElement, layout: layout})
	ui.roots = append(ui.roots, root{id: id, element: index, placed: true, options: options, bounds: ui.within, bounded: ui.isWithin})
	ui.nextOverlays = append(ui.nextOverlays, id)
	ui.Push(name)
	if content != nil {
		content()
	}
	ui.Pop()
	ui.pointerOn(id)
	ui.open = outside
}

// Within lays out the overlays and modals content opens in an area of the window instead of the
// whole window: an interface that a program shows inside a view of its own (a game played inside
// an editor). Popups and tooltips still go anywhere in the window.
func (ui *Context) Within(area draw2d.Rect, content func()) {
	outside, wasWithin := ui.within, ui.isWithin
	ui.within, ui.isWithin = area, true
	content()
	ui.within, ui.isWithin = outside, wasWithin
}

// overlayLayer is above the panels and under the sheets and the popups.
const overlayLayer = 40000

// placeInWindow puts an overlay where its options say, inside the padding of the window, or of the
// area it was opened Within.
func (ui *Context) placeInWindow(placed *root) {
	element := &ui.elements[placed.element]
	padding, bounds := placed.options.Padding, ui.boundsOf(placed)
	left, top := bounds.X+padding.Left, bounds.Y+padding.Top
	right := bounds.X + bounds.Width - padding.Right - element.rect.Width
	bottom := bounds.Y + bounds.Height - padding.Bottom - element.rect.Height
	element.rect.X = alignedBetween(placed.options.AlignX, left, right)
	element.rect.Y = alignedBetween(placed.options.AlignY, top, bottom)
}

func alignedBetween(align Align, start, end float32) float32 {
	switch align {
	case Center:
		return (start + end) / 2
	case End:
		return end
	}
	return start
}
