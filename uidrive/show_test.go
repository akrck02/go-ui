package uidrive_test

import (
	"testing"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-ui/uidrive"
	"github.com/akrck02/go-ui/uitest"
)

func TestAShownPlayerDrawsThePointerAndTheElementItFound(test *testing.T) {
	script, err := uidrive.Parse("move-to \"Light the lamp\"\nwait 1s")
	if err != nil {
		test.Fatal(err)
	}
	player := uidrive.Play(script, uidrive.PlayOptions{Show: true})
	window := keeperWindow(test, &lighthouse{}, uitest.Driver(player))
	button, err := window.Element("Light the lamp")
	if err != nil {
		test.Fatal(err)
	}
	window.Wait(500 * time.Millisecond) // the tutorial points at the lamp's button
	middleX, middleY := button.Rect.X+button.Rect.Width/2, button.Rect.Y+button.Rect.Height/2
	outlined, pointed := false, false
	for _, command := range window.DrawList().Commands {
		if command.Kind != draw2d.RectangleKind || command.Layer != uidrive.ShownLayer {
			continue
		}
		area := command.Area
		outlined = outlined || command.Border.IsVisible() && !command.Fill.IsVisible() &&
			area.X <= button.Rect.X && area.X+area.Width >= button.Rect.X+button.Rect.Width
		pointed = pointed || area.X < middleX && area.X+area.Width > middleX && area.Y < middleY &&
			area.Y+area.Height > middleY && area.Width < button.Rect.Width
	}
	if !outlined || !pointed {
		test.Fatalf("outlined the button %v, drew the pointer on it %v", outlined, pointed)
	}
}

func TestAPlayerNotAskedToShowDrawsNothing(test *testing.T) {
	script, err := uidrive.Parse("move-to \"Light the lamp\"\nwait 1s")
	if err != nil {
		test.Fatal(err)
	}
	window := keeperWindow(test, &lighthouse{}, uitest.Driver(uidrive.Play(script, uidrive.PlayOptions{})))
	window.Wait(500 * time.Millisecond)
	for _, command := range window.DrawList().Commands {
		if command.Layer == uidrive.ShownLayer {
			test.Fatal("a player that was not asked to show drew over the interface")
		}
	}
}

func TestAShownPlayerWithoutOutlineDrawsOnlyThePointer(test *testing.T) {
	script, err := uidrive.Parse("move-to \"Light the lamp\"\nwait 1s")
	if err != nil {
		test.Fatal(err)
	}
	player := uidrive.Play(script, uidrive.PlayOptions{Show: true, WithoutOutline: true})
	window := keeperWindow(test, &lighthouse{}, uitest.Driver(player))
	window.Wait(500 * time.Millisecond)
	shown := 0
	for _, command := range window.DrawList().Commands {
		if command.Layer != uidrive.ShownLayer {
			continue
		}
		shown++
		if command.Kind == draw2d.RectangleKind && command.Border.IsVisible() && !command.Fill.IsVisible() &&
			command.Area.Width > pointerWidest {
			test.Errorf("an outline was drawn around the element: %v", command.Area)
		}
	}
	if shown == 0 {
		test.Error("the pointer was not drawn either")
	}
}

// pointerWidest is wider than anything the pointer draws, in layout pixels: a wider ring is an outline.
const pointerWidest = 40

func TestAGlidingPointerTravelsToTheElementOverSeveralFrames(test *testing.T) {
	script, err := uidrive.Parse("click-at 0 0\nmove-to \"Light the lamp\"\nwait 1s")
	if err != nil {
		test.Fatal(err)
	}
	const glideFrames = 6
	player := uidrive.Play(script, uidrive.PlayOptions{Show: true, GlideFrames: glideFrames})
	window := keeperWindow(test, &lighthouse{}, uitest.Driver(player))
	button, err := window.Element("Light the lamp")
	if err != nil {
		test.Fatal(err)
	}
	middleX := button.Rect.X + button.Rect.Width/2
	var places []float32
	for range 2 * glideFrames {
		window.Frame()
		places = append(places, pointerX(window.DrawList().Commands))
	}
	between := 0
	for _, x := range places {
		if x > 1 && x < middleX-1 {
			between++
		}
	}
	if between < glideFrames/2 {
		test.Errorf("the pointer was seen on its way only %d times: %v", between, places)
	}
}

// pointerX is the middle of the pointer's ring among the shown commands, or -1 when none is drawn.
func pointerX(commands []draw2d.Command) float32 {
	for _, command := range commands {
		if command.Layer == uidrive.ShownLayer && command.Area.Width < pointerWidest {
			return command.Area.X + command.Area.Width/2
		}
	}
	return -1
}
