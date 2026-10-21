package uitest

import "github.com/akrck02/go-ui/uidrive"

// Element is what the last frame showed of a control, a text or a scroll area (uidrive.Element).
type Element = uidrive.Element

// Element finds what the last frame showed with this label: a button's text, a field's name, a
// line of text. A label shown twice gives ElementAmbiguous: use ElementIn with its scope. A label
// laid out but scrolled away or behind a dialog is found with Visible false.
func (window *Window) Element(label string) (Element, error) {
	found, err := uidrive.Find(window.context, label)
	return found, window.noted(err)
}

// ElementIn finds a control by its scopes (the names given to ui.Context.Push) and its label,
// outermost first: ElementIn("north", "Save").
func (window *Window) ElementIn(path ...string) (Element, error) {
	found, err := uidrive.FindIn(window.context, path...)
	return found, window.noted(err)
}

// ElementWithin finds an element by its label inside another found by its label: the button
// "Save" in the panel called "North". Inside means within the smallest painted box (a panel, a
// card, a row) that holds the outer element, so a title finds what its panel holds.
func (window *Window) ElementWithin(outer, inner string) (Element, error) {
	found, err := uidrive.FindWithin(window.context, outer, inner)
	return found, window.noted(err)
}

// Texts is every line of text the last frame drew, in drawing order: headings, labels, buttons'
// texts, a field's text.
func (window *Window) Texts() []string { return uidrive.Texts(window.context) }
