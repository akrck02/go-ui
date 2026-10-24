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

// Under the pointer a field's border only deepens a shade of its own line: a soft design stays
// soft, instead of jumping to a dark grey.
func TestAHoveredFieldDeepensItsOwnLine(t *testing.T) {
	soft := draw2d.MustHex("#dddddd")
	h := newHarness(t)
	h.ui.Theme.FieldLine = soft
	text := ""
	build := func(ui *Context) { ui.TextField("Name", &text) }
	h.frame(build)
	area := h.rect(idOf("Name"))
	h.moveTo(float64(area.X+5), float64(area.Y+5))
	h.frame(build)
	h.frame(build)
	want := soft.Mix(h.ui.Theme.Ink, borderHoverShade)
	for _, command := range h.ui.drawing.Commands {
		if command.Kind == draw2d.RectangleKind && command.Area == area && command.Border == want {
			return
		}
	}
	t.Errorf("the hovered field is not bordered by a shade of its own line")
}

// An empty field's example words are fainter than any value: «@name» in a field must not read as
// a name typed in.
func TestAPlaceholderIsFainterThanTheQuietestInk(t *testing.T) {
	h := newHarness(t)
	text := ""
	build := func(ui *Context) { ui.TextFieldWith("Name", &text, TextFieldOptions{Placeholder: "@name"}) }
	h.frame(build)
	h.frame(build)
	for _, command := range h.ui.drawing.Commands {
		if command.Kind == draw2d.TextKind && command.Text == "@name" {
			darkness := func(colour draw2d.Color) float32 { return 3 - colour[0] - colour[1] - colour[2] }
			if darkness(command.Fill) >= darkness(h.ui.Theme.InkQuiet) {
				t.Errorf("the placeholder is as dark as quiet ink: %v", command.Fill)
			}
			return
		}
	}
	t.Errorf("the placeholder is not drawn")
}
