# 🧪 Testing an interface

This page is for someone building a program on Go UI who wants `go test` to check its interface:
press its buttons, type in its fields and see what it shows, with no display. At the end you will
know how to open a test window, act in it, find what it drew and read its errors.

## The shortest test

Package `uitest` opens a pretend window. It runs your interface function frame by frame, exactly as
a real window does, but nothing is painted and no screen is needed. It works in CI too.

```go
package log_test

import (
	"testing"

	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

func TestAddingANoteShowsItInTheLog(t *testing.T) {
	note, entries := "", []string{}
	window, err := uitest.New(func(context *ui.Context) {
		context.TextField("note", &note)
		if context.Button("Add") {
			entries = append(entries, note)
			note = ""
		}
		for _, entry := range entries {
			context.Label(entry)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	window.Frame() // the first frame gives every control its place
	window.Context().Focus("note")
	window.Type("Fog at dawn")
	if err := window.Click("Add"); err != nil {
		t.Fatal(err)
	}
	if _, err := window.Element("Fog at dawn"); err != nil {
		t.Fatal(err)
	}
}
```

The test passes: the note is in the log. The same code, with its output checked, is
`ExampleWindow` in the package.

Why run a frame first? Go UI is immediate-mode: a control learns where it is when the frame ends.
Before the first frame nothing has a place, so nothing can be clicked.

## The window

`uitest.New(build, options...)` takes your interface function and some options:

| Option | Default | What it changes |
|---|---|---|
| `Size(width, height)` | 800 × 600 | The window, in layout pixels |
| `Scale(factor)` | 1 | Framebuffer pixels per layout pixel (2 on most phones) |
| `Theme(theme)` | `ui.LightTheme()` | The theme |
| `Measurer(measurer)` | `FixedMeasurer` | How text is measured |
| `Test(t)` | none | Names the window after its test, so the results viewer can show it (below) |

`FixedMeasurer` makes every character half the text size wide. Real fonts differ from machine to
machine; fixed widths make a layout the same everywhere, so a test that passes on your laptop
passes in CI.

`Frame()` runs one frame, `Frames(count)` several. `Resize` changes the size between frames.

## Acting in it

Every action runs whole frames, so the interface sees it as it would on a screen.

| What | Methods |
|---|---|
| The pointer | `MoveTo`, `MoveToElement`, `Press`, `Release`, `Click`, `ClickAt`, `DoubleClick`, `Drag`, `DragElement`, `Scroll` |
| The keyboard | `PressKey`, `HoldKey`, `ReleaseKey`, `Tab`, `ShiftTab`, `Type` (all in one frame), `TypeOneByOne` |
| Touch | `TouchDown`, `TouchMove`, `TouchUp`, `Tap`, `LongPress` |
| Time | `Wait(duration)` |

A click is three frames: move, press, release. Go UI answers a button when it is let go, as
desktops do, so the release frame is the one where `Button` returns true.

### Time stands still

The window has its own clock, and only `Wait` moves it. `Frames(100)` takes no time at all, so a
long press never happens by accident. `Wait(time.Second)` runs a frame every sixtieth of a second
until a second has passed, which is what long presses, double clicks and animations need. Every
run gives the same result, however slow the machine.

## Finding what it drew

`Element(label)` finds a control, a text or a scroll area in the last frame, by the name a screen
reader would read: a button's text, a described label, a line of text. You get back:

| Field | |
|---|---|
| `Rect` | Where it was laid out, in layout pixels |
| `Visible` | Some of it shows: not scrolled away, not behind a dialog |
| `Focused` | It has the focus (`ui.Context.FocusedID`) |
| `Pressable` | A click does something to it |

Two buttons called "Save" in different scopes? `ElementIn("north", "Save")` takes the scopes given
to `Push`, outermost first. `Texts()` lists every line drawn, in order, and `DrawList()` gives the
whole draw list, for tests about colours and shapes.

## Errors

Nothing in `uitest` panics or fails your test for you: methods return errors, and you decide.
Every error has a code, which a program checks with `errors.Is` or `uitest.CodeOf`, never by its
text:

| Code | When |
|---|---|
| `uitest.element-not-found` | No element has that label; the message lists the labels on screen |
| `uitest.element-ambiguous` | Several have it: use `ElementIn` |
| `uitest.element-not-visible` | It is scrolled away or behind a dialog, so it cannot be clicked |
| `uitest.element-not-pressable` | It is only text: a click would do nothing |
| `uitest.invalid-size` | A window with no width or no height |
| `uitest.recording-not-saved`, `uitest.recording-not-read` | A recording for the viewer could not be written or read |

```go
err := window.Click("Remove")
if errors.Is(err, uitest.ErrorElementNotFound) {
	// the log has no Remove button yet
}
```

## What it cannot do yet

There is no picture of a frame: Go Renderer cannot paint without a display yet. Read
`DrawList()` instead.
