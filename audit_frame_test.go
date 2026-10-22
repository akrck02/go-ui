package ui

// Performance audit: the parts of a frame measured one by one, with the real text measurer, at
// 1280×720 on a 2× screen. Run with:
//
//	go test -run Audit -bench Audit -benchmem -v

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

func newAuditContext(tb testing.TB) (*Context, *input.State) {
	tb.Helper()
	context, err := New()
	if err != nil {
		tb.Fatal(err)
	}
	state := input.NewState()
	state.WindowWidth, state.WindowHeight = 1280, 720
	state.FramebufferWidth, state.FramebufferHeight = 2560, 1440
	clock := time.Unix(1_000_000, 0)
	state.Now = func() time.Time { return clock }
	state.Feed().MoveCursor(10, 700)
	return context, state
}

// auditRows are a table's rows: two distinct texts each, like the register.
type auditRow struct{ name, place, built, rangeText string }

func makeAuditRows(count int) []auditRow {
	rows := make([]auditRow, count)
	for index := range rows {
		rows[index] = auditRow{
			name: fmt.Sprintf("Lighthouse %d", index), place: fmt.Sprintf("Island %d, outer bay", index),
			built: strconv.Itoa(1850 + index%120), rangeText: fmt.Sprintf("%d nm", 5+index%20),
		}
	}
	return rows
}

var auditColumns = []Sizing{Share(1), Share(1.4), Fixed(64), Fixed(72)}

func auditCell(ui *Context, width Sizing, content func()) {
	ui.Frame(Layout{Width: width, AlignY: Center}, Look{Clip: true}, content)
}

func auditTableRow(ui *Context, row auditRow) {
	layout := Layout{Width: Grow(), Height: Fixed(40), Padding: PadXY(16, 0), Gap: 12, AlignY: Center}
	ui.Clickable(row.name+"##row", layout, Looks{Hovered: Look{Fill: ui.Theme.Sunken}}, func() {
		auditCell(ui, auditColumns[0], func() { ui.Text(row.name, TextStyle{Weight: draw2d.Medium}) })
		auditCell(ui, auditColumns[1], func() { ui.Label(row.place) })
		auditCell(ui, auditColumns[2], func() { ui.Label(row.built) })
		auditCell(ui, auditColumns[3], func() { ui.Label(row.rangeText) })
	})
}

// buildScrolledTable builds every row inside a ScrollArea, as the register app does.
func buildScrolledTable(rows []auditRow) func(ui *Context) {
	return func(ui *Context) {
		ui.Box(Layout{Direction: Vertical, Width: Grow(), Height: Grow(), Padding: Pad(24)}, func() {
			ui.ScrollArea("rows", Layout{Width: Grow(), Height: Grow()}, func() {
				for _, row := range rows {
					auditTableRow(ui, row)
				}
			})
		})
	}
}

// buildListTable builds only the visible rows with List.
func buildListTable(rows []auditRow) func(ui *Context) {
	return func(ui *Context) {
		ui.Box(Layout{Direction: Vertical, Width: Grow(), Height: Grow(), Padding: Pad(24)}, func() {
			ui.List("rows", Layout{Width: Grow(), Height: Grow()}, len(rows), 40, func(index int) {
				auditTableRow(ui, rows[index])
			})
		})
	}
}

func auditFrame(context *Context, state *input.State, build func(ui *Context)) {
	context.Begin(state)
	build(context)
	context.End()
	state.Feed().EndFrame()
}

// TestAuditSizes logs the size of the structs copied on every element and command.
func TestAuditSizes(t *testing.T) {
	t.Logf("element %d B, Layout %d B, Look %d B, Sizing %d B, TextStyle %d B, draw2d.Command %d B, textKey %d B",
		unsafe.Sizeof(element{}), unsafe.Sizeof(Layout{}), unsafe.Sizeof(Look{}), unsafe.Sizeof(Sizing{}),
		unsafe.Sizeof(TextStyle{}), unsafe.Sizeof(draw2d.Command{}), unsafe.Sizeof(textKey{}))
	for _, count := range []int{11, 1000} {
		context, state := newAuditContext(t)
		build := buildScrolledTable(makeAuditRows(count))
		for range 3 {
			auditFrame(context, state, build)
		}
		depth := 0
		for index := range context.elements {
			level := 0
			for parent := context.elements[index].parent; parent >= 0; parent = context.elements[parent].parent {
				level++
			}
			depth = max(depth, level)
		}
		t.Logf("%d rows in a ScrollArea: %d elements (%d KB of element structs), max depth %d, %d draw commands, %d ids in current, %d memory entries",
			count, len(context.elements), len(context.elements)*int(unsafe.Sizeof(element{}))/1024, depth,
			len(context.drawing.Commands), len(context.previous), len(context.memory))
	}
}

// BenchmarkAuditTable compares a table of rows built in a ScrollArea (every row built, laid out
// and painted) with the same rows in List (only the visible ones).
func BenchmarkAuditTable(b *testing.B) {
	for _, count := range []int{11, 100, 1000, 10000} {
		rows := makeAuditRows(count)
		for _, variant := range []struct {
			name  string
			build func(ui *Context)
		}{{"ScrollArea", buildScrolledTable(rows)}, {"List", buildListTable(rows)}} {
			b.Run(fmt.Sprintf("%s/rows=%d", variant.name, count), func(b *testing.B) {
				context, state := newAuditContext(b)
				for range 3 {
					auditFrame(context, state, variant.build)
				}
				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					auditFrame(context, state, variant.build)
				}
				b.StopTimer()
				b.ReportMetric(float64(len(context.elements)), "elements/frame")
			})
		}
	}
}

// BenchmarkAuditPhases splits a frame of a 1000-row scrolled table into its phases.
func BenchmarkAuditPhases(b *testing.B) {
	rows := makeAuditRows(1000)
	build := buildScrolledTable(rows)
	context, state := newAuditContext(b)
	for range 3 {
		auditFrame(context, state, build)
	}
	b.Run("build", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			context.Begin(state)
			build(context)
			b.StopTimer()
			context.End()
			state.Feed().EndFrame()
			b.StartTimer()
		}
	})
	phase := func(name string, run func()) {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				b.StopTimer()
				context.Begin(state)
				build(context)
				context.building = false
				b.StartTimer()
				run()
				b.StopTimer()
				context.building = true
				context.End()
				state.Feed().EndFrame()
				b.StartTimer()
			}
		})
	}
	phase("layOut", func() { context.layOut() })
	phase("layOut/fitContent+shareSpace", func() {
		for rootIndex := range context.roots {
			context.size(&context.roots[rootIndex], Horizontal)
			context.size(&context.roots[rootIndex], Vertical)
		}
	})
	phase("layOut/place", func() {
		for rootIndex := range context.roots {
			context.place(context.roots[rootIndex].element)
		}
	})
	phase("layOut/surfacesAndIds", func() {
		screen := context.roots[0].element
		for index := range context.elements {
			if id := context.elements[index].id; id != 0 {
				context.current[id] = context.elements[index].rect
			}
			if context.elements[index].look.Fill.IsVisible() && index != screen && context.rootElementOf(index) == screen {
				context.nextSurfaces = append(context.nextSurfaces, context.elements[index].rect)
			}
		}
		context.nextSurfaces = context.nextSurfaces[:0]
	})
	phase("layOut/rootElementOfOnly", func() {
		sum := 0
		for index := range context.elements {
			sum += context.rootElementOf(index)
		}
		_ = sum
	})
	phase("paint", func() { context.layOut(); clear(context.current); context.paint() })
}

// BenchmarkAuditForgetUnusedMemory is the cost End pays every frame to look for stale memory, by
// how many controls keep memory (scroll areas, fields, selects…).
func BenchmarkAuditForgetUnusedMemory(b *testing.B) {
	for _, count := range []int{10, 1000, 10000} {
		b.Run(fmt.Sprintf("entries=%d", count), func(b *testing.B) {
			context := NewWithMeasurer(&fakeMeasurer{})
			for index := range count {
				context.memory[ID(index+1)] = &remembered{value: &scrollMemory{}, lastFrame: 1}
			}
			context.frame = 2
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				context.forgetUnusedMemory()
			}
		})
	}
}

// BenchmarkAuditTextField is a focused field with a long text, the caret moving one character per
// frame (arrow key held): every frame measures new prefixes of the text.
func BenchmarkAuditTextField(b *testing.B) {
	long := strings.Repeat("Rocky island on the north shore ", 6)
	for _, variant := range []struct {
		name   string
		action func(state *input.State, frame int)
	}{
		{"focused-idle", nil},
		{"caret-moving", func(state *input.State, frame int) {
			key := input.KeyLeft
			if frame/150%2 == 1 {
				key = input.KeyRight
			}
			state.Feed().SetKey(key, true)
			state.Feed().SetKey(key, false)
		}},
		{"drag-selecting", func(state *input.State, frame int) {
			state.Feed().MoveCursor(40+float64(frame%300), 30)
		}},
	} {
		b.Run(variant.name, func(b *testing.B) {
			context, state := newAuditContext(b)
			text := long
			build := func(ui *Context) {
				ui.Box(Layout{Width: Fixed(320), Padding: Pad(10)}, func() { ui.TextField("note", &text) })
			}
			auditFrame(context, state, build)
			context.Focus("note")
			auditFrame(context, state, build)
			if variant.name == "drag-selecting" {
				state.Feed().MoveCursor(40, 30)
				state.Feed().SetButton(input.MouseLeft, true)
				auditFrame(context, state, build)
			}
			measurer := &countingAuditMeasurer{inner: context.measurer}
			context.measurer = measurer
			b.ReportAllocs()
			b.ResetTimer()
			for frame := range b.N {
				if variant.action != nil {
					variant.action(state, frame)
				}
				auditFrame(context, state, build)
			}
			b.StopTimer()
			b.ReportMetric(float64(measurer.calls)/float64(b.N), "measures/frame")
			b.ReportMetric(float64(len(context.texts.entries)), "cached-texts")
		})
	}
}

type countingAuditMeasurer struct {
	inner draw2d.Measurer
	calls int
}

func (measurer *countingAuditMeasurer) MeasureText(text string, size float32, weight draw2d.Weight, scale float32) draw2d.TextMetrics {
	measurer.calls++
	return measurer.inner.MeasureText(text, size, weight, scale)
}

func (measurer *countingAuditMeasurer) HasIcon(name string) bool { return measurer.inner.HasIcon(name) }

// BenchmarkAuditMeasureMiss is one text measured by the font (a cache miss), by length.
func BenchmarkAuditMeasureMiss(b *testing.B) {
	context, _ := newAuditContext(b)
	for _, length := range []int{10, 40} {
		text := strings.Repeat("a", length)
		b.Run(fmt.Sprintf("runes=%d", length), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				context.measurer.MeasureText(text, 13, draw2d.Regular, 2)
			}
		})
	}
	b.Run("cache-hit", func(b *testing.B) {
		context.scale = 2
		context.measure("Lighthouse 42", TextStyle{Size: 13})
		b.ReportAllocs()
		for range b.N {
			context.measure("Lighthouse 42", TextStyle{Size: 13})
		}
	})
}
