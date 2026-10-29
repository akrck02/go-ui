package ui

import (
	"strings"
	"testing"

	"github.com/akrck02/go-renderer/input"
)

// areaMemory is what a text area of a name keeps between frames.
func areaMemory(h *harness, name string) *fieldMemory {
	h.t.Helper()
	return h.ui.memory[idOf(name)].value.(*fieldMemory)
}

// narrowArea builds a text area 120 pixels wide: room for 100 pixels of text, about fifteen
// characters of the fake measurer's.
func narrowArea(text *string) func(ui *Context) {
	return func(ui *Context) {
		ui.Box(Layout{Width: Fixed(120)}, func() { ui.TextArea("Notes", text, TextAreaOptions{}) })
	}
}

func TestEnterStartsANewLine(t *testing.T) {
	h := newHarness(t)
	text := ""
	build := func(ui *Context) { ui.TextArea("Notes", &text, TextAreaOptions{}) }
	h.frame(build)
	h.clickInside(idOf("Notes"), build)
	h.typeText("Lighthouse", build)
	h.key(input.KeyEnter, build)
	h.typeText("- the lamp", build)
	if text != "Lighthouse\n- the lamp" || h.ui.focus != idOf("Notes") {
		t.Fatalf("Enter keeps the keyboard and breaks the line: %q, focus %d", text, h.ui.focus)
	}
	h.key(input.KeyEscape, build)
	if text != "Lighthouse\n- the lamp" || h.ui.focus != 0 {
		t.Fatalf("Escape leaves a long text as it is, never throws it away: %q", text)
	}
}

func TestLinesTypedInOneStepKeepTheirOrder(t *testing.T) {
	h := newHarness(t)
	text := ""
	build := func(ui *Context) { ui.TextArea("Notes", &text, TextAreaOptions{}) }
	h.frame(build)
	h.clickInside(idOf("Notes"), build)
	feed := h.state.Feed()
	for _, line := range []string{"north", "", "south"} {
		for _, character := range line {
			feed.AddCharacter(character)
		}
		feed.SetKey(input.KeyEnter, true)
		feed.SetKey(input.KeyEnter, false)
	}
	h.frame(build)
	if text != "north\n\nsouth\n" {
		t.Fatalf("each Enter stays where it was typed: %q", text)
	}
}

func TestAPastedTextKeepsItsLines(t *testing.T) {
	h := newHarness(t)
	clipboard := &fakeClipboard{text: "# Harbour\r\n\r\n- north pier\n- south pier"}
	h.state.Clipboard = clipboard
	text := ""
	build := func(ui *Context) { ui.TextArea("Notes", &text, TextAreaOptions{}) }
	h.frame(build)
	h.clickInside(idOf("Notes"), build)
	h.key(input.KeyV, build, shortcutKey)
	if text != "# Harbour\n\n- north pier\n- south pier" {
		t.Fatalf("pasted %q", text)
	}
	h.key(input.KeyUp, build, input.KeyShift)
	h.key(input.KeyUp, build, input.KeyShift)
	h.key(input.KeyC, build, shortcutKey)
	if clipboard.text != "\n- north pier\n- south pier" {
		t.Fatalf("a selection across lines is copied with its line breaks: %q", clipboard.text)
	}
}

func TestTheCaretMovesUpAndDownByLines(t *testing.T) {
	h := newHarness(t)
	text := "north pier\nlamp\nsouth pier"
	build := func(ui *Context) { ui.TextArea("Notes", &text, TextAreaOptions{}) }
	h.frame(build)
	h.clickInside(idOf("Notes"), build)
	editor := &areaMemory(h, "Notes").editor
	h.key(input.KeyEnd, build, shortcutKey)
	if editor.caret != len([]rune(text)) {
		t.Fatalf("the shortcut and End go to the end of the text: %d", editor.caret)
	}
	h.key(input.KeyUp, build)
	if editor.caret != len([]rune("north pier\nlamp")) {
		t.Fatalf("up from the end of the last line lands at the end of the shorter line above: %d", editor.caret)
	}
	h.key(input.KeyHome, build)
	if editor.caret != len([]rune("north pier\n")) {
		t.Fatalf("Home goes to the start of the line, not of the text: %d", editor.caret)
	}
	h.key(input.KeyEnd, build)
	if editor.caret != len([]rune("north pier\nlamp")) {
		t.Fatalf("End goes to the end of the line: %d", editor.caret)
	}
	h.key(input.KeyUp, build)
	h.key(input.KeyUp, build)
	if editor.caret != 0 {
		t.Fatalf("up from the first line goes to the start of the text: %d", editor.caret)
	}
	h.key(input.KeyDown, build)
	h.key(input.KeyDown, build, input.KeyShift)
	if start, end := editor.selection(); start != len([]rune("north pier\n")) || end != len([]rune("north pier\nlamp\n")) {
		t.Fatalf("Shift and down selects down to the same place a line below: %d to %d", start, end)
	}
}

func TestALongLineWrapsAtTheFieldsWidth(t *testing.T) {
	h := newHarness(t)
	text := "the lamp of the lighthouse turns all night"
	build := narrowArea(&text)
	h.frame(build)
	h.frame(build) // the first frame learns the width; the second wraps at it
	lines := areaMemory(h, "Notes").paint.lines
	if len(lines) < 3 {
		t.Fatalf("forty characters in room for fifteen take three lines or more: %q", lines)
	}
	if joined := strings.Join(lines, ""); joined != text {
		t.Fatalf("wrapping loses nothing: %q", joined)
	}
	for _, line := range lines {
		if strings.HasPrefix(line, " ") {
			t.Fatalf("lines break between words, the space staying on the line before: %q", lines)
		}
		if h.ui.textWidth(strings.TrimRight(line, " "), 0, 0) > 100 {
			t.Fatalf("no line is wider than the field: %q", line)
		}
	}
	word := strings.Repeat("a", 40)
	text = word
	h.frame(build)
	h.frame(build)
	if lines := areaMemory(h, "Notes").paint.lines; len(lines) < 3 || strings.Join(lines, "") != word {
		t.Fatalf("a word longer than the field breaks where it must: %q", lines)
	}
}

func TestTheFieldScrollsToKeepTheCaretInView(t *testing.T) {
	h := newHarness(t)
	text := ""
	build := func(ui *Context) {
		ui.TextArea("Notes", &text, TextAreaOptions{Height: Fixed(100)})
	}
	h.frame(build)
	h.clickInside(idOf("Notes"), build)
	for range 20 {
		h.typeText("pier", build)
		h.key(input.KeyEnter, build)
	}
	h.frame(build)
	memory := areaMemory(h, "Notes")
	area := h.rect(idOf("Notes"))
	caretTop := memory.paint.caretY - memory.paint.lineScroll
	if memory.paint.lineScroll <= 0 || caretTop < 0 || caretTop+memory.paint.lineHeight > area.Height {
		t.Fatalf("scrolled %v leaves the caret at %v in a field %v tall", memory.paint.lineScroll, caretTop, area.Height)
	}
	h.key(input.KeyHome, build, shortcutKey)
	h.frame(build)
	if memory.paint.lineScroll != 0 {
		t.Fatalf("going to the start scrolls back to the top: %v", memory.paint.lineScroll)
	}
}

func TestAClickPlacesTheCaretOnTheLineUnderIt(t *testing.T) {
	h := newHarness(t)
	text := "north\nsouth"
	build := func(ui *Context) { ui.TextArea("Notes", &text, TextAreaOptions{}) }
	h.frame(build)
	h.frame(build)
	area := h.rect(idOf("Notes"))
	memory := areaMemory(h, "Notes")
	h.clickAt(area.X+fieldPadding+200, area.Y+fieldPadding+memory.paint.lineHeight*1.5, build)
	if memory.editor.caret != len([]rune(text)) {
		t.Fatalf("a click right of the second line puts the caret at its end: %d", memory.editor.caret)
	}
}

func TestTheWheelScrollsTheLinesWithoutTheCaretPullingThemBack(t *testing.T) {
	h := newHarness(t)
	text := strings.Repeat("pier\n", 20)
	build := func(ui *Context) {
		ui.TextArea("Notes", &text, TextAreaOptions{Height: Fixed(100)})
	}
	h.frame(build)
	h.clickInside(idOf("Notes"), build)
	h.key(input.KeyHome, build, shortcutKey)
	h.state.Feed().AddScroll(-2) // two notches down
	h.frame(build)
	h.frame(build)
	if scroll := areaMemory(h, "Notes").lineScroll; scroll <= 0 {
		t.Fatalf("the wheel over the area scrolls its lines, and they stay where it left them: %v", scroll)
	}
}

func TestTheOnScreenKeyboardsLineBreaksAreKept(t *testing.T) {
	h := newHarness(t)
	text := ""
	build := func(ui *Context) { ui.TextArea("Notes", &text, TextAreaOptions{}) }
	h.frame(build)
	h.clickInside(idOf("Notes"), build)
	field, typing := h.ui.TypedField()
	if !typing || !field.MultiLine {
		t.Fatal("the keyboard is told the field takes many lines, so it offers a new-line key")
	}
	h.ui.EditTypedField(KeyboardText{Field: field.Field, Text: "north\nsouth", SelectionStart: 11, SelectionEnd: 11})
	h.frame(build)
	if text != "north\nsouth" {
		t.Fatalf("the keyboard's new line stays one: %q", text)
	}
}

// A page that opens on a draft gives its text area the keyboard before the area was ever shown:
// it takes keys at once, and an on-screen keyboard's edit is followed by keys on the new text.
func TestAnAreaFocusedBeforeItWasShownTakesKeys(t *testing.T) {
	h := newHarness(t)
	text := "north\nsouth"
	build := func(ui *Context) {
		ui.FocusField("Notes")
		ui.TextArea("Notes", &text, TextAreaOptions{})
	}
	h.state.Feed().SetKey(input.KeyUp, true)
	h.frame(build)
	h.state.Feed().SetKey(input.KeyUp, false)
	field, _ := h.ui.TypedField()
	h.ui.EditTypedField(KeyboardText{Field: field.Field, Text: "n", SelectionStart: 1, SelectionEnd: 1})
	h.key(input.KeyEnd, build)
	if text != "n" || areaMemory(h, "Notes").editor.caret != 1 {
		t.Fatalf("got %q, caret %d", text, areaMemory(h, "Notes").editor.caret)
	}
}

// A finger that taps a long text, to go back to it, puts the caret under itself: the text does
// not first jump to its end, where the caret was loaded.
func TestATapOnALongTextPutsTheCaretUnderTheFinger(t *testing.T) {
	h := newHarness(t)
	text := strings.Repeat("pier\n", 30)
	build := func(ui *Context) { ui.TextArea("Notes", &text, TextAreaOptions{Height: Fixed(100)}) }
	h.frame(build)
	area := h.rect(idOf("Notes"))
	h.tapAt(area.X+fieldPadding+1, area.Y+fieldPadding+1, build)
	h.frame(build)
	if memory := areaMemory(h, "Notes"); memory.editor.caret != 0 || memory.lineScroll != 0 {
		t.Fatalf("the caret is at %d, the lines scrolled %v", memory.editor.caret, memory.lineScroll)
	}
}

// End on a line that wrapped stays on that line, after its last word, as in every editor; a
// click right of a short wrapped line lands on it too.
func TestTheEndOfAWrappedLineStaysOnIt(t *testing.T) {
	h := newHarness(t)
	text := "the lamp of the lighthouse turns all night"
	build := narrowArea(&text)
	h.frame(build)
	h.clickInside(idOf("Notes"), build)
	h.key(input.KeyHome, build, shortcutKey)
	memory := areaMemory(h, "Notes")
	h.key(input.KeyEnd, build)
	h.frame(build)
	if line := memory.lineOf(memory.editor.caret); line != 0 || memory.paint.caretY != 0 {
		t.Fatalf("End leaves the caret on line %d, drawn at %v", line, memory.paint.caretY)
	}
	h.key(input.KeyEnd, build)
	if line := memory.lineOf(memory.editor.caret); line != 0 {
		t.Fatalf("a second End does not walk down: line %d", line)
	}
	area := h.rect(idOf("Notes"))
	h.clickAt(area.Right()-fieldPadding-1, area.Y+fieldPadding+memory.paint.lineHeight*1.5, build)
	if line := memory.lineOf(memory.editor.caret); line != 1 {
		t.Fatalf("a click right of the second line lands on line %d", line)
	}
}

// When the area shrinks (a phone's keyboard rising under a page that grows), the caret is kept
// in view without a key being pressed.
func TestAShrinkingAreaKeepsTheCaretInView(t *testing.T) {
	h := newHarness(t)
	text, height := strings.Repeat("pier\n", 12), float32(400)
	build := func(ui *Context) { ui.TextArea("Notes", &text, TextAreaOptions{Height: Fixed(height)}) }
	h.frame(build)
	h.clickInside(idOf("Notes"), build)
	h.key(input.KeyEnd, build, shortcutKey)
	height = 100
	h.frame(build)
	h.frame(build)
	memory := areaMemory(h, "Notes")
	if top := memory.paint.caretY - memory.lineScroll; top < 0 || top+memory.paint.lineHeight > 100 {
		t.Fatalf("the caret is %v from the top of an area 100 tall", top)
	}
}
