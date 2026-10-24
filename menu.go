package ui

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// menuMemory is what a menu keeps: whether it is open, which item the keyboard is on (-1 = none),
// how many items it had in the last frame (for the arrows), and whether Enter was pressed this
// step.
type menuMemory struct {
	open      bool
	highlight int
	count     int
	choosing  bool
}

// menuMinimumWidth keeps a menu under a small button wide enough for its items.
const menuMinimumWidth = 180

// MenuButton adds a flat button with an optional icon (a Lucide name); clicking it opens a menu
// under it, over everything, whose items content adds with MenuItem (and Separator between
// groups). Click an item, or move with the arrows and press Enter, to choose it; choosing closes
// the menu, and so do Escape and a click anywhere else. With an empty label the button is a
// square with the icon alone, named by the icon.
func (ui *Context) MenuButton(icon, label string, content func()) {
	name := label
	if name == "" {
		name = icon
	}
	id := ui.idFor(name)
	memory := remember(ui, id, func() *menuMemory { return &menuMemory{highlight: -1} })
	result := ui.buttonWithID(id, displayText(label), icon, flatButton, memory.open)
	ui.nameWithoutWords(label)
	if result.clicked {
		memory.open, memory.highlight = !memory.open, -1
	}
	if !memory.open {
		if ui.focus == id {
			ui.focus = 0
		}
		return
	}
	ui.openMenu(id, memory, content)
}

// openMenu runs an open menu: closing it on a click elsewhere, following the keyboard and showing
// its items. While it is open it holds the keyboard, so Escape closes the menu and not a sheet
// around it.
func (ui *Context) openMenu(id ID, memory *menuMemory, content func()) {
	listID := childID(id, "##menu")
	button := ui.previous[id]
	x, y := float32(ui.state.CursorX), float32(ui.state.CursorY)
	if ui.state.ButtonPressed(input.MouseLeft) && ui.hoveredRoot != listID && !button.Contains(x, y) {
		memory.open = false
		return
	}
	ui.focus, ui.focusSeen = id, true
	ui.followMenuKeys(memory)
	if !memory.open {
		return
	}
	outerMenu, outerCount := ui.menu, ui.menuItems
	ui.menu, ui.menuItems = memory, 0
	ui.popup(listID, button, max(button.Width, menuMinimumWidth), false, ui.popupLook(), func() {
		ui.scopes = append(ui.scopes, listID) // items are named inside their menu
		content()
		ui.Pop()
	})
	memory.count, memory.choosing = ui.menuItems, false
	ui.menu, ui.menuItems = outerMenu, outerCount
}

// followMenuKeys moves the highlight with the arrows, closes with Escape and marks Enter, which
// the highlighted item answers when it is built.
func (ui *Context) followMenuKeys(memory *menuMemory) {
	state := ui.state
	switch {
	case state.Typed(input.KeyDown):
		memory.highlight = min(memory.highlight+1, memory.count-1)
	case state.Typed(input.KeyUp):
		memory.highlight = max(memory.highlight-1, 0)
	case state.Pressed(input.KeyEscape):
		memory.open = false
	case state.Pressed(input.KeyEnter):
		memory.choosing = true
	}
}

// MenuItem adds an item to the menu being built: an optional icon, a label and an optional hint
// on the right (a keyboard shortcut, say). It returns true on the step it is chosen, and closes
// the menu then. Outside a menu it is a plain row that returns true when clicked.
func (ui *Context) MenuItem(icon, label, hint string) bool {
	return ui.menuItem(icon, label, hint, ui.Theme.InkSoft, ui.Theme.Ink)
}

// DangerMenuItem is MenuItem for a choice that destroys something: its icon and words in red.
func (ui *Context) DangerMenuItem(icon, label, hint string) bool {
	return ui.menuItem(icon, label, hint, ui.Theme.Danger, ui.Theme.Danger)
}

// menuItem is a menu's row with its icon and words in the colours given.
func (ui *Context) menuItem(icon, label, hint string, iconInk, ink draw2d.Color) bool {
	theme := ui.Theme
	menu, position := ui.menu, ui.menuItems
	ui.menuItems++
	highlighted := menu != nil && menu.highlight == position
	layout := Layout{Width: Grow(), Height: Fixed(theme.ControlHeight), Padding: PadXY(8, 0), Gap: 8, AlignY: Center}
	row, result := ui.pressable(ui.idFor(label), layout, Look{}, func() {
		if icon != "" {
			ui.IconSized(icon, theme.IconSize, iconInk)
		} else {
			ui.Box(Layout{Width: Fixed(theme.IconSize)}, nil) // labels line up whether or not items have icons
		}
		ui.Text(displayText(label), TextStyle{Color: ink})
		ui.Spacer()
		if hint != "" {
			ui.Text(hint, TextStyle{Size: theme.SmallTextSize, Color: theme.InkQuiet, Weight: draw2d.Regular})
		}
	})
	if result.hovered {
		ui.cursor = input.CursorHand
		if menu != nil {
			menu.highlight = position // the pointer and the arrows share one highlight
		}
	}
	if result.hovered || highlighted {
		ui.elements[row].look = Look{Fill: theme.Sunken, Radius: theme.ControlRadius}
	}
	chosen := result.clicked || (highlighted && menu.choosing)
	if chosen && menu != nil {
		menu.open = false
	}
	return chosen
}
