package ui

import "github.com/akrck02/go-renderer/draw2d"

// On a touch screen a field's text is chosen as on a phone: a finger held on a word selects it,
// with a handle under each end of the selection and a bar of what can be done to it (cut, copy,
// paste, select all); a touch shows a handle under the caret. A finger drags a handle to move its
// end, or the caret: the handles hang under the text, so the finger does not hide what it chooses.
// Typing puts them away. A mouse has none: it selects by dragging, and has its own shortcuts.

// handleRadius is the radius of a handle's round part.
const handleRadius = 11

// The handles a field shows.
type handleKind int

const (
	noHandles handleKind = iota
	caretHandle
	selectionHandles
)

// Which handle a finger drags.
const (
	draggingNone = iota
	draggingStart
	draggingEnd
	draggingCaret
)

// fieldHandles is what a field keeps of its handles between frames.
type fieldHandles struct {
	shown    handleKind
	bar      bool    // the bar of actions over the selection
	dragging int     // the handle a finger holds
	grab     float32 // how far right of the handle's point the finger took it
	grabY    float32 // and how far below it
}

// EditingWords are the labels of the bar over a selection on a touch screen.
type EditingWords struct {
	Cut, Copy, Paste, SelectAll string
}

// EnglishEditingWords are the bar's labels unless a program sets others (SetEditingWords).
var EnglishEditingWords = EditingWords{Cut: "Cut", Copy: "Copy", Paste: "Paste", SelectAll: "Select all"}

// SetEditingWords sets the labels of the bar over a selection, in the person's language.
func (ui *Context) SetEditingWords(words EditingWords) { ui.editingWords = words }

// EditingWords are the labels the bar over a selection shows: the ones set, or English.
func (ui *Context) EditingWords() EditingWords {
	if ui.editingWords == (EditingWords{}) {
		return EnglishEditingWords
	}
	return ui.editingWords
}

// The ids of a field's handles and bar: popups of their own, so they can be touched where they
// hang outside the field.
func startHandleID(field ID) ID { return childID(field, "##start handle") }
func endHandleID(field ID) ID   { return childID(field, "##end handle") }
func caretHandleID(field ID) ID { return childID(field, "##caret handle") }
func selectionBarID(field ID) ID {
	return childID(field, "##selection bar")
}

// onOwnHandles tells whether the pointer is on one of a field's handles or its bar: a press there
// is no press elsewhere, which would end the edit.
func (ui *Context) onOwnHandles(field ID) bool {
	switch ui.hoveredRoot {
	case startHandleID(field), endHandleID(field), caretHandleID(field), selectionBarID(field):
		return true
	}
	return false
}

// followTouches shows and hides a field's handles from what a finger did to the field in this
// step: held on a word, it selects the word; a touch shows the caret's handle, or the selection's
// after a double touch. The mouse, or typing, puts them away.
func (ui *Context) followTouches(id ID, memory *fieldMemory, result interaction, typed bool) {
	handles, editor, press := &memory.handles, &memory.editor, &ui.press
	switch {
	case typed:
		*handles = fieldHandles{}
	case result.pressed && !press.finger:
		*handles = fieldHandles{}
	case result.held && press.finger && press.longNow:
		editor.selectWord(ui.positionUnder(id, memory, press.startX, press.startY))
		memory.dragging = false
		handles.shown, handles.bar = selectionHandles, true
	case result.clicked && press.finger:
		handles.shown, handles.bar = caretHandle, false
		if editor.hasSelection() {
			handles.shown, handles.bar = selectionHandles, true
		}
	}
	if handles.shown == selectionHandles && !editor.hasSelection() {
		handles.shown, handles.bar = caretHandle, false
	}
}

// addHandles adds the field's handles and its bar, and moves the selection or the caret with the
// handle a finger drags.
func (ui *Context) addHandles(id ID, memory *fieldMemory) {
	area, shown := ui.previous[id]
	handles := &memory.handles
	if !shown || handles.shown == noHandles {
		return
	}
	editor := &memory.editor
	start, end := editor.selection()
	if handles.shown == caretHandle {
		x, line := ui.textPoint(area, memory, editor.caret)
		ui.addHandle(id, memory, caretHandleID(id), draggingCaret, x, line)
		return
	}
	startX, startLine := ui.textPoint(area, memory, start)
	endX, endLine := ui.textPoint(area, memory, end)
	ui.addHandle(id, memory, startHandleID(id), draggingStart, startX, startLine)
	ui.addHandle(id, memory, endHandleID(id), draggingEnd, endX, endLine)
	if handles.bar && handles.dragging == draggingNone {
		middle, over := (startX+endX)/2, area.Y
		if startLine != endLine {
			middle = area.X + area.Width/2 // over lines of their own, the bar sits in the middle
		}
		if memory.multiLine { // over the selection's first line, not over the whole area
			over = max(startLine-ui.areaLineHeight(memory), area.Y)
		}
		ui.addSelectionBar(id, memory, over, middle, endLine)
	}
}

// fieldLine is the y of the bottom of a field's line of text, where the handles hang from.
func (ui *Context) fieldLine(area draw2d.Rectangle) float32 {
	metrics := ui.measure("Hg", TextStyle{Size: ui.scaledText(ui.Theme.TextSize)})
	return area.Y + (area.Height+metrics.Height())/2
}

// textPoint is where a position of a field's text is on the screen: its x, and the bottom of its
// line of text, where a handle hangs from.
func (ui *Context) textPoint(area draw2d.Rectangle, memory *fieldMemory, position int) (x, line float32) {
	if !memory.multiLine {
		return area.X + fieldPadding - memory.scroll + ui.textWidth(memory.shown(memory.editor.text[:position]), memory.weight, memory.size), ui.fieldLine(area)
	}
	ui.rewrapArea(memory)
	x, top := ui.areaPoint(memory, position)
	lineHeight := ui.areaLineHeight(memory)
	letters := ui.measure("Hg", TextStyle{Size: ui.scaledText(ui.fieldTextSize(memory.size))}).Height()
	return area.X + fieldPadding + x, area.Y + fieldPadding - memory.lineScroll + top + (lineHeight+letters)/2
}

// addHandle adds one handle with its point at x on the line, in a popup a finger can take it in,
// unless the text there is scrolled out of the field. A finger dragging it moves its end of the
// selection, or the caret, to the position under the handle's point.
func (ui *Context) addHandle(field ID, memory *fieldMemory, id ID, kind int, x, line float32) {
	area := ui.previous[field]
	if x < area.X || x > area.Right() || line < area.Y || line > area.Bottom() {
		return
	}
	box := handleBox(kind, x, line)
	// a popup sits a little under its anchor and inside its padding: the anchor makes up for both
	anchor := draw2d.Rectangle{X: box.X - popupPadding, Y: box.Y - popupPadding - popupSpacing}
	ui.popup(id, anchor, 0, false, Look{}, func() {
		touched := ui.interact(id)
		ui.frameWithID(childID(id, "drop"), Layout{Width: Fixed(box.Width), Height: Fixed(box.Height)}, Look{}, func() {
			ui.Paint(Layout{Width: Grow(), Height: Grow()}, func(drawn draw2d.Rectangle, drawing *draw2d.List) {
				paintHandle(drawing, kind, drawn.X+(x-box.X), drawn.Y+(line-box.Y), ui.Theme.Primary)
			})
		})
		ui.Describe(Description{Hidden: true}) // the field is read, not its handles
		ui.dragHandle(field, memory, kind, touched, x, line)
	})
}

// dragHandle follows a finger on a handle: it takes the handle where it came down, moves its end
// to the position under the point as it goes, and shows the bar again once it lets go.
func (ui *Context) dragHandle(field ID, memory *fieldMemory, kind int, touched interaction, x, line float32) {
	handles, editor := &memory.handles, &memory.editor
	pointer, pointerY := float32(ui.state.CursorX), float32(ui.state.CursorY)
	if touched.pressed {
		handles.dragging, handles.grab, handles.grabY = kind, pointer-x, pointerY-line
	}
	if handles.dragging != kind {
		return
	}
	if !touched.held {
		handles.dragging, handles.bar = draggingNone, kind != draggingCaret
		return
	}
	position := ui.positionUnder(field, memory, pointer-handles.grab, pointerY-handles.grabY-1) // just above the point: on its line
	start, end := editor.selection()
	switch kind {
	case draggingCaret:
		editor.moveTo(position, false)
	case draggingStart: // the ends never cross: at least one character stays chosen
		editor.anchor, editor.caret = end, min(position, end-1)
	case draggingEnd:
		editor.anchor, editor.caret = start, max(position, start+1)
	}
	memory.blinkStart = ui.now
}

// handleBox is the square a finger takes a handle in, around its drop: wider than it is drawn.
func handleBox(kind int, x, line float32) draw2d.Rectangle {
	side := float32(3 * handleRadius)
	switch kind {
	case draggingStart:
		return draw2d.Rectangle{X: x - side + handleRadius/2, Y: line, Width: side, Height: side}
	case draggingEnd:
		return draw2d.Rectangle{X: x - handleRadius/2, Y: line, Width: side, Height: side}
	}
	return draw2d.Rectangle{X: x - side/2, Y: line, Width: side, Height: side}
}

// paintHandle draws a drop with its point at x on the line: leaning left under the start of a
// selection, right under its end, straight under the caret.
func paintHandle(drawing *draw2d.List, kind int, x, line float32, colour draw2d.Color) {
	radius := float32(handleRadius)
	disc := func(centreX, centreY float32) {
		drawing.Rectangle(draw2d.Rectangle{X: centreX - radius, Y: centreY - radius, Width: 2 * radius, Height: 2 * radius},
			radius, colour, draw2d.Transparent, 0)
	}
	switch kind {
	case draggingStart:
		disc(x-radius, line+radius)
		drawing.Rectangle(draw2d.Rectangle{X: x - radius, Y: line, Width: radius, Height: radius}, 0, colour, draw2d.Transparent, 0)
	case draggingEnd:
		disc(x+radius, line+radius)
		drawing.Rectangle(draw2d.Rectangle{X: x, Y: line, Width: radius, Height: radius}, 0, colour, draw2d.Transparent, 0)
	default:
		disc(x, line+1.3*radius)
		stem := 1.1 * radius
		drawing.Rectangle(draw2d.Rectangle{X: x - stem/2, Y: line, Width: stem, Height: 1.3 * radius}, stem*0.3, colour, draw2d.Transparent, 0)
	}
}

// addSelectionBar adds the bar of what can be done to the selection, centred over it above a y
// (the field's top, or the selection's first line in a text area), or under the handles hanging
// from line when there is no room above.
func (ui *Context) addSelectionBar(field ID, memory *fieldMemory, over, middle, line float32) {
	barHeight := ui.Theme.ControlHeight + 2*popupPadding
	top := over - barHeight - 8
	if top < ui.insets.Top+4 {
		top = line + 3*handleRadius + 4
	}
	anchor := draw2d.Rectangle{X: middle, Y: top - popupSpacing}
	ui.centredPopup(selectionBarID(field), anchor, ui.popupLook(), func() {
		ui.Push("##selection bar")
		defer ui.Pop()
		ui.Box(Layout{Gap: 2}, func() {
			ui.selectionActions(memory)
		})
	})
}

// selectionActions are the bar's buttons: cut and copy a selection (not a secret one's), paste
// what the clipboard holds, select the whole text when it is not.
func (ui *Context) selectionActions(memory *fieldMemory) {
	editor, words := &memory.editor, ui.EditingWords()
	chosen := editor.hasSelection() && !memory.secret
	done := false
	if chosen && ui.button(words.Cut, words.Cut, "", flatButton) {
		ui.state.SetClipboardText(editor.selectedText())
		editor.deleteTo(editor.caret)
		done = true
	}
	if chosen && ui.button(words.Copy, words.Copy, "", flatButton) {
		ui.state.SetClipboardText(editor.selectedText())
		editor.moveTo(max(editor.caret, editor.anchor), false)
		done = true
	}
	if clipboard := ui.state.ClipboardText(); clipboard != "" && ui.button(words.Paste, words.Paste, "", flatButton) {
		editor.insert(clipboard, otherEdit)
		done = true
	}
	start, end := editor.selection()
	if (start > 0 || end < len(editor.text)) && ui.button(words.SelectAll, words.SelectAll, "", flatButton) {
		editor.selectAll()
		memory.handles.shown = selectionHandles
	}
	if done {
		memory.handles = fieldHandles{}
	}
	memory.blinkStart = ui.now
}
