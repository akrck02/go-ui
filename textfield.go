package ui

import (
	"runtime"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// TextFieldOptions shape a text field.
type TextFieldOptions struct {
	Placeholder string // shown, quiet, while the field is empty
	MaxLength   int    // the most characters it takes; 0 = no limit
	Width       Sizing // the zero value grows across the row or column
	// Weight is the face the text is written in: draw2d.Mono for a code or a key.
	Weight draw2d.Weight
}

// fieldEvent is how editing a field ended during a step, if it did.
type fieldEvent int

const (
	fieldEditing   fieldEvent = iota // nothing ended
	fieldSubmitted                   // Enter
	fieldCancelled                   // Escape: the text went back to what it was
	fieldLeft                        // the focus went elsewhere: a click outside, Tab
)

// fieldMemory is what a text field keeps between frames.
type fieldMemory struct {
	editor        editor
	hadFocus      bool
	original      string    // the text when the field took the focus, for Escape
	scroll        float32   // how far the text is moved left to keep the caret in sight
	blinkStart    time.Time // when the caret last moved: it shows steadily while typing
	dragging      bool      // a press in the field is selecting text
	measured      fieldMeasures
	measuredKey   fieldMeasureKey
	measuredValid bool
	paint         fieldPaint
	weight        draw2d.Weight // the face of its text, from its options
}

// fieldPaint is what painting a field needs, worked out while the field is built; offsets are
// from the start of the text.
type fieldPaint struct {
	text, placeholder            string
	focused, caretShown          bool
	caretX, scroll               float32
	selectionStart, selectionEnd float32
	weight                       draw2d.Weight
}

// fieldPadding is the space between a field's edges and its text.
const fieldPadding = 10

// caretBlink is how long the caret stays shown, then hidden.
const caretBlink = 530 * time.Millisecond

// TextField adds a one-line field to type text in. Click to place the caret, drag or Shift-click
// to select, double-click for a word and triple-click for everything. The usual keys work: arrows
// (with Shift to select, with Alt on a Mac or Ctrl elsewhere to jump words), Home, End, Backspace,
// Delete, and with Cmd on a Mac or Ctrl elsewhere: A, C, X, V, Z (Shift+Z or Y to redo). Enter
// or a click elsewhere ends the edit, Escape puts the text back as it was, and Tab goes to the
// next field. It returns true on the steps the text changes.
func (ui *Context) TextField(name string, text *string) bool {
	return ui.TextFieldWith(name, text, TextFieldOptions{})
}

// TextFieldWith is TextField with a placeholder, a length limit or a width.
func (ui *Context) TextFieldWith(name string, text *string, options TextFieldOptions) bool {
	before := *text
	ui.editField(ui.idFor(name), text, options)
	return *text != before
}

// editField runs a text field: focus, mouse, keys, and its box. It writes the text as it is
// typed and tells how the edit ended, if it did during this step.
func (ui *Context) editField(id ID, text *string, options TextFieldOptions) fieldEvent {
	memory := remember(ui, id, func() *fieldMemory { return &fieldMemory{} })
	memory.editor.limit, memory.weight = options.MaxLength, options.Weight
	ui.focusables = append(ui.focusables, id)
	result := ui.interact(id)
	if result.pressed && ui.focus != id {
		ui.focus = id
	}
	event := fieldEditing
	focused := ui.focus == id
	switch {
	case focused && !memory.hadFocus:
		memory.editor.load(*text)
		memory.original, memory.scroll = *text, 0
		if ui.focusByKeyboard {
			memory.editor.selectAll()
			ui.focusByKeyboard = false
		}
	case !focused && memory.hadFocus:
		event = fieldLeft
	}
	if focused {
		if *text != memory.editor.String() {
			memory.editor.load(*text) // the program changed the text while it was being edited
		}
		event = ui.followFieldMouse(id, memory, result)
		if event == fieldEditing {
			event = ui.followFieldKeys(id, memory)
		}
		switch event {
		case fieldEditing:
			*text = memory.editor.String()
		case fieldCancelled:
			*text = memory.original
		default:
			*text = memory.editor.String()
		}
		focused = ui.focus == id
	}
	if focused {
		ui.focusSeen = true
	}
	memory.hadFocus = focused
	ui.addField(id, memory, *text, options, result, focused)
	return event
}

// followFieldMouse places the caret and selects with the mouse, and ends the edit on a press
// elsewhere.
func (ui *Context) followFieldMouse(id ID, memory *fieldMemory, result interaction) fieldEvent {
	state := ui.state
	editor := &memory.editor
	if state.ButtonPressed(input.MouseLeft) && !result.pressed {
		ui.focus = 0
		return fieldLeft
	}
	if !result.pressed && !(result.held && memory.dragging) {
		memory.dragging = false
		return fieldEditing
	}
	position := ui.fieldPositionAt(id, memory, float32(state.CursorX))
	switch {
	case result.pressed:
		memory.dragging = false
		switch state.Clicks(input.MouseLeft) {
		case 0, 1:
			editor.moveTo(position, state.Down(input.KeyShift))
			memory.dragging = true
		case 2:
			editor.selectWord(position)
		default:
			editor.selectAll()
		}
		memory.blinkStart = ui.now
	case result.held && memory.dragging:
		if position != editor.caret {
			editor.moveTo(position, true)
			memory.blinkStart = ui.now
		}
	default:
		memory.dragging = false
	}
	return fieldEditing
}

// fieldPositionAt is the caret position nearest to an x on the screen, from where the field was in
// the last frame.
func (ui *Context) fieldPositionAt(id ID, memory *fieldMemory, x float32) int {
	area, shown := ui.previous[id]
	if !shown {
		return memory.editor.caret
	}
	offset := x - (area.X + fieldPadding) + memory.scroll
	runes := memory.editor.text
	low, high := 0, len(runes)
	for low < high { // the first position whose caret is right of the pointer
		middle := (low + high) / 2
		if ui.textWidth(string(runes[:middle+1]), memory.weight) < offset {
			low = middle + 1
		} else {
			high = middle
		}
	}
	if low > 0 && low <= len(runes) {
		left, right := ui.textWidth(string(runes[:low-1]), memory.weight), ui.textWidth(string(runes[:low]), memory.weight)
		if offset-left < right-offset {
			return low - 1
		}
	}
	if low < len(runes) {
		left, right := ui.textWidth(string(runes[:low]), memory.weight), ui.textWidth(string(runes[:low+1]), memory.weight)
		if offset-left > right-offset {
			return low + 1
		}
	}
	return low
}

func (ui *Context) textWidth(text string, weight draw2d.Weight) float32 {
	return ui.measure(text, TextStyle{Size: ui.Theme.TextSize, Weight: weight}).Width
}

// shortcutKey is the key held for copy, paste and undo: Cmd on a Mac, Ctrl elsewhere. wordKey is
// the one held to jump words: Alt on a Mac, Ctrl elsewhere.
var shortcutKey, wordKey = input.KeyControl, input.KeyControl

func init() {
	if runtime.GOOS == "darwin" {
		shortcutKey, wordKey = input.KeySuper, input.KeyAlt
	}
}

// followFieldKeys types, moves, selects, copies, pastes and undoes with the keyboard.
func (ui *Context) followFieldKeys(id ID, memory *fieldMemory) fieldEvent {
	state := ui.state
	editor := &memory.editor
	shift, shortcut, word := state.Down(input.KeyShift), state.Down(shortcutKey), state.Down(wordKey)
	caretBefore, textBefore := editor.caret, len(editor.text)
	lineKey := shortcut && shortcutKey == input.KeySuper // Cmd+arrow goes to the line's ends on a Mac
	switch {
	case state.Pressed(input.KeyEscape):
		ui.focus = 0
		return fieldCancelled
	case state.Pressed(input.KeyEnter):
		ui.focus = 0
		return fieldSubmitted
	case state.Pressed(input.KeyTab):
		ui.tabFrom, ui.tabBack = id, shift
		ui.focus = 0
		return fieldLeft
	}
	if !shortcut || state.Down(input.KeyAlt) && shortcutKey == input.KeyControl { // AltGr is Ctrl+Alt
		editor.insert(state.TypedText(), typingEdit)
	}
	switch {
	case state.Typed(input.KeyLeft) && lineKey, state.Typed(input.KeyHome), state.Typed(input.KeyUp):
		editor.moveTo(0, shift)
	case state.Typed(input.KeyRight) && lineKey, state.Typed(input.KeyEnd), state.Typed(input.KeyDown):
		editor.moveTo(len(editor.text), shift)
	case state.Typed(input.KeyLeft) && word:
		editor.moveTo(editor.wordBefore(editor.caret), shift)
	case state.Typed(input.KeyRight) && word:
		editor.moveTo(editor.wordAfter(editor.caret), shift)
	case state.Typed(input.KeyLeft):
		editor.moveBy(-1, shift)
	case state.Typed(input.KeyRight):
		editor.moveBy(1, shift)
	case state.Typed(input.KeyBackspace):
		editor.deleteTo(ui.deletionEnd(editor, -1, word, lineKey))
	case state.Typed(input.KeyDelete):
		editor.deleteTo(ui.deletionEnd(editor, 1, word, lineKey))
	case shortcut && state.Typed(input.KeyA):
		editor.selectAll()
	case shortcut && state.Typed(input.KeyC) && editor.hasSelection():
		state.SetClipboardText(editor.selectedText())
	case shortcut && state.Typed(input.KeyX) && editor.hasSelection():
		state.SetClipboardText(editor.selectedText())
		editor.deleteTo(editor.caret)
	case shortcut && state.Typed(input.KeyV):
		editor.insert(state.ClipboardText(), otherEdit)
	case shortcut && state.Typed(input.KeyZ) && shift, shortcut && state.Typed(input.KeyY):
		editor.redoLast()
	case shortcut && state.Typed(input.KeyZ):
		editor.undoLast()
	}
	if editor.caret != caretBefore || len(editor.text) != textBefore {
		memory.blinkStart = ui.now
	}
	return fieldEditing
}

// deletionEnd is where Backspace (direction -1) or Delete (+1) deletes to: a character, a word
// with the word key, the line's end with Cmd on a Mac.
func (ui *Context) deletionEnd(editor *editor, direction int, word, line bool) int {
	switch {
	case line && direction < 0:
		return 0
	case line:
		return len(editor.text)
	case word && direction < 0:
		return editor.wordBefore(editor.caret)
	case word:
		return editor.wordAfter(editor.caret)
	default:
		return editor.caret + direction
	}
}

// addField adds the field's box and works out what painting it needs: the caret and the
// selection, and how far to move the text so the caret stays in sight.
func (ui *Context) addField(id ID, memory *fieldMemory, text string, options TextFieldOptions, result interaction, focused bool) {
	theme := ui.Theme
	paint := &memory.paint
	*paint = fieldPaint{text: text, placeholder: options.Placeholder, focused: focused}
	if focused {
		measures := ui.measureField(memory, text)
		paint.caretX, paint.selectionStart, paint.selectionEnd = measures.caretX, measures.selectionStart, measures.selectionEnd
		paint.caretShown = ui.blinkCaret(memory)
		memory.scroll = ui.keepCaretInSight(id, memory.scroll, paint.caretX, measures.textWidth)
	} else {
		memory.scroll = 0
	}
	paint.scroll = memory.scroll
	width := options.Width
	if width == (Sizing{}) {
		width = Grow()
	}
	paint.weight = options.Weight
	look := Look{Fill: theme.Surface, Border: theme.ControlLine, BorderWidth: 1, Radius: theme.ControlRadius}
	switch {
	case focused:
		look.Border, look.BorderWidth = theme.Primary, 1.5
	case result.hovered:
		look.Border = theme.InkQuiet
	}
	if result.hovered || result.held {
		ui.cursor = input.CursorText
	}
	layout := Layout{Width: width.Between(max(width.Min, 2*fieldPadding+20), width.Max), Height: Fixed(theme.ControlHeight)}
	index := ui.frameWithID(id, layout, look, nil)
	ui.elements[index].field = paint
}

// fieldMeasures are the widths a focused field paints with, from the start of its text.
type fieldMeasures struct {
	caretX, selectionStart, selectionEnd, textWidth float32
}

// fieldMeasureKey is what fieldMeasures depend on: they are worked out again only when it changes,
// not every frame the field has the focus.
type fieldMeasureKey struct {
	text          string
	caret, anchor int
	size, scale   float32
	weight        draw2d.Weight
}

// measureField returns the caret's and the selection's offsets and the text's width.
func (ui *Context) measureField(memory *fieldMemory, text string) fieldMeasures {
	editor := &memory.editor
	key := fieldMeasureKey{text: text, caret: editor.caret, anchor: editor.anchor, size: ui.Theme.TextSize, scale: ui.scale, weight: memory.weight}
	if memory.measuredValid && memory.measuredKey == key {
		return memory.measured
	}
	start, end := editor.selection()
	memory.measured = fieldMeasures{
		caretX:         ui.textWidth(string(editor.text[:editor.caret]), memory.weight),
		selectionStart: ui.textWidth(string(editor.text[:start]), memory.weight),
		selectionEnd:   ui.textWidth(string(editor.text[:end]), memory.weight),
		textWidth:      ui.textWidth(text, memory.weight),
	}
	memory.measuredKey, memory.measuredValid = key, true
	return memory.measured
}

// caretBlinksFor is how long the caret keeps blinking after the last key or click; then it stays
// shown and the window stops waking up for it.
const caretBlinksFor = 10 * time.Second

// blinkCaret tells whether the caret shows now, and asks for the frame of its next blink.
func (ui *Context) blinkCaret(memory *fieldMemory) bool {
	since := ui.now.Sub(memory.blinkStart)
	if since >= caretBlinksFor {
		return true
	}
	ui.state.RequestFrame(caretBlink - since%caretBlink) // the next blink, in a window that draws on demand
	return since/caretBlink%2 == 0
}

// keepCaretInSight moves the text so the caret is inside the field, without leaving empty space
// after the text when it is shorter than the field allows.
func (ui *Context) keepCaretInSight(id ID, scroll, caretX, textWidth float32) float32 {
	area, shown := ui.previous[id]
	if !shown {
		return 0
	}
	room := area.Width - 2*fieldPadding
	switch {
	case caretX-scroll > room-1:
		scroll = caretX - room + 1
	case caretX < scroll:
		scroll = caretX
	}
	return clamp(scroll, 0, max(textWidth-room+1, 0))
}

// paintField paints a field's text, selection and caret inside its box.
func (ui *Context) paintField(element *element) {
	theme := ui.Theme
	paint := element.field
	rect := element.rect
	if paint.focused {
		ring := draw2d.Rect{X: rect.X - 3, Y: rect.Y - 3, Width: rect.Width + 6, Height: rect.Height + 6}
		ui.drawing.Rectangle(ring, element.look.Radius+3, draw2d.Transparent, theme.FocusRing, 3)
	}
	clip := draw2d.Rect{X: rect.X + fieldPadding - 1, Y: rect.Y, Width: rect.Width - 2*fieldPadding + 2, Height: rect.Height}
	ui.drawing.PushClip(clip)
	defer ui.drawing.PopClip()
	size := theme.TextSize
	metrics := ui.measure("Hg", TextStyle{Size: size})
	baseline := rect.Y + (rect.Height-metrics.Height())/2 + metrics.Ascent
	x := rect.X + fieldPadding - paint.scroll
	if paint.focused && paint.selectionEnd > paint.selectionStart {
		selection := draw2d.Rect{X: x + paint.selectionStart, Y: baseline - metrics.Ascent - 1, Width: paint.selectionEnd - paint.selectionStart, Height: metrics.Height() + 2}
		ui.drawing.Rectangle(selection, 2, theme.Selection, draw2d.Transparent, 0)
	}
	if paint.text == "" && paint.placeholder != "" {
		ui.drawing.Text(paint.placeholder, x, baseline, size, draw2d.Regular, theme.InkQuiet)
	} else {
		ui.drawing.Text(paint.text, x, baseline, size, paint.weight, theme.Ink)
	}
	if paint.focused && paint.caretShown {
		caret := draw2d.Rect{X: x + paint.caretX - 0.75, Y: baseline - metrics.Ascent - 1, Width: 1.5, Height: metrics.Height() + 2}
		ui.drawing.Rectangle(caret, 0, theme.Ink, draw2d.Transparent, 0)
	}
}

// Focus gives the keyboard to a text field or a number field, found by its scopes and label as
// Area finds it, with its whole text selected: to start typing in a search box on a shortcut, or
// in the first field of a form that just opened.
func (ui *Context) Focus(path ...string) {
	id := idForPath(path)
	if entry, found := ui.memory[id]; found {
		if _, isNumber := entry.value.(*numberMemory); isNumber {
			id = childID(id, "##edit")
		}
	}
	ui.focus, ui.focusByKeyboard, ui.focusSeen = id, true, true
}

// moveFocusByTab gives the focus to the field after (or, with Shift, before) the one where Tab
// was pressed, going round at the ends. The field selects its whole text when it takes the focus.
func (ui *Context) moveFocusByTab() {
	from := ui.tabFrom
	ui.tabFrom = 0
	focusables := ui.focusables
	if len(ui.nextSheets) > 0 { // a sheet is open: Tab goes round its own fields only
		focusables = focusables[ui.sheetFocusStart:ui.sheetFocusEnd]
	}
	if from == 0 || len(focusables) == 0 {
		return
	}
	for index, id := range focusables {
		if id != from {
			continue
		}
		step := 1
		if ui.tabBack {
			step = -1
		}
		next := (index + step + len(focusables)) % len(focusables)
		ui.focus, ui.focusByKeyboard = focusables[next], true
		return
	}
}
