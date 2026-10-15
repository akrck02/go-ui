package ui

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// interaction is what the mouse did to a control during this step.
type interaction struct {
	hovered bool // the pointer is over it, and nothing else holds the mouse
	pressed bool // the left button went down on it
	held    bool // it is being pressed (the button went down on it and is still down)
	clicked bool // the button went down on it and up over it
}

// interact works out what the mouse does to the control with an id, from where the control was
// drawn in the last frame. A control pressed keeps the mouse until the button is let go, even if
// the pointer leaves it, so a drag that strays does not wake up other controls. The first control
// to find the pointer over it takes it: containers that react to the mouse (a panel's title bar)
// call interact after their content, so the buttons inside them win.
func (ui *Context) interact(id ID) interaction {
	ui.noteControl(id)
	return ui.pointerOn(id)
}

// pointerOn is interact for what takes the mouse without being a control a screen reader reads: a
// modal's backdrop, a card that keeps a touch from it, a panel's title bar.
func (ui *Context) pointerOn(id ID) interaction {
	area, shown := ui.reachOf(id)
	x, y := float32(ui.state.CursorX), float32(ui.state.CursorY)
	over := shown && area.Contains(x, y) && ui.hoveredRoot == ui.currentRoot() && !ui.sceneDrag
	var result interaction
	ui.interacted = id
	result.hovered = over && (ui.active == 0 || ui.active == id) && (ui.hot == 0 || ui.hot == id)
	if result.hovered {
		ui.hot = id
	}
	if result.hovered && ui.state.ButtonPressed(input.MouseLeft) {
		ui.active = id
		result.pressed = true
	}
	if ui.active == id {
		result.held = ui.state.Button(input.MouseLeft)
		result.clicked = ui.state.ButtonReleased(input.MouseLeft) && over && !ui.press.long
	}
	if ui.readerActs(id, ReaderPress) {
		// a screen reader's double tap: pressed and let go at once, where the control is
		result.pressed, result.clicked = true, true
	}
	ui.noteLongPress(id, result)
	return result
}

// finishInteraction ends the step's mouse handling: it lets go of a released control, notes
// whether a press on empty screen started a drag that belongs to the scene, and claims the mouse
// and keyboard in the input state when the interface has them.
func (ui *Context) finishInteraction() {
	state := ui.state
	pressedNow := state.ButtonPressed(input.MouseLeft)
	if pressedNow && ui.active == 0 {
		switch {
		case ui.hoveredRoot == screenID && ui.overSurface:
			ui.active = screenID // a press on a bar or a side column keeps the mouse from the scene
		case ui.hoveredRoot == screenID:
			ui.sceneDrag = true
		default:
			ui.active = ui.hoveredRoot // a press on a panel's background keeps the mouse from the scene
		}
	}
	if !state.Button(input.MouseLeft) {
		ui.active = 0
		ui.sceneDrag = false
	}
	sheetOpen := len(ui.nextSheets) > 0
	ui.wantsMouse = !ui.sceneDrag && (ui.hoveredRoot != screenID || ui.overSurface || ui.hot != 0 || ui.active != 0)
	ui.wantsMouse = ui.wantsMouse || sheetOpen || ui.drag.holdsMouse()
	ui.wantsKeyboard = ui.focus != 0 || sheetOpen || ui.drag.phase == dragMoving
	if ui.hot != ui.lastHot {
		ui.lastHot, ui.hotSince = ui.hot, ui.now
	}
	if ui.wantsMouse {
		state.ClaimMouse()
		state.SetCursorShape(ui.cursor)
	} else if state.CursorShape() != input.CursorArrow && ui.cursorWasSet {
		state.SetCursorShape(input.CursorArrow)
	}
	ui.cursorWasSet = ui.wantsMouse
	if ui.wantsKeyboard {
		state.ClaimKeyboard()
	}
}

// reachOf is where the pointer could reach a thing in the last frame: the part that showed, when a
// box around it clipped it (a list scrolled), else all of it.
func (ui *Context) reachOf(id ID) (draw2d.Rect, bool) {
	if area, clipped := ui.previousReach[id]; clipped {
		return area, !area.IsEmpty()
	}
	area, shown := ui.previous[id]
	return area, shown
}
