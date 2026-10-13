package main

import (
	"embed"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/opengl/overlay"
	ui "github.com/akrck02/go-ui"
)

// This file builds a few pieces in the way Android's own apps look (Material Design 3): a top app
// bar, a navigation bar, list items, a switch, chips, buttons, cards and floating buttons. They
// are made of go-ui boxes, texts and clickables, as any program can make controls of its own.

// palette holds the colour roles of Material Design 3's baseline scheme.
type palette struct {
	primary, onPrimary, primaryContainer, onPrimaryContainer draw2d.Color
	secondaryContainer, onSecondaryContainer                 draw2d.Color
	surface, onSurface, onSurfaceVariant                     draw2d.Color
	containerLow, container, containerHigh, containerHighest draw2d.Color
	outline, outlineVariant, scrim                           draw2d.Color
}

func lightPalette() palette {
	return palette{
		primary: draw2d.Hex("#6750a4"), onPrimary: draw2d.Hex("#ffffff"),
		primaryContainer: draw2d.Hex("#eaddff"), onPrimaryContainer: draw2d.Hex("#21005d"),
		secondaryContainer: draw2d.Hex("#e8def8"), onSecondaryContainer: draw2d.Hex("#1d192b"),
		surface: draw2d.Hex("#fef7ff"), onSurface: draw2d.Hex("#1d1b20"), onSurfaceVariant: draw2d.Hex("#49454f"),
		containerLow: draw2d.Hex("#f7f2fa"), container: draw2d.Hex("#f3edf7"),
		containerHigh: draw2d.Hex("#ece6f0"), containerHighest: draw2d.Hex("#e6e0e9"),
		outline: draw2d.Hex("#79747e"), outlineVariant: draw2d.Hex("#cac4d0"), scrim: draw2d.Hex("#fef7ffcc"),
	}
}

func darkPalette() palette {
	return palette{
		primary: draw2d.Hex("#d0bcff"), onPrimary: draw2d.Hex("#381e72"),
		primaryContainer: draw2d.Hex("#4f378b"), onPrimaryContainer: draw2d.Hex("#eaddff"),
		secondaryContainer: draw2d.Hex("#4a4458"), onSecondaryContainer: draw2d.Hex("#e8def8"),
		surface: draw2d.Hex("#141218"), onSurface: draw2d.Hex("#e6e0e9"), onSurfaceVariant: draw2d.Hex("#cac4d0"),
		containerLow: draw2d.Hex("#1d1b20"), container: draw2d.Hex("#211f26"),
		containerHigh: draw2d.Hex("#2b2930"), containerHighest: draw2d.Hex("#36343b"),
		outline: draw2d.Hex("#938f99"), outlineVariant: draw2d.Hex("#49454f"), scrim: draw2d.Hex("#141218cc"),
	}
}

// materialTheme gives go-ui's own controls (the slider, text fields) the palette and a size for
// fingers: controls 48 pixels tall, 16-pixel text, pill-shaped corners.
func materialTheme(colours palette, dark bool) ui.Theme {
	theme := ui.LightTheme()
	if dark {
		theme = ui.DarkTheme()
	}
	theme.Background, theme.Surface, theme.Sunken = colours.surface, colours.containerLow, colours.containerHighest
	theme.Ink, theme.InkSoft, theme.InkQuiet = colours.onSurface, colours.onSurfaceVariant, colours.outline
	theme.Line, theme.ControlLine = colours.outlineVariant, colours.outline
	theme.Primary, theme.OnPrimary, theme.PrimaryTint = colours.primary, colours.onPrimary, colours.secondaryContainer
	theme.FocusRing, theme.Selection = colours.primary, colours.primaryContainer
	theme.TextSize, theme.SmallTextSize, theme.HeadingSize, theme.IconSize = 16, 14, 22, 24
	theme.ControlHeight, theme.Gap, theme.Padding = 48, 12, 16
	theme.ControlRadius, theme.PanelRadius = 20, 16
	return theme
}

//go:embed fonts/Roboto-Regular.ttf fonts/Roboto-Medium.ttf fonts/Roboto-Bold.ttf
var robotoFiles embed.FS

// useRoboto draws every text in Roboto, Android's own typeface (SIL Open Font License, in fonts/).
// It must run before the interface is made.
func useRoboto() {
	regular, _ := robotoFiles.ReadFile("fonts/Roboto-Regular.ttf")
	medium, _ := robotoFiles.ReadFile("fonts/Roboto-Medium.ttf")
	bold, _ := robotoFiles.ReadFile("fonts/Roboto-Bold.ttf")
	overlay.UseFonts(overlay.Fonts{Regular: regular, Medium: medium, Bold: bold})
}

// material draws the pieces with one palette.
type material struct {
	context *ui.Context
	colours palette
}

func (look material) text(text string, size float32, weight draw2d.Weight, colour draw2d.Color) {
	look.context.Text(text, ui.TextStyle{Size: size, Weight: weight, Color: colour})
}

// pressedOver is the layer Material puts over something pressed: its content colour, faintly.
func pressedOver(colour draw2d.Color, strength float32) draw2d.Color {
	return draw2d.Color{colour[0], colour[1], colour[2], strength}
}

// topAppBar is the bar at the top, painted up under the status bar: an optional back arrow and a
// title. It returns true when the back arrow is touched.
func (look material) topAppBar(title string, back bool) bool {
	insets := look.context.Insets()
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(64 + insets.Top),
		Padding: ui.Padding{Left: 4, Top: insets.Top, Right: 16}, Gap: 4, AlignY: ui.Center}
	wentBack := false
	look.context.Frame(layout, ui.Look{Fill: look.colours.surface}, func() {
		if back {
			wentBack = look.iconButton("arrow-left", look.colours.onSurface)
		} else {
			look.context.Box(ui.Layout{Width: ui.Fixed(12)}, nil)
		}
		look.text(title, 22, draw2d.Regular, look.colours.onSurface)
	})
	return wentBack
}

// iconButton is a round 48-pixel target with a 24-pixel icon.
func (look material) iconButton(icon string, colour draw2d.Color) bool {
	layout := ui.Layout{Width: ui.Fixed(48), Height: ui.Fixed(48), AlignX: ui.Center, AlignY: ui.Center}
	looks := ui.Looks{Pressed: ui.Look{Fill: pressedOver(colour, 0.12), Radius: 24}}
	return look.context.Clickable(icon, layout, looks, func() { look.context.IconSized(icon, 24, colour) })
}

// destination is a place the navigation bar leads to.
type destination struct{ icon, label string }

// navigationBar is the bar at the bottom, painted down under the gesture bar: the selected
// destination has its icon on a pill.
func (look material) navigationBar(destinations []destination, selected *int) {
	insets := look.context.Insets()
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(80 + insets.Bottom),
		Padding: ui.Padding{Top: 12, Bottom: 16 + insets.Bottom}}
	look.context.Frame(layout, ui.Look{Fill: look.colours.container}, func() {
		for index, place := range destinations {
			if look.navigationItem(place, index == *selected) {
				*selected = index
			}
		}
	})
}

func (look material) navigationItem(place destination, selected bool) bool {
	colours := look.colours
	pill, ink, label := ui.Look{Radius: 16}, colours.onSurfaceVariant, colours.onSurfaceVariant
	if selected {
		pill.Fill, ink, label = colours.secondaryContainer, colours.onSecondaryContainer, colours.onSurface
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Share(1), Height: ui.Grow(), Gap: 4, AlignX: ui.Center}
	return look.context.Clickable("navigate "+place.label, layout, ui.Looks{}, func() {
		indicator := ui.Layout{Width: ui.Fixed(64), Height: ui.Fixed(32), AlignX: ui.Center, AlignY: ui.Center}
		look.context.Frame(indicator, pill, func() { look.context.IconSized(place.icon, 24, ink) })
		look.text(place.label, 12, draw2d.Medium, label)
	})
}

// listItem is a row of a list: a round icon, a headline and a supporting line, and whatever goes
// at its end (a switch). It returns true when the row is touched.
func (look material) listItem(name, icon, headline, supporting string, trailing func()) bool {
	colours := look.colours
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(72), Padding: ui.PadXY(16, 0), Gap: 16, AlignY: ui.Center}
	looks := ui.Looks{Pressed: ui.Look{Fill: pressedOver(colours.onSurface, 0.1)}}
	return look.context.Clickable(name, layout, looks, func() {
		avatar := ui.Layout{Width: ui.Fixed(40), Height: ui.Fixed(40), AlignX: ui.Center, AlignY: ui.Center}
		look.context.Frame(avatar, ui.Look{Fill: colours.primaryContainer, Radius: 20}, func() {
			look.context.IconSized(icon, 20, colours.onPrimaryContainer)
		})
		look.context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
			look.text(headline, 16, draw2d.Regular, colours.onSurface)
			look.text(supporting, 14, draw2d.Regular, colours.onSurfaceVariant)
		})
		if trailing != nil {
			trailing()
		}
	})
}

// toggle is Material's switch: a 52 × 32 track, its handle growing and moving to the end when on.
// It flips the value and returns true when touched.
func (look material) toggle(name string, on *bool) bool {
	colours := look.colours
	track := ui.Look{Fill: colours.containerHighest, Border: colours.outline, BorderWidth: 2, Radius: 16}
	handle, handleSize, align := colours.outline, float32(16), ui.Start
	if *on {
		track = ui.Look{Fill: colours.primary, Radius: 16}
		handle, handleSize, align = colours.onPrimary, 24, ui.End
	}
	layout := ui.Layout{Width: ui.Fixed(52), Height: ui.Fixed(32), Padding: ui.PadXY(4, 0), AlignX: align, AlignY: ui.Center}
	looks := ui.Looks{Normal: track, Hovered: track, Pressed: track}
	touched := look.context.Clickable(name, layout, looks, func() {
		size := ui.Fixed(handleSize)
		look.context.Frame(ui.Layout{Width: size, Height: size}, ui.Look{Fill: handle, Radius: handleSize / 2}, nil)
	})
	if touched {
		*on = !*on
	}
	return touched
}

// chip is a filter chip: outlined, or filled with a tick when selected.
func (look material) chip(label string, selected bool) bool {
	colours := look.colours
	normal := ui.Look{Border: colours.outlineVariant, BorderWidth: 1, Radius: 8}
	if selected {
		normal = ui.Look{Fill: colours.secondaryContainer, Radius: 8}
	}
	layout := ui.Layout{Height: ui.Fixed(32), Padding: ui.PadXY(12, 0), Gap: 8, AlignY: ui.Center}
	pressed := normal
	pressed.Fill = pressedOver(colours.onSurfaceVariant, 0.12)
	return look.context.Clickable("chip "+label, layout, ui.Looks{Normal: normal, Hovered: normal, Pressed: pressed}, func() {
		if selected {
			look.context.IconSized("check", 18, colours.onSecondaryContainer)
		}
		look.text(label, 14, draw2d.Medium, colours.onSurfaceVariant)
	})
}

// filledButton is the main action: a 40-pixel pill in the primary colour.
func (look material) filledButton(icon, label string) bool {
	colours := look.colours
	normal := ui.Look{Fill: colours.primary, Radius: 20}
	pressed := ui.Look{Fill: colours.primaryContainer, Radius: 20}
	layout := ui.Layout{Height: ui.Fixed(40), Padding: ui.PadXY(24, 0), Gap: 8, AlignY: ui.Center}
	ink := colours.onPrimary
	return look.context.Clickable("button "+label, layout, ui.Looks{Normal: normal, Hovered: normal, Pressed: pressed}, func() {
		if icon != "" {
			look.context.IconSized(icon, 18, ink)
		}
		look.text(label, 14, draw2d.Medium, ink)
	})
}

// outlinedButton is a second action next to the main one.
func (look material) outlinedButton(label string) bool {
	colours := look.colours
	normal := ui.Look{Border: colours.outline, BorderWidth: 1, Radius: 20}
	pressed := normal
	pressed.Fill = pressedOver(colours.primary, 0.12)
	layout := ui.Layout{Height: ui.Fixed(40), Padding: ui.PadXY(24, 0), AlignY: ui.Center}
	return look.context.Clickable("button "+label, layout, ui.Looks{Normal: normal, Hovered: normal, Pressed: pressed}, func() {
		look.text(label, 14, draw2d.Medium, colours.primary)
	})
}

// card is a filled container for one subject.
func (look material) card(content func()) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(16), Gap: 12}
	look.context.Frame(layout, ui.Look{Fill: look.colours.containerHighest, Radius: 12}, content)
}

// floatingButton is a square-cornered floating action button with an icon.
func (look material) floatingButton(icon string) bool {
	colours := look.colours
	normal := ui.Look{Fill: colours.primaryContainer, Radius: 16, Shadow: true}
	pressed := normal
	pressed.Fill = colours.secondaryContainer
	layout := ui.Layout{Width: ui.Fixed(56), Height: ui.Fixed(56), AlignX: ui.Center, AlignY: ui.Center}
	return look.context.Clickable("floating "+icon, layout, ui.Looks{Normal: normal, Hovered: normal, Pressed: pressed}, func() {
		look.context.IconSized(icon, 24, colours.onPrimaryContainer)
	})
}

// sectionTitle is a small heading over a group of rows.
func (look material) sectionTitle(text string) {
	look.context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: 16, Top: 16, Right: 16, Bottom: 8}}, func() {
		look.text(text, 14, draw2d.Medium, look.colours.primary)
	})
}
