package ui

import (
	"math"
	"strconv"
	"time"

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

	glide       float32   // pixels a second the content moves on after a finger let it go
	lastTime    time.Time // the last step's time, for the glide
	pull        float32   // how far a finger pulls the content down past its top
	pulledAt    float32   // how far it was pulled when the finger let go
	pulledFrame int       // the frame it let go in
}

// scrollBarWidth is the thumb's width, and scrollBarReach how wide an area around it takes the
// pointer, so a thin bar is still easy to grab.
const (
	scrollBarWidth = 4
	scrollBarReach = 12
	minimumThumb   = 24
)

// A glide slows by glideFriction of its speed every second and stops under glideStop pixels a
// second. A pull past the top moves half as far as the finger (pullResistance), as a list resists.
const (
	glideFriction  = 4
	glideStop      = 20
	pullResistance = 0.5
)

// ScrollArea adds a box whose content scrolls up and down when it is taller than the box: with
// the wheel or the trackpad over it, by dragging the thin bar on its right, or, on a touch screen,
// by dragging the content with a finger, which glides on when let go. The box needs a
// height that does not follow its content (Grow, Fixed, Percent, or Fit with a maximum); its
// children are stacked top to bottom.
func (ui *Context) ScrollArea(name string, layout Layout, content func()) {
	ui.scrollArea(ui.idFor(name), layout, content)
}

func (ui *Context) scrollArea(id ID, layout Layout, content func()) *scrollMemory {
	memory := ui.scrollMemoryOf(id)
	layout.Direction, layout.Columns = Vertical, 0
	index := ui.frameWithID(id, layout, Look{Clip: true}, content)
	ui.elements[index].scroll = memory
	ui.followWheel(id, memory)
	ui.scrollForReader(id, memory)
	ui.followThumb(memory)
	ui.followFinger(id, memory)
	return memory
}

// followFinger drags the content with a finger that went up or down first, and lets it glide on
// when let go while moving. Pulled down past its top, the content stays and the pull is measured
// (ScrollPull), for a list that refreshes when pulled. Inner areas see the finger first (they
// finish first) and keep it.
func (ui *Context) followFinger(id ID, memory *scrollMemory) {
	seconds := float32(ui.now.Sub(memory.lastTime).Seconds())
	memory.lastTime = ui.now
	finger := &ui.finger
	if !ui.takeFinger(id, Vertical) {
		ui.glideOn(memory, seconds)
		return
	}
	if finger.released {
		memory.glide = -finger.velocityY
		if memory.pull > 0 {
			memory.glide = 0
		}
		memory.pulledAt, memory.pulledFrame, memory.pull = memory.pull, ui.frame, 0
		return
	}
	moved := finger.y - finger.previousY
	switch {
	case memory.pull > 0 || (moved > 0 && memory.offset <= 0):
		memory.pull = max(memory.pull+moved*pullResistance, 0)
		memory.offset = 0
	default:
		memory.offset -= moved
	}
	memory.glide = 0
}

// ScrollPull tells how far a finger is pulling a scroll area's content down past its top, and, in
// the frame the finger lets go, how far it was pulled then (zero otherwise): a list refreshes when
// pulled far enough and let go. name is the scroll area's, in the same scope.
func (ui *Context) ScrollPull(name string) (pulling, releasedAt float32) {
	// the area's id, worked out without counting as a use of the name (idFor would make the area's
	// own a different one)
	memory := ui.scrollMemoryOf(hashString(ui.scopeSeed(), name))
	if memory.pulledFrame >= ui.frame-1 && memory.pulledFrame > 0 {
		releasedAt = memory.pulledAt
	}
	return memory.pull, releasedAt
}

// ScrollOffset tells how far a scroll area's content is scrolled from its top, as of the last
// frame: a bar over a list shrinks or takes a colour as the list goes up under it. name is the
// scroll area's, in the same scope.
func (ui *Context) ScrollOffset(name string) float32 {
	return ui.scrollMemoryOf(hashString(ui.scopeSeed(), name)).offset
}

// scrollMemoryOf is a scroll area's memory, made the first time either the area or ScrollPull asks.
func (ui *Context) scrollMemoryOf(id ID) *scrollMemory {
	return remember(ui, id, func() *scrollMemory { return &scrollMemory{thumbID: childID(id, "##thumb")} })
}

// glideOn moves the content on after a finger let go of it, slowing to a stop.
func (ui *Context) glideOn(memory *scrollMemory, seconds float32) {
	if memory.glide == 0 || seconds <= 0 || seconds > 0.25 {
		return
	}
	before := memory.offset
	memory.offset = clamp(memory.offset+memory.glide*seconds, 0, max(memory.content-memory.view, 0))
	memory.glide *= float32(math.Exp(-glideFriction * float64(seconds)))
	if abs32(memory.glide) < glideStop || memory.offset == before {
		memory.glide = 0
		return
	}
	ui.RequestFrame(16 * time.Millisecond)
}

func abs32(value float32) float32 { return max(value, -value) }

// followWheel scrolls with the wheel when the pointer is over the area. Inner areas see the wheel
// first (they finish first) and take it, so a list inside a scrolling panel scrolls on its own.
func (ui *Context) followWheel(id ID, memory *scrollMemory) {
	delta := float32(ui.state.ScrollDelta)
	if delta == 0 || ui.scrollUsed || ui.sceneDrag || memory.content <= memory.view {
		return
	}
	area, shown := ui.reachOf(id)
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
	thumb := ui.pointerOn(memory.thumbID) // a screen reader scrolls the area itself
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
