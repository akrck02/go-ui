package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

// What is dragged must be seen in the hand: the source itself, not a label, follows the pointer,
// held at the spot where it was taken, while a faint trace stays where it came from.
func TestADraggedSourceFollowsThePointerWhereItWasTaken(t *testing.T) {
	h := newHarness(t)
	scene := &dragScene{}
	h.frame(scene.build)
	h.frame(scene.build)
	source := h.rect(idOf("Lamp"))
	grabX, grabY := source.Width-6, float32(6)
	h.moveTo(float64(source.X+grabX), float64(source.Y+grabY))
	h.frame(scene.build)
	h.press()
	h.frame(scene.build)
	pointerX, pointerY := float32(400), float32(300)
	h.moveTo(float64(pointerX), float64(pointerY))
	h.frame(scene.build)
	h.frame(scene.build)

	var lifted, label, faint bool
	for _, command := range h.ui.drawing.Commands {
		if command.Kind == draw2d.TextKind && command.Text == "Info" {
			if command.Layer == dragLayer && command.Area.X > pointerX-grabX-1 && command.Area.X < pointerX-grabX+source.Width {
				lifted = true
			}
			if command.Layer != dragLayer && command.Fill[3] < 1 {
				faint = true
			}
		}
		if command.Kind == draw2d.TextKind && command.Text == "Lamp" {
			label = true
		}
	}
	if !lifted {
		t.Errorf("the source's content does not follow the pointer")
	}
	if !faint {
		t.Errorf("the source's place does not keep a faint trace of it")
	}
	if label {
		t.Errorf("a label with the source's name still follows the pointer")
	}
}
