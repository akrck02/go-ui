package ui

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// selectMemory is what a select keeps: whether its list is open, and which option the keyboard is
// on.
type selectMemory struct {
	open      bool
	highlight int
}

// selectListHeight is the tallest a select's list grows before it scrolls.
const selectListHeight = 260

// Select adds a field showing the selected option; clicking it opens the list of options over
// everything. Click an option, or move with the arrows and press Enter, to choose it; Escape or a
// click elsewhere closes the list. It returns true on the step the selection changes.
func (ui *Context) Select(name string, selected *int, options []string) bool {
	theme := ui.Theme
	id := ui.idFor(name)
	memory := remember(ui, id, func() *selectMemory { return &selectMemory{} })
	shown := ""
	if *selected >= 0 && *selected < len(options) {
		shown = options[*selected]
	}
	layout := Layout{Width: Grow(), Height: Fixed(theme.ControlHeight), Padding: PadXY(10, 0), Gap: 8, AlignY: Center}
	ui.joinTabOrder(id)
	field, result := ui.pressable(id, layout, Look{}, func() {
		ui.Text(shown, TextStyle{})
		ui.Spacer()
		ui.IconSized("chevron-down", 14, theme.InkSoft)
	})
	look := Look{Fill: theme.Surface, Border: theme.fieldLine(), BorderWidth: 1, Radius: theme.ControlRadius}
	if result.hovered {
		look.Border = theme.fieldLine().Mix(theme.Ink, borderHoverShade)
		ui.cursor = input.CursorHand
	}
	if memory.open {
		look.Border = theme.Primary
	}
	ui.elements[field].look = look
	if result.clicked {
		memory.open, memory.highlight = !memory.open, max(*selected, 0)
	}
	if !memory.open {
		if ui.focus == id {
			ui.focus = 0
		}
		return false
	}
	return ui.openSelect(id, memory, selected, options)
}

// openSelect runs an open select: closing it on a click elsewhere, following the keyboard and
// showing its list.
func (ui *Context) openSelect(id ID, memory *selectMemory, selected *int, options []string) bool {
	listID := childID(id, "##list")
	field := ui.previous[id]
	x, y := float32(ui.state.CursorX), float32(ui.state.CursorY)
	if ui.state.ButtonPressed(input.MouseLeft) && ui.hoveredRoot != listID && !field.Contains(x, y) {
		memory.open = false
		return false
	}
	ui.focus, ui.focusSeen = id, true
	changed := ui.followSelectKeys(memory, selected, len(options))
	if !memory.open {
		return changed
	}
	ui.popup(listID, field, field.Width, false, ui.popupLook(), func() {
		ui.Push("##options")
		ui.ScrollArea("##scroll", Layout{Width: Grow(), Height: Fit().Between(0, selectListHeight), Gap: 0}, func() {
			for index, option := range options {
				if ui.selectOption(option, index == *selected, index == memory.highlight) {
					changed = changed || *selected != index
					*selected = index
					memory.open = false
				}
			}
		})
		ui.Pop()
	})
	return changed
}

// followSelectKeys moves the highlight with the arrows, chooses with Enter and closes with Escape.
func (ui *Context) followSelectKeys(memory *selectMemory, selected *int, count int) bool {
	state := ui.state
	switch {
	case state.Typed(input.KeyDown):
		memory.highlight = min(memory.highlight+1, count-1)
	case state.Typed(input.KeyUp):
		memory.highlight = max(memory.highlight-1, 0)
	case state.Pressed(input.KeyEscape):
		memory.open = false
	case state.Pressed(input.KeyEnter) && count > 0:
		changed := *selected != memory.highlight
		*selected, memory.open = memory.highlight, false
		return changed
	}
	return false
}

// selectOption is one row of an open select: highlighted under the pointer or the keyboard, with a
// tick on the selected option.
func (ui *Context) selectOption(option string, isSelected, isHighlighted bool) bool {
	theme := ui.Theme
	layout := Layout{Width: Grow(), Height: Fixed(theme.ControlHeight), Padding: PadXY(8, 0), Gap: 8, AlignY: Center}
	row, result := ui.pressable(ui.idFor(option), layout, Look{}, func() {
		weight := draw2d.Regular
		if isSelected {
			weight = draw2d.Medium
		}
		ui.Text(displayText(option), TextStyle{Weight: weight})
		ui.Spacer()
		if isSelected {
			ui.IconSized("check", 14, theme.Primary)
		} else {
			ui.Box(Layout{Width: Fixed(14)}, nil)
		}
	})
	if result.hovered || isHighlighted {
		ui.elements[row].look = Look{Fill: theme.Sunken, Radius: theme.ControlRadius}
	}
	if result.hovered {
		ui.cursor = input.CursorHand
	}
	return result.clicked
}
