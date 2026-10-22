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
	var given draw2d.Rectangle
	h.frame(func(ui *Context) {
		ui.Box(Layout{Padding: Pad(10)}, func() {
			ui.Paint(Layout{Width: Fixed(40), Height: Fixed(30)}, func(area draw2d.Rectangle, drawing *draw2d.List) {
				given = area
				drawing.Rectangle(area, 4, draw2d.MustHex("#ff0000"), draw2d.Transparent, 0)
			})
		})
	})
	if given != (draw2d.Rectangle{X: 10, Y: 10, Width: 40, Height: 30}) {
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

func TestAModalCoversAnOverlay(t *testing.T) {
	h := newHarness(t)
	h.ui.SetScreenReader(true)
	tool, closed := 0, 0
	build := func(ui *Context) {
		ui.Overlay("tools", LayerOptions{AlignX: Center, AlignY: End, Padding: Pad(16)}, func() {
			if ui.Button("Share") {
				tool++
			}
		})
		if ui.Modal("drawer", LayerOptions{AlignX: Start}, func() { ui.Label("Places") }) {
			closed++
		}
	}
	h.frame(build)
	share, _ := h.textCommand("Share")
	places, _ := h.textCommand("Places")
	if share.Layer >= places.Layer {
		t.Fatalf("the modal is drawn over the overlay: layers %d and %d", share.Layer, places.Layer)
	}
	for _, node := range h.ui.ScreenReaderNodes() {
		if node.Label == "Share" {
			t.Fatal("an overlay behind a modal is not read")
		}
	}
	button := h.rect(childID(idOf("tools"), "Share"))
	h.moveTo(float64(button.X+5), float64(button.Y+5))
	h.press()
	h.frame(build)
	h.release()
	h.frame(build)
	if tool != 0 || closed != 1 {
		t.Fatalf("a touch over the overlay is the modal's backdrop's: %d presses, %d closes", tool, closed)
	}
}

func TestLayersOpenedWithinAnAreaAreLaidOutInIt(t *testing.T) {
	h := newHarness(t)
	area := draw2d.Rectangle{X: 200, Y: 100, Width: 400, Height: 240}
	build := func(ui *Context) {
		ui.Within(area, func() {
			ui.Overlay("corner", LayerOptions{AlignX: End, AlignY: End, Padding: Pad(10)}, func() { ui.Label("account") })
			ui.Modal("dialog", LayerOptions{AlignX: Center, AlignY: Center}, func() { ui.Label("sure?") })
		})
		ui.Overlay("outside", LayerOptions{Padding: Pad(10)}, func() { ui.Label("window") })
	}
	h.frame(build)
	corner := h.rect(idOf("corner"))
	if corner.X+corner.Width != 590 || corner.Y+corner.Height != 330 {
		t.Fatalf("an overlay in the area's corner is at %+v", corner)
	}
	if dialog := h.rect(idOf("dialog")); dialog != area {
		t.Fatalf("a modal within the area covers %+v, not the area", dialog)
	}
	if outside := h.rect(idOf("outside")); outside.X != 10 || outside.Y != 10 {
		t.Fatalf("an overlay after the area is at %+v, not the window's corner", outside)
	}
}

func TestAPanelOpenedWithinAnAreaIsPlacedFromItsCorner(t *testing.T) {
	h := newHarness(t)
	area := draw2d.Rectangle{X: 200, Y: 100, Width: 400, Height: 240}
	build := func(ui *Context) {
		ui.Within(area, func() {
			ui.PanelWith("Account", PanelOptions{X: 16, Y: 16, Width: Fixed(150), Locked: true}, func() { ui.Label("name") })
		})
	}
	h.frame(build)
	if panel := h.rect(idOf("Account")); panel.X != 216 || panel.Y != 116 {
		t.Fatalf("a panel at 16, 16 within the area is at %+v", panel)
	}
}
