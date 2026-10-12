package ui

import (
	"time"

	"github.com/akrck02/go-renderer/draw2d"
)

// Theme holds the tokens every control reads its colours, sizes and corners from, so an interface
// changes look in one place. Sizes are layout pixels.
type Theme struct {
	Background  draw2d.Color // behind everything, in a window with no scene (see package window)
	Surface     draw2d.Color // panels and secondary buttons
	Sunken      draw2d.Color // hovered secondary controls, tracks, wells
	Ink         draw2d.Color // body text
	InkSoft     draw2d.Color // secondary text and icons
	InkQuiet    draw2d.Color // hints and captions
	Line        draw2d.Color // borders and separators
	Primary     draw2d.Color // the main action
	OnPrimary   draw2d.Color // text on the main action
	PrimaryTint draw2d.Color // a light wash of the primary colour (a selected row)
	Positive    draw2d.Color
	Warning     draw2d.Color
	Danger      draw2d.Color
	FocusRing   draw2d.Color
	Selection   draw2d.Color // behind selected text
	Shadow      draw2d.Color

	TextSize      float32
	SmallTextSize float32
	HeadingSize   float32
	IconSize      float32
	ControlHeight float32 // buttons, fields, selects
	Gap           float32 // between controls
	Padding       float32 // inside panels
	ControlRadius float32
	PanelRadius   float32
	PanelMinWidth float32
	ShadowBlur    float32
	ShadowDrop    float32
	ScrollStep    float32       // pixels a scroll area moves per notch of the wheel
	TooltipDelay  time.Duration // how long the pointer rests on a control before its tooltip shows
}

// LightTheme is the default: a light, quiet look that reads well over any scene.
func LightTheme() Theme {
	return Theme{
		Background:    draw2d.Hex("#f6f8fb"),
		Surface:       draw2d.Hex("#ffffff"),
		Sunken:        draw2d.Hex("#f1f4f8"),
		Ink:           draw2d.Hex("#14202f"),
		InkSoft:       draw2d.Hex("#4a5768"),
		InkQuiet:      draw2d.Hex("#6b7889"),
		Line:          draw2d.Hex("#dde3ea"),
		Primary:       draw2d.Hex("#1f5fc4"),
		OnPrimary:     draw2d.Hex("#ffffff"),
		PrimaryTint:   draw2d.Hex("#eaf1fd"),
		Positive:      draw2d.Hex("#1e7d4f"),
		Warning:       draw2d.Hex("#a86400"),
		Danger:        draw2d.Hex("#c0392b"),
		FocusRing:     draw2d.Hex("#1f5fc466"),
		Selection:     draw2d.Hex("#1f5fc45c"),
		Shadow:        draw2d.Hex("#14202f2e"),
		TextSize:      13,
		SmallTextSize: 11,
		HeadingSize:   15,
		IconSize:      16,
		ControlHeight: 28,
		Gap:           8,
		Padding:       12,
		ControlRadius: 6,
		PanelRadius:   10,
		PanelMinWidth: 200,
		ShadowBlur:    10,
		ShadowDrop:    3,
		ScrollStep:    40,
		TooltipDelay:  500 * time.Millisecond,
	}
}

// DarkTheme is the light theme's tokens for a dark look, with the same sizes.
func DarkTheme() Theme {
	theme := LightTheme()
	theme.Background = draw2d.Hex("#121a23")
	theme.Surface = draw2d.Hex("#1b2633")
	theme.Sunken = draw2d.Hex("#243241")
	theme.Ink = draw2d.Hex("#e8edf3")
	theme.InkSoft = draw2d.Hex("#b4c0ce")
	theme.InkQuiet = draw2d.Hex("#8b98a8")
	theme.Line = draw2d.Hex("#334355")
	theme.Primary = draw2d.Hex("#5b8fe8")
	theme.OnPrimary = draw2d.Hex("#0d1620")
	theme.PrimaryTint = draw2d.Hex("#22344d")
	theme.Positive = draw2d.Hex("#4cc38a")
	theme.Warning = draw2d.Hex("#e5a64a")
	theme.Danger = draw2d.Hex("#ef7a6c")
	theme.FocusRing = draw2d.Hex("#5b8fe880")
	theme.Selection = draw2d.Hex("#5b8fe880")
	theme.Shadow = draw2d.Hex("#00000066")
	return theme
}
