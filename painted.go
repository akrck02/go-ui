package ui

import "github.com/akrck02/go-renderer/draw2d"

// Paint adds a box laid out as any other, which the program paints itself once the layout knows
// where it is: a shape go-ui has no control for (a loading indicator, a wave, a chart). paint is
// given the box's area, in window pixels, and the frame's draw list, and adds rectangles, shadows,
// texts and icons to it; what it adds is clipped as the box is. A painting that moves asks for the
// next frame itself (RequestFrame).
func (ui *Context) Paint(layout Layout, paint func(area draw2d.Rect, drawing *draw2d.List)) {
	ui.openElement(element{kind: boxElement, layout: layout, painter: paint})
	ui.closeElement()
}
