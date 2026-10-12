package ui

import "github.com/akrck02/go-renderer/input"

// Side is the side of a resizable box its handle is on.
type Side int

const (
	RightSide  Side = iota // the box grows to the right: a column on the left of a window
	LeftSide               // the box grows to the left: a column on the right
	BottomSide             // the box grows down: a strip at the top
	TopSide                // the box grows up: a strip at the bottom
)

// horizontal tells whether the handle moves sideways, changing a width.
func (side Side) horizontal() bool { return side == RightSide || side == LeftSide }

// resizeHandle is the thickness of the handle: wide enough to grab, with a thin line in the middle
// that stands in for a separator.
const resizeHandle = 5

// resizeMemory is what a resizable box keeps while its handle is dragged: its size and the
// pointer's place when the drag started, so the size follows the pointer exactly and a pointer
// that goes past a bound and comes back finds the edge where it left it.
type resizeMemory struct {
	startSize, startPointer float32
}

// Resizable adds a box whose width (with its handle on the right or left side) or height (bottom
// or top) the person changes by dragging its edge, between minimum and maximum pixels (0 = no
// maximum). size is the program's variable, read and written in place, so the program keeps it
// (and may save it). layout says how the box sizes across the handle and how it arranges its
// content; its size along the handle comes from size. The handle is a thin line, drawn like a
// Separator, which it replaces. It returns true on the steps size changes.
func (ui *Context) Resizable(name string, side Side, size *float32, minimum, maximum float32, layout Layout, content func()) bool {
	id := ui.idFor(name)
	memory := remember(ui, id, func() *resizeMemory { return &resizeMemory{} })
	*size = clampSize(*size, minimum, maximum)
	inner, outer := layout, Layout{Direction: Horizontal, Height: layout.Height}
	if side.horizontal() {
		inner.Width, inner.Height = Fixed(*size), Grow()
	} else {
		inner.Height, inner.Width = Fixed(*size), Grow()
		outer = Layout{Direction: Vertical, Width: layout.Width}
	}
	var result interaction
	ui.Box(outer, func() {
		if side == LeftSide || side == TopSide {
			result = ui.resizeHandle(id, side)
		}
		ui.Box(inner, content)
		if side == RightSide || side == BottomSide {
			result = ui.resizeHandle(id, side)
		}
	})
	return ui.followResize(result, memory, side, size, minimum, maximum)
}

// resizeHandle adds the handle of a resizable box: a line that is drawn stronger under the pointer
// and while dragged, with the resize pointer.
func (ui *Context) resizeHandle(id ID, side Side) interaction {
	theme := ui.Theme
	handleLayout := Layout{Width: Fixed(resizeHandle), Height: Grow(), AlignX: Center}
	lineLayout := Layout{Width: Fixed(1), Height: Grow()}
	if !side.horizontal() {
		handleLayout = Layout{Direction: Vertical, Width: Grow(), Height: Fixed(resizeHandle), AlignY: Center}
		lineLayout = Layout{Width: Grow(), Height: Fixed(1)}
	}
	var line int
	_, result := ui.pressable(childID(id, "##handle"), handleLayout, Look{}, func() {
		line = ui.frameWithID(0, lineLayout, Look{}, nil)
	})
	colour := theme.Line
	if result.hovered || result.held {
		colour = theme.Primary
		ui.cursor = input.CursorResizeHorizontal
		if !side.horizontal() {
			ui.cursor = input.CursorResizeVertical
		}
	}
	ui.elements[line].look = Look{Fill: colour}
	return result
}

// followResize changes the size while the handle is dragged, by as much as the pointer moved since
// the press, towards the box's side.
func (ui *Context) followResize(result interaction, memory *resizeMemory, side Side, size *float32, minimum, maximum float32) bool {
	pointer := float32(ui.state.CursorX)
	if !side.horizontal() {
		pointer = float32(ui.state.CursorY)
	}
	if result.pressed {
		memory.startSize, memory.startPointer = *size, pointer
	}
	if !result.held {
		return false
	}
	moved := pointer - memory.startPointer
	if side == LeftSide || side == TopSide {
		moved = -moved
	}
	before := *size
	*size = clampSize(memory.startSize+moved, minimum, maximum)
	return *size != before
}

// clampSize keeps a size between its bounds; a maximum of 0 is no maximum.
func clampSize(size, minimum, maximum float32) float32 {
	if maximum > 0 {
		size = min(size, maximum)
	}
	return max(size, minimum)
}
