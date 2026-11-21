package ui_test

import (
	"testing"

	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

func TestTheWholeScreenIsAskedOfTheWindowUntilItIsAskedNoMore(test *testing.T) {
	asked := true
	window, err := uitest.New(func(context *ui.Context) { context.SetFullscreen(asked) }, uitest.Size(400, 300))
	if err != nil {
		test.Fatal(err)
	}
	window.Frame()
	if !window.State().Fullscreen() || !window.Context().Fullscreen() {
		test.Fatal("the window was not asked for the whole screen")
	}
	asked = false
	window.Frame()
	if window.State().Fullscreen() {
		test.Fatal("the window is still asked for the whole screen")
	}
}
