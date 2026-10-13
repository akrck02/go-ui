package ui

import (
	"testing"
	"time"

	"github.com/akrck02/go-renderer/input"
)

// clockedHarness drives frames at chosen times.
func clockedFrame(context *Context, state *input.State, at time.Time, build func()) {
	state.Now = func() time.Time { return at }
	context.Begin(state)
	build()
	context.End()
	state.EndFrame()
}

func TestAValueEasesToItsTargetAndAsksForFramesMeanwhile(t *testing.T) {
	context, state := NewWithMeasurer(&fakeMeasurer{}), input.NewState()
	start := time.Unix(1000, 0)
	target, value := float32(0), float32(0)
	build := func() { value = context.Animate("handle", target, 100*time.Millisecond) }
	clockedFrame(context, state, start, build)
	if value != 0 {
		t.Fatalf("value %v on first sight; want its target", value)
	}
	target = 1
	clockedFrame(context, state, start.Add(16*time.Millisecond), build)
	clockedFrame(context, state, start.Add(66*time.Millisecond), build)
	if value <= 0.5 || value >= 1 {
		t.Fatalf("value %v halfway through; want past half (it eases out) and short of 1", value)
	}
	if _, wanted := state.TakeFrameRequest(); !wanted {
		t.Fatal("no frame asked for while the value moves")
	}
	clockedFrame(context, state, start.Add(200*time.Millisecond), build)
	if value != 1 {
		t.Fatalf("value %v after its time; want 1", value)
	}
	if _, wanted := state.TakeFrameRequest(); wanted {
		t.Fatal("frames asked for after the value arrived")
	}
}

func TestSomethingAppearsEachTimeItIsShownAgain(t *testing.T) {
	context, state := NewWithMeasurer(&fakeMeasurer{}), input.NewState()
	start := time.Unix(1000, 0)
	shown, amount := true, float32(0)
	build := func() {
		if shown {
			amount = context.Appear("dialog", 100*time.Millisecond)
		}
	}
	clockedFrame(context, state, start, build)
	if amount != 0 {
		t.Fatalf("appeared %v on its first frame; want 0", amount)
	}
	clockedFrame(context, state, start.Add(time.Second), build)
	if amount != 1 {
		t.Fatalf("appeared %v after a second; want 1", amount)
	}
	shown = false
	clockedFrame(context, state, start.Add(2*time.Second), build)
	shown = true
	clockedFrame(context, state, start.Add(3*time.Second), build)
	if amount != 0 {
		t.Fatalf("appeared %v when shown again; want it to start again", amount)
	}
}

func TestTheBackClosesTheFrontModalAndIsAnsweredOnce(t *testing.T) {
	h := newHarness(t)
	dismissed, screenBack := 0, 0
	open := true
	build := func(ui *Context) {
		if open && ui.Modal("question", LayerOptions{AlignX: Center, AlignY: Center}, func() { ui.Label("Sure?") }) {
			dismissed++
			open = false
		}
		if ui.Back() {
			screenBack++
		}
	}
	h.frame(build)
	h.state.PressBack()
	h.frame(build)
	if dismissed != 1 || screenBack != 0 {
		t.Fatalf("modal dismissed %d, screen went back %d: the modal answers the back alone", dismissed, screenBack)
	}
	h.state.PressBack()
	h.frame(build)
	if screenBack != 1 {
		t.Fatal("with the modal gone, the screen does not get the back")
	}
}
