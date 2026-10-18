package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
)

// lighthouseControls is a small interface over a scene: a button, a check box and a field.
func lighthouseControls(lit *bool, keeper *string, pressed *int) func(ui *Context) {
	return func(ui *Context) {
		ui.Column(func() {
			if ui.Button("Ring the bell") {
				*pressed++
			}
			ui.Checkbox("Lamp lit", lit)
			ui.TextField("Keeper", keeper)
		})
	}
}

func TestAGameKeepsTabAndSpaceWhileTheInterfaceDoesNotHaveTheKeyboard(t *testing.T) {
	h := newHarness(t)
	lit, keeper, pressed := false, "", 0
	build := lighthouseControls(&lit, &keeper, &pressed)
	h.frame(build)
	h.key(input.KeyTab, build)
	h.key(input.KeySpace, build)
	h.key(input.KeyEnter, build)
	if h.ui.keyboardFocus != 0 || pressed != 0 || lit {
		t.Fatal("with the keyboard not given, Tab, Space and Enter belong to the program")
	}
	if h.ui.WantsKeyboard() {
		t.Fatal("the interface claims a keyboard it was not given")
	}
}

func TestAClickOnAButtonDoesNotTakeTheKeyboard(t *testing.T) {
	h := newHarness(t)
	lit, keeper, pressed := false, "", 0
	build := lighthouseControls(&lit, &keeper, &pressed)
	h.frame(build)
	h.clickInside(idOf("Ring the bell"), build)
	h.key(input.KeySpace, build)
	if pressed != 1 {
		t.Fatalf("Space after a click pressed the button again: %d presses", pressed)
	}
	if h.ui.HasKeyboard() {
		t.Fatal("a click on a button gave the interface the keyboard")
	}
}

func TestTabVisitsEveryControlInTheOrderItIsDrawn(t *testing.T) {
	h := newHarness(t)
	h.ui.KeyboardNavigation = Always
	lit, keeper, pressed := false, "", 0
	build := lighthouseControls(&lit, &keeper, &pressed)
	h.frame(build)
	h.key(input.KeyTab, build)
	if h.ui.keyboardFocus != idOf("Ring the bell") {
		t.Fatal("the first Tab goes to the first control")
	}
	h.key(input.KeyTab, build)
	if h.ui.keyboardFocus != idOf("Lamp lit") {
		t.Fatal("Tab goes on to the check box")
	}
	h.key(input.KeyTab, build)
	if h.ui.focus != idOf("Keeper") || h.ui.keyboardFocus != 0 {
		t.Fatal("Tab goes on to the field, which takes the keys")
	}
	h.key(input.KeyTab, build)
	if h.ui.keyboardFocus != idOf("Ring the bell") || h.ui.focus != 0 {
		t.Fatal("Tab from the last control goes round to the first")
	}
	h.key(input.KeyTab, build, input.KeyShift)
	if h.ui.focus != idOf("Keeper") {
		t.Fatal("Shift+Tab goes back")
	}
}

func TestTheKeyboardUsesTheFocusedControl(t *testing.T) {
	h := newHarness(t)
	h.ui.KeyboardNavigation = Always
	lit, keeper, pressed := false, "", 0
	build := lighthouseControls(&lit, &keeper, &pressed)
	h.frame(build)
	h.key(input.KeyTab, build)
	h.key(input.KeyEnter, build)
	h.key(input.KeySpace, build)
	if pressed != 2 {
		t.Fatalf("Enter and Space press the focused button: %d presses", pressed)
	}
	h.key(input.KeyTab, build)
	h.key(input.KeySpace, build)
	if !lit {
		t.Fatal("Space ticks the focused check box")
	}
}

func TestTakeKeyboardGivesItAndEscapeGivesItBack(t *testing.T) {
	h := newHarness(t)
	lit, keeper, pressed := false, "", 0
	build := lighthouseControls(&lit, &keeper, &pressed)
	h.frame(build)
	h.ui.TakeKeyboard()
	h.key(input.KeyTab, build)
	if h.ui.keyboardFocus != idOf("Ring the bell") || !h.ui.WantsKeyboard() {
		t.Fatal("once given the keyboard, the interface takes Tab and claims the keys")
	}
	h.key(input.KeyEscape, build)
	if h.ui.HasKeyboard() || h.ui.keyboardFocus != 0 {
		t.Fatal("Escape gives the keyboard back to the program")
	}
	h.key(input.KeySpace, build)
	if pressed != 0 {
		t.Fatal("after Escape, Space belongs to the program again")
	}
}

func TestEscapeInADialogClosesItEvenWithAButtonFocused(t *testing.T) {
	h := newHarness(t)
	open, closed := true, false
	build := func(ui *Context) {
		if !open {
			return
		}
		if ui.Modal("Leave", LayerOptions{}, func() { ui.Button("Stay") }) {
			open, closed = false, true
		}
	}
	h.frame(build)
	h.key(input.KeyTab, build)
	if h.ui.keyboardFocus != scoped("Leave", "Stay") {
		t.Fatal("an open dialog has the keyboard, so Tab reaches its button")
	}
	h.key(input.KeyEscape, build)
	if !closed {
		t.Fatal("Escape closes the dialog even while its button has the focus")
	}
}

func TestAFrameWithFocusAllocatesNothingMore(t *testing.T) {
	allocations := func(navigation KeyboardNavigation) float64 {
		h := newHarness(t)
		h.ui.KeyboardNavigation = navigation
		lit, keeper, pressed := false, "", 0
		build := lighthouseControls(&lit, &keeper, &pressed)
		h.frame(build)
		h.key(input.KeyTab, build)
		h.frame(build)
		return testing.AllocsPerRun(100, func() { h.frame(build) })
	}
	without, with := allocations(WhenGiven), allocations(Always)
	if with > without {
		t.Fatalf("a frame with a control focused allocates %v times, %v without", with, without)
	}
}
