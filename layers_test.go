package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

func TestAModalIsDismissedOutsideItsContentOnly(t *testing.T) {
	h := newHarness(t)
	dismissed, pressed := 0, 0
	build := func(ui *Context) {
		if ui.Button("behind") {
			t.Fatal("a button behind a modal was clicked")
		}
		if ui.Modal("question", LayerOptions{AlignX: Center, AlignY: Center, Width: Fixed(200)}, func() {
			if ui.Button("inside") {
				pressed++
			}
		}) {
			dismissed++
		}
	}
	h.frame(build)
	inside := h.rect(childID(idOf("question"), "inside"))
	if inside.X < 290 || inside.X > 310 {
		t.Fatalf("content at x %v: want it centred in the 800-pixel window", inside.X)
	}
	h.moveTo(float64(inside.X+5), float64(inside.Y+5))
	h.press()
	h.frame(build)
	h.release()
	h.frame(build)
	if pressed != 1 || dismissed != 0 {
		t.Fatalf("inside pressed %d, dismissed %d: a touch on the content is not a dismissal", pressed, dismissed)
	}
	h.moveTo(20, 20) // the backdrop, over the button behind
	h.press()
	h.frame(build)
	h.release()
	h.frame(build)
	if dismissed != 1 {
		t.Fatal("a touch on the backdrop did not dismiss the modal")
	}
}

func TestAnOverlaySitsWhereItsOptionsSayAndLetsTheScreenWork(t *testing.T) {
	h := newHarness(t)
	behind, action := 0, 0
	build := func(ui *Context) {
		if ui.Button("behind") {
			behind++
		}
		ui.Overlay("note", LayerOptions{AlignX: Center, AlignY: End, Padding: Pad(16), Width: Fixed(300)}, func() {
			if ui.Button("undo") {
				action++
			}
		})
	}
	h.frame(build)
	note := h.rect(idOf("note"))
	if note.X != 250 || note.Y+note.Height != 584 {
		t.Fatalf("overlay at %+v: want it centred, 16 pixels above the bottom", note)
	}
	button := h.rect(childID(idOf("note"), "undo"))
	h.moveTo(float64(button.X+5), float64(button.Y+5))
	h.press()
	h.frame(build)
	h.release()
	h.frame(build)
	behindArea := h.rect(idOf("behind"))
	h.moveTo(float64(behindArea.X+5), float64(behindArea.Y+5))
	h.press()
	h.frame(build)
	h.release()
	h.frame(build)
	if action != 1 || behind != 1 {
		t.Fatalf("action %d, behind %d: both should work", action, behind)
	}
}

func TestTheInterfaceSaysWhenAFieldIsBeingTypedIn(t *testing.T) {
	h := newHarness(t)
	text := ""
	build := func(ui *Context) { ui.TextField("name", &text) }
	h.frame(build)
	if h.ui.Typing() {
		t.Fatal("typing before any field was touched")
	}
	field := h.rect(idOf("name"))
	h.moveTo(float64(field.X+5), float64(field.Y+5))
	h.press()
	h.frame(build)
	h.release()
	h.frame(build)
	if !h.ui.Typing() {
		t.Fatal("a touched field is not being typed in")
	}
}

func TestAPlainFieldDrawsNoBoxOfItsOwn(t *testing.T) {
	h := newHarness(t)
	text := "harbour"
	build := func(ui *Context) { ui.TextFieldWith("search", &text, TextFieldOptions{Plain: true}) }
	h.frame(build)
	field := h.rect(idOf("search"))
	h.moveTo(float64(field.X+5), float64(field.Y+5)) // focused, it draws no ring either
	h.press()
	h.frame(build)
	h.release()
	h.frame(build)
	if !h.ui.Typing() {
		t.Fatal("the plain field did not take the focus")
	}
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.RectangleKind && (command.Area == field || command.Area.Width > field.Width) {
			t.Fatalf("a plain field drew its box: %+v", command)
		}
	}
}

func TestAPaintedBoxIsGivenItsArea(t *testing.T) {
	h := newHarness(t)
	var given draw2d.Rect
	h.frame(func(ui *Context) {
		ui.Box(Layout{Padding: Pad(10)}, func() {
			ui.Paint(Layout{Width: Fixed(40), Height: Fixed(30)}, func(area draw2d.Rect, drawing *draw2d.List) {
				given = area
				drawing.Rectangle(area, 4, draw2d.Hex("#ff0000"), draw2d.Transparent, 0)
			})
		})
	})
	if given != (draw2d.Rect{X: 10, Y: 10, Width: 40, Height: 30}) {
		t.Fatalf("painted at %+v; want the box's laid out area", given)
	}
	painted := false
	for _, command := range h.ui.Drawing().Commands {
		painted = painted || (command.Kind == draw2d.RectangleKind && command.Area == given)
	}
	if !painted {
		t.Fatal("what the program painted is not in the frame")
	}
}
