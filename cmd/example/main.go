// Command example shows go-ui over a small go-renderer scene: a lighthouse on an island, a panel
// to look after its light, a docked tool panel and a status bar.
//
//	go run ./cmd/example
//	go run ./cmd/example -capture example.png   (renders one frame to a PNG)
//	go run ./cmd/example -dark                  (the dark theme)
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/scene"
	"github.com/akrck02/go-renderer/viewer"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/internal/island"
)

// lighthouse is the program's own data: the interface only shows and changes it.
type lighthouse struct {
	name     string
	on       bool
	colour   int
	rangeNM  float32
	period   float32
	foghorn  bool
	keepers  []string
	keeper   int
	switches int
}

var colours = []string{"White", "Amber", "Red", "Green"}

func main() {
	options := viewer.DefaultOptions()
	options.RegisterFlags(flag.CommandLine)
	dark := flag.Bool("dark", false, "use the dark theme")
	advanced := flag.Bool("advanced", false, "start on the Advanced tab")
	demoOpen := flag.Bool("demo-open", false, "open the light colour list on the second step (for captures)")
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
	light := &lighthouse{name: "North Point", on: true, rangeNM: 18, period: 5,
		keepers: []string{"First keeper", "Second keeper", "Relief keeper", "Boatman", "Cook", "Signal officer", "Visiting engineer"}}
	tool, tab, steps := "Select", 0, 0
	if *advanced {
		tab = 1
	}

	update := func(world *scene.Scene, state *input.State, step float64) {
		steps++
		if *demoOpen {
			clickColourField(interfaceContext, state, steps)
		}
		interfaceContext.Begin(state)
		propertiesPanel(interfaceContext, light, &tab)
		toolsPanel(interfaceContext, &tool)
		statusBar(interfaceContext, light, tool)
		interfaceContext.End()
		island.ShowLamp(world, light.on)
	}
	if err := viewer.Run(world, options, "go-ui example", update); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// clickColourField clicks the light colour field on the second step, as a person would, so a
// capture shows the open list. It needs the field's place, known after the first step.
func clickColourField(interfaceContext *ui.Context, state *input.State, steps int) {
	if steps != 2 {
		return
	}
	field, found := interfaceContext.Area("Properties", "Light colour")
	if !found {
		return
	}
	state.Feed().MoveCursor(float64(field.X+field.Width/2), float64(field.Y+field.Height/2))
	state.Feed().SetButton(input.MouseLeft, true)
	state.Feed().SetButton(input.MouseLeft, false)
}

// propertiesPanel is the panel on the cover of go-ui's design: tabs, a select, a checkbox, a
// slider, a number field and buttons; the second tab has a switch, a scrolling list and a grid.
func propertiesPanel(interfaceContext *ui.Context, light *lighthouse, tab *int) {
	interfaceContext.PanelWith("Properties", ui.PanelOptions{X: 24, Y: 24, Width: ui.Fixed(300)}, func() {
		interfaceContext.Tabs("view", tab, "General", "Advanced")
		if *tab == 0 {
			generalTab(interfaceContext, light)
		} else {
			advancedTab(interfaceContext, light)
		}
		interfaceContext.Separator()
		interfaceContext.Row(func() {
			interfaceContext.Caption(fmt.Sprintf("Switched %d times", light.switches))
			interfaceContext.Spacer()
			if interfaceContext.Button("Reset") {
				light.rangeNM, light.period = 18, 5
			}
			interfaceContext.Tooltip("Back to 18 nm and 5 s")
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

func generalTab(interfaceContext *ui.Context, light *lighthouse) {
	interfaceContext.Heading(light.name)
	interfaceContext.Caption("Name")
	interfaceContext.TextFieldWith("Name", &light.name, ui.TextFieldOptions{Placeholder: "Lighthouse name", MaxLength: 32})
	interfaceContext.Caption("Light colour")
	interfaceContext.Select("Light colour", &light.colour, colours)
	interfaceContext.Checkbox("Lit at night", &light.on)
	interfaceContext.Row(func() {
		interfaceContext.Caption("Range")
		interfaceContext.Spacer()
		interfaceContext.Caption(fmt.Sprintf("%.0f nm", light.rangeNM))
	})
	interfaceContext.SliderStep("range", &light.rangeNM, 0, 30, 1)
	interfaceContext.Caption("Flash period")
	interfaceContext.NumberField("period", &light.period, 1, 30, 0.5, "%.1f s")
}

func advancedTab(interfaceContext *ui.Context, light *lighthouse) {
	interfaceContext.Toggle("Fog horn", &light.foghorn)
	interfaceContext.Caption("Crew")
	interfaceContext.ListBox("crew", ui.Layout{Width: ui.Grow(), Height: ui.Fixed(120)}, light.keepers, &light.keeper)
	interfaceContext.Caption("Signals")
	interfaceContext.Grid(4, func() {
		for _, icon := range []string{"sun", "moon", "cloud-fog", "wind", "waves", "anchor", "ship", "radio"} {
			interfaceContext.IconButton(icon, "")
			interfaceContext.Tooltip(icon)
		}
	})
}

func toolsPanel(interfaceContext *ui.Context, tool *string) {
	interfaceContext.PanelWith("Tools", ui.PanelOptions{Dock: ui.DockRight, Width: ui.Fixed(180)}, func() {
		for _, name := range []struct{ icon, label string }{
			{"mouse-pointer-2", "Select"}, {"move", "Move"}, {"ruler", "Measure"}, {"map-pin", "Mark"},
		} {
			interfaceContext.Box(ui.Layout{Width: ui.Grow()}, func() {
				if interfaceContext.IconButton(name.icon, name.label) {
					*tool = name.label
				}
			})
		}
	})
}

// statusBar sits at the bottom of the screen, outside any panel: a spacer pushes it down.
func statusBar(interfaceContext *ui.Context, light *lighthouse, tool string) {
	theme := interfaceContext.Theme
	interfaceContext.Spacer()
	look := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1}
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(12, 6), Gap: 8, AlignY: ui.Center}
	interfaceContext.Frame(layout, look, func() {
		interfaceContext.IconSized("mouse-pointer-2", 14, theme.InkSoft)
		interfaceContext.Caption("Tool: " + tool)
		interfaceContext.Separator()
		state := "light off"
		if light.on {
			state = fmt.Sprintf("light on · %.0f nm · %s", light.rangeNM, colours[light.colour])
		}
		interfaceContext.Caption(light.name + " · " + state)
	})
}
