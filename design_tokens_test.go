package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

// near says whether two layout sizes are the same, allowing for rounding.
func near(first, second float32) bool { return first-second < 0.01 && second-first < 0.01 }

// rectanglesFilled lists the rectangles of the last frame painted with a colour.
func rectanglesFilled(h *harness, fill draw2d.Color) []draw2d.Command {
	var found []draw2d.Command
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.RectangleKind && command.Fill == fill {
			found = append(found, command)
		}
	}
	return found
}

func TestAThemeThatOnlyChangesThePrimaryColourStillChoosesWithIt(t *testing.T) {
	h := newHarness(t)
	h.ui.Theme.Primary, h.ui.Theme.PrimaryTint = draw2d.MustHex("#336699"), draw2d.MustHex("#ddeeff")
	h.frame(func(ui *Context) { ui.Selectable("lamp", "North Point", true) })
	if len(rectanglesFilled(h, draw2d.MustHex("#ddeeff"))) != 1 {
		t.Fatal("a theme that leaves Chosen alone marks the chosen item with its own PrimaryTint")
	}
}

func TestChosenItemsTakeTheThemesChosenColours(t *testing.T) {
	h := newHarness(t)
	h.ui.Theme.Chosen, h.ui.Theme.ChosenInk, h.ui.Theme.ChosenWeight = draw2d.MustHex("#eeeeee"), draw2d.MustHex("#111111"), draw2d.Bold
	h.frame(func(ui *Context) { ui.Selectable("lamp", "North Point", true) })
	if len(rectanglesFilled(h, draw2d.MustHex("#eeeeee"))) != 1 {
		t.Fatal("the chosen item is filled with Chosen")
	}
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.TextKind && (command.Fill != draw2d.MustHex("#111111") || command.Weight != draw2d.Bold) {
			t.Fatalf("the chosen label is written in ChosenInk and ChosenWeight, got %v weight %d", command.Fill, command.Weight)
		}
	}
}

func TestTheLightThemeKeepsGoUIsOwnButtons(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) { ui.Button("Save") })
	button := h.rect(idOf("Save"))
	// "Save" is 4 letters of half the text size (13): 26 wide, with 12 on each side
	if button.Width != 26+2*defaultButtonPadding || button.Height != h.ui.Theme.ControlHeight {
		t.Fatalf("a button keeps go-ui's padding and height, got %v", button)
	}
	for _, command := range rectanglesFilled(h, h.ui.Theme.Surface) {
		if command.Radius != h.ui.Theme.ControlRadius {
			t.Fatalf("a button keeps the controls' corners, got %v", command.Radius)
		}
	}
}

func TestButtonsTakeTheThemesRoomAndCorners(t *testing.T) {
	h := newHarness(t)
	h.ui.Theme.ButtonPadding, h.ui.Theme.ButtonRadius = 18, 10
	h.frame(func(ui *Context) { ui.Button("Save") })
	if width := h.rect(idOf("Save")).Width; width != 26+2*18 {
		t.Fatalf("ButtonPadding sets the room on each side, got a width of %v", width)
	}
	if rectangles := rectanglesFilled(h, h.ui.Theme.Surface); len(rectangles) != 1 || rectangles[0].Radius != 10 {
		t.Fatal("ButtonRadius sets a button's corners")
	}
}

func TestTheMainButtonStandsOnItsEdgeWithoutTakingRoom(t *testing.T) {
	h := newHarness(t)
	edge := draw2d.MustHex("#224466")
	h.ui.Theme.PrimaryEdge, h.ui.Theme.EdgeWidth = edge, 2
	h.frame(func(ui *Context) { ui.PrimaryButton("Save") })
	button := h.rect(idOf("Save"))
	under := rectanglesFilled(h, edge)
	if len(under) != 1 {
		t.Fatal("the main button paints its edge under it")
	}
	if under[0].Area.Y != button.Y+2 || under[0].Area.Height != button.Height {
		t.Fatalf("the edge is the button moved down by EdgeWidth, got %v under %v", under[0].Area, button)
	}
	if button.Height != h.ui.Theme.ControlHeight {
		t.Fatal("the edge takes no room in the layout")
	}
}

func TestWithoutAnEdgeWidthNoEdgeIsDrawn(t *testing.T) {
	h := newHarness(t)
	h.ui.Theme.PrimaryEdge = draw2d.MustHex("#224466")
	h.frame(func(ui *Context) { ui.PrimaryButton("Save") })
	if len(rectanglesFilled(h, draw2d.MustHex("#224466"))) != 0 {
		t.Fatal("a zero EdgeWidth draws no edge, as go-ui always did")
	}
}

func TestALineOfTextIsAsTallAsTheThemesLineHeight(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) { ui.Box(Layout{Direction: Vertical}, func() { ui.Label("North Point") }) })
	letters := h.ui.elements[2].rect.Height // 0 is the screen, 1 the box, 2 its first line
	h.ui.Theme.LineHeight, h.ui.Theme.HeadingLineHeight = 1.5, 1.2
	h.frame(func(ui *Context) {
		ui.Box(Layout{Direction: Vertical}, func() {
			ui.Label("North Point")
			ui.Heading("The island")
		})
	})
	body, heading := h.ui.elements[2].rect.Height, h.ui.elements[3].rect.Height
	if !near(letters, h.ui.Theme.TextSize) {
		t.Fatalf("without a line height a line is as tall as its letters, got %v", letters)
	}
	if !near(body, h.ui.Theme.TextSize*1.5) || !near(heading, h.ui.Theme.HeadingSize*1.2) {
		t.Fatalf("a line is its size × LineHeight, a heading × HeadingLineHeight, got %v and %v", body, heading)
	}
}

func TestATextsOwnLineHeightWinsOverTheThemes(t *testing.T) {
	h := newHarness(t)
	h.ui.Theme.LineHeight = 1.5
	h.frame(func(ui *Context) {
		ui.Box(Layout{Direction: Vertical}, func() { ui.Text("Harbour", TextStyle{Size: 30, LineHeight: 1.2}) })
	})
	if line := h.ui.elements[2].rect.Height; !near(line, 30*1.2) {
		t.Fatalf("a title with its own line height is its size × that, got %v", line)
	}
}

func TestATextCanKeepTheHeightOfItsLettersWhateverTheTheme(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.Box(Layout{Direction: Vertical}, func() { ui.Text("Harbour", TextStyle{Size: 30}) })
	})
	letters := h.ui.elements[2].rect.Height
	h.ui.Theme.LineHeight = 1.5
	h.frame(func(ui *Context) {
		ui.Box(Layout{Direction: Vertical}, func() { ui.Text("Harbour", TextStyle{Size: 30, LineHeight: LettersLineHeight}) })
	})
	if line := h.ui.elements[2].rect.Height; !near(line, letters) {
		t.Fatalf("a text asking for its letters' height is %v tall, not %v", line, letters)
	}
}

func TestACheckBoxTakesTheThemesSizeCornersAndEdges(t *testing.T) {
	h := newHarness(t)
	theme := &h.ui.Theme
	theme.CheckSize, theme.CheckRadius, theme.CheckBorder, theme.CheckedBorder = 22, 6, 2, draw2d.MustHex("#224466")
	done := true
	h.frame(func(ui *Context) { ui.Checkbox("Done", &done) })
	for _, command := range rectanglesFilled(h, theme.Primary) {
		if command.Area.Width != 22 || command.Radius != 6 || command.Border != draw2d.MustHex("#224466") || command.BorderWidth != 2 {
			t.Fatalf("a ticked box is CheckSize wide with CheckRadius corners and the CheckedBorder edge, got %+v", command)
		}
		return
	}
	t.Fatal("a ticked box is filled with the primary colour")
}
