package window

import (
	"testing"
	"time"
)

func TestFramesAreKeptApartByTheCap(t *testing.T) {
	pacer := newFramePacer(60)
	start := time.Now()
	if pacer.wait(start) != 0 {
		t.Fatal("the first frame waits for nothing")
	}
	pacer.drawn(start)
	wait := pacer.wait(start.Add(5 * time.Millisecond))
	if wait < 11*time.Millisecond || wait > 12*time.Millisecond {
		t.Fatalf("wait = %v, want about 11.7 ms to keep 60 frames a second", wait)
	}
	if pacer.wait(start.Add(20*time.Millisecond)) != 0 {
		t.Fatal("a frame after the interval waits for nothing")
	}
}

func TestNoCapNeverWaits(t *testing.T) {
	pacer := newFramePacer(0)
	now := time.Now()
	pacer.drawn(now)
	if pacer.wait(now) != 0 {
		t.Fatal("without a cap frames are drawn at once")
	}
}
