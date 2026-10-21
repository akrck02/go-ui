// Package ui builds interfaces (panels, buttons, labels…) over a go-renderer view, in immediate
// mode: every frame the program describes the interface again from its own variables, and each
// control answers what happened to it (a button returns true on the frame it is clicked). Nothing
// has to be kept in step with the program's data, because the interface holds none of it.
//
// A frame goes:
//
//	ui.Begin(state)              // the frame's mouse and keyboard
//	ui.Panel("Light", func() {   // controls, nested in boxes
//		if ui.Button("Switch on") { light.On = true }
//	})
//	ui.End()                     // lay out, paint, claim the mouse if it is over the interface
//
// and world.Overlay.Drawing = ui.Drawing() makes go-renderer paint it over the scene.
//
// Layout is deferred: controls are boxes in a tree that is sized and placed in End, with rules
// like a flexible box (fit the content, grow into free space, fixed or a percentage, padding,
// gaps, alignment). Controls react to the mouse using where they were drawn in the previous
// frame, a frame earlier than they are drawn now: invisible at sixty frames a second, and it lets
// a box be sized by things declared after it.
package ui

import (
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/opengl/overlay"
)

// Context holds an interface between frames: what was where, what the mouse is on, which panel is
// in front, and the small state controls keep (a panel's position). Use one per window, from the
// goroutine that runs the update steps.
type Context struct {
	// within is the area overlays and modals are laid out in while isWithin (Within)
	within   draw2d.Rect
	isWithin bool

	Theme Theme

	measurer draw2d.Measurer
	texts    textCache
	state    *input.State
	scale    float32 // framebuffer pixels per layout pixel
	width    float32 // window size, in layout pixels
	height   float32
	drawing  draw2d.List
	clip     draw2d.Rect // while painting: the clip of the boxes around, when clipped
	clipped  bool
	// insideFrost is true while the children of a frosted box are painted (see isFrosted).
	insideFrost bool
	frame       int
	now         time.Time // the time of this frame, from the input state's clock
	building    bool

	elements []element
	open     []int // the boxes open now, innermost last
	roots    []root
	scopes   []ID
	seen     map[ID]int // how many times an id was used this frame, to tell repeats apart

	previous map[ID]draw2d.Rect // where things were in the last frame
	current  map[ID]draw2d.Rect
	// what of a thing showed, where a box around it clipped it: what the pointer could reach in
	// the last frame (reachOf), for things clipped; the others are reached wholly
	previousReach, currentReach map[ID]draw2d.Rect
	order                       []ID // panels from back to front
	hoveredRoot                 ID
	hot                         ID   // the control under the mouse
	active                      ID   // the control being pressed or dragged
	focus                       ID   // the control the keyboard goes to
	focusSeen                   bool // the focused control was built this frame; if not, it went away
	// focusables are the stops of the Tab order, in the order they were built this frame; tabFrom
	// is the stop where Tab was pressed (tabBack with Shift; tabPressed when no stop had it), and
	// focusByKeyboard tells a field taking the focus that it came by Tab (keyboardFocus, below, is
	// the control, not a field, that Tab reached).
	focusables      []focusable
	tabFrom         ID
	tabBack         bool
	tabPressed      bool
	focusByKeyboard bool
	// KeyboardNavigation says when the interface reads Tab, Enter and Space (focus.go).
	KeyboardNavigation KeyboardNavigation
	keyboardGiven      bool // TakeKeyboard gave the interface the keyboard
	hasKeyboard        bool // the interface reads Tab, Enter and Space in this frame
	focusAtBegin       ID   // the field, list or menu that had the keys when the frame began
	keyboardFocus      ID   // the control Tab reached, which Enter and Space press
	keyboardFocusSeen  bool // it was built this frame; if not, it went away
	// keyboardFocusPosition is where it stood in the Tab order; resumeStop is 1 + that position
	// once it went away, so the next Tab goes on from there (0: nothing to resume)
	keyboardFocusPosition, resumeStop int
	keyboardByTab                     bool // a Tab moved the focus: the interface keeps the keys until Escape or a click
	keyboardUsed                      bool // Enter, Space or Escape did something in the interface in this frame
	clickedStop                       ID   // the last Tab stop pressed with the pointer: Tab goes on from it
	outOfTabOrder                     int  // inside withoutTab: controls built now are not Tab stops
	heldControl                       ID   // the control held down in this frame, which PressScale shrinks
	// LessMotion keeps controls still when the person asks the system for less motion: a held
	// control does not shrink. The program sets it from the platform's setting.
	LessMotion   bool
	sceneDrag    bool
	insets       Insets // what the system draws over the window's edges (SetInsets)
	cursor       input.CursorShape
	cursorWasSet bool // the interface set the pointer's shape, so it puts the arrow back when it leaves
	panels       map[ID]*panelMemory
	memory       map[ID]*remembered
	popups       []ID // popups that took part in the last frame, back to front
	nextPopups   []ID
	// overlays are the overlays shown in the last frame and in this one, back to front
	overlays, nextOverlays []ID
	scrollUsed             bool                 // a scroll area took this step's wheel movement
	finger                 finger               // the one finger on a touch screen (gestures.go)
	leaving                map[ID]*leavingLayer // layers that keep their last picture to leave with (leave.go)
	textScale              float32              // the person's text size (SetTextScale); 0 = 1
	press                  press                // the pointer held down, for a long press (longpress.go)
	longPressed            ID                   // the control long-pressed in this frame
	lastHot                ID                   // the control under the pointer in the last frame
	hotSince               time.Time            // when the pointer came onto it
	interacted             ID                   // the last control that looked at the mouse, for Tooltip
	// surfaces are the painted boxes on the screen (outside panels) in the last frame: bars and
	// side columns of an application layout. The pointer over one belongs to the interface.
	surfaces, nextSurfaces []draw2d.Rect
	overSurface            bool

	wantsMouse, wantsKeyboard bool
	typing                    bool // a text field has the focus in this frame
	keyboard                  keyboardState
	editingWords              EditingWords // the bar's labels over a selection on a touch screen

	sheets, nextSheets []ID // sheets shown in the last frame and in this one, back to front
	// the fields of the front sheet, as a range of focusables, so Tab stays inside it
	sheetFocusStart, sheetFocusEnd int
	drag                           dragState
	menu                           *menuMemory // the menu whose items are being built, if any
	menuItems                      int         // items added to it so far
	table                          Table       // the table being built (see Table)
	finished                       int         // the element closed last, which Describe describes
	reader                         readerState // what a screen reader is told and asks (accessibility.go)
}

// New returns an interface that measures text with go-renderer's fonts, in the default theme.
func New() (*Context, error) {
	measurer, err := overlay.NewMeasurer()
	if err != nil {
		return nil, err
	}
	return NewWithMeasurer(measurer), nil
}

// NewWithMeasurer returns an interface that measures text with a measurer of your choice: tests
// use a fake one, so they need no fonts.
func NewWithMeasurer(measurer draw2d.Measurer) *Context {
	return &Context{
		Theme:         LightTheme(),
		measurer:      measurer,
		texts:         textCache{entries: map[textKey]draw2d.TextMetrics{}, older: map[textKey]draw2d.TextMetrics{}},
		seen:          map[ID]int{},
		previous:      map[ID]draw2d.Rect{},
		current:       map[ID]draw2d.Rect{},
		previousReach: map[ID]draw2d.Rect{},
		currentReach:  map[ID]draw2d.Rect{},
		panels:        map[ID]*panelMemory{},
		memory:        map[ID]*remembered{},
	}
}

// Drawing is the draw list the interface paints into; give it to the scene's overlay once
// (world.Overlay.Drawing = ui.Drawing()) and go-renderer draws it after every frame.
func (ui *Context) Drawing() *draw2d.List { return &ui.drawing }

// WantsMouse tells whether the interface used the mouse in the last frame: the pointer was over
// a panel or a control, or dragging one. The camera should leave the mouse alone then; End also
// claims it in the input state, which go-renderer's viewer respects.
func (ui *Context) WantsMouse() bool { return ui.wantsMouse }

// WantsKeyboard tells whether a control had the keyboard in the last frame (a text field being
// typed in).
func (ui *Context) WantsKeyboard() bool { return ui.wantsKeyboard }

// Typing tells whether a text field had the focus in the last frame: a phone shows its on-screen
// keyboard while it does. WantsKeyboard is wider (a sheet open, a list with the focus).
func (ui *Context) Typing() bool { return ui.typing }

// Begin starts describing the interface for a frame, with the frame's input. Call it in every
// update step, before the controls.
func (ui *Context) Begin(state *input.State) {
	ui.state = state
	ui.frame++
	ui.beginReading()
	ui.typing = false
	ui.beginKeyboard()
	ui.beginFocus()
	ui.now = state.Time()
	ui.texts.turnOver()
	ui.building = true
	ui.width, ui.height = float32(state.WindowWidth), float32(state.WindowHeight)
	ui.scale = 1
	if state.WindowWidth > 0 && state.FramebufferWidth > 0 {
		ui.scale = float32(state.FramebufferWidth) / float32(state.WindowWidth)
	}
	ui.elements, ui.open, ui.roots, ui.scopes = ui.elements[:0], ui.open[:0], ui.roots[:0], ui.scopes[:0]
	clear(ui.seen)
	ui.hot = 0
	ui.focusSeen, ui.focusables = false, ui.focusables[:0]
	ui.scrollUsed = false
	ui.nextPopups, ui.nextSheets, ui.nextOverlays = ui.nextPopups[:0], ui.nextSheets[:0], ui.nextOverlays[:0]
	ui.sheetFocusStart, ui.sheetFocusEnd = 0, 0
	ui.cursor = input.CursorArrow
	ui.findHoveredRoot()
	ui.trackFinger()
	ui.followPress()
	ui.followPresses()
	ui.followDrag()
	ui.openScreen()
}

// End lays the interface out, paints it into the draw list and tells the input state whether the
// interface has the mouse or the keyboard.
func (ui *Context) End() {
	if !ui.building {
		return
	}
	ui.building = false
	ui.showDrag()
	ui.layOut()
	ui.readNodes()
	ui.paint()
	ui.paintReaderFocus()
	if ui.moving() {
		ui.RequestFrame(16 * time.Millisecond) // about sixty frames a second while something eases
	}
	ui.previous, ui.current = ui.current, ui.previous
	clear(ui.current)
	ui.previousReach, ui.currentReach = ui.currentReach, ui.previousReach
	clear(ui.currentReach)
	if !ui.focusSeen {
		ui.focus = 0
	}
	ui.endFocus()
	ui.moveFocusByTab()
	ui.finishInteraction()
	ui.popups, ui.nextPopups = ui.nextPopups, ui.popups
	ui.sheets, ui.nextSheets = ui.nextSheets, ui.sheets
	ui.overlays, ui.nextOverlays = ui.nextOverlays, ui.overlays
	ui.forgetUnusedPanels()
	ui.forgetUnusedMemory()
}

// layOut sizes and places every root's tree: the screen covers the window, panels sit where
// they were left or docked.
func (ui *Context) layOut() {
	for rootIndex := range ui.roots {
		placed := &ui.roots[rootIndex]
		ui.size(placed, Horizontal)
		ui.size(placed, Vertical)
		ui.positionRoot(placed)
		ui.place(placed.element)
	}
	ui.recordScrollBars()
	ui.nextSurfaces = ui.nextSurfaces[:0]
	screen := ui.roots[0].element
	for index := range ui.elements {
		if id := ui.elements[index].id; id != 0 {
			ui.current[id] = ui.elements[index].rect
		}
		if ui.elements[index].look.Fill.IsVisible() && index != screen && ui.rootElementOf(index) == screen {
			ui.nextSurfaces = append(ui.nextSurfaces, ui.elements[index].rect)
		}
	}
	ui.surfaces, ui.nextSurfaces = ui.nextSurfaces, ui.surfaces
}

// rootElementOf is the index of the root at the top of an element's tree.
func (ui *Context) rootElementOf(index int) int {
	for ui.elements[index].parent >= 0 {
		index = ui.elements[index].parent
	}
	return index
}
