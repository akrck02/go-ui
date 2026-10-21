package uidrive

import "github.com/akrck02/go-renderer/draw2d"

// ShownLayer is the layer a shown player draws on: above the interface's dialogs, popups and the
// screen reader's ring, so a demo's pointer is never hidden.
const ShownLayer = 1 << 20

// The look of what a shown player draws, in layout pixels.
const (
	pointerRadius   = 9
	pointerBorder   = 3
	fingerRadius    = 16
	outlineMargin   = 4
	outlineBorder   = 3
	outlineRounding = 6
)

// The colours a shown player draws in: bright enough on a light and on a dark theme.
var (
	pointerColour = draw2d.Color{0.96, 0.45, 0.10, 1}    // orange
	fingerColour  = draw2d.Color{0.96, 0.45, 0.10, 0.55} // orange, half seen through
	outlineColour = draw2d.Color{0.96, 0.45, 0.10, 0.9}
	noColour      = draw2d.Color{}
)

// Show draws, when the player's options ask for it, what it does over the interface: an outline
// around the element its steps found, the pointer as a ring (filled while pressed) and each finger
// as a dot. Without Show it draws nothing.
func (player *Player) Show(drawing *draw2d.List) {
	if !player.options.Show || player.done {
		return
	}
	drawing.SetLayer(ShownLayer)
	defer drawing.SetLayer(0)
	if area := player.found.Rect; area.Width > 0 && area.Height > 0 {
		outline := draw2d.Rect{X: area.X - outlineMargin, Y: area.Y - outlineMargin,
			Width: area.Width + 2*outlineMargin, Height: area.Height + 2*outlineMargin}
		drawing.Rectangle(outline, outlineRounding, noColour, outlineColour, outlineBorder)
	}
	for _, finger := range player.hand.fingers {
		drawing.Rectangle(circle(finger.x, finger.y, fingerRadius), fingerRadius, fingerColour, noColour, 0)
	}
	if player.hand.placed && len(player.hand.fingers) == 0 {
		fill := noColour
		if player.hand.down {
			fill = pointerColour
		}
		drawing.Rectangle(circle(player.hand.x, player.hand.y, pointerRadius), pointerRadius, fill, pointerColour, pointerBorder)
	}
}

// circle is the square a circle of a radius around a place fills.
func circle(x, y, radius float32) draw2d.Rect {
	return draw2d.Rect{X: x - radius, Y: y - radius, Width: 2 * radius, Height: 2 * radius}
}
