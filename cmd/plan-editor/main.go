// Command plan-editor is a small editor over a 3D scene: it places huts on a made-up island by
// clicking on the map, and a panel lists them to select, rename and remove. It shows how a program
// edits the scene it views with go-renderer's viewer.Edit: the pointer is turned into a point on
// the ground, a click (a press and a release without moving) places or selects, a drag still moves
// the map, and the window draws only when something happens.
//
//	go run ./cmd/plan-editor
//	go run ./cmd/plan-editor -isometric
//	go run ./cmd/plan-editor -capture editor.png
package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/scene"
	"github.com/akrck02/go-renderer/scene/sample"
	"github.com/akrck02/go-renderer/viewer"
	ui "github.com/akrck02/go-ui"
)

// clickDistance is how far, in window pixels, the pointer may move between a press and its release
// for them to count as a click rather than a drag of the map.
const clickDistance = 5

// hut is one placed hut: the program's data, and the node that shows it.
type hut struct {
	name     string
	size     float32
	node     *scene.Node
	label    *scene.Label
	material *scene.Material
}

// editor is the program's state.
type editor struct {
	world    *scene.Scene
	huts     []*hut
	selected int // -1 = none
	placing  bool
	size     float32
	named    int // huts named so far, for the next default name
	// the press that may become a click, and whether the pointer was on the panel then
	pressX, pressY float64
	pressOnPanel   bool
	pointer        graphics.Vector
	pointerOnLand  bool
}

func main() {
	options := viewer.DefaultOptions()
	options.RegisterFlags(flag.CommandLine)
	dark := flag.Bool("dark", false, "use the dark theme")
	flag.Parse()
	options.Start.Map = !options.Start.Walking

	context, err := ui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *dark {
		context.Theme = ui.DarkTheme()
	}
	world := sample.Island()
	world.Overlay.Drawing = context.Drawing()
	app := &editor{world: world, selected: -1, placing: true, size: 4}
	if height, found := world.GroundHeight(14, 26); found {
		app.place(graphics.Vector{14, height, 26, 1}) // one to start with, by the square
	}

	edit := func(world *scene.Scene, state *input.State, step float64, view viewer.View) {
		context.Begin(state)
		app.panel(context)
		context.End()
		app.followPointer(state, view)
		app.handleKeys(state)
	}
	if err := viewer.Edit(world, options, "Plan editor", edit); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// followPointer finds the ground under the pointer and turns a click on the scene into placing or
// selecting a hut. The panel claims the mouse when the pointer is over it (go-ui's End).
func (app *editor) followPointer(state *input.State, view viewer.View) {
	point, onLand := view.PointOnGround(state.CursorX, state.CursorY)
	app.handlePointer(state, point, onLand)
}

// handlePointer places or selects on a click at a point of the ground (onLand false: the pointer
// is over the water or off the island).
func (app *editor) handlePointer(state *input.State, point graphics.Vector, onLand bool) {
	app.pointer, app.pointerOnLand = point, onLand
	if state.ButtonPressed(input.MouseLeft) {
		app.pressX, app.pressY, app.pressOnPanel = state.CursorX, state.CursorY, state.MouseClaimed()
	}
	if !state.ButtonReleased(input.MouseLeft) || app.pressOnPanel {
		return
	}
	if math.Hypot(state.CursorX-app.pressX, state.CursorY-app.pressY) > clickDistance {
		return // a drag: the map moved
	}
	if !app.pointerOnLand {
		return
	}
	if app.placing {
		app.place(app.pointer)
		return
	}
	app.selectNearest(app.pointer)
}

// handleKeys removes the selected hut with Delete or Backspace, unless a field is being typed in.
func (app *editor) handleKeys(state *input.State) {
	if state.KeyboardClaimed() {
		return
	}
	if state.Pressed(input.KeyDelete) || state.Pressed(input.KeyBackspace) {
		app.removeSelected()
	}
}

// place puts a new hut at a point of the ground and selects it.
func (app *editor) place(point graphics.Vector) {
	app.named++
	material := &scene.Material{Name: "hut", BaseColor: hutColour}
	half := app.size / 2
	node := &scene.Node{Name: "hut", Mesh: boxMesh(half, app.size*0.75, material),
		Transform: graphics.Translate(graphics.Vector{point[0], point[1], point[2], 0})}
	entry := &hut{name: fmt.Sprintf("Hut %d", app.named), size: app.size, node: node, material: material}
	entry.label = &scene.Label{Node: node, Height: app.size, Text: entry.name, Icon: "house"}
	app.world.Nodes = append(app.world.Nodes, node)
	app.world.Overlay.Labels = append(app.world.Overlay.Labels, entry.label)
	app.huts = append(app.huts, entry)
	app.choose(len(app.huts) - 1)
}

// selectNearest selects the hut closest to a point, if one is within its own size of it.
func (app *editor) selectNearest(point graphics.Vector) {
	nearest, best := -1, math.Inf(1)
	for index, entry := range app.huts {
		position := entry.node.Transform
		distance := math.Hypot(float64(position[12]-point[0]), float64(position[14]-point[2]))
		if distance < best && distance <= float64(entry.size) {
			nearest, best = index, distance
		}
	}
	app.choose(nearest)
}

// choose selects a hut (or none, with -1) and shows which one by its colour.
func (app *editor) choose(index int) {
	app.selected = index
	for position, entry := range app.huts {
		entry.material.BaseColor = hutColour
		if position == index {
			entry.material.BaseColor = selectedColour
		}
	}
}

func (app *editor) removeSelected() {
	if app.selected < 0 {
		return
	}
	removed := app.huts[app.selected]
	app.world.Nodes = without(app.world.Nodes, removed.node)
	app.world.Overlay.Labels = without(app.world.Overlay.Labels, removed.label)
	app.huts = append(app.huts[:app.selected], app.huts[app.selected+1:]...)
	app.choose(min(app.selected, len(app.huts)-1))
}

// without returns the list without one element.
func without[T comparable](list []T, unwanted T) []T {
	kept := list[:0]
	for _, element := range list {
		if element != unwanted {
			kept = append(kept, element)
		}
	}
	return kept
}

var (
	hutColour      = graphics.Vector{0.72, 0.52, 0.34, 1}
	selectedColour = graphics.Vector{0.95, 0.62, 0.2, 1}
)

// panel is the editor's window: the tools on the right, over the map.
func (app *editor) panel(context *ui.Context) {
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(12)}, func() {
		context.Spacer()
		layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(260), Height: ui.Grow(), Padding: ui.Pad(14), Gap: 8}
		look := ui.Look{Fill: context.Theme.Surface, Border: context.Theme.Line, BorderWidth: 1, Radius: 10}
		context.Frame(layout, look, func() {
			context.Heading("Huts")
			context.Row(func() {
				if context.Selectable("plus", "Place##tool", app.placing) {
					app.placing = true
				}
				if context.Selectable("mouse-pointer-2", "Select##tool", !app.placing) {
					app.placing = false
				}
			})
			context.NumberField("Size", &app.size, 2, 10, 0.5, "%.1f m")
			context.Separator()
			context.Caption(fmt.Sprintf("PLACED · %d", len(app.huts)))
			for index, entry := range app.huts {
				if context.Selectable("house", entry.name+fmt.Sprintf("##hut%p", entry), index == app.selected) {
					app.choose(index)
				}
			}
			context.Spacer()
			app.selectedDetails(context)
			context.Separator()
			context.Caption(app.pointerText())
		})
	})
}

// selectedDetails lets the selected hut be renamed or removed.
func (app *editor) selectedDetails(context *ui.Context) {
	if app.selected < 0 {
		context.Caption("Click a hut to select it")
		return
	}
	entry := app.huts[app.selected]
	if context.TextField("Name", &entry.name) {
		entry.label.Text = entry.name
	}
	if context.IconButton("trash-2", "Remove") {
		app.removeSelected()
	}
}

// pointerText says where the pointer is on the island.
func (app *editor) pointerText() string {
	if !app.pointerOnLand {
		return "Pointer: not on land"
	}
	return fmt.Sprintf("Pointer: x %.1f · z %.1f · height %.1f m", app.pointer[0], app.pointer[2], app.pointer[1])
}

// boxMesh is a box with its base centred on the origin: half its width on each side, height tall.
func boxMesh(half, height float32, material *scene.Material) *scene.Mesh {
	corners := [8][3]float32{
		{-half, 0, -half}, {half, 0, -half}, {half, 0, half}, {-half, 0, half},
		{-half, height, -half}, {half, height, -half}, {half, height, half}, {-half, height, half},
	}
	faces := [6][4]int{{3, 2, 6, 7}, {1, 0, 4, 5}, {0, 3, 7, 4}, {2, 1, 5, 6}, {7, 6, 5, 4}, {0, 1, 2, 3}}
	mesh := &scene.Mesh{Name: "hut", Material: material}
	for _, face := range faces {
		for _, corner := range []int{0, 1, 2, 0, 2, 3} {
			point := corners[face[corner]]
			mesh.Positions = append(mesh.Positions, point[0], point[1], point[2])
		}
	}
	mesh.ComputeNormals()
	return mesh
}
