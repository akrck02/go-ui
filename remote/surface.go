// Package remote builds an interface in one program and shows it in another. A Surface runs a
// build function's frame off any window and gives back what it drew as a Frame; the program that
// owns the window adds that frame to its own drawing with Show, and sends back the pointer and the
// keys as Input. A plugin draws its screen this way inside its host's window.
//
// Only what can travel as plain data travels: rectangles, shadows, text and icons. Pictures
// (images, and the lines of paths, which go-renderer draws as pictures) stay behind.
package remote

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// Input is the state of the pointer and the keys over a surface for one frame, in the surface's
// own layout pixels, its top left corner at 0, 0.
type Input struct {
	PointerX    float32     `json:"pointerX"`
	PointerY    float32     `json:"pointerY"`
	PointerDown bool        `json:"pointerDown"` // the main button is held
	Wheel       float32     `json:"wheel"`       // what the wheel turned since the last frame
	Pressed     []input.Key `json:"pressed"`
	Text        string      `json:"text"` // the characters typed since the last frame
}

// TODO(akrck02): send the frame's accessibility tree too, so a screen reader reads a remote
// surface as it reads the host.

// Frame is what a surface drew in one frame. Again tells that it wants another frame soon, for
// something still moving, even with no new input.
type Frame struct {
	Commands []draw2d.Command `json:"commands"`
	Again    bool             `json:"again"`
}

// Surface draws a build function into frames, with no window of its own.
type Surface struct {
	context *ui.Context
	state   *input.State
	build   func(context *ui.Context)
}

// NewSurface is a surface that draws build, measuring text with measurer: give it the measurer of
// the fonts the host draws in (ui.New's, for go-renderer's default fonts), so the lines fit.
func NewSurface(build func(context *ui.Context), measurer draw2d.Measurer) *Surface {
	return &Surface{context: ui.NewWithMeasurer(measurer), state: input.NewState(), build: build}
}

// Draw runs one frame of the surface at a size, in a theme, with the input given, and gives back
// what it drew.
func (surface *Surface) Draw(width, height float32, theme ui.Theme, given Input) Frame {
	state := surface.state
	state.WindowWidth, state.WindowHeight = int(width), int(height)
	state.FramebufferWidth, state.FramebufferHeight = int(width), int(height)
	feed := state.Feed()
	feed.MoveCursor(float64(given.PointerX), float64(given.PointerY))
	feed.SetButton(input.MouseLeft, given.PointerDown)
	feed.AddScroll(float64(given.Wheel))
	for _, key := range given.Pressed {
		feed.SetKey(key, true)
	}
	for _, character := range given.Text {
		feed.AddCharacter(character)
	}
	surface.context.Theme = theme
	surface.context.Begin(state)
	surface.build(surface.context)
	surface.context.End()
	frame := Frame{Commands: plain(surface.context.Drawing().Commands)}
	_, frame.Again = feed.TakeFrameRequest()
	for _, key := range given.Pressed {
		feed.SetKey(key, false) // a key is pressed for one frame: the host sends it once
	}
	feed.EndFrame()
	return frame
}

// plain copies the commands that can travel as data, leaving pictures behind.
// TODO(akrck02): send pictures once each, by a name, so images and path lines travel too.
func plain(commands []draw2d.Command) []draw2d.Command {
	kept := make([]draw2d.Command, 0, len(commands))
	for _, command := range commands {
		if command.Picture != nil || command.Kind == draw2d.ImageKind || command.Kind == draw2d.SpriteLightKind {
			continue
		}
		kept = append(kept, command)
	}
	return kept
}
