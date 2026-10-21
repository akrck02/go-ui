package main

import (
	"sort"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// Sizes of the replay's marks, in layout pixels of the viewer.
const (
	highlightWidth = 3  // the outline round the element an action found
	pointerRadius  = 9  // the ring where the pointer is
	pointerBorder  = 2  // that ring's line
	pictureRadius  = 4  // the grey box a recorded picture becomes
	replayRadius   = 10 // the corners of the replayed window
)

// pictureGrey stands in for a recorded picture, whose pixels are not kept.
var pictureGrey = draw2d.Color{0.5, 0.5, 0.5, 0.5}

// replay paints a recorded frame scaled into an area: what the window drew, then the pointer,
// the fingers and the element the frame's action found.
type replay struct {
	recording uitest.Recording
	frame     uitest.RecordedFrame
	theme     ui.Theme
}

// paint fits the recorded window into the area, centred, and paints the frame in it.
func (replay replay) paint(area draw2d.Rect, drawing *draw2d.List) {
	if replay.recording.Width <= 0 || replay.recording.Height <= 0 {
		return
	}
	scale := min(area.Width/replay.recording.Width, area.Height/replay.recording.Height)
	origin := draw2d.Rect{
		X: area.X + (area.Width-replay.recording.Width*scale)/2, Y: area.Y + (area.Height-replay.recording.Height*scale)/2,
		Width: replay.recording.Width * scale, Height: replay.recording.Height * scale,
	}
	place := placement{origin: origin, scale: scale}
	drawing.Rectangle(origin, replayRadius, replay.theme.Background, replay.theme.Line, 1)
	drawing.PushClip(origin)
	for _, command := range inLayers(replay.frame.Commands) {
		place.paintCommand(command, drawing)
	}
	drawing.PopClip()
	replay.paintMarks(place, drawing)
}

// paintMarks paints what the test did: the found element, the fingers and the pointer.
func (replay replay) paintMarks(place placement, drawing *draw2d.List) {
	theme := replay.theme
	if replay.frame.HasTarget {
		drawing.Rectangle(place.rect(replay.frame.Target).Inset(-highlightWidth), theme.ControlRadius, draw2d.Transparent, theme.FocusRing, highlightWidth)
	}
	for _, touch := range replay.frame.Touches {
		paintSpot(drawing, place.point(touch.X, touch.Y), theme.Warning, true)
	}
	paintSpot(drawing, place.point(replay.frame.PointerX, replay.frame.PointerY), theme.Primary, replay.frame.Pressed)
}

// paintSpot paints a ring where the pointer or a finger is; a pressed one is filled.
func paintSpot(drawing *draw2d.List, center draw2d.Rect, color draw2d.Color, pressed bool) {
	spot := draw2d.Rect{X: center.X - pointerRadius, Y: center.Y - pointerRadius, Width: 2 * pointerRadius, Height: 2 * pointerRadius}
	fill := draw2d.Transparent
	if pressed {
		fill = color
	}
	drawing.Rectangle(spot, pointerRadius, fill, color, pointerBorder)
}

// inLayers sorts recorded commands by layer, keeping their order within one, as the painter does.
func inLayers(commands []draw2d.Command) []draw2d.Command {
	sorted := append([]draw2d.Command(nil), commands...)
	sort.SliceStable(sorted, func(first, second int) bool { return sorted[first].Layer < sorted[second].Layer })
	return sorted
}

// placement maps the recorded window's layout pixels into the viewer's.
type placement struct {
	origin draw2d.Rect
	scale  float32
}

// rect places a recorded rectangle.
func (place placement) rect(area draw2d.Rect) draw2d.Rect {
	return draw2d.Rect{X: place.origin.X + area.X*place.scale, Y: place.origin.Y + area.Y*place.scale,
		Width: area.Width * place.scale, Height: area.Height * place.scale}
}

// point places a recorded point, as an empty rectangle at it.
func (place placement) point(x, y float32) draw2d.Rect {
	return draw2d.Rect{X: place.origin.X + x*place.scale, Y: place.origin.Y + y*place.scale}
}

// paintCommand paints one recorded command, scaled, inside its own clip.
func (place placement) paintCommand(command draw2d.Command, drawing *draw2d.List) {
	if command.HasClip {
		drawing.PushClip(place.rect(command.Clip))
		defer drawing.PopClip()
	}
	area, scale := place.rect(command.Area), place.scale
	switch command.Kind {
	case draw2d.RectangleKind:
		drawing.Rectangle(area, command.Radius*scale, command.Fill, command.Border, command.BorderWidth*scale)
	case draw2d.ShadowKind:
		blur := command.Blur * scale // the area is the caster grown by the blur: shrink it back
		drawing.Shadow(area.Inset(blur), command.Radius*scale, blur, 0, command.Fill)
	case draw2d.TextKind:
		baseline := place.origin.Y + command.Baseline*scale
		drawing.Text(command.Text, area.X, baseline, command.Size*scale, command.Weight, command.Fill)
	case draw2d.IconKind:
		drawing.Icon(command.Text, area.X, area.Y, command.Size*scale, command.Fill)
	case draw2d.ImageKind:
		drawing.Rectangle(area, pictureRadius, pictureGrey, draw2d.Transparent, 0) // pictures are not recorded
	}
}
