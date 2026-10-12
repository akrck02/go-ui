package ui

import "github.com/akrck02/go-renderer/draw2d"

// A popup is a root that floats over everything else, next to an anchor rectangle: the list of a
// select, a tooltip. Popups are drawn above every panel and get the mouse before them.

// popupLayer is far above the panels' layers.
const popupLayer = 100000

// popup adds a floating box under the anchor (or above it when there is no room below), at least
// minimumWidth wide. A passive popup never takes the mouse (a tooltip).
func (ui *Context) popup(id ID, anchor draw2d.Rect, minimumWidth float32, passive bool, look Look, content func()) {
	outside := ui.open
	ui.open = nil
	layout := Layout{Direction: Vertical, Width: Fit().Between(minimumWidth, 0), Padding: Pad(4)}
	index := ui.openElement(element{id: id, kind: boxElement, layout: layout, look: look})
	ui.roots = append(ui.roots, root{id: id, element: index, popup: true, passive: passive, anchor: anchor})
	if !passive {
		ui.nextPopups = append(ui.nextPopups, id)
	}
	if content != nil {
		content()
	}
	ui.open = outside
}

// positionPopup puts a popup under its anchor, or above it when it would leave the window, and
// keeps it inside the window sideways.
func (ui *Context) positionPopup(placed *root) {
	element := &ui.elements[placed.element]
	const spacing = 4
	x := placed.anchor.X
	if placed.centred {
		x -= element.rect.Width / 2
	}
	x = clamp(x, spacing, ui.width-element.rect.Width-spacing)
	y := placed.anchor.Bottom() + spacing
	if y+element.rect.Height > ui.height-spacing && placed.anchor.Y-element.rect.Height-spacing >= 0 {
		y = placed.anchor.Y - element.rect.Height - spacing
	}
	element.rect.X, element.rect.Y = x, y
}

// popupLook is the look of menus and lists that open over the interface.
func (ui *Context) popupLook() Look {
	return Look{Fill: ui.Theme.Surface, Border: ui.Theme.Line, BorderWidth: 1, Radius: ui.Theme.ControlRadius + 2, Shadow: true, Clip: true}
}

// Tooltip explains the control added just before it: when the pointer rests on that control for
// a moment (Theme.TooltipDelay), a small dark note with the text shows under the pointer.
func (ui *Context) Tooltip(text string) {
	if text == "" || ui.interacted == 0 || ui.hot != ui.interacted || ui.lastHot != ui.hot {
		return
	}
	if waited := ui.now.Sub(ui.hotSince); waited < ui.Theme.TooltipDelay {
		ui.state.RequestFrame(ui.Theme.TooltipDelay - waited) // a window that draws on demand wakes to show it
		return
	}
	theme := ui.Theme
	x, y := float32(ui.state.CursorX), float32(ui.state.CursorY)
	anchor := draw2d.Rect{X: x + 4, Y: y + 12, Width: 0, Height: 0}
	look := Look{Fill: theme.Ink.WithAlpha(0.92), Radius: theme.ControlRadius}
	ui.popup(childID(ui.interacted, "##tooltip"), anchor, 0, true, look, func() {
		ui.Box(Layout{Padding: PadXY(4, 1)}, func() {
			ui.Text(text, TextStyle{Size: theme.SmallTextSize + 1, Color: theme.Surface})
		})
	})
}

// Toast shows a short note centred near the bottom of the window, over everything, without taking
// the mouse: what an action did, or (warning) why it could not be done. The program shows it for
// as long as it wants, as with a sheet; a few seconds is usual.
func (ui *Context) Toast(text string, warning bool) {
	if text == "" {
		return
	}
	theme := ui.Theme
	icon, iconColour := "info", theme.Surface
	if warning {
		icon, iconColour = "triangle-alert", theme.Warning
	}
	id := ui.idFor("##toast")
	outside := ui.open
	ui.open = nil
	layout := Layout{Direction: Horizontal, Padding: PadXY(12, 8), Gap: 8, AlignY: Center}
	index := ui.openElement(element{id: id, kind: boxElement, layout: layout, look: Look{Fill: theme.Ink.WithAlpha(0.94), Radius: theme.ControlRadius + 2, Shadow: true}})
	anchor := draw2d.Rect{X: ui.width / 2, Y: ui.height - 64 - 2*theme.TextSize}
	ui.roots = append(ui.roots, root{id: id, element: index, popup: true, passive: true, centred: true, anchor: anchor})
	ui.IconSized(icon, theme.TextSize, iconColour)
	ui.Text(text, TextStyle{Color: theme.Surface})
	ui.open = outside
}
