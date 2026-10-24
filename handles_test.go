package ui

import (
	"testing"
	"time"
)

// fingerAt puts one finger down at a point, for one step, without lifting it.
func (h *harness) fingerAt(x, y float32, build func(ui *Context)) {
	h.state.Feed().SetTouch(1, float64(x), float64(y))
	h.moveTo(float64(x), float64(y))
	h.press()
	h.frame(build)
}

// tapAt puts a finger down at a point and lifts it.
func (h *harness) tapAt(x, y float32, build func(ui *Context)) {
	h.fingerAt(x, y, build)
	h.liftFinger(build)
}

// tapOn taps the middle of a control laid out in the last frame.
func (h *harness) tapOn(id ID, build func(ui *Context)) {
	area := h.rect(id)
	h.tapAt(area.X+area.Width/2, area.Y+area.Height/2, build)
}

// letterX is the middle of a character of a field's text, on the screen.
func (h *harness) letterX(field ID, text string, index int) float32 {
	return h.rect(field).X + fieldPadding + h.ui.textWidth(text[:index], 0, 0) + h.ui.textWidth("x", 0, 0)/2
}

func TestAFingerHeldOnAWordSelectsItWithHandlesAndABar(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	clipboard := &fakeClipboard{}
	h.state.Clipboard = clipboard
	name := "North light keeper"
	build := func(ui *Context) { ui.TextField("name", &name) }
	h.frame(build)
	field := idOf("name")
	h.tapOn(field, build)
	h.frame(build)
	if !h.isShown(caretHandleID(field)) || h.isShown(startHandleID(field)) {
		t.Fatal("a touch on the field shows the caret's handle")
	}
	area := h.rect(field)
	h.fingerAt(h.letterX(field, name, 8), area.Y+area.Height/2, build) // in "light"
	*now = now.Add(600 * time.Millisecond)
	h.frame(build)
	h.liftFinger(build)
	h.frame(build)
	field0 := &h.ui.memory[field].value.(*fieldMemory).editor
	if field0.selectedText() != "light" {
		t.Fatalf("the word held is selected: %q", field0.selectedText())
	}
	if !h.isShown(startHandleID(field)) || !h.isShown(endHandleID(field)) || !h.isShown(selectionBarID(field)) {
		t.Fatal("a handle under each end, and the bar over the selection")
	}
	if h.ui.focus != field {
		t.Fatal("the field keeps the keyboard")
	}
	copyButton, found := h.ui.Area("##selection bar", "Copy")
	if !found {
		t.Fatal("the bar has Copy")
	}
	h.tapAt(copyButton.X+5, copyButton.Y+5, build)
	h.frame(build)
	if clipboard.text != "light" || h.ui.focus != field || h.isShown(selectionBarID(field)) {
		t.Fatalf("Copy copies, keeps the field typed in and puts the bar away: %q", clipboard.text)
	}
}

func TestAHandleDraggedMovesItsEnd(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	name := "North light keeper"
	build := func(ui *Context) { ui.TextField("name", &name) }
	h.frame(build)
	field := idOf("name")
	area := h.rect(field)
	h.tapOn(field, build)
	h.fingerAt(h.letterX(field, name, 0), area.Y+area.Height/2, build) // in "North"
	*now = now.Add(600 * time.Millisecond)
	h.frame(build)
	h.liftFinger(build)
	h.frame(build)
	handle := h.rect(endHandleID(field))
	h.fingerAt(handle.X+handle.Width/2, handle.Y+handle.Height/2, build)
	target := h.rect(field).X + fieldPadding + h.ui.textWidth("North light", 0, 0)
	grabbed := handle.X + handle.Width/2 - (h.rect(field).X + fieldPadding + h.ui.textWidth("North", 0, 0))
	h.moveTo(float64(target+grabbed), float64(handle.Y+handle.Height/2))
	h.state.Feed().SetTouch(1, float64(target+grabbed), float64(handle.Y+handle.Height/2))
	h.frame(build)
	if selected := h.ui.memory[field].value.(*fieldMemory).editor.selectedText(); selected != "North light" {
		t.Fatalf("the end follows the handle's point: %q", selected)
	}
	if h.isShown(selectionBarID(field)) {
		t.Fatal("the bar is away while a handle is dragged")
	}
	h.liftFinger(build)
	h.frame(build)
	if !h.isShown(selectionBarID(field)) || h.ui.focus != field {
		t.Fatal("let go, the bar comes back and the field is still typed in")
	}
}

func TestTheMouseAndTypingShowNoHandles(t *testing.T) {
	h := newHarness(t)
	name := "North light"
	build := func(ui *Context) { ui.TextField("name", &name) }
	h.frame(build)
	field := idOf("name")
	h.clickInside(field, build)
	h.frame(build)
	if h.isShown(caretHandleID(field)) {
		t.Fatal("a click of the mouse shows no handle")
	}
	h.tapOn(field, build)
	h.frame(build)
	h.typeText("s", build)
	h.frame(build)
	if h.isShown(caretHandleID(field)) {
		t.Fatal("typing puts the handle away")
	}
}
