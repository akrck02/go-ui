package uitest

import (
	"fmt"
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

// Element finds what the last frame showed with this label: a button's text, a field's name, a
// line of text. A label shown twice gives ElementAmbiguous: use ElementIn with its scope. A label
// laid out but scrolled away or behind a dialog is found with Visible false.
func (window *Window) Element(label string) (Element, error) {
	matches := window.nodesLabelled(label)
	if len(matches) > 1 {
		return Element{}, window.noted(elementAmbiguous(label, len(matches)))
	}
	if len(matches) == 1 {
		return window.elementOf(matches[0]), nil
	}
	if texts := window.textsDrawn(label); len(texts) > 0 {
		if len(texts) > 1 {
			return Element{}, window.noted(elementAmbiguous(label, len(texts)))
		}
		return texts[0], nil
	}
	if area, found := window.context.Area(label); found {
		return Element{Label: label, Rect: area}, nil
	}
	return Element{}, window.noted(elementNotFound(label, window.labels()))
}

// textsDrawn is the lines of text the last frame drew that say exactly this: a screen reader reads
// the texts of one box as one node, so a line on its own is found in the draw list. The list drops
// lines clipped away; a line kept shows unless a layer above (a dialog's dimming, a popup) covers it.
func (window *Window) textsDrawn(text string) []Element {
	var found []Element
	for _, command := range window.DrawList().Commands {
		if command.Kind != draw2d.TextKind || command.Text != text {
			continue
		}
		area := command.Area // a text's area has no width in the list: measure it
		area.Width = window.measurer.MeasureText(text, command.Size, command.Weight, 0).Width
		found = append(found, Element{Label: text, Rect: area, Visible: !window.coveredFromAbove(area, command.Layer)})
	}
	return found
}

// ElementIn finds a control by its scopes (the names given to ui.Context.Push) and its label,
// outermost first: ElementIn("north", "Save").
func (window *Window) ElementIn(path ...string) (Element, error) {
	name := strings.Join(path, " / ")
	if len(path) == 0 {
		return Element{}, window.noted(elementNotFound(name, window.labels()))
	}
	area, found := window.context.Area(path...)
	if !found {
		return Element{}, window.noted(elementNotFound(name, window.labels()))
	}
	for _, node := range window.nodesLabelled(path[len(path)-1]) {
		if contains(area, node.Area) {
			return window.elementOf(node), nil
		}
	}
	return Element{Label: path[len(path)-1], Rect: area}, nil
}

// nodesLabelled is the screen reader's nodes of the last frame with this label.
func (window *Window) nodesLabelled(label string) []ui.AccessibleNode {
	var matches []ui.AccessibleNode
	for _, node := range window.context.ScreenReaderNodes() {
		if node.Label == label {
			matches = append(matches, node)
		}
	}
	return matches
}

// elementOf turns a screen reader's node into an Element.
func (window *Window) elementOf(node ui.AccessibleNode) Element {
	return Element{
		ID: node.ID, Label: node.Label, Rect: node.Area,
		Visible:   node.Area.Width > 0 && node.Area.Height > 0,
		Focused:   node.ID != 0 && node.ID == window.context.FocusedID(),
		Pressable: node.Pressable,
	}
}

// labels is every label of the last frame, for error messages.
func (window *Window) labels() []string {
	var labels []string
	for _, node := range window.context.ScreenReaderNodes() {
		labels = append(labels, node.Label)
	}
	return labels
}

// Texts is every line of text the last frame drew, in drawing order: headings, labels, buttons'
// texts, a field's text.
func (window *Window) Texts() []string {
	var texts []string
	for _, command := range window.DrawList().Commands {
		if command.Kind == draw2d.TextKind {
			texts = append(texts, command.Text)
		}
	}
	return texts
}

// coveredFromAbove tells whether a filled rectangle of a higher layer covers an area.
func (window *Window) coveredFromAbove(area draw2d.Rect, layer int) bool {
	for _, command := range window.DrawList().Commands {
		covers := command.Kind == draw2d.RectangleKind && command.Layer > layer && command.Fill.IsVisible()
		if covers && contains(command.Area, area) {
			return true
		}
	}
	return false
}

// contains tells whether a rectangle holds another.
func contains(outer, inner draw2d.Rect) bool {
	return inner.X >= outer.X && inner.Y >= outer.Y &&
		inner.X+inner.Width <= outer.X+outer.Width && inner.Y+inner.Height <= outer.Y+outer.Height
}

// describe is a short form of a label for messages.
func describe(label string) string { return fmt.Sprintf("%q", label) }
