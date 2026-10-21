package main

import "time"

// Playing a test's frames one after another, as a film: the viewer moves to the next frame each
// time a frame's time has passed, and stops at the last one. A test's frames are its steps (a
// click, a key, a wait), not moments of equal length, so each is shown for the same time, long
// enough to see what changed.

// playbackFrame is how long each frame is shown while the replay plays.
const playbackFrame = 100 * time.Millisecond

// playback is whether the replay plays, and when it last moved to the next frame.
type playback struct {
	playing bool
	movedAt time.Time
	now     func() time.Time // the viewer's clock: time.Now, or a test's
}

// clock is the time now, by the viewer's clock.
func (results *results) clock() time.Time {
	if results.playback.now == nil {
		return time.Now()
	}
	return results.playback.now()
}

// togglePlaying plays the replay, or pauses it; playing from the last frame starts again from the
// first.
func (results *results) togglePlaying() {
	playback := &results.playback
	if playback.playing {
		playback.playing = false
		return
	}
	recording, _ := results.current()
	if results.frame >= len(recording.Frames)-1 {
		results.goTo(0)
	}
	playback.playing, playback.movedAt = true, results.clock()
}

// advance moves the replay on by the frames whose time has passed while it plays, and stops it at
// the last frame. It tells how long until the next frame is due, for the window to come back then.
func (results *results) advance() time.Duration {
	playback := &results.playback
	if !playback.playing {
		return 0
	}
	now := results.clock()
	passed := int(now.Sub(playback.movedAt) / playbackFrame)
	if passed > 0 {
		results.step(passed)
		playback.movedAt = playback.movedAt.Add(time.Duration(passed) * playbackFrame)
	}
	recording, _ := results.current()
	if results.frame >= len(recording.Frames)-1 {
		playback.playing = false
		return 0
	}
	return playbackFrame - now.Sub(playback.movedAt)
}
