package window

import "github.com/akrck02/go-renderer/input"

// touchPointer turns fingers into the pointer every control already understands: the first finger
// down moves the cursor and presses the left button, lifting it releases the button. Other fingers
// are ignored until it lifts. It works in layout pixels, so it divides the touch's screen pixels by
// the screen's density.
type touchPointer struct {
	finger int64 // the touch sequence that is the pointer
	down   bool
}

// begin starts a touch; the first finger becomes the pointer.
func (pointer *touchPointer) begin(state *input.State, finger int64, x, y, scale float32) {
	if pointer.down {
		return
	}
	pointer.finger, pointer.down = finger, true
	state.MoveCursor(float64(x/scale), float64(y/scale))
	state.SetButton(input.MouseLeft, true)
}

// move follows the pointer's finger.
func (pointer *touchPointer) move(state *input.State, finger int64, x, y, scale float32) {
	if !pointer.down || finger != pointer.finger {
		return
	}
	state.MoveCursor(float64(x/scale), float64(y/scale))
}

// end lifts the pointer's finger: a tap is a click where the finger was.
func (pointer *touchPointer) end(state *input.State, finger int64, x, y, scale float32) {
	if !pointer.down || finger != pointer.finger {
		return
	}
	pointer.down = false
	state.MoveCursor(float64(x/scale), float64(y/scale))
	state.SetButton(input.MouseLeft, false)
}

// densityScale is layout pixels to screen pixels for a screen of this many pixels per point (a
// point is 1/72 inch). Layout pixels are Android's density-independent pixels, 1/160 inch, so an
// interface is as big on every phone.
func densityScale(pixelsPerPoint float32) float32 {
	if pixelsPerPoint <= 0 {
		return 1
	}
	return pixelsPerPoint * 72 / 160
}
