package ui

import (
	"testing"

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
