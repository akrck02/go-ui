package ui

import (
	"testing"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// fakeClipboard keeps what is copied, as the system clipboard would.
type fakeClipboard struct{ text string }

func (clipboard *fakeClipboard) ClipboardText() string        { return clipboard.text }
func (clipboard *fakeClipboard) SetClipboardText(text string) { clipboard.text = text }

// typeText types characters in one step, as the backend reports them after the keyboard layout.
func (h *harness) typeText(text string, build func(ui *Context)) {
	for _, character := range text {
		h.state.Feed().AddCharacter(character)
	}
	h.frame(build)
}

// key presses a key, with modifiers held, for one step.
func (h *harness) key(key input.Key, build func(ui *Context), modifiers ...input.Key) {
	for _, modifier := range modifiers {
		h.state.Feed().SetKey(modifier, true)
	}
	h.state.Feed().SetKey(key, true)
	h.frame(build)
	h.state.Feed().SetKey(key, false)
	for _, modifier := range modifiers {
		h.state.Feed().SetKey(modifier, false)
	}
}

// clickAt presses and lets go at a point, in one step.
func (h *harness) clickAt(x, y float32, build func(ui *Context)) {
	h.moveTo(float64(x), float64(y))
	h.press()
	h.release()
	h.frame(build)
}

// withClock gives the harness a clock that only moves when told, so double clicks are certain.
func (h *harness) withClock() *time.Time {
	now := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	h.state.Now = func() time.Time { return now }
	return &now
}

func TestEditorTypesDeletesAndUndoesWordByWord(t *testing.T) {
	var editor editor
	editor.load("")
	for _, character := range "Faro del" {
		editor.insert(string(character), typingEdit)
	}
	editor.deleteTo(editor.caret - 1)
	editor.deleteTo(editor.caret - 1)
	editor.insert("Norte", otherEdit)
	if editor.String() != "Faro dNorte" {
		t.Fatalf("got %q", editor.String())
	}
	editor.undoLast() // the paste
	editor.undoLast() // both Backspaces
	if editor.String() != "Faro del" {
		t.Fatalf("after two undos: %q", editor.String())
	}
	editor.undoLast() // the typing
	if editor.String() != "" {
		t.Fatalf("after three undos: %q", editor.String())
	}
	editor.redoLast()
	if editor.String() != "Faro del" || editor.caret != 8 {
		t.Fatalf("after redo: %q, caret %d", editor.String(), editor.caret)
	}
}

func TestEditorWordsAndSelections(t *testing.T) {
	var editor editor
	editor.load("Faro del Norte, ñandú")
	if editor.wordBefore(len(editor.text)) != 16 || editor.wordAfter(0) != 4 || editor.wordAfter(4) != 8 {
		t.Fatalf("word jumps: %d %d %d", editor.wordBefore(len(editor.text)), editor.wordAfter(0), editor.wordAfter(4))
	}
	editor.selectWord(18)
	if editor.selectedText() != "ñandú" {
		t.Fatalf("selected %q", editor.selectedText())
	}
	editor.selectWord(4) // just after "Faro"
	if editor.selectedText() != "Faro" {
		t.Fatalf("selected %q", editor.selectedText())
	}
	editor.moveBy(1, false) // an arrow on a selection lands on its edge
	if editor.caret != 4 || editor.hasSelection() {
		t.Fatalf("caret %d", editor.caret)
	}
	editor.insert("\nX\t", typingEdit) // one line: breaks and tabs become spaces
	if editor.String() != "Faro X  del Norte, ñandú" {
		t.Fatalf("got %q", editor.String())
	}
}

func TestEditorKeepsToItsLimit(t *testing.T) {
	editor := editor{limit: 5}
	editor.load("abc")
	editor.insert("defgh", otherEdit)
	if editor.String() != "abcde" {
		t.Fatalf("got %q", editor.String())
	}
	editor.selectAll()
	editor.insert("123456", otherEdit)
	if editor.String() != "12345" {
		t.Fatalf("a selection replaced within the limit: %q", editor.String())
	}
}

func TestTypingCopyingAndPastingBetweenFields(t *testing.T) {
	h := newHarness(t)
	h.withClock()
	clipboard := &fakeClipboard{}
	h.state.Clipboard = clipboard
	name, copied := "", ""
	build := func(ui *Context) {
		ui.Box(Layout{Direction: Vertical, Width: Fixed(300), Padding: Pad(10), Gap: 10}, func() {
			ui.TextField("Name", &name)
			ui.TextField("Copy", &copied)
		})
	}
	h.frame(build)
	h.clickInside(idOf("Name"), build)
	if h.ui.focus != idOf("Name") || !h.state.KeyboardClaimed() && !h.ui.WantsKeyboard() {
		t.Fatal("a click gives the field the keyboard")
	}
	h.typeText("Faro del Norte", build)
	if name != "Faro del Norte" {
		t.Fatalf("typed %q", name)
	}
	// double-click on "del": characters 5–8, 6.5 pixels each, after 10 of padding
	area := h.rect(idOf("Name"))
	x, y := area.X+10+6.5*6.5, area.Y+area.Height/2
	h.clickAt(x, y, build)
	h.clickAt(x, y, build)
	if got := h.ui.memory[idOf("Name")].value.(*fieldMemory).editor.selectedText(); got != "del" {
		t.Fatalf("double click selected %q", got)
	}
	h.key(input.KeyC, build, shortcutKey)
	if clipboard.text != "del" {
		t.Fatalf("copied %q", clipboard.text)
	}
	h.clickInside(idOf("Copy"), build)
	if h.ui.focus != idOf("Copy") {
		t.Fatal("a click in the other field moves the focus")
	}
	h.key(input.KeyV, build, shortcutKey)
	if copied != "del" || name != "Faro del Norte" {
		t.Fatalf("pasted %q, the first field has %q", copied, name)
	}
	h.typeText("ñá", build)
	h.key(input.KeyBackspace, build)
	if copied != "delñ" {
		t.Fatalf("got %q", copied)
	}
	h.key(input.KeyZ, build, shortcutKey)
	h.key(input.KeyZ, build, shortcutKey)
	if copied != "del" {
		t.Fatalf("undone to %q", copied)
	}
}

func TestEscapePutsTheTextBackAndEnterKeepsIt(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	text := "Faro"
	build := func(ui *Context) { ui.TextField("Name", &text) }
	h.frame(build)
	h.clickInside(idOf("Name"), build)
	h.typeText(" Sur", build)
	h.key(input.KeyEscape, build)
	if text != "Faro" || h.ui.focus != 0 {
		t.Fatalf("Escape: %q, focus %d", text, h.ui.focus)
	}
	*now = now.Add(time.Second) // not a double click
	h.clickInside(idOf("Name"), build)
	h.typeText(" Sur", build)
	h.key(input.KeyEnter, build)
	if text != "Faro Sur" || h.ui.focus != 0 || h.ui.WantsKeyboard() {
		t.Fatalf("Enter: %q, focus %d", text, h.ui.focus)
	}
}

func TestAClickElsewhereOrAFieldThatGoesAwayDropsTheFocus(t *testing.T) {
	h := newHarness(t)
	text, shown := "", true
	build := func(ui *Context) {
		if shown {
			ui.TextField("Name", &text)
		}
	}
	h.frame(build)
	h.clickInside(idOf("Name"), build)
	h.clickAt(700, 500, build)
	if h.ui.focus != 0 {
		t.Fatal("a click elsewhere ends the edit")
	}
	h.clickInside(idOf("Name"), build)
	shown = false
	h.frame(build)
	if h.ui.focus != 0 || h.ui.WantsKeyboard() {
		t.Fatal("a field no longer built lets go of the keyboard")
	}
}

func TestTabGoesThroughTheFieldsInOrder(t *testing.T) {
	h := newHarness(t)
	first, second, number := "uno", "dos", float32(3)
	build := func(ui *Context) {
		ui.Column(func() {
			ui.TextField("First", &first)
			ui.TextField("Second", &second)
			ui.NumberField("Number", &number, 0, 10, 0.5, "%.1f m")
		})
	}
	h.frame(build)
	h.clickInside(idOf("First"), build)
	h.key(input.KeyTab, build)
	h.frame(build)
	if h.ui.focus != idOf("Second") {
		t.Fatal("Tab goes to the next field")
	}
	h.typeText("2", build) // the whole text was selected, so typing replaces it
	if second != "2" {
		t.Fatalf("got %q", second)
	}
	h.key(input.KeyTab, build)
	h.typeText("7,5", build)
	h.key(input.KeyTab, build)
	if number != 7.5 {
		t.Fatalf("the number field took %v", number)
	}
	if h.ui.focus != idOf("First") {
		t.Fatal("Tab goes round to the first field")
	}
	h.key(input.KeyTab, build, input.KeyShift)
	h.frame(build)
	if h.ui.focus != childID(idOf("Number"), "##edit") {
		t.Fatal("Shift+Tab goes back round to the last field")
	}
}

func TestClickingANumberOpensItForTyping(t *testing.T) {
	h := newHarness(t)
	value := float32(2)
	build := func(ui *Context) {
		ui.Box(Layout{Width: Fixed(200)}, func() { ui.NumberField("Range", &value, 0, 10, 0.5, "%.1f km") })
	}
	h.frame(build)
	h.clickInside(childID(idOf("Range"), "##value"), build)
	h.frame(build)
	h.typeText("12 km", build)
	h.key(input.KeyEnter, build)
	if value != 10 {
		t.Fatalf("typed past the maximum: %v", value)
	}
	h.frame(build)
	h.clickInside(childID(idOf("Range"), "##value"), build)
	h.frame(build)
	h.typeText("4", build)
	h.key(input.KeyEscape, build)
	if value != 10 {
		t.Fatalf("Escape keeps the number: %v", value)
	}
	h.frame(build)
	h.clickInside(childID(idOf("Range"), "##value"), build)
	h.frame(build)
	h.typeText("faro", build)
	h.key(input.KeyEnter, build)
	if value != 10 {
		t.Fatalf("text that is no number changes nothing: %v", value)
	}
}

func TestALongTextScrollsToKeepTheCaretInSight(t *testing.T) {
	h := newHarness(t)
	text := ""
	build := func(ui *Context) {
		ui.Box(Layout{Width: Fixed(120)}, func() { ui.TextField("Name", &text) }) // room for 100 px: 15 characters
	}
	h.frame(build)
	h.clickInside(idOf("Name"), build)
	h.typeText("Faro del Norte de la isla", build)
	h.frame(build)
	memory := h.ui.memory[idOf("Name")].value.(*fieldMemory)
	caret := memory.paint.caretX - memory.scroll
	if memory.scroll <= 0 || caret < 0 || caret > 100 {
		t.Fatalf("scroll %v leaves the caret at %v", memory.scroll, caret)
	}
	h.key(input.KeyHome, build)
	h.frame(build)
	if memory.scroll != 0 {
		t.Fatalf("Home scrolls back to the start: %v", memory.scroll)
	}
}

func TestParseNumber(t *testing.T) {
	for text, want := range map[string]float32{"2.5": 2.5, " 2,5 ": 2.5, "-3": -3, "12 km": 12, "1e2": 100} {
		if got, ok := parseNumber(text); !ok || got != want {
			t.Errorf("%q: got %v, %v", text, got, ok)
		}
	}
	for _, text := range []string{"", "faro", "NaN", "Inf"} {
		if _, ok := parseNumber(text); ok {
			t.Errorf("%q should not be a number", text)
		}
	}
}

func TestFocusSelectsTheFieldsText(t *testing.T) {
	h := newHarness(t)
	text, number := "Gull Rock", float32(4)
	build := func(ui *Context) {
		ui.Panel("Search", func() {
			ui.TextField("Name", &text)
			ui.NumberField("Range", &number, 0, 30, 1, "%.0f nm")
		})
	}
	h.frame(build)
	h.ui.Focus("Search", "Name")
	h.frame(build)
	h.typeText("Seal", build)
	if text != "Seal" || !h.ui.WantsKeyboard() {
		t.Fatalf("got %q", text)
	}
	h.ui.Focus("Search", "Range")
	h.frame(build)
	h.typeText("12", build)
	h.key(input.KeyEnter, build)
	if number != 12 {
		t.Fatalf("got %v", number)
	}
}

func TestAnIdleInterfaceAsksForNoFramesButACaretBlinks(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	text, on := "Faro", false
	build := func(ui *Context) {
		ui.Panel("Light", func() {
			ui.TextField("Name", &text)
			ui.Toggle("On", &on)
			ui.Button("Save")
		})
	}
	h.frame(build)
	h.frame(build)
	if _, wanted := h.state.Feed().TakeFrameRequest(); wanted {
		t.Fatal("nothing moves, so nothing needs drawing")
	}
	h.clickInside(scoped("Light", "Name"), build)
	*now = now.Add(200 * time.Millisecond)
	h.frame(build)
	delay, wanted := h.state.Feed().TakeFrameRequest()
	if !wanted || delay != caretBlink-200*time.Millisecond {
		t.Fatalf("the caret asks for its next blink: %v, %v", delay, wanted)
	}
	shown := h.ui.memory[scoped("Light", "Name")].value.(*fieldMemory).paint.caretShown
	*now = now.Add(delay)
	h.frame(build)
	if h.ui.memory[scoped("Light", "Name")].value.(*fieldMemory).paint.caretShown == shown {
		t.Fatal("at that time the caret blinks")
	}
}

func TestASecretFieldShowsDotsAndCopiesNothing(t *testing.T) {
	h := newHarness(t)
	clipboard := &fakeClipboard{text: "before"}
	h.state.Clipboard = clipboard
	password := ""
	build := func(ui *Context) {
		ui.TextFieldWith("password", &password, TextFieldOptions{Secret: true})
	}
	h.frame(build)
	area := h.rect(idOf("password"))
	h.clickAt(area.X+10, area.Y+area.Height/2, build)
	h.typeText("lamp", build)
	if password != "lamp" {
		t.Fatalf("the field holds %q; want what was typed", password)
	}
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.TextKind && command.Text == "lamp" {
			t.Fatal("a secret field shows its text")
		}
	}
	drawn := false
	for _, command := range h.ui.Drawing().Commands {
		drawn = drawn || (command.Kind == draw2d.TextKind && command.Text == "••••")
	}
	if !drawn {
		t.Fatal("a secret field does not show a dot for each character")
	}
	h.key(input.KeyA, build, shortcutKey)
	h.key(input.KeyC, build, shortcutKey)
	if clipboard.text != "before" {
		t.Fatalf("copying from a secret field put %q on the clipboard", clipboard.text)
	}
}

func TestAProgramKnowsWhichFieldIsTypedIn(t *testing.T) {
	h := newHarness(t)
	name, place := "", ""
	var nameFocused, placeFocused bool
	build := func(ui *Context) {
		nameFocused, placeFocused = ui.FieldFocused("name"), ui.FieldFocused("place")
		ui.Column(func() {
			ui.TextField("name", &name)
			ui.TextField("place", &place)
		})
	}
	h.frame(build)
	area := h.rect(idOf("place"))
	h.clickAt(area.X+10, area.Y+area.Height/2, build)
	h.frame(build)
	if nameFocused || !placeFocused {
		t.Fatalf("name focused %v, place focused %v after touching place; want only place", nameFocused, placeFocused)
	}
}

func TestAFieldGivenTheKeyboardIsTypedIn(t *testing.T) {
	h := newHarness(t)
	query, opened := "", false
	build := func(ui *Context) {
		if !opened {
			ui.FocusField("query")
			opened = true
		}
		ui.TextField("query", &query)
	}
	h.frame(build)
	h.typeText("north", build)
	if query != "north" || !h.ui.Typing() {
		t.Fatalf("query %q, typing %v; want the field typed in without a touch", query, h.ui.Typing())
	}
}
