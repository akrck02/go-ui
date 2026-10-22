package ui

import (
	"errors"
	"math"
	"testing"
	"unicode/utf8"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/opengl/overlay"
)

// fakeMeasurer gives every character half the text size, so widths are easy to work out.
type fakeMeasurer struct{ calls int }

func (measurer *fakeMeasurer) MeasureText(text string, size float32, _ draw2d.Weight, _ float32) draw2d.TextMetrics {
	measurer.calls++
	return draw2d.TextMetrics{Width: float32(utf8.RuneCountInString(text)) * size / 2, Ascent: size * 0.8, Descent: size * 0.2}
}

func (measurer *fakeMeasurer) HasIcon(string) bool { return true }

// harness drives an interface frame by frame with a fake window and mouse.
type harness struct {
	t     *testing.T
	ui    *Context
	state *input.State
}

func newHarness(t *testing.T) *harness {
	state := input.NewState()
	state.WindowWidth, state.WindowHeight = 800, 600
	state.FramebufferWidth, state.FramebufferHeight = 800, 600
	return &harness{t: t, ui: NewWithMeasurer(&fakeMeasurer{}), state: state}
}

// frame runs one update step: the interface, then the end of the step as the backend does it.
func (h *harness) frame(build func(ui *Context)) {
	h.ui.Begin(h.state)
	build(h.ui)
	h.ui.End()
	h.state.Feed().EndFrame()
}

func (h *harness) moveTo(x, y float64) { h.state.Feed().MoveCursor(x, y) }
func (h *harness) press()              { h.state.Feed().SetButton(input.MouseLeft, true) }
func (h *harness) release()            { h.state.Feed().SetButton(input.MouseLeft, false) }

// rect returns where the element with an id was laid out in the last frame.
func (h *harness) rect(id ID) draw2d.Rectangle {
	h.t.Helper()
	area, found := h.ui.previous[id]
	if !found {
		h.t.Fatalf("no element with id %d was laid out", id)
	}
	return area
}

// idOf works out the id a label gets in the screen root (no scope).
func idOf(label string) ID { return hashString(fnvOffset, label) }

func TestGrowingChildrenShareTheFreeSpace(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.frameWithID(idOf("row"), Layout{Width: Fixed(300), Height: Fixed(40), Padding: Pad(10), Gap: 10}, Look{}, func() {
			ui.frameWithID(idOf("fixed"), Layout{Width: Fixed(60)}, Look{}, nil)
			ui.frameWithID(idOf("one"), Layout{Width: Grow()}, Look{}, nil)
			ui.frameWithID(idOf("two"), Layout{Width: GrowBy(2)}, Look{}, nil)
		})
	})
	// inner 280, minus 60 fixed and two gaps of 10: 200 left, shared 1:2
	if one, two := h.rect(idOf("one")), h.rect(idOf("two")); one.Width < 66 || one.Width > 67 || two.Width < 133 || two.Width > 134 {
		t.Fatalf("got %v and %v", one.Width, two.Width)
	}
	if one := h.rect(idOf("one")); one.X != 10+60+10 || one.Y != 10 || one.Height != 0 {
		t.Fatalf("placed after the fixed box, inside the padding: %+v", one)
	}
}

func TestAGrowingChildStopsAtItsMaximum(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.Box(Layout{Width: Fixed(300)}, func() {
			ui.frameWithID(idOf("capped"), Layout{Width: Grow().Between(0, 50)}, Look{}, nil)
			ui.frameWithID(idOf("free"), Layout{Width: Grow()}, Look{}, nil)
		})
	})
	if capped, free := h.rect(idOf("capped")), h.rect(idOf("free")); capped.Width != 50 || free.Width != 250 {
		t.Fatalf("the capped box keeps 50, the other takes the rest: %v, %v", capped.Width, free.Width)
	}
}

func TestColumnsFitTheWidestChildAndCentre(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.frameWithID(idOf("column"), Layout{Direction: Vertical, Padding: Pad(5), Gap: 4, AlignX: Center}, Look{}, func() {
			ui.frameWithID(idOf("wide"), Layout{Width: Fixed(100), Height: Fixed(20)}, Look{}, nil)
			ui.frameWithID(idOf("narrow"), Layout{Width: Fixed(40), Height: Fixed(20)}, Look{}, nil)
			ui.frameWithID(idOf("filling"), Layout{Width: Grow(), Height: Fixed(10)}, Look{}, nil)
		})
	})
	column := h.rect(idOf("column"))
	if column.Width != 110 || column.Height != 5+20+4+20+4+10+5 {
		t.Fatalf("the column fits its content: %+v", column)
	}
	if narrow := h.rect(idOf("narrow")); narrow.X != 5+30 {
		t.Fatalf("the narrow child is centred: %+v", narrow)
	}
	if filling := h.rect(idOf("filling")); filling.Width != 100 || filling.X != 5 {
		t.Fatalf("a growing child fills the column's width: %+v", filling)
	}
}

func TestPercentages(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.Box(Layout{Width: Fixed(200), Height: Fixed(100)}, func() {
			ui.frameWithID(idOf("quarter"), Layout{Width: Percent(0.25), Height: Percent(0.5)}, Look{}, nil)
		})
	})
	if quarter := h.rect(idOf("quarter")); quarter.Width != 50 || quarter.Height != 50 {
		t.Fatalf("got %+v", quarter)
	}
}

func TestTextIsMeasuredAndCentred(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.frameWithID(idOf("box"), Layout{Height: Fixed(40), AlignY: Center}, Look{}, func() { ui.Label("Lighthouse") })
	})
	if box := h.rect(idOf("box")); box.Width != 10*13/2 {
		t.Fatalf("the box fits the text: %+v", box)
	}
	text := h.ui.drawing.Commands[0]
	if text.Kind != draw2d.TextKind || math.Abs(float64(text.Baseline)-(13.5+10.4)) > 0.001 {
		t.Fatalf("the line is centred in the box: %+v", text)
	}
}

func TestPanelsClaimTheMouse(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.PanelWith("Light", PanelOptions{X: 100, Y: 100}, func() { ui.Label("Range") })
	}
	h.frame(build)
	h.moveTo(110, 110)
	h.frame(build)
	if !h.ui.WantsMouse() {
		t.Fatal("the pointer over a panel belongs to the interface")
	}
	h.ui.Begin(h.state)
	build(h.ui)
	h.ui.End()
	if !h.state.MouseClaimed() {
		t.Fatal("the claim reaches the input state, for the camera")
	}
	h.state.Feed().EndFrame()
	h.moveTo(700, 500)
	h.frame(build)
	if h.ui.WantsMouse() {
		t.Fatal("off the panel the mouse is the scene's")
	}
}

func TestADragStartedOnTheSceneStaysWithTheScene(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.PanelWith("Light", PanelOptions{X: 100, Y: 100}, func() { ui.Label("Range") })
	}
	h.frame(build)
	h.moveTo(700, 500)
	h.press()
	h.frame(build)
	h.moveTo(110, 110) // the camera drag passes over the panel
	h.frame(build)
	if h.ui.WantsMouse() {
		t.Fatal("a drag that began on the scene is not taken by a panel it crosses")
	}
	h.release()
	h.frame(build)
	h.frame(build)
	if !h.ui.WantsMouse() {
		t.Fatal("once let go, the panel takes the mouse again")
	}
}

func TestDraggingTheTitleMovesThePanel(t *testing.T) {
	h := newHarness(t)
	panelID := idOf("Light")
	build := func(ui *Context) {
		ui.PanelWith("Light", PanelOptions{X: 100, Y: 100}, func() { ui.Label("Range") })
	}
	h.frame(build)
	h.moveTo(110, 110)
	h.frame(build)
	h.press()
	h.frame(build)
	h.moveTo(160, 140)
	h.frame(build)
	h.frame(build) // the move shows in the layout of the following frame
	if moved := h.rect(panelID); moved.X != 150 || moved.Y != 130 {
		t.Fatalf("the panel follows the pointer: %+v", moved)
	}
	h.release()
	h.frame(build)
}

func TestTheFoldButtonCollapsesThePanel(t *testing.T) {
	h := newHarness(t)
	panelID := idOf("Light")
	build := func(ui *Context) {
		ui.PanelWith("Light", PanelOptions{X: 100, Y: 100}, func() { ui.Label("Range"); ui.Label("Colour") })
	}
	h.frame(build)
	open := h.rect(panelID)
	fold := h.rect(hashString(uint64(idOf("Light")), "##fold"))
	h.moveTo(float64(fold.X+3), float64(fold.Y+3))
	h.frame(build)
	h.press()
	h.release()
	h.frame(build)
	h.frame(build)
	if folded := h.rect(panelID); folded.Height >= open.Height {
		t.Fatalf("a folded panel shows only its title bar: %v then %v", open.Height, folded.Height)
	}
}

func TestPressedPanelComesToTheFront(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.PanelWith("Back", PanelOptions{X: 100, Y: 100}, func() { ui.Label("first") })
		ui.PanelWith("Front", PanelOptions{X: 150, Y: 110}, func() { ui.Label("second") })
	}
	h.frame(build)
	if h.ui.layerOf(idOf("Front")) <= h.ui.layerOf(idOf("Back")) {
		t.Fatal("a panel added later starts in front")
	}
	h.moveTo(105, 105) // only the back panel is under the pointer here
	h.frame(build)
	h.press()
	h.frame(build)
	if h.ui.layerOf(idOf("Back")) <= h.ui.layerOf(idOf("Front")) {
		t.Fatal("pressing the back panel brings it to the front")
	}
	h.release()
	h.frame(build)
}

func TestRepeatedLabelsGetTheirOwnIDs(t *testing.T) {
	h := newHarness(t)
	h.ui.Begin(h.state)
	first, second := h.ui.idFor("Apply"), h.ui.idFor("Apply")
	h.ui.Push("north")
	scoped := h.ui.idFor("Apply")
	h.ui.Pop()
	h.ui.End()
	if first == second || first == scoped || second == scoped {
		t.Fatal("repeats and scopes give different ids")
	}
	h.ui.Begin(h.state)
	again := h.ui.idFor("Apply")
	h.ui.End()
	if again != first {
		t.Fatal("ids stay the same from frame to frame")
	}
	if displayText("Apply##north") != "Apply" || displayText("Apply") != "Apply" {
		t.Fatal("what follows ## is not shown")
	}
}

func TestTextMeasuresAreCached(t *testing.T) {
	h := newHarness(t)
	measurer := h.ui.measurer.(*fakeMeasurer)
	for range 3 {
		h.frame(func(ui *Context) { ui.Label("Lighthouse") })
	}
	if measurer.calls != 1 {
		t.Fatalf("the same text is measured once, got %d", measurer.calls)
	}
}

func TestPanelsAreDrawnOnTheirLayersAndClipped(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.Label("on the screen")
		ui.PanelWith("Light", PanelOptions{X: 100, Y: 100}, func() { ui.Label("Range") })
	})
	var screenText, panelText *draw2d.Command
	for index := range h.ui.drawing.Commands {
		command := &h.ui.drawing.Commands[index]
		if command.Kind != draw2d.TextKind {
			continue
		}
		switch command.Text {
		case "on the screen":
			screenText = command
		case "Range":
			panelText = command
		}
	}
	if screenText == nil || panelText == nil {
		t.Fatal("both texts are drawn")
	}
	if screenText.Layer != 0 || panelText.Layer <= 0 {
		t.Fatalf("the screen is under the panels: %d, %d", screenText.Layer, panelText.Layer)
	}
	if !panelText.HasClip || screenText.HasClip {
		t.Fatal("a panel's content is cut to the panel")
	}
}

func TestDockedPanelsSpanTheirEdge(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.PanelWith("Tools", PanelOptions{Dock: DockRight, Width: Fixed(240)}, func() { ui.Label("Select") })
	})
	if docked := h.rect(idOf("Tools")); docked != (draw2d.Rectangle{X: 560, Y: 0, Width: 240, Height: 600}) {
		t.Fatalf("got %+v", docked)
	}
}

func TestTheInnermostControlTakesThePointer(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.PanelWith("Light", PanelOptions{X: 100, Y: 100}, func() { ui.Label("Range") })
	}
	h.frame(build)
	fold := h.rect(hashString(uint64(idOf("Light")), "##fold"))
	h.moveTo(float64(fold.X+3), float64(fold.Y+3))
	h.frame(build)
	if h.ui.hot != hashString(uint64(idOf("Light")), "##fold") || h.ui.cursor != input.CursorHand {
		t.Fatal("over the fold button, the button is hot, not the title bar around it")
	}
}

func TestPaintedScreenBoxesHoldTheMouse(t *testing.T) {
	h := newHarness(t)
	build := func(ui *Context) {
		ui.Box(Layout{Width: Grow(), Height: Grow()}, func() {
			ui.Frame(Layout{Width: Fixed(200), Height: Grow()}, Look{Fill: ui.Theme.Surface}, func() { ui.Label("side") })
			ui.Box(Layout{Width: Grow(), Height: Grow()}, nil) // the view, left transparent
		})
	}
	h.frame(build)
	h.moveTo(100, 300)
	h.frame(build)
	if !h.ui.WantsMouse() {
		t.Fatal("the side column is the interface's")
	}
	h.moveTo(500, 300)
	h.frame(build)
	if h.ui.WantsMouse() {
		t.Fatal("the transparent middle is the scene's")
	}
	h.moveTo(100, 300)
	h.frame(build)
	h.press()
	h.frame(build)
	if h.ui.sceneDrag || !h.state.MouseClaimed() && !h.ui.WantsMouse() {
		t.Fatal("a press on the side column does not start a camera drag")
	}
	h.release()
	h.frame(build)
}

func TestSelectableRowsReportClicks(t *testing.T) {
	h := newHarness(t)
	selected := "North Point"
	build := func(ui *Context) {
		for _, name := range []string{"North Point", "South Cape"} {
			if ui.Selectable("lamp", name, name == selected) {
				selected = name
			}
		}
	}
	h.frame(build)
	row := h.rect(idOf("South Cape"))
	h.moveTo(float64(row.X+4), float64(row.Y+4))
	h.frame(build)
	h.press()
	h.release()
	h.frame(build)
	if selected != "South Cape" {
		t.Fatalf("clicking a row selects it, got %q", selected)
	}
}

func TestClickableBoxesFollowThePointer(t *testing.T) {
	h := newHarness(t)
	clicked := 0
	looks := Looks{Normal: Look{Fill: draw2d.MustHex("#ffffff")}, Hovered: Look{Fill: draw2d.MustHex("#eeeeee")}, Pressed: Look{Fill: draw2d.MustHex("#dddddd")}}
	build := func(ui *Context) {
		if ui.Clickable("row", Layout{Width: Fixed(200), Height: Fixed(30)}, looks, func() { ui.Label("North Point") }) {
			clicked++
		}
	}
	h.frame(build)
	h.moveTo(10, 10)
	h.frame(build)
	if h.ui.drawing.Commands[0].Fill != looks.Hovered.Fill {
		t.Fatal("under the pointer the box takes its hovered look")
	}
	h.press()
	h.release()
	h.frame(build)
	if clicked != 1 {
		t.Fatalf("got %d clicks", clicked)
	}
}

func TestSharesIgnoreTheContent(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.Box(Layout{Width: Fixed(400)}, func() {
			ui.frameWithID(idOf("short"), Layout{Width: Share(1)}, Look{}, func() { ui.Label("ab") })
			ui.frameWithID(idOf("long"), Layout{Width: Share(1)}, Look{}, func() { ui.Label("a much longer text") })
		})
	})
	if short, long := h.rect(idOf("short")), h.rect(idOf("long")); short.Width != 200 || long.Width != 200 {
		t.Fatalf("equal shares are equal whatever they hold: %v, %v", short.Width, long.Width)
	}
}

func TestGrowingChildrenShrinkWhenTheContentDoesNotFit(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.frameWithID(idOf("row"), Layout{Width: Fixed(300)}, Look{}, func() {
			ui.frameWithID(idOf("side"), Layout{Width: Fixed(100)}, Look{}, nil)
			ui.frameWithID(idOf("main"), Layout{Width: Grow()}, Look{}, func() {
				ui.Box(Layout{Width: Fixed(300)}, nil)
			})
			ui.frameWithID(idOf("bounded"), Layout{Width: Grow().Between(80, 0)}, Look{}, func() {
				ui.Box(Layout{Width: Fixed(100)}, nil)
			})
		})
	})
	// 500 wanted in 300: main gives 300 of slack and bounded 20, in proportion
	side, main, bounded := h.rect(idOf("side")), h.rect(idOf("main")), h.rect(idOf("bounded"))
	if side.Width != 100 || math.Abs(float64(main.Width+bounded.Width-200)) > 0.01 || bounded.Width < 80 || bounded.X+bounded.Width > 300.01 {
		t.Fatalf("side %v, main %v, bounded %v at %v", side.Width, main.Width, bounded.Width, bounded.X)
	}
}

func TestTextInAClippingBoxEndsInAnEllipsis(t *testing.T) {
	h := newHarness(t)
	// 6.5 pixels a character: "Rocky island" is 78 wide; 50 holds "Rock…" (32.5) and "Rocky…" (39), not "Rocky …"
	if got := h.ui.elide("Rocky island", TextStyle{Size: 13}, 50); got != "Rocky…" {
		t.Fatalf("got %q", got)
	}
	if got := h.ui.elide("Rocky", TextStyle{Size: 13}, 4); got != "" {
		t.Fatalf("no room: %q", got)
	}
}

func TestAnInterfaceInFontsThatAreNotFontsIsFontNotRead(test *testing.T) {
	if _, err := NewInFonts(overlay.Fonts{Regular: []byte("not a font")}); !errors.Is(err, overlay.FontNotRead) {
		test.Fatalf("got %v, want %s", err, overlay.FontNotRead.Name())
	}
	if _, err := NewInFonts(overlay.Fonts{}); err != nil {
		test.Fatalf("the default fonts: %v", err)
	}
}
