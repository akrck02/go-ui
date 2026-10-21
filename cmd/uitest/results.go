package main

import (
	"sort"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-ui/uitest"
)

// results is what the viewer shows: the recorded tests, the one chosen and the frame shown.
type results struct {
	recordings []uitest.Recording
	selected   int
	frame      int
	frameValue float32 // the frame as the slider holds it
	playback   playback
	// measurer measures text with the viewer's fonts, to fit the replay's lines (replay.measurer)
	measurer draw2d.Measurer
}

// newResults orders the recordings for the viewer: failed tests first, each group by name.
func newResults(recordings []uitest.Recording) *results {
	sorted := append([]uitest.Recording(nil), recordings...)
	sort.SliceStable(sorted, func(first, second int) bool {
		if sorted[first].Failed != sorted[second].Failed {
			return sorted[first].Failed
		}
		return sorted[first].Test < sorted[second].Test
	})
	return &results{recordings: sorted}
}

// current is the chosen test, if there is any test.
func (results *results) current() (uitest.Recording, bool) {
	if results.selected < 0 || results.selected >= len(results.recordings) {
		return uitest.Recording{}, false
	}
	return results.recordings[results.selected], true
}

// currentFrame is the frame shown of the chosen test.
func (results *results) currentFrame() (uitest.RecordedFrame, bool) {
	recording, found := results.current()
	if !found || results.frame >= len(recording.Frames) {
		return uitest.RecordedFrame{}, false
	}
	return recording.Frames[results.frame], true
}

// choose shows another test, from its first frame, paused.
func (results *results) choose(index int) {
	results.selected = index
	results.playback.playing = false
	results.goTo(0)
}

// step moves through the frames of the chosen test, stopping at the first and the last.
func (results *results) step(delta int) { results.goTo(results.frame + delta) }

// goTo shows a frame of the chosen test, kept within its frames.
func (results *results) goTo(frame int) {
	recording, _ := results.current() // no test: no frames, and the frame stays 0
	results.frame = max(0, min(frame, len(recording.Frames)-1))
	results.frameValue = float32(results.frame)
}

// counts is how many tests passed and how many failed.
func (results *results) counts() (passed, failed int) {
	for _, recording := range results.recordings {
		if recording.Failed {
			failed++
			continue
		}
		passed++
	}
	return passed, failed
}
