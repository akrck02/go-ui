package ui

import "github.com/akrck02/go-renderer/draw2d"

// Look is how a box is painted: a fill, a border, rounded corners and a shadow. The zero value
// paints nothing, so a box is only a way to arrange its children.
type Look struct {
	Fill        draw2d.Color
	Border      draw2d.Color
	BorderWidth float32
	Radius      float32
	Shadow      bool // a soft shadow under it, for things that float over others
	Clip        bool // children are cut to the box's area
	// Backdrop, when set, draws more of the box's background, over Fill and Border and under its
	// children: a frame of several colours, a pattern. A pointer to a style kept from frame to
	// frame allocates nothing.
	Backdrop Backdrop
	// Edge is a solid band of this colour, EdgeWidth tall, just under the box, as if it stood on a
	// block: a design system's main button. It takes no room in the layout, like a shadow.
	Edge      draw2d.Color
	EdgeWidth float32
	// Fade makes the box and all it holds that much see-through, from 0 (as it is) to 1 (gone): a
	// screen fading in with Context.Appear.
	Fade float32
	// ShiftX and ShiftY draw the box and all it holds that many pixels away, without moving it in
	// the layout or for the pointer: a dialog rising into place, a note sliding in.
	ShiftX, ShiftY float32
}

// Backdrop draws a box's background in its area.
type Backdrop interface {
	Paint(area draw2d.Rectangle, drawing *draw2d.List)
}

type elementKind int

const (
	boxElement elementKind = iota
	textElement
	iconElement
)

// TextStyle is how a text is written.
type TextStyle struct {
	Size   float32 // in pixels; 0 = the theme's text size
	Weight draw2d.Weight
	Color  draw2d.Color // the zero colour = the theme's ink
	// Shadow, when visible, draws the text a second time a pixel right and down, in this colour,
	// under it: the hard shadow pixel-art games give their text.
	Shadow draw2d.Color
	// LineHeight, when above 0, makes this line as tall as its size × this, with the letters in
	// the middle, as CSS's line-height does; 0 follows the theme's LineHeight, and
	// LettersLineHeight keeps the line as tall as its letters.
	LineHeight float32
}

// LettersLineHeight is a TextStyle.LineHeight that keeps a line as tall as its letters, whatever
// the theme says: CSS's line-height: normal, which buttons and tabs often use.
const LettersLineHeight float32 = -1

// element is one box of this frame's tree: a container, a line of text or an icon. Elements live
// in one slice reused every frame and point at each other by index (-1 = none), so building the
// tree allocates nothing once the slice has grown.
type element struct {
	id     ID // 0 for boxes nobody asks about again
	kind   elementKind
	layout Layout
	look   Look

	text    string
	style   TextStyle
	metrics draw2d.TextMetrics
	icon    string
	color   draw2d.Color // the icon's colour

	parent, firstChild, lastChild, nextSibling int
	rect                                       draw2d.Rectangle
	capped                                     [2]bool                                           // reached its maximum while growing, per axis
	content                                    [2]float32                                        // what the content needs, per axis, padding included
	scroll                                     *scrollMemory                                     // set on scroll areas
	field                                      *fieldPaint                                       // set on text fields
	painter                                    func(area draw2d.Rectangle, drawing *draw2d.List) // set on painted boxes (Paint)
	description                                int                                               // 1 + its index in the screen reader's descriptions; 0 = none
	lineBox                                    float32                                           // a text's height in the layout: its letters, or the theme's line height
	filledWithPrimary                          bool                                              // a primary button: its focus ring goes inside
}

// intrinsic is the size of what the element shows itself (not its children) along an axis.
func (element *element) intrinsic(axis Direction) float32 {
	switch element.kind {
	case textElement:
		if axis == Horizontal {
			return element.metrics.Width
		}
		return max(element.metrics.Height(), element.lineBox)
	case iconElement:
		return element.style.Size
	default:
		return 0
	}
}

func (element *element) sizeAlong(axis Direction) float32 {
	if axis == Horizontal {
		return element.rect.Width
	}
	return element.rect.Height
}

func (element *element) setSize(axis Direction, size float32) {
	if axis == Horizontal {
		element.rect.Width = size
	} else {
		element.rect.Height = size
	}
}

func (element *element) positionAlong(axis Direction) float32 {
	if axis == Horizontal {
		return element.rect.X
	}
	return element.rect.Y
}

func (element *element) setPosition(axis Direction, position float32) {
	if axis == Horizontal {
		element.rect.X = position
	} else {
		element.rect.Y = position
	}
}

func (element *element) isCapped(axis Direction) bool { return element.capped[axis] }

func (element *element) capAxis(axis Direction) { element.capped[axis] = true }

// openElement adds an element under the box open now and opens it, so what is added next goes
// inside it until closeElement. It returns the element's index.
func (ui *Context) openElement(added element) int {
	added.parent, added.firstChild, added.lastChild, added.nextSibling = -1, -1, -1, -1
	index := len(ui.elements)
	if len(ui.open) > 0 {
		parentIndex := ui.open[len(ui.open)-1]
		added.parent = parentIndex
		parent := &ui.elements[parentIndex]
		if parent.lastChild >= 0 {
			ui.elements[parent.lastChild].nextSibling = index
		} else {
			parent.firstChild = index
		}
		parent.lastChild = index
	}
	ui.elements = append(ui.elements, added)
	ui.open = append(ui.open, index)
	return index
}

// closeElement closes the box opened last.
func (ui *Context) closeElement() {
	ui.finished = ui.open[len(ui.open)-1] // what Describe describes
	if len(ui.open) > 1 {
		ui.open = ui.open[:len(ui.open)-1]
	}
}

// addLeaf adds an element with no children (a text, an icon).
func (ui *Context) addLeaf(added element) int {
	index := ui.openElement(added)
	ui.closeElement()
	return index
}

// parentDirection is the direction of the box open now, so spacers know which way to grow.
func (ui *Context) parentDirection() Direction {
	if len(ui.open) == 0 {
		return Vertical
	}
	return ui.elements[ui.open[len(ui.open)-1]].layout.Direction
}

// Box adds a box with a layout, and what content adds goes inside it.
func (ui *Context) Box(layout Layout, content func()) {
	ui.Frame(layout, Look{}, content)
}

// Frame adds a box with a layout and a look (fill, border, corners, shadow).
func (ui *Context) Frame(layout Layout, look Look, content func()) {
	ui.frameWithID(0, layout, look, content)
}

func (ui *Context) frameWithID(id ID, layout Layout, look Look, content func()) int {
	index := ui.openElement(element{id: id, kind: boxElement, layout: layout, look: look})
	if content != nil {
		content()
	}
	ui.closeElement()
	return index
}

// Row lines up what content adds from left to right, across the whole width, centred vertically.
func (ui *Context) Row(content func()) {
	ui.Box(Layout{Direction: Horizontal, Width: Grow(), Gap: ui.Theme.Gap, AlignY: Center}, content)
}

// Column stacks what content adds from top to bottom, across the whole width.
func (ui *Context) Column(content func()) {
	ui.Box(Layout{Direction: Vertical, Width: Grow(), Gap: ui.Theme.Gap}, content)
}

// Spacer takes the free space of the row or column it is in, pushing what follows to the end.
func (ui *Context) Spacer() {
	layout := Layout{Width: Grow()}
	if ui.parentDirection() == Vertical {
		layout = Layout{Height: Grow()}
	}
	ui.addLeaf(element{kind: boxElement, layout: layout})
}

// Separator draws a thin line across the row or column it is in.
func (ui *Context) Separator() {
	layout := Layout{Width: Grow(), Height: Fixed(1)}
	if ui.parentDirection() == Horizontal {
		layout = Layout{Width: Fixed(1), Height: Grow()}
	}
	ui.addLeaf(element{kind: boxElement, layout: layout, look: Look{Fill: ui.Theme.Line}})
}

// Text adds a line of text in a style.
func (ui *Context) Text(text string, style TextStyle) {
	ui.addLeaf(ui.textElement(text, style))
}

func (ui *Context) textElement(text string, style TextStyle) element {
	if style.Size <= 0 {
		style.Size = ui.Theme.TextSize
	}
	style.Size = ui.scaledText(style.Size)
	if !style.Color.IsVisible() {
		style.Color = ui.Theme.Ink
	}
	metrics := ui.measure(text, style)
	return element{kind: textElement, text: text, style: style, metrics: metrics, lineBox: ui.Theme.lineBoxOf(style, metrics.Height())}
}

// Label adds a line of body text.
func (ui *Context) Label(text string) { ui.Text(text, TextStyle{}) }

// Heading adds a line of bold, larger text: the title of a group.
func (ui *Context) Heading(text string) {
	ui.Text(text, TextStyle{Size: ui.Theme.HeadingSize, Weight: draw2d.Bold})
	ui.Describe(Description{Role: HeadingRole})
}

// Caption adds a line of small, quiet text: a hint or a unit.
func (ui *Context) Caption(text string) {
	ui.Text(text, TextStyle{Size: ui.Theme.SmallTextSize, Color: ui.Theme.InkQuiet})
}

// Icon adds an icon (a Lucide name, https://lucide.dev) at the theme's icon size.
func (ui *Context) Icon(name string) { ui.IconSized(name, ui.Theme.IconSize, ui.Theme.InkSoft) }

// IconSized adds an icon of a size and colour.
func (ui *Context) IconSized(name string, size float32, color draw2d.Color) {
	ui.addLeaf(ui.iconElement(name, size, color))
}

func (ui *Context) iconElement(name string, size float32, color draw2d.Color) element {
	return element{kind: iconElement, icon: name, style: TextStyle{Size: size}, color: color}
}
