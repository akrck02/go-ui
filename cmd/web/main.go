//go:build js && wasm

// Command web shows go-ui over a 3D scene in a web browser: the lighthouse island, a panel to look
// after its light and a status bar, drawn through WebGL2. Drag the view to turn around the
// island; the panel keeps the pointer while it is over it. Build it and serve the page with:
//
//	GOOS=js GOARCH=wasm go build -o web/web.wasm ./cmd/web
//	cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/
//	python3 -m http.server -d web 8080      # then open http://localhost:8080/?program=web
package main

import (
	"fmt"
	"log"
	"math"

	"github.com/akrck02/go-renderer/browser"
	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/gles"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/opengl/overlay"
	"github.com/akrck02/go-renderer/scene"
	rendererwindow "github.com/akrck02/go-renderer/window"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/internal/island"
)

// lighthouse is the program's own data: the interface only shows and changes it.
type lighthouse struct {
	name     string
	on       bool
	rangeNM  float32
	switches int
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	canvas, err := browser.OpenCanvas("view")
	if err != nil {
		return err
	}
	functions, err := canvas.WebGL(true)
	if err != nil {
		return err
	}
	renderer, err := gles.NewSceneRenderer(functions, overlay.Fonts{})
	if err != nil {
		return err
	}
	interfaceContext, err := ui.New()
	if err != nil {
		return err
	}
	world := island.Build()
	world.Overlay.Drawing = interfaceContext.Drawing()
	light := &lighthouse{name: "North Point", on: true, rangeNM: 18}
	orbit, size := orbitAround(world)
	app := &rendererwindow.Application{Input: input.NewState(), Options: rendererwindow.Options{DrawOnDemand: true},
		Camera: camera.Lens{Near: size * 0.001, Far: size * 20}}
	var shown camera.Lens
	app.Update = func(app *rendererwindow.Application, step float64) error {
		interfaceContext.Begin(app.Input)
		propertiesPanel(interfaceContext, light)
		statusBar(interfaceContext, light)
		interfaceContext.End()
		island.ShowLamp(world, light.on)
		// the camera gets only what the interface did not claim: dragging a slider does not turn it
		unclaimed := app.Input.Unclaimed()
		orbit.Update(&unclaimed, step, &app.Camera)
		if app.Camera != shown {
			app.Input.RequestFrame(0) // the camera eases to a stop: keep drawing until it does
			shown = app.Camera
		}
		return nil
	}
	draw := func() error {
		state := app.Input
		renderer.OverlayScale = float32(state.FramebufferWidth) / float32(max(state.WindowWidth, 1))
		renderer.Draw(world, app.Camera, state.FramebufferWidth, state.FramebufferHeight)
		return nil
	}
	return browser.Run(canvas, app, draw)
}

func propertiesPanel(interfaceContext *ui.Context, light *lighthouse) {
	interfaceContext.PanelWith("Lighthouse", ui.PanelOptions{X: 24, Y: 24, Width: ui.Fixed(280)}, func() {
		interfaceContext.Heading(light.name)
		interfaceContext.Caption("Name")
		interfaceContext.TextFieldWith("Name", &light.name, ui.TextFieldOptions{Placeholder: "Lighthouse name", MaxLength: 32})
		interfaceContext.Checkbox("Lit at night", &light.on)
		interfaceContext.Row(func() {
			interfaceContext.Caption("Range")
			interfaceContext.Spacer()
			interfaceContext.Caption(fmt.Sprintf("%.0f nm", light.rangeNM))
		})
		interfaceContext.SliderStep("range", &light.rangeNM, 0, 30, 1)
		interfaceContext.Separator()
		interfaceContext.Row(func() {
			interfaceContext.Caption(fmt.Sprintf("Switched %d times", light.switches))
			interfaceContext.Spacer()
			label := "Switch off"
			if !light.on {
				label = "Switch on"
			}
			if interfaceContext.PrimaryButton(label + "##switch") {
				light.on = !light.on
				light.switches++
			}
		})
	})
}

// statusBar sits at the bottom of the page, outside any panel: a spacer pushes it down.
func statusBar(interfaceContext *ui.Context, light *lighthouse) {
	theme := interfaceContext.Theme
	interfaceContext.Spacer()
	look := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1}
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(12, 6), Gap: 8, AlignY: ui.Center}
	interfaceContext.Frame(layout, look, func() {
		interfaceContext.IconSized("globe", 14, theme.InkSoft)
		interfaceContext.Caption("In a web browser, through WebGL2")
		interfaceContext.Separator()
		state := "light off"
		if light.on {
			state = fmt.Sprintf("light on · %.0f nm", light.rangeNM)
		}
		interfaceContext.Caption(light.name + " · " + state)
	})
}

// orbitAround places an orbit camera over the middle of the scene, as the desktop viewer does.
func orbitAround(world *scene.Scene) (*camera.Orbit, float32) {
	minimum, maximum := world.Bounds()
	size := float32(math.Max(float64(maximum[0]-minimum[0]), float64(maximum[2]-minimum[2])))
	if size <= 0 || math.IsInf(float64(size), 0) {
		size = 10
	}
	center := graphics.Vector{(minimum[0] + maximum[0]) / 2, 0, (minimum[2] + maximum[2]) / 2, 1}
	orbit := camera.NewOrbit(center, size*1.6)
	orbit.MaximumDistance = float64(size) * 4
	orbit.Ground = world.GroundHeight
	return orbit, size
}
