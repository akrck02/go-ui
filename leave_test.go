package ui

import (
	"testing"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
)

// noteLayer shows a modal with a note while shown is true, leaving as told.
func noteLayer(shown *bool, leave Leave) func(ui *Context) {
	return func(ui *Context) {
		if !*shown {
			return
		}
		options := LayerOptions{AlignX: Center, AlignY: Center, Leave: leave}
		ui.Modal("note", options, func() { ui.Label("The north light is lit") })
	}
}

// textCommand finds the note's line in the last frame's drawing.
func (h *harness) textCommand(text string) (draw2d.Command, bool) {
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.TextKind && command.Text == text {
			return command, true
		}
	}
	return draw2d.Command{}, false
}

func TestALayerLeavesFadingAndMovingAfterItIsNoLongerShown(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	shown := true
	build := noteLayer(&shown, Leave{Duration: 200 * time.Millisecond, Y: 100})
	h.frame(build)
	before, _ := h.textCommand("The north light is lit")

	shown = false
	h.frame(build)
	*now = now.Add(100 * time.Millisecond)
	h.frame(build)
	halfway, found := h.textCommand("The north light is lit")
	if !found {
		t.Fatal("the note went at once")
	}
	if halfway.Fill[3] >= before.Fill[3] || halfway.Area.Y <= before.Area.Y {
		t.Fatalf("half way the note is at %v with alpha %v; want it lower than %v and fainter than %v",
			halfway.Area.Y, halfway.Fill[3], before.Area.Y, before.Fill[3])
	}
	if _, asked := h.state.TakeFrameRequest(); !asked {
		t.Fatal("a leaving note asked for no next frame")
	}
	*now = now.Add(150 * time.Millisecond)
	h.frame(build)
	if _, found := h.textCommand("The north light is lit"); found {
		t.Fatal("the note is still there after it had time to leave")
	}
}

func TestALayerWithNoLeaveGoesAtOnce(t *testing.T) {
	h := newHarness(t)
	h.withClock()
	shown := true
	build := noteLayer(&shown, Leave{})
	h.frame(build)
	shown = false
	h.frame(build)
	if _, found := h.textCommand("The north light is lit"); found {
		t.Fatal("a layer that does not leave with motion is still drawn")
	}
}

func TestALayerShownAgainIsDrawnOnce(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	shown := true
	build := noteLayer(&shown, Leave{Duration: 200 * time.Millisecond})
	h.frame(build)
	shown = false
	h.frame(build)
	*now = now.Add(50 * time.Millisecond)
	shown = true
	h.frame(build)
	count := 0
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.TextKind && command.Text == "The north light is lit" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("the note shown again is drawn %d times; want once", count)
	}
}
