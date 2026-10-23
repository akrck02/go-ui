package ui

// A horizontal box with Layout.Wrap breaks its children into lines: a child that would pass the
// right edge starts the next line. Lines are as tall as their tallest child and stacked with the
// box's Gap; inside a line, AlignX places the line and AlignY places each child in it.

// wrappedLine is one line of a wrapping box: its children from first up to (not including) next,
// and its size.
type wrappedLine struct {
	first, next   int // child indices; next is -1 after the last line
	width, height float32
}

// lineFrom gathers the children that fit on one line, starting at the child first, within a width.
// A line always takes at least one child, so a child wider than the box gets a line of its own.
func (ui *Context) lineFrom(element *element, first int, width float32) wrappedLine {
	line := wrappedLine{first: first, next: -1}
	for child := first; child >= 0; child = ui.elements[child].nextSibling {
		childElement := &ui.elements[child]
		needed := childElement.rect.Width
		if child != first {
			needed += element.layout.Gap
		}
		if child != first && line.width+needed > width+0.5 {
			line.next = child
			return line
		}
		line.width += needed
		line.height = max(line.height, childElement.rect.Height)
	}
	return line
}

// innerWidth is the width a box's children have between its padding.
func (element *element) innerWidth() float32 {
	return max(element.rect.Width-element.layout.Padding.along(Horizontal), 0)
}

// fitWrapped sets a wrapping box's height from its lines, once the widths are known.
func (ui *Context) fitWrapped(element *element) {
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		ui.fitContent(child, Vertical)
	}
	inner, content, lines := element.innerWidth(), float32(0), 0
	for first := element.firstChild; first >= 0; {
		line := ui.lineFrom(element, first, inner)
		content += line.height
		lines++
		first = line.next
	}
	content += element.layout.lineGap()*float32(max(lines-1, 0)) + element.layout.Padding.along(Vertical)
	ui.sizeFromContent(element, Vertical, content)
}

// shareWrapped sizes the children of a wrapping box: across, no child is wider than the box and
// percentages take their share; down, growing children fill their line's height.
func (ui *Context) shareWrapped(element *element, axis Direction, inner float32) {
	if axis == Horizontal {
		for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
			childElement := &ui.elements[child]
			if sizing := childElement.layout.Width; sizing.kind == percentSizing {
				childElement.setSize(Horizontal, sizing.clamp(inner*sizing.value))
			}
			childElement.setSize(Horizontal, min(childElement.rect.Width, inner))
		}
		return
	}
	for first := element.firstChild; first >= 0; {
		line := ui.lineFrom(element, first, element.innerWidth())
		for child := line.first; child != line.next; child = ui.elements[child].nextSibling {
			if sizing := ui.elements[child].layout.Height; sizing.kind == growSizing {
				ui.elements[child].setSize(Vertical, sizing.clamp(line.height))
			}
		}
		first = line.next
	}
}

// placeWrapped puts the children of a wrapping box line by line.
func (ui *Context) placeWrapped(element *element) {
	inner := element.innerWidth()
	left := element.rect.X + element.layout.Padding.Left
	top := element.rect.Y + element.layout.Padding.Top
	for first := element.firstChild; first >= 0; {
		line := ui.lineFrom(element, first, inner)
		x := left + alignOffset(element.layout.AlignX, inner-line.width)
		for child := line.first; child != line.next; child = ui.elements[child].nextSibling {
			childElement := &ui.elements[child]
			childElement.rect.X = x
			childElement.rect.Y = top + alignOffset(element.layout.AlignY, line.height-childElement.rect.Height)
			x += childElement.rect.Width + element.layout.Gap
		}
		top += line.height + element.layout.lineGap()
		first = line.next
	}
}
