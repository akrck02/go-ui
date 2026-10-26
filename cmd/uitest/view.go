package main

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Sizes of the viewer's layout, in layout pixels.
const (
	sidebarWidth = 380
	cardRadius   = 10
	pagePadding  = 24
)

// draw builds the viewer: the list of tests on the left, the chosen one on the right.
func (results *results) draw(context *ui.Context) {
	theme := context.Theme
	page := ui.Layout{Width: ui.Grow(), Height: ui.Grow()}
	context.Frame(page, ui.Look{Fill: theme.Background}, func() {
		results.sidebar(context)
		results.chosenTest(context)
	})
}

// sidebar lists the tests by package, failed first, with how many passed, failed and were
// skipped, and which tests to show: all, the failed ones or those with a window.
func (results *results) sidebar(context *ui.Context) {
	theme := context.Theme
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(sidebarWidth), Height: ui.Grow(), Padding: ui.Pad(12), Gap: 8}
	context.Frame(layout, ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1}, func() {
		context.Heading("Tests")
		results.filterChips(context)
		context.Separator()
		context.ScrollArea("tests", ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: 2}, func() {
			results.testList(context)
		})
	})
}

// filterChips choose which tests the list shows, each saying how many it shows: all of them,
// the failed ones (and the skipped, when there are any), and those with a window. The chosen one
// stands out in the theme's main colour.
func (results *results) filterChips(context *ui.Context) {
	passed, failed, skipped := results.counts()
	context.Box(ui.Layout{Width: ui.Grow(), Wrap: true, Gap: 6}, func() {
		results.filterChip(context, allTests, "list", fmt.Sprintf("All %d", passed+failed+skipped))
		results.filterChip(context, failedTests, "circle-x", fmt.Sprintf("Failed %d", failed))
		if skipped > 0 {
			results.filterChip(context, skippedTests, "circle-minus", fmt.Sprintf("Skipped %d", skipped))
		}
		results.filterChip(context, windowTests, "app-window", fmt.Sprintf("With a window %d", results.windowCount()))
	})
}

// filterChip is one filter: a chip that shows its tests when clicked.
func (results *results) filterChip(context *ui.Context, filter testFilter, icon, label string) {
	theme := context.Theme
	look := ui.ChipLook{Fill: theme.Sunken, Ink: theme.InkSoft}
	if results.shown == filter {
		look = ui.ChipLook{Fill: theme.PrimaryTint, Ink: theme.Primary}
	}
	if context.Chip(icon, label, look) {
		results.shown = filter
	}
}

// testList is the tests the filter shows, each package's under its name.
func (results *results) testList(context *ui.Context) {
	theme := context.Theme
	lastPackage, listed := "", false
	for index, test := range results.tests {
		if !results.isShown(test) {
			continue
		}
		listed = true
		if test.Package != lastPackage {
			context.Text(results.shortPackage(test.Package), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkQuiet, Weight: draw2d.Medium})
			lastPackage = test.Package
		}
		context.Push(test.Package)
		if context.Selectable(stateIcon(test), test.Name, index == results.selected) {
			results.choose(index)
		}
		context.Tooltip(test.Name) // a long name is cut at the list's edge
		context.Pop()
	}
	if !listed {
		context.Caption("No test here.")
	}
}

// stateIcon is a test's icon: a window's when it made one and passed, else how it ended.
func stateIcon(test testResult) string {
	switch {
	case test.State == failedTest:
		return "circle-x"
	case test.State == skippedTest:
		return "circle-minus"
	case test.recording >= 0:
		return "app-window"
	}
	return "circle-check"
}

// chosenTest shows the chosen test: its name, state and package, and then its window's errors and
// replay, or what it wrote when it made no window.
func (results *results) chosenTest(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(pagePadding), Gap: 16}
	context.Box(layout, func() {
		test, found := results.currentTest()
		if !found {
			context.Heading("No tests ran")
			context.Caption("go test found no test in the packages given.")
			return
		}
		title(context, test)
		if _, hasWindow := results.current(); hasWindow {
			results.errorCard(context)
			results.replayCard(context)
			return
		}
		outputCard(context, test)
	})
}

// title is a test's name, a chip saying how it ended, and its package and time under them.
func title(context *ui.Context, test testResult) {
	theme := context.Theme
	context.Row(func() {
		context.Text(test.Name, ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
		switch test.State {
		case failedTest:
			context.Chip("circle-x", "Failed", ui.ChipLook{Fill: theme.Sunken, Ink: theme.Danger})
		case skippedTest:
			context.Chip("circle-minus", "Skipped", ui.ChipLook{Fill: theme.Sunken, Ink: theme.InkSoft})
		default:
			context.Chip("circle-check", "Passed", ui.ChipLook{Fill: theme.PrimaryTint, Ink: theme.Positive})
		}
	})
	if test.Package != "" {
		context.Caption(fmt.Sprintf("%s · %.2f s", test.Package, test.Seconds))
	}
}

// outputCard shows what a test with no window wrote, line by line, as go test printed it.
func outputCard(context *ui.Context, test testResult) {
	theme := context.Theme
	card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 8}, func() {
		context.Row(func() {
			context.IconSized("terminal", theme.IconSize, theme.InkSoft)
			context.Text("Output", ui.TextStyle{Weight: draw2d.Medium})
		})
		if len(test.Output) == 0 {
			context.Caption("It wrote nothing: a test of code, with no window, that " + endedAs(test.State) + ".")
			return
		}
		context.ScrollArea("output", ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: 2}, func() {
			for index, line := range test.Output {
				context.Push(fmt.Sprint(index))
				context.Text(line, ui.TextStyle{Weight: draw2d.Mono, Size: theme.SmallTextSize})
				context.Pop()
			}
		})
	})
}

// endedAs says how a test ended, in words.
func endedAs(state testState) string {
	switch state {
	case failedTest:
		return "failed"
	case skippedTest:
		return "was skipped"
	}
	return "passed"
}

// card is a box in the theme's surface, with a border and round corners.
func card(context *ui.Context, layout ui.Layout, content func()) {
	theme := context.Theme
	context.Frame(layout, ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: cardRadius}, content)
}

// errorCard lists the errors the test's window returned: the code as a chip, then the message.
func (results *results) errorCard(context *ui.Context) {
	recording, _ := results.current() // chosenTest only calls it with a test chosen
	if len(recording.Errors) == 0 {
		return
	}
	theme := context.Theme
	card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(16), Gap: 8}, func() {
		for index, recorded := range recording.Errors {
			context.Push(fmt.Sprint(index))
			context.Row(func() {
				context.IconSized("triangle-alert", theme.IconSize, theme.Danger)
				if context.Chip("", string(recorded.Code), ui.ChipLook{Fill: theme.Sunken, Ink: theme.Danger}) {
					results.goTo(recorded.Frame - 1) // the frame the error came after
				}
				context.Caption(fmt.Sprintf("after frame %d", recorded.Frame))
			})
			context.Label(recorded.Message)
			context.Pop()
		}
	})
}

// replayCard shows the frame chosen and the controls to step through the frames.
func (results *results) replayCard(context *ui.Context) {
	recording, _ := results.current() // chosenTest only calls it with a test chosen
	frame, hasFrame := results.currentFrame()
	theme := context.Theme
	card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 12}, func() {
		if !hasFrame {
			context.Caption("This test ran no frame.")
			return
		}
		shown := replay{recording: recording, frame: frame, theme: theme, measurer: results.measurer}
		context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, shown.paint)
		results.stepControls(context, len(recording.Frames), frame.Action)
	})
}

// stepControls are the buttons and the slider that move through the frames, and play and pause
// them.
func (results *results) stepControls(context *ui.Context, frames int, action string) {
	if wait := results.advance(); wait > 0 {
		context.RequestFrame(wait)
	}
	context.Row(func() {
		playIcon, playLabel := "play", "Play"
		if results.playback.playing {
			playIcon, playLabel = "pause", "Pause"
		}
		if context.FlatIconButton(playIcon, playLabel) {
			results.togglePlaying()
		}
		if context.FlatIconButton("chevron-left", "Previous frame") {
			results.step(-1)
		}
		context.Box(ui.Layout{Width: ui.Grow()}, func() {
			if context.Slider("frame", &results.frameValue, 0, float32(max(frames-1, 0))) {
				results.goTo(int(results.frameValue + 0.5))
			}
		})
		if context.FlatIconButton("chevron-right", "Next frame") {
			results.step(1)
		}
	})
	context.Row(func() {
		context.Caption(fmt.Sprintf("Frame %d of %d", results.frame+1, frames))
		context.Checkbox("At the host's speed", &results.playback.atHostSpeed)
		if action != "" {
			context.Chip("mouse-pointer-click", action, ui.ChipLook{})
		}
	})
}
