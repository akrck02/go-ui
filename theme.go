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
	ControlLine draw2d.Color // borders of fields, selects and secondary buttons: stronger than Line, so a control reads as one
	// FieldLine is the border of text and number fields and selects, when a design wants them
	// softer than buttons; left empty, they take ControlLine.
	FieldLine   draw2d.Color
	Primary     draw2d.Color // the main action
	OnPrimary   draw2d.Color // text on the main action
	PrimaryTint draw2d.Color // a light wash of the primary colour (a selected row)
	Positive    draw2d.Color
	Warning     draw2d.Color
	Danger      draw2d.Color
	FocusRing   draw2d.Color
	Selection   draw2d.Color // behind selected text
	Shadow      draw2d.Color
	Scrim       draw2d.Color // over everything behind a sheet, dimming it

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

	// What a design system tunes (docs/5-components/9-design-systems.md). The light theme's values
	// keep go-ui's own look, and a zero colour or size follows the token it names, so a theme that
	// only changes Primary or ControlRadius still looks the same.
	ButtonRadius      float32       // corners of buttons and selectable items; 0 = ControlRadius
	ButtonPadding     float32       // left and right inside a button with a label; 0 = 12
	ButtonGap         float32       // between a button's icon and its label; 0 = 6
	ButtonWeight      draw2d.Weight // the weight of a button's label
	FlatHeight        float32       // a flat button's height; 0 = a little under ControlHeight, for toolbars
	FlatPadding       float32       // left and right inside a flat button; 0 = 8
	FlatInk           draw2d.Color  // a flat button's label at rest; invisible = InkSoft
	PrimaryEdge       draw2d.Color  // a darker edge under the main button, a block it stands on
	EdgeWidth         float32       // how tall that edge is; 0 draws none
	Chosen            draw2d.Color  // behind a selected item of a list; invisible = PrimaryTint
	ChosenInk         draw2d.Color  // a selected item's label and icon; invisible = Primary
	ChosenWeight      draw2d.Weight // a selected item's label
	ItemPadding       float32       // left and right inside an item of a list; 0 = 8
	ItemGap           float32       // between an item's icon and its label; 0 = 8
	ItemIconInk       draw2d.Color  // an item's icon at rest; invisible = InkSoft
	LineHeight        float32       // a line of text as tall as its size × this; 0 keeps it as tall as its letters
	HeadingLineHeight float32       // the same for text at HeadingSize and above
	CheckSize         float32       // the side of a check box
	CheckMark         float32       // the size of its tick; 0 = 12
	CheckLine         draw2d.Color  // an empty check box's edge; invisible = InkQuiet
	CheckRadius       float32
	CheckBorder       float32      // the width of an empty check box's edge
	CheckedBorder     draw2d.Color // the edge of a ticked box; an invisible colour draws none
	// BackdropBlur blurs what is behind every translucent panel by this many layout pixels (a
	// standard deviation): frosted glass. Panels inside a translucent one are not blurred again; 0
	// blurs nothing, at no cost. (Look.Backdrop is another thing: a painter for a box's background.)
	BackdropBlur float32
	// FocusWidth is the width of the ring around the control Tab reached, lying on its edge with no
	// gap; 0 keeps go-ui's 3 pixels. On a button filled with Primary the ring is drawn inside it,
	// in OnPrimary, where FocusRing could not be told from the fill.
	FocusWidth float32
	// FieldFocusEdge lights a field's own edge in FocusRing while it is typed in, instead of a
	// ring around it.
	FieldFocusEdge bool
	// PressScale is how big a button is drawn while it is held, around its centre (0.97 gives a
	// little); its content shrinks with it and the layout does not move. 0 keeps it still.
	PressScale float32
}

// LightTheme is the default: a light, quiet look that reads well over any scene.
func LightTheme() Theme {
	return Theme{
		Background:    draw2d.MustHex("#f6f8fb"),
		Surface:       draw2d.MustHex("#ffffff"),
		Sunken:        draw2d.MustHex("#f1f4f8"),
		Ink:           draw2d.MustHex("#14202f"),
		InkSoft:       draw2d.MustHex("#4a5768"),
		InkQuiet:      draw2d.MustHex("#6b7889"),
		Line:          draw2d.MustHex("#dde3ea"),
		ControlLine:   draw2d.MustHex("#dde3ea"),
		Primary:       draw2d.MustHex("#1f5fc4"),
		OnPrimary:     draw2d.MustHex("#ffffff"),
		PrimaryTint:   draw2d.MustHex("#eaf1fd"),
		Positive:      draw2d.MustHex("#1e7d4f"),
		Warning:       draw2d.MustHex("#a86400"),
		Danger:        draw2d.MustHex("#c0392b"),
		FocusRing:     draw2d.MustHex("#1f5fc466"),
		Selection:     draw2d.MustHex("#1f5fc45c"),
		Shadow:        draw2d.MustHex("#14202f2e"),
		Scrim:         draw2d.MustHex("#14202f59"),
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

		ButtonWeight: draw2d.Medium,
		ChosenWeight: draw2d.Medium,
		CheckSize:    16,
		CheckRadius:  4,
		CheckBorder:  1.5,
	}
}

// DarkTheme is the light theme's tokens for a dark look, with the same sizes.
func DarkTheme() Theme {
	theme := LightTheme()
	theme.Background = draw2d.MustHex("#121a23")
	theme.Surface = draw2d.MustHex("#1b2633")
	theme.Sunken = draw2d.MustHex("#243241")
	theme.Ink = draw2d.MustHex("#e8edf3")
	theme.InkSoft = draw2d.MustHex("#b4c0ce")
	theme.InkQuiet = draw2d.MustHex("#8b98a8")
	theme.Line = draw2d.MustHex("#334355")
	theme.ControlLine = draw2d.MustHex("#334355")
	theme.Primary = draw2d.MustHex("#5b8fe8")
	theme.OnPrimary = draw2d.MustHex("#0d1620")
	theme.PrimaryTint = draw2d.MustHex("#22344d")
	theme.Positive = draw2d.MustHex("#4cc38a")
	theme.Warning = draw2d.MustHex("#e5a64a")
	theme.Danger = draw2d.MustHex("#ef7a6c")
	theme.FocusRing = draw2d.MustHex("#5b8fe880")
	theme.Selection = draw2d.MustHex("#5b8fe880")
	theme.Shadow = draw2d.MustHex("#00000066")
	theme.Scrim = draw2d.MustHex("#0000008c")
	return theme
}

// focusWidth is FocusWidth, or go-ui's own ring.
func (theme Theme) focusWidth() float32 { return orDefault(theme.FocusWidth, defaultFocusWidth) }

// buttonRadius is the corner of buttons and selectable items: ButtonRadius, or the controls' own.
func (theme Theme) buttonRadius() float32 { return orDefault(theme.ButtonRadius, theme.ControlRadius) }

// go-ui's own measures, kept when a theme leaves the matching field at zero.
const (
	defaultButtonPadding = 12
	defaultButtonGap     = 6
	flatButtonShorter    = 4
	defaultFlatPadding   = 8
	defaultItemPadding   = 8
	defaultItemGap       = 8
	defaultCheckMark     = 12
)

// orDefault is a size the theme sets, or go-ui's own when it leaves it at zero.
func orDefault(size, fallback float32) float32 {
	if size > 0 {
		return size
	}
	return fallback
}

// orColour is a colour the theme sets, or the token it follows when it leaves it invisible.
func orColour(colour, fallback draw2d.Color) draw2d.Color {
	if colour.IsVisible() {
		return colour
	}
	return fallback
}

func (theme Theme) buttonPadding() float32 {
	return orDefault(theme.ButtonPadding, defaultButtonPadding)
}
func (theme Theme) buttonGap() float32 { return orDefault(theme.ButtonGap, defaultButtonGap) }
func (theme Theme) flatHeight() float32 {
	return orDefault(theme.FlatHeight, theme.ControlHeight-flatButtonShorter)
}
func (theme Theme) flatPadding() float32 { return orDefault(theme.FlatPadding, defaultFlatPadding) }
func (theme Theme) itemPadding() float32 { return orDefault(theme.ItemPadding, defaultItemPadding) }
func (theme Theme) itemGap() float32     { return orDefault(theme.ItemGap, defaultItemGap) }

// chosen is what a selected item looks like: Chosen and ChosenInk, or the primary colour's tint.
func (theme Theme) chosen() (fill, ink draw2d.Color) {
	return orColour(theme.Chosen, theme.PrimaryTint), orColour(theme.ChosenInk, theme.Primary)
}

// lineBoxOf is how tall a line of text in style is laid out: its own line height when it has one,
// or the theme's.
func (theme Theme) lineBoxOf(style TextStyle, letters float32) float32 {
	if style.LineHeight == LettersLineHeight {
		return letters
	}
	if style.LineHeight > 0 {
		return max(letters, style.Size*style.LineHeight)
	}
	return theme.lineBox(style.Size, letters)
}

// lineBox is how tall a line of text of this size is laid out: the letters' own height, or the
// theme's line height when it asks for more.
func (theme Theme) lineBox(size, letters float32) float32 {
	multiple := theme.LineHeight
	if size >= theme.HeadingSize && theme.HeadingLineHeight > 0 {
		multiple = theme.HeadingLineHeight
	}
	return max(letters, size*multiple)
}

// fieldLine is the border of a field or a select: FieldLine when the theme gives one, ControlLine
// otherwise.
func (theme Theme) fieldLine() draw2d.Color {
	if theme.FieldLine.IsVisible() {
		return theme.FieldLine
	}
	return theme.ControlLine
}
