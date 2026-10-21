package main

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Sizes of the viewer's layout, in layout pixels.
const (
	sidebarWidth = 320
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

// sidebar lists the tests, failed first, with a chip for each error code.
func (results *results) sidebar(context *ui.Context) {
	theme := context.Theme
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(sidebarWidth), Height: ui.Grow(), Padding: ui.Pad(12), Gap: 8}
	context.Frame(layout, ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1}, func() {
		context.Heading("Window tests")
		passed, failed := results.counts()
		context.Row(func() {
			context.Chip("circle-check", fmt.Sprintf("%d passed", passed), ui.ChipLook{Fill: theme.PrimaryTint, Ink: theme.Positive})
			context.Chip("circle-x", fmt.Sprintf("%d failed", failed), ui.ChipLook{Fill: theme.Sunken, Ink: theme.Danger})
		})
		context.Separator()
		context.ScrollArea("tests", ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: 2}, func() {
			for index, recording := range results.recordings {
				icon := "circle-check"
				if recording.Failed {
					icon = "circle-x"
				}
				if context.Selectable(icon, recording.Test, index == results.selected) {
					results.choose(index)
				}
			}
		})
	})
}

// chosenTest shows the chosen test: its name and state, its errors and the replay of its frames.
func (results *results) chosenTest(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(pagePadding), Gap: 16}
	context.Box(layout, func() {
		recording, found := results.current()
		if !found {
			context.Heading("No window tests were recorded")
			context.Caption("A test records when its window is made with uitest.Test(t).")
			return
		}
		title(context, recording.Test, recording.Failed)
		results.errorCard(context)
		results.replayCard(context)
	})
}

// title is a test's name and a chip saying whether it passed.
func title(context *ui.Context, name string, failed bool) {
	theme := context.Theme
	context.Row(func() {
		context.Text(name, ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
		if failed {
			context.Chip("circle-x", "Failed", ui.ChipLook{Fill: theme.Sunken, Ink: theme.Danger})
			return
		}
		context.Chip("circle-check", "Passed", ui.ChipLook{Fill: theme.PrimaryTint, Ink: theme.Positive})
	})
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
		shown := replay{recording: recording, frame: frame, theme: theme}
		context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, shown.paint)
		results.stepControls(context, len(recording.Frames), frame.Action)
	})
}

// stepControls are the buttons and the slider that move through the frames.
func (results *results) stepControls(context *ui.Context, frames int, action string) {
	context.Row(func() {
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
		if action != "" {
			context.Chip("mouse-pointer-click", action, ui.ChipLook{})
		}
	})
}
