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
| `func KeyName(key input.Key) string` | KeyName is the name a script gives a key ("Enter", "A"), or "" for a key it has no name for. |
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

### Action

Action is what a step does, by the word a script writes for it.

```go
type Action string
```

```go
const (
	MoveTo      Action = "move-to"      // move-to "Help": the pointer goes over an element
	Click       Action = "click"        // click "Save"
	DoubleClick Action = "double-click" // double-click "Entry"
	ClickAt     Action = "click-at"     // click-at 40 300: a place, for what has no name
	Tap         Action = "tap"          // tap "Light the lamp": a finger, as on a phone
	Hold        Action = "hold"         // hold "Lens" 1s: a long press, one second when no time is given
	Drag        Action = "drag"         // drag "Boat" 120 0, or drag "Boat" to "Harbour"
	Scroll      Action = "scroll"       // scroll "Log" -3: a negative amount goes down the page
	Type        Action = "type"         // type "Fog at dawn": the focused control takes the text
	Press       Action = "press"        // press Enter, press Shift+Tab
	Wait        Action = "wait"         // wait 0.5s
	Until       Action = "until"        // until "Saved" 10s: wait for an element, five seconds when no time is given
)
```

The actions a script knows.

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

### PlayOptions

PlayOptions choose how a Player plays.

```go
type PlayOptions struct {
	// Report, when set, gets a line for each step and one at the end (report.go): standard
	// output, Android's log, the 3DS's console.
	Report io.Writer
	// Show draws the pointer, the fingers and the element each step found over the interface, so
	// a demo or a tutorial is seen, not only run (show.go).
	Show bool
}
```

### Player

Player plays a script in a window, one frame at a time: it is a Driver. Each step finds its element in what the last frame showed, then gives its input over the frames that follow, as a person would, and lets one more frame show what it changed. A step that fails ends the script.

```go
type Player struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func Play(script Script, options PlayOptions) *Player` | Play makes a player for a script; give it to window.Options.Driver, or to uitest's Play. |
| `func (player *Player) Done() bool` | Done tells whether the script has ended: every step played, or one failed. |
| `func (player *Player) Drive(state *input.State, context *ui.Context)` | Drive plays a frame of the script. |
| `func (player *Player) Err() error` | Err is why the script ended early: the failed step's error, with its code. |
| `func (player *Player) Show(drawing *draw2d.List)` | Show draws what the player does over the interface, when its options ask for it. |

### Recorder

Recorder writes down what a person does in a window, as a script that plays it back: give it to window.Options.Driver, and take Script when the session ends. It adds no input. Steps name the elements pressed, so a replay works when the layout has moved; a press on nothing named is written as a click-at its place, and a drag that starts on nothing named is left out.

```go
type Recorder struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func Record() *Recorder` | Record makes a recorder. |
| `func (recorder *Recorder) Drive(state *input.State, context *ui.Context)` | Drive watches the input the person gave this frame, against what the last frame showed. |
| `func (recorder *Recorder) Script() Script` | Script is what the person has done so far, with the text they were typing. |
| `func (recorder *Recorder) Show(*draw2d.List)` | Show draws nothing: a recording is not seen. |

### Script

Script is a list of steps, played one after another.

```go
type Script []Step
```

| Function or method | |
|---|---|
| `func Parse(text string) (Script, error)` | Parse reads a script: one step a line, blank lines and what follows a # left out. |
| `func (script Script) String() string` | String writes the script in the form Parse reads, one step a line. |

### Step

Step is one line of a script.

```go
type Step struct {
	Action   Action
	Label    string        // the element the step acts on, by the name people see
	Target   string        // where a drag ends, when it ends on another element
	X, Y     float32       // a click-at's place, or a drag's distance, in layout pixels
	Amount   float32       // a scroll's turn of the wheel
	Text     string        // what a type step types
	Keys     []input.Key   // what a press step presses: the modifiers held, then the key
	Duration time.Duration // a wait's time, a hold's, or how long until waits at most
	Line     int           // the line it was read from, for errors; 0 when made in Go
}
```

| Function or method | |
|---|---|
| `func (step Step) String() string` | String writes a step as a script's line. |

---

[← Reference](0-reference.md) · generated from the doc comments by `generate.py`
