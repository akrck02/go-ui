package uitest

import (
	"errors"
	"testing"

	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

func TestAClickOnAButtonPressesIt(t *testing.T) {
	presses := 0
	window := newWindow(t, func(context *ui.Context) {
		if context.Button("Light the lamp") {
			presses++
		}
	})
	if err := window.Click("Light the lamp"); err != nil {
		t.Fatal(err)
	}
	if presses != 1 {
		t.Fatalf("one click pressed the button %d times", presses)
	}
}

func TestAClickOnATextIsNotPressable(t *testing.T) {
	window := newWindow(t, keeperLog)
	err := window.Click("Wind: calm")
	if !errors.Is(err, ErrorElementNotPressable) {
		t.Fatalf("got %v, want %s", err, ElementNotPressable)
	}
}

func TestAClickOnAHiddenElementIsNotVisible(t *testing.T) {
	window := newWindow(t, func(context *ui.Context) {
		context.Button("Light the lamp")
		context.Modal("storm", ui.LayerOptions{AlignX: ui.Center, AlignY: ui.Center}, func() {
			context.Label("A storm is coming")
		})
	})
	window.Frame()
	err := window.Click("Light the lamp")
	if !errors.Is(err, ErrorElementNotVisible) {
		t.Fatalf("got %v, want %s", err, ElementNotVisible)
	}
}

func TestADoubleClickIsSeenAsOne(t *testing.T) {
	doubles := 0
	state := input.NewState()
	window := newWindow(t, func(context *ui.Context) {
		context.Button("Light the lamp")
		if state.Clicks(input.MouseLeft) == 2 {
			doubles++
		}
	})
	state = window.State()
	if err := window.DoubleClick("Light the lamp"); err != nil {
		t.Fatal(err)
	}
	if doubles == 0 {
		t.Fatal("a double click was not seen")
	}
}

func TestADragMovesASlider(t *testing.T) {
	brightness := float32(0)
	window := newWindow(t, func(context *ui.Context) {
		context.Slider("Brightness", &brightness, 0, 100)
	})
	slider, err := window.Element("Brightness")
	if err != nil {
		t.Fatal(err)
	}
	from := center(slider.Rect)
	window.Drag(slider.Rect.X+1, from.y, slider.Rect.X+slider.Rect.Width-1, from.y)
	if brightness < 90 {
		t.Fatalf("dragging the slider to its end gave %v, want about 100", brightness)
	}
}

func TestScrollingShowsWhatWasBelow(t *testing.T) {
	window := newWindow(t, func(context *ui.Context) {
		context.ScrollArea("entries", ui.Layout{Direction: ui.Vertical, Height: ui.Fixed(100)}, func() {
			for _, night := range []string{"Calm", "Fog", "Rain", "Wind", "Storm", "Snow", "Clear"} {
				context.Button(night)
			}
		})
	})
	if err := window.Scroll("Calm", -10); err != nil {
		t.Fatal(err)
	}
	window.Frames(30)
	clear, err := window.Element("Clear")
	if err != nil {
		t.Fatal(err)
	}
	if !clear.Visible {
		t.Fatal("scrolling down did not show the last entry")
	}
}
