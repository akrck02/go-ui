package ui

import (
	"unicode"
)

// editor is the text of a field being edited, with its caret and selection, and what can be undone.
// Positions count characters (runes), not bytes, so "ñ" is one step of the caret. The selection
// runs from anchor to caret; when they are equal nothing is selected.
type editor struct {
	text          []rune
	caret, anchor int
	undo, redo    []editorSnapshot
	lastEdit      editKind // what the last change was, so typing a word is undone in one go
	limit         int      // the most characters the text may have; 0 = no limit
	// asString is the text as a string, kept until the text changes: a focused field compares and
	// writes it every frame.
	asString      string
	asStringValid bool
}

// editorSnapshot is the editor before a change, to go back to.
type editorSnapshot struct {
	text          string
	caret, anchor int
}

type editKind int

const (
	noEdit editKind = iota
	typingEdit
	deletingEdit
	otherEdit // paste, cut, replacing a selection: each is undone on its own
)

// maximumUndoSteps bounds the history a field keeps.
const maximumUndoSteps = 100

// load starts editing a text, with the caret at its end and no history.
func (editor *editor) load(text string) {
	editor.setText([]rune(text))
	editor.caret, editor.anchor = len(editor.text), len(editor.text)
	editor.undo, editor.redo = editor.undo[:0], editor.redo[:0]
	editor.lastEdit = noEdit
}

func (editor *editor) String() string {
	if !editor.asStringValid {
		editor.asString, editor.asStringValid = string(editor.text), true
	}
	return editor.asString
}

// setText replaces the text, forgetting its string form.
func (editor *editor) setText(text []rune) {
	editor.text, editor.asStringValid = text, false
}

// hasSelection tells whether some text is selected.
func (editor *editor) hasSelection() bool { return editor.caret != editor.anchor }

// selection is the selected range, start first.
func (editor *editor) selection() (int, int) {
	return min(editor.caret, editor.anchor), max(editor.caret, editor.anchor)
}

func (editor *editor) selectedText() string {
	start, end := editor.selection()
	return string(editor.text[start:end])
}

// moveTo puts the caret at a position; extending keeps the anchor, so the selection grows.
func (editor *editor) moveTo(position int, extending bool) {
	editor.caret = clampInt(position, 0, len(editor.text))
	if !extending {
		editor.anchor = editor.caret
	}
	editor.lastEdit = noEdit
}

// moveBy moves the caret a number of characters. Without extending, an arrow on a selection lands
// on the selection's edge in that direction, as in every text box.
func (editor *editor) moveBy(steps int, extending bool) {
	if !extending && editor.hasSelection() {
		start, end := editor.selection()
		if steps < 0 {
			editor.moveTo(start, false)
		} else {
			editor.moveTo(end, false)
		}
		return
	}
	editor.moveTo(editor.caret+steps, extending)
}

// selectAll selects the whole text, with the caret at the end.
func (editor *editor) selectAll() {
	editor.anchor, editor.caret = 0, len(editor.text)
	editor.lastEdit = noEdit
}

// selectWord selects the word (or the run of spaces or of punctuation) at a position.
func (editor *editor) selectWord(position int) {
	position = clampInt(position, 0, len(editor.text))
	if len(editor.text) == 0 {
		return
	}
	at := min(position, len(editor.text)-1)
	if position > 0 && (position == len(editor.text) || characterClass(editor.text[at]) != characterClass(editor.text[position-1])) {
		at = position - 1 // a click just after a word, on its end, takes that word
	}
	class := characterClass(editor.text[at])
	start, end := at, at+1
	for start > 0 && characterClass(editor.text[start-1]) == class {
		start--
	}
	for end < len(editor.text) && characterClass(editor.text[end]) == class {
		end++
	}
	editor.anchor, editor.caret = start, end
	editor.lastEdit = noEdit
}

// characterClass groups characters into words: letters and digits, spaces, everything else.
func characterClass(character rune) int {
	switch {
	case unicode.IsLetter(character) || unicode.IsDigit(character) || character == '_':
		return 0
	case unicode.IsSpace(character):
		return 1
	default:
		return 2
	}
}

// wordBefore is where the caret goes when jumping a word back: past any spaces, then to the start
// of the word before them.
func (editor *editor) wordBefore(position int) int {
	for position > 0 && unicode.IsSpace(editor.text[position-1]) {
		position--
	}
	if position == 0 {
		return 0
	}
	class := characterClass(editor.text[position-1])
	for position > 0 && characterClass(editor.text[position-1]) == class {
		position--
	}
	return position
}

// wordAfter is where the caret goes when jumping a word forward: to the end of the next word.
func (editor *editor) wordAfter(position int) int {
	for position < len(editor.text) && unicode.IsSpace(editor.text[position]) {
		position++
	}
	if position == len(editor.text) {
		return position
	}
	class := characterClass(editor.text[position])
	for position < len(editor.text) && characterClass(editor.text[position]) == class {
		position++
	}
	return position
}

// insert types text at the caret, over the selection if there is one. Line breaks and other
// control characters are left out (the field has one line), and the limit cuts what does not fit.
// It returns whether the text changed.
func (editor *editor) insert(text string, kind editKind) bool {
	if text == "" {
		return false
	}
	added := make([]rune, 0, len(text))
	for _, character := range text {
		if character == '\n' || character == '\r' || character == '\t' {
			character = ' '
		}
		if unicode.IsControl(character) {
			continue
		}
		added = append(added, character)
	}
	start, end := editor.selection()
	if editor.limit > 0 {
		room := editor.limit - (len(editor.text) - (end - start))
		added = added[:clampInt(len(added), 0, max(room, 0))]
	}
	if len(added) == 0 && start == end {
		return false
	}
	if start != end {
		kind = otherEdit // replacing a selection is its own step
	}
	editor.remember(kind)
	replaced := make([]rune, 0, len(editor.text)-(end-start)+len(added))
	replaced = append(replaced, editor.text[:start]...)
	replaced = append(replaced, added...)
	replaced = append(replaced, editor.text[end:]...)
	editor.setText(replaced)
	editor.caret = start + len(added)
	editor.anchor = editor.caret
	return true
}

// deleteTo deletes from the caret to a position (the selection instead, if there is one). It
// returns whether the text changed.
func (editor *editor) deleteTo(position int) bool {
	start, end := editor.selection()
	kind := otherEdit
	if start == end {
		start, end = min(editor.caret, position), max(editor.caret, position)
		start, end = clampInt(start, 0, len(editor.text)), clampInt(end, 0, len(editor.text))
		kind = deletingEdit
	}
	if start == end {
		return false
	}
	editor.remember(kind)
	editor.setText(append(editor.text[:start:start], editor.text[end:]...))
	editor.caret, editor.anchor = start, start
	return true
}

// remember keeps the editor as it is before a change, unless the change continues the last one
// (another letter of the same word, another Backspace).
func (editor *editor) remember(kind editKind) {
	editor.redo = editor.redo[:0]
	continues := kind != otherEdit && kind == editor.lastEdit
	editor.lastEdit = kind
	if continues {
		return
	}
	if len(editor.undo) == maximumUndoSteps {
		editor.undo = append(editor.undo[:0], editor.undo[1:]...)
	}
	editor.undo = append(editor.undo, editor.snapshot())
}

func (editor *editor) snapshot() editorSnapshot {
	return editorSnapshot{text: string(editor.text), caret: editor.caret, anchor: editor.anchor}
}

func (editor *editor) restore(snapshot editorSnapshot) {
	editor.setText([]rune(snapshot.text))
	editor.caret, editor.anchor = snapshot.caret, snapshot.anchor
	editor.lastEdit = noEdit
}

// undoLast goes back one step; redoLast goes forward again. They return whether anything changed.
func (editor *editor) undoLast() bool {
	if len(editor.undo) == 0 {
		return false
	}
	last := editor.undo[len(editor.undo)-1]
	editor.undo = editor.undo[:len(editor.undo)-1]
	editor.redo = append(editor.redo, editor.snapshot())
	editor.restore(last)
	return true
}

func (editor *editor) redoLast() bool {
	if len(editor.redo) == 0 {
		return false
	}
	next := editor.redo[len(editor.redo)-1]
	editor.redo = editor.redo[:len(editor.redo)-1]
	editor.undo = append(editor.undo, editor.snapshot())
	editor.restore(next)
	return true
}
