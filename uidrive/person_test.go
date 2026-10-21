package uidrive_test

import (
	"errors"
	"testing"
	"time"

	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-ui/uidrive"
	"github.com/akrck02/go-ui/uitest"
)

func TestAKeyPressedDuringAScriptStopsIt(test *testing.T) {
	script, err := uidrive.Parse("hold \"Lens\" 2s\nclick \"Light the lamp\"")
	if err != nil {
		test.Fatal(err)
	}
	player := uidrive.Play(script, uidrive.PlayOptions{})
	keeper := &lighthouse{}
	window := keeperWindow(test, keeper, uitest.Driver(player))
	window.Wait(500 * time.Millisecond) // the script holds the lens
	window.PressKey(input.KeyEscape)    // a person presses a key
	window.Wait(3 * time.Second)
	if !player.Done() || !errors.Is(player.Err(), uidrive.ErrorScriptStopped) {
		test.Fatalf("done %v, error %v: the person's key should stop the script", player.Done(), player.Err())
	}
	if keeper.lampClicks != 0 || keeper.lensLongPresses != 0 {
		test.Fatalf("%d clicks, %d long presses after the script stopped", keeper.lampClicks, keeper.lensLongPresses)
	}
	if len(window.State().Touches()) != 0 || window.State().Button(input.MouseLeft) {
		test.Fatal("the script's finger stayed on the screen")
	}
}
