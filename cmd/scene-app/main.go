// Command scene-app is a traditional desktop application built with go-ui around a 3D view: a menu bar, a toolbar, a
// list on the left, the 3D view in the middle, an inspector on the right and a status bar. The
// whole layout is one tree of boxes on the screen; the middle box is left transparent, so the
// scene shows through it and the camera works there.
//
//	go run ./cmd/scene-app
//	go run ./cmd/scene-app -capture app.png
//	go run ./cmd/scene-app -dark
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/scene"
	"github.com/akrck02/go-renderer/viewer"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/internal/island"
)

// lighthouse is one entry of the register: the program's data, which the interface only shows.
type lighthouse struct {
	name, place string
	built       int
	height      int // metres
	focalHeight int // metres above the sea
	character   string
	rangeNM     int
	lit         bool
	band        graphics.Vector
}

// register is the application's state.
type register struct {
	lighthouses []*lighthouse
	selected    int
	tool        string
	view        string
	saved       bool
}

func main() {
	options := viewer.DefaultOptions()
	options.RegisterFlags(flag.CommandLine)
	dark := flag.Bool("dark", false, "use the dark theme")
	flag.Parse()

	interfaceContext, err := ui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *dark {
		interfaceContext.Theme = ui.DarkTheme()
	}
	world := island.Build()
	world.Overlay.Drawing = interfaceContext.Drawing()
	app := newRegister()
	shown := -1

	update := func(world *scene.Scene, state *input.State, step float64) {
		interfaceContext.Begin(state)
		app.draw(interfaceContext)
		interfaceContext.End()
		current := app.current()
		if shown != app.selected {
			island.Paint(world, current.band)
			shown = app.selected
		}
		island.ShowLamp(world, current.lit)
	}
	if err := viewer.Run(world, options, "Lighthouse register", update); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRegister() *register {
	return &register{
		lighthouses: []*lighthouse{
			{name: "North Point", place: "Rocky island, north shore", built: 1850, height: 30, focalHeight: 40, character: "Fl W 5s", rangeNM: 20, lit: true, band: graphics.Vector{0.8, 0.2, 0.18, 1}},
			{name: "South Cape", place: "Headland, south coast", built: 1870, height: 20, focalHeight: 60, character: "Fl(2) W 10s", rangeNM: 18, lit: true, band: graphics.Vector{0.15, 0.25, 0.55, 1}},
			{name: "Gull Rock", place: "Reef, outer bay", built: 1900, height: 15, focalHeight: 20, character: "Iso R 4s", rangeNM: 10, lit: false, band: graphics.Vector{0.1, 0.1, 0.1, 1}},
			{name: "Harbour Light", place: "Breakwater, harbour mouth", built: 1920, height: 10, focalHeight: 12, character: "Oc G 3s", rangeNM: 6, lit: true, band: graphics.Vector{0.15, 0.5, 0.3, 1}},
		},
		tool: "Select", view: "Perspective", saved: true,
	}
}

func (app *register) current() *lighthouse { return app.lighthouses[app.selected] }

// draw describes the whole window, top to bottom.
func (app *register) draw(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		app.menuBar(context)
		context.Separator()
		app.toolbar(context)
		context.Separator()
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
			app.sidebar(context)
			context.Separator()
			app.viewport(context)
			context.Separator()
			app.inspector(context)
		})
		context.Separator()
		app.statusBar(context)
	})
}

// bar is a full-width strip painted with the surface colour.
func bar(context *ui.Context, height, padding float32, content func()) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(height), Padding: ui.PadXY(padding, 0), Gap: 2, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, content)
}

func (app *register) menuBar(context *ui.Context) {
	bar(context, 30, 6, func() {
		context.IconSized("lamp", 16, context.Theme.Primary)
		context.Box(ui.Layout{Width: ui.Fixed(6)}, nil)
		for _, menu := range []string{"File", "Edit", "View", "Help"} {
			context.FlatButton(menu)
		}
		context.Spacer()
		context.Caption("Lighthouse register")
		context.Box(ui.Layout{Width: ui.Fixed(8)}, nil)
	})
}

func (app *register) toolbar(context *ui.Context) {
	bar(context, 40, 8, func() {
		if context.FlatIconButton("save", "") {
			app.saved = true
		}
		context.FlatIconButton("undo-2", "")
		context.FlatIconButton("redo-2", "")
		app.toolbarGap(context)
		for _, tool := range []struct{ icon, name string }{{"mouse-pointer-2", "Select"}, {"move", "Move"}, {"ruler", "Measure"}} {
			if context.Selectable(tool.icon, tool.name+"##tool", app.tool == tool.name) {
				app.tool = tool.name
			}
		}
		app.toolbarGap(context)
		for _, view := range []string{"Perspective", "Map"} {
			if context.Selectable("", view+"##view", app.view == view) {
				app.view = view
			}
		}
		context.Spacer()
		if context.PrimaryButton("New lighthouse") {
			app.addLighthouse()
		}
	})
}

// toolbarGap is a short vertical line between groups of tools.
func (app *register) toolbarGap(context *ui.Context) {
	context.Box(ui.Layout{Height: ui.Fixed(22), Padding: ui.PadXY(6, 0)}, func() { context.Separator() })
}

func (app *register) sidebar(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(240), Height: ui.Grow(), Padding: ui.Pad(8), Gap: 2}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, func() {
		context.Box(ui.Layout{Padding: ui.PadXY(8, 6)}, func() { context.Caption(fmt.Sprintf("LIGHTHOUSES · %d", len(app.lighthouses))) })
		for index, entry := range app.lighthouses {
			icon := "lamp"
			if !entry.lit {
				icon = "lamp-wall-down"
			}
			if context.Selectable(icon, entry.name+"##list", index == app.selected) {
				app.selected = index
			}
		}
		context.Spacer()
		context.Separator()
		context.Selectable("settings", "Settings", false)
	})
}

// viewport leaves the middle transparent, so the scene shows through and the camera gets the
// mouse; a small chip names the view.
func (app *register) viewport(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(12)}, func() {
		chip := ui.Look{Fill: context.Theme.Surface.WithAlpha(0.92), Border: context.Theme.Line, BorderWidth: 1, Radius: 13}
		context.Frame(ui.Layout{Height: ui.Fixed(26), Padding: ui.PadXY(10, 0), Gap: 6, AlignY: ui.Center}, chip, func() {
			context.IconSized("box", 14, context.Theme.InkSoft)
			context.Caption(app.view + " · " + app.current().name)
		})
	})
}

func (app *register) inspector(context *ui.Context) {
	entry := app.current()
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(280), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 10}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, func() {
		context.Heading(entry.name)
		context.Caption(entry.place)
		context.Separator()
		property(context, "Built", fmt.Sprint(entry.built))
		property(context, "Tower height", fmt.Sprintf("%d m", entry.height))
		property(context, "Focal height", fmt.Sprintf("%d m", entry.focalHeight))
		property(context, "Character", entry.character)
		context.Separator()
		context.Row(func() {
			context.Icon("lightbulb")
			context.Label("Light")
			context.Spacer()
			state, colour := "Off", context.Theme.InkQuiet
			if entry.lit {
				state, colour = "On", context.Theme.Positive
			}
			context.Text(state, ui.TextStyle{Color: colour})
		})
		context.Row(func() {
			context.Icon("radar")
			context.Label("Range")
			context.Spacer()
			context.Label(fmt.Sprintf("%d nm", entry.rangeNM))
			if context.IconButton("minus", "") && entry.rangeNM > 0 {
				entry.rangeNM--
				app.saved = false
			}
			if context.IconButton("plus", "") && entry.rangeNM < 30 {
				entry.rangeNM++
				app.saved = false
			}
		})
		context.Spacer()
		context.Row(func() {
			if context.IconButton("trash-2", "Remove") {
				app.removeSelected()
			}
			context.Spacer()
			label := "Switch off"
			if !entry.lit {
				label = "Switch on"
			}
			if context.PrimaryButton(label + "##switch") {
				entry.lit = !entry.lit
				app.saved = false
			}
		})
	})
}

// property is one line of the inspector: a name on the left, its value on the right.
func property(context *ui.Context, name, value string) {
	context.Row(func() {
		context.Text(name, ui.TextStyle{Color: context.Theme.InkSoft})
		context.Spacer()
		context.Label(value)
	})
}

func (app *register) statusBar(context *ui.Context) {
	bar(context, 26, 12, func() {
		state := "All changes saved"
		if !app.saved {
			state = "Unsaved changes"
		}
		context.Caption(state)
		context.Spacer()
		lit := 0
		for _, entry := range app.lighthouses {
			if entry.lit {
				lit++
			}
		}
		context.Caption(fmt.Sprintf("%d of %d lit · Tool: %s", lit, len(app.lighthouses), app.tool))
	})
}

func (app *register) addLighthouse() {
	number := len(app.lighthouses) + 1
	app.lighthouses = append(app.lighthouses, &lighthouse{
		name: fmt.Sprintf("Lighthouse %d", number), place: "Not placed yet", built: 2000,
		height: 10, focalHeight: 10, character: "Fl W 5s", rangeNM: 10, band: graphics.Vector{0.8, 0.6, 0.1, 1},
	})
	app.selected = len(app.lighthouses) - 1
	app.saved = false
}

func (app *register) removeSelected() {
	if len(app.lighthouses) <= 1 {
		return
	}
	app.lighthouses = append(app.lighthouses[:app.selected], app.lighthouses[app.selected+1:]...)
	app.selected = min(app.selected, len(app.lighthouses)-1)
	app.saved = false
}
