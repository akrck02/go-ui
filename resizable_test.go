package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
)

func TestDraggingTheEdgeResizesBetweenBounds(t *testing.T) {
	h := newHarness(t)
	width, changes := float32(200), 0
	build := func(ui *Context) {
		ui.Box(Layout{Width: Grow(), Height: Fixed(300)}, func() {
			if ui.Resizable("list", RightSide, &width, 120, 320, Layout{Direction: Vertical, Height: Grow()}, func() {
				ui.frameWithID(idOf("content"), Layout{Width: Grow(), Height: Grow()}, Look{}, nil)
			}) {
				changes++
			}
		})
	}
	h.frame(build)
	if content := h.rect(idOf("content")); content.Width != 200 || content.Height != 300 {
		t.Fatalf("the box is as wide as the program's width: %+v", content)
	}
	handle := h.rect(childID(idOf("list"), "##handle"))
	if handle.X != 200 {
		t.Fatalf("the handle is on the right edge: %+v", handle)
	}
	h.moveTo(float64(handle.X+2), 150)
	h.frame(build)
	if h.state.CursorShape() != input.CursorResizeHorizontal {
		t.Fatal("over the handle the pointer shows it resizes sideways")
	}
	h.press()
	h.frame(build)
	h.moveTo(float64(handle.X+2+50), 150)
	h.frame(build)
	if width != 250 || changes != 1 {
		t.Fatalf("the width follows the pointer: %v", width)
	}
	h.moveTo(float64(handle.X+2+500), 150)
	h.frame(build)
	if width != 320 {
		t.Fatalf("up to the maximum: %v", width)
	}
	h.moveTo(float64(handle.X+2-500), 150)
	h.frame(build)
	if width != 120 {
		t.Fatalf("down to the minimum: %v", width)
	}
	h.release()
	h.frame(build)
	h.frame(build)
	if content := h.rect(idOf("content")); content.Width != 120 {
		t.Fatalf("the box follows: %+v", content)
	}
}

func TestAResizableOnTheRightGrowsToTheLeftAndStripsGrowUp(t *testing.T) {
	h := newHarness(t)
	width, height := float32(200), float32(100)
	build := func(ui *Context) {
		ui.Box(Layout{Width: Grow(), Height: Grow()}, func() {
			ui.Box(Layout{Direction: Vertical, Width: Grow(), Height: Grow()}, func() {
				ui.Spacer()
				ui.Resizable("log", TopSide, &height, 50, 0, Layout{Width: Grow()}, nil)
			})
			ui.Resizable("inspector", LeftSide, &width, 100, 400, Layout{Height: Grow()}, nil)
		})
	}
	h.frame(build)
	drag := func(handle ID, deltaX, deltaY float64) {
		area := h.rect(handle)
		x, y := float64(area.X+area.Width/2), float64(area.Y+area.Height/2)
		h.moveTo(x, y)
		h.frame(build)
		h.press()
		h.frame(build)
		h.moveTo(x+deltaX, y+deltaY)
		h.frame(build)
		h.release()
		h.frame(build)
	}
	drag(childID(idOf("inspector"), "##handle"), -40, 0)
	if width != 240 {
		t.Fatalf("dragging a left edge to the left widens the box: %v", width)
	}
	drag(childID(idOf("log"), "##handle"), 0, -30)
	if height != 130 {
		t.Fatalf("dragging a top edge up makes the strip taller: %v", height)
	}
}
