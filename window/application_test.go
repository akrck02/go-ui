package window

import (
	"hash/maphash"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/scene"
	ui "github.com/akrck02/go-ui"
)

func TestTheInterfaceApplicationBuildsOneStepAtATime(t *testing.T) {
	context, err := ui.New()
	if err != nil {
		t.Fatal(err)
	}
	builds := 0
	shown := shownFrame{seed: maphash.MakeSeed()}
	app := interfaceApplication(context, Options{Title: "Lighthouse"}, func(context *ui.Context) {
		builds++
		context.Label("Hello")
	}, &shown, &shownScene{})
	if !app.DrawOnDemand || !app.FlatOnly || app.Title != "Lighthouse" {
		t.Fatalf("draw on demand %v, flat %v, title %q", app.DrawOnDemand, app.FlatOnly, app.Title)
	}
	app.Input.FramebufferWidth, app.Input.FramebufferHeight = 800, 600
	app.Input.WindowWidth, app.Input.WindowHeight = 400, 300
	if err := app.Update(app, 1.0/60); err != nil {
		t.Fatal(err)
	}
	if builds != 1 || context.Drawing().IsEmpty() {
		t.Fatalf("builds %d, commands %d: one step should build and paint once", builds, len(context.Drawing().Commands))
	}
	if app.Unchanged(app) {
		t.Fatal("the first frame is always drawn")
	}
	if !app.Unchanged(app) {
		t.Fatal("the same interface at the same size should not be drawn again")
	}
}

func TestContinuousWindowsDrawEveryFrame(t *testing.T) {
	context, err := ui.New()
	if err != nil {
		t.Fatal(err)
	}
	shown := shownFrame{seed: maphash.MakeSeed()}
	app := interfaceApplication(context, Options{Continuous: true}, func(*ui.Context) {}, &shown, &shownScene{})
	if app.DrawOnDemand {
		t.Fatal("a continuous window should not draw on demand")
	}
}

func TestASceneIsAskedForWithTheUnclaimedInputAndDrawnUnderTheInterface(t *testing.T) {
	context, err := ui.New()
	if err != nil {
		t.Fatal(err)
	}
	world := scene.New()
	var sawClaimed bool
	options := Options{Scene: func(state *input.State, step float64) *SceneView {
		sawClaimed = state.MouseClaimed()
		state.RequestFrame(0) // the camera is still moving
		return &SceneView{World: world}
	}}
	shown, view := shownFrame{seed: maphash.MakeSeed()}, &shownScene{}
	app := interfaceApplication(context, options, func(context *ui.Context) { context.Label("Over the scene") }, &shown, view)
	if app.FlatOnly {
		t.Fatal("a window with a scene needs a depth buffer")
	}
	app.Input.WindowWidth, app.Input.WindowHeight = 400, 300
	if err := app.Update(app, 1.0/60); err != nil {
		t.Fatal(err)
	}
	if view.view == nil || world.Overlay.Drawing != context.Drawing() {
		t.Fatal("the scene should be shown with the interface painted over it")
	}
	if sawClaimed {
		t.Fatal("the scene got input the interface had claimed")
	}
	if _, wanted := app.Input.TakeFrameRequest(); !wanted {
		t.Fatal("the frame the scene asked for was lost")
	}
	if app.Unchanged(app) {
		t.Fatal("a frame with a scene is always drawn when asked for")
	}
}

func TestTheProgramsDrawingIsPaintedOverTheSceneAndUnderTheInterface(t *testing.T) {
	context, err := ui.New()
	if err != nil {
		t.Fatal(err)
	}
	world := scene.New()
	var own draw2d.List
	options := Options{Scene: func(state *input.State, step float64) *SceneView {
		own.Reset()
		own.Rectangle(draw2d.Rect{Width: 100, Height: 10}, 0, draw2d.Hex("#ff0000"), draw2d.Transparent, 0)
		return &SceneView{World: world, Drawing: &own}
	}}
	shown, view := shownFrame{seed: maphash.MakeSeed()}, &shownScene{}
	app := interfaceApplication(context, options, func(context *ui.Context) { context.Label("Over the bar") }, &shown, view)
	app.Input.WindowWidth, app.Input.WindowHeight = 400, 300
	if err := app.Update(app, 1.0/60); err != nil {
		t.Fatal(err)
	}
	painted := world.Overlay.Drawing.Commands
	interfaceCommands := len(context.Drawing().Commands)
	if len(painted) != 1+interfaceCommands || painted[0].Area.Width != 100 {
		t.Fatalf("over the scene go %d commands, the program's first; want 1 and the interface's %d", len(painted), interfaceCommands)
	}
}

func TestAScenesCapturedPointerReachesTheWindow(t *testing.T) {
	context, err := ui.New()
	if err != nil {
		t.Fatal(err)
	}
	world, capture := scene.New(), true
	options := Options{Scene: func(state *input.State, step float64) *SceneView {
		state.LockCursor(capture)
		return &SceneView{World: world}
	}}
	shown, view := shownFrame{seed: maphash.MakeSeed()}, &shownScene{}
	app := interfaceApplication(context, options, func(context *ui.Context) {}, &shown, view)
	if err := app.Update(app, 1.0/60); err != nil || !app.Input.CursorLocked() {
		t.Fatalf("a scene that captures the pointer leaves the window's %v, %v", app.Input.CursorLocked(), err)
	}
	capture = false
	if err := app.Update(app, 1.0/60); err != nil || app.Input.CursorLocked() {
		t.Fatal("a scene that lets the pointer go leaves it captured")
	}
}

func TestQuittingStopsTheWindowAfterTheFrame(t *testing.T) {
	context, err := ui.New()
	if err != nil {
		t.Fatal(err)
	}
	shown, view := shownFrame{seed: maphash.MakeSeed()}, &shownScene{}
	app := interfaceApplication(context, Options{}, func(context *ui.Context) { Quit() }, &shown, view)
	if err := app.Update(app, 1.0/60); err != nil || !app.Stop {
		t.Fatalf("a window asked to quit keeps running (%v)", err)
	}
	if takeQuit() {
		t.Fatal("the quit asked for is not forgotten once taken")
	}
}
