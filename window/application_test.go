package window

import (
	"hash/maphash"
	"testing"

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
	}, &shown)
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
	app := interfaceApplication(context, Options{Continuous: true}, func(*ui.Context) {}, &shown)
	if app.DrawOnDemand {
		t.Fatal("a continuous window should not draw on demand")
	}
}
