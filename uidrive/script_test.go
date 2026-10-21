package uidrive_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-ui/uidrive"
)

// keeperScript is a lighthouse keeper's day, with every kind of step.
const keeperScript = `# a lighthouse keeper writes the day's log
click "Add an entry"
type "Fog at dawn, \"thick\""
press Enter
press Shift+Tab
wait 0.5s
until "Saved" 10s
tap "Light the lamp"
hold "Lens" 2s
drag "Boat" 120 -40
drag "Boat" to "Harbour"
scroll "Log" -3
double-click "Entry"
move-to "Help"
click-at 40 300
`

func TestAScriptIsReadAsOneStepALine(test *testing.T) {
	script, err := uidrive.Parse(keeperScript)
	if err != nil {
		test.Fatal(err)
	}
	if len(script) != 14 {
		test.Fatalf("%d steps, want 14", len(script))
	}
	typed, pressed, until := script[1], script[3], script[5]
	if typed.Action != uidrive.Type || typed.Text != `Fog at dawn, "thick"` || typed.Line != 3 {
		test.Fatalf("the second step is %+v", typed)
	}
	if len(pressed.Keys) != 2 || pressed.Keys[0] != input.KeyShift || pressed.Keys[1] != input.KeyTab {
		test.Fatalf("Shift+Tab read as %v", pressed.Keys)
	}
	if until.Label != "Saved" || until.Duration != 10*time.Second {
		test.Fatalf("until read as %+v", until)
	}
}

func TestAScriptWrittenAndReadAgainIsTheSame(test *testing.T) {
	script, err := uidrive.Parse(keeperScript)
	if err != nil {
		test.Fatal(err)
	}
	again, err := uidrive.Parse(script.String())
	if err != nil {
		test.Fatal(err)
	}
	if again.String() != script.String() {
		test.Fatalf("written:\n%s\nread again and written:\n%s", script, again)
	}
	if !strings.Contains(script.String(), "drag \"Boat\" to \"Harbour\"\n") {
		test.Fatalf("written as:\n%s", script)
	}
}

func TestAnUnreadableLineIsAnErrorWithItsLine(test *testing.T) {
	cases := map[string]string{
		"an unknown step":         "click \"Save\"\nsing \"Shanty\"",
		"a label not quoted":      "click \"Save\"\nclick Save",
		"an unknown key":          "click \"Save\"\npress Banana",
		"a duration with no unit": "click \"Save\"\nwait 5",
		"an unclosed quote":       "click \"Save\"\ntype \"Fog",
	}
	for name, text := range cases {
		_, err := uidrive.Parse(text)
		if !errors.Is(err, uidrive.ErrorScriptNotRead) || !strings.Contains(err.Error(), "line 2") {
			test.Errorf("%s gave %v", name, err)
		}
	}
}
