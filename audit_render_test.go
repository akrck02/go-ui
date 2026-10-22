package ui

// Audit benchmarks for a whole on-demand frame of a go-ui window, minus OpenGL: the update step
// (Begin, build, End: layout + draw list) and go-renderer's painter (draw list -> quads), at
// 2560x1440 framebuffer / 1280x720 window (scale 2), with the real fonts.
//
//	go test . -run Audit -bench Audit -benchmem -v

import (
	"fmt"
	"hash/maphash"
	"testing"
	"unsafe"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/opengl/overlay"
	"github.com/akrck02/go-renderer/scene"
)

type auditWindow struct {
	ui      *Context
	state   *input.State
	painter *overlay.Painter
	layer   scene.Overlay
	build   func(ui *Context)
}

func newAuditWindow(tb testing.TB) *auditWindow {
	context, err := New()
	if err != nil {
		tb.Fatal(err)
	}
	painter, err := overlay.NewPainter(overlay.Fonts{})
	if err != nil {
		tb.Fatal(err)
	}
	state := input.NewState()
	state.WindowWidth, state.WindowHeight = 1280, 720
	state.FramebufferWidth, state.FramebufferHeight = 2560, 1440
	labels := make([]string, 50)
	for index := range labels {
		labels[index] = fmt.Sprintf("Item %d", index)
	}
	checked := make([]bool, 50)
	build := func(ui *Context) {
		for _, panel := range []string{"Scene", "Light", "Tools"} {
			ui.Panel(panel, func() {
				for index, label := range labels {
					ui.Row(func() {
						ui.Checkbox(label+"##check"+panel, &checked[index])
						ui.Spacer()
						ui.Button(label + "##" + panel)
					})
				}
			})
		}
	}
	window := &auditWindow{ui: context, state: state, painter: painter, build: build}
	window.update()
	window.paint()
	return window
}

func (window *auditWindow) update() {
	window.ui.Begin(window.state)
	window.build(window.ui)
	window.ui.End()
	window.state.Feed().EndFrame()
}

func (window *auditWindow) paint() []overlay.Quad {
	window.layer.Drawing = window.ui.Drawing()
	return window.painter.Paint(&window.layer, nil, overlay.Frame{Width: 2560, Height: 1440, Scale: 2})
}

func hashDrawing(seed maphash.Seed, list *draw2d.List) uint64 {
	var hash maphash.Hash
	hash.SetSeed(seed)
	for index := range list.Commands {
		command := list.Commands[index]
		text := command.Text
		command.Text = ""
		hash.Write(unsafe.Slice((*byte)(unsafe.Pointer(&command)), unsafe.Sizeof(command)))
		hash.WriteString(text)
	}
	return hash.Sum64()
}

// TestAuditFrame logs the shape of the frame and whether a pointer move that changes nothing
// visible produces an identical draw list (so the GPU frame could be skipped).
func TestAuditFrame(t *testing.T) {
	window := newAuditWindow(t)
	quads := window.paint()
	list := window.ui.Drawing()
	kinds := map[draw2d.Kind]int{}
	for _, command := range list.Commands {
		kinds[command.Kind]++
	}
	area := 0.0
	for _, quad := range quads {
		area += float64(quad.Screen[2]-quad.Screen[0]) * float64(quad.Screen[3]-quad.Screen[1])
	}
	t.Logf("commands=%d (rect %d, shadow %d, text %d, icon %d) quads=%d vertexBytes=%d KiB",
		len(list.Commands), kinds[draw2d.RectangleKind], kinds[draw2d.ShadowKind], kinds[draw2d.TextKind], kinds[draw2d.IconKind],
		len(quads), len(quads)*6*8*4/1024)
	t.Logf("blended quad area %.2f Mpx = %.2fx of 2560x1440 (plus the clear)", area/1e6, area/(2560*1440))
	// texts are not dropped by draw2d when their clip hides them (draw2d/list.go add), so the
	// painter lays out their glyphs and clipQuads throws them away: count how many
	unclipped := draw2d.List{Commands: append([]draw2d.Command(nil), list.Commands...)}
	for index := range unclipped.Commands {
		unclipped.Commands[index].HasClip = false
	}
	window.layer.Drawing = &unclipped
	before := len(window.painter.Paint(&window.layer, nil, overlay.Frame{Width: 2560, Height: 1440, Scale: 2}))
	t.Logf("quads before clipping %d, after %d: %d%% of the painter's quads are thrown away", before, len(quads), 100*(before-len(quads))/before)
	atlas := window.painter.Atlas()
	t.Logf("atlas %d px, version %d", atlas.Size(), atlas.Version)

	seed := maphash.MakeSeed()
	first := hashDrawing(seed, list)
	window.state.Feed().MoveCursor(1270, 710) // over empty background, far from every control
	window.update()
	window.state.Feed().MoveCursor(1271, 711)
	window.update()
	afterMove := hashDrawing(seed, window.ui.Drawing())
	window.update() // the on-demand follow-up frame, no input
	afterFollowUp := hashDrawing(seed, window.ui.Drawing())
	t.Logf("draw list unchanged after an idle pointer move: %v; after the follow-up frame: %v", first == afterMove, afterMove == afterFollowUp)
}

// BenchmarkAuditUpdateStep is the CPU of one update step (build + layout + draw list).
func BenchmarkAuditUpdateStep(b *testing.B) {
	window := newAuditWindow(b)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		window.update()
	}
}

// BenchmarkAuditPaint is the painter's share: draw list -> quads.
func BenchmarkAuditPaint(b *testing.B) {
	window := newAuditWindow(b)
	b.ReportMetric(float64(len(window.ui.Drawing().Commands)), "commands")
	b.ReportMetric(float64(len(window.paint())), "quads")
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		window.paint()
	}
}

// BenchmarkAuditWholeFrameCPU is update + paint: what every on-demand wake-up costs before GL.
func BenchmarkAuditWholeFrameCPU(b *testing.B) {
	window := newAuditWindow(b)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		window.update()
		window.paint()
	}
}

// BenchmarkAuditHashDrawing is the cost of the check that would let an unchanged frame be skipped.
func BenchmarkAuditHashDrawing(b *testing.B) {
	window := newAuditWindow(b)
	seed := maphash.MakeSeed()
	list := window.ui.Drawing()
	b.ReportAllocs()
	b.ResetTimer()
	var sink uint64
	for range b.N {
		sink ^= hashDrawing(seed, list)
	}
	_ = sink
}
