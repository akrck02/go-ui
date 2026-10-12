package main

import (
	"testing"

	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/scene"
)

func newEditor() *editor {
	return &editor{world: &scene.Scene{}, selected: -1, placing: true, size: 4}
}

// click presses and releases the left button, moving the pointer by distance in between.
func click(app *editor, state *input.State, x, y, distance float64, onLand bool, point graphics.Vec4) {
	state.MoveCursor(x, y)
	state.SetButton(input.MouseLeft, true)
	app.handlePointer(state, point, onLand)
	state.EndFrame()
	state.MoveCursor(x+distance, y)
	state.SetButton(input.MouseLeft, false)
	app.handlePointer(state, point, onLand)
	state.EndFrame()
}

func TestAClickOnLandPlacesAHut(t *testing.T) {
	app, state := newEditor(), input.NewState()
	click(app, state, 300, 200, 0, true, graphics.Vec4{5, 2, 7, 1})
	if len(app.huts) != 1 || app.selected != 0 {
		t.Fatalf("%d huts, selected %d; want one, selected", len(app.huts), app.selected)
	}
	position := app.huts[0].node.Transform
	if position[12] != 5 || position[13] != 2 || position[14] != 7 {
		t.Errorf("the hut stands at %v, want where the click was", position[12:15])
	}
	if len(app.world.Nodes) != 1 || len(app.world.Overlay.Labels) != 1 {
		t.Error("the hut and its label belong in the scene")
	}
}

func TestADragMovesTheMapAndPlacesNothing(t *testing.T) {
	app, state := newEditor(), input.NewState()
	click(app, state, 300, 200, 40, true, graphics.Vec4{5, 2, 7, 1})
	if len(app.huts) != 0 {
		t.Error("a drag must not place a hut")
	}
}

func TestAClickOnTheWaterPlacesNothing(t *testing.T) {
	app, state := newEditor(), input.NewState()
	click(app, state, 300, 200, 0, false, graphics.Vec4{})
	if len(app.huts) != 0 {
		t.Error("a hut cannot stand in the sea")
	}
}

func TestSelectingPicksTheNearestHutAndRemovingDropsIt(t *testing.T) {
	app, state := newEditor(), input.NewState()
	click(app, state, 100, 100, 0, true, graphics.Vec4{0, 0, 0, 1})
	click(app, state, 200, 100, 0, true, graphics.Vec4{20, 0, 0, 1})
	app.placing = false
	click(app, state, 100, 100, 0, true, graphics.Vec4{1, 0, 1, 1})
	if app.selected != 0 {
		t.Fatalf("selected %d, want the first hut, next to the click", app.selected)
	}
	if app.huts[0].material.BaseColor != selectedColour || app.huts[1].material.BaseColor != hutColour {
		t.Error("only the selected hut takes the selection colour")
	}
	app.removeSelected()
	if len(app.huts) != 1 || len(app.world.Nodes) != 1 || len(app.world.Overlay.Labels) != 1 {
		t.Error("removing a hut takes it, its node and its label away")
	}
	click(app, state, 100, 100, 0, true, graphics.Vec4{50, 0, 50, 1})
	if app.selected != -1 {
		t.Error("a click far from every hut selects none")
	}
}
