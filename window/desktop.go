//go:build !js && !android

package window

import (
	"errors"
	"hash/maphash"
	"log"
	"os/exec"
	"runtime"
	"time"

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

// Graphics describes a phone's OpenGL ES and GPU; on the desktop it is "".
func Graphics() string { return "" }

// LastCrash returns the report of the last run's crash; on the desktop the terminal shows it, and
// this is always "".
func LastCrash() string { return "" }

// SystemPrefersDark tells whether the system is set to a dark theme. On the desktop it is not
// asked yet, and the answer is false: use Options.Dark.
func SystemPrefersDark() bool { return false }

// OpenLink opens a link in the application that shows it (the browser for a web page), through
// the computer's own command for that. It returns false when the command failed.
func OpenLink(link string) bool {
	opener := map[string][]string{"darwin": {"open"}, "windows": {"rundll32", "url.dll,FileProtocolHandler"}}[runtime.GOOS]
	if opener == nil {
		opener = []string{"xdg-open"}
	}
	return exec.Command(opener[0], append(opener[1:], link)...).Start() == nil
}

// Share offers a text to other applications; a desktop has no sheet for that, so it returns false
// (put the text on the clipboard instead).
func Share(title, text string) bool { return false }

// Notify shows a notification; on the desktop it does not yet, and returns false.
func Notify(id int, title, text string) bool { return false }

// FeelLongPress gives the buzz of a long press on a phone; a desktop has nothing to buzz.
func FeelLongPress() {}

// Feel gives a buzz on a phone; a desktop has nothing to buzz.
func Feel(Feeling) {}

// NotificationsAllowed tells whether the program may show notifications: not yet on the desktop.
func NotificationsAllowed() bool { return false }

// AskForNotifications asks the person to allow notifications; on the desktop there is nothing to ask.
func AskForNotifications() {}

// SystemPalettes returns the colours a phone made from its wallpaper; a desktop has none.
func SystemPalettes() (Palettes, bool) { return Palettes{}, false }

// RunWith is Run with an interface context of your own (a custom theme or measurer).
func RunWith(context *ui.Context, options Options, build func(context *ui.Context)) error {
	if options.Hidden != nil {
		defer options.Hidden() // the window closed
	}
	width, height := options.Width, options.Height
	if width <= 0 || height <= 0 {
		width, height = 1280, 800
	}
	shown := shownFrame{seed: maphash.MakeSeed()}
	view := &shownScene{}
	app := interfaceApplication(context, options, build, &shown, view)
	app.Width, app.Height = width, height
	app.Renderer = &opengl.OpenGLRenderer{}
	if options.CapturePath != "" {
		app.Type, app.CapturePath = models.HeadlessApplication, options.CapturePath
	}
	if options.Backend.usesMetal() && options.Scene == nil { // Metal's window draws flat interfaces only
		err := runMetal(context, app)
		if !errors.Is(err, metal.ErrUnavailable) || options.Backend == Metal {
			return err
		}
		log.Printf("window: %v; drawing through OpenGL", err)
		shown = shownFrame{seed: shown.seed}
		app.Input = input.NewState()
	}
	return runOpenGL(context, app, view)
}

// runOpenGL draws the window through OpenGL: the interface alone, or over a scene.
func runOpenGL(context *ui.Context, app *models.Application, view *shownScene) error {
	var flat *opengl.FlatRenderer
	var sceneRenderer *opengl.SceneRenderer
	started := time.Now()
	app.Draw = func(app *models.Application) error {
		framebufferWidth, framebufferHeight, scale := frameSize(app)
		if view.view != nil {
			if sceneRenderer == nil {
				created, err := opengl.NewSceneRenderer()
				if err != nil {
					return err
				}
				sceneRenderer = created
			}
			sceneRenderer.OverlayScale = scale
			sceneRenderer.Draw(view.view.World, view.view.Camera, framebufferWidth, framebufferHeight, time.Since(started).Seconds())
			return nil
		}
		if flat == nil {
			created, err := opengl.NewFlatRenderer()
			if err != nil {
				return err
			}
			flat = created
		}
		flat.Draw(context.Drawing(), context.Theme.Background, framebufferWidth, framebufferHeight, scale)
		return nil
	}
	return (&opengl.OpenGL{}).StartLoop(app)
}
