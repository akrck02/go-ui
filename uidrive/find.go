package uidrive

import (
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Element is what the last frame showed of a control, a text or a scroll area.
type Element struct {
	ID        ui.ID       // the one the interface gives it, 0 when it was not on screen
	Label     string      // what a screen reader calls it: a button's text, a described label
	Rect      draw2d.Rect // where it was laid out, in layout pixels
	Visible   bool        // some of it shows: not scrolled away, not behind a dialog or sheet
	Focused   bool        // it has the focus (ui.Context.FocusedID)
	Pressable bool        // a click or a tap does something to it
}

// Find finds what the last frame showed with this label: a button's text, a field's name, a line
// of text. The context's screen reader must be on (ui.Context.SetScreenReader), since its nodes
// are what is searched. A label shown twice gives ElementAmbiguous: use FindIn with its scope. A
// label laid out but scrolled away or behind a dialog is found with Visible false.
func Find(context *ui.Context, label string) (Element, error) {
	matches := nodesLabelled(context, label)
	if len(matches) > 1 {
		return Element{}, elementAmbiguous(label, len(matches))
	}
	if len(matches) == 1 {
		return elementOf(context, matches[0]), nil
	}
	if texts := textsDrawn(context, label); len(texts) > 0 {
		if len(texts) > 1 {
			return Element{}, elementAmbiguous(label, len(texts))
		}
		return texts[0], nil
	}
	if area, found := context.Area(label); found {
		return Element{Label: label, Rect: area}, nil
	}
	return Element{}, elementNotFound(label, labels(context))
}

// FindIn finds a control by its scopes (the names given to ui.Context.Push) and its label,
// outermost first: FindIn(context, "north", "Save").
func FindIn(context *ui.Context, path ...string) (Element, error) {
	name := strings.Join(path, " / ")
	if len(path) == 0 {
		return Element{}, elementNotFound(name, labels(context))
	}
	area, found := context.Area(path...)
	if !found {
		return Element{}, elementNotFound(name, labels(context))
	}
	for _, node := range nodesLabelled(context, path[len(path)-1]) {
		if contains(area, node.Area) {
			return elementOf(context, node), nil
		}
	}
	return Element{Label: path[len(path)-1], Rect: area}, nil
}

// FindWithin finds an element by its label inside another found by its label: the button "Save"
// in the panel called "North". Inside means within the smallest painted box (a panel, a card, a
// row) that holds the outer element, so a title finds what its panel holds.
func FindWithin(context *ui.Context, outer, inner string) (Element, error) {
	container, err := Find(context, outer)
	if err != nil {
		return Element{}, err
	}
	area := boxAround(context.Drawing(), container.Rect)
	var found []Element
	for _, node := range nodesLabelled(context, inner) {
		if contains(area, node.Area) {
			found = append(found, elementOf(context, node))
		}
	}
	if len(found) == 0 { // a line of text on its own, as Find looks for it
		for _, text := range textsDrawn(context, inner) {
			if contains(area, text.Rect) {
				found = append(found, text)
			}
		}
	}
	name := outer + " > " + inner
	if len(found) == 0 {
		return Element{}, elementNotFound(name, labels(context))
	}
	if len(found) > 1 {
		return Element{}, elementAmbiguous(name, len(found))
	}
	return found[0], nil
}

// FindVisible finds an element that shows, or says why it cannot be used: ElementNotVisible when
// it is scrolled away or behind a dialog.
func FindVisible(context *ui.Context, label string) (Element, error) {
	found, err := Find(context, label)
	if err != nil {
		return Element{}, err
	}
	if !found.Visible {
		return Element{}, elementNotVisible(label)
	}
	return found, nil
}

// FindPressable finds an element that shows and can be pressed: ElementNotPressable when a click
// on it would do nothing.
func FindPressable(context *ui.Context, label string) (Element, error) {
	found, err := FindVisible(context, label)
	if err != nil {
		return Element{}, err
	}
	if !found.Pressable {
		return Element{}, elementNotPressable(label)
	}
	return found, nil
}

// FindAt finds the smallest named element of the last frame under a place in the window, in
// layout pixels: what a person pressed there. It tells false when nothing named is there.
func FindAt(context *ui.Context, x, y float32) (Element, bool) {
	var smallest ui.AccessibleNode
	found := false
	for _, node := range context.ScreenReaderNodes() {
		area := node.Area
		inside := x >= area.X && y >= area.Y && x < area.X+area.Width && y < area.Y+area.Height
		if !inside || node.Label == "" {
			continue
		}
		if !found || area.Width*area.Height < smallest.Area.Width*smallest.Area.Height {
			smallest, found = node, true
		}
	}
	if !found {
		return Element{}, false
	}
	return elementOf(context, smallest), true
}

// Texts is every line of text the last frame drew, in drawing order: headings, labels, buttons'
// texts, a field's text.
func Texts(context *ui.Context) []string {
	var texts []string
	for _, command := range context.Drawing().Commands {
		if command.Kind == draw2d.TextKind {
			texts = append(texts, command.Text)
		}
	}
	return texts
}

// textsDrawn is the lines of text the last frame drew that say exactly this: a screen reader reads
// the texts of one box as one node, so a line on its own is found in the draw list. The list drops
// lines clipped away; a line kept shows unless a layer above (a dialog's dimming, a popup) covers it.
func textsDrawn(context *ui.Context, text string) []Element {
	var found []Element
	drawing := context.Drawing()
	for _, command := range drawing.Commands {
		if command.Kind != draw2d.TextKind || command.Text != text {
			continue
		}
		area := command.Area // a text's area has no width in the list: measure it
		area.Width = context.Measurer().MeasureText(text, command.Size, command.Weight, 0).Width
		found = append(found, Element{Label: text, Rect: area, Visible: !coveredFromAbove(drawing, area, command.Layer)})
	}
	return found
}

// nodesLabelled is the screen reader's nodes of the last frame with this label.
func nodesLabelled(context *ui.Context, label string) []ui.AccessibleNode {
	var matches []ui.AccessibleNode
	for _, node := range context.ScreenReaderNodes() {
		if node.Label == label {
			matches = append(matches, node)
		}
	}
	return matches
}

// elementOf turns a screen reader's node into an Element.
func elementOf(context *ui.Context, node ui.AccessibleNode) Element {
	return Element{
		ID: node.ID, Label: node.Label, Rect: node.Area,
		Visible:   node.Area.Width > 0 && node.Area.Height > 0,
		Focused:   node.ID != 0 && node.ID == context.FocusedID(),
		Pressable: node.Pressable,
	}
}

// labels is every label of the last frame, for error messages.
func labels(context *ui.Context) []string {
	var labels []string
	for _, node := range context.ScreenReaderNodes() {
		labels = append(labels, node.Label)
	}
	return labels
}

// coveredFromAbove tells whether a filled rectangle of a higher layer covers an area.
func coveredFromAbove(drawing *draw2d.List, area draw2d.Rect, layer int) bool {
	for _, command := range drawing.Commands {
		covers := command.Kind == draw2d.RectangleKind && command.Layer > layer && command.Fill.IsVisible()
		if covers && contains(command.Area, area) {
			return true
		}
	}
	return false
}

// boxAround is the smallest rectangle a drawing painted that holds an area, or the area itself
// when nothing painted holds it.
func boxAround(drawing *draw2d.List, area draw2d.Rect) draw2d.Rect {
	smallest, found := area, false
	for _, command := range drawing.Commands {
		painted := command.Fill.IsVisible() || command.Border.IsVisible()
		if command.Kind != draw2d.RectangleKind || !painted || !contains(command.Area, area) {
			continue
		}
		if !found || command.Area.Width*command.Area.Height < smallest.Width*smallest.Height {
			smallest, found = command.Area, true
		}
	}
	return smallest
}

// contains tells whether a rectangle holds another.
func contains(outer, inner draw2d.Rect) bool {
	return inner.X >= outer.X && inner.Y >= outer.Y &&
		inner.X+inner.Width <= outer.X+outer.Width && inner.Y+inner.Height <= outer.Y+outer.Height
}
