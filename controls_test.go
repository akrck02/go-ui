package ui

import (
	"math"
	"testing"
	"time"

	"github.com/akrck02/go-renderer/input"
)

// clickAt moves to a point inside an area, presses and lets go in one step, and runs the frame.
func (h *harness) clickInside(id ID, build func(ui *Context)) {
	h.t.Helper()
	area := h.rect(id)
	h.moveTo(float64(area.X+area.Width/2), float64(area.Y+area.Height/2))
	h.frame(build)
	h.press()
	h.release()
	h.frame(build)
}

func scoped(scope string, label string) ID { return hashString(uint64(idOf(scope)), label) }

func TestCheckboxAndToggleFlip(t *testing.T) {
	h := newHarness(t)
	ticked, switched, changes := false, true, 0
	build := func(ui *Context) {
		ui.Column(func() {
			if ui.Checkbox("Lamp lit", &ticked) {
				changes++
			}
			if ui.Toggle("Fog horn", &switched) {
				changes++
			}
		})
	}
	h.frame(build)
	h.clickInside(idOf("Lamp lit"), build)
	h.clickInside(idOf("Fog horn"), build)
	if !ticked || switched || changes != 2 {
		t.Fatalf("each click flips its value once: %v %v %d", ticked, switched, changes)
	}
}

func TestSliderFollowsThePointerAndSteps(t *testing.T) {
	h := newHarness(t)
	value := float32(0)
	build := func(ui *Context) {
		ui.Box(Layout{Width: Fixed(216)}, func() { ui.SliderStep("range", &value, 0, 100, 5) })
	}
	h.frame(build)
	track := h.rect(idOf("range"))
	// the knob's centre travels from 8 to 208: three quarters is 158
	h.moveTo(float64(track.X+158), float64(track.Y+10))
	h.frame(build)
	h.press()
	h.frame(build)
	if value != 75 {
		t.Fatalf("pressing the track jumps there: %v", value)
	}
	h.moveTo(float64(track.X+161), float64(track.Y+10))
	h.frame(build)
	if value != 75 {
		t.Fatalf("the value keeps to steps of 5: %v", value)
	}
	h.moveTo(float64(track.X+500), float64(track.Y+10))
	h.frame(build)
	if value != 100 {
		t.Fatalf("dragging past the end stops at the maximum: %v", value)
	}
	h.release()
	h.frame(build)
}

func TestNumberFieldStepsAndDrags(t *testing.T) {
	h := newHarness(t)
	value := float32(28)
	build := func(ui *Context) {
		ui.Box(Layout{Width: Fixed(200)}, func() { ui.NumberField("range", &value, 0, 30, 1, "%.0f nm") })
	}
	h.frame(build)
	plus := hashString(uint64(idOf("range")), "plus")
	h.clickInside(plus, build)
	h.clickInside(plus, build)
	h.clickInside(plus, build)
	if value != 30 {
		t.Fatalf("plus adds a step and stops at the maximum: %v", value)
	}
	middle := h.rect(childID(idOf("range"), "##value"))
	h.moveTo(float64(middle.X+20), float64(middle.Y+5))
	h.frame(build)
	h.press()
	h.frame(build)
	h.moveTo(float64(middle.X+20-9), float64(middle.Y+5)) // nine pixels left: two whole steps
	h.frame(build)
	if value != 28 {
		t.Fatalf("dragging the number moves a step every 4 pixels: %v", value)
	}
	h.release()
	h.frame(build)
}

func TestTabsSelect(t *testing.T) {
	h := newHarness(t)
	tab := 0
	build := func(ui *Context) {
		ui.Box(Layout{Width: Fixed(300)}, func() { ui.Tabs("views", &tab, "General", "Advanced") })
	}
	h.frame(build)
	h.clickInside(scoped("views", "Advanced"), build)
	if tab != 1 {
		t.Fatalf("clicking a tab selects it: %d", tab)
	}
}

func TestSelectOpensChoosesAndCloses(t *testing.T) {
	h := newHarness(t)
	colour, changes := 0, 0
	options := []string{"Amber", "White", "Red"}
	build := func(ui *Context) {
		ui.Box(Layout{Width: Fixed(200), Padding: Pad(20)}, func() {
			if ui.Select("light", &colour, options) {
				changes++
			}
		})
	}
	h.frame(build)
	h.clickInside(idOf("light"), build)
	list := childID(idOf("light"), "##list")
	h.frame(build)
	if _, open := h.ui.previous[list]; !open {
		t.Fatal("clicking the field opens its list")
	}
	field, popup := h.rect(idOf("light")), h.rect(list)
	if popup.Y < field.Bottom() || popup.Width < field.Width {
		t.Fatalf("the list opens under the field, at least as wide: %+v under %+v", popup, field)
	}
	if layer := h.ui.layerOf(list); layer < popupLayer {
		t.Fatal("the list is drawn over everything")
	}
	// options are scoped by the select's scroll area inside the list
	h.moveTo(float64(popup.X+20), float64(popup.Y+4+28*2+10))
	h.frame(build)
	h.press()
	h.release()
	h.frame(build)
	h.frame(build)
	if colour != 2 || changes != 1 {
		t.Fatalf("clicking an option chooses it: %d after %d changes", colour, changes)
	}
	if _, open := h.ui.previous[list]; open {
		t.Fatal("choosing closes the list")
	}
}

func TestSelectWithTheKeyboard(t *testing.T) {
	h := newHarness(t)
	colour := 0
	build := func(ui *Context) {
		ui.Box(Layout{Width: Fixed(200)}, func() { ui.Select("light", &colour, []string{"Amber", "White", "Red"}) })
	}
	h.frame(build)
	h.clickInside(idOf("light"), build)
	if !h.ui.WantsKeyboard() {
		t.Fatal("an open select takes the keyboard")
	}
	h.state.SetKey(input.KeyDown, true)
	h.frame(build)
	h.state.SetKey(input.KeyDown, false)
	h.state.SetKey(input.KeyEnter, true)
	h.frame(build)
	h.state.SetKey(input.KeyEnter, false)
	h.frame(build)
	if colour != 1 || h.ui.WantsKeyboard() {
		t.Fatalf("down and enter choose the next option and close: %d", colour)
	}
}

func TestAClickElsewhereClosesASelect(t *testing.T) {
	h := newHarness(t)
	colour := 0
	build := func(ui *Context) {
		ui.Box(Layout{Width: Fixed(200)}, func() { ui.Select("light", &colour, []string{"Amber", "White"}) })
	}
	h.frame(build)
	h.clickInside(idOf("light"), build)
	h.moveTo(700, 500)
	h.frame(build)
	h.press()
	h.release()
	h.frame(build)
	h.frame(build)
	if _, open := h.ui.previous[childID(idOf("light"), "##list")]; open || colour != 0 {
		t.Fatal("a click outside closes the list without choosing")
	}
}

func TestScrollAreaScrollsWithTheWheelAndTheThumb(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.ScrollArea("log", Layout{Width: Fixed(200), Height: Fixed(100)}, func() {
			for index := range 10 {
				ui.frameWithID(idOf("line"+string(rune('0'+index))), Layout{Width: Grow(), Height: Fixed(30)}, Look{}, nil)
			}
		})
	}
	h.frame(build)
	h.moveTo(50, 50)
	h.frame(build)
	h.state.AddScroll(-1) // one notch down
	h.frame(build)
	h.frame(build)
	if first := h.rect(idOf("line0")); first.Y != -40 {
		t.Fatalf("a notch scrolls the content up by the theme's step: %+v", first)
	}
	h.state.AddScroll(-100)
	h.frame(build)
	h.frame(build)
	if last := h.rect(idOf("line9")); last.Bottom() != 100 {
		t.Fatalf("scrolling stops when the end shows: %+v", last)
	}
	thumb := h.rect(childID(idOf("log"), "##thumb"))
	h.moveTo(float64(thumb.X+thumb.Width/2), float64(thumb.Y+thumb.Height/2))
	h.frame(build)
	h.press()
	h.frame(build)
	h.moveTo(float64(thumb.X+thumb.Width/2), float64(thumb.Y+thumb.Height/2)-1000)
	h.frame(build)
	h.frame(build)
	if first := h.rect(idOf("line0")); first.Y != 0 {
		t.Fatalf("dragging the thumb to the top shows the start: %+v", first)
	}
	h.release()
	h.frame(build)
}

func TestListBuildsOnlyTheVisibleRows(t *testing.T) {
	h := newHarness(t)
	built := 0
	build := func(ui *Context) {
		built = 0
		ui.List("rows", Layout{Width: Fixed(200), Height: Fixed(300)}, 10000, 30, func(index int) {
			built++
			ui.Label("row")
		})
	}
	h.frame(build)
	h.frame(build)
	if built > 12 {
		t.Fatalf("only the rows that fit are built, got %d", built)
	}
	memory := remember(h.ui, idOf("rows"), func() *scrollMemory { return nil })
	if memory.content != 300000 {
		t.Fatalf("the list is as tall as all its rows: %v", memory.content)
	}
}

func TestGridPlacesChildrenInRows(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.Box(Layout{Width: Fixed(320), Columns: 3, Gap: 10}, func() {
			for index := range 5 {
				height := float32(20)
				if index == 1 {
					height = 40
				}
				ui.frameWithID(idOf("cell"+string(rune('0'+index))), Layout{Width: Grow(), Height: Fixed(height)}, Look{}, nil)
			}
		})
	})
	cell := float32(100) // (320 - 2 gaps of 10) / 3
	if second := h.rect(idOf("cell1")); second.X != cell+10 || second.Width != cell {
		t.Fatalf("cells are equal: %+v", second)
	}
	if fourth := h.rect(idOf("cell3")); fourth.X != 0 || fourth.Y != 40+10 {
		t.Fatalf("the second row starts under the tallest cell of the first: %+v", fourth)
	}
}

func TestTooltipWaitsForThePointerToRest(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	build := func(ui *Context) {
		ui.Button("Save")
		ui.Tooltip("Save the register")
	}
	h.frame(build)
	area := h.rect(idOf("Save"))
	h.moveTo(float64(area.X+5), float64(area.Y+5))
	tooltip := childID(idOf("Save"), "##tooltip")
	for step := 0; step < 5; step++ {
		h.frame(build)
		*now = now.Add(80 * time.Millisecond)
		if _, shown := h.ui.previous[tooltip]; shown {
			t.Fatalf("the tooltip waits; it showed after %d steps", step)
		}
	}
	if delay, wanted := h.state.TakeFrameRequest(); !wanted || delay > h.ui.Theme.TooltipDelay {
		t.Fatalf("it asks for a frame when it is due: %v, %v", delay, wanted)
	}
	*now = now.Add(h.ui.Theme.TooltipDelay)
	h.frame(build)
	h.frame(build)
	if _, shown := h.ui.previous[tooltip]; !shown {
		t.Fatal("after resting, the tooltip shows")
	}
	h.moveTo(700, 500)
	h.frame(build)
	h.frame(build)
	if _, shown := h.ui.previous[tooltip]; shown {
		t.Fatal("moving away hides it")
	}
}

func TestSnap(t *testing.T) {
	if snap(7.4, 0, 10, 0.5) != 7.5 || snap(-3, 0, 10, 0) != 0 || math.IsNaN(float64(snap(5, 0, 10, 0))) {
		t.Fatal("snap rounds to steps and clamps")
	}
}
