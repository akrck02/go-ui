package ui

import (
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// scrollMemory is what a scroll area keeps: how far it is scrolled, and how tall its content and
// its view were in the last frame.
type scrollMemory struct {
	offset  float32
	content float32
	view    float32
	thumbID ID
	thumbOn bool // the pointer is on the thumb or dragging it: drawn stronger
}

// scrollBarWidth is the thumb's width, and scrollBarReach how wide an area around it takes the
// pointer, so a thin bar is still easy to grab.
const (
	scrollBarWidth = 4
	scrollBarReach = 12
	minimumThumb   = 24
)

// ScrollArea adds a box whose content scrolls up and down when it is taller than the box: with
// the wheel or the trackpad over it, or by dragging the thin bar on its right. The box needs a
// height that does not follow its content (Grow, Fixed, Percent, or Fit with a maximum); its
// children are stacked top to bottom.
func (ui *Context) ScrollArea(name string, layout Layout, content func()) {
	ui.scrollArea(ui.idFor(name), layout, content)
}

func (ui *Context) scrollArea(id ID, layout Layout, content func()) *scrollMemory {
	memory := remember(ui, id, func() *scrollMemory { return &scrollMemory{thumbID: childID(id, "##thumb")} })
	layout.Direction, layout.Columns = Vertical, 0
	index := ui.frameWithID(id, layout, Look{Clip: true}, content)
	ui.elements[index].scroll = memory
	ui.followWheel(id, memory)
	ui.followThumb(memory)
	return memory
}

// followWheel scrolls with the wheel when the pointer is over the area. Inner areas see the wheel
// first (they finish first) and take it, so a list inside a scrolling panel scrolls on its own.
func (ui *Context) followWheel(id ID, memory *scrollMemory) {
	delta := float32(ui.state.ScrollDelta)
	if delta == 0 || ui.scrollUsed || ui.sceneDrag || memory.content <= memory.view {
		return
	}
	area, shown := ui.previous[id]
	x, y := float32(ui.state.CursorX), float32(ui.state.CursorY)
	if !shown || !area.Contains(x, y) || ui.hoveredRoot != ui.currentRoot() {
		return
	}
	memory.offset -= delta * ui.Theme.ScrollStep
	ui.scrollUsed = true
}

// followThumb moves the content while the thumb is dragged, as far as the thumb moves times how
// much taller the content is than the view.
func (ui *Context) followThumb(memory *scrollMemory) {
	thumb := ui.interact(memory.thumbID)
	memory.thumbOn = thumb.hovered || ui.active == memory.thumbID
	if memory.thumbOn {
		ui.cursor = input.CursorArrow
	}
	if thumb.held && memory.view > 0 {
		memory.offset += float32(ui.state.CursorDeltaY) * memory.content / memory.view
	}
}

// keepInRange clamps a scroll area's offset to its content, once the content is measured.
func (memory *scrollMemory) keepInRange(content, view float32) {
	memory.content, memory.view = content, view
	memory.offset = clamp(memory.offset, 0, max(content-view, 0))
}

// thumbArea is where the thumb is drawn in an area, and whether there is one.
func (memory *scrollMemory) thumbArea(area draw2d.Rect) (draw2d.Rect, bool) {
	if memory.content <= memory.view+0.5 || memory.view <= 0 {
		return draw2d.Rect{}, false
	}
	length := max(area.Height*memory.view/memory.content, minimumThumb)
	travel := area.Height - length - 4
	position := memory.offset / (memory.content - memory.view) * travel
	return draw2d.Rect{X: area.Right() - scrollBarWidth - 3, Y: area.Y + 2 + position, Width: scrollBarWidth, Height: length}, true
}

// recordScrollBars notes where every thumb is (with its wider reach) for the next frame's mouse.
func (ui *Context) recordScrollBars() {
	for index := range ui.elements {
		element := &ui.elements[index]
		if element.scroll == nil {
			continue
		}
		if thumb, shown := element.scroll.thumbArea(element.rect); shown {
			reach := draw2d.Rect{X: thumb.Right() - scrollBarReach, Y: thumb.Y, Width: scrollBarReach + 3, Height: thumb.Height}
			ui.current[element.scroll.thumbID] = reach
		}
	}
}

// paintScrollBar draws the thumb over a scroll area's content.
func (ui *Context) paintScrollBar(element *element) {
	thumb, shown := element.scroll.thumbArea(element.rect)
	if !shown {
		return
	}
	colour := ui.Theme.InkQuiet.WithAlpha(0.45)
	if element.scroll.thumbOn {
		colour = ui.Theme.InkSoft.WithAlpha(0.8)
	}
	ui.drawing.Rectangle(thumb, scrollBarWidth/2, colour, draw2d.Transparent, 0)
}

// List shows count rows of the same height in a scroll area and builds only the rows that can be
// seen, so a list of ten thousand rows costs what a list of twenty does. row builds one row; ids
// inside it are scoped by the row's number.
func (ui *Context) List(name string, layout Layout, count int, rowHeight float32, row func(index int)) {
	id := ui.idFor(name)
	memory := remember(ui, id, func() *scrollMemory { return &scrollMemory{thumbID: childID(id, "##thumb")} })
	view := memory.view
	if view <= 0 {
		view = ui.height // the first frame: the area's height is not known yet
	}
	first := clampInt(int(memory.offset/rowHeight), 0, count)
	last := clampInt(int((memory.offset+view)/rowHeight)+1, first, count)
	layout.Gap = 0
	ui.scrollArea(id, layout, func() {
		ui.Box(Layout{Width: Grow(), Height: Fixed(float32(first) * rowHeight)}, nil)
		for index := first; index < last; index++ {
			ui.Push(strconv.Itoa(index))
			ui.Box(Layout{Direction: Vertical, Width: Grow(), Height: Fixed(rowHeight), AlignY: Center}, func() { row(index) })
			ui.Pop()
		}
		ui.Box(Layout{Width: Grow(), Height: Fixed(float32(count-last) * rowHeight)}, nil)
	})
}

// ListBox shows items as a scrolling list of selectable rows and returns true when the selection
// changes.
func (ui *Context) ListBox(name string, layout Layout, items []string, selected *int) bool {
	changed := false
	ui.List(name, layout, len(items), ui.Theme.ControlHeight+2, func(index int) {
		if ui.Selectable("", items[index], index == *selected) && *selected != index {
			*selected = index
			changed = true
		}
	})
	return changed
}

func clampInt(value, low, high int) int { return min(max(value, low), high) }
