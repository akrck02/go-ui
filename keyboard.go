package ui

// KeyboardText is the field being typed in, as an on-screen keyboard sees and edits it: a phone's
// keyboard suggests words from the text around the caret, keeps the word it is still making
// (composing: a prediction, a Chinese or Japanese reading) until it commits it, and deletes around
// the caret. Positions count characters (runes); the selection runs from SelectionStart, where it
// began, to SelectionEnd, the caret.
type KeyboardText struct {
	Field                        ID // which field: an edit of another field is dropped
	Text                         string
	SelectionStart, SelectionEnd int
	ComposingStart, ComposingEnd int  // the word being composed, underlined; equal when none
	Secret                       bool // a password: the keyboard suggests nothing
}

// keyboardState is what goes between the fields and an on-screen keyboard.
type keyboardState struct {
	typed       KeyboardText // the field typed in this frame
	typing      bool
	asked, edit KeyboardText // the keyboard's edit, for the next frame and for this one
	askedSet    bool
	editSet     bool
}

// TypedField is the field typed in during the frame, for an on-screen keyboard (the window hands it
// on): its text, its selection and the word being composed. False while no field is typed in.
func (ui *Context) TypedField() (KeyboardText, bool) {
	return ui.keyboard.typed, ui.keyboard.typing
}

// EditTypedField puts what an on-screen keyboard made of the field's text in it, in the next frame:
// the field takes the whole text as typing (one line, cut at its MaxLength), the selection and the
// word being composed. An edit of a field no longer typed in is dropped.
func (ui *Context) EditTypedField(edit KeyboardText) {
	ui.keyboard.asked, ui.keyboard.askedSet = edit, true
}

// beginKeyboard starts a frame: the edit asked since the last frame is this frame's.
func (ui *Context) beginKeyboard() {
	keyboard := &ui.keyboard
	keyboard.typing = false
	keyboard.edit, keyboard.editSet = keyboard.asked, keyboard.askedSet
	keyboard.askedSet = false
}

// keyboardEditFor is the keyboard's edit of a field in this frame, if it made one.
func (ui *Context) keyboardEditFor(id ID) (KeyboardText, bool) {
	keyboard := &ui.keyboard
	if !keyboard.editSet || keyboard.edit.Field != id {
		return KeyboardText{}, false
	}
	keyboard.editSet = false
	return keyboard.edit, true
}

// noteTyped remembers the field typed in this frame, for TypedField.
func (ui *Context) noteTyped(id ID, memory *fieldMemory) {
	editor := &memory.editor
	ui.keyboard.typed = KeyboardText{Field: id, Text: editor.String(), SelectionStart: editor.anchor,
		SelectionEnd: editor.caret, ComposingStart: editor.composeStart, ComposingEnd: editor.composeEnd,
		Secret: memory.secret}
	ui.keyboard.typing = true
}
