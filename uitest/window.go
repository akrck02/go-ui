// Package uitest drives a Go UI interface in go test, with no display: it opens a pretend window,
// presses, types and waits in it, and tells what it shows. Every action runs whole frames, as a
// backend does, and returns an error with a code when what it looks for is not there.
package uitest

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// Window is a pretend window that runs an interface function frame by frame.
type Window struct {
	context *ui.Context
	state   *input.State
	build   func(context *ui.Context)
	clock   clock
	// measurer is the interface's, to measure texts found in the draw list (find.go)
	measurer draw2d.Measurer
	// pointerFinger is the finger that also moves the pointer (touch.go), or noFinger
	pointerFinger int64
}

// New makes a test window for an interface function: 800x600 layout pixels at scale 1, ui's
// default theme and FixedMeasurer, unless options say otherwise. No frame has run yet.
func New(build func(context *ui.Context), options ...Option) (*Window, error) {
	chosen := settings{width: defaultWidth, height: defaultHeight, scale: 1, measurer: FixedMeasurer{}}
	for _, option := range options {
		option(&chosen)
	}
	if chosen.width <= 0 || chosen.height <= 0 || chosen.scale <= 0 {
		return nil, invalidSize(chosen.width, chosen.height)
	}
	window := &Window{context: ui.NewWithMeasurer(chosen.measurer), build: build, measurer: chosen.measurer, pointerFinger: noFinger}
	window.state = newState(chosen.width, chosen.height)
	if chosen.theme != nil {
		window.context.Theme = *chosen.theme
	}
	window.context.SetScreenReader(true) // its nodes are how elements are found (find.go)
	window.clock.now = clockStart
	window.state.Now = window.clock.read
	window.resize(chosen.width, chosen.height, chosen.scale)
	return window, nil
}

// newState makes an input state for a window of a size in layout pixels, at scale 1.
func newState(width, height float32) *input.State {
	state := input.NewState()
	state.WindowWidth, state.WindowHeight = int(width), int(height)
	state.FramebufferWidth, state.FramebufferHeight = int(width), int(height)
	return state
}

// resize sets the window's and the framebuffer's sizes in the input state.
func (window *Window) resize(width, height, scale float32) {
	window.state.WindowWidth, window.state.WindowHeight = int(width), int(height)
	window.state.FramebufferWidth, window.state.FramebufferHeight = int(width*scale), int(height*scale)
}

// Resize changes the window's size in layout pixels, keeping its scale; the next frame sees it.
func (window *Window) Resize(width, height float32) error {
	if width <= 0 || height <= 0 {
		return invalidSize(width, height)
	}
	scale := float32(window.state.FramebufferWidth) / float32(window.state.WindowWidth)
	window.resize(width, height, scale)
	return nil
}

// Frame runs one frame as a backend does: Begin, the interface, End, then the end of the input
// step, which forgets this frame's presses and typed text.
func (window *Window) Frame() {
	window.context.Begin(window.state)
	window.build(window.context)
	window.context.End()
	window.state.EndFrame()
}

// Frames runs a number of frames, with no time passing between them (see Wait).
func (window *Window) Frames(count int) {
	for range count {
		window.Frame()
	}
}

// Context is the interface's context, for what this package does not cover yet.
func (window *Window) Context() *ui.Context { return window.context }

// State is the window's input state, for what this package does not cover yet.
func (window *Window) State() *input.State { return window.state }

// DrawList is what the last frame drew, as go-renderer would paint it.
func (window *Window) DrawList() *draw2d.List { return window.context.Drawing() }
