//go:build !js && !android && !nintendo3ds

package window

import (
	"errors"
	"hash/maphash"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/metal"
	"github.com/akrck02/go-renderer/opengl"
	"github.com/akrck02/go-renderer/opengl/overlay"
	rendererwindow "github.com/akrck02/go-renderer/window"
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

// SystemPrefersDark tells whether the system is set to a dark theme, asking it through its own
// command: the global interface style on macOS, GNOME's colour scheme on Linux (which KDE and
// most desktops keep too), the apps' theme in the registry on Windows. A system that cannot be
// asked answers false.
func SystemPrefersDark() bool {
	asking := darkQuestions[runtime.GOOS]
	if asking == nil {
		asking = darkQuestions["linux"]
	}
	// a command that fails (macOS in light mode, a Linux without gsettings) has asked nothing:
	// its output is empty, and empty means light
	output, _ := exec.Command(asking[0], asking[1:]...).Output()
	return darkAnswer(runtime.GOOS, string(output))
}

// darkQuestions are the commands that ask each system for its theme.
var darkQuestions = map[string][]string{
	"darwin":  {"defaults", "read", "-g", "AppleInterfaceStyle"},
	"linux":   {"gsettings", "get", "org.gnome.desktop.interface", "color-scheme"},
	"windows": {"reg", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, "/v", "AppsUseLightTheme"},
}

// darkAnswer reads what a system's command answered about its theme: true for a dark one.
func darkAnswer(system, output string) bool {
	switch system {
	case "darwin":
		return strings.TrimSpace(output) == "Dark"
	case "windows":
		return strings.Contains(output, "0x0")
	}
	return strings.Contains(output, "prefer-dark")
}

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
	if options.CapturePath != "" {
		app.Kind, app.CapturePath = rendererwindow.Headless, options.CapturePath
	}
	if options.Backend.usesMetal() && options.Scene == nil { // Metal's window draws flat interfaces only
		err := runMetal(context, app, options.Fonts)
		if !errors.Is(err, metal.Unavailable) || options.Backend == Metal {
			return err
		}
		log.Printf("window: %v; drawing through OpenGL", err)
		shown = shownFrame{seed: shown.seed}
		app.Input = input.NewState()
	}
	return runOpenGL(context, app, view, options.Fonts)
}

// runOpenGL draws the window through OpenGL: the interface alone, or over a scene.
func runOpenGL(context *ui.Context, app *rendererwindow.Application, view *shownScene, fonts overlay.Fonts) error {
	var flat *opengl.FlatRenderer
	var sceneRenderer *opengl.SceneRenderer
	started := time.Now()
	app.Draw = func(app *rendererwindow.Application) error {
		framebufferWidth, framebufferHeight, scale := frameSize(app)
		if view.view != nil {
			if sceneRenderer == nil {
				created, err := opengl.NewSceneRenderer(fonts)
				if err != nil {
					return err
				}
				sceneRenderer = created
			}
			sceneRenderer.OverlayScale = scale
			seconds := time.Since(started).Seconds()
			if !drawsInAreas(view.view) {
				sceneRenderer.Draw(view.view.World, view.view.Camera, framebufferWidth, framebufferHeight, seconds)
				return nil
			}
			if flat == nil {
				created, err := opengl.NewFlatRenderer(fonts)
				if err != nil {
					return err
				}
				flat = created
			}
			drawInAreas(sceneRenderer, flat, view.view, context.Theme.Background, framebufferWidth, framebufferHeight, scale, seconds)
			return nil
		}
		if flat == nil {
			created, err := opengl.NewFlatRenderer(fonts)
			if err != nil {
				return err
			}
			flat = created
		}
		flat.Draw(context.Drawing(), context.Theme.Background, framebufferWidth, framebufferHeight, scale)
		return nil
	}
	return (&opengl.Window{}).Run(app)
}

// nothingToPaint is an empty list, painted to clear the frame.
var nothingToPaint draw2d.List

// drawsInAreas tells whether a view needs areas of the window: the scene has one of its own, or
// the one below has a place to be shown in (a corner of the window, over the scene).
func drawsInAreas(view *SceneView) bool {
	return view.Area != nil || (view.Below != nil && view.Below.Area != nil)
}

// drawInAreas draws a scene, and the one below it, in their areas of the window over the theme's
// background, then the interface over all of them. A scene with no area of its own fills the
// window, with the one below in its corner of it.
func drawInAreas(renderer *opengl.SceneRenderer, flat *opengl.FlatRenderer, view *SceneView, background draw2d.Color, width, height int, scale float32, seconds float64) {
	flat.Draw(&nothingToPaint, background, width, height, scale) // clears the frame to the background
	whole := opengl.Area{Width: width, Height: height}
	if view.Area != nil {
		whole = framebufferArea(*view.Area, height, scale)
	}
	renderer.DrawIn(view.World, view.Camera, whole, seconds)
	if below := view.Below; below != nil && below.Area != nil {
		drawBelow(renderer, below, framebufferArea(*below.Area, height, scale), seconds)
	}
	renderer.DrawOverlay(view.World, width, height)
}

// drawBelow draws the scene below in its area: over the scene behind it, or with its own sky.
func drawBelow(renderer *opengl.SceneRenderer, below *BelowView, area opengl.Area, seconds float64) {
	if below.Over {
		renderer.DrawOver(below.World, below.Camera, area, seconds)
		return
	}
	renderer.DrawIn(below.World, below.Camera, area, seconds)
}

// framebufferArea is an area of the window in the framebuffer's pixels, counted from its bottom as
// OpenGL counts.
func framebufferArea(area draw2d.Rectangle, framebufferHeight int, scale float32) opengl.Area {
	left, top := int(area.X*scale), int(area.Y*scale)
	width, height := int(area.Width*scale), int(area.Height*scale)
	return opengl.Area{X: left, Y: framebufferHeight - top - height, Width: width, Height: height}
}
