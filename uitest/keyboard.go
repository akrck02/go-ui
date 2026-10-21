package uitest

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// PressKey presses a key and lets it go: one frame with it down, one with it up.
func (window *Window) PressKey(key input.Key) {
	defer window.doing("Press a key", draw2d.Rect{}, false)()
	window.HoldKey(key)
	window.ReleaseKey(key)
}

// HoldKey puts a key down and runs a frame; it stays down until ReleaseKey.
func (window *Window) HoldKey(key input.Key) {
	defer window.doing("Hold a key", draw2d.Rect{}, false)()
	window.state.SetKey(key, true)
	window.step()
}

// ReleaseKey lets a key go and runs a frame.
func (window *Window) ReleaseKey(key input.Key) {
	defer window.doing("Let a key go", draw2d.Rect{}, false)()
	window.state.SetKey(key, false)
	window.step()
}

// Tab moves the focus to the next control, as the Tab key does.
func (window *Window) Tab() {
	defer window.doing("Tab", draw2d.Rect{}, false)()
	window.PressKey(input.KeyTab)
}

// ShiftTab moves the focus to the control before, as Shift+Tab does.
func (window *Window) ShiftTab() {
	defer window.doing("Shift+Tab", draw2d.Rect{}, false)()
	window.state.SetKey(input.KeyShift, true)
	window.PressKey(input.KeyTab)
	window.ReleaseKey(input.KeyShift)
}

// Type types a text in one frame, as a paste or a fast typist does: the control with the focus
// takes it.
func (window *Window) Type(text string) {
	defer window.doing(fmt.Sprintf("Type %q", text), draw2d.Rect{}, false)()
	for _, character := range text {
		window.state.AddCharacter(character)
	}
	window.step()
}

// TypeOneByOne types a text one character a frame, as a person types.
func (window *Window) TypeOneByOne(text string) {
	defer window.doing(fmt.Sprintf("Type %q", text), draw2d.Rect{}, false)()
	for _, character := range text {
		window.state.AddCharacter(character)
		window.step()
	}
}
