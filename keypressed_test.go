package ui_test

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

func TestAKeyIsPressedOnlyInTheStepItGoesDown(test *testing.T) {
	steps := 0
	window, err := uitest.New(func(context *ui.Context) {
		if context.KeyPressed(input.KeyF11) {
			steps++
		}
	}, uitest.Size(400, 300))
	if err != nil {
		test.Fatal(err)
	}
	window.HoldKey(input.KeyF11)
	window.Frame()
	window.ReleaseKey(input.KeyF11)
	if steps != 1 {
		test.Fatalf("F11 held over several frames was pressed in %d of them, want 1", steps)
	}
}
