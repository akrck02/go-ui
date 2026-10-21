package uitest

import "github.com/akrck02/go-renderer/input"

// PressKey presses a key and lets it go: one frame with it down, one with it up.
func (window *Window) PressKey(key input.Key) {
	window.HoldKey(key)
	window.ReleaseKey(key)
}

// HoldKey puts a key down and runs a frame; it stays down until ReleaseKey.
func (window *Window) HoldKey(key input.Key) {
	window.state.SetKey(key, true)
	window.Frame()
}

// ReleaseKey lets a key go and runs a frame.
func (window *Window) ReleaseKey(key input.Key) {
	window.state.SetKey(key, false)
	window.Frame()
}

// Tab moves the focus to the next control, as the Tab key does.
func (window *Window) Tab() { window.PressKey(input.KeyTab) }

// ShiftTab moves the focus to the control before, as Shift+Tab does.
func (window *Window) ShiftTab() {
	window.state.SetKey(input.KeyShift, true)
	window.PressKey(input.KeyTab)
	window.ReleaseKey(input.KeyShift)
}

// Type types a text in one frame, as a paste or a fast typist does: the control with the focus
// takes it.
func (window *Window) Type(text string) {
	for _, character := range text {
		window.state.AddCharacter(character)
	}
	window.Frame()
}

// TypeOneByOne types a text one character a frame, as a person types.
func (window *Window) TypeOneByOne(text string) {
	for _, character := range text {
		window.state.AddCharacter(character)
		window.Frame()
	}
}
