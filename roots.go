package ui

import (
	"slices"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// A root is the top of a tree of boxes laid out on its own: the screen (the whole window, for
// controls outside any panel) or a panel floating over it.
type root struct {
	id      ID
	element int
	panel   *panelMemory // nil for the screen and popups
	dock    Edge
	popup   bool         // floats over everything, next to anchor
	sheet   bool         // covers the window over every panel, under the popups (see Sheet)
	passive bool         // a popup that never takes the mouse (a tooltip)
	centred bool         // a popup centred on its anchor's x (a toast)
	anchor  draw2d.Rect  // what a popup opens next to
	placed  bool         // an overlay, placed in the window by options (see Overlay)
	options LayerOptions // where an overlay goes, and how a modal or an overlay leaves
}

// Edge is a side of the window a panel can be docked to.
type Edge int

const (
	Floating Edge = iota // not docked: the panel floats where it is put, and can be dragged
	DockLeft
	DockRight
	DockTop
	DockBottom
)

// PanelOptions place and size a panel.
type PanelOptions struct {
	X, Y   float32 // where it first appears, in pixels from the window's top-left corner
	Width  Sizing  // zero = as wide as its content
	Height Sizing  // zero = as tall as its content
	Dock   Edge    // docked panels fill a side of the window and cannot be dragged
	Locked bool    // a floating panel that cannot be dragged
}

// panelMemory is what a panel keeps between frames.
type panelMemory struct {
	x, y      float32
	collapsed bool
	lastFrame int
}

// panelsForgottenAfter is how many frames a panel may go unseen before its position is forgotten.
const panelsForgottenAfter = 3600

// screenID names the screen root.
var screenID = hashString(fnvOffset, "##screen")

// openScreen starts the screen root, where controls outside any panel go.
func (ui *Context) openScreen() {
	index := ui.openElement(element{id: 0, kind: boxElement, layout: Layout{Direction: Vertical}})
	ui.roots = append(ui.roots, root{id: screenID, element: index})
}

// currentRoot is the root whose boxes are being added now.
func (ui *Context) currentRoot() ID {
	if len(ui.open) == 0 {
		return screenID
	}
	for rootIndex := len(ui.roots) - 1; rootIndex >= 0; rootIndex-- {
		if ui.roots[rootIndex].element == ui.open[0] {
			return ui.roots[rootIndex].id
		}
	}
	return screenID
}

// Panel adds a floating panel with a title bar; drag the bar to move it, click the chevron to fold
// it. What content adds goes inside, in a column.
func (ui *Context) Panel(title string, content func()) {
	cascade := float32(len(ui.roots)-1) * 24
	ui.PanelWith(title, PanelOptions{X: 16 + cascade, Y: 16 + cascade}, content)
}

// PanelWith adds a panel placed and sized by options.
func (ui *Context) PanelWith(title string, options PanelOptions, content func()) {
	id := ui.idFor(title)
	memory := ui.panelMemoryFor(id, options)
	if !slices.Contains(ui.order, id) {
		ui.order = append(ui.order, id)
	}
	outside := ui.open
	ui.open = nil
	index := ui.openElement(element{id: id, kind: boxElement, layout: ui.panelLayout(options), look: ui.panelLook(options)})
	ui.roots = append(ui.roots, root{id: id, element: index, panel: memory, dock: options.Dock})
	ui.Push(title)
	ui.titleBar(title, memory, options)
	if !memory.collapsed {
		ui.Separator()
		ui.Box(Layout{Direction: Vertical, Width: Grow(), Height: Grow(), Padding: Pad(ui.Theme.Padding), Gap: ui.Theme.Gap}, content)
	}
	ui.Pop()
	ui.open = outside
}

func (ui *Context) panelMemoryFor(id ID, options PanelOptions) *panelMemory {
	memory, known := ui.panels[id]
	if !known {
		memory = &panelMemory{x: options.X, y: options.Y}
		ui.panels[id] = memory
	}
	memory.lastFrame = ui.frame
	return memory
}

func (ui *Context) panelLayout(options PanelOptions) Layout {
	layout := Layout{Direction: Vertical, Width: options.Width, Height: options.Height}
	if layout.Width.kind == fitSizing && layout.Width.Min == 0 {
		layout.Width.Min = ui.Theme.PanelMinWidth
	}
	return layout
}

func (ui *Context) panelLook(options PanelOptions) Look {
	look := Look{Fill: ui.Theme.Surface, Border: ui.Theme.Line, BorderWidth: 1, Radius: ui.Theme.PanelRadius, Shadow: true, Clip: true}
	if options.Dock != Floating {
		look.Radius, look.Shadow = 0, false
	}
	return look
}

// titleBar draws the panel's title and fold button, and moves the panel while the bar is dragged.
// The bar looks at the mouse after its fold button, so a press on the button is the button's: the
// innermost control gets the mouse first.
func (ui *Context) titleBar(title string, memory *panelMemory, options PanelOptions) {
	barID := ui.idFor("##title")
	layout := Layout{Direction: Horizontal, Width: Grow(), Padding: Padding{ui.Theme.Padding, ui.Theme.Padding / 2, ui.Theme.Padding / 2, ui.Theme.Padding / 2}, Gap: ui.Theme.Gap, AlignY: Center}
	ui.frameWithID(barID, layout, Look{}, func() {
		ui.Text(displayText(title), TextStyle{Weight: draw2d.Bold})
		ui.Spacer()
		fold := "chevron-down"
		if memory.collapsed {
			fold = "chevron-right"
		}
		if ui.flatIconButton("##fold", fold) {
			memory.collapsed = !memory.collapsed
		}
	})
	bar := ui.pointerOn(barID)
	canMove := options.Dock == Floating && !options.Locked
	if canMove && (bar.hovered || ui.active == barID) {
		ui.cursor = input.CursorMove
	}
	if canMove && bar.held {
		memory.x += float32(ui.state.CursorDeltaX)
		memory.y += float32(ui.state.CursorDeltaY)
	}
}

// rootSize is a root's size along an axis once its content is measured.
func (ui *Context) rootSize(placed *root, axis Direction) float32 {
	window := ui.windowAlong(axis)
	if placed.popup {
		return min(ui.elements[placed.element].sizeAlong(axis), window)
	}
	if placed.panel == nil {
		return window
	}
	element := &ui.elements[placed.element]
	if fillsAxis(placed.dock, axis) {
		return window
	}
	sizing := element.layout.sizing(axis)
	switch sizing.kind {
	case percentSizing:
		return sizing.clamp(window * sizing.value)
	case growSizing:
		return sizing.clamp(window)
	default:
		return min(element.sizeAlong(axis), window)
	}
}

// fillsAxis tells whether a panel docked to an edge spans the window along an axis.
func fillsAxis(dock Edge, axis Direction) bool {
	switch dock {
	case DockLeft, DockRight:
		return axis == Vertical
	case DockTop, DockBottom:
		return axis == Horizontal
	default:
		return false
	}
}

func (ui *Context) windowAlong(axis Direction) float32 {
	if axis == Horizontal {
		return ui.width
	}
	return ui.height
}

// positionRoot puts a root's box in the window: the screen at the corner, a docked panel against
// its edge, a floating panel where it was left, kept within reach (its title bar on screen).
func (ui *Context) positionRoot(placed *root) {
	element := &ui.elements[placed.element]
	if placed.placed {
		ui.placeInWindow(placed)
		return
	}
	if placed.popup {
		ui.positionPopup(placed)
		return
	}
	if placed.panel == nil {
		element.rect.X, element.rect.Y = 0, 0
		return
	}
	switch placed.dock {
	case DockLeft, DockTop:
		element.rect.X, element.rect.Y = 0, 0
	case DockRight:
		element.rect.X, element.rect.Y = ui.width-element.rect.Width, 0
	case DockBottom:
		element.rect.X, element.rect.Y = 0, ui.height-element.rect.Height
	default:
		reach := ui.Theme.ControlHeight * 2
		placed.panel.x = clamp(placed.panel.x, reach-element.rect.Width, ui.width-reach)
		placed.panel.y = clamp(placed.panel.y, 0, ui.height-reach)
		element.rect.X, element.rect.Y = placed.panel.x, placed.panel.y
	}
}

func clamp(value, low, high float32) float32 {
	if high < low {
		return low
	}
	return min(max(value, low), high)
}

// layerOf is the drawing layer of a root: the screen at the bottom, panels from back to front.
// Panels are ten layers apart, so a panel's menus and tooltips can sit just above it.
func (ui *Context) layerOf(id ID) int {
	if id == screenID {
		return 0
	}
	if position := slices.Index(ui.nextPopups, id); position >= 0 {
		return popupLayer + position
	}
	if position := slices.Index(ui.nextSheets, id); position >= 0 {
		return sheetLayer + position*10
	}
	for _, placed := range ui.roots {
		if placed.id == id && placed.popup {
			return popupLayer + len(ui.nextPopups) + 1 // a passive popup, over everything
		}
	}
	return (slices.Index(ui.order, id) + 1) * 10
}

// findHoveredRoot finds the front-most panel under the pointer, from where panels were in the
// last frame; with none, the pointer is over the screen.
func (ui *Context) findHoveredRoot() {
	ui.hoveredRoot = screenID
	x, y := float32(ui.state.CursorX), float32(ui.state.CursorY)
	ui.overSurface = false
	for _, surface := range ui.surfaces {
		ui.overSurface = ui.overSurface || surface.Contains(x, y)
	}
	for position := len(ui.popups) - 1; position >= 0; position-- {
		if area, shown := ui.previous[ui.popups[position]]; shown && area.Contains(x, y) {
			ui.hoveredRoot = ui.popups[position]
			return
		}
	}
	if len(ui.sheets) > 0 {
		ui.hoveredRoot = ui.sheets[len(ui.sheets)-1] // the front sheet covers the window
		return
	}
	for position := len(ui.order) - 1; position >= 0; position-- {
		if area, shown := ui.previous[ui.order[position]]; shown && area.Contains(x, y) {
			ui.hoveredRoot = ui.order[position]
			return
		}
	}
}

// followPresses brings a panel to the front when it is pressed.
func (ui *Context) followPresses() {
	if !ui.state.ButtonPressed(input.MouseLeft) || ui.hoveredRoot == screenID || ui.sceneDrag {
		return
	}
	position := slices.Index(ui.order, ui.hoveredRoot)
	if position >= 0 {
		ui.order = append(slices.Delete(ui.order, position, position+1), ui.hoveredRoot)
	}
}

// forgetUnusedPanels drops what panels unseen for a long time remembered.
func (ui *Context) forgetUnusedPanels() {
	for id, memory := range ui.panels {
		if ui.frame-memory.lastFrame > panelsForgottenAfter {
			delete(ui.panels, id)
			ui.order = slices.DeleteFunc(ui.order, func(ordered ID) bool { return ordered == id })
		}
	}
}
