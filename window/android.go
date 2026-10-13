//go:build android

package window

import (
	"hash/maphash"
	"log"
	"sync"
	"time"

	"github.com/akrck02/go-renderer/frameloop"
	"github.com/akrck02/go-renderer/gles"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/models"
	ui "github.com/akrck02/go-ui"
	mobileapp "golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

// Wake makes the window draw a frame now: call it when the program's data changes on another
// goroutine. Safe from any goroutine.
func Wake() {
	running.Lock()
	window := current
	running.Unlock()
	if window != nil {
		window.requestPaint()
	}
}

var (
	running sync.Mutex
	current *androidWindow
)

// RunWith is Run with an interface context of your own (a custom theme or measurer). On Android
// the window is the application's activity, drawn through OpenGL ES with go-renderer's
// gles.FlatRenderer. It never returns: the system ends the program.
func RunWith(context *ui.Context, options Options, build func(context *ui.Context)) error {
	shown := shownFrame{seed: maphash.MakeSeed()}
	view := &shownScene{}
	rate := options.MaxFramesPerSecond
	if rate == 0 {
		rate = defaultPhoneFrameRate
	}
	window := &androidWindow{context: context, shown: &shown, view: view, pacer: newFramePacer(rate),
		app: interfaceApplication(context, options, build, &shown, view)}
	window.clock = frameloop.NewClock(window.app.FixedStep)
	mobileapp.Main(window.run)
	return nil
}

// androidWindow is the state of the activity's event loop.
type androidWindow struct {
	context  *ui.Context
	shown    *shownFrame
	view     *shownScene
	app      *models.Application
	clock    *frameloop.Clock
	events   mobileapp.App
	graphics gl.Context
	renderer *gles.FlatRenderer
	scene    *gles.SceneRenderer
	pointer  touchPointer
	scale    float32
	// insetsKnown is whether the window's insets were read; until it is shown, Android only knows
	// the bars' standard heights, so the window asks again at the next paint
	insetsKnown bool

	mutex        sync.Mutex
	paintPending bool // a paint event is on its way: more requests add nothing
	timer        *time.Timer
	pacer        framePacer // keeps frames apart (MaxFramesPerSecond)
	paced        bool       // a paint waits for the pacer: more requests add nothing
}

func (window *androidWindow) run(events mobileapp.App) {
	window.events = events
	running.Lock()
	current = window
	running.Unlock()
	for event := range events.Events() {
		switch event := events.Filter(event).(type) {
		case lifecycle.Event:
			window.changeStage(event)
		case size.Event:
			window.resize(event)
		case touch.Event:
			window.touched(event)
		case key.Event:
			window.keyPressed(event)
		case paint.Event:
			window.paint(event)
		}
	}
}

// changeStage creates the renderer when the activity becomes visible and lets it go when it is
// hidden: the system destroys the graphics context then, and the program stops drawing.
func (window *androidWindow) changeStage(event lifecycle.Event) {
	switch event.Crosses(lifecycle.StageVisible) {
	case lifecycle.CrossOn:
		window.graphics, _ = event.DrawContext.(gl.Context)
		window.renderer, window.scene = nil, nil
		readSystemDark()
		*window.shown = shownFrame{seed: window.shown.seed} // the new surface is empty: draw again
		window.requestPaint()
	case lifecycle.CrossOff:
		window.graphics, window.renderer, window.scene = nil, nil, nil
	}
}

// resize records the window's size: the framebuffer in pixels, the layout in density-independent
// pixels.
func (window *androidWindow) resize(event size.Event) {
	window.scale = densityScale(event.PixelsPerPt)
	state := window.app.Input
	state.FramebufferWidth, state.FramebufferHeight = event.WidthPx, event.HeightPx
	state.WindowWidth = int(float32(event.WidthPx) / window.scale)
	state.WindowHeight = int(float32(event.HeightPx) / window.scale)
	window.insetsKnown = false
	window.followInsets()
	window.requestPaint()
}

// followInsets tells the interface how much of each edge the system's bars cover, in layout
// pixels, until the window's own insets are known.
func (window *androidWindow) followInsets() {
	if window.insetsKnown || window.scale <= 0 {
		return
	}
	insets, known := readInsets()
	window.insetsKnown = known
	window.context.SetInsets(ui.Insets{Top: float32(insets.top) / window.scale, Bottom: float32(insets.bottom) / window.scale,
		Left: float32(insets.left) / window.scale, Right: float32(insets.right) / window.scale})
}

func (window *androidWindow) touched(event touch.Event) {
	finger := int64(event.Sequence)
	switch event.Type {
	case touch.TypeBegin:
		window.pointer.begin(window.app.Input, finger, event.X, event.Y, window.scale)
	case touch.TypeMove:
		window.pointer.move(window.app.Input, finger, event.X, event.Y, window.scale)
	case touch.TypeEnd:
		window.pointer.end(window.app.Input, finger, event.X, event.Y, window.scale)
	}
	window.requestPaint()
}

// keyPressed passes on a hardware keyboard (or the emulator's): typed characters and the keys
// that edit text.
func (window *androidWindow) keyPressed(event key.Event) {
	state := window.app.Input
	if mapped, known := keysFromAndroid[event.Code]; known {
		switch event.Direction {
		case key.DirPress:
			state.SetKey(mapped, true)
		case key.DirRelease:
			state.SetKey(mapped, false)
		case key.DirNone:
			state.RepeatKey(mapped)
		}
	}
	if event.Direction != key.DirRelease && event.Rune >= ' ' {
		state.AddCharacter(event.Rune)
	}
	window.requestPaint()
}

var keysFromAndroid = map[key.Code]input.Key{
	key.CodeDeleteBackspace: input.KeyBackspace, key.CodeDeleteForward: input.KeyDelete,
	key.CodeReturnEnter: input.KeyEnter, key.CodeTab: input.KeyTab, key.CodeEscape: input.KeyEscape,
	key.CodeLeftArrow: input.KeyLeft, key.CodeRightArrow: input.KeyRight,
	key.CodeUpArrow: input.KeyUp, key.CodeDownArrow: input.KeyDown,
	key.CodeHome: input.KeyHome, key.CodeEnd: input.KeyEnd,
	key.CodeLeftShift: input.KeyShift, key.CodeRightShift: input.KeyShift,
	key.CodeSpacebar: input.KeySpace,
}

// paint runs one interface step and draws it, unless it would look like the frame shown, then
// decides when the next one comes: a phone's battery needs the program to sleep in between.
func (window *androidWindow) paint(event paint.Event) {
	window.mutex.Lock()
	window.paintPending = false
	window.mutex.Unlock()
	if window.graphics == nil || event.External {
		return
	}
	window.followInsets()
	started := time.Now() // the pacer counts from a frame's start, so its work is inside the interval
	hadInput := window.app.Input.HasInput()
	if err := window.clock.StepOnce(window.app); err != nil {
		log.Printf("window: %v", err)
	}
	if !window.app.Unchanged(window.app) {
		if err := window.draw(); err != nil {
			log.Printf("window: %v", err)
			return
		}
		window.events.Publish()
		window.mutex.Lock()
		window.pacer.drawn(started) // only frames really drawn count: the system's own paints do not
		window.mutex.Unlock()
	}
	window.scheduleNext(hadInput)
}

func (window *androidWindow) draw() error {
	width, height, scale := frameSize(window.app)
	if view := window.view.view; view != nil {
		if window.scene == nil {
			scene, err := gles.NewSceneRenderer(mobileGL{window.graphics})
			if err != nil {
				return err
			}
			window.scene = scene
		}
		window.scene.OverlayScale = scale
		window.scene.Draw(view.World, view.Camera, width, height)
		return nil
	}
	if window.renderer == nil {
		renderer, err := gles.NewFlatRenderer(mobileGL{window.graphics})
		if err != nil {
			return err
		}
		window.renderer = renderer
	}
	window.renderer.Draw(window.context.Drawing(), window.context.Theme.Background, width, height, scale)
	return nil
}

// scheduleNext asks for the next frame: right away when drawing continuously or after a frame with
// input (the looks that depend on the new layout catch up), at the time a control asked for (a
// caret's blink), or not until the next event.
func (window *androidWindow) scheduleNext(hadInput bool) {
	if !window.app.DrawOnDemand || hadInput {
		window.requestPaint()
		return
	}
	delay, wanted := window.app.Input.TakeFrameRequest()
	if !wanted {
		return
	}
	window.mutex.Lock()
	defer window.mutex.Unlock()
	if window.timer != nil {
		window.timer.Stop()
	}
	window.timer = time.AfterFunc(delay, window.requestPaint)
}

// requestPaint sends a paint event to the loop, once however many times it is asked before the
// event arrives. Safe from any goroutine.
func (window *androidWindow) requestPaint() {
	window.mutex.Lock()
	defer window.mutex.Unlock()
	if window.paintPending || window.paced || window.events == nil {
		return
	}
	if wait := window.pacer.wait(time.Now()); wait > 0 {
		// too soon after the last frame: a finger dragging asks far faster than the cap
		window.paced = true
		time.AfterFunc(wait, window.paceOver)
		return
	}
	window.paintPending = true
	window.events.Send(paint.Event{})
}

// paceOver sends the paint that waited for the pacer.
func (window *androidWindow) paceOver() {
	window.mutex.Lock()
	window.paced = false
	window.mutex.Unlock()
	window.requestPaint()
}
