package ui_test

import (
	"testing"

	ui "github.com/akrck02/go-ui"
)

func TestButtonClicksOnRelease(t *testing.T) {
	clicks := 0
	window := openWindow(t, func(context *ui.Context) {
		if context.Button("Switch on") {
			clicks++
		}
	})
	x, y := middleOf(t, window, "Switch on")
	window.MoveTo(x, y)
	window.Press()
	if clicks != 0 {
		t.Fatal("pressing clicked before the button was let go")
	}
	window.Release()
	if clicks != 1 {
		t.Fatalf("letting go over it clicks, got %d", clicks)
	}
	window.Frame()
	if clicks != 1 {
		t.Fatal("one click per release")
	}
}

func TestLettingGoElsewhereDoesNotClick(t *testing.T) {
	clicked := false
	window := openWindow(t, func(context *ui.Context) { clicked = context.Button("Switch on") || clicked })
	x, y := middleOf(t, window, "Switch on")
	window.MoveTo(x, y)
	window.Press()
	window.MoveTo(700, 500)
	window.Release()
	if clicked {
		t.Fatal("a press dragged away and let go is not a click")
	}
}
