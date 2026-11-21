package ui

import (
	"runtime"

	"github.com/akrck02/go-renderer/input"
)

// Shortcut tells whether a key was pressed in this step with the system's shortcut key held: Cmd
// on a Mac, Ctrl elsewhere, and either in a browser, which does not tell which system it runs on.
// With shift, Shift must be held too, and without it, not. A program's own commands answer to it:
// Shortcut(input.KeyO, true) for Ctrl+Shift+O, a new document.
func (ui *Context) Shortcut(key input.Key, shift bool) bool {
	state := ui.state
	if !state.Pressed(key) || state.Down(input.KeyShift) != shift {
		return false
	}
	if runtime.GOOS == "js" {
		return state.Down(input.KeyControl) || state.Down(input.KeySuper)
	}
	return state.Down(shortcutKey)
}

// KeyPressed tells whether a key was pressed in this step, whatever else is held: for keys that
// act on their own, as F11 does for the whole screen. Keys that type or move in a field reach the
// field too, so a program's own commands with letters use Shortcut instead.
func (ui *Context) KeyPressed(key input.Key) bool {
	return ui.state != nil && ui.state.Pressed(key)
}
