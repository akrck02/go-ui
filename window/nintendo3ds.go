//go:build nintendo3ds

package window

import (
	"math"
	"runtime"
	"time"

	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// On the Nintendo 3DS (homebrew, built with TinyGo's nintendo3ds target) the window is the
// console's two screens, drawn through go-renderer's ctr: the scene on the top one, in 3D with the
// slider, and the interface on the bottom one, the touch screen, whose touches are the pointer.
// SELECT swaps them: the scene goes below, where it can be touched, and the interface above, to be
// read. START ends the program. The cross is the arrow keys, L and R minus and plus, A Enter and B
// Escape (ctr.Controls).

// buttonsPressed are the console's buttons pressed this frame (ButtonsPressed).
var buttonsPressed ctr.Buttons

// ButtonsPressed are the console's buttons pressed this frame, for a program that answers them
// itself (A to pick, B to go back) whatever the interface does with the keys they stand for.
// Only on the 3DS.
func ButtonsPressed() ctr.Buttons { return buttonsPressed }

// Wake: the console draws every frame, so there is nothing to wake.
func Wake() {}

// Graphics names the console's GPU.
func Graphics() string { return "PICA200, through citro3d" }

// LastCrash: the console keeps no crash reports for a program.
func LastCrash() string { return "" }

// SystemPrefersDark: the console has no dark theme to follow.
func SystemPrefersDark() bool { return false }

// OpenLink, Share, Notify and the rest have nothing to go to on the console.
func OpenLink(link string) bool              { return false }
func Share(title, text string) bool          { return false }
func Notify(id int, title, text string) bool { return false }
func FeelLongPress()                         {}
func Feel(Feeling)                           {}
func NotificationsAllowed() bool             { return false }
func AskForNotifications()                   {}
func SystemPalettes() (Palettes, bool)       { return Palettes{}, false }

// separationAtFullSlider is how far apart the eyes stand with the 3D slider at its top, as a share
// of the distance to what the camera looks at.
const separationAtFullSlider = 0.03

// RunWith is Run with an interface context of your own. It returns when START is pressed or the
// system asks the program to close (the HOME menu, the power button).
func RunWith(context *ui.Context, options Options, build func(context *ui.Context)) error {
	top, err := ctr.Open()
	if err != nil {
		return err
	}
	defer top.Close()
	bottom, err := ctr.OpenBottomScreen()
	if err != nil {
		return err
	}
	flat, err := ctr.NewFlatRenderer(bottom)
	if err != nil {
		return err
	}
	run := consoleRun{context: context, options: options, build: build, top: top, bottom: bottom, flat: flat, state: input.NewState()}
	if options.Scene != nil {
		if run.scene, err = ctr.NewSceneRenderer(top); err != nil {
			return err
		}
	}
	run.loop()
	if options.Hidden != nil {
		options.Hidden()
	}
	return nil
}

// consoleRun is a program running on the console's two screens.
type consoleRun struct {
	context  *ui.Context
	options  Options
	build    func(context *ui.Context)
	top      *ctr.Screen
	bottom   *ctr.BottomScreen
	flat     *ctr.FlatRenderer
	scene    *ctr.SceneRenderer
	state    *input.State
	controls ctr.Controls
	swapped  bool // the scene below and the interface above
	// forScene and keysOnly are the input the scene and the interface get, kept here: made in
	// step, they would go to the heap every frame
	forScene, keysOnly input.State
}

func (run *consoleRun) loop() {
	last := time.Now()
	for ctr.Running() {
		reading := ctr.ReadControls()
		pressed := run.controls.Pressed(reading)
		if pressed&ctr.ButtonStart != 0 && !run.options.KeepStart || takeQuit() {
			return
		}
		if pressed&ctr.ButtonSelect != 0 {
			run.swapped = !run.swapped
		}
		run.controls.Apply(run.state, reading)
		buttonsPressed = pressed
		now := time.Now()
		step := now.Sub(last).Seconds()
		last = now
		view := run.step(step)
		run.draw(view)
		run.state.EndFrame()
		// TinyGo's scheduler on the console is cooperative, and waiting for the screen happens
		// in C, which does not yield: without this the program's other goroutines (reads in the
		// background) would never run.
		runtime.Gosched()
	}
}

// step builds the interface and asks the program for its scene. In the usual layout the touches
// are the interface's, and the scene gets the input the interface did not claim (the keys); with
// the screens swapped the touches are the scene's, and the interface, shown above, only the keys.
func (run *consoleRun) step(step float64) *SceneView {
	interfaceWidth, sceneWidth := ctr.BottomWidth, ctr.ScreenWidth
	if run.swapped {
		interfaceWidth, sceneWidth = ctr.ScreenWidth, ctr.BottomWidth
	}
	shaped := func(state *input.State, width int) {
		state.WindowWidth, state.WindowHeight = width, ctr.BottomHeight
		state.FramebufferWidth, state.FramebufferHeight = width, ctr.BottomHeight
	}
	shaped(run.state, interfaceWidth)
	forInterface := run.state
	if run.swapped {
		run.keysOnly = run.state.Unclaimed()
		keysOnly := &run.keysOnly
		keysOnly.SetButton(input.MouseLeft, false)
		keysOnly.CursorX, keysOnly.CursorY, keysOnly.CursorDeltaX, keysOnly.CursorDeltaY = -1e6, -1e6, 0, 0
		keysOnly.ClearTouches()
		forInterface = keysOnly
	}
	run.context.Begin(forInterface)
	run.build(run.context)
	run.context.End()
	if run.options.Scene == nil {
		return nil
	}
	run.forScene = run.state.Unclaimed()
	forScene := &run.forScene
	if !run.swapped {
		// the interface has the touch screen; the scene turns with the keys, and its pointer is
		// off it (no hand to show over what is under the finger on the other screen)
		forScene.SetButton(input.MouseLeft, false)
		forScene.ClearTouches()
		forScene.CursorX, forScene.CursorY, forScene.CursorDeltaX, forScene.CursorDeltaY = -1e6, -1e6, 0, 0
	}
	shaped(forScene, sceneWidth)
	return run.options.Scene(forScene, step)
}

// FrameTimes are how long the parts of the last frame took on the console, for a program that
// shows its own resources: waiting for the GPU to finish the frame before (the frame's GPU time,
// when it is the bottleneck), painting the scene, the scene's own drawing over it and the
// interface (the processor's work of making their draw calls), and handing the frame over.
type FrameTimes struct {
	WaitForGPU, Scene, OverScene, Interface, Submit time.Duration
}

var frameTimes FrameTimes

// LastFrameTimes are the parts of the last frame drawn.
func LastFrameTimes() FrameTimes { return frameTimes }

// draw shows the frame: the scene and the interface, each on its screen.
func (run *consoleRun) draw(view *SceneView) {
	drawing, background := run.context.Drawing(), run.context.Theme.Background
	times := &frameTimes
	mark := time.Now()
	ctr.BeginFrame()
	lap(&times.WaitForGPU, &mark)
	switch {
	case view == nil:
		run.top.Clear(packedBackground(background))
		times.Scene, times.OverScene = 0, 0
		run.flat.Paint(drawing, background)
		lap(&times.Interface, &mark)
	case run.swapped:
		run.flat.PaintOn(run.top, drawing, background)
		lap(&times.Interface, &mark)
		run.scene.PaintOn(run.bottom, view.World, view.Camera)
		lap(&times.Scene, &mark)
		times.OverScene = 0
	default:
		focus := distance(view.Camera.Position, view.Camera.Target)
		separation := ctr.Slider3D() * focus * separationAtFullSlider
		run.scene.PaintStereo(view.World, view.Camera, separation, focus)
		lap(&times.Scene, &mark)
		run.paintOverScene(view, separation > 0)
		lap(&times.OverScene, &mark)
		run.paintBottom(view.Below, drawing, background)
		lap(&times.Interface, &mark)
	}
	ctr.EndFrame()
	lap(&times.Submit, &mark)
}

// paintBottom paints the bottom screen: the interface alone, or over a scene of its own, drawn by
// the top screen's renderer as the swapped layout does.
func (run *consoleRun) paintBottom(below *BelowView, drawing *draw2d.List, background draw2d.Color) {
	if below == nil {
		run.flat.Paint(drawing, background)
		return
	}
	run.scene.PaintOn(run.bottom, below.World, below.Camera)
	run.flat.PaintOver(run.bottom, drawing)
}

// paintOverScene paints the program's drawing over the scene (SceneView.Drawing, or the older
// World.Overlay.Drawing: a sight, a name) on the top screen, in each eye when in 3D.
func (run *consoleRun) paintOverScene(view *SceneView, stereo bool) {
	over := view.World.Overlay.Drawing
	if !view.Drawing.IsEmpty() {
		over = view.Drawing
	}
	if over.IsEmpty() {
		return
	}
	if !stereo {
		run.flat.PaintOver(run.top, over)
		return
	}
	for _, eye := range []ctr.Eye{ctr.LeftEye, ctr.RightEye} {
		run.top.DrawOn(eye)
		run.flat.PaintOver(run.top, over)
	}
}

// lap puts the time since mark into a frame time, and starts the next lap now.
func lap(into *time.Duration, mark *time.Time) {
	now := time.Now()
	*into, *mark = now.Sub(*mark), now
}

func distance(from, to [4]float32) float32 {
	x, y, z := to[0]-from[0], to[1]-from[1], to[2]-from[2]
	return float32(math.Sqrt(float64(x*x + y*y + z*z)))
}

// packedBackground packs a theme colour as 0xRRGGBBAA, to clear the screen with.
func packedBackground(colour [4]float32) uint32 {
	channel := func(value float32) uint32 { return uint32(min(max(value, 0), 1)*255 + 0.5) }
	return channel(colour[0])<<24 | channel(colour[1])<<16 | channel(colour[2])<<8 | channel(colour[3])
}
