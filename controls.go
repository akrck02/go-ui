package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// Checkbox adds a box to tick with a label after it; clicking either flips value. It returns true
// on the step the value changes.
func (ui *Context) Checkbox(label string, value *bool) bool {
	theme := ui.Theme
	layout := Layout{Height: Fixed(theme.ControlHeight), Gap: 8, AlignY: Center}
	var box int
	side := theme.CheckSize
	_, result := ui.pressable(ui.idFor(label), layout, Look{}, func() {
		box = ui.frameWithID(0, Layout{Width: Fixed(side), Height: Fixed(side), AlignX: Center, AlignY: Center}, Look{}, func() {
			if *value {
				ui.IconSized("check", orDefault(theme.CheckMark, defaultCheckMark), theme.OnPrimary)
			}
		})
		ui.Text(displayText(label), TextStyle{})
	})
	look := Look{Fill: theme.Surface, Border: orColour(theme.CheckLine, theme.InkQuiet), BorderWidth: theme.CheckBorder, Radius: theme.CheckRadius}
	if *value {
		look = Look{Fill: theme.Primary, Border: theme.CheckedBorder, BorderWidth: theme.CheckBorder, Radius: theme.CheckRadius}
	}
	if result.hovered && !*value {
		look.Border = theme.Primary
	}
	ui.elements[box].look = look
	ui.Describe(Description{Role: CheckboxRole, Checked: *value})
	return ui.flipOnClick(result, value)
}

// Toggle adds a row with a label on the left and a switch on the right; clicking the row flips
// value. It returns true on the step the value changes.
func (ui *Context) Toggle(label string, value *bool) bool {
	theme := ui.Theme
	layout := Layout{Width: Grow(), Height: Fixed(theme.ControlHeight), Gap: 8, AlignY: Center}
	var track, knob int
	_, result := ui.pressable(ui.idFor(label), layout, Look{}, func() {
		ui.Text(displayText(label), TextStyle{})
		ui.Spacer()
		align := Start
		if *value {
			align = End
		}
		track = ui.frameWithID(0, Layout{Width: Fixed(32), Height: Fixed(18), Padding: Pad(2), AlignX: align, AlignY: Center}, Look{}, func() {
			knob = ui.frameWithID(0, Layout{Width: Fixed(14), Height: Fixed(14)}, Look{}, nil)
		})
	})
	trackFill := theme.Line
	if *value {
		trackFill = theme.Primary
	}
	if result.hovered {
		trackFill = trackFill.Mix(theme.Ink, 0.1)
	}
	ui.elements[track].look = Look{Fill: trackFill, Radius: 9}
	ui.elements[knob].look = Look{Fill: theme.Surface, Radius: 7}
	ui.Describe(Description{Role: SwitchRole, Checked: *value})
	return ui.flipOnClick(result, value)
}

func (ui *Context) flipOnClick(result interaction, value *bool) bool {
	if result.hovered {
		ui.cursor = input.CursorHand
	}
	if !result.clicked {
		return false
	}
	*value = !*value
	return true
}

// sliderThumb is the slider's knob size.
const sliderThumb = 16

// Slider adds a track with a knob to drag between minimum and maximum, across the width it is
// given. Clicking the track jumps there; holding Shift while dragging moves ten times slower. It
// returns true on the steps the value changes.
func (ui *Context) Slider(name string, value *float32, minimum, maximum float32) bool {
	return ui.SliderStep(name, value, minimum, maximum, 0)
}

// SliderStep is Slider with the value kept to whole steps (0 = any value).
func (ui *Context) SliderStep(name string, value *float32, minimum, maximum, step float32) bool {
	theme := ui.Theme
	id := ui.idFor(name)
	fraction := float32(0)
	if maximum > minimum {
		fraction = clamp((*value-minimum)/(maximum-minimum), 0, 1)
	}
	var thumb int
	bar := func(fill draw2d.Color, share float32) {
		ui.Box(Layout{Width: Share(share), Height: Grow(), AlignY: Center}, func() {
			ui.Frame(Layout{Width: Grow(), Height: Fixed(4)}, Look{Fill: fill, Radius: 2}, nil)
		})
	}
	slider, result := ui.pressable(id, Layout{Width: Grow(), Height: Fixed(20), AlignY: Center}, Look{}, func() {
		bar(theme.Primary, fraction)
		thumb = ui.frameWithID(0, Layout{Width: Fixed(sliderThumb), Height: Fixed(sliderThumb)}, Look{}, nil)
		bar(theme.Line, 1-fraction)
	})
	thumbLook := Look{Fill: theme.Surface, Border: theme.Primary, BorderWidth: 1.5, Radius: sliderThumb / 2, Shadow: false}
	if result.hovered || ui.active == id {
		thumbLook.BorderWidth = 2.5
		ui.cursor = input.CursorHand
	}
	ui.elements[thumb].look = thumbLook
	ui.describeElement(slider, Description{Role: SliderRole, Label: displayText(name), Range: Range{Minimum: minimum, Maximum: maximum, Current: *value}})
	before := *value
	if steps := ui.readerSteps(id); steps != 0 {
		*value = snap(clamp(*value+steps*readerStep(minimum, maximum, step), minimum, maximum), minimum, maximum, step)
		return *value != before
	}
	if !result.held {
		return false
	}
	*value = ui.slideTo(id, result, *value, minimum, maximum)
	*value = snap(*value, minimum, maximum, step)
	return *value != before
}

// slideTo is where a dragged slider points: under the pointer, or, with Shift, the old value
// moved by a tenth of the pointer's movement.
func (ui *Context) slideTo(id ID, result interaction, value, minimum, maximum float32) float32 {
	area := ui.previous[id]
	travel := max(area.Width-sliderThumb, 1)
	if ui.state.Down(input.KeyShift) && !result.pressed {
		return value + float32(ui.state.CursorDeltaX)/travel*(maximum-minimum)*0.1
	}
	fraction := clamp((float32(ui.state.CursorX)-area.X-sliderThumb/2)/travel, 0, 1)
	return minimum + fraction*(maximum-minimum)
}

// snap keeps a value within its range and on whole steps from the minimum.
func snap(value, minimum, maximum, step float32) float32 {
	if step > 0 {
		value = minimum + float32(math.Round(float64((value-minimum)/step)))*step
	}
	return clamp(value, minimum, maximum)
}

// numberMemory is what a number field keeps: the movement of a drag not yet worth a whole step,
// whether the pointer moved during the press (a press that did not move is a click, which opens
// the number for typing), and the text being typed.
type numberMemory struct {
	carried float32
	moved   bool
	typed   string
}

// numberDragPixels is how far the pointer moves for one step while dragging a number.
const numberDragPixels = 4

// NumberField adds a number between minimum and maximum, shown with format (for example "%.0f nm"),
// with buttons to take a step off or add one. Drag the number sideways to change it a step every
// few pixels, or click it (or reach it with Tab) to type a value: Enter or a click elsewhere keeps
// it, Escape leaves the number as it was. It returns true on the steps the value changes.
func (ui *Context) NumberField(name string, value *float32, minimum, maximum, step float32, format string) bool {
	id := ui.idFor(name)
	memory := remember(ui, id, func() *numberMemory { return &numberMemory{} })
	before := *value
	editID := childID(id, "##edit")
	if ui.focus == editID {
		switch ui.editField(editID, &memory.typed, TextFieldOptions{}) {
		case fieldSubmitted, fieldLeft:
			if typed, ok := parseNumber(memory.typed); ok {
				*value = snap(typed, minimum, maximum, step)
			}
		}
		return *value != before
	}
	ui.focusables = append(ui.focusables, editID)
	memory.typed = strconv.FormatFloat(float64(*value), 'f', -1, 32)
	ui.stepNumber(id, name, memory, value, step, format)
	*value = snap(*value, minimum, maximum, step)
	return *value != before
}

// stepNumber is a number field when nobody is typing in it: the number between its buttons.
func (ui *Context) stepNumber(id ID, name string, memory *numberMemory, value *float32, step float32, format string) {
	theme := ui.Theme
	layout := Layout{Width: Grow(), Height: Fixed(theme.ControlHeight), Padding: PadXY(2, 0), AlignY: Center}
	ui.Frame(layout, Look{Fill: theme.Surface, Border: theme.ControlLine, BorderWidth: 1, Radius: theme.ControlRadius}, func() {
		ui.Push(name)
		if ui.FlatIconButton("minus", "") {
			*value -= step
		}
		valueLayout := Layout{Width: Grow(), Height: Grow(), AlignX: Center, AlignY: Center}
		_, drag := ui.pressable(childID(id, "##value"), valueLayout, Look{}, func() {
			ui.Text(fmt.Sprintf(format, *value), TextStyle{Weight: draw2d.Medium})
		})
		if drag.hovered || drag.held {
			ui.cursor = input.CursorResizeHorizontal
		}
		if drag.pressed {
			memory.moved = false
		}
		if drag.held {
			memory.moved = memory.moved || ui.state.CursorDeltaX != 0
			memory.carried += float32(ui.state.CursorDeltaX) / numberDragPixels
			steps := float32(math.Trunc(float64(memory.carried)))
			*value += steps * step
			memory.carried -= steps
		} else {
			memory.carried = 0
		}
		if drag.clicked && !memory.moved {
			ui.focus, ui.focusByKeyboard, ui.focusSeen = childID(id, "##edit"), true, true
		}
		if ui.FlatIconButton("plus", "") {
			*value += step
		}
		ui.Pop()
	})
}

// parseNumber reads the number at the start of a typed text, with a point or a comma before the
// decimals; what follows it (a unit) is ignored.
func parseNumber(text string) (float32, bool) {
	text = strings.ReplaceAll(strings.TrimSpace(text), ",", ".")
	var number float64
	if _, err := fmt.Sscanf(text, "%g", &number); err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, false
	}
	return float32(number), true
}

// readerStep is how far a screen reader moves a slider at a time: a step, or a twentieth of it all.
func readerStep(minimum, maximum, step float32) float32 {
	if step > 0 {
		return step
	}
	return (maximum - minimum) / 20
}

// Tabs adds a row of segments, one per label, with the selected one raised; clicking a segment
// selects it. It returns true on the step the selection changes.
func (ui *Context) Tabs(name string, selected *int, labels ...string) bool {
	theme := ui.Theme
	changed := false
	layout := Layout{Width: Grow(), Height: Fixed(theme.ControlHeight), Padding: Pad(2), Gap: 2, AlignY: Center}
	ui.Frame(layout, Look{Fill: theme.Sunken, Radius: theme.ControlRadius}, func() {
		ui.Push(name)
		for index, label := range labels {
			isSelected := index == *selected
			var text int
			tabLayout := Layout{Width: Share(1), Height: Grow(), AlignX: Center, AlignY: Center}
			tab, result := ui.pressable(ui.idFor(label), tabLayout, Look{}, func() {
				text = ui.addLeaf(ui.textElement(displayText(label), TextStyle{Weight: draw2d.Medium, Color: theme.InkSoft}))
			})
			ui.Describe(Description{Role: TabRole, Selected: isSelected})
			if isSelected {
				ui.elements[tab].look = Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: theme.ControlRadius - 1}
				ui.elements[text].style.Color = theme.Ink
			} else if result.hovered {
				ui.elements[text].style.Color = theme.Ink
				ui.cursor = input.CursorHand
			}
			if result.clicked && !isSelected {
				*selected, changed = index, true
			}
		}
		ui.Pop()
	})
	return changed
}
