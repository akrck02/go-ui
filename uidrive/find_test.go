package uidrive_test

import (
	"errors"
	"testing"

	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uidrive"
)

func TestAButtonIsFoundByItsLabelInTheLiveInterface(test *testing.T) {
	live := newLiveInterface(func(context *ui.Context) {
		context.Label("Lighthouse log")
		context.Button("Light the lamp")
	})
	live.frame()
	button, err := uidrive.Find(live.context, "Light the lamp")
	if err != nil {
		test.Fatal(err)
	}
	if !button.Pressable || !button.Visible || button.Rect.Width == 0 {
		test.Fatalf("found %+v: the button should be visible and pressable", button)
	}
	_, err = uidrive.Find(live.context, "Put out the lamp")
	if !errors.Is(err, uidrive.ErrorElementNotFound) {
		test.Fatalf("a missing label gave %v", err)
	}
}

func TestTheElementUnderAPointIsTheSmallestNamedOne(test *testing.T) {
	live := newLiveInterface(func(context *ui.Context) { context.Button("Light the lamp") })
	live.frame()
	button, err := uidrive.Find(live.context, "Light the lamp")
	if err != nil {
		test.Fatal(err)
	}
	found, isFound := uidrive.FindAt(live.context, button.Rect.X+1, button.Rect.Y+1)
	if !isFound || found.Label != "Light the lamp" {
		test.Fatalf("found %+v at the button's corner", found)
	}
	if _, isFound := uidrive.FindAt(live.context, 790, 590); isFound {
		test.Fatal("nothing named is at the window's far corner")
	}
}
