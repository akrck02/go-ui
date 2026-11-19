package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
)

// A page that opens to be written in gives its area the keyboard at once.
func TestAnEntryAreaTakesTheKeyboardWhenTheProgramGivesIt(t *testing.T) {
	h := newHarness(t)
	text := ""
	build := func(ui *Context) { ui.EntryArea("Prompt", &text, TextAreaOptions{}) }
	h.frame(build)
	h.ui.FocusField("Prompt")
	h.frame(build)
	h.typeText("hello", build)
	if text != "hello" {
		t.Errorf("typed after FocusField: %q", text)
	}
}

// A button that opens a page to write in gives the page's area the keyboard, whether the area is
// built in the same frame as the click or only from the next: the click's own press does not take
// the keyboard back, nor does a frame end before the area is built.
func TestAButtonGivesTheKeyboardToAnAreaItOpens(t *testing.T) {
	for _, later := range []bool{false, true} {
		h := newHarness(t)
		text, opened, built := "", false, false
		build := func(ui *Context) {
			if built {
				ui.EntryArea("Prompt", &text, TextAreaOptions{})
			}
			if ui.Button("New") {
				opened = true
				ui.FocusField("Prompt")
			}
			if opened && !built && !later {
				ui.EntryArea("Prompt", &text, TextAreaOptions{})
			}
			built = opened
		}
		h.frame(build)
		h.clickInside(idOf("New"), build)
		h.frame(build)
		h.typeText("hello", build)
		if text != "hello" {
			t.Errorf("built the frame after the click %v: typed %q", later, text)
		}
	}
}

// A chat's box sends with Enter and breaks the line with Shift+Enter, so a person writes several
// lines and sends them without reaching for the mouse.
func TestAnEntryAreaSendsWithEnterAndBreaksTheLineWithShiftEnter(t *testing.T) {
	h := newHarness(t)
	text, sent := "", 0
	build := func(ui *Context) {
		if ui.EntryArea("Prompt", &text, TextAreaOptions{}) {
			sent++
		}
	}
	h.frame(build)
	h.clickInside(idOf("Prompt"), build)
	h.typeText("A site", build)
	h.key(input.KeyEnter, build, input.KeyShift)
	h.typeText("with prices", build)
	if sent != 0 || text != "A site\nwith prices" {
		t.Fatalf("Shift+Enter breaks the line and sends nothing: sent %d, %q", sent, text)
	}
	h.key(input.KeyEnter, build)
	h.frame(build)
	if sent != 1 || text != "A site\nwith prices" {
		t.Errorf("Enter sends the text once, as it is: sent %d, %q", sent, text)
	}
}

// A chat's box is one line until more are written, then grows with them up to its most, and
// scrolls after: it never pushes the page away.
func TestAnAreaWithMostLinesGrowsWithItsTextUpToThem(t *testing.T) {
	h := newHarness(t)
	text := ""
	build := func(ui *Context) {
		ui.Box(Layout{Width: Fixed(300)}, func() { ui.EntryArea("Prompt", &text, TextAreaOptions{MaxLines: 3}) })
	}
	h.frame(build)
	one := h.rect(idOf("Prompt")).Height
	h.clickInside(idOf("Prompt"), build)
	h.typeText("north", build)
	h.key(input.KeyEnter, build, input.KeyShift)
	h.typeText("south", build)
	h.frame(build)
	two := h.rect(idOf("Prompt")).Height
	for range 4 {
		h.key(input.KeyEnter, build, input.KeyShift)
		h.typeText("more", build)
	}
	h.frame(build)
	most := h.rect(idOf("Prompt")).Height
	if !(one < two && two < most) {
		t.Fatalf("the area grows with its lines: %v, %v, %v", one, two, most)
	}
	lineHeight := two - one
	if most-one > 2*lineHeight+0.5 {
		t.Errorf("the area stops at three lines: one %v, most %v, a line %v", one, most, lineHeight)
	}
}

// What sits beside an area can line up with it: TextAreaHeight tells the height the area takes for
// one line and for more.
func TestTextAreaHeightIsTheHeightAnAreaTakes(t *testing.T) {
	h := newHarness(t)
	text := ""
	build := func(ui *Context) {
		ui.Box(Layout{Width: Fixed(300)}, func() { ui.EntryArea("Prompt", &text, TextAreaOptions{MaxLines: 3}) })
	}
	h.frame(build)
	if got, want := h.rect(idOf("Prompt")).Height, h.ui.TextAreaHeight(1, 0); got != want {
		t.Errorf("one line: the area takes %v, TextAreaHeight says %v", got, want)
	}
	h.clickInside(idOf("Prompt"), build)
	h.typeText("north", build)
	h.key(input.KeyEnter, build, input.KeyShift)
	h.typeText("south", build)
	h.frame(build)
	if got, want := h.rect(idOf("Prompt")).Height, h.ui.TextAreaHeight(2, 0); got != want {
		t.Errorf("two lines: the area takes %v, TextAreaHeight says %v", got, want)
	}
}
