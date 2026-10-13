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

// RunWith is Run with an interface context of your own (a custom theme or measurer). In a web page
// the window is the page's canvas (Options.Canvas); it draws through WebGL2 and returns only when
// the program stops, which keeps the WebAssembly program alive.
func RunWith(context *ui.Context, options Options, build func(context *ui.Context)) error {
	canvas, err := browser.OpenCanvas(options.Canvas)
	if err != nil {
		return err
	}
	functions, err := canvas.WebGL(false) // the atlas already smooths every edge of an interface
	if err != nil {
		return err
	}
	renderer, err := gles.NewFlatRenderer(functions)
	if err != nil {
		return err
	}
	if options.Title != "" {
		js.Global().Get("document").Set("title", options.Title)
	}
	shown := shownFrame{seed: maphash.MakeSeed()}
	app := interfaceApplication(context, options, build, &shown)
	draw := func() error {
		width, height, scale := frameSize(app)
		renderer.Draw(context.Drawing(), context.Theme.Background, width, height, scale)
		return nil
	}
	return browser.Run(canvas, app, draw)
}
