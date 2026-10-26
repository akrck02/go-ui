package main

import (
	"testing"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// sampleRecordings are two tests of a made-up interface: one passed with three frames, one failed.
func sampleRecordings() []uitest.Recording {
	button := draw2d.Rectangle{X: 20, Y: 20, Width: 120, Height: 32}
	frame := func(action string, hasTarget bool) uitest.RecordedFrame {
		return uitest.RecordedFrame{Action: action, Target: button, HasTarget: hasTarget, PointerX: 80, PointerY: 36,
			Commands: []draw2d.Command{{Kind: draw2d.RectangleKind, Area: button, Fill: draw2d.Color{0.2, 0.4, 0.8, 1}}}}
	}
	return []uitest.Recording{
		{Test: "TestALampLights", Width: 400, Height: 300, Frames: []uitest.RecordedFrame{
			frame("", false), frame(`Click "Light the lamp"`, true), frame(`Click "Light the lamp"`, true)}},
		{Test: "TestTheBellRings", Failed: true, Width: 400, Height: 300,
			Errors: []uitest.RecordedError{{Code: uitest.ElementNotFound, Message: `no element "Ring the bell"`}},
			Frames: []uitest.RecordedFrame{frame("", false)}},
	}
}

func TestFailedTestsComeFirst(t *testing.T) {
	results := newResults(nil, sampleRecordings())
	recording, found := results.current()
	if !found || recording.Test != "TestTheBellRings" {
		t.Fatalf("the first test shown is %q, want the failed one", recording.Test)
	}
	if passed, failed, _ := results.counts(); passed != 1 || failed != 1 {
		t.Fatalf("counted %d passed and %d failed", passed, failed)
	}
}

func TestSteppingStopsAtTheEnds(t *testing.T) {
	results := newResults(nil, sampleRecordings())
	results.choose(1)
	results.step(-1)
	if results.frame != 0 {
		t.Fatalf("stepping back from the first frame went to %d", results.frame)
	}
	results.step(5)
	if results.frame != 2 {
		t.Fatalf("stepping past the last frame went to %d", results.frame)
	}
}

// openViewer shows the viewer in a test window, on the passed test.
func openViewer(t *testing.T) (*uitest.Window, *results) {
	t.Helper()
	results := newResults(nil, sampleRecordings())
	results.choose(1)
	window, err := uitest.New(results.draw, uitest.Size(1200, 800))
	if err != nil {
		t.Fatal(err)
	}
	window.Frames(2)
	return window, results
}

func TestTheStepButtonsMoveThroughTheFrames(t *testing.T) {
	window, results := openViewer(t)
	if err := window.Click("Next frame"); err != nil {
		t.Fatal(err)
	}
	if results.frame != 1 {
		t.Fatalf("Next went to frame %d, want 1", results.frame)
	}
	if err := window.Click("Previous frame"); err != nil {
		t.Fatal(err)
	}
	if results.frame != 0 {
		t.Fatalf("Previous went to frame %d, want 0", results.frame)
	}
}

func TestChoosingATestShowsItsErrors(t *testing.T) {
	window, _ := openViewer(t)
	if err := window.Click("TestTheBellRings"); err != nil {
		t.Fatal(err)
	}
	if _, err := window.Element(string(uitest.ElementNotFound)); err != nil {
		t.Fatalf("the failed test's error code is not shown: %v", err)
	}
}

func TestTheFoundElementIsHighlighted(t *testing.T) {
	window, results := openViewer(t)
	results.step(1) // the click, which found the button
	window.Frame()
	ring := window.Context().Theme.FocusRing
	for _, command := range window.DrawList().Commands {
		if command.Kind == draw2d.RectangleKind && command.Border == ring && command.BorderWidth >= highlightWidth {
			return
		}
	}
	t.Fatal("no outline in the focus colour around the element the click found")
}

func TestAFrameWithNoTargetHasNoHighlight(t *testing.T) {
	window, _ := openViewer(t)
	theme := ui.LightTheme()
	for _, command := range window.DrawList().Commands {
		if command.Kind == draw2d.RectangleKind && command.Border == theme.FocusRing && command.BorderWidth == highlightWidth {
			t.Fatal("a frame with no action has a highlight")
		}
	}
}

func TestPlayMovesThroughTheFramesAndStopsAtTheLast(t *testing.T) {
	window, results := openViewer(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	results.playback.now = func() time.Time { return now }
	if err := window.Click("Play"); err != nil {
		t.Fatal(err)
	}
	if !results.playback.playing {
		t.Fatal("Play does not play the replay")
	}
	now = now.Add(playbackFrame)
	window.Frame()
	if results.frame != 1 {
		t.Fatalf("after a frame's time the replay shows frame %d, want 1", results.frame)
	}
	recording, _ := results.current()
	now = now.Add(time.Duration(len(recording.Frames)) * playbackFrame)
	window.Frame()
	if results.frame != len(recording.Frames)-1 || results.playback.playing {
		t.Fatalf("the replay is at frame %d, playing %v; want the last frame, stopped", results.frame, results.playback.playing)
	}
}

func TestPauseKeepsTheFrameShown(t *testing.T) {
	window, results := openViewer(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	results.playback.now = func() time.Time { return now }
	if err := window.Click("Play"); err != nil {
		t.Fatal(err)
	}
	if err := window.Click("Pause"); err != nil {
		t.Fatal(err)
	}
	now = now.Add(10 * playbackFrame)
	window.Frame()
	if results.frame != 0 || results.playback.playing {
		t.Fatalf("a paused replay moved to frame %d", results.frame)
	}
}

func TestAtTheHostsSpeedEachDrawnFrameShowsTheNextRecordedOne(t *testing.T) {
	window, results := openViewer(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	results.playback.now = func() time.Time { return now } // the clock stays still: frames drawn pace it
	if err := window.Click("At the host's speed"); err != nil {
		t.Fatal(err)
	}
	if err := window.Click("Play"); err != nil {
		t.Fatal(err)
	}
	start := results.frame
	window.Frame()
	if results.frame != start+1 {
		t.Fatalf("a drawn frame moved the replay from %d to %d, want one frame on", start, results.frame)
	}
	recording, _ := results.current()
	for range recording.Frames {
		window.Frame()
	}
	if results.frame != len(recording.Frames)-1 || results.playback.playing {
		t.Fatalf("the replay is at frame %d, playing %v; want the last frame, stopped", results.frame, results.playback.playing)
	}
}
