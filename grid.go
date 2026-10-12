package ui

// A grid (Layout.Columns above 0) places its children in rows of equal cells. Every cell of a row
// is as tall as the row's tallest child; every column is as wide as the widest child when the
// grid fits its content, or an equal share of the grid's width when it is wider.

// gridShape counts the children and the rows they make.
func (ui *Context) gridShape(element *element) (children, rows int) {
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		children++
	}
	columns := element.layout.Columns
	return children, (children + columns - 1) / columns
}

func (ui *Context) fitGrid(element *element, axis Direction) {
	columns := element.layout.Columns
	children, rows := ui.gridShape(element)
	rowSizes := make([]float32, max(rows, 1))
	widest := float32(0)
	position := 0
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		ui.fitContent(child, axis)
		size := ui.elements[child].sizeAlong(axis)
		widest = max(widest, size)
		rowSizes[position/columns] = max(rowSizes[position/columns], size)
		position++
	}
	content := element.layout.Padding.along(axis)
	if axis == Horizontal {
		cells := min(columns, max(children, 1))
		content += widest*float32(cells) + element.layout.Gap*float32(cells-1)
	} else {
		for _, size := range rowSizes {
			content += size
		}
		content += element.layout.Gap * float32(max(rows-1, 0))
	}
	ui.sizeFromContent(element, axis, content)
}

// shareGrid gives every child its cell's width (unless it has a fixed one), and lets growing
// children fill their row's height.
func (ui *Context) shareGrid(element *element, axis Direction, inner float32) {
	columns := element.layout.Columns
	if axis == Horizontal {
		cell := max((inner-element.layout.Gap*float32(columns-1))/float32(columns), 0)
		for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
			childElement := &ui.elements[child]
			if sizing := childElement.layout.Width; sizing.kind != fixedSizing {
				childElement.setSize(Horizontal, sizing.clamp(cell))
			}
		}
		return
	}
	rowHeights := ui.gridRowHeights(element)
	position := 0
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		childElement := &ui.elements[child]
		if sizing := childElement.layout.Height; sizing.kind == growSizing {
			childElement.setSize(Vertical, sizing.clamp(rowHeights[position/columns]))
		}
		position++
	}
}

func (ui *Context) gridRowHeights(element *element) []float32 {
	columns := element.layout.Columns
	_, rows := ui.gridShape(element)
	heights := make([]float32, max(rows, 1))
	position := 0
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		heights[position/columns] = max(heights[position/columns], ui.elements[child].rect.Height)
		position++
	}
	return heights
}

// placeGrid puts every child at the top-left corner of its cell.
func (ui *Context) placeGrid(element *element) {
	columns := element.layout.Columns
	gap := element.layout.Gap
	inner := element.rect.Width - element.layout.Padding.along(Horizontal)
	cell := max((inner-gap*float32(columns-1))/float32(columns), 0)
	rowHeights := ui.gridRowHeights(element)
	left := element.rect.X + element.layout.Padding.Left
	top := element.rect.Y + element.layout.Padding.Top
	position := 0
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		column, row := position%columns, position/columns
		if column == 0 && row > 0 {
			top += rowHeights[row-1] + gap
		}
		ui.elements[child].rect.X = left + float32(column)*(cell+gap)
		ui.elements[child].rect.Y = top
		position++
	}
}

// Grid lays out what content adds in rows of a number of equal columns, across the whole width.
func (ui *Context) Grid(columns int, content func()) {
	ui.Box(Layout{Width: Grow(), Gap: ui.Theme.Gap, Columns: max(columns, 1)}, content)
}
