package window

import (
	"github.com/akrck02/go-renderer/metal"
	"github.com/akrck02/go-renderer/models"
	ui "github.com/akrck02/go-ui"
)

// runMetal draws the window through Metal.
func runMetal(context *ui.Context, app *models.Application) error {
	flat := &metal.FlatWindow{}
	app.Draw = func(app *models.Application) error {
		framebufferWidth, framebufferHeight, scale := frameSize(app)
		flat.Renderer().Draw(context.Drawing(), context.Theme.Background, framebufferWidth, framebufferHeight, scale)
		return nil
	}
	return flat.StartLoop(app)
}
