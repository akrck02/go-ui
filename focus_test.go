package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
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
	if !h.ui.HasKeyboard() {
		t.Fatal("the frame Escape was used in still counts as the interface's")
	}
	h.frame(build)
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

// allocationsBeforeFocus is what a still frame of lighthouseControls allocated before keyboard
// focus existed (measured on the commit that proposed it): one, in the text field.
const allocationsBeforeFocus = 1

func TestAFrameWithFocusAllocatesNothingMore(t *testing.T) {
	for _, navigation := range []KeyboardNavigation{WhenGiven, Always} {
		h := newHarness(t)
		h.ui.KeyboardNavigation = navigation
		lit, keeper, pressed := false, "", 0
		build := lighthouseControls(&lit, &keeper, &pressed)
		h.frame(build)
		h.key(input.KeyTab, build)
		h.frame(build)
		if allocations := testing.AllocsPerRun(100, func() { h.frame(build) }); allocations > allocationsBeforeFocus {
			t.Fatalf("a still frame of a few controls, one focused, allocates %v times, %v before", allocations, allocationsBeforeFocus)
		}
	}
}

// ringAround finds the stroke the focus ring draws: a rectangle with no fill and a border.
func (h *harness) ringAround(colour draw2d.Color) (draw2d.Command, bool) {
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.RectangleKind && !command.Fill.IsVisible() && command.Border == colour && command.BorderWidth > 0 {
			return command, true
		}
	}
	return draw2d.Command{}, false
}

func TestAClickGivesFocusWithoutARing(t *testing.T) {
	h := newHarness(t)
	h.ui.KeyboardNavigation = Always
	lit, keeper, pressed := false, "", 0
	build := lighthouseControls(&lit, &keeper, &pressed)
	h.frame(build)
	h.clickInside(idOf("Ring the bell"), build)
	if _, found := h.ringAround(h.ui.Theme.FocusRing); found {
		t.Fatal("a click shows no focus ring")
	}
	h.key(input.KeyTab, build)
	h.frame(build) // the frame Tab asks for
	if _, found := h.ringAround(h.ui.Theme.FocusRing); !found {
		t.Fatal("Tab shows the focus ring")
	}
}

func TestTheRingLiesOnTheEdge(t *testing.T) {
	h := newHarness(t)
	h.ui.KeyboardNavigation = Always
	h.ui.Theme.FocusWidth = 2
	lit, keeper, pressed := false, "", 0
	build := lighthouseControls(&lit, &keeper, &pressed)
	h.frame(build)
	h.key(input.KeyTab, build)
	h.frame(build)
	ring, found := h.ringAround(h.ui.Theme.FocusRing)
	button := h.rect(idOf("Ring the bell"))
	grown := draw2d.Rect{X: button.X - 2, Y: button.Y - 2, Width: button.Width + 4, Height: button.Height + 4}
	if !found || ring.Area != grown || ring.BorderWidth != 2 {
		t.Fatalf("the ring is %v wide around %v, want 2 around %v", ring.BorderWidth, ring.Area, grown)
	}
}

func TestTheRingOfAPrimaryButtonIsInside(t *testing.T) {
	h := newHarness(t)
	h.ui.KeyboardNavigation = Always
	build := func(ui *Context) { ui.PrimaryButton("Light the lamp") }
	h.frame(build)
	h.key(input.KeyTab, build)
	h.frame(build)
	ring, found := h.ringAround(h.ui.Theme.OnPrimary)
	if !found || ring.Area != h.rect(idOf("Light the lamp")) {
		t.Fatal("on a primary button the ring is drawn inside it, in the colour of its text")
	}
}

func TestAFieldCanLightItsOwnBorder(t *testing.T) {
	h := newHarness(t)
	h.ui.Theme.FieldFocusEdge, h.ui.Theme.FocusWidth = true, 2
	keeper := ""
	build := func(ui *Context) { ui.TextField("Keeper", &keeper) }
	h.frame(build)
	h.clickInside(idOf("Keeper"), build)
	ring, found := h.ringAround(h.ui.Theme.FocusRing)
	if !found || ring.Area != h.rect(idOf("Keeper")) || ring.BorderWidth != 2 {
		t.Fatalf("the field's own edge lights up, with no ring around it: %+v", ring)
	}
}

// holdOn presses a control and keeps the button down for a frame.
func (h *harness) holdOn(id ID, build func(ui *Context)) {
	area := h.rect(id)
	h.moveTo(float64(area.X+area.Width/2), float64(area.Y+area.Height/2))
	h.frame(build)
	h.press()
	h.frame(build)
	h.frame(build)
}

// fillOf finds the rectangle a button fills, by its colour.
func (h *harness) fillOf(colour draw2d.Color) draw2d.Rect {
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.RectangleKind && command.Fill == colour {
			return command.Area
		}
	}
	return draw2d.Rect{}
}

func TestAHeldButtonShrinks(t *testing.T) {
	h := newHarness(t)
	h.ui.Theme.PressScale = 0.9
	build := func(ui *Context) { ui.Button("Ring the bell") }
	h.frame(build)
	button := h.rect(idOf("Ring the bell"))
	h.holdOn(idOf("Ring the bell"), build)
	drawn := h.fillOf(h.ui.Theme.Line) // a secondary button held is filled with Line
	if drawn.Width >= button.Width || drawn.Width < button.Width*0.89 {
		t.Fatalf("a held button is drawn at 0.9 of its width: %v of %v", drawn.Width, button.Width)
	}
	if h.rect(idOf("Ring the bell")) != button {
		t.Fatal("shrinking is only drawn: the layout does not move")
	}
}

func TestLessMotionKeepsAHeldButtonStill(t *testing.T) {
	h := newHarness(t)
	h.ui.Theme.PressScale, h.ui.LessMotion = 0.9, true
	build := func(ui *Context) { ui.Button("Ring the bell") }
	h.frame(build)
	button := h.rect(idOf("Ring the bell"))
	h.holdOn(idOf("Ring the bell"), build)
	if drawn := h.fillOf(h.ui.Theme.Line); drawn != button {
		t.Fatalf("with less motion a held button keeps its size: %v, want %v", drawn, button)
	}
}

// TestThePauseMenuOfTheGuideWorks runs the game of docs/5-components/10-the-keyboard.md.
func TestThePauseMenuOfTheGuideWorks(t *testing.T) {
	h := newHarness(t)
	paused, jumps := false, 0
	step := func() {
		if !paused && h.state.Pressed(input.KeyP) {
			paused = true
			h.ui.TakeKeyboard()
		}
		h.ui.Begin(h.state)
		if paused {
			h.ui.Panel("Paused", func() {
				if h.ui.PrimaryButton("Resume") {
					paused = false
					h.ui.ReleaseKeyboard()
				}
				h.ui.Button("Settings")
			})
		}
		h.ui.End()
		if paused && !h.ui.HasKeyboard() {
			paused = false
		}
		if !h.ui.HasKeyboard() && h.state.Pressed(input.KeySpace) {
			jumps++
		}
		h.state.EndFrame()
	}
	press := func(key input.Key) {
		h.state.SetKey(key, true)
		step()
		h.state.SetKey(key, false)
	}
	press(input.KeySpace)
	press(input.KeyP)
	press(input.KeySpace) // the menu has the keyboard: no jump
	if !paused || jumps != 1 {
		t.Fatalf("paused %v with %d jumps, want paused with 1", paused, jumps)
	}
	press(input.KeyEscape)
	press(input.KeySpace)
	if paused || jumps != 2 {
		t.Fatalf("Escape resumes and Space jumps again: paused %v, %d jumps", paused, jumps)
	}
	press(input.KeyP)
	press(input.KeyTab) // the panel's fold button comes first, in its title bar
	press(input.KeyTab)
	press(input.KeyEnter) // Tab reached Resume, Enter presses it
	if paused {
		t.Fatal("Tab and Enter press Resume")
	}
	press(input.KeyP)
	press(input.KeyTab)
	press(input.KeyTab)
	press(input.KeySpace) // Space presses Resume, and must not jump too
	if paused || jumps != 2 {
		t.Fatalf("Space on Resume resumes without a jump: paused %v, %d jumps", paused, jumps)
	}
}

func TestTabFromAFieldToAButtonKeepsTheKeyboard(t *testing.T) {
	h := newHarness(t)
	keeper, pressed := "", 0
	build := func(ui *Context) {
		ui.Column(func() {
			ui.TextField("Keeper", &keeper)
			if ui.Button("Save") {
				pressed++
			}
		})
	}
	h.frame(build)
	h.clickInside(idOf("Keeper"), build)
	h.key(input.KeyTab, build)
	h.frame(build)
	h.key(input.KeyEnter, build)
	if pressed != 1 || !h.ui.HasKeyboard() {
		t.Fatalf("Tab from a field reaches the button, and Enter presses it: %d presses", pressed)
	}
}

func TestTabGoesOnFromAControlThatWentAway(t *testing.T) {
	h := newHarness(t)
	h.ui.KeyboardNavigation = Always
	removed := false
	build := func(ui *Context) {
		ui.Column(func() {
			ui.Button("North")
			if !removed && ui.Button("Middle") {
				removed = true
			}
			ui.Button("South")
		})
	}
	h.frame(build)
	h.key(input.KeyTab, build)
	h.key(input.KeyTab, build)
	h.key(input.KeyEnter, build) // Middle removes itself
	h.frame(build)
	h.key(input.KeyTab, build)
	if h.ui.keyboardFocus != idOf("South") {
		t.Fatal("Tab goes on to the control after the one that went away")
	}
}

func TestTabGoesOnFromTheControlLastClicked(t *testing.T) {
	h := newHarness(t)
	h.ui.KeyboardNavigation = Always
	build := func(ui *Context) {
		ui.Column(func() {
			ui.Button("North")
			ui.Button("Middle")
			ui.Button("South")
		})
	}
	h.frame(build)
	h.clickInside(idOf("Middle"), build)
	h.key(input.KeyTab, build)
	if h.ui.keyboardFocus != idOf("South") {
		t.Fatal("Tab after a click goes on from the control clicked")
	}
}

func TestAHeldButtonShrinksItsLabelToo(t *testing.T) {
	h := newHarness(t)
	h.ui.Theme.PressScale = 0.5
	build := func(ui *Context) { ui.Button("Ring the bell") }
	h.frame(build)
	before, _ := h.textCommand("Ring the bell")
	h.holdOn(idOf("Ring the bell"), build)
	after, _ := h.textCommand("Ring the bell")
	if after.Size != before.Size*0.5 || after.Area.Width != before.Area.Width*0.5 {
		t.Fatalf("the label shrinks with the button: size %v from %v", after.Size, before.Size)
	}
}

func TestFocusedIDNamesTheControlTabReachedAndTheFieldBeingTyped(t *testing.T) {
	h := newHarness(t)
	h.ui.KeyboardNavigation = Always
	text := ""
	build := func(ui *Context) {
		ui.Button("Light the lamp")
		ui.TextField("note", &text)
	}
	h.frame(build)
	if h.ui.FocusedID() != 0 {
		t.Fatal("something has the focus before any key")
	}
	h.state.SetKey(input.KeyTab, true)
	h.frame(build)
	h.state.SetKey(input.KeyTab, false)
	h.frame(build)
	if h.ui.FocusedID() != idOf("Light the lamp") {
		t.Fatal("FocusedID does not name the button Tab reached")
	}
	h.ui.Focus("note")
	h.frame(build)
	if h.ui.FocusedID() != idOf("note") {
		t.Fatal("FocusedID does not name the field being typed in")
	}
}
