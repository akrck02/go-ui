package ui

// SetFullscreen asks for the window on the whole screen, or in a window again: from a shortcut
// (F11, as browsers use) or a button. The window applies it after the frame; a page applies it at
// the next key or click, as browsers require. A window with no screen to fill stays a window, and
// Fullscreen then answers false.
func (ui *Context) SetFullscreen(on bool) {
	if ui.state != nil {
		ui.state.SetFullscreen(on)
	}
}

// Fullscreen tells whether the window is asked to be on the whole screen.
func (ui *Context) Fullscreen() bool { return ui.state != nil && ui.state.Fullscreen() }
