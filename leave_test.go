package ui

import (
	"testing"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
)

// noteLayer shows a modal with a note while shown is true, leaving as told.
func noteLayer(shown *bool, leave Leave) func(ui *Context) {
	return func(ui *Context) {
		if !*shown {
			return
		}
		options := LayerOptions{AlignX: Center, AlignY: Center, Leave: leave}
		ui.Modal("note", options, func() { ui.Label("The north light is lit") })
	}
}

// textCommand finds the note's line in the last frame's drawing.
func (h *harness) textCommand(text string) (draw2d.Command, bool) {
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.TextKind && command.Text == text {
			return command, true
		}
	}
	return draw2d.Command{}, false
}

func TestALayerLeavesFadingAndMovingAfterItIsNoLongerShown(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	shown := true
	build := noteLayer(&shown, Leave{Duration: 200 * time.Millisecond, Y: 100})
	h.frame(build)
	before, _ := h.textCommand("The north light is lit")

	shown = false
	h.frame(build)
	*now = now.Add(100 * time.Millisecond)
	h.frame(build)
	halfway, found := h.textCommand("The north light is lit")
	if !found {
		t.Fatal("the note went at once")
	}
	if halfway.Fill[3] >= before.Fill[3] || halfway.Area.Y <= before.Area.Y {
		t.Fatalf("half way the note is at %v with alpha %v; want it lower than %v and fainter than %v",
			halfway.Area.Y, halfway.Fill[3], before.Area.Y, before.Fill[3])
	}
	if _, asked := h.state.TakeFrameRequest(); !asked {
		t.Fatal("a leaving note asked for no next frame")
	}
	*now = now.Add(150 * time.Millisecond)
	h.frame(build)
	if _, found := h.textCommand("The north light is lit"); found {
		t.Fatal("the note is still there after it had time to leave")
	}
}

func TestALayerWithNoLeaveGoesAtOnce(t *testing.T) {
	h := newHarness(t)
	h.withClock()
	shown := true
	build := noteLayer(&shown, Leave{})
	h.frame(build)
	shown = false
	h.frame(build)
	if _, found := h.textCommand("The north light is lit"); found {
		t.Fatal("a layer that does not leave with motion is still drawn")
	}
}

func TestALayerShownAgainIsDrawnOnce(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	shown := true
	build := noteLayer(&shown, Leave{Duration: 200 * time.Millisecond})
	h.frame(build)
	shown = false
	h.frame(build)
	*now = now.Add(50 * time.Millisecond)
	shown = true
	h.frame(build)
	count := 0
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.TextKind && command.Text == "The north light is lit" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("the note shown again is drawn %d times; want once", count)
	}
}

func TestAFoldOpensToItsContentAndClosesToNothing(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	open, built := false, false
	build := func(ui *Context) {
		built = false
		ui.Fold("answer", open, 200*time.Millisecond, Layout{Direction: Vertical, Width: Fixed(300)}, func() {
			built = true
			for range 4 {
				ui.Box(Layout{Width: Grow(), Height: Fixed(30)}, nil)
			}
		})
	}
	step := func(milliseconds int) {
		*now = now.Add(time.Duration(milliseconds) * time.Millisecond)
		h.frame(build)
	}
	step(16)
	if built {
		t.Fatal("a closed fold built its content")
	}
	open = true
	step(16)
	step(100)
	half := h.rect(idOf("answer")).Height
	step(200)
	full := h.rect(idOf("answer")).Height
	if half <= 0 || half >= 120 || full != 120 {
		t.Fatalf("fold %v tall half way and %v open; want between 0 and 120, then 120", half, full)
	}
	open = false
	step(16)
	step(300)
	if built {
		t.Fatal("a fold closed some time ago still builds its content")
	}
}

func TestAnImageIsDrawnInItsBox(t *testing.T) {
	h := newHarness(t)
	photo := draw2d.NewBlankImage(40, 30)
	build := func(ui *Context) {
		ui.Box(Layout{Padding: Pad(10)}, func() { ui.Image(photo, Layout{Width: Fixed(120), Height: Fixed(90)}, 12) })
	}
	h.frame(build)
	h.frame(build)
	for _, command := range h.ui.Drawing().Commands {
		if command.Kind == draw2d.ImageKind && command.Picture == photo {
			if command.Area != (draw2d.Rect{X: 10, Y: 10, Width: 120, Height: 90}) || command.Radius != 12 {
				t.Fatalf("image drawn at %+v with radius %v; want its box, rounded by 12", command.Area, command.Radius)
			}
			return
		}
	}
	t.Fatal("the image was not drawn")
}

func TestTextGrowsWithThePersonsTextSize(t *testing.T) {
	h := newHarness(t)
	size := func() float32 {
		for _, command := range h.ui.Drawing().Commands {
			if command.Kind == draw2d.TextKind && command.Text == "North light" {
				return command.Size
			}
		}
		t.Fatal("the text was not drawn")
		return 0
	}
	build := func(ui *Context) { ui.Text("North light", TextStyle{Size: 16}) }
	h.frame(build)
	if got := size(); got != 16 || h.ui.TextScale() != 1 {
		t.Fatalf("text %v at scale %v by default; want 16 at 1", got, h.ui.TextScale())
	}
	h.ui.SetTextScale(1.3)
	h.frame(build)
	if got := size(); got < 20.7 || got > 20.9 {
		t.Fatalf("text %v at a text size of 1.3; want 20.8", got)
	}
}

// noteClipboard is a clipboard in memory, as a backend gives one.
type noteClipboard struct{ text string }

func (clipboard *noteClipboard) ClipboardText() string        { return clipboard.text }
func (clipboard *noteClipboard) SetClipboardText(text string) { clipboard.text = text }

func TestAProgramCopiesToTheClipboard(t *testing.T) {
	h := newHarness(t)
	clipboard := &noteClipboard{}
	h.state.Clipboard = clipboard
	pasted := ""
	h.frame(func(ui *Context) {
		ui.Copy("54°N 8°W")
		pasted = ui.ClipboardText()
	})
	if clipboard.text != "54°N 8°W" || pasted != "54°N 8°W" {
		t.Fatalf("clipboard %q, read back %q; want the text copied", clipboard.text, pasted)
	}
}
