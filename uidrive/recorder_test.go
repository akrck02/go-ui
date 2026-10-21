package uidrive_test

import (
	"strings"
	"testing"

	"github.com/akrck02/go-ui/uidrive"
	"github.com/akrck02/go-ui/uitest"
)

// keepersEvening is what a person does in the keeper's window, through uitest as a backend would
// give it: a note added to the log, the lamp lit by a tap, a long press on the lens, the
// brightness dragged up.
func keepersEvening(test *testing.T, window *uitest.Window) {
	test.Helper()
	slider, err := window.Element("Brightness")
	if err != nil {
		test.Fatal(err)
	}
	actions := []func() error{
		func() error { return window.Click("Note") },
		func() error { window.Type("Fog at dawn"); return nil },
		func() error { return window.Click("Add an entry") },
		func() error { return window.Tap("Light the lamp") },
		func() error { return window.LongPress("Lens") },
		func() error { return window.DragElement("Brightness", slider.Rect.Width/2, 0) },
	}
	for _, action := range actions {
		if err := action(); err != nil {
			test.Fatal(err)
		}
	}
}

func TestARecordedSessionPlayedBackEndsInTheSameState(test *testing.T) {
	recorder := uidrive.Record()
	person := &lighthouse{}
	keepersEvening(test, keeperWindow(test, person, uitest.Driver(recorder)))
	script := recorder.Script()

	replayed := &lighthouse{}
	if err := keeperWindow(test, replayed).Play(script); err != nil {
		test.Fatalf("%v, playing:\n%s", err, script)
	}
	same := strings.Join(replayed.entries, ",") == strings.Join(person.entries, ",") &&
		replayed.lampClicks == person.lampClicks && replayed.lensLongPresses == person.lensLongPresses &&
		replayed.brightness-person.brightness < 1 && person.brightness-replayed.brightness < 1
	if !same || person.lampClicks != 1 || person.brightness == 0 {
		test.Fatalf("the person left %+v, the replay %+v; the script was:\n%s", *person, *replayed, script)
	}
}

func TestARecordingNamesWhatThePersonPressed(test *testing.T) {
	recorder := uidrive.Record()
	keepersEvening(test, keeperWindow(test, &lighthouse{}, uitest.Driver(recorder)))
	written := recorder.Script().String()
	for _, line := range []string{`click "Note"`, `type "Fog at dawn"`, `click "Add an entry"`, `tap "Light the lamp"`, `hold "Lens" 1s`, `drag "Brightness" `} {
		if !strings.Contains(written, line) {
			test.Errorf("the recording has no %s; it is:\n%s", line, written)
		}
	}
}
