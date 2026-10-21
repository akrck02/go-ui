# uidrive

```go
import "github.com/akrck02/go-ui/uidrive"
```

Package uidrive drives a Go UI interface by the names people see: it finds what
the last frame showed ("Add an entry", "Light the lamp"), makes the input a
person would give it, and plays a script of such steps frame by frame. A real
window runs a driver through window.Options.Driver (a demo, a tutorial, a replay
of a reported bug, a test on the device); uitest runs the same code in go test,
with no display.

### Functions

| Function | |
|---|---|
| `func Texts(context *ui.Context) []string` | Texts is every line of text the last frame drew, in drawing order: headings, labels, buttons' texts, a field's text. |

### Constants and variables

```go
var (
	ErrorElementNotFound     = &Error{Code: ElementNotFound}
	ErrorElementAmbiguous    = &Error{Code: ElementAmbiguous}
	ErrorElementNotVisible   = &Error{Code: ElementNotVisible}
	ErrorElementNotPressable = &Error{Code: ElementNotPressable}
	ErrorScriptNotRead       = &Error{Code: ScriptNotRead}
	ErrorScriptStopped       = &Error{Code: ScriptStopped}
)
```

The errors to compare with errors.Is; comparing them allocates nothing.

### Code

Code is the stable name of an error: programs compare it, people search for it.

```go
type Code string
```

```go
const (
	ElementNotFound     Code = "uitest.element-not-found"
	ElementAmbiguous    Code = "uitest.element-ambiguous"
	ElementNotVisible   Code = "uitest.element-not-visible"
	ElementNotPressable Code = "uitest.element-not-pressable"
	ScriptNotRead       Code = "uidrive.script-not-read"
	ScriptStopped       Code = "uidrive.script-stopped"
)
```

The codes this package returns.

| Function or method | |
|---|---|
| `func CodeOf(err error) Code` | CodeOf is the code of an Error, or "" for any other error (and for nil). |

### Driver

Driver is what a window runs at each frame when it is driven. Drive comes after the window gathered the frame's real input and before the interface runs: it may add input to the state, and read what the last frame showed through the context. Show comes after the interface ran, and may draw over what it drew.

```go
type Driver interface {
	Drive(state *input.State, context *ui.Context)
	Show(drawing *draw2d.List)
}
```

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

| Function or method | |
|---|---|
| `func Find(context *ui.Context, label string) (Element, error)` | Find finds what the last frame showed with this label: a button's text, a field's name, a line of text. |
| `func FindAt(context *ui.Context, x, y float32) (Element, bool)` | FindAt finds the smallest named element of the last frame under a place in the window, in layout pixels: what a person pressed there. |
| `func FindIn(context *ui.Context, path ...string) (Element, error)` | FindIn finds a control by its scopes (the names given to ui.Context.Push) and its label, outermost first: FindIn(context, "north", "Save"). |
| `func FindPressable(context *ui.Context, label string) (Element, error)` | FindPressable finds an element that shows and can be pressed: ElementNotPressable when a click on it would do nothing. |
| `func FindVisible(context *ui.Context, label string) (Element, error)` | FindVisible finds an element that shows, or says why it cannot be used: ElementNotVisible when it is scrolled away or behind a dialog. |
| `func FindWithin(context *ui.Context, outer, inner string) (Element, error)` | FindWithin finds an element by its label inside another found by its label: the button "Save" in the panel called "North". |

### Error

Error is an error with a code, and a message for people.

```go
type Error struct {
	Code    Code
	Message string
	Cause   error // what went wrong underneath, when something outside this package failed
}
```

| Function or method | |
|---|---|
| `func (err *Error) Error() string` | Error gives the code and the message. |
| `func (err *Error) Is(target error) bool` | Is tells whether target is an Error with the same code, so errors.Is compares codes, never messages. |
| `func (err *Error) Unwrap() error` | Unwrap is the cause underneath, for errors.Is and errors.As. |

### Hand

Hand makes the input a person gives with a mouse, fingers and a keyboard, on an input state, as a backend would: the first finger down also moves and presses the pointer, as a phone's backend does, so controls answer a tap as a click. It remembers what it holds, to let go of all of it. The zero Hand holds nothing.

```go
type Hand struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func (hand *Hand) Down() bool` | Down tells whether the main button, or the finger that is the pointer, is down. |
| `func (hand *Hand) HoldKey(state *input.State, key input.Key)` | HoldKey puts a key down; it stays down until ReleaseKey or LetGo. |
| `func (hand *Hand) LetGo(state *input.State)` | LetGo lifts every finger and lets go of the button and every key the hand holds: what a script does when it stops halfway, so nothing stays pressed. |
| `func (hand *Hand) MoveTo(state *input.State, x, y float32)` | MoveTo moves the pointer to a place in the window, in layout pixels. |
| `func (hand *Hand) Placed() bool` | Placed tells whether the hand has put the pointer somewhere in the window yet. |
| `func (hand *Hand) Press(state *input.State)` | Press holds the main button down where the pointer is. |
| `func (hand *Hand) Release(state *input.State)` | Release lets the main button go. |
| `func (hand *Hand) ReleaseKey(state *input.State, key input.Key)` | ReleaseKey lets a key go. |
| `func (hand *Hand) Scroll(state *input.State, amount float32)` | Scroll turns the wheel where the pointer is: a negative amount goes down the page, as a wheel turned toward the person does. |
| `func (hand *Hand) TouchDown(state *input.State, finger int64, x, y float32)` | TouchDown puts a finger on the screen. |
| `func (hand *Hand) TouchMove(state *input.State, finger int64, x, y float32)` | TouchMove moves a finger that is down. |
| `func (hand *Hand) TouchUp(state *input.State, finger int64)` | TouchUp lifts a finger. |
| `func (hand *Hand) Type(state *input.State, text string)` | Type types a text at once, as a paste or a fast typist does: the control with the focus takes it. |

---

[← Reference](0-reference.md) · generated from the doc comments by `generate.py`
