package uidrive

import (
	"fmt"
	"io"
)

// The report is one line a step, then one for the script, each starting with "uidrive:" so a test
// runner finds them among a device's other logs:
//
//	uidrive: step 1 ok: click "Add an entry"
//	uidrive: step 2 failed uitest.element-not-found: click "Remove": no element "Remove" in …
//	uidrive: script failed at step 2: uitest.element-not-found
//	uidrive: script passed: 5 steps
//
// A report is written only when PlayOptions.Report is set, and only when a step ends: never at
// every frame. A writer that fails loses lines, not the script: Player.Err still tells the result.

// reportPassed writes the line of a step that played through.
func reportPassed(report io.Writer, number int, step Step) {
	if report == nil {
		return
	}
	_, _ = fmt.Fprintf(report, "uidrive: step %d ok: %s\n", number, step) // see above: the report is best effort
}

// reportFailed writes the line of a step that failed, then the script's.
func reportFailed(report io.Writer, number int, step Step, err error) {
	if report == nil {
		return
	}
	_, _ = fmt.Fprintf(report, "uidrive: step %d failed %s: %s: %s\n", number, CodeOf(err), step, messageOf(err))
	_, _ = fmt.Fprintf(report, "uidrive: script failed at step %d: %s\n", number, CodeOf(err))
}

// reportScriptPassed writes the script's line when every step passed.
func reportScriptPassed(report io.Writer, steps int) {
	if report == nil {
		return
	}
	_, _ = fmt.Fprintf(report, "uidrive: script passed: %d steps\n", steps)
}

// messageOf is an error's message for people, without its code, which the report gives apart.
func messageOf(err error) string {
	if codedError, isError := err.(*Error); isError {
		return codedError.Message
	}
	return err.Error()
}
