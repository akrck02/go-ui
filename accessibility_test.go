package ui

import "testing"

// readerHarness is a harness with a screen reader on.
func readerHarness(t *testing.T) *harness {
	h := newHarness(t)
	h.ui.SetScreenReader(true)
	return h
}

// node finds the node read with a label, or fails.
func (h *harness) node(label string) AccessibleNode {
	h.t.Helper()
	for _, node := range h.ui.ScreenReaderNodes() {
		if node.Label == label {
			return node
		}
	}
	h.t.Fatalf("no node reads %q in %+v", label, h.labels())
	return AccessibleNode{}
}

// labels is what the screen reader would read, in order.
func (h *harness) labels() []string {
	var labels []string
	for _, node := range h.ui.ScreenReaderNodes() {
		labels = append(labels, node.Label)
	}
	return labels
}

func TestNothingIsReadWithNoScreenReader(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) { ui.Button("Light the lamp") })
	if nodes := h.ui.ScreenReaderNodes(); len(nodes) != 0 {
		t.Fatalf("no screen reader, no nodes: %+v", nodes)
	}
}

func TestAControlIsReadByItsTextsAndTextsBesideEachOtherTogether(t *testing.T) {
	h := readerHarness(t)
	h.frame(func(ui *Context) {
		ui.Column(func() {
			ui.Label("North lighthouse")
			ui.Caption("On the cliff")
		})
		ui.Button("Light the lamp")
	})
	if labels := h.labels(); len(labels) != 2 {
		t.Fatalf("one node for the two texts of a column, one for the button: %q", labels)
	}
	texts := h.node("North lighthouse On the cliff")
	if texts.Role != TextRole || texts.Pressable {
		t.Fatalf("texts are read, not pressed: %+v", texts)
	}
	button := h.node("Light the lamp")
	if button.Role != ButtonRole || !button.Pressable || button.ID != idOf("Light the lamp") {
		t.Fatalf("a button is a button, named by its id: %+v", button)
	}
	if button.Area != h.rect(idOf("Light the lamp")) {
		t.Fatalf("read where it shows: %+v, laid out at %+v", button.Area, h.rect(idOf("Light the lamp")))
	}
}

func TestControlsSayWhatTheyAreAndTheirState(t *testing.T) {
	h := readerHarness(t)
	lamp, fog, tab := true, false, 1
	value := float32(4)
	h.frame(func(ui *Context) {
		ui.Toggle("Lamp", &lamp)
		ui.Checkbox("Fog horn", &fog)
		ui.Tabs("views", &tab, "Map", "List")
		ui.SliderStep("Beam", &value, 0, 10, 1)
		ui.Heading("Keepers")
	})
	if node := h.node("Lamp"); node.Role != SwitchRole || !node.Checked {
		t.Fatalf("a switch, on: %+v", node)
	}
	if node := h.node("Fog horn"); node.Role != CheckboxRole || node.Checked {
		t.Fatalf("a checkbox, not ticked: %+v", node)
	}
	if node := h.node("List"); node.Role != TabRole || !node.Selected {
		t.Fatalf("the tab shown: %+v", node)
	}
	if node := h.node("Map"); node.Selected {
		t.Fatalf("the other tab is not selected: %+v", node)
	}
	slider := h.node("Beam")
	if slider.Role != SliderRole || !slider.Scrollable || slider.Range != (Range{Maximum: 10, Current: 4}) {
		t.Fatalf("a slider with its range: %+v", slider)
	}
	if node := h.node("Keepers"); node.Role != HeadingRole {
		t.Fatalf("a heading: %+v", node)
	}
}

func TestDescribeNamesAControlOfTheProgramsOwn(t *testing.T) {
	h := readerHarness(t)
	on := true
	h.frame(func(ui *Context) {
		ui.Clickable("beacon", Layout{Width: Fixed(52), Height: Fixed(32)}, Looks{}, nil)
		ui.Describe(Description{Role: SwitchRole, Label: "Beacon", Checked: on})
		ui.Clickable("back", Layout{Width: Fixed(40), Height: Fixed(40)}, Looks{}, func() { ui.Icon("arrow-left") })
	})
	if node := h.node("Beacon"); node.Role != SwitchRole || !node.Checked || !node.Pressable {
		t.Fatalf("described as a switch: %+v", node)
	}
	if node := h.node("arrow left"); node.Role != ButtonRole {
		t.Fatalf("a control with an icon alone is named by it, for want of a label: %+v", node)
	}
}

func TestFieldsAreReadWithTheirTextUnlessSecret(t *testing.T) {
	h := readerHarness(t)
	name, password := "Keeper", "tide"
	h.frame(func(ui *Context) {
		ui.TextFieldWith("name", &name, TextFieldOptions{Placeholder: "Name"})
		ui.TextFieldWith("password", &password, TextFieldOptions{Placeholder: "Password", Secret: true})
	})
	if node := h.node("Name"); node.Role != FieldRole || node.Value != "Keeper" {
		t.Fatalf("a field with its text: %+v", node)
	}
	if node := h.node("Password"); !node.Secret || node.Value != "" {
		t.Fatalf("a secret field keeps its text: %+v", node)
	}
}

func TestAPressFromTheScreenReaderClicksTheControl(t *testing.T) {
	h := readerHarness(t)
	clicks, longPresses := 0, 0
	build := func(ui *Context) {
		if ui.Button("Light the lamp") {
			clicks++
		}
		if ui.LongPressed() {
			longPresses++
		}
		if ui.Button("Ring the bell") {
			clicks += 10
		}
	}
	h.frame(build)
	if !h.node("Light the lamp").LongPressable || h.node("Ring the bell").LongPressable {
		t.Fatalf("only the button the program reads a long press of says it has one: %+v", h.ui.ScreenReaderNodes())
	}
	h.ui.ScreenReaderAct(h.node("Light the lamp").ID, ReaderPress)
	h.frame(build)
	if clicks != 1 {
		t.Fatalf("the press clicks that button only, with no pointer on it: %d", clicks)
	}
	h.frame(build)
	if clicks != 1 {
		t.Fatalf("once: %d", clicks)
	}
	h.ui.ScreenReaderAct(h.node("Light the lamp").ID, ReaderLongPress)
	h.frame(build)
	if longPresses != 1 || clicks != 1 {
		t.Fatalf("a long press and no click: %d long presses, %d clicks", longPresses, clicks)
	}
}

func TestAPressFocusesAField(t *testing.T) {
	h := readerHarness(t)
	name := "Keeper"
	build := func(ui *Context) { ui.TextFieldWith("name", &name, TextFieldOptions{Placeholder: "Name"}) }
	h.frame(build)
	h.moveTo(1, 1) // the pointer at the field's start, where a tap would put the caret
	h.ui.ScreenReaderAct(h.node("Name").ID, ReaderPress)
	h.frame(build)
	h.frame(build)
	if !h.ui.Typing() || !h.node("Name").Typing {
		t.Fatal("pressed by the screen reader, the field takes the keyboard")
	}
	h.state.AddCharacter('!')
	h.frame(build)
	if name != "Keeper!" {
		t.Fatalf("typing goes on at the end of the text: %q", name)
	}
}

func TestAScrollAreaHoldsItsRowsAndScrollsForTheReader(t *testing.T) {
	h := readerHarness(t)
	build := func(ui *Context) {
		ui.ScrollArea("lighthouses", Layout{Width: Fixed(200), Height: Fixed(100)}, func() {
			for row := range 10 {
				ui.Push(string(rune('a' + row)))
				ui.frameWithID(0, Layout{Width: Grow(), Height: Fixed(40)}, Look{}, func() {
					ui.Label("Lighthouse " + string(rune('A'+row)))
				})
				ui.Pop()
			}
		})
	}
	h.frame(build)
	h.frame(build) // the area knows its content's height from the last frame
	area := h.node("")
	if area.Role != ScrollRole || !area.Scrollable || area.ID != idOf("lighthouses") {
		t.Fatalf("a scroll area: %+v", area)
	}
	first := h.node("Lighthouse A")
	if first.Parent != area.ID {
		t.Fatalf("its rows are inside it: %+v", first)
	}
	for _, label := range h.labels() {
		if label == "Lighthouse D" {
			t.Fatalf("a row out of sight is not read: %q", h.labels())
		}
	}
	h.ui.ScreenReaderAct(area.ID, ReaderForward)
	h.frame(build)
	h.frame(build)
	if node := h.node("Lighthouse D"); node.Area.Y >= 100 || node.Area.IsEmpty() {
		t.Fatalf("scrolled forward, the next rows show: %+v", node.Area)
	}
}

func TestTheSliderMovesAStepForTheReader(t *testing.T) {
	h := readerHarness(t)
	value := float32(4)
	changed := false
	build := func(ui *Context) { changed = ui.SliderStep("Beam", &value, 0, 10, 1) }
	h.frame(build)
	h.ui.ScreenReaderAct(idOf("Beam"), ReaderForward)
	h.frame(build)
	if value != 5 || !changed {
		t.Fatalf("a step up: %v, changed %v", value, changed)
	}
	h.ui.ScreenReaderAct(idOf("Beam"), ReaderBackward)
	h.frame(build)
	h.ui.ScreenReaderAct(idOf("Beam"), ReaderBackward)
	h.frame(build)
	if value != 3 {
		t.Fatalf("two steps down: %v", value)
	}
}

func TestWhatIsBehindAModalIsNotRead(t *testing.T) {
	h := readerHarness(t)
	h.frame(func(ui *Context) {
		ui.Button("Light the lamp")
		ui.Modal("confirm", LayerOptions{AlignX: Center, AlignY: Center}, func() {
			ui.Label("Put the lamp out?")
			ui.Button("Put it out")
		})
	})
	labels := h.labels()
	if len(labels) != 2 || labels[0] != "Put the lamp out?" || labels[1] != "Put it out" {
		t.Fatalf("only the modal, without its backdrop: %q", labels)
	}
}

func TestNodesKeepTheirIDsAndRepeatsGetTheirOwn(t *testing.T) {
	h := readerHarness(t)
	build := func(ui *Context) {
		ui.Column(func() { ui.Label("Lighthouse") })
		ui.Column(func() { ui.Label("Lighthouse") })
	}
	h.frame(build)
	first := h.ui.ScreenReaderNodes()
	firstIDs := []ID{first[0].ID, first[1].ID}
	if firstIDs[0] == firstIDs[1] {
		t.Fatalf("two texts alike are two nodes: %+v", first)
	}
	h.frame(build)
	again := h.ui.ScreenReaderNodes()
	if again[0].ID != firstIDs[0] || again[1].ID != firstIDs[1] {
		t.Fatalf("the same ids in the next frame: %v, then %v", firstIDs, []ID{again[0].ID, again[1].ID})
	}
}

func TestTheScreenReadersNodeIsRinged(t *testing.T) {
	h := readerHarness(t)
	build := func(ui *Context) { ui.Button("Light the lamp") }
	h.frame(build)
	before := len(h.ui.Drawing().Commands)
	h.ui.SetScreenReaderFocus(idOf("Light the lamp"))
	h.frame(build)
	after := h.ui.Drawing().Commands
	if len(after) != before+2 {
		t.Fatalf("two rings, dark and light: %d commands, then %d", before, len(after))
	}
	ring := after[len(after)-2]
	if ring.Layer != readerLayer || ring.Area != h.node("Light the lamp").Area {
		t.Fatalf("over everything, around the button: %+v", ring)
	}
	h.ui.SetScreenReader(false)
	h.frame(build)
	if len(h.ui.Drawing().Commands) != before {
		t.Fatal("no screen reader, no ring")
	}
}

func TestDescribeAfterAnOpenMenuNamesItsButton(t *testing.T) {
	h := readerHarness(t)
	build := func(ui *Context) {
		ui.MenuButton("ellipsis-vertical", "", func() { ui.MenuItem("", "Tides", "") })
		ui.Describe(Description{Label: "More options"})
	}
	h.frame(build)
	h.ui.ScreenReaderAct(h.node("More options").ID, ReaderPress)
	h.frame(build)
	h.node("Tides")
	if node := h.node("More options"); node.Role != ButtonRole {
		t.Fatalf("the button keeps its name while its menu is open: %+v", node)
	}
}
