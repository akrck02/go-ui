package remote

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// A host only ever sees a frame as data it received: what it shows must be what the surface drew.
func TestAFrameSentAsDataDrawsTheSameCommands(test *testing.T) {
	surface := NewSurface(func(context *ui.Context) {
		context.Text("Lighthouse", ui.TextStyle{})
		context.Button("Light it")
	}, uitest.FixedMeasurer{})
	drawn := surface.Draw(400, 300, ui.LightTheme(), Input{})
	if len(drawn.Commands) == 0 {
		test.Fatal("the surface drew nothing")
	}
	data, err := json.Marshal(drawn)
	if err != nil {
		test.Fatal(err)
	}
	var received Frame
	if err := json.Unmarshal(data, &received); err != nil {
		test.Fatal(err)
	}
	var shown draw2d.List
	Show(received, draw2d.Rectangle{X: 0, Y: 0, Width: 400, Height: 300}, &shown)
	for index, command := range shown.Commands {
		original := drawn.Commands[index]
		original.Clip, original.HasClip = command.Clip, command.HasClip // the host keeps it in its area
		if !reflect.DeepEqual(command, original) {
			test.Fatalf("command %d is shown as %+v, drawn as %+v", index, command, original)
		}
	}
}

// The host sends the pointer as input; a press and a release over a button is a click.
func TestAClickSentAsInputReachesAButton(test *testing.T) {
	clicked := false
	surface := NewSurface(func(context *ui.Context) {
		if context.Button("Light it") {
			clicked = true
		}
	}, uitest.FixedMeasurer{})
	theme := ui.LightTheme()
	first := surface.Draw(400, 300, theme, Input{})
	button := firstRectangle(first)
	x, y := button.X+button.Width/2, button.Y+button.Height/2
	surface.Draw(400, 300, theme, Input{PointerX: x, PointerY: y})
	surface.Draw(400, 300, theme, Input{PointerX: x, PointerY: y, PointerDown: true})
	surface.Draw(400, 300, theme, Input{PointerX: x, PointerY: y})
	if !clicked {
		test.Error("a press and a release over the button did not click it")
	}
}

// firstRectangle is the area of a frame's first rectangle: the button's, in a frame of one button.
func firstRectangle(frame Frame) draw2d.Rectangle {
	for _, command := range frame.Commands {
		if command.Kind == draw2d.RectangleKind {
			return command.Area
		}
	}
	return draw2d.Rectangle{}
}
