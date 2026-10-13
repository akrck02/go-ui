//go:build !js && !android

package window

import (
	"errors"
	"hash/maphash"
	"log"

	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/metal"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/opengl"
	ui "github.com/akrck02/go-ui"
)

// Wake makes the window draw a frame now: call it when the program's data changes on another
// goroutine (a file loaded, a message arrived), since the window otherwise sleeps until input.
// Safe from any goroutine.
func Wake() { opengl.Wake() } // both backends wait for events in GLFW

// RunWith is Run with an interface context of your own (a custom theme or measurer).
func RunWith(context *ui.Context, options Options, build func(context *ui.Context)) error {
	width, height := options.Width, options.Height
	if width <= 0 || height <= 0 {
		width, height = 1280, 800
	}
	shown := shownFrame{seed: maphash.MakeSeed()}
	app := interfaceApplication(context, options, build, &shown)
	app.Width, app.Height = width, height
	app.Renderer = &opengl.OpenGLRenderer{}
	if options.CapturePath != "" {
		app.Type, app.CapturePath = models.HeadlessApplication, options.CapturePath
	}
	if options.Backend.usesMetal() {
		err := runMetal(context, app)
		if !errors.Is(err, metal.ErrUnavailable) || options.Backend == Metal {
			return err
		}
		log.Printf("window: %v; drawing through OpenGL", err)
		shown = shownFrame{seed: shown.seed}
		app.Input = input.NewState()
	}
	return runOpenGL(context, app)
}

// runOpenGL draws the window through OpenGL.
func runOpenGL(context *ui.Context, app *models.Application) error {
	var renderer *opengl.FlatRenderer
	app.Draw = func(app *models.Application) error {
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
	}
	return (&opengl.OpenGL{}).StartLoop(app)
}
