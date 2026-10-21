package main

import (
	"sort"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-ui/uitest"
)

// results is what the viewer shows: every test go test ran, the window tests' recordings, the
// test chosen and the frame shown.
type results struct {
	tests      []testResult
	recordings []uitest.Recording
	shown      testFilter // which tests the list shows
	common     string     // the start every package's path shares, which the list leaves out
	selected   int        // the test chosen, by its place in tests
	frame      int
	frameValue float32 // the frame as the slider holds it
	playback   playback
	// measurer measures text with the viewer's fonts, to fit the replay's lines (replay.measurer)
	measurer draw2d.Measurer
}

// testFilter is which tests the list shows.
type testFilter int

const (
	allTests testFilter = iota
	failedTests
	skippedTests
	windowTests
)

// newResults joins the tests go test ran with the recordings their windows made, by the tests'
// names, and orders them: failed tests first, then by package and name. With no tests read (the
// viewer's own tests), each recording stands for its test.
func newResults(tests []testResult, recordings []uitest.Recording) *results {
	if len(tests) == 0 {
		tests = testsOfRecordings(recordings)
	}
	tests = append([]testResult(nil), tests...)
	taken := make([]bool, len(recordings))
	for index := range tests {
		tests[index].recording = recordingOf(tests[index].Name, recordings, taken)
	}
	sort.SliceStable(tests, func(first, second int) bool {
		if (tests[first].State == failedTest) != (tests[second].State == failedTest) {
			return tests[first].State == failedTest
		}
		if tests[first].Package != tests[second].Package {
			return tests[first].Package < tests[second].Package
		}
		return tests[first].Name < tests[second].Name
	})
	return &results{tests: tests, recordings: recordings, common: commonFolder(tests)}
}

// commonFolder is the folders every test's package path starts with ("example.com/game/"), so the
// list names packages short ("engine", "editor").
func commonFolder(tests []testResult) string {
	if len(tests) == 0 {
		return ""
	}
	common := tests[0].Package
	for _, test := range tests[1:] {
		for !strings.HasPrefix(test.Package, common) {
			common = common[:len(common)-1]
		}
	}
	return common[:strings.LastIndex(common, "/")+1] // whole folders: a package keeps its own name
}

// shortPackage is a package's path without the folders every package shares.
func (results *results) shortPackage(path string) string {
	return strings.TrimPrefix(path, results.common)
}

// testsOfRecordings are the tests some recordings tell of.
func testsOfRecordings(recordings []uitest.Recording) []testResult {
	tests := make([]testResult, 0, len(recordings))
	for _, recording := range recordings {
		state := passedTest
		if recording.Failed {
			state = failedTest
		}
		tests = append(tests, testResult{Name: recording.Test, State: state})
	}
	return tests
}

// recordingOf is the place of a test's recording, the first not taken by another test of the
// same name; -1 when the test made none.
func recordingOf(name string, recordings []uitest.Recording, taken []bool) int {
	for index, recording := range recordings {
		if recording.Test == name && !taken[index] {
			taken[index] = true
			return index
		}
	}
	return -1
}

// currentTest is the test chosen, if there is any test.
func (results *results) currentTest() (testResult, bool) {
	if results.selected < 0 || results.selected >= len(results.tests) {
		return testResult{}, false
	}
	return results.tests[results.selected], true
}

// current is the chosen test's recording, when it made a window.
func (results *results) current() (uitest.Recording, bool) {
	test, found := results.currentTest()
	if !found || test.recording < 0 {
		return uitest.Recording{}, false
	}
	return results.recordings[test.recording], true
}

// isShown tells whether the list shows a test, by the filter chosen.
func (results *results) isShown(test testResult) bool {
	switch results.shown {
	case failedTests:
		return test.State == failedTest
	case skippedTests:
		return test.State == skippedTest
	case windowTests:
		return test.recording >= 0
	}
	return true
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

// windowCount is how many tests made a window.
func (results *results) windowCount() int {
	count := 0
	for _, test := range results.tests {
		if test.recording >= 0 {
			count++
		}
	}
	return count
}

// counts is how many tests passed, failed and were skipped.
func (results *results) counts() (passed, failed, skipped int) {
	for _, test := range results.tests {
		switch test.State {
		case failedTest:
			failed++
		case skippedTest:
			skipped++
		default:
			passed++
		}
	}
	return passed, failed, skipped
}
