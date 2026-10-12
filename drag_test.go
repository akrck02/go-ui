package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
)

// dragScene has a drag source carrying a word (with a button inside), a drop target that takes
// words, and one that takes only numbers, on the screen around open space.
type dragScene struct {
	dropped        []any
	refused        int
	sourceClicks   int
	buttonClicks   int
	screenDrops    []Drag
	draggingInside bool
}

func (scene *dragScene) build(ui *Context) {
	ui.Box(Layout{Gap: 100, Padding: Pad(20)}, func() {
		if ui.DragSource("Lamp", "lamp", Layout{Width: Fixed(100), Height: Fixed(40), Padding: Pad(4)}, func() {
			if ui.Button("Info") {
				scene.buttonClicks++
			}
		}) {
			scene.sourceClicks++
		}
		if payload, dropped := ui.DropTarget("Shelf", Layout{Width: Fixed(150), Height: Fixed(100)}, nil, nil); dropped {
			scene.dropped = append(scene.dropped, payload)
		}
		onlyNumbers := func(payload any) bool { _, isNumber := payload.(int); return isNumber }
		if _, dropped := ui.DropTarget("Scale", Layout{Width: Fixed(150), Height: Fixed(100)}, onlyNumbers, nil); dropped {
			scene.refused++
		}
	})
	_, scene.draggingInside = ui.Dragging()
}

// dragTo presses inside an area, moves to a point over a few steps and lets go there.
func (h *harness) dragTo(from ID, x, y float32, build func(ui *Context)) {
	area := h.rect(from)
	h.moveTo(float64(area.X+area.Width-6), float64(area.Y+6)) // on the source, off the button inside
	h.frame(build)
	h.press()
	h.frame(build)
	h.moveTo(float64(x), float64(y))
	h.frame(build)
	h.frame(build)
	h.release()
	h.frame(build)
}

func TestDraggingOntoATargetDropsThePayload(t *testing.T) {
	h := newHarness(t)
	scene := &dragScene{}
	h.frame(scene.build)
	h.frame(scene.build)
	source, shelf := h.rect(idOf("Lamp")), h.rect(idOf("Shelf"))
	h.moveTo(float64(source.Right()-6), float64(source.Y+6))
	h.frame(scene.build)
	h.press()
	h.frame(scene.build)
	h.moveTo(float64(source.Right()-4), float64(source.Y+7)) // under the threshold
	h.frame(scene.build)
	if scene.draggingInside {
		t.Fatal("a move of a few pixels is not a drag yet")
	}
	h.moveTo(float64(shelf.X+20), float64(shelf.Y+20))
	h.frame(scene.build)
	h.frame(scene.build)
	if !scene.draggingInside || !h.ui.WantsMouse() || h.state.CursorShape() != input.CursorMove {
		t.Fatal("past the threshold the payload is dragged, the interface holds the mouse and the pointer shows it")
	}
	highlighted := false
	for _, element := range h.ui.elements {
		highlighted = highlighted || (element.id == idOf("Shelf") && element.look.Fill == h.ui.Theme.PrimaryTint)
	}
	if !highlighted {
		t.Fatal("a target under a payload it takes is highlighted")
	}
	if drag, _ := h.ui.Dragging(); drag.OverScreen {
		t.Fatal("over a target is not over the open screen")
	}
	h.release()
	h.frame(scene.build)
	if len(scene.dropped) != 1 || scene.dropped[0] != "lamp" {
		t.Fatalf("letting go over the target hands it the payload: %v", scene.dropped)
	}
	if _, onScreen := h.ui.DroppedOnScreen(); onScreen {
		t.Fatal("a drop on a target is not a drop on the screen")
	}
	if scene.sourceClicks != 0 {
		t.Fatal("a drag is not a click on its source")
	}
	h.frame(scene.build)
	if scene.draggingInside || len(scene.dropped) != 1 {
		t.Fatal("the drag is over")
	}
}

func TestDroppingOnTheOpenScreenIsLeftToTheProgram(t *testing.T) {
	h := newHarness(t)
	scene := &dragScene{}
	h.frame(scene.build)
	h.frame(scene.build)
	h.dragTo(idOf("Lamp"), 400, 400, scene.build)
	drop, onScreen := h.ui.DroppedOnScreen()
	if !onScreen || drop.Payload != "lamp" || drop.X != 400 || drop.Y != 400 {
		t.Fatalf("after End, the step of the release tells the payload and where it fell: %+v", drop)
	}
	if !h.state.MouseClaimed() && !h.ui.WantsMouse() {
		t.Fatal("the release is the drop's, not a click on the scene")
	}
	h.frame(scene.build)
	if _, onScreen := h.ui.DroppedOnScreen(); onScreen {
		t.Fatal("only on the step of the release")
	}
}

func TestATargetRefusesWhatItDoesNotTake(t *testing.T) {
	h := newHarness(t)
	scene := &dragScene{}
	h.frame(scene.build)
	h.frame(scene.build)
	scale := h.rect(idOf("Scale"))
	h.dragTo(idOf("Lamp"), scale.X+20, scale.Y+20, scene.build)
	if scene.refused != 0 {
		t.Fatal("a target does not take a payload it refuses")
	}
	if _, onScreen := h.ui.DroppedOnScreen(); onScreen {
		t.Fatal("nor does it fall through to the scene")
	}
}

func TestEscapeCancelsADrag(t *testing.T) {
	h := newHarness(t)
	scene := &dragScene{}
	h.frame(scene.build)
	h.frame(scene.build)
	source := h.rect(idOf("Lamp"))
	h.moveTo(float64(source.Right()-6), float64(source.Y+6))
	h.frame(scene.build)
	h.press()
	h.frame(scene.build)
	h.moveTo(400, 400)
	h.frame(scene.build)
	if !h.ui.WantsKeyboard() {
		t.Fatal("a drag holds the keyboard, for Escape")
	}
	h.key(input.KeyEscape, scene.build)
	if _, dragging := h.ui.Dragging(); dragging {
		t.Fatal("Escape cancels the drag")
	}
	h.release()
	h.frame(scene.build)
	if _, onScreen := h.ui.DroppedOnScreen(); onScreen || len(scene.dropped) != 0 || scene.sourceClicks != 0 {
		t.Fatal("letting go after Escape drops nothing and clicks nothing")
	}
	if !h.ui.WantsMouse() {
		t.Fatal("the release after a cancelled drag is still not the scene's")
	}
}

func TestADragSourceStillTakesClicks(t *testing.T) {
	h := newHarness(t)
	scene := &dragScene{}
	h.frame(scene.build)
	h.frame(scene.build)
	h.clickInside(hashString(fnvOffset, "Info"), scene.build)
	source := h.rect(idOf("Lamp"))
	h.clickAt(source.Right()-6, source.Y+6, scene.build)
	if scene.buttonClicks != 1 || scene.sourceClicks != 1 {
		t.Fatalf("a click without moving is a click, on the button or on the source: %d, %d", scene.buttonClicks, scene.sourceClicks)
	}
	button := h.rect(idOf("Info"))
	h.moveTo(float64(button.X+4), float64(button.Y+4))
	h.frame(scene.build)
	h.press()
	h.frame(scene.build)
	h.moveTo(float64(button.X+30), float64(button.Y+30))
	h.frame(scene.build)
	h.moveTo(float64(button.X+6), float64(button.Y+6)) // back over the button
	h.frame(scene.build)
	h.release()
	h.frame(scene.build)
	if scene.buttonClicks != 1 {
		t.Fatal("a drag that starts on a button inside takes the mouse from it: no click")
	}
}
