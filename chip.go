package ui

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// ChipLook is a chip's own colours: its fill and the ink of its text and icon. The zero value uses
// the theme's quiet colours (Sunken and InkSoft).
type ChipLook struct {
	Fill, Ink draw2d.Color
}

// chipHeight is a chip's height: smaller than a button, so chips sit inside rows and table cells.
const chipHeight = 22

// Chip adds a small rounded label with an optional icon (a Lucide name), in its own colours: a
// tag, a state, a filter. It returns true on the step it is clicked. Chips in a row stay on one
// line; to let them flow onto more lines, put them in a box with Layout.Wrap.
func (ui *Context) Chip(icon, label string, look ChipLook) bool {
	clicked, _ := ui.chip(icon, label, look, false)
	return clicked
}

// RemovableChip is a Chip with a small × after its label. It tells apart a click on the chip and
// a click on the ×, which asks the program to remove what the chip stands for.
func (ui *Context) RemovableChip(icon, label string, look ChipLook) (clicked, removed bool) {
	return ui.chip(icon, label, look, true)
}

func (ui *Context) chip(icon, label string, look ChipLook, removable bool) (clicked, removed bool) {
	look = ui.chipColours(look)
	id := ui.idFor(label)
	padding := Padding{Left: 8, Right: 8}
	if removable {
		padding.Right = 3
	}
	layout := Layout{Height: Fixed(chipHeight), Padding: padding, Gap: 4, AlignY: Center}
	ui.joinTabOrder(id)
	index, result := ui.pressable(id, layout, Look{}, func() {
		if icon != "" {
			ui.IconSized(icon, 12, look.Ink)
		}
		ui.Text(displayText(label), TextStyle{Size: ui.Theme.SmallTextSize + 1, Weight: draw2d.Medium, Color: look.Ink})
		if removable {
			removed = ui.chipRemoveButton(id, look.Ink)
		}
	})
	fill := look.Fill
	if result.hovered {
		fill = fill.Mix(look.Ink, 0.12)
		ui.cursor = input.CursorHand
	}
	ui.elements[index].look = Look{Fill: fill, Radius: chipHeight / 2}
	return result.clicked, removed
}

// chipRemoveButton is the × of a removable chip: a round button inside the chip, which it finishes
// before the chip does, so a click on it is the button's and not the chip's.
func (ui *Context) chipRemoveButton(chipID ID, ink draw2d.Color) bool {
	const size = 16
	layout := Layout{Width: Fixed(size), Height: Fixed(size), AlignX: Center, AlignY: Center}
	index, result := ui.pressable(childID(chipID, "##remove"), layout, Look{}, func() {
		ui.IconSized("x", 11, ink)
	})
	if result.hovered {
		ui.elements[index].look = Look{Fill: ink.WithAlpha(0.18), Radius: size / 2}
		ui.cursor = input.CursorHand
	}
	return result.clicked
}

// chipColours fills in the theme's colours for a chip look left at zero.
func (ui *Context) chipColours(look ChipLook) ChipLook {
	if !look.Fill.IsVisible() {
		look.Fill = ui.Theme.Sunken
	}
	if !look.Ink.IsVisible() {
		look.Ink = ui.Theme.InkSoft
	}
	return look
}
