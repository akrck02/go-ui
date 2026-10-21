package uidrive

import (
	"math"
	"slices"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// What the recorder counts as a pause, a long press and a drag.
const (
	shortestPause      = 300 * time.Millisecond // less is the time a person takes between steps anyway
	longPressTime      = 500 * time.Millisecond // as long as the interface's own long press needs
	pauseResolution    = 100 * time.Millisecond // waits and holds are written to a tenth of a second
	dragDistancePixels = 4                      // a press that moves further is a drag (input.MultiClickDistance)
)

// pressedKeys are the keys a recording writes as press steps; the rest arrive as typed text.
var pressedKeys = []input.Key{
	input.KeyEnter, input.KeyTab, input.KeyEscape, input.KeyBackspace, input.KeyDelete,
	input.KeyUp, input.KeyDown, input.KeyLeft, input.KeyRight, input.KeyHome, input.KeyEnd,
	input.KeyPageUp, input.KeyPageDown,
}

// Recorder writes down what a person does in a window, as a script that plays it back: give it to
// window.Options.Driver, and take Script when the session ends. It adds no input. Steps name the
// elements pressed, so a replay works when the layout has moved; a press on nothing named is
// written as a click-at its place, and a drag that starts on nothing named is left out.
type Recorder struct {
	steps     Script
	typed     []rune    // text typed since the last step, written as one type step
	scrolled  Step      // a scroll being gathered, while the wheel turns over one element
	pressing  *pressing // the main button or the first finger, while it is down
	lastEvent time.Time // when the person last did something, to write the pauses between
}

// pressing is a press being watched until it is let go.
type pressing struct {
	x, y   float32
	at     time.Time
	found  Element
	named  bool // found is an element a script can find again by its name
	touch  bool // a finger, written as a tap or a hold
	double bool // the second press of a double click
}

// Record makes a recorder.
func Record() *Recorder { return &Recorder{} }

// Script is what the person has done so far, with the text they were typing.
func (recorder *Recorder) Script() Script {
	recorder.flush()
	return append(Script(nil), recorder.steps...)
}

// Show draws nothing: a recording is not seen.
func (recorder *Recorder) Show(*draw2d.List) {}

// Drive watches the input the person gave this frame, against what the last frame showed.
func (recorder *Recorder) Drive(state *input.State, context *ui.Context) {
	context.SetScreenReader(true) // what was pressed is named by the screen reader's nodes
	now := state.Time()
	recorder.watchTyping(state, now)
	recorder.watchKeys(state, now)
	recorder.watchScroll(state, context, now)
	recorder.watchPress(state, context, now)
}

// watchTyping gathers typed text, written as one step when something else happens.
func (recorder *Recorder) watchTyping(state *input.State, now time.Time) {
	text := state.TypedText()
	if text == "" {
		return
	}
	if len(recorder.typed) == 0 {
		recorder.flush()
		recorder.pause(now)
	}
	recorder.typed = append(recorder.typed, []rune(text)...)
}

// watchKeys writes the keys that do something of their own (Enter, Tab, the arrows), and letters
// pressed with Control or Super (a shortcut), each with the modifiers held.
func (recorder *Recorder) watchKeys(state *input.State, now time.Time) {
	for key := firstKey; key <= lastKey; key++ {
		if !state.Pressed(key) || slices.Contains(modifierKeys, key) {
			continue
		}
		modifiers := heldModifiers(state)
		shortcut := state.Down(input.KeyControl) || state.Down(input.KeySuper)
		if !shortcut && !slices.Contains(pressedKeys, key) {
			continue // a letter or a digit: it arrives as typed text
		}
		recorder.add(Step{Action: Press, Keys: append(modifiers, key)}, now)
	}
}

// watchScroll gathers the wheel turned over one element into one scroll step.
func (recorder *Recorder) watchScroll(state *input.State, context *ui.Context, now time.Time) {
	if state.ScrollDelta == 0 {
		return
	}
	found, named := namedAt(context, float32(state.CursorX), float32(state.CursorY))
	if !named {
		return // nothing a script could find again
	}
	if recorder.scrolled.Label != found.Label {
		recorder.flush()
		recorder.pause(now)
		recorder.scrolled = Step{Action: Scroll, Label: found.Label}
	}
	recorder.scrolled.Amount += float32(state.ScrollDelta)
}

// watchPress follows the main button, or the first finger, from its press to its release.
func (recorder *Recorder) watchPress(state *input.State, context *ui.Context, now time.Time) {
	x, y := float32(state.CursorX), float32(state.CursorY)
	if state.ButtonPressed(input.MouseLeft) {
		recorder.flush()
		recorder.pause(now)
		found, named := namedAt(context, x, y)
		recorder.pressing = &pressing{x: x, y: y, at: now, found: found, named: named,
			touch: len(state.Touches()) > 0, double: recorder.secondClickOn(state, found)}
	}
	if state.ButtonReleased(input.MouseLeft) && recorder.pressing != nil {
		recorder.released(recorder.pressing, x, y, now)
		recorder.pressing = nil
	}
}

// secondClickOn tells whether a press is the second of a double click on the element the last
// step clicked.
func (recorder *Recorder) secondClickOn(state *input.State, found Element) bool {
	last := len(recorder.steps) - 1
	return state.Clicks(input.MouseLeft) == 2 && last >= 0 &&
		recorder.steps[last].Action == Click && recorder.steps[last].Label == found.Label
}

// released writes the step a press made once it is let go: a click, a tap, a hold, a double click
// or a drag.
func (recorder *Recorder) released(press *pressing, x, y float32, now time.Time) {
	distanceX, distanceY := x-press.x, y-press.y
	held := now.Sub(press.at)
	label := press.found.Label
	switch {
	case math.Hypot(float64(distanceX), float64(distanceY)) > dragDistancePixels:
		if press.named {
			recorder.add(Step{Action: Drag, Label: label, X: distanceX, Y: distanceY}, now)
		}
	case !press.named:
		recorder.add(Step{Action: ClickAt, X: press.x, Y: press.y}, now)
	case press.touch && held >= longPressTime:
		recorder.add(Step{Action: Hold, Label: label, Duration: held.Round(pauseResolution)}, now)
	case press.touch:
		recorder.add(Step{Action: Tap, Label: label}, now)
	case press.double:
		recorder.steps[len(recorder.steps)-1].Action = DoubleClick
		recorder.lastEvent = now
	default:
		recorder.add(Step{Action: Click, Label: label}, now)
	}
}

// add writes a step, after what was being gathered.
func (recorder *Recorder) add(step Step, now time.Time) {
	recorder.flush()
	recorder.steps = append(recorder.steps, step)
	recorder.lastEvent = now
}

// pause writes the time the person waited before what they do now, when it was a real pause.
func (recorder *Recorder) pause(now time.Time) {
	waited := now.Sub(recorder.lastEvent)
	if !recorder.lastEvent.IsZero() && waited >= shortestPause {
		recorder.steps = append(recorder.steps, Step{Action: Wait, Duration: waited.Round(pauseResolution)})
	}
	recorder.lastEvent = now
}

// flush writes the text being typed and the scroll being gathered as their steps.
func (recorder *Recorder) flush() {
	if len(recorder.typed) > 0 {
		recorder.steps = append(recorder.steps, Step{Action: Type, Text: string(recorder.typed)})
		recorder.typed = recorder.typed[:0]
	}
	if recorder.scrolled.Label != "" {
		recorder.steps = append(recorder.steps, recorder.scrolled)
		recorder.scrolled = Step{}
	}
}

// namedAt finds the element under a place, and tells whether a script can find it again by its
// name: it has one, and no other element shares it.
func namedAt(context *ui.Context, x, y float32) (Element, bool) {
	found, isFound := FindAt(context, x, y)
	if !isFound {
		return Element{}, false
	}
	again, err := Find(context, found.Label)
	return found, err == nil && again.ID == found.ID
}

// heldModifiers are the modifier keys down now, in the order a script writes them.
func heldModifiers(state *input.State) []input.Key {
	var held []input.Key
	for _, key := range modifierKeys {
		if state.Down(key) {
			held = append(held, key)
		}
	}
	return held
}
