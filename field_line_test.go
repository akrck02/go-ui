package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

// A design may want its fields softer than its buttons: FieldLine borders fields and selects, and
// a theme without one keeps ControlLine, as before it existed.
func TestFieldsAreBorderedByFieldLineWhenTheThemeGivesOne(t *testing.T) {
	soft := draw2d.MustHex("#dddddd")
	for _, each := range []struct {
		fieldLine, want draw2d.Color
	}{{draw2d.Color{}, LightTheme().ControlLine}, {soft, soft}} {
		h := newHarness(t)
		h.ui.Theme.FieldLine = each.fieldLine
		h.moveTo(700, 500) // away from both, which hovered take another border
		text, choice := "", 0
		build := func(ui *Context) {
			ui.TextField("Name", &text)
			ui.Select("Kind", &choice, []string{"One", "Two"})
		}
		h.frame(build)
		h.frame(build)
		bordered := 0
		for _, command := range h.ui.drawing.Commands {
			if command.Kind == draw2d.RectangleKind && command.Border == each.want {
				bordered++
			}
		}
		if bordered < 2 {
			t.Errorf("with FieldLine %v, %d boxes bordered %v, not the field and the select", each.fieldLine, bordered, each.want)
		}
	}
}
