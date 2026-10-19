package ui

import (
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
)

// paint turns the laid-out trees into the draw list, each root on its layer.
func (ui *Context) paint() {
	ui.drawing.Reset()
	for _, placed := range ui.roots {
		ui.drawing.SetLayer(ui.layerOf(placed.id))
		ui.clip, ui.clipped = draw2d.Rect{}, false
		first := len(ui.drawing.Commands)
		ui.paintElement(placed.element)
		ui.keepForLeaving(placed.id, placed.options.Leave, first)
	}
	ui.drawing.SetLayer(0)
	ui.paintLeaving()
}

func (ui *Context) paintElement(index int) {
	element := &ui.elements[index]
	if element.id != 0 && ui.clipped {
		// what the pointer can reach is what shows: a row scrolled out of a list's view must not
		// take a touch meant for the bar over it
		ui.currentReach[element.id] = element.rect.Intersect(ui.clip)
	}
	switch element.kind {
	case textElement:
		ui.paintText(element)
	case iconElement:
		ui.paintIcon(element)
	default:
		ui.paintBox(element)
	}
}

func (ui *Context) paintBox(element *element) {
	if ui.shrinks(element) {
		first := len(ui.drawing.Commands)
		defer ui.scaleCommands(first, element.rect, ui.Theme.PressScale)
	}
	look := element.look
	if look.Shadow {
		ui.drawing.Shadow(element.rect, look.Radius, ui.Theme.ShadowBlur, ui.Theme.ShadowDrop, ui.Theme.Shadow)
	}
	frosted := ui.isFrosted(look)
	if frosted {
		ui.drawing.Backdrop(element.rect, look.Radius, ui.Theme.Backdrop)
	}
	if look.EdgeWidth > 0 && look.Edge.IsVisible() {
		under := element.rect
		under.Y += look.EdgeWidth
		ui.drawing.Rectangle(under, look.Radius, look.Edge, draw2d.Transparent, 0)
	}
	ui.drawing.Rectangle(element.rect, look.Radius, look.Fill, look.Border, look.BorderWidth)
	if look.Backdrop != nil {
		look.Backdrop.Paint(element.rect, &ui.drawing)
	}
	if element.field != nil {
		ui.paintField(element)
	}
	if element.painter != nil {
		element.painter(element.rect, &ui.drawing)
	}
	outerClip, outerClipped := ui.clip, ui.clipped
	if look.Clip {
		ui.drawing.PushClip(element.rect)
		ui.clip, ui.clipped = element.rect, true
		if outerClipped {
			ui.clip = element.rect.Intersect(outerClip)
		}
	}
	outerFrost := ui.insideFrost
	ui.insideFrost = outerFrost || frosted
	right := element.rect.Right() - element.layout.Padding.Right
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		if ui.isOutOfSight(&ui.elements[child]) {
			ui.outOfReach(child)
			continue
		}
		if look.Clip && ui.elements[child].kind == textElement {
			ui.paintTextWithin(&ui.elements[child], right)
			continue
		}
		ui.paintElement(child)
	}
	if element.scroll != nil {
		ui.paintScrollBar(element)
	}
	ui.paintFocusRing(element)
	ui.insideFrost = outerFrost
	if look.Clip {
		ui.drawing.PopClip()
		ui.clip, ui.clipped = outerClip, outerClipped
	}
}

// isFrosted tells whether a box blurs what is behind it: with a Backdrop in the theme, a box with
// a translucent fill that is not inside another frosted one. A frosted panel's own hovered or
// chosen rows are translucent too, but blurring again under them would cost a blur for nothing.
func (ui *Context) isFrosted(look Look) bool {
	if ui.Theme.Backdrop <= 0 || ui.insideFrost {
		return false
	}
	alpha := look.Fill[3]
	return alpha > 0 && alpha < 1
}

// isOutOfSight tells whether an element lies wholly outside the clip of a box around it, with
// nothing overflowing it that could reach back in: the rows of a scrolled table out of view. Such
// elements are not painted at all.
func (ui *Context) isOutOfSight(element *element) bool {
	if !ui.clipped || !ui.clip.Intersect(element.rect).IsEmpty() {
		return false
	}
	fits := element.content[Horizontal] <= element.rect.Width+0.5 && element.content[Vertical] <= element.rect.Height+0.5
	return fits
}

// paintText centres the line vertically in its box.
func (ui *Context) paintText(element *element) {
	metrics := element.metrics
	baseline := element.rect.Y + (element.rect.Height-metrics.Height())/2 + metrics.Ascent
	if element.style.Shadow.IsVisible() {
		ui.drawing.Text(element.text, element.rect.X+1, baseline+1, element.style.Size, element.style.Weight, element.style.Shadow)
	}
	ui.drawing.Text(element.text, element.rect.X, baseline, element.style.Size, element.style.Weight, element.style.Color)
}

// paintTextWithin paints a line of text in a box that clips, ending it with "…" where it would
// pass the box's inner edge, so a cell of a narrow column reads "Rocky isl…" instead of a word cut
// in half or running under the next column.
func (ui *Context) paintTextWithin(element *element, right float32) {
	room := right - element.rect.X
	if element.metrics.Width <= room+0.5 {
		ui.paintText(element)
		return
	}
	shortened := *element
	shortened.text = ui.elide(element.text, element.style, room)
	ui.paintText(&shortened)
}

// elide is the longest start of a text that fits a width with "…" after it.
func (ui *Context) elide(text string, style TextStyle, width float32) string {
	runes := []rune(text)
	low, high := 0, len(runes)
	for low < high { // the most characters that fit
		middle := (low + high + 1) / 2
		if ui.measure(string(runes[:middle])+"…", style).Width <= width {
			low = middle
		} else {
			high = middle - 1
		}
	}
	if low == 0 {
		return ""
	}
	return strings.TrimRight(string(runes[:low]), " ") + "…"
}

// paintIcon centres the icon in its box.
func (ui *Context) paintIcon(element *element) {
	size := element.style.Size
	x := element.rect.X + (element.rect.Width-size)/2
	y := element.rect.Y + (element.rect.Height-size)/2
	ui.drawing.Icon(element.icon, x, y, size, element.color)
}

// textKey names a measured text.
type textKey struct {
	text   string
	size   float32
	weight draw2d.Weight
	scale  float32
}

// textCache keeps the measures of texts, which an interface repeats every frame. It has two
// generations: the texts measured since the last turnover, and the ones before. A text found in the
// older one moves to the newer one, so whatever a frame uses is never dropped in the middle of it;
// texts that stop being used (a counter's old values) go when the older generation is dropped.
type textCache struct {
	entries, older map[textKey]draw2d.TextMetrics
}

// maximumCachedTexts is how many texts the newer generation gathers before it becomes the older
// one, at the start of a frame. A frame that shows more keeps them all until it ends.
const maximumCachedTexts = 4096

// turnOver starts a new generation when the current one is full; called at the start of a frame.
func (cache *textCache) turnOver() {
	if len(cache.entries) < maximumCachedTexts {
		return
	}
	clear(cache.older)
	cache.entries, cache.older = cache.older, cache.entries
}

func (ui *Context) measure(text string, style TextStyle) draw2d.TextMetrics {
	key := textKey{text: text, size: style.Size, weight: style.Weight, scale: ui.scale}
	if metrics, found := ui.texts.entries[key]; found {
		return metrics
	}
	metrics, found := ui.texts.older[key]
	if !found {
		metrics = ui.measurer.MeasureText(text, style.Size, style.Weight, ui.scale)
	}
	ui.texts.entries[key] = metrics
	return metrics
}

// outOfReach takes an element out of sight, and everything in it, from the pointer's reach.
func (ui *Context) outOfReach(index int) {
	element := &ui.elements[index]
	if element.id != 0 {
		ui.currentReach[element.id] = draw2d.Rect{}
	}
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		ui.outOfReach(child)
	}
}
