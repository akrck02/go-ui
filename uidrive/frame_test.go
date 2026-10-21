package uidrive_test

import (
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// liveInterface is a made-up window loop for these tests: an interface run frame by frame on an
// input state, as a real window runs it.
type liveInterface struct {
	context *ui.Context
	state   *input.State
	build   func(context *ui.Context)
}

// newLiveInterface makes a loop of 800x600 layout pixels for an interface function.
func newLiveInterface(build func(context *ui.Context)) *liveInterface {
	live := &liveInterface{context: ui.NewWithMeasurer(uitest.FixedMeasurer{}), state: input.NewState(), build: build}
	live.state.WindowWidth, live.state.WindowHeight = 800, 600
	live.state.FramebufferWidth, live.state.FramebufferHeight = 800, 600
	live.context.SetScreenReader(true)
	return live
}

// frame runs one frame.
func (live *liveInterface) frame() {
	live.context.Begin(live.state)
	live.build(live.context)
	live.context.End()
	live.state.EndFrame()
}
