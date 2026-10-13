package ui

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// sheetLayer is above every panel and below the popups, so a select inside a sheet opens over it.
const sheetLayer = 50000

// SheetOptions size a sheet.
type SheetOptions struct {
	Width Sizing // the card's width; zero = 420 pixels
}

// sheetWidth is a sheet's width when the options leave it out: room for a sentence and a row of
// buttons.
const sheetWidth = 420

// sheetMemory is what a sheet keeps: the last frame it was shown, to tell when it opens.
type sheetMemory struct {
	lastFrame int
}

// Sheet shows a card in the middle of the window, over everything, with what is behind it dimmed:
// a question to confirm, a short form. While it is shown, nothing behind it reacts to the mouse or
// the keyboard, and the interface claims both, so cameras and shortcuts leave them alone too. Tab
// goes round the sheet's own fields. The program shows it with an if, for as long as it wants; it
// returns true on the step the person asks to close it (Escape, the system's back, the × in its
// title, or a click on the dimmed backdrop), and the program then stops showing it.
func (ui *Context) Sheet(title string, options SheetOptions, content func()) (dismissed bool) {
	id := ui.idFor(title)
	ui.dropOutsideFocusWhenOpening(id)
	escaped := ui.state.Pressed(input.KeyEscape) && ui.focus == 0 // a field or a list in the sheet takes Escape first
	escaped = escaped || (ui.frontSheet(id) && ui.Back())
	outside := ui.open
	ui.open = nil
	backdrop := Layout{Direction: Vertical, Width: Grow(), Height: Grow(), Padding: Pad(24), AlignX: Center, AlignY: Center}
	index := ui.openElement(element{id: id, kind: boxElement, layout: backdrop, look: Look{Fill: ui.Theme.Scrim}})
	ui.roots = append(ui.roots, root{id: id, element: index, sheet: true})
	ui.nextSheets = append(ui.nextSheets, id)
	ui.Push(title)
	closed := ui.sheetCard(id, title, options, content)
	ui.Pop()
	backdropClicked := ui.interact(id).clicked // after the card, so a click on the card is not on the backdrop
	ui.open = outside
	return escaped || closed || backdropClicked
}

// dropOutsideFocusWhenOpening takes the keyboard from a field behind a sheet on the step the sheet
// opens, so the letters typed next do not go to it.
func (ui *Context) dropOutsideFocusWhenOpening(id ID) {
	memory := remember(ui, id, func() *sheetMemory { return &sheetMemory{lastFrame: -1} })
	if memory.lastFrame != ui.frame-1 {
		ui.focus = 0
	}
	memory.lastFrame = ui.frame
}

// sheetCard is the card of a sheet: a title with a close button, and the content. It takes the
// mouse over its own background, so only a click outside it reaches the backdrop. It returns true
// when the close button is clicked.
func (ui *Context) sheetCard(id ID, title string, options SheetOptions, content func()) bool {
	theme := ui.Theme
	width := options.Width
	if width == (Sizing{}) {
		width = Fixed(sheetWidth)
	}
	cardID := childID(id, "##card")
	look := Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: theme.PanelRadius, Shadow: true, Clip: true}
	closed := false
	ui.frameWithID(cardID, Layout{Direction: Vertical, Width: width}, look, func() {
		titleLayout := Layout{Width: Grow(), Padding: Padding{theme.Padding, theme.Padding / 2, theme.Padding / 2, theme.Padding / 2}, Gap: theme.Gap, AlignY: Center}
		ui.Box(titleLayout, func() {
			ui.Text(displayText(title), TextStyle{Weight: draw2d.Bold})
			ui.Spacer()
			closed = ui.flatIconButton("##close", "x")
		})
		ui.Separator()
		start := len(ui.focusables)
		ui.Box(Layout{Direction: Vertical, Width: Grow(), Padding: Pad(theme.Padding), Gap: theme.Gap}, content)
		ui.sheetFocusStart, ui.sheetFocusEnd = start, len(ui.focusables)
	})
	ui.interact(cardID)
	return closed
}
