package main

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// wideMeasurer measures every character as wide as the text size: twice what FixedMeasurer does.
type wideMeasurer struct{}

func (wideMeasurer) MeasureText(text string, size float32, _ draw2d.Weight, _ float32) draw2d.TextMetrics {
	return draw2d.TextMetrics{Width: float32(len(text)) * size, Ascent: size * 0.8, Descent: size * 0.2}
}

func (wideMeasurer) HasIcon(string) bool { return true }

func TestTheReplayKeepsTheWindowsShape(t *testing.T) {
	recording := uitest.Recording{Width: 400, Height: 300}
	place, fits := fitWindow(recording, draw2d.Rect{X: 0, Y: 0, Width: 1000, Height: 300})
	if !fits {
		t.Fatal("a window did not fit an area bigger than it")
	}
	if place.scale != 1 || place.origin.X != 300 || place.origin.Width != 400 || place.origin.Height != 300 {
		t.Fatalf("placed at %+v, scale %v; want centred at scale 1", place.origin, place.scale)
	}
	big, _ := fitWindow(recording, draw2d.Rect{Width: 1000, Height: 900})
	if big.scale != 1 {
		t.Fatalf("a big area enlarged the window to scale %v; want 1", big.scale)
	}
	small, _ := fitWindow(recording, draw2d.Rect{Width: 200, Height: 300})
	if small.scale != 0.5 || small.origin.Height != 150 {
		t.Fatalf("a narrow area gave scale %v and %+v; want 0.5, the shape kept", small.scale, small.origin)
	}
}

func TestATextIsShrunkToTheWidthItWasMeasuredAt(t *testing.T) {
	frame := uitest.RecordedFrame{Commands: []draw2d.Command{{
		Kind: draw2d.TextKind, Text: "Light", Size: 10, Baseline: 20, Fill: draw2d.Color{0, 0, 0, 1},
		Area: draw2d.Rect{X: 10, Y: 10, Width: 25}, // FixedMeasurer's width: 5 characters of half the size
	}}}
	shown := replay{recording: uitest.Recording{Width: 400, Height: 300}, frame: frame, theme: ui.LightTheme(), measurer: wideMeasurer{}}
	var drawing draw2d.List
	shown.paint(draw2d.Rect{Width: 400, Height: 300}, &drawing)
	for _, command := range drawing.Commands {
		if command.Kind == draw2d.TextKind && command.Text == "Light" {
			if command.Size != 5 {
				t.Fatalf("the text was painted at size %v; want 5, so its real width is the 25 measured", command.Size)
			}
			return
		}
	}
	t.Fatal("the text was not painted")
}

func TestARecordedTextKeepsItsMeasuredWidth(t *testing.T) {
	folder := t.TempDir()
	t.Setenv(uitest.RecordVariable, folder)
	t.Run("label", func(test *testing.T) {
		window, err := uitest.New(func(context *ui.Context) { context.Label("Light") }, uitest.Test(test))
		if err != nil {
			test.Fatal(err)
		}
		window.Frame()
	})
	recordings, err := uitest.LoadRecordings(folder)
	if err != nil || len(recordings) != 1 || len(recordings[0].Frames) == 0 {
		t.Fatalf("recordings %v, %v", recordings, err)
	}
	for _, command := range recordings[0].Frames[0].Commands {
		if command.Kind == draw2d.TextKind && command.Area.Width <= 0 {
			t.Fatalf("the text %q was recorded with no width", command.Text)
		}
	}
}
