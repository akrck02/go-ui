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
	ui.keyboardGiven, ui.keyboardByTab, ui.keyboardFocus, ui.resumeStop = false, false, 0, 0
	ui.hasKeyboard = ui.keyboardWanted()
}

// keyboardWanted tells whether something gives the interface the keyboard: its mode, the program,
// an open dialog or sheet, a field, list or menu taking the keys, or a Tab that moved the focus
// while the interface had them (it keeps them until Escape or a click).
func (ui *Context) keyboardWanted() bool {
	return ui.KeyboardNavigation == Always || ui.keyboardGiven || ui.keyboardByTab || len(ui.sheets) > 0 || ui.focus != 0
}

// HasKeyboard tells whether the interface reads Tab, Enter and Space in this frame. It stays true
// for the rest of a frame in which the interface used one of those keys, even if it gave the
// keyboard back on it: Space on a Resume button must not also reach the game.
func (ui *Context) HasKeyboard() bool { return ui.hasKeyboard || ui.keyboardUsed }

// FocusedID is the control that has the focus after the last frame: the field being typed in, or
// the control Tab reached; 0 when none has it. A control clicked with the pointer has it only if it
// is a field.
func (ui *Context) FocusedID() ID {
	if ui.focus != 0 {
		return ui.focus
	}
	return ui.keyboardFocus
}

// beginFocus works out, at the start of a frame, whether the interface has the keyboard, and
// follows Tab when no field is taking the keys (a field follows its own).
func (ui *Context) beginFocus() {
	if ui.state.ButtonPressed(input.MouseLeft) {
		// the pointer moves the focus by itself and shows no ring; Tab goes on from what it pressed
		ui.keyboardFocus, ui.keyboardByTab, ui.resumeStop = 0, false, 0
	}
	ui.hasKeyboard = ui.keyboardWanted()
	ui.keyboardFocusSeen, ui.focusAtBegin, ui.heldControl, ui.keyboardUsed = false, ui.focus, 0, false
	if !ui.hasKeyboard {
		ui.keyboardFocus = 0
	}
	if ui.hasKeyboard && ui.focus == 0 && ui.state.Pressed(input.KeyTab) {
		from := ui.keyboardFocus
		if from == 0 {
			from = ui.clickedStop
		}
		ui.tabFrom, ui.tabBack, ui.tabPressed = from, ui.state.Down(input.KeyShift), true
	}
}

// joinTabOrder adds a control to the Tab order of this frame. Controls only join while the
// interface has the keyboard, so a game's frame does no work for them.
func (ui *Context) joinTabOrder(id ID) {
	if id == ui.keyboardFocus {
		ui.keyboardFocusSeen = true
	}
	if !ui.hasKeyboard || ui.outOfTabOrder > 0 {
		return
	}
	if id == ui.keyboardFocus {
		ui.keyboardFocusPosition = len(ui.focusables)
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
	pressed := ui.state.Pressed(input.KeyEnter) || ui.state.Pressed(input.KeySpace)
	ui.keyboardUsed = ui.keyboardUsed || pressed
	return pressed
}

// endFocus forgets a focused control that was not built, and gives the keyboard back on an Escape
// that nothing else took: no field, list or menu had the keys and no dialog was open.
func (ui *Context) endFocus() {
	if !ui.keyboardFocusSeen && ui.keyboardFocus != 0 {
		// the control went away (a row deleted): Tab goes on from where it was
		ui.keyboardFocus, ui.resumeStop = 0, ui.keyboardFocusPosition+1
	}
	if ui.state.ButtonPressed(input.MouseLeft) && ui.isTabStop(ui.active) {
		ui.clickedStop = ui.active
	}
	escaped := ui.state.Pressed(input.KeyEscape) && ui.focusAtBegin == 0
	if escaped && len(ui.sheets) == 0 && len(ui.nextSheets) == 0 {
		ui.ReleaseKeyboard()
		ui.keyboardUsed = true
	}
}

// moveFocusByTab gives the focus to the stop after (or, with Shift, before) the one where Tab was
// pressed, going round at the ends; a first Tab goes to the first stop (or the last). A field
// selects its whole text when it takes the focus.
func (ui *Context) moveFocusByTab() {
	from, pressed, resume := ui.tabFrom, ui.tabPressed || ui.tabFrom != 0, ui.resumeStop
	ui.tabFrom, ui.tabPressed = 0, false
	focusables, offset := ui.focusables, 0
	if len(ui.nextSheets) > 0 { // a sheet is open: Tab goes round its own controls only
		focusables, offset = focusables[ui.sheetFocusStart:ui.sheetFocusEnd], ui.sheetFocusStart
	}
	if !pressed || len(focusables) == 0 {
		return
	}
	ui.resumeStop = 0
	step := 1
	if ui.tabBack {
		step = -1
	}
	index := -1
	switch {
	case resume > 0 && !ui.tabBack:
		index = clampIndex(resume-1-offset, len(focusables)) - 1 // the stop now where the lost one was
	case resume > 0:
		index = clampIndex(resume-1-offset, len(focusables))
	case ui.tabBack:
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
	ui.keyboardByTab = true
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
	ring := draw2d.Rectangle{X: area.X - width, Y: area.Y - width, Width: area.Width + 2*width, Height: area.Height + 2*width}
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
func (ui *Context) scaleCommands(first int, area draw2d.Rectangle, scale float32) {
	centreX, centreY := area.X+area.Width/2, area.Y+area.Height/2
	around := func(rect draw2d.Rectangle) draw2d.Rectangle {
		return draw2d.Rectangle{X: centreX + (rect.X-centreX)*scale, Y: centreY + (rect.Y-centreY)*scale,
			Width: rect.Width * scale, Height: rect.Height * scale}
	}
	commands := ui.drawing.Commands[first:]
	for index := range commands {
		command := &commands[index]
		command.Area, command.Radius, command.BorderWidth = around(command.Area), command.Radius*scale, command.BorderWidth*scale
		switch command.Kind {
		case draw2d.TextKind, draw2d.IconKind:
			command.Size, command.Baseline = command.Size*scale, centreY+(command.Baseline-centreY)*scale
		case draw2d.ShadowKind:
			command.Blur *= scale
		}
		if command.HasClip && area.Intersect(command.Clip) == command.Clip {
			command.Clip = around(command.Clip) // a clip of the control's own; one from a list around it stays
		}
	}
}

// clampIndex keeps a position inside a list of length stops.
func clampIndex(position, length int) int { return min(max(position, 0), length-1) }

// isTabStop tells whether a control is in this frame's Tab order.
func (ui *Context) isTabStop(id ID) bool {
	for _, stop := range ui.focusables {
		if stop.id == id {
			return id != 0
		}
	}
	return false
}
