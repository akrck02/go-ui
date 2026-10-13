package ui

import (
	"strconv"
	"testing"
	"time"
)

// dragFinger puts a finger down at a point and moves it by steps, a frame each, then lifts it.
func dragFinger(h *harness, now *time.Time, build func(ui *Context), fromX, fromY, byX, byY float64, lift bool) {
	h.state.SetTouch(1, fromX, fromY)
	h.moveTo(fromX, fromY)
	h.press()
	h.frame(build)
	const steps = 10
	for step := 1; step <= steps; step++ {
		*now = now.Add(16 * time.Millisecond)
		x, y := fromX+byX*float64(step)/steps, fromY+byY*float64(step)/steps
		h.state.SetTouch(1, x, y)
		h.moveTo(x, y)
		h.frame(build)
	}
	if lift {
		*now = now.Add(16 * time.Millisecond)
		h.release()
		h.state.LiftTouch(1)
		h.frame(build)
	}
}

// pages is a swipe box holding a scrolling list: the case where the two ways meet.
func pages(swipes *[]Swipe, clicked *int) func(ui *Context) {
	return func(ui *Context) {
		ui.SwipeBox("pages", Layout{Width: Fixed(400), Height: Fixed(300)}, Horizontal, func(swipe Swipe) {
			if swipe.Held || swipe.Released {
				*swipes = append(*swipes, swipe)
			}
			ui.ScrollArea("list", Layout{Width: Grow(), Height: Grow()}, func() {
				for index := range 40 {
					if ui.Button("row " + strconv.Itoa(index)) {
						*clicked++
					}
				}
			})
		})
	}
}

func TestASidewaysDragGoesToTheSwipeBoxAndAnUpwardOneToTheListInIt(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	var swipes []Swipe
	clicked := 0
	build := pages(&swipes, &clicked)
	h.frame(build)
	h.frame(build)
	dragFinger(h, now, build, 20, 100, -150, 4, true)
	if len(swipes) == 0 || !swipes[len(swipes)-1].Released || swipes[len(swipes)-1].Offset > -130 {
		t.Fatalf("swipes %+v: want the box to follow the finger 150 pixels left and say when it let go", swipes)
	}
	list := remember(h.ui, childID(idOf("pages"), "list"), func() *scrollMemory { return nil })
	if list != nil && list.offset != 0 {
		t.Fatalf("the list scrolled by %v on a sideways drag", list.offset)
	}
	swipes = nil
	dragFinger(h, now, build, 20, 250, 3, -120, true)
	if len(swipes) != 0 {
		t.Fatalf("the box took an upward drag: %+v", swipes)
	}
	if clicked != 0 {
		t.Fatalf("%d rows clicked by drags", clicked)
	}
}

func TestPullingAListPastItsTopIsMeasured(t *testing.T) {
	h := newHarness(t)
	now := h.withClock()
	var pulling, released float32
	build := func(ui *Context) {
		pulling, released = ui.ScrollPull("list")
		ui.ScrollArea("list", Layout{Width: Fixed(300), Height: Fixed(200)}, func() {
			for index := range 40 {
				ui.Label("row " + strconv.Itoa(index))
			}
		})
	}
	h.frame(build)
	h.frame(build)
	dragFinger(h, now, build, 50, 20, 0, 160, false)
	*now = now.Add(16 * time.Millisecond)
	h.frame(build)
	if pulling < 60 {
		t.Fatalf("pulled %v while 160 pixels down; want it measured, at half the finger's way", pulling)
	}
	h.release()
	h.state.LiftTouch(1)
	h.frame(build)
	h.frame(build)
	if released < 60 || pulling != 0 {
		t.Fatalf("released at %v, still pulling %v: want the pull at the lift, and none after", released, pulling)
	}
}
