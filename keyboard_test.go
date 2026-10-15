package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

func TestAnOnScreenKeyboardEditsTheFieldTypedIn(t *testing.T) {
	h := newHarness(t)
	name := "Harbour"
	changed := false
	build := func(ui *Context) {
		changed = ui.TextField("name", &name)
	}
	h.frame(build)
	if _, typing := h.ui.TypedField(); typing {
		t.Fatal("no field is typed in yet")
	}
	h.ui.Focus("name")
	h.frame(build)
	field, typing := h.ui.TypedField()
	if !typing || field.Field != idOf("name") || field.Text != "Harbour" || field.SelectionStart != 0 || field.SelectionEnd != 7 {
		t.Fatalf("the field focused, its text selected: %+v", field)
	}
	h.ui.EditTypedField(KeyboardText{Field: field.Field, Text: "Harbour li", SelectionStart: 10, SelectionEnd: 10, ComposingStart: 8, ComposingEnd: 10})
	h.frame(build)
	field, _ = h.ui.TypedField()
	if name != "Harbour li" || !changed || field.ComposingStart != 8 || field.ComposingEnd != 10 || field.SelectionEnd != 10 {
		t.Fatalf("the keyboard's text, its caret and the word it composes: %q %+v", name, field)
	}
	if _, underlined := h.underline(); !underlined {
		t.Fatal("the word being composed is underlined")
	}
	h.typeText("g", build)
	field, _ = h.ui.TypedField()
	if name != "Harbour lig" || field.ComposingEnd != field.ComposingStart {
		t.Fatalf("a key typed ends the composing: %q %+v", name, field)
	}
	if _, underlined := h.underline(); underlined {
		t.Fatal("nothing is underlined then")
	}
	h.key(input.KeyBackspace, build)
	h.key(input.KeyBackspace, build)
	h.key(input.KeyBackspace, build)
	h.ui.EditTypedField(KeyboardText{Field: field.Field, Text: "Harbour light", SelectionStart: 13, SelectionEnd: 13})
	h.frame(build)
	h.key(input.KeyZ, build, shortcutKey)
	if name != "Harbour " {
		t.Fatalf("the keyboard's words are undone as typing is: %q", name)
	}
}

func TestAKeyboardEditGoesToItsFieldOnly(t *testing.T) {
	h := newHarness(t)
	name, code := "Harbour", "12"
	build := func(ui *Context) {
		ui.TextField("name", &name)
		ui.TextFieldWith("code", &code, TextFieldOptions{MaxLength: 3})
	}
	h.ui.Focus("code")
	h.frame(build)
	h.ui.EditTypedField(KeyboardText{Field: idOf("name"), Text: "lost"})
	h.frame(build)
	if name != "Harbour" || code != "12" {
		t.Fatalf("an edit of a field not typed in is dropped: %q %q", name, code)
	}
	h.ui.EditTypedField(KeyboardText{Field: idOf("code"), Text: "12\n45", SelectionStart: 5, SelectionEnd: 5, ComposingStart: 3, ComposingEnd: 5})
	h.frame(build)
	field, _ := h.ui.TypedField()
	if code != "12 " || field.SelectionEnd != 3 || field.ComposingEnd != field.ComposingStart {
		t.Fatalf("one line, cut at the limit, the caret kept inside: %q %+v", code, field)
	}
}

func TestASecretFieldSaysSo(t *testing.T) {
	h := newHarness(t)
	password := ""
	build := func(ui *Context) {
		ui.TextFieldWith("password", &password, TextFieldOptions{Secret: true})
	}
	h.ui.Focus("password")
	h.frame(build)
	if field, _ := h.ui.TypedField(); !field.Secret {
		t.Fatal("a password asks for a keyboard with no suggestions")
	}
}

// underline finds the line under a word being composed: a thin rectangle in the ink's colour.
func (h *harness) underline() (draw2d.Command, bool) {
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.RectangleKind && command.Area.Height == composingLine && command.Fill == h.ui.Theme.Ink {
			return command, true
		}
	}
	return draw2d.Command{}, false
}
