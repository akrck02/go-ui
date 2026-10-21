# uitest

```go
import "github.com/akrck02/go-ui/uitest"
```

Package uitest drives a Go UI interface in go test, with no display: it opens
a pretend window, presses, types and waits in it, and tells what it shows.
Every action runs whole frames, as a backend does, and returns an error with a
code when what it looks for is not there.

### Constants and variables

```go
var (
	ErrorInvalidSize         = &Error{Code: InvalidSize}
	ErrorElementNotFound     = &Error{Code: ElementNotFound}
	ErrorElementAmbiguous    = &Error{Code: ElementAmbiguous}
	ErrorElementNotVisible   = &Error{Code: ElementNotVisible}
	ErrorElementNotPressable = &Error{Code: ElementNotPressable}
)
```

The errors to compare with errors.Is; comparing them allocates nothing.

### Code

Code is the stable name of an error of this package: programs compare it, people search for it.

```go
type Code string
```

```go
const (
	InvalidSize         Code = "uitest.invalid-size"
	ElementNotFound     Code = "uitest.element-not-found"
	ElementAmbiguous    Code = "uitest.element-ambiguous"
	ElementNotVisible   Code = "uitest.element-not-visible"
	ElementNotPressable Code = "uitest.element-not-pressable"
)
```

The codes this package returns.

| Function or method | |
|---|---|
| `func CodeOf(err error) Code` | CodeOf is the code of an error of this package, or "" for any other error (and for nil). |

### Element

Element is what the last frame showed of a control, a text or a scroll area.

```go
type Element struct {
	ID        ui.ID       // the one the interface gives it, 0 when it was not on screen
	Label     string      // what a screen reader calls it: a button's text, a described label
	Rect      draw2d.Rect // where it was laid out, in layout pixels
	Visible   bool        // some of it shows: not scrolled away, not behind a dialog or sheet
	Focused   bool        // it has the focus (ui.Context.FocusedID)
	Pressable bool        // a click or a tap does something to it
}
```

### Error

Error is an error of this package: its code, and a message for people.

```go
type Error struct {
	Code    Code
	Message string
}
```

| Function or method | |
|---|---|
| `func (err *Error) Error() string` | Error gives the code and the message. |
| `func (err *Error) Is(target error) bool` | Is tells whether target is an error of this package with the same code, so errors.Is compares codes, never messages. |

### FixedMeasurer

FixedMeasurer measures text with no fonts: every character is half the text size wide. Tests that use it give the same layout on every machine. It has every icon.

```go
type FixedMeasurer struct{}
```

| Function or method | |
|---|---|
| `func (FixedMeasurer) HasIcon(string) bool` | HasIcon says every icon exists. |
| `func (FixedMeasurer) MeasureText(text string, size float32, _ draw2d.Weight, _ float32) draw2d.TextMetrics` | MeasureText gives each character half the size in width. |

### Option

Option chooses something about a test window; give them to New.

```go
type Option func(*settings)
```

| Function or method | |
|---|---|
| `func Measurer(measurer draw2d.Measurer) Option` | Measurer sets how text is measured; without it the window uses FixedMeasurer. |
| `func Scale(factor float32) Option` | Scale sets how many framebuffer pixels a layout pixel takes (2 on most phones). |
| `func Size(width, height float32) Option` | Size sets the window's size in layout pixels: what the interface's Size reports. |
| `func Theme(theme ui.Theme) Option` | Theme sets the interface's theme; without it the window uses ui's default. |

### Window

Window is a pretend window that runs an interface function frame by frame.

```go
type Window struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func New(build func(context *ui.Context), options ...Option) (*Window, error)` | New makes a test window for an interface function: 800x600 layout pixels at scale 1, ui's default theme and FixedMeasurer, unless options say otherwise. |
| `func (window *Window) Context() *ui.Context` | Context is the interface's context, for what this package does not cover yet. |
| `func (window *Window) DrawList() *draw2d.List` | DrawList is what the last frame drew, as go-renderer would paint it. |
| `func (window *Window) Element(label string) (Element, error)` | Element finds what the last frame showed with this label: a button's text, a field's name, a line of text. |
| `func (window *Window) ElementIn(path ...string) (Element, error)` | ElementIn finds a control by its scopes (the names given to ui.Context.Push) and its label, outermost first: ElementIn("north", "Save"). |
| `func (window *Window) Frame()` | Frame runs one frame as a backend does: Begin, the interface, End, then the end of the input step, which forgets this frame's presses and typed text. |
| `func (window *Window) Frames(count int)` | Frames runs a number of frames, with no time passing between them (see Wait). |
| `func (window *Window) Now() time.Time` | Now is the window's time. |
| `func (window *Window) Resize(width, height float32) error` | Resize changes the window's size in layout pixels, keeping its scale; the next frame sees it. |
| `func (window *Window) State() *input.State` | State is the window's input state, for what this package does not cover yet. |
| `func (window *Window) Texts() []string` | Texts is every line of text the last frame drew, in drawing order: headings, labels, buttons' texts, a field's text. |
| `func (window *Window) Wait(duration time.Duration)` | Wait lets time pass: it runs a frame about every sixtieth of a second until the duration has gone, so long presses, double-click intervals and animations see time move as on a screen. |

---

[← Reference](0-reference.md) · generated from the doc comments by `generate.py`
