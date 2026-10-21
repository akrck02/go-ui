package uidrive

import (
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// dragSteps is how many frames a drag takes to go from where it starts to where it ends, so
// controls that follow the pointer see it move.
const dragSteps = 10

// touchFinger is the finger a script's taps and holds use.
const touchFinger = 1

// plan finds a step's element in what the last frame showed and makes its moves, a frame each,
// ending with one frame that only lets the interface show what the step changed.
func plan(hand *Hand, step Step, context *ui.Context) ([]move, Element, error) {
	found, err := findFor(step, context)
	if err != nil {
		return nil, Element{}, err
	}
	x, y := center(found)
	moves, err := movesFor(hand, step, context, x, y)
	if err != nil {
		return nil, Element{}, err
	}
	return append(moves, settle), found, nil
}

// findFor finds the element a step acts on: a pressable one for a click or a tap, a visible one
// for the rest. Steps on no element find nothing.
func findFor(step Step, context *ui.Context) (Element, error) {
	switch step.Action {
	case Click, DoubleClick, Tap, Hold:
		return FindPressable(context, step.Label)
	case MoveTo, Drag, Scroll:
		return FindVisible(context, step.Label)
	case ClickAt:
		return Element{Rect: rectangleAt(step.X, step.Y)}, nil
	}
	return Element{}, nil
}

// movesFor makes the moves of a step at a place, the middle of its element.
func movesFor(hand *Hand, step Step, context *ui.Context, x, y float32) ([]move, error) {
	switch step.Action {
	case MoveTo:
		return []move{moveTo(hand, x, y)}, nil
	case Click, ClickAt:
		return []move{moveTo(hand, x, y), press(hand), release(hand)}, nil
	case DoubleClick:
		return []move{moveTo(hand, x, y), press(hand), release(hand), press(hand), release(hand)}, nil
	case Tap:
		return []move{touchDown(hand, x, y), touchUp(hand)}, nil
	case Hold:
		return []move{touchDown(hand, x, y), waitFor(step.Duration), touchUp(hand)}, nil
	case Drag:
		return dragMoves(hand, step, context, x, y)
	case Scroll:
		return []move{moveTo(hand, x, y), scroll(hand, step.Amount)}, nil
	case Type:
		return []move{typeText(hand, step.Text)}, nil
	case Press:
		return []move{holdKeys(hand, step.Keys), releaseKeys(hand, step.Keys)}, nil
	case Wait:
		return []move{waitFor(step.Duration)}, nil
	}
	return []move{waitUntilShown(step.Label, step.Duration)}, nil // Until
}

// dragMoves presses at a place, moves over several frames by the step's distance or to its target
// element, and lets go there.
func dragMoves(hand *Hand, step Step, context *ui.Context, fromX, fromY float32) ([]move, error) {
	toX, toY := fromX+step.X, fromY+step.Y
	if step.Target != "" {
		target, err := FindVisible(context, step.Target)
		if err != nil {
			return nil, err
		}
		toX, toY = center(target)
	}
	moves := []move{moveTo(hand, fromX, fromY), press(hand)}
	for part := 1; part <= dragSteps; part++ {
		share := float32(part) / dragSteps
		moves = append(moves, moveTo(hand, fromX+(toX-fromX)*share, fromY+(toY-fromY)*share))
	}
	return append(moves, release(hand)), nil
}

// The moves of one frame each.

func moveTo(hand *Hand, x, y float32) move {
	return func(state *input.State, _ *ui.Context) (bool, error) { hand.MoveTo(state, x, y); return true, nil }
}

func press(hand *Hand) move {
	return func(state *input.State, _ *ui.Context) (bool, error) { hand.Press(state); return true, nil }
}

func release(hand *Hand) move {
	return func(state *input.State, _ *ui.Context) (bool, error) { hand.Release(state); return true, nil }
}

func touchDown(hand *Hand, x, y float32) move {
	return func(state *input.State, _ *ui.Context) (bool, error) {
		hand.TouchDown(state, touchFinger, x, y)
		return true, nil
	}
}

func touchUp(hand *Hand) move {
	return func(state *input.State, _ *ui.Context) (bool, error) {
		hand.TouchUp(state, touchFinger)
		return true, nil
	}
}

func scroll(hand *Hand, amount float32) move {
	return func(state *input.State, _ *ui.Context) (bool, error) { hand.Scroll(state, amount); return true, nil }
}

func typeText(hand *Hand, text string) move {
	return func(state *input.State, _ *ui.Context) (bool, error) { hand.Type(state, text); return true, nil }
}

// holdKeys puts down every key of a press, the modifiers first.
func holdKeys(hand *Hand, keys []input.Key) move {
	return func(state *input.State, _ *ui.Context) (bool, error) {
		for _, key := range keys {
			hand.HoldKey(state, key)
		}
		return true, nil
	}
}

// releaseKeys lets go of the keys of a press, the last first.
func releaseKeys(hand *Hand, keys []input.Key) move {
	return func(state *input.State, _ *ui.Context) (bool, error) {
		for index := len(keys) - 1; index >= 0; index-- {
			hand.ReleaseKey(state, keys[index])
		}
		return true, nil
	}
}

// settle gives no input: the frame after a step's last input, where what it changed shows.
func settle(*input.State, *ui.Context) (bool, error) { return true, nil }

// waitFor lets a time pass on the state's clock, counted from the move's first frame.
func waitFor(duration time.Duration) move {
	var end time.Time
	return func(state *input.State, _ *ui.Context) (bool, error) {
		if end.IsZero() {
			end = state.Time().Add(duration)
		}
		return !state.Time().Before(end), nil
	}
}

// waitUntilShown waits until an element shows, at most a time; then it fails with the error that
// finding it gave.
func waitUntilShown(label string, limit time.Duration) move {
	var deadline time.Time
	return func(state *input.State, context *ui.Context) (bool, error) {
		if deadline.IsZero() {
			deadline = state.Time().Add(limit)
		}
		_, err := FindVisible(context, label)
		if err == nil {
			return true, nil
		}
		if state.Time().Before(deadline) {
			return false, nil
		}
		return false, err
	}
}

// center is the middle of an element.
func center(found Element) (x, y float32) {
	return found.Rect.X + found.Rect.Width/2, found.Rect.Y + found.Rect.Height/2
}

// rectangleAt is an empty rectangle at a place, whose middle is the place.
func rectangleAt(x, y float32) draw2d.Rect { return draw2d.Rect{X: x, Y: y} }
