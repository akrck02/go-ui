package ui

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// Dragging and dropping: a box made a drag source (DragSource) carries a value of the program's
// (the payload). Pressing it and moving the pointer a few pixels starts a drag: a small label with
// the source's name follows the pointer, and the interface holds the mouse, so cameras stay still.
// Letting go over a drop target (DropTarget) hands it the payload; letting go over the open screen
// (no panel, no painted box, no target) leaves the payload for the program, which may drop it on
// its scene (DroppedOnScreen). Escape cancels the drag.

// dragThreshold is how far, in pixels, the pointer moves with the button down before a press on a
// drag source becomes a drag; less is a click.
const dragThreshold = 4

type dragPhase int

const (
	dragIdle      dragPhase = iota
	dragMoving              // a payload follows the pointer
	dragDropped             // the button was let go this step: the payload is being dropped
	dragCancelled           // Escape cancelled the drag; waiting for the button to be let go
	dragFinished            // the button was let go after a cancelled drag, this step
)

// dragState is the drag in progress, if any.
type dragState struct {
	phase dragPhase
	// armed is the source a press landed on, until the button is let go; pressX and pressY are
	// where, to tell a drag from a click.
	armed          ID
	pressX, pressY float32
	source         ID
	payload        any
	label          string
	// targets are the drop targets built this step, and lastTargets those of the last one, to
	// tell whether the pointer is over one before they are built again; over is the target under
	// the pointer this step (the innermost one).
	targets, lastTargets []ID
	over                 ID
	overScreen           bool // the pointer is over the open screen: a release there drops on the scene
}

// Drag is a payload being dragged, or dropped, and where the pointer is, in layout pixels.
type Drag struct {
	Payload any
	X, Y    float32
	// OverScreen tells whether the pointer is over the open screen, where no panel, painted box
	// or drop target is: where letting go drops the payload on the program's scene.
	OverScreen bool
}

// holdsMouse tells whether a drag keeps the mouse from the scene: while it moves, on the step it
// drops (the release is the drop's, not a click on the scene) and until the button is let go
// after Escape.
func (drag *dragState) holdsMouse() bool { return drag.phase != dragIdle }

// followDrag moves a drag on at the start of a step: it ends the drop of the last step, cancels
// with Escape, drops when the button is let go, and notes whether the pointer is over the open
// screen.
func (ui *Context) followDrag() {
	drag := &ui.drag
	drag.targets, drag.lastTargets = drag.lastTargets[:0], drag.targets
	drag.over = 0
	held := ui.state.Button(input.MouseLeft)
	switch drag.phase {
	case dragDropped, dragFinished: // the step of the release is over
		ui.endDrag()
	case dragCancelled:
		if !held {
			drag.phase = dragFinished // the release still belongs to the drag, not to a click
		}
	}
	if !held && drag.phase == dragIdle {
		drag.armed = 0
	}
	if drag.phase != dragMoving {
		return
	}
	drag.overScreen = ui.pointerOverOpenScreen()
	switch {
	case ui.state.Pressed(input.KeyEscape):
		drag.phase, drag.payload = dragCancelled, nil
	case !held:
		drag.phase = dragDropped
	}
}

// endDrag forgets a finished drag.
func (ui *Context) endDrag() {
	ui.drag.phase, ui.drag.source, ui.drag.payload, ui.drag.label = dragIdle, 0, nil, ""
	ui.drag.armed, ui.drag.overScreen = 0, false
}

// pointerOverOpenScreen tells whether the pointer is where the scene shows: over no panel, popup
// or sheet, no painted box of the screen, and no drop target of the last step.
func (ui *Context) pointerOverOpenScreen() bool {
	if ui.hoveredRoot != screenID || ui.overSurface {
		return false
	}
	x, y := float32(ui.state.CursorX), float32(ui.state.CursorY)
	for _, target := range ui.drag.lastTargets {
		if area, shown := ui.previous[target]; shown && area.Contains(x, y) {
			return false
		}
	}
	return true
}

// DragSource adds a box whose content can be dragged away, carrying payload: press it and move
// the pointer a few pixels. name labels the drag (it follows the pointer) and names the box. The
// content may hold buttons and chips, which still take clicks; a drag that starts on them takes
// the mouse from them. Controls that drag themselves (a slider) do not belong in one. It returns
// true on the step it is clicked without dragging.
func (ui *Context) DragSource(name string, payload any, layout Layout, content func()) bool {
	id := ui.idFor(name)
	_, result := ui.pressable(id, layout, Look{}, content)
	ui.armDrag(id)
	drag := &ui.drag
	if drag.phase == dragIdle && drag.armed == id && ui.state.Button(input.MouseLeft) && ui.movedPastThreshold() {
		drag.phase, drag.source, drag.payload, drag.label = dragMoving, id, payload, displayText(name)
		drag.overScreen = false
		ui.active = id // the drag takes the mouse from whatever inside was pressed
	}
	dragged := drag.source == id && drag.phase != dragIdle
	return result.clicked && !dragged
}

// armDrag notes that a press landed on a source, the innermost one first: the press may become a
// drag once the pointer moves.
func (ui *Context) armDrag(id ID) {
	drag := &ui.drag
	if drag.armed != 0 || drag.phase != dragIdle || ui.sceneDrag || !ui.state.ButtonPressed(input.MouseLeft) {
		return
	}
	x, y := float32(ui.state.CursorX), float32(ui.state.CursorY)
	if area, shown := ui.previous[id]; !shown || !area.Contains(x, y) || ui.hoveredRoot != ui.currentRoot() {
		return
	}
	drag.armed, drag.pressX, drag.pressY = id, x, y
}

// movedPastThreshold tells whether the pointer went far enough from the press to make it a drag.
func (ui *Context) movedPastThreshold() bool {
	deltaX := float32(ui.state.CursorX) - ui.drag.pressX
	deltaY := float32(ui.state.CursorY) - ui.drag.pressY
	return deltaX*deltaX+deltaY*deltaY >= dragThreshold*dragThreshold
}

// DropTarget adds a box that takes dragged payloads. While a payload it accepts is over it, it is
// highlighted; on the step one is let go over it, it returns the payload and true. accepts tells
// which payloads it takes (nil takes any), so a list of names can refuse a colour. Targets inside
// targets work: the innermost under the pointer takes the drop.
func (ui *Context) DropTarget(name string, layout Layout, accepts func(payload any) bool, content func()) (payload any, dropped bool) {
	id := ui.idFor(name)
	index := ui.frameWithID(id, layout, Look{}, content)
	drag := &ui.drag
	drag.targets = append(drag.targets, id)
	if drag.phase != dragMoving && drag.phase != dragDropped {
		return nil, false
	}
	if drag.over != 0 || (accepts != nil && !accepts(drag.payload)) {
		return nil, false
	}
	x, y := float32(ui.state.CursorX), float32(ui.state.CursorY)
	if area, shown := ui.previous[id]; !shown || !area.Contains(x, y) || ui.hoveredRoot != ui.currentRoot() {
		return nil, false
	}
	drag.over = id
	if drag.phase == dragDropped {
		return drag.payload, true
	}
	theme := ui.Theme
	ui.elements[index].look = Look{Fill: theme.PrimaryTint, Border: theme.Primary, BorderWidth: 1.5, Radius: theme.ControlRadius}
	return nil, false
}

// Dragging tells whether a payload is being dragged, and where the pointer is: a program shows
// where it would land on its scene while Drag.OverScreen is true.
func (ui *Context) Dragging() (Drag, bool) {
	if ui.drag.phase != dragMoving {
		return Drag{}, false
	}
	return ui.dragNow(), true
}

// DroppedOnScreen tells, on the step the button is let go, whether a payload was dropped on the
// open screen, with the pointer's place: the program drops it on its scene there. It answers from
// Begin to the next Begin, so it can be asked after End, next to the scene's own code.
func (ui *Context) DroppedOnScreen() (Drag, bool) {
	if ui.drag.phase != dragDropped || !ui.drag.overScreen {
		return Drag{}, false
	}
	return ui.dragNow(), true
}

func (ui *Context) dragNow() Drag {
	return Drag{Payload: ui.drag.payload, X: float32(ui.state.CursorX), Y: float32(ui.state.CursorY), OverScreen: ui.drag.overScreen}
}

// showDrag adds, at the end of a step, the label that follows the pointer while dragging, and the
// pointer's shape. go-renderer has no closed-hand shape, so a drag shows the four arrows of
// something being moved.
func (ui *Context) showDrag() {
	drag := &ui.drag
	if drag.phase != dragMoving {
		return
	}
	ui.cursor = input.CursorMove
	theme := ui.Theme
	x, y := float32(ui.state.CursorX), float32(ui.state.CursorY)
	anchor := draw2d.Rectangle{X: x + 12, Y: y + 4}
	look := Look{Fill: theme.Surface, Border: theme.Primary, BorderWidth: 1, Radius: chipHeight / 2, Shadow: true}
	ui.popup(childID(drag.source, "##dragging"), anchor, 0, true, look, func() {
		ui.Box(Layout{Height: Fixed(chipHeight - 8), Padding: PadXY(6, 0), AlignY: Center}, func() {
			ui.Text(drag.label, TextStyle{Size: theme.SmallTextSize + 1, Weight: draw2d.Medium})
		})
	})
}
