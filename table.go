package ui

import (
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// Column is one column of a table: its title, its width and where its cells sit.
type Column struct {
	Title string
	// Width is the column's width. Use Fixed for narrow columns and Share for the rest: shares
	// start from nothing, so the header and every row come out the same.
	Width Sizing
	Align Align // Start for words, End for numbers
	// Weight is the face of its cells: draw2d.Mono for a column of codes.
	Weight draw2d.Weight
}

// Table is a table being built, handed to the function that adds its rows. Rows and group
// captions out of sight are only counted, not built, so a table of a few thousand rows costs what
// the rows in sight do.
type Table struct {
	ui       *Context
	columns  []Column
	captions map[string]string // group captions in capitals, kept so they are not made every frame
	// top is how far down the next row goes, counting every row so far; visibleTop and
	// visibleBottom are the part of the rows in sight; skipped is the height of rows out of
	// sight not yet stood in for by an empty box.
	top, visibleTop, visibleBottom, skipped float32
}

// tableMemory is what a table keeps between frames, besides its scroll position.
type tableMemory struct {
	captions map[string]string
}

// groupRowHeight is the height of a group caption, a little less than a row.
const groupRowHeight = 26

// accentWidth is the strip on the left of every row that the selected row fills.
const accentWidth = 3

// Table adds a table with a header and rows that scroll under it: rows adds the rows, with
// table.Row, and table.Group for a caption between groups of rows. The layout needs a height
// that does not follow its content (Grow, Fixed, Percent) for the rows to scroll; with Fit every
// row is built.
func (ui *Context) Table(name string, layout Layout, columns []Column, rows func(table *Table)) {
	theme := ui.Theme
	id := ui.idFor(name)
	bodyID := childID(id, "##rows")
	scroll := remember(ui, bodyID, func() *scrollMemory { return &scrollMemory{thumbID: childID(bodyID, "##thumb")} })
	memory := remember(ui, childID(id, "##table"), func() *tableMemory { return &tableMemory{captions: map[string]string{}} })
	view := scroll.view
	if view <= 0 {
		view = ui.height // the first frame: the table's height is not known yet
	}
	outer := ui.table
	ui.table = Table{ui: ui, columns: columns, captions: memory.captions, visibleTop: scroll.offset, visibleBottom: scroll.offset + view}
	layout.Direction, layout.Gap, layout.Columns = Vertical, 0, 0
	look := Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: theme.ControlRadius + 2, Clip: true}
	ui.frameWithID(id, layout, look, func() {
		ui.tableHeader(columns)
		ui.Separator()
		ui.Push(name)
		ui.scrollArea(bodyID, Layout{Width: Grow(), Height: Grow()}, func() {
			rows(&ui.table)
			ui.table.standInForSkipped()
		})
		ui.Pop()
	})
	ui.table = outer
}

// tableHeader is the row of column titles, which stays while the rows scroll.
func (ui *Context) tableHeader(columns []Column) {
	theme := ui.Theme
	ui.Frame(Layout{Width: Grow(), Height: Fixed(groupRowHeight), AlignY: Center}, Look{Fill: theme.Sunken}, func() {
		ui.Box(Layout{Width: Fixed(accentWidth)}, nil)
		for _, column := range columns {
			ui.tableCell(column, column.Title, TextStyle{Size: theme.SmallTextSize + 1, Weight: draw2d.Medium, Color: theme.InkSoft})
		}
	})
}

// tableCell is one cell: its column's width, its text cut with "…" when it does not fit.
// leadingCell is a row's first cell with its icon, indent and second line.
func (ui *Context) leadingCell(column Column, text string, look RowLook) {
	theme := ui.Theme
	layout := Layout{Width: column.Width, Height: Grow(), Padding: Padding{Left: 8 + look.Indent, Right: 8}, Gap: 10, AlignY: Center}
	ui.Frame(layout, Look{Clip: true}, func() {
		if look.Icon != "" {
			ink := look.IconInk
			if !ink.IsVisible() {
				ink = theme.InkSoft
			}
			size := theme.ControlHeight * 0.66
			ui.Frame(Layout{Width: Fixed(size), Height: Fixed(size), AlignX: Center, AlignY: Center}, Look{Fill: look.Tile, Radius: theme.ControlRadius * 0.7}, func() {
				ui.IconSized(look.Icon, size*0.58, ink)
			})
		}
		ui.Box(Layout{Direction: Vertical, Gap: 1}, func() {
			ui.Text(text, TextStyle{Weight: column.Weight})
			if look.Secondary != "" {
				ui.Text(look.Secondary, TextStyle{Size: theme.SmallTextSize, Weight: look.SecondaryWeight, Color: theme.InkQuiet})
			}
		})
	})
}

func (ui *Context) tableCell(column Column, text string, style TextStyle) {
	if style.Weight == draw2d.Regular {
		style.Weight = column.Weight
	}
	layout := Layout{Width: column.Width, Height: Grow(), Padding: PadXY(8, 0), AlignX: column.Align, AlignY: Center}
	ui.Frame(layout, Look{Clip: true}, func() { ui.Text(text, style) })
}

// inSight counts a row of a height and tells whether any of it can be seen. A row in sight first
// gets an empty box for the rows skipped above it, so it lands where it belongs.
func (table *Table) inSight(height float32) bool {
	top := table.top
	table.top += height
	if top+height <= table.visibleTop || top >= table.visibleBottom {
		table.skipped += height
		return false
	}
	table.standInForSkipped()
	return true
}

// standInForSkipped adds an empty box as tall as the rows skipped since the last one built.
func (table *Table) standInForSkipped() {
	if table.skipped <= 0 {
		return
	}
	table.ui.Box(Layout{Width: Grow(), Height: Fixed(table.skipped)}, nil)
	table.skipped = 0
}

// Group adds a caption across all the columns, in small capitals, above the rows of a group.
func (table *Table) Group(caption string) {
	if !table.inSight(groupRowHeight) {
		return
	}
	ui := table.ui
	shown, known := table.captions[caption]
	if !known {
		shown = strings.ToUpper(displayText(caption))
		table.captions[caption] = shown
	}
	layout := Layout{Width: Grow(), Height: Fixed(groupRowHeight), Padding: Padding{Left: accentWidth + 8, Top: 4}, AlignY: Center}
	ui.Box(layout, func() {
		ui.Text(shown, TextStyle{Size: ui.Theme.SmallTextSize, Weight: draw2d.Bold, Color: ui.Theme.InkQuiet})
	})
}

// Row adds a row with a text per column, named by id (unique in the table: the item's key), and
// returns true on the step it is clicked. A selected row is tinted, with a strip of the primary
// colour on its left. Cells beyond the columns are ignored; missing ones are empty.
func (table *Table) Row(id string, selected bool, cells ...string) bool {
	return table.RowWith(id, selected, RowLook{}, cells...)
}

// RowLook is how a row's first cell starts: an icon on a tinted tile, and an indent for a row that
// belongs to the one above it (an attribute under its type).
type RowLook struct {
	Icon            string
	IconInk, Tile   draw2d.Color // the icon's colour and its tile's; zero = the theme's soft ink, no tile
	Indent          float32
	Secondary       string // a quieter second line under the first cell's text
	SecondaryWeight draw2d.Weight
}

// RowWith adds a row as Row does, its first cell starting as look says.
func (table *Table) RowWith(id string, selected bool, look RowLook, cells ...string) bool {
	ui := table.ui
	if !table.inSight(ui.Theme.ControlHeight) {
		return false
	}
	theme := ui.Theme
	var accent int
	layout := Layout{Width: Grow(), Height: Fixed(theme.ControlHeight), AlignY: Center}
	row, result := ui.pressable(ui.idFor(id), layout, Look{}, func() {
		accent = ui.frameWithID(0, Layout{Width: Fixed(accentWidth), Height: Grow()}, Look{}, nil)
		for index, column := range table.columns {
			text := ""
			if index < len(cells) {
				text = cells[index]
			}
			if index == 0 && (look.Icon != "" || look.Indent > 0 || look.Secondary != "") {
				ui.leadingCell(column, text, look)
				continue
			}
			ui.tableCell(column, text, TextStyle{})
		}
	})
	switch {
	case selected:
		ui.elements[row].look = Look{Fill: theme.PrimaryTint}
		ui.elements[accent].look = Look{Fill: theme.Primary}
	case result.held && result.hovered:
		ui.elements[row].look = Look{Fill: theme.Line}
	case result.hovered:
		ui.elements[row].look = Look{Fill: theme.Sunken}
	}
	if result.hovered {
		ui.cursor = input.CursorHand
	}
	return result.clicked
}
