# go-ui

```go
import "github.com/akrck02/go-ui"
```

Package ui builds interfaces (panels, buttons, labels…) over a go-renderer view,
in immediate mode: every frame the program describes the interface again from
its own variables, and each control answers what happened to it (a button
returns true on the frame it is clicked). Nothing has to be kept in step with
the program's data, because the interface holds none of it.

A frame goes:

    ui.Begin(state)              // the frame's mouse and keyboard
    ui.Panel("Light", func() {   // controls, nested in boxes
    	if ui.Button("Switch on") { light.On = true }
    })
    ui.End()                     // lay out, paint, claim the mouse if it is over the interface

and world.Overlay.Drawing = ui.Drawing() makes go-renderer paint it over the
scene.

Layout is deferred: controls are boxes in a tree that is sized and placed in
End, with rules like a flexible box (fit the content, grow into free space,
fixed or a percentage, padding, gaps, alignment). Controls react to the mouse
using where they were drawn in the previous frame, a frame earlier than they
are drawn now: invisible at sixty frames a second, and it lets a box be sized by
things declared after it.

### Constants and variables

```go
var EnglishEditingWords = EditingWords{Cut: "Cut", Copy: "Copy", Paste: "Paste", SelectAll: "Select all"}
```

EnglishEditingWords are the bar's labels unless a program sets others (SetEditingWords).

### AccessibleNode

AccessibleNode is one thing a screen reader can read and act on.

```go
type AccessibleNode struct {
	ID     ID          // the same from one frame to the next while the thing stays
	Parent ID          // the scroll area it is in, or 0 for the window
	Area   draw2d.Rect // where it shows, in layout pixels: the part a scroll area does not hide
	Description
	Pressable     bool // a double tap presses it (ReaderPress)
	LongPressable bool // the program reads LongPressed after it (ReaderLongPress)
	Scrollable    bool // ReaderForward and ReaderBackward scroll it, or move a slider
	Typing        bool // the keyboard types in it now
}
```

### Align

Align places children inside the free space of their box, along one axis.

```go
type Align int
```

```go
const (
	Start  Align = iota // left or top
	Center              // in the middle
	End                 // right or bottom
)
```

### ChipLook

ChipLook is a chip's own colours: its fill and the ink of its text and icon. The zero value uses the theme's quiet colours (Sunken and InkSoft).

```go
type ChipLook struct {
	Fill, Ink draw2d.Color
}
```

### Column

Column is one column of a table: its title, its width and where its cells sit.

```go
type Column struct {
	Title string
	// Width is the column's width. Use Fixed for narrow columns and Share for the rest: shares
	// start from nothing, so the header and every row come out the same.
	Width Sizing
	Align Align // Start for words, End for numbers
	// Weight is the face of its cells: draw2d.Mono for a column of codes.
	Weight draw2d.Weight
}
```

### Context

Context holds an interface between frames: what was where, what the mouse is on, which panel is in front, and the small state controls keep (a panel's position). Use one per window, from the goroutine that runs the update steps.

```go
type Context struct {
	Theme Theme

	// KeyboardNavigation says when the interface reads Tab, Enter and Space (focus.go).
	KeyboardNavigation KeyboardNavigation

	// LessMotion keeps controls still when the person asks the system for less motion: a held
	// control does not shrink. The program sets it from the platform's setting.
	LessMotion bool

	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func New() (*Context, error)` | New returns an interface that measures text with go-renderer's fonts, in the default theme. |
| `func NewWithMeasurer(measurer draw2d.Measurer) *Context` | NewWithMeasurer returns an interface that measures text with a measurer of your choice: tests use a fake one, so they need no fonts. |
| `func (ui *Context) Animate(name string, target float32, duration time.Duration) float32` | Animate eases a value towards a target over a duration, starting again from where it is when the target changes: a switch's handle going across, a tab's underline moving to the tab chosen. |
| `func (ui *Context) Appear(name string, duration time.Duration) float32` | Appear goes from 0 to 1 over a duration from the frame something was first shown after it was not: a dialog fading and rising in, a note sliding up. |
| `func (ui *Context) Area(path ...string) (draw2d.Rect, bool)` | Area tells where a control was drawn in the last frame, found by its label and the scopes around it, outermost first: Area("Properties", "Apply") is the Apply button of the Properties panel (panels open a scope with their title). |
| `func (ui *Context) Back() bool` | Back answers the system's back (a phone's back gesture or button) in this frame, and tells whether there was one: the part of the interface the person is in goes back (a screen returns to the one before) and returns true. |
| `func (ui *Context) Begin(state *input.State)` | Begin starts describing the interface for a frame, with the frame's input. |
| `func (ui *Context) Box(layout Layout, content func())` | Box adds a box with a layout, and what content adds goes inside it. |
| `func (ui *Context) Button(label string) bool` | Button adds a button and returns true on the frame it is clicked (pressed and let go over it). |
| `func (ui *Context) Caption(text string)` | Caption adds a line of small, quiet text: a hint or a unit. |
| `func (ui *Context) Checkbox(label string, value *bool) bool` | Checkbox adds a box to tick with a label after it; clicking either flips value. |
| `func (ui *Context) Chip(icon, label string, look ChipLook) bool` | Chip adds a small rounded label with an optional icon (a Lucide name), in its own colours: a tag, a state, a filter. |
| `func (ui *Context) Clickable(name string, layout Layout, looks Looks, content func()) bool` | Clickable adds a box that reacts to the mouse, with whatever content, and returns true on the step it is clicked: the building block for controls of your own (a table row, a card). |
| `func (ui *Context) ClipboardText() string` | ClipboardText is the text on the system's clipboard, "" when there is none. |
| `func (ui *Context) Column(content func())` | Column stacks what content adds from top to bottom, across the whole width. |
| `func (ui *Context) Copy(text string)` | Copy puts a text on the system's clipboard (a link, a code, what a row holds), to paste in another application. |
| `func (ui *Context) Describe(description Description)` | Describe tells a screen reader more about the box added just before it (a control, a text): what it is, its name, its value and its state. |
| `func (ui *Context) DragSource(name string, payload any, layout Layout, content func()) bool` | DragSource adds a box whose content can be dragged away, carrying payload: press it and move the pointer a few pixels. |
| `func (ui *Context) Dragging() (Drag, bool)` | Dragging tells whether a payload is being dragged, and where the pointer is: a program shows where it would land on its scene while Drag.OverScreen is true. |
| `func (ui *Context) Drawing() *draw2d.List` | Drawing is the draw list the interface paints into; give it to the scene's overlay once (world.Overlay.Drawing = ui.Drawing()) and go-renderer draws it after every frame. |
| `func (ui *Context) DropTarget(name string, layout Layout, accepts func(payload any) bool, content func()) (payload any, dropped bool)` | DropTarget adds a box that takes dragged payloads. |
| `func (ui *Context) DroppedOnScreen() (Drag, bool)` | DroppedOnScreen tells, on the step the button is let go, whether a payload was dropped on the open screen, with the pointer's place: the program drops it on its scene there. |
| `func (ui *Context) EditTypedField(edit KeyboardText)` | EditTypedField puts what an on-screen keyboard made of the field's text in it, in the next frame: the field takes the whole text as typing (one line, cut at its MaxLength), the selection and the word being composed. |
| `func (ui *Context) EditingWords() EditingWords` | EditingWords are the labels the bar over a selection shows: the ones set, or English. |
| `func (ui *Context) End()` | End lays the interface out, paints it into the draw list and tells the input state whether the interface has the mouse or the keyboard. |
| `func (ui *Context) FieldFocused(name string) bool` | FieldFocused tells whether the text field of a name, in the same scope, has the keyboard: a program that draws the field's box itself (a Material field, with its text before and after) colours it by that. |
| `func (ui *Context) FlatButton(label string) bool` | FlatButton adds a button with no fill or border until hovered: menu bars and toolbars. |
| `func (ui *Context) FlatIconButton(icon, label string) bool` | FlatIconButton is FlatButton with an icon; with an empty label it is a square icon button. |
| `func (ui *Context) Focus(path ...string)` | Focus gives the keyboard to a text field or a number field, found by its scopes and label as Area finds it, with its whole text selected: to start typing in a search box on a shortcut, or in the first field of a form that just opened. |
| `func (ui *Context) FocusField(name string)` | FocusField gives the keyboard to the text field of a name, in the same scope, from the next time it is built: a search that opens ready to type in. |
| `func (ui *Context) Fold(name string, open bool, duration time.Duration, layout Layout, content func())` | Fold lays content out in a box that opens to the content's height and closes to nothing with motion, clipping what does not show yet: the answer under a question, more under a list row. |
| `func (ui *Context) Frame(layout Layout, look Look, content func())` | Frame adds a box with a layout and a look (fill, border, corners, shadow). |
| `func (ui *Context) Grid(columns int, content func())` | Grid lays out what content adds in rows of a number of equal columns, across the whole width. |
| `func (ui *Context) HasKeyboard() bool` | HasKeyboard tells whether the interface reads Tab, Enter and Space in this frame. |
| `func (ui *Context) Heading(text string)` | Heading adds a line of bold, larger text: the title of a group. |
| `func (ui *Context) Icon(name string)` | Icon adds an icon (a Lucide name, https://lucide.dev) at the theme's icon size. |
| `func (ui *Context) IconButton(icon, label string) bool` | IconButton adds a button with an icon (a Lucide name) before its label; with an empty label it is a square button with the icon alone, named by the icon. |
| `func (ui *Context) IconSized(name string, size float32, color draw2d.Color)` | IconSized adds an icon of a size and colour. |
| `func (ui *Context) Image(picture *draw2d.Image, layout Layout, radius float32)` | Image adds a picture filling a box laid out as any other (a photo in a card, a person's avatar): scaled to cover it and cropped evenly where their shapes differ, its corners rounded by radius. |
| `func (ui *Context) Insets() Insets` | Insets returns what the system draws over the window's edges. |
| `func (ui *Context) Label(text string)` | Label adds a line of body text. |
| `func (ui *Context) List(name string, layout Layout, count int, rowHeight float32, row func(index int))` | List shows count rows of the same height in a scroll area and builds only the rows that can be seen, so a list of ten thousand rows costs what a list of twenty does. |
| `func (ui *Context) ListBox(name string, layout Layout, items []string, selected *int) bool` | ListBox shows items as a scrolling list of selectable rows and returns true when the selection changes. |
| `func (ui *Context) LongPressed() bool` | LongPressed tells whether the control added just before it was long-pressed in this frame: a finger held still on it for half a second, or the right mouse button pressed on it. |
| `func (ui *Context) MenuButton(icon, label string, content func())` | MenuButton adds a flat button with an optional icon (a Lucide name); clicking it opens a menu under it, over everything, whose items content adds with MenuItem (and Separator between groups). |
| `func (ui *Context) MenuItem(icon, label, hint string) bool` | MenuItem adds an item to the menu being built: an optional icon, a label and an optional hint on the right (a keyboard shortcut, say). |
| `func (ui *Context) Modal(name string, options LayerOptions, content func()) (dismissed bool)` | Modal shows content over everything with what is behind it dimmed, as Sheet does, but drawn entirely by the program: a dialog, a sheet from the bottom of a phone. |
| `func (ui *Context) NumberField(name string, value *float32, minimum, maximum, step float32, format string) bool` | NumberField adds a number between minimum and maximum, shown with format (for example "%.0f nm"), with buttons to take a step off or add one. |
| `func (ui *Context) Overlay(name string, options LayerOptions, content func())` | Overlay shows content over the screen and its panels without dimming or blocking what is behind it: a note at the bottom with an action in it, a floating toolbar. |
| `func (ui *Context) Paint(layout Layout, paint func(area draw2d.Rect, drawing *draw2d.List))` | Paint adds a box laid out as any other, which the program paints itself once the layout knows where it is: a shape go-ui has no control for (a loading indicator, a wave, a chart). |
| `func (ui *Context) Panel(title string, content func())` | Panel adds a floating panel with a title bar; drag the bar to move it, click the chevron to fold it. |
| `func (ui *Context) PanelWith(title string, options PanelOptions, content func())` | PanelWith adds a panel placed and sized by options. |
| `func (ui *Context) Pointer() (x, y float32)` | Pointer is where the pointer is, in the window's layout pixels: a finger on a touch screen. |
| `func (ui *Context) Pop()` | Pop closes the scope opened last. |
| `func (ui *Context) Popover(name string, open *bool, look Look, content func())` | Popover floats a box of the program's own under the control added just before it, over everything, at least as wide as the control: a dropdown's choices, a small form about a row. |
| `func (ui *Context) Pressing() bool` | Pressing tells whether the pointer is pressed now: a finger is on a touch screen, the mouse's left button is down. |
| `func (ui *Context) PrimaryButton(label string) bool` | PrimaryButton adds a button for the main action of a panel. |
| `func (ui *Context) PrimaryIconButton(icon, label string) bool` | PrimaryIconButton is PrimaryButton with an icon (a Lucide name) before its label: the main action of a panel, such as adding a new thing. |
| `func (ui *Context) Push(name string)` | Push opens a scope: ids of the controls added until Pop are made from this name too. |
| `func (ui *Context) ReleaseKeyboard()` | ReleaseKeyboard gives the keyboard back to the program and takes the focus off the controls. |
| `func (ui *Context) RemovableChip(icon, label string, look ChipLook) (clicked, removed bool)` | RemovableChip is a Chip with a small × after its label. |
| `func (ui *Context) RequestFrame(delay time.Duration)` | RequestFrame asks the window for another frame within a delay, even if no input arrives: for a program that animates something of its own (a clock, a replay). |
| `func (ui *Context) Resizable(name string, side Side, size *float32, minimum, maximum float32, layout Layout, content func()) bool` | Resizable adds a box whose width (with its handle on the right or left side) or height (bottom or top) the person changes by dragging its edge, between minimum and maximum pixels (0 = no maximum). |
| `func (ui *Context) Row(content func())` | Row lines up what content adds from left to right, across the whole width, centred vertically. |
| `func (ui *Context) ScreenReaderAct(id ID, action ReaderAction)` | ScreenReaderAct does what a screen reader asks of a node, in the next frame: a press clicks the control as a tap would, so whatever the program does on a click happens. |
| `func (ui *Context) ScreenReaderNodes() []AccessibleNode` | ScreenReaderNodes is what the last frame showed, for a screen reader, in reading order: the window's own first, then what panels, sheets and popups hold, front-most last. |
| `func (ui *Context) ScreenReaderOn() bool` | ScreenReaderOn tells whether a screen reader is on (SetScreenReader). |
| `func (ui *Context) ScreenReaderSetText(id ID, text string)` | ScreenReaderSetText replaces a field's whole text with what a screen reader typed, pasted or dictated into it, in the next frame, without bringing the keyboard up. |
| `func (ui *Context) ScrollArea(name string, layout Layout, content func())` | ScrollArea adds a box whose content scrolls up and down when it is taller than the box: with the wheel or the trackpad over it, by dragging the thin bar on its right, or, on a touch screen, by dragging the content with a finger, which glides on when let go. |
| `func (ui *Context) ScrollOffset(name string) float32` | ScrollOffset tells how far a scroll area's content is scrolled from its top, as of the last frame: a bar over a list shrinks or takes a colour as the list goes up under it. |
| `func (ui *Context) ScrollPull(name string) (pulling, releasedAt float32)` | ScrollPull tells how far a finger is pulling a scroll area's content down past its top, and, in the frame the finger lets go, how far it was pulled then (zero otherwise): a list refreshes when pulled far enough and let go. |
| `func (ui *Context) Select(name string, selected *int, options []string) bool` | Select adds a field showing the selected option; clicking it opens the list of options over everything. |
| `func (ui *Context) Selectable(icon, label string, selected bool) bool` | Selectable adds an item with an optional icon that shows whether it is selected, and returns true on the step it is clicked. |
| `func (ui *Context) Separator()` | Separator draws a thin line across the row or column it is in. |
| `func (ui *Context) SetEditingWords(words EditingWords)` | SetEditingWords sets the labels of the bar over a selection, in the person's language. |
| `func (ui *Context) SetInsets(insets Insets)` | SetInsets tells the interface what the system draws over the window; the window calls it. |
| `func (ui *Context) SetScreenReader(on bool)` | SetScreenReader tells the interface whether a screen reader is on; the window does, on a phone. |
| `func (ui *Context) SetScreenReaderFocus(id ID)` | SetScreenReaderFocus says which node the screen reader is on (0 for none), so the interface rings it, as a phone rings what TalkBack reads. |
| `func (ui *Context) SetTextScale(scale float32)` | SetTextScale sets how much larger than its size each text is drawn: the person's text size, as a phone's settings give it (1.3 for a larger text). |
| `func (ui *Context) Sheet(title string, options SheetOptions, content func()) (dismissed bool)` | Sheet shows a card in the middle of the window, over everything, with what is behind it dimmed: a question to confirm, a short form. |
| `func (ui *Context) Size() (width, height float32)` | Size returns the window's size in this frame, in layout pixels: a program chooses a layout by it (a list alone on a phone, a list and its details side by side on a tablet). |
| `func (ui *Context) Slider(name string, value *float32, minimum, maximum float32) bool` | Slider adds a track with a knob to drag between minimum and maximum, across the width it is given. |
| `func (ui *Context) SliderStep(name string, value *float32, minimum, maximum, step float32) bool` | SliderStep is Slider with the value kept to whole steps (0 = any value). |
| `func (ui *Context) Spacer()` | Spacer takes the free space of the row or column it is in, pushing what follows to the end. |
| `func (ui *Context) SwipeBox(name string, layout Layout, axis Direction, content func(swipe Swipe))` | SwipeBox lays content out in a box a finger drags along an axis (sideways, for pages, cards in a row, a row swiped away): content is given what the finger does in this frame, and places itself by it. |
| `func (ui *Context) Table(name string, layout Layout, columns []Column, rows func(table *Table))` | Table adds a table with a header and rows that scroll under it: rows adds the rows, with table.Row, and table.Group for a caption between groups of rows. |
| `func (ui *Context) Tabs(name string, selected *int, labels ...string) bool` | Tabs adds a row of segments, one per label, with the selected one raised; clicking a segment selects it. |
| `func (ui *Context) TakeKeyboard()` | TakeKeyboard gives the interface the keyboard, as a game does when it opens its own menu: Tab then moves between the controls until Escape or ReleaseKeyboard gives the keys back. |
| `func (ui *Context) Text(text string, style TextStyle)` | Text adds a line of text in a style. |
| `func (ui *Context) TextField(name string, text *string) bool` | TextField adds a one-line field to type text in. |
| `func (ui *Context) TextFieldWith(name string, text *string, options TextFieldOptions) bool` | TextFieldWith is TextField with a placeholder, a length limit or a width. |
| `func (ui *Context) TextScale() float32` | TextScale is the text size SetTextScale gave, 1 by default: a control of a program's own whose height holds text grows by it. |
| `func (ui *Context) Toast(text string, warning bool)` | Toast shows a short note centred near the bottom of the window, over everything, without taking the mouse: what an action did, or (warning) why it could not be done. |
| `func (ui *Context) Toggle(label string, value *bool) bool` | Toggle adds a row with a label on the left and a switch on the right; clicking the row flips value. |
| `func (ui *Context) Tooltip(text string)` | Tooltip explains the control added just before it: when the pointer rests on that control for a moment (Theme.TooltipDelay), a small dark note with the text shows under the pointer. |
| `func (ui *Context) Touches() []input.Touch` | Touches returns the fingers on a touch screen in this frame, in layout pixels, all of them: those on controls too. |
| `func (ui *Context) TypedField() (KeyboardText, bool)` | TypedField is the field typed in during the frame, for an on-screen keyboard (the window hands it on): its text, its selection and the word being composed. |
| `func (ui *Context) Typing() bool` | Typing tells whether a text field had the focus in the last frame: a phone shows its on-screen keyboard while it does. |
| `func (ui *Context) WantsKeyboard() bool` | WantsKeyboard tells whether a control had the keyboard in the last frame (a text field being typed in). |
| `func (ui *Context) WantsMouse() bool` | WantsMouse tells whether the interface used the mouse in the last frame: the pointer was over a panel or a control, or dragging one. |

### Description

Description is what a screen reader says about a thing, beyond what its boxes tell.

```go
type Description struct {
	Role  Role
	Label string // what it is called; empty = the texts inside it
	Value string // what it holds, read after the label: a field's text
	// Checked is a checkbox ticked, a switch on or a radio button chosen.
	Checked  bool
	Selected bool // the tab, the destination or the choice shown
	Disabled bool // shown, but does nothing now
	Secret   bool // a field whose text is not read aloud (a password)
	// Hidden leaves it out, with all it holds: a decoration, or something repeated just beside it.
	Hidden bool
	Range  Range // a slider's; Maximum above Minimum when it has one
}
```

### Direction

Direction is the way a box lines up its children.

```go
type Direction int
```

```go
const (
	Horizontal Direction = iota // left to right
	Vertical                    // top to bottom
)
```

### Drag

Drag is a payload being dragged, or dropped, and where the pointer is, in layout pixels.

```go
type Drag struct {
	Payload any
	X, Y    float32
	// OverScreen tells whether the pointer is over the open screen, where no panel, painted box
	// or drop target is: where letting go drops the payload on the program's scene.
	OverScreen bool
}
```

### Edge

Edge is a side of the window a panel can be docked to.

```go
type Edge int
```

```go
const (
	Floating Edge = iota // not docked: the panel floats where it is put, and can be dragged
	DockLeft
	DockRight
	DockTop
	DockBottom
)
```

### EditingWords

EditingWords are the labels of the bar over a selection on a touch screen.

```go
type EditingWords struct {
	Cut, Copy, Paste, SelectAll string
}
```

### ID

ID names a control from one frame to the next, so the interface can remember that the mouse is pressing it or where a panel was left. It is a hash of the control's label and of the scopes it is in (see Push), so two "Apply" buttons in different panels do not mix.

```go
type ID uint64
```

### Insets

Insets are how far in from each edge of the window the system draws over it, in layout pixels: a phone's status bar at the top and its gesture bar at the bottom. A window that covers the whole screen paints its own bars' backgrounds under them, and keeps its controls clear of them. On a desktop and in a browser they are zero.

```go
type Insets struct {
	Top, Bottom, Left, Right float32
}
```

### KeyboardNavigation

KeyboardNavigation says when the interface reads Tab, Enter and Space to move between controls and press them, and when those keys belong to the program (a game, a 3D view).

```go
type KeyboardNavigation int
```

```go
const (
	// WhenGiven, the default: the interface has the keyboard only while a field is typed in, a
	// dialog, sheet or menu is open, or the program gave it with TakeKeyboard. Escape gives it
	// back. A click on a button does not take it, so Space in a game never presses the last button
	// clicked.
	WhenGiven KeyboardNavigation = iota
	// Always: the program is only an interface (a register, a settings app), so Tab always moves
	// between its controls.
	Always
)
```

### KeyboardText

KeyboardText is the field being typed in, as an on-screen keyboard sees and edits it: a phone's keyboard suggests words from the text around the caret, keeps the word it is still making (composing: a prediction, a Chinese or Japanese reading) until it commits it, and deletes around the caret. Positions count characters (runes); the selection runs from SelectionStart, where it began, to SelectionEnd, the caret.

```go
type KeyboardText struct {
	Field                        ID // which field: an edit of another field is dropped
	Text                         string
	SelectionStart, SelectionEnd int
	ComposingStart, ComposingEnd int  // the word being composed, underlined; equal when none
	Secret                       bool // a password: the keyboard suggests nothing
}
```

### LayerOptions

LayerOptions place what a layer holds in the window.

```go
type LayerOptions struct {
	AlignX, AlignY Align   // Center, Center for a dialog; Center, End for a sheet from the bottom
	Padding        Padding // how far it keeps from the window's edges
	Width          Sizing  // its width; the zero value is as wide as its content
	Height         Sizing  // its height; the zero value is as tall as its content, Grow() all of it
	// Hidden is how much of a modal's dimming is not shown yet, from 0 (all of it, the zero value)
	// to 1 (none): a modal fading in as it opens (see Appear).
	Hidden float32
	// Leave is how it goes when the program stops showing it (leave.go); at once by default.
	Leave Leave
}
```

### Layout

Layout says how a box sizes itself and places its children. The zero value is a box that lines its children up left to right, fits them, and has no padding or gap.

```go
type Layout struct {
	Direction     Direction
	Width, Height Sizing
	Padding       Padding
	Gap           float32 // space between two children, in pixels
	// Wrap, on a horizontal box, starts a new line when the next child would pass the box's
	// right edge, like words in a paragraph (a row of tags). Gap is kept between children and
	// between lines. The box needs a width that does not follow its content (Grow in a column,
	// Fixed, Percent), or it is one line as wide as all its children.
	// (Next to Gap, it takes the padding the compiler leaves there: boxes stay as small.)
	Wrap   bool
	AlignX Align // where the children sit horizontally when there is room to spare
	AlignY Align // where the children sit vertically when there is room to spare
	// Columns, when above 0, makes a grid: children fill rows of this many equal cells, left to
	// right and then down, with Gap between cells both ways. Direction is not used then.
	Columns int
}
```

### Leave

Leave is how a layer goes when the program stops showing it: over Duration it fades out, dimming and all, as it moves by X and Y (a sheet sliding back down, a drawer back off the side). Its zero value goes at once.

```go
type Leave struct {
	Duration time.Duration
	X, Y     float32
}
```

### Look

Look is how a box is painted: a fill, a border, rounded corners and a shadow. The zero value paints nothing, so a box is only a way to arrange its children.

```go
type Look struct {
	Fill        draw2d.Color
	Border      draw2d.Color
	BorderWidth float32
	Radius      float32
	Shadow      bool // a soft shadow under it, for things that float over others
	Clip        bool // children are cut to the box's area
	// Edge is a solid band of this colour, EdgeWidth tall, just under the box, as if it stood on a
	// block: a design system's main button. It takes no room in the layout, like a shadow.
	Edge      draw2d.Color
	EdgeWidth float32
}
```

### Looks

Looks are a box's looks while it is left alone, under the pointer and pressed.

```go
type Looks struct {
	Normal, Hovered, Pressed Look
}
```

### Padding

Padding is the empty space between a box's edges and its children, in pixels.

```go
type Padding struct {
	Left, Top, Right, Bottom float32
}
```

| Function or method | |
|---|---|
| `func Pad(all float32) Padding` | Pad is the same padding on every side. |
| `func PadXY(horizontal, vertical float32) Padding` | PadXY is one padding on the left and right and another on the top and bottom. |

### PanelOptions

PanelOptions place and size a panel.

```go
type PanelOptions struct {
	X, Y   float32 // where it first appears, in pixels from the window's top-left corner
	Width  Sizing  // zero = as wide as its content
	Height Sizing  // zero = as tall as its content
	Dock   Edge    // docked panels fill a side of the window and cannot be dragged
	Locked bool    // a floating panel that cannot be dragged
}
```

### Range

Range is a slider's span and where it is now.

```go
type Range struct {
	Minimum, Maximum, Current float32
}
```

### ReaderAction

ReaderAction is what a screen reader asks of a node.

```go
type ReaderAction int
```

```go
const (
	ReaderPress     ReaderAction = iota + 1 // a double tap: the control is clicked
	ReaderLongPress                         // a double tap held: the control is long-pressed
	ReaderForward                           // a scroll area shows what comes next; a slider goes up a step
	ReaderBackward                          // a scroll area shows what came before; a slider goes down a step
)
```

### Role

Role is what a thing on the screen is, which a screen reader says after its name: "Wi-Fi, switch, on".

```go
type Role int
```

```go
const (
	NoRole       Role = iota // worked out from the box: a control is a button, anything else text
	ButtonRole               // pressed to do something
	CheckboxRole             // ticked or not
	SwitchRole               // on or off
	RadioRole                // one of a few choices, of which one is chosen
	SliderRole               // a value between two ends (Range)
	FieldRole                // typed in
	TabRole                  // one of the pages over a screen, one of them selected
	HeadingRole              // the title of what follows
	ImageRole                // a picture, read by its label
	TextRole                 // something to read
	ScrollRole               // an area that scrolls (ScrollArea)
)
```

### RowLook

RowLook is how a row's first cell starts: an icon on a tinted tile, and an indent for a row that belongs to the one above it (an attribute under its type).

```go
type RowLook struct {
	Icon            string
	IconInk, Tile   draw2d.Color // the icon's colour and its tile's; zero = the theme's soft ink, no tile
	Indent          float32
	Secondary       string // a quieter second line under the first cell's text
	SecondaryWeight draw2d.Weight
}
```

### SheetOptions

SheetOptions size a sheet.

```go
type SheetOptions struct {
	Width Sizing // the card's width; zero = 420 pixels
}
```

### Side

Side is the side of a resizable box its handle is on.

```go
type Side int
```

```go
const (
	RightSide  Side = iota // the box grows to the right: a column on the left of a window
	LeftSide               // the box grows to the left: a column on the right
	BottomSide             // the box grows down: a strip at the top
	TopSide                // the box grows up: a strip at the bottom
)
```

### Sizing

Sizing says how big a box is along one axis. The zero value fits the content.

```go
type Sizing struct {

	// Min and Max bound the size, in pixels; 0 = no bound.
	Min, Max float32
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func Fit() Sizing` | Fit makes a box as big as its content (the default). |
| `func Fixed(pixels float32) Sizing` | Fixed makes a box exactly this many pixels. |
| `func Grow() Sizing` | Grow makes a box take the space its parent has left, shared equally with its growing siblings. |
| `func GrowBy(weight float32) Sizing` | GrowBy is Grow with a weight: a box with weight 2 takes twice the share of one with weight 1. |
| `func Percent(share float32) Sizing` | Percent makes a box a share (0 to 1) of its parent's inner size. |
| `func Share(weight float32) Sizing` | Share makes a box take weight shares of the free space, starting from nothing rather than from its content: boxes with equal shares come out equal whatever they hold, so table columns line up from row to row and cards in a row match. |
| `func (sizing Sizing) Between(minimum, maximum float32) Sizing` | Between returns the sizing bounded by a minimum and a maximum, in pixels (0 = no bound). |

### Swipe

Swipe is what a finger does to a swipe box in a frame.

```go
type Swipe struct {
	Offset   float32 // how far the finger has moved along the box's axis since it took the box
	Velocity float32 // pixels a second along the axis, as it last moved
	Held     bool    // a finger is dragging the box
	Released bool    // the finger let go in this frame; Offset and Velocity are its last
	// Area is where the box was laid out in the last frame (zero in its first): its width is a
	// page's, for content that moves a page at a time.
	Area draw2d.Rect
}
```

### Table

Table is a table being built, handed to the function that adds its rows. Rows and group captions out of sight are only counted, not built, so a table of a few thousand rows costs what the rows in sight do.

```go
type Table struct {
	// Has unexported fields.
}
```

| Function or method | |
|---|---|
| `func (table *Table) Group(caption string)` | Group adds a caption across all the columns, in small capitals, above the rows of a group. |
| `func (table *Table) Row(id string, selected bool, cells ...string) bool` | Row adds a row with a text per column, named by id (unique in the table: the item's key), and returns true on the step it is clicked. |
| `func (table *Table) RowWith(id string, selected bool, look RowLook, cells ...string) bool` | RowWith adds a row as Row does, its first cell starting as look says. |

### TextFieldOptions

TextFieldOptions shape a text field.

```go
type TextFieldOptions struct {
	Placeholder string // shown, quiet, while the field is empty
	MaxLength   int    // the most characters it takes; 0 = no limit
	Width       Sizing // the zero value grows across the row or column
	// Weight is the face the text is written in: draw2d.Mono for a code or a key.
	Weight draw2d.Weight
	// Plain draws no box of the field's own, neither fill nor border nor focus ring: for a field
	// inside a box of the program's (a search bar's pill), which shows where it is.
	Plain bool
	// Secret shows a dot for each character (a password), and lets nothing be copied or cut.
	Secret bool
}
```

### TextStyle

TextStyle is how a text is written.

```go
type TextStyle struct {
	Size   float32 // in pixels; 0 = the theme's text size
	Weight draw2d.Weight
	Color  draw2d.Color // the zero colour = the theme's ink
}
```

### Theme

Theme holds the tokens every control reads its colours, sizes and corners from, so an interface changes look in one place. Sizes are layout pixels.

```go
type Theme struct {
	Background  draw2d.Color // behind everything, in a window with no scene (see package window)
	Surface     draw2d.Color // panels and secondary buttons
	Sunken      draw2d.Color // hovered secondary controls, tracks, wells
	Ink         draw2d.Color // body text
	InkSoft     draw2d.Color // secondary text and icons
	InkQuiet    draw2d.Color // hints and captions
	Line        draw2d.Color // borders and separators
	ControlLine draw2d.Color // borders of fields, selects and secondary buttons: stronger than Line, so a control reads as one
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
	// Backdrop blurs what is behind every translucent panel by this many layout pixels (a standard
	// deviation): frosted glass. Panels inside a translucent one are not blurred again; 0 blurs
	// nothing, at no cost.
	Backdrop float32
	// FocusWidth is the width of the ring around the control Tab reached, lying on its edge with no
	// gap; 0 keeps go-ui's 3 pixels. On a button filled with Primary the ring is drawn inside it,
	// in OnPrimary, where FocusRing could not be told from the fill.
	FocusWidth float32
	// FieldFocusEdge lights a field's own edge in FocusRing while it is typed in, instead of a
	// ring around it.
	FieldFocusEdge bool
	// PressScale is how big a control is drawn while it is held, around its centre (0.97 gives a
	// little); its content shrinks with it and the layout does not move. 0 keeps it still.
	PressScale float32
}
```

| Function or method | |
|---|---|
| `func DarkTheme() Theme` | DarkTheme is the light theme's tokens for a dark look, with the same sizes. |
| `func LightTheme() Theme` | LightTheme is the default: a light, quiet look that reads well over any scene. |

---

[← Reference](0-reference.md) · generated from the doc comments by `generate.py`
