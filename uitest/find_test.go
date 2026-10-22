package uitest

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// keeperLog is a small interface for the tests: a lighthouse keeper's log.
func keeperLog(context *ui.Context) {
	context.Column(func() {
		context.Heading("Lighthouse log")
		context.Button("Light the lamp")
		context.Label("Wind: calm")
	})
}

func newWindow(t *testing.T, build func(*ui.Context), options ...Option) *Window {
	t.Helper()
	window, err := New(build, options...)
	if err != nil {
		t.Fatal(err)
	}
	window.Frame()
	return window
}

func TestAnElementIsFoundByItsLabel(t *testing.T) {
	window := newWindow(t, keeperLog)
	element, err := window.Element("Light the lamp")
	if err != nil {
		t.Fatal(err)
	}
	if !element.Visible || !element.Pressable || element.Rect.Width <= 0 || element.Rect.Height <= 0 {
		t.Fatalf("the button was found as %+v", element)
	}
}

func TestAMissingLabelGivesElementNotFound(t *testing.T) {
	window := newWindow(t, keeperLog)
	_, err := window.Element("Ring the bell")
	if !errors.Is(err, ErrorElementNotFound) {
		t.Fatalf("got %v, want %s", err, ElementNotFound)
	}
	if !strings.Contains(err.Error(), "Light the lamp") {
		t.Fatalf("the message does not list what is on screen: %v", err)
	}
}

func TestARepeatedLabelGivesElementAmbiguous(t *testing.T) {
	window := newWindow(t, func(context *ui.Context) {
		context.Button("Save")
		context.Button("Save")
	})
	_, err := window.Element("Save")
	if !errors.Is(err, ErrorElementAmbiguous) {
		t.Fatalf("got %v, want %s", err, ElementAmbiguous)
	}
}

func TestAnElementIsFoundInItsScope(t *testing.T) {
	window := newWindow(t, func(context *ui.Context) {
		context.Push("north")
		context.Button("Save")
		context.Pop()
		context.Push("south")
		context.Button("Save")
		context.Pop()
	})
	north, err := window.ElementIn("north", "Save")
	if err != nil {
		t.Fatal(err)
	}
	south, err := window.ElementIn("south", "Save")
	if err != nil {
		t.Fatal(err)
	}
	if north.Rect == south.Rect || !north.Visible || !south.Visible {
		t.Fatalf("the two buttons were not told apart: %+v and %+v", north, south)
	}
}

func TestAnElementScrolledAwayIsNotVisible(t *testing.T) {
	window := newWindow(t, func(context *ui.Context) {
		context.ScrollArea("entries", ui.Layout{Direction: ui.Vertical, Height: ui.Fixed(100)}, func() {
			for night := 1; night <= 40; night++ {
				context.Button(fmt.Sprintf("Night %d", night))
			}
		})
	})
	first, err := window.Element("Night 1")
	if err != nil {
		t.Fatal(err)
	}
	last, err := window.Element("Night 40")
	if err != nil {
		t.Fatal(err)
	}
	if !first.Visible || last.Visible {
		t.Fatalf("first visible %v, last visible %v; want true and false", first.Visible, last.Visible)
	}
}

func TestAnElementBehindADialogIsNotVisible(t *testing.T) {
	window := newWindow(t, func(context *ui.Context) {
		context.Button("Light the lamp")
		context.Modal("storm", ui.LayerOptions{AlignX: ui.Center, AlignY: ui.Center}, func() {
			context.Button("Close the shutters")
		})
	})
	window.Frame()
	behind, err := window.Element("Light the lamp")
	if err != nil {
		t.Fatal(err)
	}
	front, err := window.Element("Close the shutters")
	if err != nil {
		t.Fatal(err)
	}
	if behind.Visible || !front.Visible {
		t.Fatalf("behind visible %v, front visible %v; want false and true", behind.Visible, front.Visible)
	}
}

func TestTheTextsOfAFrameComeInDrawingOrder(t *testing.T) {
	window := newWindow(t, keeperLog)
	want := []string{"Lighthouse log", "Light the lamp", "Wind: calm"}
	if got := window.Texts(); !reflect.DeepEqual(got, want) {
		t.Fatalf("texts %q, want %q", got, want)
	}
}

func TestTheDrawListOfTheLastFrameCanBeRead(t *testing.T) {
	window := newWindow(t, keeperLog)
	if len(window.DrawList().Commands) == 0 {
		t.Fatal("the draw list is empty")
	}
}

func TestTheFocusedElementSaysSo(t *testing.T) {
	window := newWindow(t, func(context *ui.Context) {
		context.KeyboardNavigation = ui.Always
		context.Button("Light the lamp")
		context.Button("Ring the bell")
	})
	window.State().Feed().SetKey(input.KeyTab, true)
	window.Frame()
	window.State().Feed().SetKey(input.KeyTab, false)
	window.Frame()
	element, err := window.Element("Light the lamp")
	if err != nil {
		t.Fatal(err)
	}
	other, _ := window.Element("Ring the bell") // found above in the same frame
	if !element.Focused || other.Focused {
		t.Fatalf("focused %v and %v; want true and false", element.Focused, other.Focused)
	}
}

func TestTheScreenReaderTreeDoesNotChangeTheLayout(t *testing.T) {
	without := ui.NewWithMeasurer(FixedMeasurer{})
	with := ui.NewWithMeasurer(FixedMeasurer{})
	with.SetScreenReader(true)
	for _, context := range []*ui.Context{without, with} {
		window := &Window{context: context, build: keeperLog, measurer: FixedMeasurer{}}
		window.state = newState(defaultWidth, defaultHeight)
		window.Frame()
	}
	if !reflect.DeepEqual(without.Drawing().Commands, with.Drawing().Commands) {
		t.Fatal("switching the screen reader on changed what was drawn")
	}
}

// twoPanels shows two panels with a button of the same name in each.
func twoPanels(context *ui.Context) {
	context.PanelWith("North", ui.PanelOptions{X: 20, Y: 20}, func() { context.Button("Save") })
	context.PanelWith("South", ui.PanelOptions{X: 20, Y: 300}, func() { context.Button("Save") })
}

func TestAnElementIsFoundWithinAnother(t *testing.T) {
	window := newWindow(t, twoPanels)
	window.Frame()
	north, err := window.ElementWithin("North", "Save")
	if err != nil {
		t.Fatal(err)
	}
	south, err := window.ElementWithin("South", "Save")
	if err != nil {
		t.Fatal(err)
	}
	if north.Rect.Y >= 300 || south.Rect.Y < 300 || !north.Pressable {
		t.Fatalf("found %+v in North and %+v in South", north, south)
	}
}

func TestAnElementMissingWithinAnotherIsNotFound(t *testing.T) {
	window := newWindow(t, twoPanels)
	window.Frame()
	if _, err := window.ElementWithin("North", "Delete"); !errors.Is(err, ErrorElementNotFound) {
		t.Fatalf("a missing inner element gave %v", err)
	}
	if _, err := window.ElementWithin("East", "Save"); !errors.Is(err, ErrorElementNotFound) {
		t.Fatalf("a missing outer element gave %v", err)
	}
}
