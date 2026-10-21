# window

```go
import "github.com/akrck02/go-ui/window"
```

Package window runs an application that is only an interface: a window with no
3D scene, where go-ui draws everything on the theme's background, in one draw
call. On the desktop it opens the window through go-renderer (so it needs cgo,
and Metal or OpenGL 4.1); built for WebAssembly (GOOS=js) the window is a canvas
in a web page, drawn through WebGL2; built for Android (with gomobile) it is the
application's activity, drawn through OpenGL ES.

The window draws only when something happens: input, a resize, a frame a control
asks for (a caret's blink, a tooltip due) or Wake. In between the program sleeps
and uses no processor, which a laptop's or a phone's battery needs from an
application that is mostly still. Set Options.Continuous for one that animates
all the time.

    window.Run(window.Options{Title: "Lighthouses", Width: 1280, Height: 800}, func(context *ui.Context) {
    	context.Label("Hello")
    })

### Functions

| Function | |
|---|---|
| `func AskForNotifications()` | AskForNotifications asks the person to allow notifications; on the desktop there is nothing to ask. |
| `func Feel(Feeling)` | Feel gives a buzz on a phone; a desktop has nothing to buzz. |
| `func FeelLongPress()` | FeelLongPress gives the buzz of a long press on a phone; a desktop has nothing to buzz. |
| `func ForgetPreference(name string) error` | ForgetPreference lets go of the value kept under a name. |
| `func Graphics() string` | Graphics describes a phone's OpenGL ES and GPU; on the desktop it is "". |
| `func KeepPreference(name, value string) error` | KeepPreference keeps a value under a name, replacing the one there was. |
| `func LastCrash() string` | LastCrash returns the report of the last run's crash; on the desktop the terminal shows it, and this is always "". |
| `func NotificationsAllowed() bool` | NotificationsAllowed tells whether the program may show notifications: not yet on the desktop. |
| `func Notify(id int, title, text string) bool` | Notify shows a notification; on the desktop it does not yet, and returns false. |
| `func OpenLink(link string) bool` | OpenLink opens a link in the application that shows it (the browser for a web page), through the computer's own command for that. |
| `func Preference(name string) (string, bool)` | Preference is the value kept under a name, and whether there is one. |
| `func Run(options Options, build func(context *ui.Context)) error` | Run opens the window and calls build in every update step, between the interface's Begin and End, until the window is closed. |
| `func RunWith(context *ui.Context, options Options, build func(context *ui.Context)) error` | RunWith is Run with an interface context of your own (a custom theme or measurer). |
| `func Share(title, text string) bool` | Share offers a text to other applications; a desktop has no sheet for that, so it returns false (put the text on the clipboard instead). |
| `func SystemPrefersDark() bool` | SystemPrefersDark tells whether the system is set to a dark theme. |
| `func Wake()` | Wake makes the window draw a frame now: call it when the program's data changes on another goroutine (a file loaded, a message arrived), since the window otherwise sleeps until input. |

### Constants and variables

```go
var ErrVideoUnsupported = errors.New("window: films do not play here yet")
```

ErrVideoUnsupported is OpenVideo's answer where films do not play yet.

```go
var PaletteTones = [13]int{0, 10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}
```

PaletteTones are the shades' numbers in Android's names, in the order of each palette: 0 is white, 500 the middle, 1000 black.

### Backend

Backend is the graphics interface a window draws through.

```go
type Backend string
```

```go
const (
	Automatic Backend = ""       // Metal on macOS, falling back to OpenGL if it cannot start; OpenGL elsewhere
	Metal     Backend = "metal"  // only on macOS
	OpenGL    Backend = "opengl" // OpenGL 4.1, on every desktop system
)
```

| Function or method | |
|---|---|
| `func (backend *Backend) Set(text string) error` | Set reads a backend from a flag: automatic, metal or opengl. |
| `func (backend *Backend) String() string` |  |

### BelowView

BelowView is a scene seen from a camera on the 3DS's bottom screen.

```go
type BelowView struct {
	World  *scene.Scene
	Camera models.Camera
}
```

### Feeling

Feeling is a kind of buzz a touch gives, as a phone's own applications give them (Feel).

```go
type Feeling int
```

```go
const (
	// FeelingLongPress is the buzz of a touch that became a long press.
	FeelingLongPress Feeling = iota
	// FeelingTick is a light tick: something chosen, a switch flipped, a step on a dial.
	FeelingTick
	// FeelingConfirm says something went through.
	FeelingConfirm
	// FeelingReject says something went wrong or needs a look.
	FeelingReject
)
```

### Options

Options configure the window.

```go
type Options struct {
	// KeepStart gives the 3DS's START button to the program (a pause) instead of ending it; the
	// HOME menu still ends it. Elsewhere it does nothing.
	KeepStart     bool
	Title         string
	Width, Height int // window size in window pixels; 0 = 1280 × 800
	// MinWidth and MinHeight stop the window from being made smaller than the layout needs; 0 = no
	// limit.
	MinWidth, MinHeight int
	CapturePath         string // draw one frame to this PNG (hidden window) and return, for documentation and tests
	NoVsync             bool   // draw as fast as possible instead of waiting for the display
	Dark                bool   // start with the dark theme
	// Continuous draws every frame, at the display's rate, instead of only when something
	// happens: for an interface that animates all the time. It uses far more battery.
	Continuous bool
	// Backend is what draws the window: Automatic (Metal on macOS, OpenGL elsewhere), Metal or
	// OpenGL. In a web page it is always WebGL2.
	Backend Backend
	// Canvas is the id of the page's canvas to draw in, in a web page; empty = the first canvas.
	// The window is the canvas: Width, Height and Title follow the page.
	Canvas string
	// Scene, when set, puts a 3D scene behind the interface. It is called at every step after the
	// interface, with the input the interface did not claim (so the program can turn its camera
	// without fighting a slider), and returns what to show this frame: nil for no scene. On the
	// desktop a window with a scene draws through OpenGL.
	Scene func(state *input.State, step float64) *SceneView
	// MaxFramesPerSecond caps how often the window draws, to spare a battery; 0 = 60 on a phone,
	// the display's own rate elsewhere (where vsync keeps it).
	MaxFramesPerSecond int
	// Hidden is called when the program leaves the screen: on a phone, another application comes
	// up or the phone locks, and the system may end the program after that without a word; in a
	// browser, the page is hidden; on the desktop, the window closes. Keep what must survive then
	// (KeepPreference). It runs on the goroutine that builds the interface, except in a browser.
	Hidden func()
	// Shown is called when the program comes back on the screen after Hidden: on a phone, back from
	// another application or the lock screen (and at its start); in a browser, when the page shows
	// again. It is the moment to play again what Hidden stopped, such as a sound. On the desktop,
	// where a hidden window has closed, it is not called.
	Shown func()
}
```

| Function or method | |
|---|---|
| `func (options *Options) RegisterFlags(flags *flag.FlagSet)` | RegisterFlags declares -width, -height, -capture, -no-vsync, -dark, -continuous and -backend. |

### Palettes

Palettes are the colours a phone made from its wallpaper (Android 12 and later, Material You): three accent palettes and two neutral ones, each in thirteen shades from white to black, as Android names them (system_accent1_0 … system_accent1_1000). A program that builds its colours from them matches the phone, as the system's own applications do.

```go
type Palettes struct {
	Accent1, Accent2, Accent3, Neutral1, Neutral2 [13]draw2d.Color
}
```

| Function or method | |
|---|---|
| `func SystemPalettes() (Palettes, bool)` | SystemPalettes returns the colours a phone made from its wallpaper; a desktop has none. |

### SceneView

SceneView is a 3D scene seen from a camera, drawn behind the interface.

```go
type SceneView struct {
	World  *scene.Scene
	Camera models.Camera
	// Drawing is the program's own drawing over the scene and under the interface: a game's life
	// bar, a sight. On the 3DS it is painted on the top screen, in each eye when in 3D. Keep the
	// list between frames (Reset it), so drawing it allocates nothing.
	Drawing *draw2d.List
	// Below is a second scene for the 3DS's bottom screen, with the interface painted over it (a
	// map, a turning model); nil leaves the bottom screen to the interface alone. Elsewhere it is
	// not drawn.
	Below *BelowView
}
```

### Video

Video plays a film's pictures (not yet its sound) in a draw2d.Image the program draws, with ui.Context.Image or any piece that shows a picture. A program opens it once, calls Update in every frame it is shown, and plays, pauses or loops it. On Android the phone's own decoders read MP4 and WebM films (H.264, H.265, VP8, VP9, AV1); in a browser, those the browser plays; on the desktop, films do not play yet (OpenVideo says so).

```go
type Video struct {
	Loop bool // start again at the end

	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func OpenVideo(film []byte) (*Video, error)` | OpenVideo opens a film (the bytes of an MP4 or WebM file, often embedded in the program), paused at its start. |
| `func (video *Video) Close()` | Close lets go of the film's decoder. |
| `func (video *Video) Duration() time.Duration` | Duration is how long the film lasts; 0 until known. |
| `func (video *Video) Ended() bool` | Ended tells whether the film reached its end (and does not loop). |
| `func (video *Video) Pause()` | Pause stops the film where it is. |
| `func (video *Video) Picture() *draw2d.Image` | Picture is the film's current picture, to draw; nil until the film's size is known. |
| `func (video *Video) Play()` | Play plays the film from where it is; at its end, from its start. |
| `func (video *Video) Playing() bool` | Playing tells whether the film is playing. |
| `func (video *Video) Position() time.Duration` | Position is how far into the film it is. |
| `func (video *Video) Update(context *ui.Context)` | Update brings the picture to the film's position, and asks for the next frame while it plays. |

---

[← Reference](0-reference.md) · generated from the doc comments by `generate.py`
