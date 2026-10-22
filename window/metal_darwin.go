package window

import (
	"github.com/akrck02/go-renderer/metal"
	"github.com/akrck02/go-renderer/opengl/overlay"
	rendererwindow "github.com/akrck02/go-renderer/window"
	ui "github.com/akrck02/go-ui"
)

// runMetal draws the window through Metal.
func runMetal(context *ui.Context, app *rendererwindow.Application, fonts overlay.Fonts) error {
	flat := &metal.FlatWindow{Fonts: fonts}
	app.Draw = func(app *rendererwindow.Application) error {
		framebufferWidth, framebufferHeight, scale := frameSize(app)
		flat.Renderer().Draw(context.Drawing(), context.Theme.Background, framebufferWidth, framebufferHeight, scale)
		return nil
	}
	return flat.Run(app)
}
