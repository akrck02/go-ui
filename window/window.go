// Package window runs an application that is only an interface: a window with no 3D scene, where
// go-ui draws everything on the theme's background, in one draw call. On the desktop it opens the
// window through go-renderer (so it needs cgo, and Metal or OpenGL 4.1); built for WebAssembly
// (GOOS=js) the window is a canvas in a web page, drawn through WebGL2; built for Android (with
// gomobile) it is the application's activity, drawn through OpenGL ES.
//
// The window draws only when something happens: input, a resize, a frame a control asks for (a
// caret's blink, a tooltip due) or Wake. In between the program sleeps and uses no processor,
// which a laptop's or a phone's battery needs from an application that is mostly still. Set
// Options.Continuous for one that animates all the time.
//
//	window.Run(window.Options{Title: "Lighthouses", Width: 1280, Height: 800}, func(context *ui.Context) {
//		context.Label("Hello")
//	})
package window

import (
	"flag"
	"fmt"
	"hash/maphash"
	"runtime"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/scene"
	ui "github.com/akrck02/go-ui"
)

// Options configure the window.
type Options struct {
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
}

// SceneView is a 3D scene seen from a camera, drawn behind the interface.
type SceneView struct {
	World  *scene.Scene
	Camera models.Camera
}

// shownScene is the scene the last step asked for, shared by the step and the drawing.
type shownScene struct{ view *SceneView }

// Backend is the graphics interface a window draws through.
type Backend string

const (
	Automatic Backend = ""       // Metal on macOS, falling back to OpenGL if it cannot start; OpenGL elsewhere
	Metal     Backend = "metal"  // only on macOS
	OpenGL    Backend = "opengl" // OpenGL 4.1, on every desktop system
)

// Set reads a backend from a flag: automatic, metal or opengl.
func (backend *Backend) Set(text string) error {
	switch text {
	case "automatic", "":
		*backend = Automatic
	case string(Metal), string(OpenGL):
		*backend = Backend(text)
	default:
		return fmt.Errorf("unknown backend %q: use automatic, metal or opengl", text)
	}
	return nil
}

func (backend *Backend) String() string {
	if *backend == Automatic {
		return "automatic"
	}
	return string(*backend)
}

// usesMetal tells whether a window should try Metal first.
func (backend Backend) usesMetal() bool {
	return backend == Metal || backend == Automatic && runtime.GOOS == "darwin"
}

// RegisterFlags declares -width, -height, -capture, -no-vsync, -dark, -continuous and -backend.
func (options *Options) RegisterFlags(flags *flag.FlagSet) {
	flags.IntVar(&options.Width, "width", 1280, "window width")
	flags.IntVar(&options.Height, "height", 800, "window height")
	flags.StringVar(&options.CapturePath, "capture", "", "draw one frame to this PNG and exit")
	flags.BoolVar(&options.NoVsync, "no-vsync", false, "do not wait for the display")
	flags.BoolVar(&options.Dark, "dark", false, "use the dark theme")
	flags.BoolVar(&options.Continuous, "continuous", false, "draw every frame instead of only when something happens")
	flags.Var(&options.Backend, "backend", "what draws the window: automatic, metal or opengl")
}

// Run opens the window and calls build in every update step, between the interface's Begin and
// End, until the window is closed. It must be called from the main goroutine.
func Run(options Options, build func(context *ui.Context)) error {
	context, err := ui.New()
	if err != nil {
		return err
	}
	if options.Dark {
		context.Theme = ui.DarkTheme()
	}
	return RunWith(context, options, build)
}

// interfaceApplication is the application every window runs: one interface step (Begin, build,
// End) per update step, drawing only when something happens unless Continuous, and no frame when
// it would look like the one shown.
func interfaceApplication(context *ui.Context, options Options, build func(context *ui.Context), shown *shownFrame, view *shownScene) *models.Application {
	return &models.Application{
		Type:         models.WindowApplication,
		Title:        options.Title,
		MinWidth:     options.MinWidth,
		MinHeight:    options.MinHeight,
		Input:        input.NewState(),
		DisableVsync: options.NoVsync,
		DrawOnDemand: !options.Continuous,
		FlatOnly:     options.Scene == nil,
		Unchanged: func(app *models.Application) bool {
			if view.view != nil {
				// a scene can change without the interface changing: the program asks for frames
				// while its camera moves, and is idle otherwise
				*shown = shownFrame{seed: shown.seed}
				return false
			}
			width, height, scale := frameSize(app)
			return shown.same(context.Drawing(), context.Theme.Background, width, height, scale)
		},
		Update: func(app *models.Application, step float64) error {
			context.Begin(app.Input)
			build(context)
			context.End()
			view.view = nil
			if options.Scene != nil {
				unclaimed := app.Input.Unclaimed()
				view.view = options.Scene(&unclaimed, step)
				// the program asks for frames on its copy (while its camera moves): pass them on
				if delay, wanted := unclaimed.TakeFrameRequest(); wanted {
					app.Input.RequestFrame(delay)
				}
			}
			if view.view != nil {
				view.view.World.Overlay.Drawing = context.Drawing() // the interface is painted over the scene
			}
			return nil
		},
	}
}

// frameSize is the framebuffer's size in pixels and its pixels per layout pixel.
func frameSize(app *models.Application) (width, height int, scale float32) {
	state := app.Input
	width, height = state.FramebufferWidth, state.FramebufferHeight
	if width == 0 || height == 0 {
		width, height = app.Width, app.Height
	}
	scale = 1
	if state.WindowWidth > 0 {
		scale = float32(width) / float32(state.WindowWidth)
	}
	return width, height, scale
}

// shownFrame remembers what the window shows, to tell when the next frame would look the same:
// the pointer moving over still parts of the interface, the frame that follows each input.
type shownFrame struct {
	seed          maphash.Seed
	fingerprint   uint64
	background    draw2d.Color
	width, height int
	scale         float32
	valid         bool
}

// same tells whether a frame would look like the one shown, and otherwise remembers it as the
// one about to be shown.
func (shown *shownFrame) same(list *draw2d.List, background draw2d.Color, width, height int, scale float32) bool {
	fingerprint := list.Fingerprint(shown.seed)
	if shown.valid && fingerprint == shown.fingerprint && background == shown.background &&
		width == shown.width && height == shown.height && scale == shown.scale {
		return true
	}
	*shown = shownFrame{seed: shown.seed, fingerprint: fingerprint, background: background, width: width, height: height, scale: scale, valid: true}
	return false
}
