package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

// An action that destroys must never look like an ordinary one: its words are red, on a soft
// red fill, whichever form of button it takes.
func TestADestructiveButtonIsRedOnASoftRedFill(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.Box(Layout{Direction: Vertical, Padding: Pad(20)}, func() {
			ui.DangerButton("Delete")
			ui.DangerIconButton("trash-2", "Empty")
		})
	}
	h.moveTo(700, 500) // away from the buttons
	h.frame(build)
	h.frame(build)
	for _, name := range []string{"Delete", "Empty"} {
		if fill := h.fillOver(idOf(name)); fill != h.ui.Theme.dangerTint() {
			t.Errorf("«%s» is filled with %v, not the soft red", name, fill)
		}
		if !h.drewText(name, h.ui.Theme.Danger) {
			t.Errorf("«%s» is not written in red", name)
		}
	}
}

// A flat delete beside a row stays quiet until the pointer comes: red, with no fill.
func TestAFlatDestructiveButtonHasNoFillUntilHovered(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) { ui.DangerFlatIconButton("trash-2", "Remove") }
	h.moveTo(700, 500) // away from the button
	h.frame(build)
	h.frame(build)
	if fill := h.fillOver(idOf("Remove")); fill.IsVisible() {
		t.Errorf("the flat delete is filled with %v before the pointer comes", fill)
	}
	if !h.drewText("Remove", h.ui.Theme.Danger) {
		t.Errorf("the flat delete is not written in red")
	}
}

// drewText tells whether the last frame wrote these words in this colour.
func (h *harness) drewText(words string, colour draw2d.Color) bool {
	for _, command := range h.ui.drawing.Commands {
		if command.Kind == draw2d.TextKind && command.Text == words && command.Fill == colour {
			return true
		}
	}
	return false
}

// fillOver is the fill of the rectangle drawn over an element's whole area in the last frame.
func (h *harness) fillOver(id ID) draw2d.Color {
	area := h.rect(id)
	for _, command := range h.ui.drawing.Commands {
		if command.Kind == draw2d.RectangleKind && command.Area == area {
			return command.Fill
		}
	}
	return draw2d.Transparent
}
