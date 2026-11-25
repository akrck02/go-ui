package remote

import "github.com/akrck02/go-renderer/draw2d"

// Show adds a frame's commands to a drawing, moved into an area and kept inside it: the host
// calls it each frame with the last frame the surface gave, so showing costs no message.
func Show(frame Frame, area draw2d.Rectangle, drawing *draw2d.List) {
	for _, command := range frame.Commands {
		command.Area.X += area.X
		command.Area.Y += area.Y
		command.Baseline += area.Y
		clip := area
		if command.HasClip {
			command.Clip.X += area.X
			command.Clip.Y += area.Y
			clip = command.Clip.Intersect(area)
		}
		command.Clip, command.HasClip = clip, true
		drawing.Commands = append(drawing.Commands, command)
	}
}
