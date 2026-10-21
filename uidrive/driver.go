// Package uidrive drives a Go UI interface by the names people see: it finds what the last frame
// showed ("Add an entry", "Light the lamp"), makes the input a person would give it, and plays a
// script of such steps frame by frame. A real window runs a driver through window.Options.Driver
// (a demo, a tutorial, a replay of a reported bug, a test on the device); uitest runs the same
// code in go test, with no display.
package uidrive

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// Driver is what a window runs at each frame when it is driven. Drive comes after the window
// gathered the frame's real input and before the interface runs: it may add input to the state,
// and read what the last frame showed through the context. Show comes after the interface ran,
// and may draw over what it drew.
type Driver interface {
	Drive(state *input.State, context *ui.Context)
	Show(drawing *draw2d.List)
}
