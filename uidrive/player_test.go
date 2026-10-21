package uidrive_test

import (
	"errors"
	"testing"
	"time"

	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uidrive"
	"github.com/akrck02/go-ui/uitest"
)

// lighthouse is what a made-up lighthouse keeper's interface keeps, for the steps to change.
type lighthouse struct {
	lampClicks, lensLongPresses, doubleClicks int
	note                                      string
	entries                                   []string
	brightness                                float32
	shiftTabs                                 int
}

// keeperInterface is the made-up interface the steps act on.
func (keeper *lighthouse) keeperInterface(state func() *input.State) func(*ui.Context) {
	return func(context *ui.Context) {
		if context.Button("Light the lamp") {
			keeper.lampClicks++
		}
		context.Button("Lens")
		if context.LongPressed() {
			keeper.lensLongPresses++
		}
		if state().Clicks(input.MouseLeft) == 2 {
			keeper.doubleClicks++
		}
		if state().Pressed(input.KeyTab) && state().Down(input.KeyShift) {
			keeper.shiftTabs++
		}
		context.TextFieldWith("note", &keeper.note, ui.TextFieldOptions{Placeholder: "Note"}) // a field is found by its placeholder
		if context.Button("Add an entry") {
			keeper.entries = append(keeper.entries, keeper.note)
		}
		context.Slider("Brightness", &keeper.brightness, 0, 100)
	}
}

// keeperWindow opens a test window on the keeper's interface.
func keeperWindow(test *testing.T, keeper *lighthouse) *uitest.Window {
	test.Helper()
	var window *uitest.Window
	window, err := uitest.New(keeper.keeperInterface(func() *input.State { return window.State() }), uitest.Test(test))
	if err != nil {
		test.Fatal(err)
	}
	window.Frame()
	return window
}

// play reads a script and plays it in a window.
func play(test *testing.T, window *uitest.Window, text string) error {
	test.Helper()
	script, err := uidrive.Parse(text)
	if err != nil {
		test.Fatal(err)
	}
	return window.Play(script)
}

func TestAClickStepPressesTheButtonByItsName(test *testing.T) {
	keeper := &lighthouse{}
	if err := play(test, keeperWindow(test, keeper), `click "Light the lamp"`); err != nil {
		test.Fatal(err)
	}
	if keeper.lampClicks != 1 {
		test.Fatalf("the lamp was clicked %d times", keeper.lampClicks)
	}
}

func TestATypeStepTypesInTheFieldClickedBefore(test *testing.T) {
	keeper := &lighthouse{}
	err := play(test, keeperWindow(test, keeper), "click \"Note\"\ntype \"Fog at dawn\"\nclick \"Add an entry\"")
	if err != nil {
		test.Fatal(err)
	}
	if len(keeper.entries) != 1 || keeper.entries[0] != "Fog at dawn" {
		test.Fatalf("the log holds %q", keeper.entries)
	}
}

func TestAPressStepHoldsItsModifiersWhileItPressesTheKey(test *testing.T) {
	keeper := &lighthouse{}
	if err := play(test, keeperWindow(test, keeper), "press Shift+Tab"); err != nil {
		test.Fatal(err)
	}
	if keeper.shiftTabs != 1 {
		test.Fatalf("Shift+Tab seen %d times", keeper.shiftTabs)
	}
}

func TestATapIsAClickAndAHoldIsALongPress(test *testing.T) {
	keeper := &lighthouse{}
	if err := play(test, keeperWindow(test, keeper), "tap \"Light the lamp\"\nhold \"Lens\" 1s"); err != nil {
		test.Fatal(err)
	}
	if keeper.lampClicks != 1 || keeper.lensLongPresses != 1 {
		test.Fatalf("%d taps on the lamp, %d long presses on the lens", keeper.lampClicks, keeper.lensLongPresses)
	}
}

func TestADoubleClickStepIsSeenAsOne(test *testing.T) {
	keeper := &lighthouse{}
	if err := play(test, keeperWindow(test, keeper), `double-click "Light the lamp"`); err != nil {
		test.Fatal(err)
	}
	if keeper.doubleClicks != 1 {
		test.Fatalf("%d double clicks seen", keeper.doubleClicks)
	}
}

func TestADragStepMovesASliderByItsDistance(test *testing.T) {
	keeper := &lighthouse{}
	window := keeperWindow(test, keeper)
	slider, err := window.Element("Brightness")
	if err != nil {
		test.Fatal(err)
	}
	script := uidrive.Script{{Action: uidrive.Drag, Label: "Brightness", X: slider.Rect.Width}}
	if err := window.Play(script); err != nil {
		test.Fatal(err)
	}
	if keeper.brightness < 90 {
		test.Fatalf("the slider dragged its width to the right is at %v", keeper.brightness)
	}
}

func TestAWaitStepLetsItsTimePass(test *testing.T) {
	window := keeperWindow(test, &lighthouse{})
	start := window.Now()
	if err := play(test, window, "wait 2s"); err != nil {
		test.Fatal(err)
	}
	if waited := window.Now().Sub(start); waited < 2*time.Second || waited > 2100*time.Millisecond {
		test.Fatalf("waited %v for 2s", waited)
	}
}

func TestAnUntilStepWaitsForAnElementToShow(test *testing.T) {
	var window *uitest.Window
	var start time.Time
	window, err := uitest.New(func(context *ui.Context) {
		if window.Now().Sub(start) >= time.Second { // the log is saved a second after the start
			context.Label("Saved")
		}
	})
	if err != nil {
		test.Fatal(err)
	}
	start = window.Now()
	window.Frame()
	if err := play(test, window, `until "Saved" 3s`); err != nil {
		test.Fatal(err)
	}
	err = play(test, window, `until "Lost" 1s`)
	if !errors.Is(err, uidrive.ErrorElementNotFound) {
		test.Fatalf("waiting for what never shows gave %v", err)
	}
}

func TestAStepOnAMissingElementStopsTheScriptWithItsCode(test *testing.T) {
	keeper := &lighthouse{}
	err := play(test, keeperWindow(test, keeper), "click \"Put out the lamp\"\nclick \"Light the lamp\"")
	if uidrive.CodeOf(err) != uidrive.ElementNotFound {
		test.Fatalf("got %v", err)
	}
	if keeper.lampClicks != 0 {
		test.Fatal("the steps after a failed one ran")
	}
}
