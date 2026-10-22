// Command mobile is a small application in the way Android's own apps look: a list of lighthouses
// to switch on and off with their details, a 3D map of an island to turn around and light at any
// time of day, and settings. The same code runs everywhere window.Run does:
//
//	go run ./cmd/mobile                                           # a phone-sized desktop window
//	GOOS=js GOARCH=wasm go build -o web/mobile.wasm ./cmd/mobile  # a web page: ?program=mobile
//	cmd/mobile/build-apk.sh mobile.apk                            # an Android package
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/scene"
	"github.com/akrck02/go-renderer/scene/sample"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"
)

// lighthouse is the program's own data: the interface only shows and changes it.
type lighthouse struct {
	name, place string
	built       int
	lit         bool
	reach       float32 // nautical miles
}

// filter chooses which lighthouses the list shows.
type filter int

const (
	allLighthouses filter = iota
	litLighthouses
	darkLighthouses
)

// The destinations of the navigation bar.
const (
	listScreen = iota
	mapScreen
	settingsScreen
)

// application is everything the program remembers between frames.
type application struct {
	lighthouses []*lighthouse
	screen      int
	chosen      int // the lighthouse whose details are open; -1 = the list
	showing     filter
	// themeChosen is set once the theme is chosen in the settings (or with -dark); until then the
	// interface follows the phone's or the browser's theme
	themeChosen bool
	dark        bool
	hour        float32 // time of day on the map, 0–24

	world  *scene.Scene
	orbit  *camera.Orbit
	camera camera.Camera
	shown  camera.Camera // the camera of the frame on screen, to tell when it still moves
}

func main() {
	options := window.Options{Title: "Lighthouses"}
	options.RegisterFlags(flag.CommandLine)
	flag.Parse()
	options.Width, options.Height = 412, 892 // a phone's size on the desktop

	options.Fonts = roboto() // the interface measures and draws its text in Roboto
	program := newApplication(options.Dark)
	options.Scene = program.scene
	if err := window.Run(options, program.build); err != nil {
		log.Fatal(err)
	}
}

func newApplication(dark bool) *application {
	program := &application{chosen: -1, dark: dark, themeChosen: dark, hour: 15,
		lighthouses: []*lighthouse{
			{name: "North Point", place: "Rocky island, north shore", built: 1850, lit: true, reach: 20},
			{name: "Harbour Light", place: "Breakwater, harbour mouth", built: 1920, lit: true, reach: 6},
			{name: "Long Shoal", place: "Shoal, outer channel", built: 1940, reach: 12},
			{name: "Seal Island", place: "Island, northern channel", built: 1880, lit: true, reach: 14},
			{name: "Gull Rock", place: "Reef, outer bay", built: 1900, reach: 10},
			{name: "West Ledge", place: "Ledge, western approach", built: 1890, lit: true, reach: 16},
		}}
	program.world = sample.Island()
	program.orbit, program.camera = orbitAround(program.world)
	return program
}

// build describes the whole interface for this frame.
func (program *application) build(context *ui.Context) {
	dark := program.isDark()
	colours := lightPalette()
	if dark {
		colours = darkPalette()
	}
	context.Theme = materialTheme(colours, dark)
	look := material{context: context, colours: colours}
	screen := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}
	context.Box(screen, func() {
		switch {
		case program.screen == listScreen && program.chosen >= 0:
			program.details(look)
		case program.screen == listScreen:
			program.list(look)
		case program.screen == mapScreen:
			program.mapOverlay(look)
		default:
			program.settings(look)
		}
		before := program.screen
		look.navigationBar([]destination{{"lighthouse", "Lighthouses"}, {"map", "Map"}, {"settings", "Settings"}}, &program.screen)
		if program.screen != before {
			program.chosen = -1
		}
	})
}

// list is the first screen: filter chips and a row per lighthouse, each with its switch.
func (program *application) list(look material) {
	look.topAppBar("Lighthouses", false)
	context := look.context
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		if window.LastCrash() != "" {
			context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: 16, Right: 16, Bottom: 8}}, func() {
				look.card(func() {
					look.text("The app closed unexpectedly last time", 16, draw2d.Medium, look.colours.onSurface)
					look.text("Settings shows what happened:", 14, draw2d.Regular, look.colours.onSurfaceVariant)
					look.text("a screenshot of it helps fix it", 14, draw2d.Regular, look.colours.onSurfaceVariant)
				})
			})
		}
		context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: 16, Right: 16, Bottom: 8}, Gap: 8}, func() {
			for _, choice := range []struct {
				label string
				which filter
			}{{"All", allLighthouses}, {fmt.Sprintf("Lit · %d", program.countLit()), litLighthouses}, {"Dark", darkLighthouses}} {
				if look.chip(choice.label, program.showing == choice.which) {
					program.showing = choice.which
				}
			}
		})
		for index, light := range program.lighthouses {
			if !program.shows(light) {
				continue
			}
			context.Push(light.name)
			supporting := fmt.Sprintf("%s · %.0f nm", light.place, light.reach)
			if look.listItem("row", "lighthouse", light.name, supporting, func() { look.toggle("lit", &light.lit) }) {
				program.chosen = index
			}
			context.Pop()
		}
	})
}

func (program *application) shows(light *lighthouse) bool {
	switch program.showing {
	case litLighthouses:
		return light.lit
	case darkLighthouses:
		return !light.lit
	}
	return true
}

// details is one lighthouse: what it is, its light and its reach.
func (program *application) details(look material) {
	light := program.lighthouses[program.chosen]
	if look.topAppBar(light.name, true) {
		program.chosen = -1
		return
	}
	context, colours := look.context, look.colours
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 16}, func() {
		look.card(func() {
			look.text(light.place, 16, draw2d.Medium, colours.onSurface)
			look.text(fmt.Sprintf("Built in %d", light.built), 14, draw2d.Regular, colours.onSurfaceVariant)
		})
		look.card(func() {
			context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
				context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
					look.text("Shown at night", 16, draw2d.Regular, colours.onSurface)
					look.text(litText(light.lit), 14, draw2d.Regular, colours.onSurfaceVariant)
				})
				look.toggle("lit", &light.lit)
			})
			context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
				look.text("Reach", 16, draw2d.Regular, colours.onSurface)
				context.Spacer()
				look.text(fmt.Sprintf("%.0f nm", light.reach), 16, draw2d.Medium, colours.primary)
			})
			context.SliderStep("reach", &light.reach, 0, 30, 1)
		})
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8}, func() {
			if look.filledButton("map", "Show on map") {
				program.screen, program.chosen = mapScreen, -1
			}
			label := "Switch on"
			if light.lit {
				label = "Switch off"
			}
			if look.outlinedButton(label) {
				light.lit = !light.lit
			}
		})
	})
}

func litText(lit bool) string {
	if lit {
		return "The light is on"
	}
	return "The light is off"
}

// mapOverlay is what floats over the island: a title card, buttons to come closer and the time
// of day. Where nothing is drawn, a drag turns the camera (see scene).
func (program *application) mapOverlay(look material) {
	context, colours := look.context, look.colours
	insets := context.Insets()
	context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(insets.Top)}, ui.Look{Fill: colours.scrim}, nil)
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 12}, func() {
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(16, 12), Gap: 2},
			ui.Look{Fill: colours.containerHigh, Radius: 28, Shadow: true}, func() {
				look.text("Sample island", 16, draw2d.Medium, colours.onSurface)
				look.text("Drag to turn around it", 14, draw2d.Regular, colours.onSurfaceVariant)
			})
		context.Spacer()
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 12, AlignX: ui.End}, func() {
			if look.floatingButton("plus") {
				program.zoom(0.8)
			}
			if look.floatingButton("minus") {
				program.zoom(1.25)
			}
		})
		look.card(func() {
			context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center, Gap: 8}, func() {
				context.IconSized(dayIcon(program.hour), 20, colours.onSurfaceVariant)
				look.text("Time of day", 16, draw2d.Regular, colours.onSurface)
				context.Spacer()
				look.text(clock(program.hour), 16, draw2d.Medium, colours.primary)
			})
			context.SliderStep("hour", &program.hour, 0, 24, 0.25)
		})
	})
}

func (program *application) zoom(factor float64) {
	orbit := program.orbit
	orbit.Distance = math.Min(math.Max(orbit.Distance*factor, orbit.MinimumDistance), orbit.MaximumDistance)
}

func dayIcon(hour float32) string {
	if hour < 6.5 || hour > 19.5 {
		return "moon"
	}
	return "sun"
}

func clock(hour float32) string {
	minutes := int(math.Round(float64(hour) * 60))
	return fmt.Sprintf("%02d:%02d", minutes/60%24, minutes%60)
}

// scene shows the island on the map screen, turned by the input the interface left over.
func (program *application) scene(state *input.State, step float64) *window.SceneView {
	if program.screen != mapScreen {
		return nil
	}
	program.world.Environment.SunDirection = sunAt(program.hour)
	program.orbit.Update(state, step, &program.camera)
	if program.camera != program.shown {
		state.RequestFrame(0) // the camera still moves: draw the next frame too
		program.shown = program.camera
	}
	return &window.SceneView{World: program.world, Camera: program.camera}
}

// sunAt is where the sun is at an hour: rising in the east at six, highest at noon, setting in the
// west at six in the evening, and under the horizon at night.
func sunAt(hour float32) graphics.Vector {
	angle := (float64(hour) - 6) / 12 * math.Pi
	return graphics.Vector{float32(math.Cos(angle)), float32(math.Sin(angle)), 0.35, 0}.Normalize()
}

// orbitAround places an orbit camera over the middle of the scene, as the desktop viewer does.
func orbitAround(world *scene.Scene) (*camera.Orbit, camera.Camera) {
	minimum, maximum := world.Bounds()
	size := float32(math.Max(float64(maximum[0]-minimum[0]), float64(maximum[2]-minimum[2])))
	if size <= 0 || math.IsInf(float64(size), 0) {
		size = 10
	}
	center := graphics.Vector{(minimum[0] + maximum[0]) / 2, 0, (minimum[2] + maximum[2]) / 2, 1}
	orbit := camera.NewOrbit(center, size*1.3)
	orbit.MaximumDistance = float64(size) * 3
	orbit.Ground = world.GroundHeight
	return orbit, camera.Camera{Near: size * 0.001, Far: size * 10}
}

// settings holds the theme and what the program knows about the screen.
func (program *application) settings(look material) {
	look.topAppBar("Settings", false)
	context, colours := look.context, look.colours
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		look.sectionTitle("Appearance")
		dark := program.isDark()
		explanation := "Follows the phone's setting"
		if program.themeChosen {
			explanation = "Chosen here"
		}
		touched := false
		if look.listItem("dark row", "moon", "Dark theme", explanation, func() { touched = look.toggle("dark", &dark) }) {
			dark, touched = !dark, true
		}
		if touched {
			program.dark, program.themeChosen = dark, true
		}
		look.sectionTitle("This screen")
		width, height := context.Size()
		insets := context.Insets()
		look.listItem("size row", "smartphone", fmt.Sprintf("%.0f × %.0f", width, height), "Size in density-independent pixels", nil)
		look.listItem("insets row", "panel-top", fmt.Sprintf("%.0f at the top, %.0f at the bottom", insets.Top, insets.Bottom),
			"Covered by the system's bars", nil)
		if graphics := window.Graphics(); graphics != "" {
			look.listItem("graphics row", "cpu", "Graphics", graphics, nil)
		}
		if crash := window.LastCrash(); crash != "" {
			look.sectionTitle("Last time the app closed")
			crashReport(look, crash)
		}
		look.sectionTitle("About")
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(16, 0), Gap: 4}, func() {
			look.text("Drawn with go-ui and go-renderer", 14, draw2d.Regular, colours.onSurface)
			look.text("OpenGL ES on Android, WebGL2 in a browser", 14, draw2d.Regular, colours.onSurfaceVariant)
		})
	})
}

// isDark tells which theme to draw in: the one chosen in the settings, or else the system's, asked
// every frame so a phone switching to its dark theme at night is followed.
func (program *application) isDark() bool {
	if program.themeChosen {
		return program.dark
	}
	return window.SystemPrefersDark()
}

// crashReport shows the first lines of a crash's report, small, for a screenshot.
func crashReport(look material, crash string) {
	lines := strings.Split(crash, "\n")
	if len(lines) > 18 {
		lines = lines[:18]
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(16, 0), Gap: 2}
	look.context.Frame(layout, ui.Look{Clip: true}, func() {
		for _, line := range lines {
			look.text(strings.TrimSpace(line), 11, draw2d.Regular, look.colours.onSurfaceVariant)
		}
	})
}

func (program *application) countLit() int {
	lit := 0
	for _, light := range program.lighthouses {
		if light.lit {
			lit++
		}
	}
	return lit
}
