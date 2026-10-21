package uitest

import "time"

// frameDuration is how much time one frame of a Wait takes: sixty frames a second.
const frameDuration = time.Second / 60

// clockStart is when every test window's clock starts, so runs are alike.
var clockStart = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC)

// clock is a test window's time: it moves only when a test waits.
type clock struct {
	now time.Time
}

// read tells the time; the window gives it to the input state.
func (clock *clock) read() time.Time { return clock.now }

// Now is the window's time.
func (window *Window) Now() time.Time { return window.clock.now }

// Wait lets time pass: it runs a frame about every sixtieth of a second until the duration has
// gone, so long presses, double-click intervals and animations see time move as on a screen. The
// last frame sees the time the wait ends at.
func (window *Window) Wait(duration time.Duration) {
	if duration <= 0 {
		return
	}
	start := window.clock.now
	frames := max(1, int((duration+frameDuration/2)/frameDuration))
	for frame := 1; frame <= frames; frame++ {
		window.clock.now = start.Add(duration * time.Duration(frame) / time.Duration(frames))
		window.Frame()
	}
}
