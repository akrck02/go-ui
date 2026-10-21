package ui

import (
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
)

// A screen reader (TalkBack on a phone) speaks what is on the screen and lets a person who cannot
// see it act on it: a finger moved over the screen reads what is under it, a double tap presses
// it. It needs to know what the screen holds as things with names, not as pixels: while one is
// on (SetScreenReader), each frame also makes a list of what can be read (ScreenReaderNodes),
// worked out from the boxes themselves. A control is a node read by the texts inside it; texts
// outside controls are read together with the texts beside them in the same box. Describe says
// more where the boxes cannot tell: that a box is a switch and is on.

// Role is what a thing on the screen is, which a screen reader says after its name: "Wi-Fi,
// switch, on".
type Role int

const (
	NoRole       Role = iota // worked out from the box: a control is a button, anything else text
	ButtonRole               // pressed to do something
	CheckboxRole             // ticked or not
	SwitchRole               // on or off
	RadioRole                // one of a few choices, of which one is chosen
	SliderRole               // a value between two ends (Range)
	FieldRole                // typed in
	TabRole                  // one of the pages over a screen, one of them selected
	HeadingRole              // the title of what follows
	ImageRole                // a picture, read by its label
	TextRole                 // something to read
	ScrollRole               // an area that scrolls (ScrollArea)
)

// Range is a slider's span and where it is now.
type Range struct {
	Minimum, Maximum, Current float32
}

// Description is what a screen reader says about a thing, beyond what its boxes tell.
type Description struct {
	Role  Role
	Label string // what it is called; empty = the texts inside it
	Value string // what it holds, read after the label: a field's text
	// Checked is a checkbox ticked, a switch on or a radio button chosen.
	Checked  bool
	Selected bool // the tab, the destination or the choice shown
	Disabled bool // shown, but does nothing now
	Secret   bool // a field whose text is not read aloud (a password)
	// Hidden leaves it out, with all it holds: a decoration, or something repeated just beside it.
	Hidden bool
	Range  Range // a slider's; Maximum above Minimum when it has one
}

// merge adds what another description says to this one: a control's own (a field's text) and what
// its caller adds (the label above it). What the other leaves empty stays.
func (description *Description) merge(other Description) {
	if other.Role != NoRole {
		description.Role = other.Role
	}
	if other.Label != "" {
		description.Label = other.Label
	}
	if other.Value != "" {
		description.Value = other.Value
	}
	if other.Range.Maximum > other.Range.Minimum {
		description.Range = other.Range
	}
	description.Checked = description.Checked || other.Checked
	description.Selected = description.Selected || other.Selected
	description.Disabled = description.Disabled || other.Disabled
	description.Secret = description.Secret || other.Secret
	description.Hidden = description.Hidden || other.Hidden
}

// AccessibleNode is one thing a screen reader can read and act on.
type AccessibleNode struct {
	ID     ID          // the same from one frame to the next while the thing stays
	Parent ID          // the scroll area it is in, or 0 for the window
	Area   draw2d.Rect // where it shows, in layout pixels: the part a scroll area does not hide
	Description
	Pressable     bool // a double tap presses it (ReaderPress)
	LongPressable bool // the program reads LongPressed after it (ReaderLongPress)
	Scrollable    bool // ReaderForward and ReaderBackward scroll it, or move a slider
	Typing        bool // the keyboard types in it now
}

// ReaderAction is what a screen reader asks of a node.
type ReaderAction int

const (
	ReaderPress     ReaderAction = iota + 1 // a double tap: the control is clicked
	ReaderLongPress                         // a double tap held: the control is long-pressed
	ReaderForward                           // a scroll area shows what comes next; a slider goes up a step
	ReaderBackward                          // a scroll area shows what came before; a slider goes down a step
)

// readerState is what the interface keeps for a screen reader.
type readerState struct {
	on            bool
	descriptions  []Description // this frame's, pointed at by element.description
	controls      map[ID]bool   // controls that looked at the mouse this frame
	longPressable map[ID]bool   // controls the program asked LongPressed of this frame
	nodes         []AccessibleNode
	seen          map[ID]int // ids given to nodes this frame, to tell repeats apart
	focus         ID         // the node the screen reader is on, ringed
	asked         map[ID]ReaderAction
	acting        map[ID]ReaderAction // the actions of this frame
	textAsked     map[ID]string       // fields' new texts, by ScreenReaderSetText
	texts         map[ID]string       // the new texts of this frame
	tooltips      map[ID]string       // the tooltips of this frame, by control: an icon button's name
}

// SetScreenReader tells the interface whether a screen reader is on; the window does, on a phone.
// While one is on, each frame makes its list of nodes (ScreenReaderNodes); while none is, nothing
// is spent on it.
func (ui *Context) SetScreenReader(on bool) {
	ui.reader.on = on
	if !on {
		ui.reader.nodes, ui.reader.focus = ui.reader.nodes[:0], 0
	}
}

// ScreenReaderOn tells whether a screen reader is on (SetScreenReader).
func (ui *Context) ScreenReaderOn() bool { return ui.reader.on }

// ScreenReaderNodes is what the last frame showed, for a screen reader, in reading order: the
// window's own first, then what panels, sheets and popups hold, front-most last. While a modal or
// a sheet is shown, what is behind it is left out, as it cannot be touched. Empty while no screen
// reader is on. The list is reused by the next frame.
func (ui *Context) ScreenReaderNodes() []AccessibleNode { return ui.reader.nodes }

// SetScreenReaderFocus says which node the screen reader is on (0 for none), so the interface
// rings it, as a phone rings what TalkBack reads.
func (ui *Context) SetScreenReaderFocus(id ID) { ui.reader.focus = id }

// ScreenReaderAct does what a screen reader asks of a node, in the next frame: a press clicks the
// control as a tap would, so whatever the program does on a click happens.
func (ui *Context) ScreenReaderAct(id ID, action ReaderAction) {
	if ui.reader.asked == nil {
		ui.reader.asked = map[ID]ReaderAction{}
	}
	ui.reader.asked[id] = action
}

// ScreenReaderSetText replaces a field's whole text with what a screen reader typed, pasted or
// dictated into it, in the next frame, without bringing the keyboard up. The field takes it as
// typing: one line, cut at its MaxLength.
func (ui *Context) ScreenReaderSetText(id ID, text string) {
	if ui.reader.textAsked == nil {
		ui.reader.textAsked = map[ID]string{}
	}
	ui.reader.textAsked[id] = text
}

// Describe tells a screen reader more about the box added just before it (a control, a text): what
// it is, its name, its value and its state. What it leaves empty stays as the box tells it. Call it
// in every frame, after the control: a switch says it is a switch, and whether it is on.
//
//	if look.Context.Clickable("wifi", layout, looks, drawSwitch) {
//		on = !on
//	}
//	look.Context.Describe(ui.Description{Role: ui.SwitchRole, Label: "Wi-Fi", Checked: on})
func (ui *Context) Describe(description Description) {
	ui.describeElement(ui.finished, description)
}

// describeElement adds a description to an element of this frame.
func (ui *Context) describeElement(index int, description Description) {
	if !ui.reader.on || index < 0 || index >= len(ui.elements) {
		return
	}
	element := &ui.elements[index]
	if element.description == 0 {
		ui.reader.descriptions = append(ui.reader.descriptions, description)
		element.description = len(ui.reader.descriptions)
		return
	}
	ui.reader.descriptions[element.description-1].merge(description)
}

// beginReading starts a frame's reading: the actions asked since the last frame are this frame's.
func (ui *Context) beginReading() {
	reader := &ui.reader
	reader.descriptions = reader.descriptions[:0]
	clear(reader.controls)
	clear(reader.longPressable)
	clear(reader.tooltips)
	reader.acting, reader.asked = reader.asked, reader.acting
	clear(reader.asked)
	reader.texts, reader.textAsked = reader.textAsked, reader.texts
	clear(reader.textAsked)
}

// readerText is the text a screen reader put in a field for this frame, if it did.
func (ui *Context) readerText(id ID) (string, bool) {
	text, asked := ui.reader.texts[id]
	return text, asked
}

// readerActs tells whether a screen reader asked something of a control in this frame.
func (ui *Context) readerActs(id ID, action ReaderAction) bool {
	return id != 0 && ui.reader.acting[id] == action
}

// readerSteps is how many steps a screen reader moves a slider in this frame: 1, -1 or 0.
func (ui *Context) readerSteps(id ID) float32 {
	switch {
	case ui.readerActs(id, ReaderForward):
		return 1
	case ui.readerActs(id, ReaderBackward):
		return -1
	}
	return 0
}

// noteControl remembers that a control looked at the mouse, which makes it a node.
func (ui *Context) noteControl(id ID) {
	if !ui.reader.on {
		return
	}
	if ui.reader.controls == nil {
		ui.reader.controls = map[ID]bool{}
	}
	ui.reader.controls[id] = true
}

// noteTooltip remembers the tooltip of the control that looked at the mouse last, which names it
// when it has no words of its own.
func (ui *Context) noteTooltip(text string) {
	if !ui.reader.on || ui.interacted == 0 {
		return
	}
	if ui.reader.tooltips == nil {
		ui.reader.tooltips = map[ID]string{}
	}
	ui.reader.tooltips[ui.interacted] = text
}

// noteLongPressable remembers that the program reads a control's long press.
func (ui *Context) noteLongPressable(id ID) {
	if !ui.reader.on || id == 0 {
		return
	}
	if ui.reader.longPressable == nil {
		ui.reader.longPressable = map[ID]bool{}
	}
	ui.reader.longPressable[id] = true
}

// readerLayer is where the ring around the screen reader's node is painted: over everything.
const readerLayer = popupLayer * 2

// readNodes makes the frame's list of nodes from the laid-out boxes.
func (ui *Context) readNodes() {
	reader := &ui.reader
	reader.nodes = reader.nodes[:0]
	if !reader.on {
		return
	}
	if reader.seen == nil {
		reader.seen = map[ID]int{}
	}
	clear(reader.seen)
	front := ui.frontModalLayer()
	for _, placed := range ui.rootsInLayers() {
		if ui.layerOf(placed.id) < front {
			continue // behind a modal: it cannot be touched, so it is not read
		}
		walk := readerWalk{ui: ui}
		walk.box(placed.element, 0, draw2d.Rect{}, false)
	}
}

// frontModalLayer is the layer of the front-most modal or sheet, or 0 when none is shown.
func (ui *Context) frontModalLayer() int {
	if len(ui.nextSheets) == 0 {
		return 0
	}
	return ui.layerOf(ui.nextSheets[len(ui.nextSheets)-1])
}

// rootsInLayers is the frame's roots from the back to the front, the order they are read in.
func (ui *Context) rootsInLayers() []root {
	sorted := make([]root, len(ui.roots))
	copy(sorted, ui.roots)
	// a few roots: an insertion sort keeps those in one layer in the order they were added
	for index := 1; index < len(sorted); index++ {
		for back := index; back > 0 && ui.layerOf(sorted[back-1].id) > ui.layerOf(sorted[back].id); back-- {
			sorted[back-1], sorted[back] = sorted[back], sorted[back-1]
		}
	}
	return sorted
}

// readerWalk goes down a root's boxes making nodes.
type readerWalk struct {
	ui *Context
}

// box reads a box that is not a control: its controls and scroll areas become nodes, and its own
// texts are read together, as one node where the first of them is.
func (walk readerWalk) box(index int, parent ID, clip draw2d.Rect, clipped bool) {
	ui := walk.ui
	element := &ui.elements[index]
	area, shown := visibleArea(element.rect, clip, clipped)
	description := ui.descriptionOf(element)
	if !shown || description.Hidden {
		return
	}
	if walk.isNode(element, description) {
		walk.control(index, parent, clip, clipped)
		return
	}
	if element.kind == textElement {
		walk.texts(index, parent, area, description)
		return
	}
	clip, clipped = innerClip(element, clip, clipped)
	texts := -1 // the node of this box's own texts, once one is found
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		childElement := &ui.elements[child]
		if childElement.kind != textElement || walk.isNode(childElement, ui.descriptionOf(childElement)) {
			walk.box(child, parent, clip, clipped)
			continue
		}
		childArea, childShown := visibleArea(childElement.rect, clip, clipped)
		childDescription := ui.descriptionOf(childElement)
		if !childShown || childDescription.Hidden || childElement.text == "" {
			continue
		}
		if texts < 0 {
			texts = walk.texts(child, parent, childArea, childDescription)
			continue
		}
		walk.addText(texts, childArea, childElement.text, childDescription)
	}
}

// isNode tells whether a box is read on its own, with what it holds: a control, a scroll area, or
// something given a role or a label.
func (walk readerWalk) isNode(element *element, description Description) bool {
	if element.kind == textElement {
		return false
	}
	return walk.ui.reader.controls[element.id] && element.id != 0 || element.scroll != nil ||
		description.Role != NoRole || description.Label != ""
}

// texts starts a node of text where a text is.
func (walk readerWalk) texts(index int, parent ID, area draw2d.Rect, description Description) int {
	ui := walk.ui
	element := &ui.elements[index]
	node := AccessibleNode{Parent: parent, Area: area, Description: description}
	if node.Role == NoRole {
		node.Role = TextRole
	}
	if node.Label == "" {
		node.Label = element.text
	}
	node.ID = ui.nodeID(hashString(uint64(parent)^fnvOffset, "##text "+node.Label))
	ui.reader.nodes = append(ui.reader.nodes, node)
	return len(ui.reader.nodes) - 1
}

// addText reads another text of a box with those before it.
func (walk readerWalk) addText(position int, area draw2d.Rect, text string, description Description) {
	node := &walk.ui.reader.nodes[position]
	node.Label += " " + text
	node.Area = union(node.Area, area)
	if description.Role == HeadingRole {
		node.Role = HeadingRole
	}
}

// control makes a node of a control, a scroll area or a box described, named by the texts inside it
// that are not in another control.
func (walk readerWalk) control(index int, parent ID, clip draw2d.Rect, clipped bool) {
	ui := walk.ui
	element := &ui.elements[index]
	area, _ := visibleArea(element.rect, clip, clipped)
	description := ui.descriptionOf(element)
	reader := &ui.reader
	node := AccessibleNode{ID: ui.nodeID(element.id), Parent: parent, Area: area, Description: description,
		Pressable: reader.controls[element.id], LongPressable: reader.longPressable[element.id],
		Scrollable: element.scroll != nil || description.Range.Maximum > description.Range.Minimum,
		Typing:     ui.focus != 0 && ui.focus == element.id}
	if element.id == 0 {
		node.ID = ui.nodeID(hashString(uint64(parent)^fnvOffset, "##described "+description.Label))
	}
	switch {
	case node.Role != NoRole:
	case element.scroll != nil:
		node.Role = ScrollRole
	case node.Pressable:
		node.Role = ButtonRole
	default:
		node.Role = TextRole
	}
	position := len(reader.nodes)
	reader.nodes = append(reader.nodes, node)
	clip, clipped = innerClip(element, clip, clipped)
	if element.scroll != nil {
		// a scroll area is read as a list: what it holds are nodes of their own inside it
		for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
			walk.box(child, node.ID, clip, clipped)
		}
		return
	}
	var words, icons []string
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		walk.gather(child, node.ID, clip, clipped, &words, &icons)
	}
	named := &reader.nodes[position]
	if named.Label == "" {
		named.Label = strings.Join(words, " ")
	}
	if named.Label == "" {
		named.Label = reader.tooltips[element.id] // a button with only an icon is called what its tooltip says
	}
	if named.Label == "" {
		named.Label = strings.Join(icons, " ")
	}
}

// gather collects the texts inside a control that name it; a control inside it is a node of its
// own, and its texts are its own.
func (walk readerWalk) gather(index int, parent ID, clip draw2d.Rect, clipped bool, words, icons *[]string) {
	ui := walk.ui
	element := &ui.elements[index]
	_, shown := visibleArea(element.rect, clip, clipped)
	description := ui.descriptionOf(element)
	if !shown || description.Hidden {
		return
	}
	if walk.isNode(element, description) {
		walk.control(index, parent, clip, clipped)
		return
	}
	switch element.kind {
	case textElement:
		if description.Label != "" {
			*words = append(*words, description.Label)
		} else if element.text != "" {
			*words = append(*words, element.text)
		}
		return
	case iconElement:
		// a Lucide name ("arrow-left") is a poor name, but better than none for a control with an
		// icon alone and no label given
		*icons = append(*icons, strings.ReplaceAll(element.icon, "-", " "))
		return
	}
	clip, clipped = innerClip(element, clip, clipped)
	for child := element.firstChild; child >= 0; child = ui.elements[child].nextSibling {
		walk.gather(child, parent, clip, clipped, words, icons)
	}
}

// descriptionOf is what an element was described as, or nothing.
func (ui *Context) descriptionOf(element *element) Description {
	if element.description == 0 || element.description > len(ui.reader.descriptions) {
		return Description{}
	}
	return ui.reader.descriptions[element.description-1]
}

// nodeID makes an id unique in the frame's nodes: the same thing twice (two texts alike in one
// box) gets a different id each time, in order.
func (ui *Context) nodeID(id ID) ID {
	repeats := ui.reader.seen[id]
	ui.reader.seen[id] = repeats + 1
	if repeats > 0 {
		return mix(id, repeats)
	}
	return id
}

// visibleArea is the part of a box a clip leaves, and whether any of it shows.
func visibleArea(rect, clip draw2d.Rect, clipped bool) (draw2d.Rect, bool) {
	if clipped {
		rect = rect.Intersect(clip)
	}
	return rect, !rect.IsEmpty()
}

// innerClip is the clip inside a box: its own area, when it cuts its children to it.
func innerClip(element *element, clip draw2d.Rect, clipped bool) (draw2d.Rect, bool) {
	if !element.look.Clip {
		return clip, clipped
	}
	if clipped {
		return element.rect.Intersect(clip), true
	}
	return element.rect, true
}

// union is the smallest rectangle around two.
func union(first, second draw2d.Rect) draw2d.Rect {
	left, top := min(first.X, second.X), min(first.Y, second.Y)
	right, bottom := max(first.Right(), second.Right()), max(first.Bottom(), second.Bottom())
	return draw2d.Rect{X: left, Y: top, Width: right - left, Height: bottom - top}
}

// scrollForReader moves a scroll area by most of its view when a screen reader asks.
func (ui *Context) scrollForReader(id ID, memory *scrollMemory) {
	page := memory.view * 0.8
	switch {
	case ui.readerActs(id, ReaderForward):
		memory.offset = min(memory.offset+page, max(memory.content-memory.view, 0))
	case ui.readerActs(id, ReaderBackward):
		memory.offset = max(memory.offset-page, 0)
	default:
		return
	}
	memory.glide = 0
}

// paintReaderFocus rings the node the screen reader is on, over everything: dark outside and light
// inside, so it shows on any colour.
func (ui *Context) paintReaderFocus() {
	reader := &ui.reader
	if !reader.on || reader.focus == 0 {
		return
	}
	for _, node := range reader.nodes {
		if node.ID != reader.focus {
			continue
		}
		ui.drawing.SetLayer(readerLayer)
		ring := node.Area
		ui.drawing.Rectangle(ring, 4, draw2d.Transparent, ui.Theme.Ink, 4)
		inner := draw2d.Rect{X: ring.X + 3, Y: ring.Y + 3, Width: ring.Width - 6, Height: ring.Height - 6}
		ui.drawing.Rectangle(inner, 2, draw2d.Transparent, ui.Theme.Background, 2)
		ui.drawing.SetLayer(0)
		return
	}
}
