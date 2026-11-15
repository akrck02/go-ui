package ui

import (
	"runtime"
	"testing"

	"github.com/akrck02/go-renderer/input"
)

// A program's command answers to its keys with the shortcut key held, and not to the letter
// alone (typed in a field) nor with Shift missing, which may be another command.
func TestAShortcutAnswersOnlyToItsKeys(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("the shortcut key is Cmd on a Mac; this test presses Ctrl")
	}
	h := newHarness(t)
	pressed := 0
	build := func(ui *Context) {
		if ui.Shortcut(input.KeyO, true) {
			pressed++
		}
	}
	h.frame(build)
	h.key(input.KeyO, build)
	h.key(input.KeyO, build, input.KeyControl)
	h.key(input.KeyO, build, input.KeyShift)
	if pressed != 0 {
		t.Fatalf("O without Ctrl and Shift together ran the command %d times", pressed)
	}
	h.key(input.KeyO, build, input.KeyControl, input.KeyShift)
	if pressed != 1 {
		t.Errorf("Ctrl+Shift+O ran the command %d times", pressed)
	}
}
