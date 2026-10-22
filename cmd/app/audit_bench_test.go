package main

// Performance audit benchmarks: a full frame of the register app (Begin, build, End) with the
// real text measurer, at 1280×720 on a 2× screen, in the situations a desktop app spends its
// frames in. Run with:
//
//	go test ./cmd/app -run xxx -bench Audit -benchmem

import (
	"fmt"
	"testing"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/opengl/overlay"
	"github.com/akrck02/go-renderer/scene"
	ui "github.com/akrck02/go-ui"
)

type auditHarness struct {
	app     *register
	context *ui.Context
	state   *input.State
	clock   time.Time
}

func newAuditHarness(b *testing.B, lighthouses int) *auditHarness {
	b.Helper()
	context, err := ui.New()
	if err != nil {
		b.Fatal(err)
	}
	state := input.NewState()
	state.WindowWidth, state.WindowHeight = 1280, 720
	state.FramebufferWidth, state.FramebufferHeight = 2560, 1440
	harness := &auditHarness{app: newRegister(), context: context, state: state, clock: time.Unix(1_000_000, 0)}
	state.Now = func() time.Time { return harness.clock }
	if lighthouses > 0 {
		harness.app.lighthouses = manyLighthouses(lighthouses)
	}
	state.Feed().MoveCursor(10, 700) // the status bar: over the interface, on no control
	return harness
}

// manyLighthouses makes a register of count entries with distinct names and places.
func manyLighthouses(count int) []*lighthouse {
	entries := make([]*lighthouse, count)
	for index := range entries {
		entries[index] = &lighthouse{
			name: fmt.Sprintf("Lighthouse %d", index), place: fmt.Sprintf("Island %d, outer bay", index),
			built: 1850 + index%120, height: 10 + index%20, focalHeight: 12 + index%40,
			character: "Fl W 5s", rangeNM: float32(5 + index%20), lit: index%3 != 0,
		}
	}
	return entries
}

func (harness *auditHarness) frame() {
	harness.clock = harness.clock.Add(16 * time.Millisecond)
	harness.context.Begin(harness.state)
	harness.app.draw(harness.context)
	harness.context.End()
	harness.state.Feed().EndFrame()
}

func (harness *auditHarness) warm() {
	for range 5 {
		harness.frame()
	}
}

func (harness *auditHarness) report(b *testing.B) {
	b.ReportMetric(float64(len(harness.context.Drawing().Commands)), "commands/frame")
}

func runAudit(b *testing.B, harness *auditHarness, beforeFrame func(frame int)) {
	harness.warm()
	b.ReportAllocs()
	b.ResetTimer()
	for frame := range b.N {
		if beforeFrame != nil {
			beforeFrame(frame)
		}
		harness.frame()
	}
	b.StopTimer()
	harness.report(b)
}

// BenchmarkAuditIdle is a frame where nothing changes: the pointer rests on the status bar.
func BenchmarkAuditIdle(b *testing.B) {
	runAudit(b, newAuditHarness(b, 0), nil)
}

// BenchmarkAuditPointerOverRows moves the pointer down and up the table, a row hovered per frame.
func BenchmarkAuditPointerOverRows(b *testing.B) {
	harness := newAuditHarness(b, 0)
	runAudit(b, harness, func(frame int) {
		y := 330 + float64(frame%40)*8
		harness.state.Feed().MoveCursor(500, y)
	})
}

// BenchmarkAuditFocusedName has the inspector's Name field focused (caret blinking).
func BenchmarkAuditFocusedName(b *testing.B) {
	harness := newAuditHarness(b, 0)
	harness.frame()
	harness.context.Focus("0", "Name")
	runAudit(b, harness, nil)
}

// BenchmarkAuditTypingInName types a letter and deletes it, one key per frame, in the Name field.
func BenchmarkAuditTypingInName(b *testing.B) {
	harness := newAuditHarness(b, 0)
	harness.frame()
	harness.context.Focus("0", "Name")
	harness.frame()
	harness.state.Feed().SetKey(input.KeyEnd, true) // caret to the end
	harness.frame()
	harness.state.Feed().SetKey(input.KeyEnd, false)
	runAudit(b, harness, func(frame int) {
		if frame%2 == 0 {
			harness.state.Feed().AddCharacter('a')
			return
		}
		harness.state.Feed().SetKey(input.KeyBackspace, true)
		harness.state.Feed().SetKey(input.KeyBackspace, false)
		harness.state.Feed().RepeatKey(input.KeyBackspace)
	})
}

// BenchmarkAuditSearchActive filters the table with a search text, the search box focused.
func BenchmarkAuditSearchActive(b *testing.B) {
	harness := newAuditHarness(b, 0)
	harness.app.search = "Rock"
	harness.frame()
	harness.context.Focus("Search")
	runAudit(b, harness, nil)
}

func benchmarkMany(b *testing.B, count int, moving bool) {
	harness := newAuditHarness(b, count)
	var before func(int)
	if moving {
		before = func(frame int) { harness.state.Feed().MoveCursor(500, 330+float64(frame%40)*8) }
	}
	runAudit(b, harness, before)
}

// BenchmarkAuditMany scales the register: the table builds every row inside its scroll area.
func BenchmarkAuditMany(b *testing.B) {
	for _, count := range []int{11, 100, 1000, 1900, 2100, 3000} {
		b.Run(fmt.Sprintf("lighthouses=%d", count), func(b *testing.B) { benchmarkMany(b, count, false) })
		b.Run(fmt.Sprintf("lighthouses=%d/pointer", count), func(b *testing.B) { benchmarkMany(b, count, true) })
	}
}

// BenchmarkAuditPainter turns the idle frame's draw list into quads with the real overlay painter
// (no OpenGL needed): the CPU half of drawing a frame after go-ui.
func BenchmarkAuditPainter(b *testing.B) {
	for _, count := range []int{11, 1000} {
		b.Run(fmt.Sprintf("lighthouses=%d", count), func(b *testing.B) {
			harness := newAuditHarness(b, count)
			if count == 11 {
				harness = newAuditHarness(b, 0)
			}
			harness.warm()
			painter, err := overlay.NewPainter(overlay.Fonts{})
			if err != nil {
				b.Fatal(err)
			}
			layer := &scene.Overlay{Drawing: harness.context.Drawing()}
			frame := overlay.Frame{Width: 2560, Height: 1440, Scale: 2}
			quads := painter.Paint(layer, nil, frame)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				quads = painter.Paint(layer, nil, frame)
			}
			b.StopTimer()
			b.ReportMetric(float64(len(quads)), "quads/frame")
			b.ReportMetric(float64(len(harness.context.Drawing().Commands)), "commands/frame")
			b.ReportMetric(float64(len(quads)*6*8*4), "vertexbytes/frame")
		})
	}
}

// countingMeasurer counts the texts that missed go-ui's cache and reached the font.
type countingMeasurer struct {
	inner *overlay.Measurer
	calls int
	runes int
}

func (measurer *countingMeasurer) MeasureText(text string, size float32, weight draw2d.Weight, scale float32) draw2d.TextMetrics {
	measurer.calls++
	measurer.runes += len([]rune(text))
	return measurer.inner.MeasureText(text, size, weight, scale)
}

func (measurer *countingMeasurer) HasIcon(name string) bool { return measurer.inner.HasIcon(name) }

// TestAuditMeasureMisses prints how many texts reach the font per frame (cache misses), per
// scenario: with more unique texts than the cache holds (4096) it is cleared mid-frame and every
// text is measured again, every frame.
func TestAuditMeasureMisses(t *testing.T) {
	scenarios := []struct {
		name  string
		count int
		setUp func(harness *auditHarness)
		step  func(harness *auditHarness, frame int)
	}{
		{name: "idle", count: 0},
		{name: "focused name", count: 0, setUp: func(harness *auditHarness) { harness.context.Focus("0", "Name") }},
		{name: "typing in name", count: 0, setUp: func(harness *auditHarness) { harness.context.Focus("0", "Name") },
			step: func(harness *auditHarness, frame int) {
				if frame%2 == 0 {
					harness.state.Feed().AddCharacter('a')
				} else {
					harness.state.Feed().SetKey(input.KeyBackspace, true)
					harness.state.Feed().SetKey(input.KeyBackspace, false)
				}
			}},
		{name: "1000 lighthouses", count: 1000},
		{name: "1900 lighthouses", count: 1900},
		{name: "2000 lighthouses", count: 2000},
		{name: "3000 lighthouses", count: 3000},
	}
	for _, scenario := range scenarios {
		inner, err := overlay.NewMeasurer(overlay.Fonts{})
		if err != nil {
			t.Fatal(err)
		}
		measurer := &countingMeasurer{inner: inner}
		state := input.NewState()
		state.WindowWidth, state.WindowHeight = 1280, 720
		state.FramebufferWidth, state.FramebufferHeight = 2560, 1440
		harness := &auditHarness{app: newRegister(), context: ui.NewWithMeasurer(measurer), state: state, clock: time.Unix(1_000_000, 0)}
		state.Now = func() time.Time { return harness.clock }
		if scenario.count > 0 {
			harness.app.lighthouses = manyLighthouses(scenario.count)
		}
		state.Feed().MoveCursor(10, 700)
		harness.frame()
		if scenario.setUp != nil {
			scenario.setUp(harness)
		}
		for range 5 {
			harness.frame()
		}
		measurer.calls, measurer.runes = 0, 0
		const frames = 20
		for frame := range frames {
			if scenario.step != nil {
				scenario.step(harness, frame)
			}
			harness.frame()
		}
		t.Logf("%-18s %7.1f measures/frame  %8.1f runes measured/frame", scenario.name,
			float64(measurer.calls)/frames, float64(measurer.runes)/frames)
	}
}

// TestAuditDrawListComposition logs what the register's draw list is made of, and how many text
// commands lie wholly outside their clip (rows scrolled out of a ScrollArea, which draw2d keeps
// because it never culls text), with the quads the painter makes from it.
func TestAuditDrawListComposition(t *testing.T) {
	for _, count := range []int{0, 1000} {
		context, err := ui.New()
		if err != nil {
			t.Fatal(err)
		}
		state := input.NewState()
		state.WindowWidth, state.WindowHeight = 1280, 720
		state.FramebufferWidth, state.FramebufferHeight = 2560, 1440
		harness := &auditHarness{app: newRegister(), context: context, state: state, clock: time.Unix(1_000_000, 0)}
		state.Now = func() time.Time { return harness.clock }
		if count > 0 {
			harness.app.lighthouses = manyLighthouses(count)
		}
		for range 3 {
			harness.frame()
		}
		kinds := map[draw2d.Kind]int{}
		clipped, textRunes, hiddenTexts, hiddenRunes := 0, 0, 0, 0
		for _, command := range context.Drawing().Commands {
			kinds[command.Kind]++
			if command.HasClip {
				clipped++
			}
			if command.Kind == draw2d.TextKind {
				runes := len([]rune(command.Text))
				textRunes += runes
				if command.HasClip && (command.Clip.IsEmpty() || command.Baseline < command.Clip.Y || command.Baseline-command.Size > command.Clip.Bottom()) {
					hiddenTexts++
					hiddenRunes += runes
				}
			}
		}
		painter, err := overlay.NewPainter(overlay.Fonts{})
		if err != nil {
			t.Fatal(err)
		}
		quads := painter.Paint(&scene.Overlay{Drawing: context.Drawing()}, nil, overlay.Frame{Width: 2560, Height: 1440, Scale: 2})
		t.Logf("lighthouses=%d: %d commands (rect %d, shadow %d, text %d, icon %d), %d clipped; %d text runes; %d texts (%d runes) wholly outside their clip; %d quads",
			max(count, 11), len(context.Drawing().Commands), kinds[draw2d.RectangleKind], kinds[draw2d.ShadowKind], kinds[draw2d.TextKind], kinds[draw2d.IconKind],
			clipped, textRunes, hiddenTexts, hiddenRunes, len(quads))
	}
}
