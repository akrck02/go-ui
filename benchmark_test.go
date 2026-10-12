package ui

import (
	"fmt"
	"testing"
)

// BenchmarkThreeHundredControls builds, lays out and paints a frame of three panels with a
// hundred controls each, the size of a busy editor.
func BenchmarkThreeHundredControls(b *testing.B) {
	h := newHarness(&testing.T{})
	labels := make([]string, 50)
	for index := range labels {
		labels[index] = fmt.Sprintf("Item %d", index)
	}
	build := func(ui *Context) {
		for _, panel := range []string{"Scene", "Light", "Tools"} {
			ui.Panel(panel, func() {
				for _, label := range labels {
					ui.Row(func() {
						ui.Label(label)
						ui.Spacer()
						ui.Button(label + "##" + panel)
					})
				}
			})
		}
	}
	h.frame(build)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		h.frame(build)
	}
}

// BenchmarkNewControls builds, lays out and paints a frame of a resizable side column with a table
// of three hundred rows in three groups, a menu button, a row of tags that wraps and drag sources
// over a drop target: the pieces of an editor's side list and inspector.
func BenchmarkNewControls(b *testing.B) {
	h := newHarness(&testing.T{})
	keys := make([]string, 300)
	counts := make([]string, 300)
	for index := range keys {
		keys[index] = fmt.Sprintf("part-%d", index)
		counts[index] = fmt.Sprintf("%d", index*10)
	}
	columns := []Column{{Title: "Part", Width: Share(1)}, {Title: "Count", Width: Fixed(60), Align: End}}
	width := float32(280)
	selected := 3
	build := func(ui *Context) {
		ui.Box(Layout{Width: Grow(), Height: Grow()}, func() {
			ui.Resizable("side", RightSide, &width, 200, 400, Layout{Direction: Vertical, Height: Grow(), Gap: 8}, func() {
				ui.MenuButton("plus", "Add", func() { ui.MenuItem("", "Hut", "") })
				ui.Table("parts", Layout{Width: Grow(), Height: Grow()}, columns, func(table *Table) {
					for index := range keys {
						if index%100 == 0 {
							table.Group("Group")
						}
						if table.Row(keys[index], index == selected, keys[index], counts[index]) {
							selected = index
						}
					}
				})
			})
			ui.Box(Layout{Direction: Vertical, Width: Grow(), Gap: 8}, func() {
				ui.Box(Layout{Width: Grow(), Gap: 4, Wrap: true}, func() {
					for index := range 12 {
						ui.DragSource(keys[index], index, Layout{}, func() { ui.Chip("tag", keys[index], ChipLook{}) })
					}
				})
				ui.DropTarget("shelf", Layout{Width: Grow(), Height: Fixed(80)}, nil, nil)
			})
		})
	}
	h.frame(build)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		h.frame(build)
	}
}
