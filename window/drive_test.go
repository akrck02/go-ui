package window

import (
	"hash/maphash"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// enterPresser is a made-up driver: it presses Enter at every frame, and counts the drawings it
// was shown.
type enterPresser struct{ shown int }

func (driver *enterPresser) Drive(state *input.State, _ *ui.Context) {
	state.SetKey(input.KeyEnter, true)
}

func (driver *enterPresser) Show(drawing *draw2d.List) {
	if !drawing.IsEmpty() {
		driver.shown++
	}
}

// lampStep makes the application of a lighthouse's window, whose interface counts the frames it
// saw Enter pressed in.
func lampStep(options Options, enters *int) func() {
	context := ui.NewWithMeasurer(uitest.FixedMeasurer{})
	shown := shownFrame{seed: maphash.MakeSeed()}
	var state *input.State
	app := interfaceApplication(context, options, func(context *ui.Context) {
		context.Label("Lamp")
		if state.Pressed(input.KeyEnter) {
			*enters++
		}
	}, &shown, &shownScene{})
	state = app.Input
	app.Input.WindowWidth, app.Input.WindowHeight = 400, 300
	app.Input.FramebufferWidth, app.Input.FramebufferHeight = 400, 300
	return func() {
		_ = app.Update(app, 1.0/60) // a flat interface's step returns no error
		app.Input.EndFrame()
	}
}

func TestADriverAddsInputBeforeTheInterfaceRuns(test *testing.T) {
	driver := &enterPresser{}
	enters := 0
	step := lampStep(Options{Driver: driver}, &enters)
	step()
	if enters != 1 || driver.shown != 1 {
		test.Fatalf("the interface saw Enter %d times and the driver was shown %d drawings; want 1 and 1", enters, driver.shown)
	}
}

func TestAWindowWithoutADriverAllocatesNothingForIt(test *testing.T) {
	enters := 0
	step := lampStep(Options{}, &enters)
	step() // the first frame makes what the interface keeps
	if allocations := testing.AllocsPerRun(100, step); allocations != 0 {
		test.Fatalf("a window step with no driver allocates %v times", allocations)
	}
}

func BenchmarkAWindowStepWithoutADriver(benchmark *testing.B) {
	enters := 0
	step := lampStep(Options{}, &enters)
	benchmark.ReportAllocs()
	for benchmark.Loop() {
		step()
	}
}

func BenchmarkAWindowStepWithADriver(benchmark *testing.B) {
	enters := 0
	step := lampStep(Options{Driver: &enterPresser{}}, &enters)
	benchmark.ReportAllocs()
	for benchmark.Loop() {
		step()
	}
}
