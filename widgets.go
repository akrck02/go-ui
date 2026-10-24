package ui

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

type buttonVariant int

const (
	secondaryButton  buttonVariant = iota // a white button with a border
	primaryButton                         // the main action, filled with the primary colour
	flatButton                            // no fill or border until hovered: icons in title bars
	dangerButton                          // an action that destroys: red ink on a soft red fill
	dangerFlatButton                      // a flat button that destroys: red ink, the soft red fill when hovered
)

// dangerHover is how much more of Danger a destructive button takes under the pointer, and twice
// that while pressed.
const dangerHover = 0.08

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
// is a square button with the icon alone, named by the icon. A label of only a hidden part
// ("##Close") shows the icon alone too, and a screen reader reads it by that name.
func (ui *Context) IconButton(icon, label string) bool {
	name := label
	if name == "" {
		name = icon
	}
	clicked := ui.button(name, displayText(label), icon, secondaryButton)
	ui.nameWithoutWords(label)
	return clicked
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
	clicked := ui.button(name, displayText(label), icon, flatButton)
	ui.nameWithoutWords(label)
	return clicked
}

// DangerButton adds a button for an action that destroys something (delete, empty): red ink on
// a soft red fill, so it is never taken for an ordinary action.
func (ui *Context) DangerButton(label string) bool {
	return ui.button(label, displayText(label), "", dangerButton)
}

// DangerIconButton is DangerButton with an icon before its label; a label of only a hidden part
// ("##Delete") shows the icon alone, read by that name.
func (ui *Context) DangerIconButton(icon, label string) bool {
	clicked := ui.button(label, displayText(label), icon, dangerButton)
	ui.nameWithoutWords(label)
	return clicked
}

// DangerFlatIconButton is FlatIconButton for an action that destroys: its icon and words in red,
// on the soft red fill when hovered. For a delete beside each row of a list.
func (ui *Context) DangerFlatIconButton(icon, label string) bool {
	clicked := ui.button(label, displayText(label), icon, dangerFlatButton)
	ui.nameWithoutWords(label)
	return clicked
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
	if variant == flatButton || variant == dangerFlatButton {
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
		return fill, shadeBorder(fill, theme.Ink, result), theme.OnPrimary
	case dangerButton, dangerFlatButton:
		return ui.dangerColours(variant, result)
	case flatButton:
		if pressed {
			return theme.Line, draw2d.Transparent, theme.Ink
		}
		if result.hovered {
			return theme.Sunken, draw2d.Transparent, theme.Ink
		}
		return draw2d.Transparent, draw2d.Transparent, orColour(theme.FlatInk, theme.InkSoft)
	default:
		fill = theme.Surface
		if pressed {
			fill = theme.Line
		} else if result.hovered {
			fill = theme.Sunken
		}
		return fill, shadeBorder(theme.ControlLine, theme.Ink, result), theme.Ink
	}
}

// dangerColours are a destructive button's fill, border and ink: red ink always, the soft red
// fill (always, or under the pointer for a flat one), deeper when hovered and pressed.
func (ui *Context) dangerColours(variant buttonVariant, result interaction) (fill, border, ink draw2d.Color) {
	theme := ui.Theme
	tint := theme.dangerTint()
	fill = tint
	if variant == dangerFlatButton && !result.hovered {
		fill = draw2d.Transparent
	}
	switch {
	case result.held && result.hovered:
		fill = tint.Mix(theme.Danger, 2*dangerHover)
	case result.hovered:
		fill = tint.Mix(theme.Danger, dangerHover)
	}
	return fill, shadeBorder(fill, theme.Danger, result), theme.Danger
}

// How much deeper a button's border turns, towards a darker colour of its own family: a step
// under the pointer, two while pressed. The border answers the finger without changing its hue.
const (
	borderHoverShade = 0.15
	borderPressShade = 0.3
)

// shadeBorder is a button's border in its state: the border itself at rest, a shade deeper
// towards ink when hovered, and deeper still when pressed.
func shadeBorder(border, ink draw2d.Color, result interaction) draw2d.Color {
	switch {
	case result.held && result.hovered:
		return border.Mix(ink, borderPressShade)
	case result.hovered:
		return border.Mix(ink, borderHoverShade)
	}
	return border
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
