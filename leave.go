package ui

import (
	"time"

	"github.com/akrck02/go-renderer/draw2d"
)

// Layers that leave with motion. A program stops showing a dialog or a note by not drawing it,
// so nothing of it is left to move; a layer whose options say how it leaves (LayerOptions.Leave)
// keeps what it painted in its last frame, and that picture goes on fading and moving for a moment
// after the program stopped. It is only a picture: nothing in it can be touched.

// Leave is how a layer goes when the program stops showing it: over Duration it fades out,
// dimming and all, as it moves by X and Y (a sheet sliding back down, a drawer back off the side).
// Its zero value goes at once.
type Leave struct {
	Duration time.Duration
	X, Y     float32
}

// leavingLayer is what a layer painted in the last frame it was shown, and how it leaves.
type leavingLayer struct {
	commands  []draw2d.Command
	leave     Leave
	lastFrame int       // the last frame it was shown in
	started   time.Time // when it began leaving; zero while it is shown
}

// keepForLeaving keeps what a layer painted in this frame (the commands from first on), for when
// the program stops showing it.
func (ui *Context) keepForLeaving(id ID, leave Leave, first int) {
	if leave.Duration <= 0 {
		return
	}
	if ui.leaving == nil {
		ui.leaving = map[ID]*leavingLayer{}
	}
	layer, found := ui.leaving[id]
	if !found {
		layer = &leavingLayer{}
		ui.leaving[id] = layer
	}
	layer.commands = append(layer.commands[:0], ui.drawing.Commands[first:]...)
	layer.leave, layer.lastFrame, layer.started = leave, ui.frame, time.Time{}
}

// paintLeaving paints the layers the program stopped showing, a step further on their way, and
// forgets those that are gone.
func (ui *Context) paintLeaving() {
	for id, layer := range ui.leaving {
		if layer.lastFrame == ui.frame {
			continue
		}
		if layer.started.IsZero() {
			layer.started = ui.now
		}
		progress := float32(ui.now.Sub(layer.started)) / float32(layer.leave.Duration)
		if progress >= 1 {
			delete(ui.leaving, id)
			continue
		}
		ui.paintOnItsWay(layer, max(progress, 0))
		ui.state.RequestFrame(16 * time.Millisecond)
	}
}

// paintOnItsWay adds a leaving layer's picture to the draw list, faded and moved as far as it has
// gone: it starts slowly and speeds up, as things that leave do.
func (ui *Context) paintOnItsWay(layer *leavingLayer, progress float32) {
	moved := progress * progress
	shiftX, shiftY := layer.leave.X*moved, layer.leave.Y*moved
	shown := 1 - progress
	for _, command := range layer.commands {
		command.Area.X += shiftX
		command.Area.Y += shiftY
		command.Clip.X += shiftX
		command.Clip.Y += shiftY
		command.Baseline += shiftY
		command.Fill[3] *= shown
		command.Border[3] *= shown
		ui.drawing.Commands = append(ui.drawing.Commands, command)
	}
}
