package window

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
)

func TestATapIsAClickWhereTheFingerWas(t *testing.T) {
	state := input.NewState()
	var pointer touchPointer
	pointer.begin(state, 1, 300, 600, 3)
	if !state.ButtonPressed(input.MouseLeft) || state.CursorX != 100 || state.CursorY != 200 {
		t.Fatalf("pressed %v at %v, %v; want a press at 100, 200 layout pixels", state.ButtonPressed(input.MouseLeft), state.CursorX, state.CursorY)
	}
	pointer.end(state, 1, 303, 600, 3)
	if !state.ButtonReleased(input.MouseLeft) || state.CursorX != 101 {
		t.Fatalf("released %v at %v; want a release at 101", state.ButtonReleased(input.MouseLeft), state.CursorX)
	}
}

func TestASecondFingerIsIgnored(t *testing.T) {
	state := input.NewState()
	var pointer touchPointer
	pointer.begin(state, 1, 100, 100, 1)
	pointer.begin(state, 2, 500, 500, 1)
	pointer.move(state, 2, 600, 600, 1)
	if state.CursorX != 100 || state.CursorY != 100 {
		t.Fatalf("cursor at %v, %v; the second finger should not move it", state.CursorX, state.CursorY)
	}
	pointer.end(state, 2, 600, 600, 1)
	if !state.Button(input.MouseLeft) {
		t.Fatal("lifting the second finger released the first one's press")
	}
	pointer.move(state, 1, 140, 100, 1)
	if state.CursorX != 140 {
		t.Fatalf("cursor at %v; the first finger should still drag", state.CursorX)
	}
}

func TestDensityScaleIsAndroidDensityIndependentPixels(t *testing.T) {
	// a 480 dots-per-inch screen is 480/72 pixels per point and three pixels per layout pixel
	if scale := densityScale(480.0 / 72); scale < 2.999 || scale > 3.001 {
		t.Fatalf("scale = %v, want 3", scale)
	}
	if densityScale(0) != 1 {
		t.Fatal("an unknown density should draw at one pixel per layout pixel")
	}
}
