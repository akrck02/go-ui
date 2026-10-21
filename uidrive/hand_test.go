package uidrive_test

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-ui/uidrive"
)

func TestTheFirstFingerDownAlsoPressesThePointer(test *testing.T) {
	state := input.NewState()
	var hand uidrive.Hand
	hand.TouchDown(state, 1, 40, 30)
	hand.TouchDown(state, 2, 200, 30)
	if !state.Button(input.MouseLeft) || state.CursorX != 40 || !hand.Down() {
		test.Fatalf("pointer at %v pressed %v: the first finger should press it where it is", state.CursorX, state.Button(input.MouseLeft))
	}
	hand.TouchMove(state, 2, 300, 30)
	if state.CursorX != 40 {
		test.Fatal("the second finger moved the pointer")
	}
	hand.TouchUp(state, 1)
	if state.Button(input.MouseLeft) || hand.Down() {
		test.Fatal("lifting the first finger should let the pointer go")
	}
}

func TestLettingGoLiftsEverythingTheHandHolds(test *testing.T) {
	state := input.NewState()
	var hand uidrive.Hand
	hand.TouchDown(state, 1, 40, 30)
	hand.HoldKey(state, input.KeyShift)
	hand.LetGo(state)
	if len(state.Touches()) != 0 || state.Button(input.MouseLeft) || state.Down(input.KeyShift) {
		test.Fatalf("touches %d, button %v, shift %v after letting go", len(state.Touches()), state.Button(input.MouseLeft), state.Down(input.KeyShift))
	}
}
