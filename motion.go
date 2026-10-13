package ui

import "time"

// Motion for controls of a program's own: a value that eases to where it is going instead of
// jumping, and a control that eases in as it appears. Both ask for frames only while they move, so
// a window that draws on demand still sleeps when nothing does.

// animation is what an animated value keeps between frames.
type animation struct {
	from, to  float32
	started   time.Time
	lastFrame int
}

// Animate eases a value towards a target over a duration, starting again from where it is when the
// target changes: a switch's handle going across, a tab's underline moving to the tab chosen. The
// first time a name is seen, the value is at its target. It returns the value for this frame.
func (ui *Context) Animate(name string, target float32, duration time.Duration) float32 {
	memory := remember(ui, ui.idFor("##animate "+name), func() *animation {
		return &animation{from: target, to: target}
	})
	if target != memory.to {
		memory.from, memory.to, memory.started = memory.valueAt(ui.now, duration), target, ui.now
	}
	return memory.valueAt(ui.now, duration)
}

// Appear goes from 0 to 1 over a duration from the frame something was first shown after it was
// not: a dialog fading and rising in, a note sliding up. It starts again when the thing comes back.
func (ui *Context) Appear(name string, duration time.Duration) float32 {
	memory := remember(ui, ui.idFor("##appear "+name), func() *animation { return &animation{lastFrame: -2} })
	if memory.lastFrame != ui.frame-1 { // not shown in the last frame: it appears now
		memory.from, memory.to, memory.started = 0, 1, ui.now
	}
	memory.lastFrame = ui.frame
	return memory.valueAt(ui.now, duration)
}

// valueAt is the value at a time, asking for the next frame while it moves.
func (memory *animation) valueAt(now time.Time, duration time.Duration) float32 {
	if memory.from == memory.to || duration <= 0 {
		return memory.to
	}
	progress := float32(now.Sub(memory.started)) / float32(duration)
	if progress >= 1 {
		memory.from = memory.to
		return memory.to
	}
	return memory.from + (memory.to-memory.from)*easeOut(max(progress, 0))
}

// easeOut starts fast and settles gently, as things that arrive do (a cubic).
func easeOut(progress float32) float32 {
	left := 1 - progress
	return 1 - left*left*left
}

// moving tells whether any value animated in this frame is still on its way: the interface asks
// for the next frame then.
func (ui *Context) moving() bool {
	for _, entry := range ui.memory {
		if memory, isAnimation := entry.value.(*animation); isAnimation && entry.lastFrame == ui.frame && memory.from != memory.to {
			return true
		}
	}
	return false
}

// Back answers the system's back (a phone's back gesture or button) in this frame, and tells
// whether there was one: the part of the interface the person is in goes back (a screen returns to
// the one before) and returns true. Dialogs and sheets answer it themselves. A back nobody answers
// leaves the program, as a phone's own applications do at their first screen.
func (ui *Context) Back() bool {
	if ui.state == nil {
		return false
	}
	return ui.state.TakeBack()
}

// frontSheet tells whether a sheet or modal was the front one in the last frame, the one a back
// closes.
func (ui *Context) frontSheet(id ID) bool {
	return len(ui.sheets) > 0 && ui.sheets[len(ui.sheets)-1] == id
}
