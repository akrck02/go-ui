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

// A row of a list draws no background of its own; lifted as it is, its words would float over
// the page and be lost in it. It is lifted on a card of the theme's surface, a little wider than it.
func TestASourceWithoutABackgroundIsLiftedOnACard(t *testing.T) {
	h := newHarness(t)
	scene := &dragScene{}
	h.frame(scene.build)
	h.frame(scene.build)
	source := h.rect(idOf("Lamp"))
	h.moveTo(float64(source.X+source.Width-6), float64(source.Y+6))
	h.frame(scene.build)
	h.press()
	h.frame(scene.build)
	h.moveTo(400, 300)
	h.frame(scene.build)
	h.frame(scene.build)

	for _, command := range h.ui.drawing.Commands {
		if command.Kind == draw2d.RectangleKind && command.Layer == dragLayer &&
			command.Fill == h.ui.Theme.Surface && command.Area.Width > source.Width {
			return
		}
	}
	t.Errorf("the source is lifted with no card under it")
}

// A target holding one card of its own has the card marked when a drop would land there: a
// mark painted under the card would be hidden by it.
func TestATargetHoldingACardHasTheCardMarked(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.Box(Layout{Gap: 100, Padding: Pad(20)}, func() {
			ui.DragSource("Lamp", "lamp", Layout{Width: Fixed(100), Height: Fixed(40)}, func() { ui.Text("Lamp", TextStyle{}) })
			ui.DropTarget("Shelf", Layout{Width: Fixed(150), Height: Fixed(100)}, nil, func() {
				ui.Frame(Layout{Width: Grow(), Height: Grow()}, Look{Fill: ui.Theme.Surface, Radius: 18}, nil)
			})
		})
	}
	h.frame(build)
	h.frame(build)
	source, shelf := h.rect(idOf("Lamp")), h.rect(idOf("Shelf"))
	h.moveTo(float64(source.X+10), float64(source.Y+10))
	h.frame(build)
	h.press()
	h.frame(build)
	h.moveTo(float64(shelf.X+20), float64(shelf.Y+20))
	h.frame(build)
	h.frame(build)
	for _, command := range h.ui.drawing.Commands {
		if command.Kind == draw2d.RectangleKind && command.Area == shelf && command.Radius == 18 {
			if command.BorderWidth != dropLineWidth || command.Fill == h.ui.Theme.Surface {
				t.Errorf("the card is not marked while a drop would land on it")
			}
			return
		}
	}
	t.Errorf("the card is not drawn")
}

// A source that is a short card with a margin under it, and its shadow, already has a background:
// it is lifted as it is, not on a second card around the first.
func TestACardInsideItsSourceIsNotLiftedOnAnotherCard(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.Box(Layout{Padding: Pad(20)}, func() {
			ui.DragSource("Note", "note", Layout{Width: Fixed(200), Padding: Padding{Bottom: 16}}, func() {
				ui.Frame(Layout{Width: Grow(), Height: Fixed(60)}, Look{Fill: ui.Theme.Background, Radius: 12, Shadow: true}, nil)
			})
		})
	}
	h.frame(build)
	h.frame(build)
	source := h.rect(idOf("Note"))
	h.moveTo(float64(source.X+20), float64(source.Y+20))
	h.frame(build)
	h.press()
	h.frame(build)
	h.moveTo(500, 400)
	h.frame(build)
	h.frame(build)
	for _, command := range h.ui.drawing.Commands {
		if command.Kind == draw2d.RectangleKind && command.Layer == dragLayer && command.Fill == h.ui.Theme.Surface {
			t.Fatalf("the card is lifted on a second card")
		}
	}
}
