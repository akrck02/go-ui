package window

import "time"

// defaultPhoneFrameRate is how many frames a second a phone's window draws at most unless told
// otherwise. Phones refresh at 90 or 120 Hz; drawing every refresh while a finger drags a 3D view
// costs over a whole core (measured on a 120 Hz phone), and 60 looks as smooth to the eye.
const defaultPhoneFrameRate = 60

// framePacer keeps frames at least an interval apart: a cap on the frames a second.
type framePacer struct {
	interval time.Duration // 0 = no cap
	last     time.Time     // when the last frame was drawn
}

func newFramePacer(framesPerSecond int) framePacer {
	if framesPerSecond <= 0 {
		return framePacer{}
	}
	return framePacer{interval: time.Second / time.Duration(framesPerSecond)}
}

// wait is how long until the next frame may be drawn; 0 when it may be drawn now.
func (pacer *framePacer) wait(now time.Time) time.Duration {
	if pacer.interval == 0 || pacer.last.IsZero() {
		return 0
	}
	return max(pacer.last.Add(pacer.interval).Sub(now), 0)
}

// drawn records that a frame was drawn now.
func (pacer *framePacer) drawn(now time.Time) { pacer.last = now }
