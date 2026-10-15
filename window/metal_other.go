//go:build !darwin && !js && !android

package window

import (
	"github.com/akrck02/go-renderer/metal"
	"github.com/akrck02/go-renderer/models"
	ui "github.com/akrck02/go-ui"
)

// runMetal fails everywhere but macOS, and the window falls back to OpenGL.
func runMetal(context *ui.Context, app *models.Application) error {
	return (&metal.FlatWindow{}).StartLoop(app)
}
