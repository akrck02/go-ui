package ui

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

type buttonVariant int

const (
	secondaryButton buttonVariant = iota // a white button with a border
	primaryButton                        // the main action, filled with the primary colour
	flatButton                           // no fill or border until hovered: icons in title bars
)

// Button adds a button and returns true on the frame it is clicked (pressed and let go over it).
func (ui *Context) Button(label string) bool {
	return ui.button(label, displayText(label), "", secondaryButton)
}

// PrimaryButton adds a button for the main action of a panel.
func (ui *Context) PrimaryButton(label string) bool {
	return ui.button(label, displayText(label), "", primaryButton)
}

// PrimaryIconButton is PrimaryButton with an icon (a Lucide name) before its label: the main action
// of a panel, such as adding a new thing.
func (ui *Context) PrimaryIconButton(icon, label string) bool {
	return ui.button(label, displayText(label), icon, primaryButton)
}

// IconButton adds a button with an icon (a Lucide name) before its label; with an empty label it
// is a square button with the icon alone, named by the icon.
func (ui *Context) IconButton(icon, label string) bool {
	name := label
	if name == "" {
		name = icon
	}
	return ui.button(name, displayText(label), icon, secondaryButton)
}

// FlatButton adds a button with no fill or border until hovered: menu bars and toolbars.
func (ui *Context) FlatButton(label string) bool {
	return ui.button(label, displayText(label), "", flatButton)
}

// FlatIconButton is FlatButton with an icon; with an empty label it is a square icon button.
func (ui *Context) FlatIconButton(icon, label string) bool {
	name := label
	if name == "" {
		name = icon
	}
	return ui.button(name, displayText(label), icon, flatButton)
}

// flatIconButton is a borderless icon button, for title bars.
func (ui *Context) flatIconButton(name, icon string) bool {
	return ui.button(name, "", icon, flatButton)
}

func (ui *Context) button(name, text, icon string, variant buttonVariant) bool {
	return ui.buttonWithID(ui.idFor(name), text, icon, variant, false).clicked
}

// buttonWithID adds a button with a known id; a lit button looks pressed (the button of an open
// menu).
func (ui *Context) buttonWithID(id ID, text, icon string, variant buttonVariant, lit bool) interaction {
	ui.joinTabOrder(id)
	if id == ui.active {
		ui.heldControl = id // only buttons shrink while held, as Bubble's .button:active does
	}
	result := ui.interact(id)
	if result.hovered {
		ui.cursor = input.CursorHand
	}
	shown := result
	if lit {
		shown.held, shown.hovered = true, true
	}
	theme := ui.Theme
	fill, border, ink := ui.buttonColours(variant, shown)
	height := theme.ControlHeight
	padding := PadXY(theme.buttonPadding(), 0)
	width := Fit().Between(height, 0)
	if text == "" {
		padding = Padding{}
		width = Fixed(height)
	}
	if variant == flatButton {
		height = theme.flatHeight()
		padding = PadXY(theme.flatPadding(), 0)
		if text == "" {
			width, padding = Fixed(height), Padding{}
		}
	}
	layout := Layout{Direction: Horizontal, Width: width, Height: Fixed(height), Padding: padding, Gap: theme.buttonGap(), AlignX: Center, AlignY: Center}
	look := Look{Fill: fill, Border: border, BorderWidth: 1, Radius: theme.buttonRadius()}
	if variant == primaryButton {
		look.Edge, look.EdgeWidth = theme.PrimaryEdge, theme.EdgeWidth
	}
	index := ui.frameWithID(id, layout, look, func() {
		if icon != "" {
			ui.IconSized(icon, theme.IconSize, ink)
		}
		if text != "" {
			ui.Text(text, TextStyle{Weight: theme.ButtonWeight, Color: ink})
		}
	})
	ui.elements[index].filledWithPrimary = variant == primaryButton
	return result
}

// buttonColours picks fill, border and ink for a variant in its state: darker while hovered,
// darker still while pressed.
func (ui *Context) buttonColours(variant buttonVariant, result interaction) (fill, border, ink draw2d.Color) {
	theme := ui.Theme
	pressed := result.held && result.hovered
	switch variant {
	case primaryButton:
		fill = theme.Primary
		if pressed {
			fill = fill.Mix(theme.Ink, 0.25)
		} else if result.hovered {
			fill = fill.Mix(theme.Ink, 0.12)
		}
		return fill, fill, theme.OnPrimary
	case flatButton:
		if pressed {
			return theme.Line, draw2d.Transparent, theme.Ink
		}
		if result.hovered {
			return theme.Sunken, draw2d.Transparent, theme.Ink
		}
		return draw2d.Transparent, draw2d.Transparent, orColour(theme.FlatInk, theme.InkSoft)
	default:
		fill, border = theme.Surface, theme.ControlLine
		if pressed {
			fill = theme.Line
		} else if result.hovered {
			fill = theme.Sunken
		}
		return fill, border, theme.Ink
	}
}

// Selectable adds an item with an optional icon that shows whether it is selected, and returns
// true on the step it is clicked. In a column it spans the width (the items of a side list); in a
// row it fits its content (a group of tools in a toolbar).
func (ui *Context) Selectable(icon, label string, selected bool) bool {
	id := ui.idFor(label)
	ui.joinTabOrder(id)
	result := ui.interact(id)
	if result.hovered {
		ui.cursor = input.CursorHand
	}
	theme := ui.Theme
	fill, ink, iconInk, weight := draw2d.Transparent, theme.Ink, orColour(theme.ItemIconInk, theme.InkSoft), draw2d.Regular
	switch {
	case selected:
		chosenFill, chosenInk := theme.chosen()
		fill, ink, iconInk, weight = chosenFill, chosenInk, chosenInk, theme.ChosenWeight
	case result.held && result.hovered:
		fill = theme.Line
	case result.hovered:
		fill = theme.Sunken
	}
	layout := Layout{Direction: Horizontal, Width: Grow(), Height: Fixed(theme.ControlHeight), Padding: PadXY(theme.itemPadding(), 0), Gap: theme.itemGap(), AlignY: Center}
	if ui.parentDirection() == Horizontal {
		layout.Width = Fit()
	}
	ui.frameWithID(id, layout, Look{Fill: fill, Radius: theme.buttonRadius()}, func() {
		if icon != "" {
			ui.IconSized(icon, theme.IconSize, iconInk)
		}
		ui.Text(displayText(label), TextStyle{Weight: weight, Color: ink})
	})
	ui.Describe(Description{Selected: selected})
	return result.clicked
}

// Looks are a box's looks while it is left alone, under the pointer and pressed.
type Looks struct {
	Normal, Hovered, Pressed Look
}

// Clickable adds a box that reacts to the mouse, with whatever content, and returns true on the
// step it is clicked: the building block for controls of your own (a table row, a card). Its look
// follows the pointer. It looks at the mouse after its content, so buttons inside it still work.
func (ui *Context) Clickable(name string, layout Layout, looks Looks, content func()) bool {
	id := ui.idFor(name)
	ui.joinTabOrder(id)
	index, result := ui.pressable(id, layout, looks.Normal, content)
	switch {
	case result.held && result.hovered:
		ui.elements[index].look = looks.Pressed
	case result.hovered:
		ui.elements[index].look = looks.Hovered
	}
	if result.hovered {
		ui.cursor = input.CursorHand
	}
	return result.clicked
}
