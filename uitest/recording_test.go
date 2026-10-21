package uitest

import (
	"testing"

	ui "github.com/akrck02/go-ui"
)

func TestAWindowWithoutRecordingKeepsNothing(t *testing.T) {
	t.Setenv(RecordVariable, "")
	window, err := New(keeperLog, Test(t))
	if err != nil {
		t.Fatal(err)
	}
	window.Frame()
	if err := window.Click("Light the lamp"); err != nil {
		t.Fatal(err)
	}
	if window.recording != nil {
		t.Fatal("a window recorded with the mode off")
	}
}

func TestARecordedTestWritesItsFramesAndErrors(t *testing.T) {
	folder := t.TempDir()
	t.Setenv(RecordVariable, folder)
	t.Run("keeper", func(test *testing.T) {
		window, err := New(func(context *ui.Context) { keeperLog(context) }, Test(test))
		if err != nil {
			test.Fatal(err)
		}
		window.Frame()
		if err := window.Click("Light the lamp"); err != nil {
			test.Fatal(err)
		}
		_, _ = window.Element("Ring the bell") // the error is what this test records
	})
	recordings, err := LoadRecordings(folder)
	if err != nil {
		t.Fatal(err)
	}
	if len(recordings) != 1 {
		t.Fatalf("%d recordings, want 1", len(recordings))
	}
	recording := recordings[0]
	if recording.Test != "TestARecordedTestWritesItsFramesAndErrors/keeper" || recording.Failed {
		t.Fatalf("recorded %q, failed %v", recording.Test, recording.Failed)
	}
	if len(recording.Errors) != 1 || recording.Errors[0].Code != ElementNotFound {
		t.Fatalf("recorded errors %+v", recording.Errors)
	}
	clicked := false
	for _, frame := range recording.Frames {
		if frame.Action == `Click "Light the lamp"` && frame.HasTarget && len(frame.Commands) > 0 {
			clicked = true
		}
	}
	if !clicked {
		t.Fatalf("no frame of the click with its target among %d frames", len(recording.Frames))
	}
}
