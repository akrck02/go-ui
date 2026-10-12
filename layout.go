package ui

// Direction is the way a box lines up its children.
type Direction int

const (
	Horizontal Direction = iota // left to right
	Vertical                    // top to bottom
)

// Align places children inside the free space of their box, along one axis.
type Align int

const (
	Start  Align = iota // left or top
	Center              // in the middle
	End                 // right or bottom
)

type sizingKind int

const (
	fitSizing     sizingKind = iota // as big as the content
	growSizing                      // the content, plus a share of the space left in the parent
	fixedSizing                     // a set number of pixels
	percentSizing                   // a share of the parent's inner size
)

// Sizing says how big a box is along one axis. The zero value fits the content.
type Sizing struct {
	kind      sizingKind
	value     float32 // grow: weight; fixed: pixels; percent: share from 0 to 1
	fromEmpty bool    // grow: start from nothing instead of the content (Share)
	// Min and Max bound the size, in pixels; 0 = no bound.
	Min, Max float32
}

// Fit makes a box as big as its content (the default).
func Fit() Sizing { return Sizing{kind: fitSizing} }

// Grow makes a box take the space its parent has left, shared equally with its growing siblings.
func Grow() Sizing { return Sizing{kind: growSizing, value: 1} }

// GrowBy is Grow with a weight: a box with weight 2 takes twice the share of one with weight 1.
func GrowBy(weight float32) Sizing { return Sizing{kind: growSizing, value: max(weight, 0)} }

// Share makes a box take weight shares of the free space, starting from nothing rather than from
// its content: boxes with equal shares come out equal whatever they hold, so table columns line up
// from row to row and cards in a row match. Content wider than its share overflows.
func Share(weight float32) Sizing {
	return Sizing{kind: growSizing, value: max(weight, 0), fromEmpty: true}
}

// Fixed makes a box exactly this many pixels.
func Fixed(pixels float32) Sizing { return Sizing{kind: fixedSizing, value: max(pixels, 0)} }

// Percent makes a box a share (0 to 1) of its parent's inner size.
func Percent(share float32) Sizing { return Sizing{kind: percentSizing, value: max(share, 0)} }

// Between returns the sizing bounded by a minimum and a maximum, in pixels (0 = no bound).
func (sizing Sizing) Between(minimum, maximum float32) Sizing {
	sizing.Min, sizing.Max = minimum, maximum
	return sizing
}

func (sizing Sizing) clamp(size float32) float32 {
	if sizing.Max > 0 {
		size = min(size, sizing.Max)
	}
	return max(size, sizing.Min, 0)
}

// Padding is the empty space between a box's edges and its children, in pixels.
type Padding struct {
	Left, Top, Right, Bottom float32
}

// Pad is the same padding on every side.
func Pad(all float32) Padding { return Padding{all, all, all, all} }

// PadXY is one padding on the left and right and another on the top and bottom.
func PadXY(horizontal, vertical float32) Padding {
	return Padding{horizontal, vertical, horizontal, vertical}
}

func (padding Padding) along(axis Direction) float32 {
	if axis == Horizontal {
		return padding.Left + padding.Right
	}
	return padding.Top + padding.Bottom
}

func (padding Padding) before(axis Direction) float32 {
	if axis == Horizontal {
		return padding.Left
	}
	return padding.Top
}

// Layout says how a box sizes itself and places its children. The zero value is a box that lines
// its children up left to right, fits them, and has no padding or gap.
type Layout struct {
	Direction     Direction
	Width, Height Sizing
	Padding       Padding
	Gap           float32 // space between two children, in pixels
	// Wrap, on a horizontal box, starts a new line when the next child would pass the box's
	// right edge, like words in a paragraph (a row of tags). Gap is kept between children and
	// between lines. The box needs a width that does not follow its content (Grow in a column,
	// Fixed, Percent), or it is one line as wide as all its children.
	// (Next to Gap, it takes the padding the compiler leaves there: boxes stay as small.)
	Wrap   bool
	AlignX Align // where the children sit horizontally when there is room to spare
	AlignY Align // where the children sit vertically when there is room to spare
	// Columns, when above 0, makes a grid: children fill rows of this many equal cells, left to
	// right and then down, with Gap between cells both ways. Direction is not used then.
	Columns int
}

// wraps tells whether a box breaks its children into lines. Callers test layout.Wrap on its own
// first: that one byte keeps the layout of boxes that do not wrap as fast as before (measured with
// BenchmarkAuditPhases).
func (layout Layout) wraps() bool {
	return layout.Wrap && layout.Direction == Horizontal && layout.Columns == 0
}

func (layout Layout) sizing(axis Direction) Sizing {
	if axis == Horizontal {
		return layout.Width
	}
	return layout.Height
}

func (layout Layout) align(axis Direction) Align {
	if axis == Horizontal {
		return layout.AlignX
	}
	return layout.AlignY
}

// The layout runs once per frame, when the whole tree of boxes is known (Context.End), axis by
// axis: first every box's size from its content, from the leaves up; then the space left over
// is shared out to growing boxes, from the root down; then positions, from the root down. Text
// does not wrap, so heights never depend on widths and the two axes are independent, except in
// boxes that wrap their children into lines (Layout.Wrap): their height is worked out from the
// widths, which are all known by then, because the horizontal axis is done first.

// size works out the sizes of a root's tree along an axis: the content first, then the root's
// own size (which may depend on it), then the space shared out inside.
func (ui *Context) size(placed *root, axis Direction) {
	ui.fitContent(placed.element, axis)
	ui.elements[placed.element].setSize(axis, ui.rootSize(placed, axis))
	ui.shareSpace(placed.element, axis)
}

// fitContent sets every box to the size of its content along an axis, children first. Growing
// boxes start at their content's size too; percentages wait for their parent's size.
func (ui *Context) fitContent(index int, axis Direction) {
	element := &ui.elements[index]
	if element.layout.Columns > 0 {
		ui.fitGrid(element, axis)
		return
	}
	if element.layout.Wrap && axis == Vertical && element.layout.wraps() {
		ui.fitWrapped(element)
		return
	}
	content := element.intrinsic(axis)
	isMainAxis := element.layout.Direction == axis
	childCount := 0
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		ui.fitContent(child, axis)
		childSize := ui.elements[child].sizeAlong(axis)
		if isMainAxis {
			content += childSize
		} else {
			content = max(content, childSize)
		}
		childCount++
	}
	if isMainAxis && childCount > 1 {
		content += element.layout.Gap * float32(childCount-1)
	}
	content += element.layout.Padding.along(axis)
	ui.sizeFromContent(element, axis, content)
}

// sizeFromContent sets a box's size along an axis from what its content needs, as its sizing
// says, and remembers that need (a scroll area is smaller than its content).
func (ui *Context) sizeFromContent(element *element, axis Direction, content float32) {
	element.content[axis] = content
	sizing := element.layout.sizing(axis)
	switch sizing.kind {
	case fixedSizing:
		element.setSize(axis, sizing.clamp(sizing.value))
	case percentSizing:
		element.setSize(axis, 0)
	case growSizing:
		if sizing.fromEmpty {
			element.setSize(axis, sizing.clamp(0))
			break
		}
		element.setSize(axis, sizing.clamp(content))
	default:
		element.setSize(axis, sizing.clamp(content))
	}
}

// shareSpace sizes the children of a box from its final size: percentages take their share,
// and growing children split what is left along the main axis, or fill the box across it.
func (ui *Context) shareSpace(index int, axis Direction) {
	element := &ui.elements[index]
	inner := max(element.sizeAlong(axis)-element.layout.Padding.along(axis), 0)
	if element.layout.Columns > 0 {
		ui.shareGrid(element, axis, inner)
	} else if element.layout.Wrap && element.layout.wraps() {
		ui.shareWrapped(element, axis, inner)
	} else if element.layout.Direction == axis {
		ui.shareAlongMainAxis(element, axis, inner)
	} else {
		ui.fillAcross(element, axis, inner)
	}
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		ui.shareSpace(child, axis)
	}
}

func (ui *Context) shareAlongMainAxis(element *element, axis Direction, inner float32) {
	used, weights, childCount := float32(0), float32(0), 0
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		childElement := &ui.elements[child]
		sizing := childElement.layout.sizing(axis)
		if sizing.kind == percentSizing {
			childElement.setSize(axis, sizing.clamp(inner*sizing.value))
		}
		used += childElement.sizeAlong(axis)
		if sizing.kind == growSizing {
			weights += sizing.value
		}
		childCount++
	}
	if childCount > 1 {
		used += element.layout.Gap * float32(childCount-1)
	}
	if free := inner - used; free < 0 {
		ui.shrinkChildren(element, axis, -free)
		return
	}
	ui.growChildren(element, axis, inner-used, weights)
}

// shrinkChildren takes the space missing from the growing children, when their content does not
// fit the box: each gives up a part of what it has above its minimum, in proportion to it, so
// the fixed and fitted boxes around them (a side column, an inspector, a button) keep their
// size. What the growing children cannot give overflows at the end.
func (ui *Context) shrinkChildren(element *element, axis Direction, missing float32) {
	slack := float32(0)
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		childElement := &ui.elements[child]
		if sizing := childElement.layout.sizing(axis); sizing.kind == growSizing {
			slack += max(childElement.sizeAlong(axis)-sizing.clamp(0), 0)
		}
	}
	if slack <= 0 {
		return
	}
	share := min(missing/slack, 1)
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		childElement := &ui.elements[child]
		if sizing := childElement.layout.sizing(axis); sizing.kind == growSizing {
			size := childElement.sizeAlong(axis)
			childElement.setSize(axis, size-max(size-sizing.clamp(0), 0)*share)
		}
	}
}

// growChildren adds the free space to the growing children, in proportion to their weights. A
// child that reaches its maximum keeps it, and the rest is shared again among the others.
func (ui *Context) growChildren(element *element, axis Direction, free, weights float32) {
	for free > 0.5 && weights > 0 {
		perWeight, given, capped := free/weights, float32(0), float32(0)
		for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
			childElement := &ui.elements[child]
			sizing := childElement.layout.sizing(axis)
			if sizing.kind != growSizing || sizing.value == 0 || childElement.isCapped(axis) {
				continue
			}
			before := childElement.sizeAlong(axis)
			after := sizing.clamp(before + perWeight*sizing.value)
			childElement.setSize(axis, after)
			given += after - before
			if sizing.Max > 0 && after >= sizing.Max {
				childElement.capAxis(axis)
				capped += sizing.value
			}
		}
		free -= given
		if capped == 0 || given <= 0 {
			return
		}
		weights -= capped
	}
}

func (ui *Context) fillAcross(element *element, axis Direction, inner float32) {
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		childElement := &ui.elements[child]
		sizing := childElement.layout.sizing(axis)
		switch sizing.kind {
		case growSizing:
			childElement.setSize(axis, sizing.clamp(inner)) // the box's size, even below the content: it shrinks too
		case percentSizing:
			childElement.setSize(axis, sizing.clamp(inner*sizing.value))
		}
	}
}

// place sets the position of every child of a box, along both axes, from the box's own.
func (ui *Context) place(index int) {
	element := &ui.elements[index]
	if element.layout.Columns > 0 || (element.layout.Wrap && element.layout.wraps()) {
		if element.layout.Columns > 0 {
			ui.placeGrid(element)
		} else {
			ui.placeWrapped(element)
		}
		for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
			ui.place(child)
		}
		return
	}
	for _, axis := range []Direction{Horizontal, Vertical} {
		if element.layout.Direction == axis {
			ui.placeAlongMainAxis(element, axis)
		} else {
			ui.placeAcross(element, axis)
		}
	}
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		ui.place(child)
	}
}

func (ui *Context) placeAlongMainAxis(element *element, axis Direction) {
	inner := element.sizeAlong(axis) - element.layout.Padding.along(axis)
	used, childCount := float32(0), 0
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		used += ui.elements[child].sizeAlong(axis)
		childCount++
	}
	if childCount > 1 {
		used += element.layout.Gap * float32(childCount-1)
	}
	position := element.positionAlong(axis) + element.layout.Padding.before(axis) + alignOffset(element.layout.align(axis), inner-used)
	if element.scroll != nil && axis == Vertical {
		element.scroll.keepInRange(element.content[Vertical], element.rect.Height)
		position -= element.scroll.offset
	}
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		ui.elements[child].setPosition(axis, position)
		position += ui.elements[child].sizeAlong(axis) + element.layout.Gap
	}
}

func (ui *Context) placeAcross(element *element, axis Direction) {
	inner := element.sizeAlong(axis) - element.layout.Padding.along(axis)
	start := element.positionAlong(axis) + element.layout.Padding.before(axis)
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		childElement := &ui.elements[child]
		childElement.setPosition(axis, start+alignOffset(element.layout.align(axis), inner-childElement.sizeAlong(axis)))
	}
}

// alignOffset is where content starts inside free space; with no room to spare it starts at the
// beginning, so overflowing content is cut at the end rather than at both sides.
func alignOffset(align Align, free float32) float32 {
	if free <= 0 {
		return 0
	}
	switch align {
	case Center:
		return free / 2
	case End:
		return free
	default:
		return 0
	}
}
