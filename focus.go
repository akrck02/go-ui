package ui

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// defaultFocusWidth is go-ui's own focus ring, in layout pixels.
const defaultFocusWidth = 3

// KeyboardNavigation says when the interface reads Tab, Enter and Space to move between controls
// and press them, and when those keys belong to the program (a game, a 3D view).
type KeyboardNavigation int

const (
	// WhenGiven, the default: the interface has the keyboard only while a field is typed in, a
	// dialog, sheet or menu is open, or the program gave it with TakeKeyboard. Escape gives it
	// back. A click on a button does not take it, so Space in a game never presses the last button
	// clicked.
	WhenGiven KeyboardNavigation = iota
	// Always: the program is only an interface (a register, a settings app), so Tab always moves
	// between its controls.
	Always
)

// focusable is a stop of the Tab order: a control, or a field that takes the typed keys.
type focusable struct {
	id    ID
	field bool
}

// TakeKeyboard gives the interface the keyboard, as a game does when it opens its own menu: Tab
// then moves between the controls until Escape or ReleaseKeyboard gives the keys back.
func (ui *Context) TakeKeyboard() { ui.keyboardGiven = true }

// ReleaseKeyboard gives the keyboard back to the program and takes the focus off the controls.
func (ui *Context) ReleaseKeyboard() {
	ui.keyboardGiven, ui.keyboardFocus = false, 0
	ui.hasKeyboard = ui.keyboardWanted()
}

// keyboardWanted tells whether something gives the interface the keyboard: its mode, the program,
// an open dialog or sheet, or a field, list or menu taking the keys.
func (ui *Context) keyboardWanted() bool {
	return ui.KeyboardNavigation == Always || ui.keyboardGiven || len(ui.sheets) > 0 || ui.focus != 0
}

// HasKeyboard tells whether the interface reads Tab, Enter and Space in this frame.
func (ui *Context) HasKeyboard() bool { return ui.hasKeyboard }

// beginFocus works out, at the start of a frame, whether the interface has the keyboard, and
// follows Tab when no field is taking the keys (a field follows its own).
func (ui *Context) beginFocus() {
	ui.hasKeyboard = ui.keyboardWanted()
	ui.keyboardFocusSeen, ui.focusAtBegin, ui.heldControl = false, ui.focus, 0
	if !ui.hasKeyboard || ui.state.ButtonPressed(input.MouseLeft) {
		ui.keyboardFocus = 0 // the pointer moves the focus by itself, and shows no ring
	}
	if ui.hasKeyboard && ui.focus == 0 && ui.state.Pressed(input.KeyTab) {
		ui.tabFrom, ui.tabBack, ui.tabPressed = ui.keyboardFocus, ui.state.Down(input.KeyShift), true
	}
}

// joinTabOrder adds a control to the Tab order of this frame. Controls only join while the
// interface has the keyboard, so a game's frame does no work for them.
func (ui *Context) joinTabOrder(id ID) {
	if id == ui.active {
		ui.heldControl = id
	}
	if id == ui.keyboardFocus {
		ui.keyboardFocusSeen = true
	}
	if !ui.hasKeyboard || ui.outOfTabOrder > 0 {
		return
	}
	ui.focusables = append(ui.focusables, focusable{id: id})
}

// joinTabOrderAsField adds a field, which takes the typed keys when it has the focus.
func (ui *Context) joinTabOrderAsField(id ID) {
	ui.focusables = append(ui.focusables, focusable{id: id, field: true})
}

// withoutTab builds controls that Tab skips: the step buttons of a number field, whose number
// can be typed.
func (ui *Context) withoutTab(content func()) {
	ui.outOfTabOrder++
	content()
	ui.outOfTabOrder--
}

// keyboardPresses tells whether Enter or Space presses the control with this id in this frame.
func (ui *Context) keyboardPresses(id ID) bool {
	if id == 0 || id != ui.keyboardFocus || ui.focus != 0 {
		return false
	}
	return ui.state.Pressed(input.KeyEnter) || ui.state.Pressed(input.KeySpace)
}

// endFocus forgets a focused control that was not built, and gives the keyboard back on an Escape
// that nothing else took: no field, list or menu had the keys and no dialog was open.
func (ui *Context) endFocus() {
	if !ui.keyboardFocusSeen {
		ui.keyboardFocus = 0
	}
	escaped := ui.state.Pressed(input.KeyEscape) && ui.focusAtBegin == 0
	if escaped && len(ui.sheets) == 0 && len(ui.nextSheets) == 0 {
		ui.ReleaseKeyboard()
	}
}

// moveFocusByTab gives the focus to the stop after (or, with Shift, before) the one where Tab was
// pressed, going round at the ends; a first Tab goes to the first stop (or the last). A field
// selects its whole text when it takes the focus.
func (ui *Context) moveFocusByTab() {
	from, pressed := ui.tabFrom, ui.tabPressed || ui.tabFrom != 0
	ui.tabFrom, ui.tabPressed = 0, false
	focusables := ui.focusables
	if len(ui.nextSheets) > 0 { // a sheet is open: Tab goes round its own controls only
		focusables = focusables[ui.sheetFocusStart:ui.sheetFocusEnd]
	}
	if !pressed || len(focusables) == 0 {
		return
	}
	step := 1
	if ui.tabBack {
		step = -1
	}
	index := -1
	if ui.tabBack {
		index = len(focusables)
	}
	for position, stop := range focusables {
		if stop.id == from {
			index = position
			break
		}
	}
	next := focusables[(index+step+len(focusables))%len(focusables)]
	ui.RequestFrame(0) // the focus moved after this frame was painted: draw it where it is now
	if next.field {
		ui.focus, ui.focusByKeyboard, ui.keyboardFocus = next.id, true, 0
		return
	}
	ui.focus, ui.keyboardFocus = 0, next.id
}

// paintFocusRing rings the control Tab reached, over its content: on its edge with no gap, or, on a
// primary button, inside it in the colour of its text. A control reached by the pointer gets none.
func (ui *Context) paintFocusRing(element *element) {
	if element.id == 0 || element.id != ui.keyboardFocus || !ui.hasKeyboard {
		return
	}
	theme := ui.Theme
	width := theme.focusWidth()
	if element.filledWithPrimary {
		ui.drawing.Rectangle(element.rect, element.look.Radius, draw2d.Transparent, theme.OnPrimary, width)
		return
	}
	area := element.rect
	ring := draw2d.Rect{X: area.X - width, Y: area.Y - width, Width: area.Width + 2*width, Height: area.Height + 2*width}
	ui.drawing.Rectangle(ring, element.look.Radius+width, draw2d.Transparent, theme.FocusRing, width)
}

// shrinks tells whether a box is a control held down under the pointer that the theme shrinks.
func (ui *Context) shrinks(element *element) bool {
	scale := ui.Theme.PressScale
	held := element.id != 0 && element.id == ui.heldControl && element.id == ui.hot
	return held && scale > 0 && scale != 1 && !ui.LessMotion
}

// scaleCommands draws what was painted since first smaller (or larger) around the centre of area:
// only the drawing changes, so the layout and where the pointer reaches stay where they were.
func (ui *Context) scaleCommands(first int, area draw2d.Rect, scale float32) {
	centreX, centreY := area.X+area.Width/2, area.Y+area.Height/2
	around := func(rect draw2d.Rect) draw2d.Rect {
		return draw2d.Rect{X: centreX + (rect.X-centreX)*scale, Y: centreY + (rect.Y-centreY)*scale,
			Width: rect.Width * scale, Height: rect.Height * scale}
	}
	commands := ui.drawing.Commands[first:]
	for index := range commands {
		command := &commands[index]
		command.Area, command.Radius, command.BorderWidth = around(command.Area), command.Radius*scale, command.BorderWidth*scale
		command.Size, command.Baseline = command.Size*scale, centreY+(command.Baseline-centreY)*scale
		if command.HasClip && area.Intersect(command.Clip) == command.Clip {
			command.Clip = around(command.Clip) // a clip of the control's own; one from a list around it stays
		}
	}
}
