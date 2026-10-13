//go:build js && wasm

package window

import (
	"hash/maphash"
	"syscall/js"

	"github.com/akrck02/go-renderer/browser"
	"github.com/akrck02/go-renderer/gles"
	ui "github.com/akrck02/go-ui"
)

// Wake makes the page draw a frame now: call it when the program's data changes on another
// goroutine (a fetch finished). Safe from any goroutine.
func Wake() { browser.Wake() }

// Graphics describes a phone's OpenGL ES and GPU; in a browser it is "".
func Graphics() string { return "" }

// LastCrash returns the report of the last run's crash; in a browser the console shows it, and this
// is always "".
func LastCrash() string { return "" }

// Stop ends the window in a web page: Run returns after one more frame and the canvas is let go.
// For a page whose own application shows the program on one of its screens, and takes the screen
// away. Safe from any goroutine.
func Stop() { browser.Stop() }

// SystemPrefersDark tells whether the browser is set to a dark theme (prefers-color-scheme).
func SystemPrefersDark() bool {
	query := js.Global().Call("matchMedia", "(prefers-color-scheme: dark)")
	return query.Truthy() && query.Get("matches").Bool()
}

// RunWith is Run with an interface context of your own (a custom theme or measurer). In a web page
// the window is the page's canvas (Options.Canvas); it draws through WebGL2 and returns only when
// the program stops, which keeps the WebAssembly program alive.
func RunWith(context *ui.Context, options Options, build func(context *ui.Context)) error {
	canvas, err := browser.OpenCanvas(options.Canvas)
	if err != nil {
		return err
	}
	// the atlas already smooths every edge of an interface; a scene's geometry needs antialiasing
	functions, err := canvas.WebGL(options.Scene != nil)
	if err != nil {
		return err
	}
	renderer, err := gles.NewFlatRenderer(functions)
	if err != nil {
		return err
	}
	var sceneRenderer *gles.SceneRenderer
	if options.Title != "" {
		js.Global().Get("document").Set("title", options.Title)
	}
	shown := shownFrame{seed: maphash.MakeSeed()}
	view := &shownScene{}
	app := interfaceApplication(context, options, build, &shown, view)
	draw := func() error {
		width, height, scale := frameSize(app)
		if view.view == nil {
			renderer.Draw(context.Drawing(), context.Theme.Background, width, height, scale)
			return nil
		}
		if sceneRenderer == nil {
			created, err := gles.NewSceneRenderer(functions)
			if err != nil {
				return err
			}
			sceneRenderer = created
		}
		sceneRenderer.OverlayScale = scale
		sceneRenderer.Draw(view.view.World, view.view.Camera, width, height)
		return nil
	}
	return browser.Run(canvas, app, draw)
}
