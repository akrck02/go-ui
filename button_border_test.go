package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

// A button answers the finger with its border: a shade deeper under the pointer and deeper still
// while pressed, in its own colour family, so the design's soft lines stay soft at rest.
func TestAButtonsBorderDeepensWhenHoveredAndMoreWhenPressed(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.Box(Layout{Padding: Pad(20)}, func() { ui.Button("Save") })
	}
	borderOf := func() draw2d.Color {
		area := h.rect(idOf("Save"))
		for _, command := range h.ui.drawing.Commands {
			if command.Kind == draw2d.RectangleKind && command.Area == area {
				return command.Border
			}
		}
		t.Fatalf("the button is not drawn")
		return draw2d.Color{}
	}
	h.moveTo(700, 500)
	h.frame(build)
	h.frame(build)
	rest := borderOf()
	area := h.rect(idOf("Save"))
	h.moveTo(float64(area.X+5), float64(area.Y+5))
	h.frame(build)
	h.frame(build)
	hovered := borderOf()
	h.press()
	h.frame(build)
	h.frame(build)
	pressed := borderOf()
	if rest != h.ui.Theme.ControlLine {
		t.Errorf("at rest the border is %v, not the controls' line", rest)
	}
	darkness := func(colour draw2d.Color) float32 { return 3 - colour[0] - colour[1] - colour[2] }
	if !(darkness(hovered) > darkness(rest) && darkness(pressed) > darkness(hovered)) {
		t.Errorf("the border does not deepen: %v at rest, %v hovered, %v pressed", rest, hovered, pressed)
	}
}
