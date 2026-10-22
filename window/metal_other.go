//go:build !darwin && !js && !android && !nintendo3ds

package window

import (
	"github.com/akrck02/go-renderer/metal"
	"github.com/akrck02/go-renderer/opengl/overlay"
	rendererwindow "github.com/akrck02/go-renderer/window"
	ui "github.com/akrck02/go-ui"
)

// runMetal fails everywhere but macOS, and the window falls back to OpenGL.
func runMetal(context *ui.Context, app *rendererwindow.Application, fonts overlay.Fonts) error {
	return (&metal.FlatWindow{Fonts: fonts}).Run(app)
}
