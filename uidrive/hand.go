package uidrive

import (
	"slices"

	"github.com/akrck02/go-renderer/input"
)

// Hand makes the input a person gives with a mouse, fingers and a keyboard, on an input state, as
// a backend would: the first finger down also moves and presses the pointer, as a phone's backend
// does, so controls answer a tap as a click. It remembers what it holds, to let go of all of it.
// The zero Hand holds nothing.
type Hand struct {
	pointerFinger    int64 // the finger that also moves the pointer, when hasPointerFinger
	hasPointerFinger bool
	down             bool // the main button or the pointer's finger is down
	placed           bool // the pointer has been put somewhere
	keys             []input.Key
	fingers          []int64
}

// Down tells whether the main button, or the finger that is the pointer, is down.
func (hand *Hand) Down() bool { return hand.down }

// Placed tells whether the hand has put the pointer somewhere in the window yet.
func (hand *Hand) Placed() bool { return hand.placed }

// MoveTo moves the pointer to a place in the window, in layout pixels.
func (hand *Hand) MoveTo(state *input.State, x, y float32) {
	state.MoveCursor(float64(x), float64(y))
	hand.placed = true
}

// Press holds the main button down where the pointer is.
func (hand *Hand) Press(state *input.State) {
	state.SetButton(input.MouseLeft, true)
	hand.down = true
}

// Release lets the main button go.
func (hand *Hand) Release(state *input.State) {
	state.SetButton(input.MouseLeft, false)
	hand.down = false
}

// TouchDown puts a finger on the screen. The first finger down is also the pointer.
func (hand *Hand) TouchDown(state *input.State, finger int64, x, y float32) {
	if len(state.Touches()) == 0 {
		hand.pointerFinger, hand.hasPointerFinger = finger, true
	}
	state.SetTouch(finger, float64(x), float64(y))
	if !slices.Contains(hand.fingers, finger) {
		hand.fingers = append(hand.fingers, finger)
	}
	if hand.isPointer(finger) {
		hand.MoveTo(state, x, y)
		hand.Press(state)
	}
}

// TouchMove moves a finger that is down.
func (hand *Hand) TouchMove(state *input.State, finger int64, x, y float32) {
	state.SetTouch(finger, float64(x), float64(y))
	if hand.isPointer(finger) {
		state.MoveCursor(float64(x), float64(y))
	}
}

// TouchUp lifts a finger.
func (hand *Hand) TouchUp(state *input.State, finger int64) {
	state.LiftTouch(finger)
	hand.fingers = slices.DeleteFunc(hand.fingers, func(held int64) bool { return held == finger })
	if hand.isPointer(finger) {
		hand.Release(state)
		hand.hasPointerFinger = false
	}
}

// isPointer tells whether a finger is the one that moves the pointer.
func (hand *Hand) isPointer(finger int64) bool {
	return hand.hasPointerFinger && finger == hand.pointerFinger
}

// HoldKey puts a key down; it stays down until ReleaseKey or LetGo.
func (hand *Hand) HoldKey(state *input.State, key input.Key) {
	state.SetKey(key, true)
	if !slices.Contains(hand.keys, key) {
		hand.keys = append(hand.keys, key)
	}
}

// ReleaseKey lets a key go.
func (hand *Hand) ReleaseKey(state *input.State, key input.Key) {
	state.SetKey(key, false)
	hand.keys = slices.DeleteFunc(hand.keys, func(held input.Key) bool { return held == key })
}

// Type types a text at once, as a paste or a fast typist does: the control with the focus takes it.
func (hand *Hand) Type(state *input.State, text string) {
	for _, character := range text {
		state.AddCharacter(character)
	}
}

// LetGo lifts every finger and lets go of the button and every key the hand holds: what a script
// does when it stops halfway, so nothing stays pressed.
func (hand *Hand) LetGo(state *input.State) {
	for len(hand.fingers) > 0 {
		hand.TouchUp(state, hand.fingers[0])
	}
	for len(hand.keys) > 0 {
		hand.ReleaseKey(state, hand.keys[0])
	}
	if hand.down {
		hand.Release(state)
	}
}

// Scroll turns the wheel where the pointer is: a negative amount goes down the page, as a wheel
// turned toward the person does.
func (hand *Hand) Scroll(state *input.State, amount float32) { state.AddScroll(float64(amount)) }
