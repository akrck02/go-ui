package uitest

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

func TestTypedTextReachesTheFocusedField(t *testing.T) {
	note := ""
	window := newWindow(t, func(context *ui.Context) {
		context.TextField("note", &note)
	})
	window.Context().Focus("note")
	window.Type("Fog at dawn")
	window.TypeOneByOne("!")
	if note != "Fog at dawn!" {
		t.Fatalf("the field holds %q", note)
	}
	window.PressKey(input.KeyBackspace)
	if note != "Fog at dawn" {
		t.Fatalf("after Backspace the field holds %q", note)
	}
}

func TestTabMovesTheFocusToTheNextControl(t *testing.T) {
	window := newWindow(t, func(context *ui.Context) {
		context.KeyboardNavigation = ui.Always
		context.Button("Light the lamp")
		context.Button("Ring the bell")
	})
	window.Tab()
	window.Tab()
	bell, err := window.Element("Ring the bell")
	if err != nil {
		t.Fatal(err)
	}
	if !bell.Focused {
		t.Fatal("two Tabs did not reach the second button")
	}
	window.ShiftTab()
	lamp, _ := window.Element("Light the lamp") // found in this same frame above
	if !lamp.Focused {
		t.Fatal("Shift+Tab did not go back to the first button")
	}
}

func TestAHeldKeyStaysDownUntilReleased(t *testing.T) {
	window := newWindow(t, keeperLog)
	window.HoldKey(input.KeyW)
	window.Frames(2)
	if !window.State().Down(input.KeyW) {
		t.Fatal("a held key was let go")
	}
	window.ReleaseKey(input.KeyW)
	if window.State().Down(input.KeyW) {
		t.Fatal("a released key is still down")
	}
}
