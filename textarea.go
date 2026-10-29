package ui

import (
	"sort"
	"unicode"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// A text area is a text field of many lines: the same editor, keys, clipboard and on-screen
// keyboard, over a text that keeps its line breaks and wraps at the field's width. It differs
// where one line and many must: Enter starts a new line instead of ending the edit; Up, Down,
// Home and End move by the lines shown; Escape leaves the field with its text as it is (a page of
// writing is never thrown away by one key); and the text scrolls up and down, by the wheel, a
// finger or the caret, instead of sideways.

// TextAreaOptions shape a text area.
type TextAreaOptions struct {
	Placeholder string // shown, quiet, while the area is empty
	Label       string // what a screen reader calls the area; the placeholder when empty
	MaxLength   int    // the most characters it takes, line breaks included; 0 = no limit
	Width       Sizing // the zero value grows across the row or column
	// Height is the area's; the zero value is tall enough for four lines (textAreaLines). Grow() fills
	// what the column leaves, for a page that is all writing.
	Height Sizing
}

// textAreaLines is how many lines a text area shows when its options give no height.
const textAreaLines = 4

// textAreaLineHeight is a text area's line as tall as its text's size × this: looser than a
// label's, for reading many lines.
const textAreaLineHeight = 1.45

// TextArea adds a field of many lines to type text in, with the keys of TextField, Enter starting
// a new line, and Up and Down moving between lines. It returns true on the steps the text changes.
func (ui *Context) TextArea(name string, text *string, options TextAreaOptions) bool {
	before := *text
	ui.editField(ui.idFor(name), text, TextFieldOptions{Placeholder: options.Placeholder, Label: options.Label,
		MaxLength: options.MaxLength, Width: options.Width, Height: options.Height, multiLine: true})
	return *text != before
}

// textLine is one line of a text area as it is shown: the characters from start up to end. The
// next line starts at end when the line wrapped, or after the line break at end.
type textLine struct{ start, end int }

// wrapKey is what a text area's lines depend on: they are wrapped again only when it changes.
type wrapKey struct {
	text        string
	width       float32
	size, scale float32
	weight      draw2d.Weight
}

// wrapArea breaks a text area's text into the lines it shows at a width, unless it already has.
func (ui *Context) wrapArea(memory *fieldMemory, text string, width float32) {
	key := wrapKey{text: text, width: width, size: ui.fieldTextSize(memory.size), scale: ui.scale, weight: memory.weight}
	if memory.wrapValid && memory.wrapped == key {
		return
	}
	runes := []rune(text)
	memory.lines = ui.wrapText(memory.lines[:0], runes, width, memory)
	memory.lineTexts = memory.lineTexts[:0]
	for _, line := range memory.lines {
		memory.lineTexts = append(memory.lineTexts, string(runes[line.start:line.end]))
	}
	memory.wrapped, memory.wrapValid = key, true
}

// wrapText breaks a text into lines: at each line break, and where a line would be wider than
// width. There is always one line, if empty, and one after a final line break.
func (ui *Context) wrapText(lines []textLine, runes []rune, width float32, memory *fieldMemory) []textLine {
	start := 0
	for {
		end := start
		for end < len(runes) && runes[end] != '\n' {
			end++
		}
		lines = ui.wrapParagraph(lines, runes, start, end, width, memory)
		if end == len(runes) {
			return lines
		}
		start = end + 1
	}
}

// wrapParagraph breaks the characters from start to end, which hold no line break, into lines no
// wider than width: between words, the spaces after a word staying on its line, and inside a word
// only when the word alone is wider than the line.
func (ui *Context) wrapParagraph(lines []textLine, runes []rune, start, end int, width float32, memory *fieldMemory) []textLine {
	measure := func(from, to int) float32 { return ui.textWidth(string(runes[from:to]), memory.weight, memory.size) }
	lineStart, lineWidth := start, float32(0)
	for position := start; position < end; {
		wordEnd := position
		for wordEnd < end && !unicode.IsSpace(runes[wordEnd]) {
			wordEnd++
		}
		spaceEnd := wordEnd
		for spaceEnd < end && unicode.IsSpace(runes[spaceEnd]) {
			spaceEnd++
		}
		wordWidth := measure(position, wordEnd)
		switch {
		case lineWidth+wordWidth > width && position > lineStart:
			lines = append(lines, textLine{lineStart, position}) // the word starts the next line
			lineStart, lineWidth = position, 0
		case wordWidth > width:
			cut := ui.mostThatFit(runes, position, wordEnd, width, memory)
			lines = append(lines, textLine{lineStart, cut})
			lineStart, position = cut, cut
		default:
			lineWidth += measure(position, spaceEnd)
			position = spaceEnd
		}
	}
	return append(lines, textLine{lineStart, end})
}

// mostThatFit is where a word from start to end is cut so its first part fits in width: always
// after one character at least, so a very narrow field still moves on.
func (ui *Context) mostThatFit(runes []rune, start, end int, width float32, memory *fieldMemory) int {
	fitting := sort.Search(end-start, func(count int) bool {
		return ui.textWidth(string(runes[start:start+count+1]), memory.weight, memory.size) > width
	})
	return start + max(fitting, 1)
}

// rewrapArea breaks the text being edited into lines again at the width they were last wrapped
// at, if it changed since: an on-screen keyboard's edit, or the letters typed in this step, came
// before the keys that move by lines, and an area given the keyboard before it was ever shown has
// no lines yet. Positions then always fall inside the lines.
func (ui *Context) rewrapArea(memory *fieldMemory) {
	ui.wrapArea(memory, memory.editor.String(), memory.wrapped.width)
}

// lineOf is the shown line a position of the text is on. A position where a line wrapped is shown
// at the start of the next line, as in every editor.
func (memory *fieldMemory) lineOf(position int) int {
	after := sort.Search(len(memory.lines), func(index int) bool { return memory.lines[index].start > position })
	return max(after-1, 0)
}

// caretEnd is the furthest the caret goes on a shown line: its end, or, on a line that wrapped,
// before the space it wrapped at, since the position at its end is the next line's start.
func (memory *fieldMemory) caretEnd(index int) int {
	line := memory.lines[index]
	wrapped := index+1 < len(memory.lines) && memory.lines[index+1].start == line.end
	if wrapped && line.end > line.start && unicode.IsSpace(memory.editor.text[line.end-1]) {
		return line.end - 1
	}
	return line.end
}

// areaLineHeight is how tall each line of a text area is.
func (ui *Context) areaLineHeight(memory *fieldMemory) float32 {
	size := ui.scaledText(ui.fieldTextSize(memory.size))
	letters := ui.measure("Hg", TextStyle{Size: size}).Height()
	return max(ui.Theme.lineBox(size, letters), size*textAreaLineHeight)
}

// areaPoint is where a position of a text area's text is, from the top left of its text: the
// caret's x and the top of its line.
func (ui *Context) areaPoint(memory *fieldMemory, position int) (x, top float32) {
	index := memory.lineOf(position)
	line := memory.lines[index]
	position = clampInt(position, line.start, line.end)
	x = ui.textWidth(string(memory.editor.text[line.start:position]), memory.weight, memory.size)
	return x, float32(index) * ui.areaLineHeight(memory)
}

// areaPositionAt is the position nearest to a point given from the top left of a text area's
// text: on the line under it, nearest to its x.
func (ui *Context) areaPositionAt(memory *fieldMemory, x, y float32) int {
	index := clampInt(int(y/ui.areaLineHeight(memory)), 0, len(memory.lines)-1)
	start := memory.lines[index].start
	return start + ui.nearestPosition(memory, memory.editor.text[start:memory.caretEnd(index)], x)
}

// areaPositionUnder is the position under a point on the screen, from where the area was in the
// last frame.
func (ui *Context) areaPositionUnder(id ID, memory *fieldMemory, x, y float32) int {
	ui.rewrapArea(memory)
	area, shown := ui.previous[id]
	if !shown {
		return memory.editor.caret
	}
	return ui.areaPositionAt(memory, x-(area.X+fieldPadding), y-(area.Y+fieldPadding)+memory.lineScroll)
}

// moveByLines moves the caret with the keys that go by lines in a text area: Up and Down to the
// line above or below, Home and End to the ends of the line, and with the shortcut key (Cmd's
// arrows on a Mac) to the ends of the text. It tells whether one of them was typed.
func (ui *Context) moveByLines(memory *fieldMemory, shift, shortcut, lineKey bool) bool {
	state, editor := ui.state, &memory.editor
	ui.rewrapArea(memory)
	index := memory.lineOf(editor.caret)
	switch {
	case shortcut && (state.Typed(input.KeyHome) || state.Typed(input.KeyUp) && lineKey):
		editor.moveTo(0, shift)
	case shortcut && (state.Typed(input.KeyEnd) || state.Typed(input.KeyDown) && lineKey):
		editor.moveTo(len(editor.text), shift)
	case state.Typed(input.KeyHome) || state.Typed(input.KeyLeft) && lineKey:
		editor.moveTo(memory.lines[index].start, shift)
	case state.Typed(input.KeyEnd) || state.Typed(input.KeyRight) && lineKey:
		editor.moveTo(memory.caretEnd(index), shift)
	case state.Typed(input.KeyUp):
		editor.moveTo(ui.positionOnNextLine(memory, -1), shift)
	case state.Typed(input.KeyDown):
		editor.moveTo(ui.positionOnNextLine(memory, 1), shift)
	default:
		return false
	}
	return true
}

// positionOnNextLine is where the caret lands a line up (direction -1) or down (+1): under the
// caret, or at the start or the end of the text when there is no line there.
func (ui *Context) positionOnNextLine(memory *fieldMemory, direction int) int {
	index := memory.lineOf(memory.editor.caret) + direction
	switch {
	case index < 0:
		return 0
	case index >= len(memory.lines):
		return len(memory.editor.text)
	}
	x, _ := ui.areaPoint(memory, memory.editor.caret)
	return ui.areaPositionAt(memory, x, float32(index)*ui.areaLineHeight(memory))
}

// followAreaScrolling scrolls a text area's lines with the wheel over it, or a finger dragged on
// it; inner areas take the wheel first, as scroll areas do.
func (ui *Context) followAreaScrolling(memory *fieldMemory, result interaction) {
	state := ui.state
	if result.held && ui.press.finger && ui.press.moved {
		memory.lineScroll -= float32(state.CursorDeltaY)
	}
	if delta := float32(state.ScrollDelta); delta != 0 && result.hovered && !ui.scrollUsed {
		memory.lineScroll -= delta * ui.Theme.ScrollStep
		ui.scrollUsed = true
	}
}

// followAreaFinger places the caret where a finger tapped a text area. A finger that moves scrolls
// the lines instead (followAreaScrolling), and one held still selects a word (followTouches).
func (ui *Context) followAreaFinger(id ID, memory *fieldMemory, result interaction) {
	memory.dragging = false
	if result.clicked && !ui.press.moved {
		memory.editor.moveTo(ui.areaPositionUnder(id, memory, ui.press.startX, ui.press.startY), false)
		memory.blinkStart = ui.now
	}
}

// measureArea works out what painting a text area needs: its lines at the area's width, the caret
// and the selection, and how far to scroll its lines to keep the caret in view.
func (ui *Context) measureArea(id ID, memory *fieldMemory, text string, focused bool) {
	paint := &memory.paint
	area, shown := ui.previous[id]
	width := float32(0)
	if shown {
		width = max(area.Width-2*fieldPadding, 1)
	}
	ui.wrapArea(memory, text, width)
	lineHeight := ui.areaLineHeight(memory)
	paint.multiLine, paint.lines, paint.lineHeight = true, memory.lineTexts, lineHeight
	room := area.Height - 2*fieldPadding
	content := float32(len(memory.lines)) * lineHeight
	if focused {
		ui.measureAreaSelection(memory)
		caret := memory.editor.caret
		paint.caretX, paint.caretY = ui.areaPoint(memory, caret)
		paint.caretShown = ui.blinkCaret(memory)
		if caret != memory.followedCaret || len(memory.editor.text) != memory.followedLength || room != memory.followedRoom {
			memory.lineScroll = keepLineInView(memory.lineScroll, paint.caretY, lineHeight, room)
			memory.followedCaret, memory.followedLength, memory.followedRoom = caret, len(memory.editor.text), room
		}
	}
	if !focused {
		memory.followedRoom = room // a field taking the focus follows its caret in the room it has
	}
	memory.lineScroll = clamp(memory.lineScroll, 0, max(content-room, 0))
	paint.lineScroll = memory.lineScroll
}

// keepLineInView scrolls as little as keeps a line, from top, inside a room's height.
func keepLineInView(scroll, top, lineHeight, room float32) float32 {
	switch {
	case top < scroll:
		return top
	case top+lineHeight > scroll+room:
		return top + lineHeight - room
	}
	return scroll
}

// measureAreaSelection works out the band of each line the selection covers, and the line under
// the word being composed, from the top left of the text.
func (ui *Context) measureAreaSelection(memory *fieldMemory) {
	paint, editor := &memory.paint, &memory.editor
	key := bandsKey{caret: editor.caret, anchor: editor.anchor, composeStart: editor.composeStart,
		composeEnd: editor.composeEnd, wrapped: memory.wrapped}
	if !memory.bandsValid || memory.bandsFor != key {
		ui.measureBands(memory)
		memory.bandsFor, memory.bandsValid = key, true
	}
	paint.selections, paint.composing = memory.selectionBands, memory.composeBands
}

// bandsKey is what a text area's bands depend on: they are worked out again only when it
// changes, not every frame the area has the focus.
type bandsKey struct {
	caret, anchor            int
	composeStart, composeEnd int
	wrapped                  wrapKey
}

// measureBands works out the selection's bands and the composing word's.
func (ui *Context) measureBands(memory *fieldMemory) {
	editor := &memory.editor
	memory.selectionBands = memory.selectionBands[:0]
	start, end := editor.selection()
	memory.selectionBands = ui.bandsBetween(memory, memory.selectionBands, start, end)
	memory.composeBands = memory.composeBands[:0]
	if editor.composeEnd > editor.composeStart && editor.composeEnd <= len(editor.text) {
		memory.composeBands = ui.bandsBetween(memory, memory.composeBands, editor.composeStart, editor.composeEnd)
	}
}

// bandsBetween adds a band for each line the text from start to end covers.
func (ui *Context) bandsBetween(memory *fieldMemory, bands []draw2d.Rectangle, start, end int) []draw2d.Rectangle {
	if end <= start {
		return bands
	}
	lineHeight, first := ui.areaLineHeight(memory), memory.lineOf(start)
	for index := first; index < len(memory.lines); index++ {
		line := memory.lines[index]
		if line.start > end || line.start == end && index > first {
			break // a selection that ends where a line starts covers nothing of it
		}
		from, to := max(start, line.start), min(end, line.end)
		left := ui.textWidth(string(memory.editor.text[line.start:from]), memory.weight, memory.size)
		right := ui.textWidth(string(memory.editor.text[line.start:to]), memory.weight, memory.size)
		if end > line.end {
			right += ui.textWidth(" ", memory.weight, memory.size) // the line's end is chosen too
		}
		if right <= left {
			continue
		}
		bands = append(bands, draw2d.Rectangle{X: left, Y: float32(index) * lineHeight, Width: right - left, Height: lineHeight})
	}
	return bands
}

// paintArea paints a text area's lines, selection and caret inside its box: only the lines in
// view, so a long text costs what a screenful does.
func (ui *Context) paintArea(element *element) {
	theme, paint, rect := ui.Theme, element.field, element.rect
	clip := draw2d.Rectangle{X: rect.X + fieldPadding - 1, Y: rect.Y + 1, Width: rect.Width - 2*fieldPadding + 2, Height: rect.Height - 2}
	ui.drawing.PushClip(clip)
	defer ui.drawing.PopClip()
	size := ui.scaledText(ui.fieldTextSize(paint.size))
	metrics := ui.measure("Hg", TextStyle{Size: size})
	left, top := rect.X+fieldPadding, rect.Y+fieldPadding-paint.lineScroll
	baseline := (paint.lineHeight-metrics.Height())/2 + metrics.Ascent
	for _, band := range paint.selections {
		band.X, band.Y = band.X+left, band.Y+top
		ui.drawing.Rectangle(band, 2, theme.Selection, draw2d.Transparent, 0)
	}
	if len(paint.lines) == 1 && paint.lines[0] == "" && paint.placeholder != "" {
		ui.drawing.Text(paint.placeholder, left, top+baseline, size, draw2d.Regular, theme.placeholder())
	}
	first := clampInt(int(paint.lineScroll/paint.lineHeight), 0, len(paint.lines))
	for index := first; index < len(paint.lines); index++ {
		lineTop := top + float32(index)*paint.lineHeight
		if lineTop > rect.Bottom() {
			break
		}
		ui.drawing.Text(paint.lines[index], left, lineTop+baseline, size, paint.weight, theme.Ink)
	}
	for _, band := range paint.composing {
		line := draw2d.Rectangle{X: left + band.X, Y: top + band.Y + baseline + 3, Width: band.Width, Height: composingLine}
		ui.drawing.Rectangle(line, 0, theme.Ink, draw2d.Transparent, 0)
	}
	if paint.focused && paint.caretShown {
		caret := draw2d.Rectangle{X: left + paint.caretX - 0.75, Y: top + paint.caretY + baseline - metrics.Ascent - 1, Width: 1.5, Height: metrics.Height() + 2}
		ui.drawing.Rectangle(caret, 0, theme.Ink, draw2d.Transparent, 0)
	}
}
