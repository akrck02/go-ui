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

A click is a move, a press and a release. Go UI answers a button when it is let go, as desktops
do, so the release frame is the one where `Button` returns true. After each input the window runs
one more frame, and any frame the interface asked for at once, as a real window does. So what a
click changed is on screen when `Click` returns, even when it was drawn before the button: a tab
bar at the bottom that changes the page above it.

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

Two buttons called "Save" in different panels? `ElementWithin("North", "Save")` finds the one
inside the panel whose title is "North": inside the smallest painted box (a panel, a card) that
holds the outer element. `ElementIn("north", "Save")` does it by the scopes given to `Push`,
outermost first, when you know them. A button with only an icon is found by the `Tooltip` that
follows it: `Click("Add a lighthouse")`. `Texts()` lists every line drawn, in order, and `DrawList()` gives the
whole draw list, for tests about colours and shapes.

## Playing a script

A test can also play a *script*: the same few lines a real window plays for a demo or a tutorial
(see [Driving a window](12-driving-a-window.md)). `uidrive.Parse` reads it and `Play` runs it, a
sixtieth of a second a frame on the window's clock, and returns the error of the first step that
failed:

```go
script, err := uidrive.Parse(`
click "Note"            # a field is found by its placeholder
type "Fog at dawn"
click "Add"
until "Fog at dawn" 1s
`)
if err != nil {
	t.Fatal(err)
}
if err := window.Play(script); err != nil {
	t.Fatal(err)
}
```

Both use one piece of code to find elements and make input (package `uidrive`), so a script that
passes here does the same steps on a device. `uitest.Driver` gives the window any other driver,
which it runs at every frame as a real window does.

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

## Seeing the results in a window

When a test fails, it helps to see what its window showed. Make the window with `Test(t)`:

```go
window, err := uitest.New(build, uitest.Test(t))
```

Then run your tests through the viewer instead of `go test`:

```bash
go run github.com/akrck02/go-ui/cmd/uitest ./...
go run github.com/akrck02/go-ui/cmd/uitest -dark ./editor/... -run TestSaving
```

It runs `go test` with the arguments you give, with recording on, and opens a Go UI window. The
tests are listed on the left, failed ones first. On the right you see the chosen test's error
codes and messages, and a replay of its frames. *Play* shows them one after another, a tenth of
a second each, and stops at the last; *Pause* holds the frame shown. Step through them with
*Previous frame*, *Next frame* or the slider. The pointer is a ring, filled while pressed; fingers are orange dots; the
element an action found has an outline in the theme's focus colour. Click an error's code to jump
to the frame it came after.

Recording costs nothing in a plain `go test`: a window only keeps its frames while the
`UITEST_RECORD` variable names a folder, which the viewer sets for you. The replay shows the test's
window at its own size (smaller only when it does not fit), on its theme's background. The test
measured text with its own measurer and the viewer paints with real fonts, so each recording keeps
how wide every line was measured, and the viewer shrinks a line that would come out wider: the
frame looks as the test laid it out. The viewer needs a display.

## What it cannot do yet

There is no picture of a frame without a display: Go Renderer cannot paint without one yet. Read
`DrawList()` instead, or look at the frames in the viewer.
