package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
)

// sheetScene is a panel with a button and a field behind an optional sheet with a button and a
// field of its own.
type sheetScene struct {
	open                     bool
	behindClicks, sheetClick int
	dismissals               int
	behindText, sheetText    string
}

func (scene *sheetScene) build(ui *Context) {
	ui.PanelWith("Island", PanelOptions{X: 20, Y: 20, Width: Fixed(300)}, func() {
		if ui.Button("Count") {
			scene.behindClicks++
		}
		ui.TextField("Behind", &scene.behindText)
	})
	if !scene.open {
		return
	}
	if ui.Sheet("Confirm", SheetOptions{}, func() {
		ui.TextField("Reason", &scene.sheetText)
		if ui.PrimaryButton("Remove") {
			scene.sheetClick++
		}
	}) {
		scene.dismissals++
		scene.open = false
	}
}

func TestASheetTakesTheMouseAndTheKeyboard(t *testing.T) {
	h := newHarness(t)
	scene := &sheetScene{}
	h.frame(scene.build)
	button := h.rect(scoped("Island", "Count"))
	scene.open = true
	h.frame(scene.build)
	h.moveTo(700, 500) // over nothing but the backdrop
	h.frame(scene.build)
	if !h.ui.WantsMouse() || !h.ui.WantsKeyboard() {
		t.Fatal("a sheet claims the mouse and the keyboard, so cameras and shortcuts leave them alone")
	}
	sheetLayer, panelLayer := h.ui.layerOf(idOf("Confirm")), h.ui.layerOf(idOf("Island"))
	if sheetLayer <= panelLayer || sheetLayer >= popupLayer {
		t.Fatalf("a sheet is drawn over the panels and under the popups: %d", sheetLayer)
	}
	card := h.rect(childID(idOf("Confirm"), "##card"))
	if card.Width != sheetWidth || card.X != (800-sheetWidth)/2 {
		t.Fatalf("the card is centred in the window: %+v", card)
	}
	h.clickInside(scoped("Confirm", "Remove"), scene.build)
	if scene.sheetClick != 1 || scene.dismissals != 0 {
		t.Fatal("buttons inside the sheet work, and a click on the card does not dismiss it")
	}
	h.clickAt(button.X+button.Width/2, button.Y+button.Height/2, scene.build)
	h.frame(scene.build)
	if scene.behindClicks != 0 || scene.dismissals != 1 {
		t.Fatal("a click on a button behind the sheet is a click on the backdrop: it dismisses, the button does not react")
	}
}

func TestASheetIsDismissedWithEscapeOrTheBackdrop(t *testing.T) {
	h := newHarness(t)
	scene := &sheetScene{open: true}
	h.frame(scene.build)
	h.frame(scene.build)
	h.key(input.KeyEscape, scene.build)
	if scene.dismissals != 1 || scene.open {
		t.Fatal("Escape dismisses the sheet")
	}
	scene.open = true
	h.frame(scene.build)
	h.frame(scene.build)
	h.clickAt(5, 590, scene.build)
	if scene.dismissals != 2 {
		t.Fatal("a click on the dimmed backdrop dismisses the sheet")
	}
	h.frame(scene.build)
	if h.ui.WantsKeyboard() {
		t.Fatal("once closed, the sheet lets go of the keyboard")
	}
}

func TestEscapeInAFieldOfASheetLeavesTheField(t *testing.T) {
	h := newHarness(t)
	scene := &sheetScene{open: true}
	h.frame(scene.build)
	h.frame(scene.build)
	h.clickInside(scoped("Confirm", "Reason"), scene.build)
	h.typeText("Storm", scene.build)
	h.key(input.KeyEscape, scene.build)
	if scene.dismissals != 0 || scene.sheetText != "" {
		t.Fatalf("the first Escape is the field's (its text goes back): %q, %d dismissals", scene.sheetText, scene.dismissals)
	}
	h.key(input.KeyEscape, scene.build)
	if scene.dismissals != 1 {
		t.Fatal("the next Escape dismisses the sheet")
	}
}

func TestOpeningASheetTakesTheFocusFromAFieldBehind(t *testing.T) {
	h := newHarness(t)
	scene := &sheetScene{}
	h.frame(scene.build)
	h.clickInside(scoped("Island", "Behind"), scene.build)
	if h.ui.focus != scoped("Island", "Behind") {
		t.Fatal("the field behind has the focus before the sheet opens")
	}
	scene.open = true
	h.frame(scene.build)
	h.typeText("x", scene.build)
	if scene.behindText != "" {
		t.Fatalf("letters typed while a sheet is open do not reach a field behind: %q", scene.behindText)
	}
	h.clickInside(scoped("Confirm", "Reason"), scene.build)
	h.key(input.KeyTab, scene.build)
	if h.ui.focus != scoped("Confirm", "Reason") {
		t.Fatal("Tab goes round the sheet's own fields only")
	}
}
