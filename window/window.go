// Package window runs an application that is only an interface: a window with no 3D scene, where
// go-ui draws everything on the theme's background. It opens the window through go-renderer (so it
// needs cgo and OpenGL 4.1) and paints the interface in one draw call.
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
	"hash/maphash"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/opengl"
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
}

// RegisterFlags declares -width, -height, -capture, -no-vsync, -dark and -continuous.
func (options *Options) RegisterFlags(flags *flag.FlagSet) {
	flags.IntVar(&options.Width, "width", 1280, "window width")
	flags.IntVar(&options.Height, "height", 800, "window height")
	flags.StringVar(&options.CapturePath, "capture", "", "draw one frame to this PNG and exit")
	flags.BoolVar(&options.NoVsync, "no-vsync", false, "do not wait for the display")
	flags.BoolVar(&options.Dark, "dark", false, "use the dark theme")
	flags.BoolVar(&options.Continuous, "continuous", false, "draw every frame instead of only when something happens")
}

// Wake makes the window draw a frame now: call it when the program's data changes on another
// goroutine (a file loaded, a message arrived), since the window otherwise sleeps until input.
// Safe from any goroutine.
func Wake() { opengl.Wake() }

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

// RunWith is Run with an interface context of your own (a custom theme or measurer).
func RunWith(context *ui.Context, options Options, build func(context *ui.Context)) error {
	width, height := options.Width, options.Height
	if width <= 0 || height <= 0 {
		width, height = 1280, 800
	}
	var renderer *opengl.FlatRenderer
	shown := shownFrame{seed: maphash.MakeSeed()}
	app := &models.Application{
		Type:         models.WindowApplication,
		Title:        options.Title,
		Width:        width,
		Height:       height,
		MinWidth:     options.MinWidth,
		MinHeight:    options.MinHeight,
		Input:        input.NewState(),
		Renderer:     &opengl.OpenGLRenderer{},
		DisableVsync: options.NoVsync,
		DrawOnDemand: !options.Continuous,
		FlatOnly:     true,
		Unchanged: func(app *models.Application) bool {
			width, height, scale := frameSize(app)
			return shown.same(context.Drawing(), context.Theme.Background, width, height, scale)
		},
		Update: func(app *models.Application, step float64) error {
			context.Begin(app.Input)
			build(context)
			context.End()
			return nil
		},
		Draw: func(app *models.Application) error {
			if renderer == nil {
				created, err := opengl.NewFlatRenderer()
				if err != nil {
					return err
				}
				renderer = created
			}
			framebufferWidth, framebufferHeight, scale := frameSize(app)
			renderer.Draw(context.Drawing(), context.Theme.Background, framebufferWidth, framebufferHeight, scale)
			return nil
		},
	}
	if options.CapturePath != "" {
		app.Type, app.CapturePath = models.HeadlessApplication, options.CapturePath
	}
	return (&opengl.OpenGL{}).StartLoop(app)
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
