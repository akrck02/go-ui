package ui

import (
	"testing"
	"time"

	"github.com/akrck02/go-renderer/input"
)

func TestTheInterfaceKnowsTheWindowAndWhatCoversIt(t *testing.T) {
	context := NewWithMeasurer(&fakeMeasurer{})
	context.SetInsets(Insets{Top: 24, Bottom: 48})
	state := input.NewState()
	state.WindowWidth, state.WindowHeight = 400, 800
	context.Begin(state)
	context.End()
	width, height := context.Size()
	if width != 400 || height != 800 {
		t.Fatalf("size = %v × %v, want 400 × 800", width, height)
	}
	if insets := context.Insets(); insets.Top != 24 || insets.Bottom != 48 {
		t.Fatalf("insets = %+v, want 24 at the top and 48 at the bottom", insets)
	}
}

func TestAProgramCanAskForTheNextFrame(t *testing.T) {
	context := NewWithMeasurer(&fakeMeasurer{})
	state := input.NewState()
	context.Begin(state)
	context.RequestFrame(250 * time.Millisecond)
	context.End()
	if delay, wanted := state.Feed().TakeFrameRequest(); !wanted || delay > 250*time.Millisecond {
		t.Fatalf("frame wanted %v in %v; want one within 250 ms", wanted, delay)
	}
}

func TestTheInterfaceListsTheFingers(t *testing.T) {
	context := NewWithMeasurer(&fakeMeasurer{})
	if len(context.Touches()) != 0 {
		t.Fatal("fingers before any frame")
	}
	state := input.NewState()
	state.Feed().SetTouch(1, 10, 20)
	state.Feed().SetTouch(2, 30, 40)
	context.Begin(state)
	touches := context.Touches()
	context.End()
	if len(touches) != 2 || touches[1].X != 30 {
		t.Fatalf("touches = %+v, want the two fingers", touches)
	}
}
