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
const (
	InvalidSize         Code = "uitest.invalid-size"
	ElementNotFound          = uidrive.ElementNotFound
	ElementAmbiguous         = uidrive.ElementAmbiguous
	ElementNotVisible        = uidrive.ElementNotVisible
	ElementNotPressable      = uidrive.ElementNotPressable
	RecordingNotSaved   Code = "uitest.recording-not-saved"
	RecordingNotRead    Code = "uitest.recording-not-read"
)
```

The codes this package returns; the element codes are uidrive's, under the same names.

```go
const RecordVariable = "UITEST_RECORD"
```

RecordVariable is the environment variable that turns recording on: a folder where each test made with the Test option writes what its window showed, for the results viewer (cmd/uitest).

```go
var (
	ErrorInvalidSize         = &Error{Code: InvalidSize}
	ErrorElementNotFound     = uidrive.ErrorElementNotFound
	ErrorElementAmbiguous    = uidrive.ErrorElementAmbiguous
	ErrorElementNotVisible   = uidrive.ErrorElementNotVisible
	ErrorElementNotPressable = uidrive.ErrorElementNotPressable
	ErrorRecordingNotSaved   = &Error{Code: RecordingNotSaved}
	ErrorRecordingNotRead    = &Error{Code: RecordingNotRead}
)
```

The errors to compare with errors.Is; comparing them allocates nothing.

### Code

Code is the stable name of an error: programs compare it, people search for it. It is uidrive's, which finds elements for this package.

```go
type Code = uidrive.Code
```

| Function or method | |
|---|---|
| `func CodeOf(err error) Code` | CodeOf is the code of an error of this package, or "" for any other error (and for nil). |

### Element

Element is what the last frame showed of a control, a text or a scroll area (uidrive.Element).

```go
type Element = uidrive.Element
```

### Error

Error is an error with a code, and a message for people (uidrive.Error).

```go
type Error = uidrive.Error
```

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
| `func Driver(driver uidrive.Driver) Option` | Driver gives the window a driver, which it runs at every frame as a real window does (window.Options.Driver): before the interface, to add input, and after it, to draw over it. |
| `func Measurer(measurer draw2d.Measurer) Option` | Measurer sets how text is measured; without it the window uses FixedMeasurer. |
| `func Scale(factor float32) Option` | Scale sets how many framebuffer pixels a layout pixel takes (2 on most phones). |
| `func Size(width, height float32) Option` | Size sets the window's size in layout pixels: what the interface's Size reports. |
| `func Test(test testing.TB) Option` | Test names the window after a test. |
| `func Theme(theme ui.Theme) Option` | Theme sets the interface's theme; without it the window uses ui's default. |

### RecordedError

RecordedError is an error a window returned, and the frame it came after.

```go
type RecordedError struct {
	Code    Code
	Message string
	Frame   int
}
```

### RecordedFrame

RecordedFrame is one frame: what was drawn, where the pointer and the fingers were, and the action that ran it with the element it found.

```go
type RecordedFrame struct {
	Action     string // `Click "Save"`, empty for a frame run on its own
	Target     draw2d.Rectangle
	HasTarget  bool
	HasPointer bool // the pointer has been somewhere in the window
	PointerX   float32
	PointerY   float32
	Pressed    bool // the main button or the first finger is down
	Touches    []RecordedTouch
	// Commands are what was drawn; pictures are kept by their area only, and a text's Area.Width is
	// its width as the window's measurer measured it
	Commands []draw2d.Command
}
```

### RecordedTouch

RecordedTouch is a finger on the screen.

```go
type RecordedTouch struct{ X, Y float32 }
```

### Recording

Recording is what a test's window showed, frame by frame, and the errors it returned.

```go
type Recording struct {
	Test          string // the test's name, as go test gives it
	Failed        bool
	Width, Height float32      // the window, in layout pixels
	Background    draw2d.Color // the theme's, which a window paints behind the interface
	Errors        []RecordedError
	Frames        []RecordedFrame
}
```

| Function or method | |
|---|---|
| `func LoadRecordings(folder string) ([]Recording, error)` | LoadRecordings reads every recording in a folder, sorted by test name. |

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
| `func (window *Window) Click(label string) error` | Click clicks the middle of an element that can be pressed. |
| `func (window *Window) ClickAt(x, y float32)` | ClickAt moves the pointer to a place, presses and releases there, each with its frames, as a control answers when the button is let go. |
| `func (window *Window) Context() *ui.Context` | Context is the interface's context, for what this package does not cover yet. |
| `func (window *Window) DoubleClick(label string) error` | DoubleClick clicks an element twice with no time between, which the input state counts as a double click. |
| `func (window *Window) Drag(fromX, fromY, toX, toY float32)` | Drag presses at one place, moves to another over several frames and lets go there. |
| `func (window *Window) DragElement(label string, distanceX, distanceY float32) error` | DragElement drags an element's middle by a distance in layout pixels. |
| `func (window *Window) DrawList() *draw2d.List` | DrawList is what the last frame drew, as go-renderer would paint it. |
| `func (window *Window) Element(label string) (Element, error)` | Element finds what the last frame showed with this label: a button's text, a field's name, a line of text. |
| `func (window *Window) ElementIn(path ...string) (Element, error)` | ElementIn finds a control by its scopes (the names given to ui.Context.Push) and its label, outermost first: ElementIn("north", "Save"). |
| `func (window *Window) ElementWithin(outer, inner string) (Element, error)` | ElementWithin finds an element by its label inside another found by its label: the button "Save" in the panel called "North". |
| `func (window *Window) Frame()` | Frame runs one frame as a backend does: the driver if there is one, Begin, the interface, End, then the end of the input step, which forgets this frame's presses and typed text. |
| `func (window *Window) Frames(count int)` | Frames runs a number of frames, with no time passing between them (see Wait). |
| `func (window *Window) HoldKey(key input.Key)` | HoldKey puts a key down and runs a frame; it stays down until ReleaseKey. |
| `func (window *Window) LongPress(label string) error` | LongPress holds a finger still on an element for a second, letting the time pass, then lifts it. |
| `func (window *Window) MoveTo(x, y float32)` | MoveTo moves the pointer to a place in the window and runs a frame. |
| `func (window *Window) MoveToElement(label string) error` | MoveToElement moves the pointer to the middle of an element (see Element) and runs a frame. |
| `func (window *Window) Now() time.Time` | Now is the window's time. |
| `func (window *Window) Play(script uidrive.Script) error` | Play plays a script in the window, as a driven window would (uidrive.Player), with a sixtieth of a second passing on the window's clock at each frame. |
| `func (window *Window) Press()` | Press holds the main button down where the pointer is and runs a frame. |
| `func (window *Window) PressKey(key input.Key)` | PressKey presses a key and lets it go: one frame with it down, one with it up. |
| `func (window *Window) Release()` | Release lets the main button go and runs a frame. |
| `func (window *Window) ReleaseKey(key input.Key)` | ReleaseKey lets a key go and runs a frame. |
| `func (window *Window) Resize(width, height float32) error` | Resize changes the window's size in layout pixels, keeping its scale; the next frame sees it. |
| `func (window *Window) Scroll(label string, amount float32) error` | Scroll puts the pointer over an element and turns the wheel: a negative amount goes down the page, as a wheel turned toward the person does. |
| `func (window *Window) ShiftTab()` | ShiftTab moves the focus to the control before, as Shift+Tab does. |
| `func (window *Window) State() *input.State` | State is the window's input state, for what this package does not cover yet. |
| `func (window *Window) Tab()` | Tab moves the focus to the next control, as the Tab key does. |
| `func (window *Window) Tap(label string) error` | Tap touches the middle of an element that can be pressed and lifts the finger at once. |
| `func (window *Window) Texts() []string` | Texts is every line of text the last frame drew, in drawing order: headings, labels, buttons' texts, a field's text. |
| `func (window *Window) TouchDown(finger int64, x, y float32)` | TouchDown puts a finger on the screen and runs a frame. |
| `func (window *Window) TouchMove(finger int64, x, y float32)` | TouchMove moves a finger that is down and runs a frame. |
| `func (window *Window) TouchUp(finger int64)` | TouchUp lifts a finger and runs a frame. |
| `func (window *Window) Type(text string)` | Type types a text in one frame, as a paste or a fast typist does: the control with the focus takes it. |
| `func (window *Window) TypeOneByOne(text string)` | TypeOneByOne types a text one character a frame, as a person types. |
| `func (window *Window) Wait(duration time.Duration)` | Wait lets time pass: it runs a frame about every sixtieth of a second until the duration has gone, so long presses, double-click intervals and animations see time move as on a screen. |

---

[← Reference](0-reference.md) · generated from the doc comments by `generate.py`
