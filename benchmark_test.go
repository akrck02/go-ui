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
