package ui

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
)

func TestRowsOutOfSightAreNotPainted(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.ScrollArea("rows", Layout{Width: Fixed(300), Height: Fixed(200)}, func() {
			for index := 0; index < 500; index++ {
				ui.Push(strconv.Itoa(index))
				ui.Frame(Layout{Width: Grow(), Height: Fixed(20)}, Look{Fill: draw2d.MustHex("#ffffff")}, func() {
					ui.Label(fmt.Sprintf("Row %d", index))
				})
				ui.Pop()
			}
		})
	}
	h.frame(build)
	h.frame(build)
	texts := 0
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.TextKind {
			texts++
		}
	}
	if texts == 0 || texts > 12 {
		t.Fatalf("only the rows in sight (about 10 of 500) are painted, got %d texts", texts)
	}
}

func TestTheTextCacheKeepsWhatAFrameUses(t *testing.T) {
	measurer := &fakeMeasurer{}
	h := newHarness(t)
	h.ui = NewWithMeasurer(measurer)
	build := func(ui *Context) {
		for index := 0; index < maximumCachedTexts+500; index++ {
			ui.Label("Row " + strconv.Itoa(index))
		}
	}
	h.frame(build)
	h.frame(build)
	before := measurer.calls
	h.frame(build)
	if measurer.calls != before {
		t.Fatalf("a frame with more texts than the cache holds measures them again: %d new measures", measurer.calls-before)
	}
}

func TestTheCaretStopsBlinkingAfterAWhile(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	text := "Faro"
	build := func(ui *Context) { ui.TextField("Name", &text) }
	h.frame(build)
	h.clickInside(idOf("Name"), build)
	h.frame(build)
	if _, wanted := h.state.Feed().TakeFrameRequest(); !wanted {
		t.Fatal("a fresh caret blinks")
	}
	*now = now.Add(caretBlinksFor + time.Second)
	h.frame(build)
	if _, wanted := h.state.Feed().TakeFrameRequest(); wanted {
		t.Fatal("after a while without typing the caret stays shown and asks for no frames")
	}
	if !h.ui.memory[idOf("Name")].value.(*fieldMemory).paint.caretShown {
		t.Fatal("it stays shown, not hidden")
	}
	h.typeText("s", build)
	h.frame(build)
	if _, wanted := h.state.Feed().TakeFrameRequest(); !wanted {
		t.Fatal("typing starts the blinking again")
	}
}

func TestAFocusedFieldDoesNotMeasureWhileNothingChanges(t *testing.T) {
	measurer := &fakeMeasurer{}
	h := newHarness(t)
	h.ui = NewWithMeasurer(measurer)
	text := "Faro del Norte"
	build := func(ui *Context) { ui.TextField("Name", &text) }
	h.frame(build)
	h.clickInside(idOf("Name"), build)
	h.frame(build)
	h.frame(build)
	before := measurer.calls
	for range 5 {
		h.frame(build)
	}
	if measurer.calls != before {
		t.Fatalf("idle frames measured %d texts", measurer.calls-before)
	}
}
