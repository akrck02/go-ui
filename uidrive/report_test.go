package uidrive_test

import (
	"strings"
	"testing"
	"time"

	"github.com/akrck02/go-ui/uidrive"
	"github.com/akrck02/go-ui/uitest"
)

// maximumFrames bounds a test's script, so a player that never ends fails instead of hanging.
const maximumFrames = 600

// playReported plays a script in a keeper's window through the Driver option, as a real window
// would, and gives what it reported.
func playReported(test *testing.T, text string) string {
	test.Helper()
	script, err := uidrive.Parse(text)
	if err != nil {
		test.Fatal(err)
	}
	var report strings.Builder
	player := uidrive.Play(script, uidrive.PlayOptions{Report: &report})
	window := keeperWindow(test, &lighthouse{}, uitest.Driver(player))
	for frame := 0; !player.Done() && frame < maximumFrames; frame++ {
		window.Wait(time.Second / 60)
	}
	return report.String()
}

func TestTheReportSaysEachStepAndTheCodeOfAFailure(test *testing.T) {
	report := playReported(test, "click \"Light the lamp\"\nclick \"Put out the lamp\"\nclick \"Lens\"")
	lines := strings.Split(strings.TrimSuffix(report, "\n"), "\n")
	want := []string{
		"uidrive: step 1 ok: click \"Light the lamp\"",
		"uidrive: step 2 failed uitest.element-not-found: click \"Put out the lamp\": no element",
		"uidrive: script failed at step 2: uitest.element-not-found",
	}
	if len(lines) != len(want) {
		test.Fatalf("reported:\n%s", report)
	}
	for index, line := range lines {
		if !strings.HasPrefix(line, want[index]) {
			test.Errorf("line %d is %q, want it to start with %q", index+1, line, want[index])
		}
	}
}

func TestTheReportEndsWithAPassWhenEveryStepPassed(test *testing.T) {
	report := playReported(test, "click \"Light the lamp\"\nwait 0.1s")
	if !strings.HasSuffix(report, "uidrive: script passed: 2 steps\n") {
		test.Fatalf("reported:\n%s", report)
	}
}

func TestAnEmptyScriptPassesAtOnce(test *testing.T) {
	if report := playReported(test, "# nothing to do yet\n"); report != "uidrive: script passed: 0 steps\n" {
		test.Fatalf("reported:\n%s", report)
	}
}
