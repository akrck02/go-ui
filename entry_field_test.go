package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
)

// A line sent with Enter tells the program on that step, and only then, so a tag or a message is
// added once.
func TestAnEntryFieldTellsWhenEnterSendsIt(t *testing.T) {
	h := newHarness(t)
	text, sent := "", 0
	build := func(ui *Context) {
		if ui.EntryField("Tag", &text, TextFieldOptions{}) {
			sent++
		}
	}
	h.frame(build)
	h.ui.FocusField("Tag")
	h.frame(build)
	h.typeText("lighthouse", build)
	if sent != 0 {
		t.Fatalf("typing sent the line %d times", sent)
	}
	h.key(input.KeyEnter, build)
	h.frame(build)
	if sent != 1 || text != "lighthouse" {
		t.Errorf("Enter sent the line %d times, with %q", sent, text)
	}
}
