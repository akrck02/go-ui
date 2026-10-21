package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// The viewer shows every test go test ran, not only the window tests: it runs go test -json and
// reads its events, one JSON object a line, as go doc test2json describes them. A test's output
// lines and its end (pass, fail or skip) come as events of their own.

// testState is how a test ended.
type testState int

const (
	passedTest testState = iota
	failedTest
	skippedTest
)

// testResult is a test go test ran: its package and name, how it ended, how long it took, what it
// wrote, and its window's recording, by its place among the recordings (-1 when it made none).
type testResult struct {
	Package, Name string
	State         testState
	Seconds       float64
	Output        []string
	recording     int
}

// testEvent is one line of go test -json.
type testEvent struct {
	Action  string
	Package string
	Test    string
	Output  string
	Elapsed float64
}

// readTestEvents reads go test -json's events into the tests they tell of, in the order they
// started, and writes their output to echo as go test would have shown it.
func readTestEvents(events io.Reader, echo io.Writer) ([]testResult, error) {
	var tests []testResult
	where := map[string]int{} // a test's place in tests, by its package and name
	lines := bufio.NewScanner(events)
	lines.Buffer(make([]byte, 0, 64*1024), 16*1024*1024) // a test may write long lines
	for lines.Scan() {
		var event testEvent
		if err := json.Unmarshal(lines.Bytes(), &event); err != nil {
			fmt.Fprintln(echo, lines.Text()) // not an event: go's own message, as a build error
			continue
		}
		fmt.Fprint(echo, event.Output)
		if event.Test == "" {
			continue
		}
		key := event.Package + " " + event.Test
		index, found := where[key]
		if !found {
			index = len(tests)
			where[key] = index
			tests = append(tests, testResult{Package: event.Package, Name: event.Test, recording: -1})
		}
		tests[index].apply(event)
	}
	return tests, lines.Err()
}

// apply takes one event of a test into what is known of it.
func (test *testResult) apply(event testEvent) {
	switch event.Action {
	case "output":
		if line := strings.TrimRight(event.Output, "\n"); !isFrameworkLine(line) {
			test.Output = append(test.Output, line)
		}
	case "pass":
		test.State, test.Seconds = passedTest, event.Elapsed
	case "fail":
		test.State, test.Seconds = failedTest, event.Elapsed
	case "skip":
		test.State, test.Seconds = skippedTest, event.Elapsed
	}
}

// isFrameworkLine tells whether an output line is go test's own (=== RUN, --- PASS…), which the
// viewer shows as the test's state instead.
func isFrameworkLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "=== ") || strings.HasPrefix(trimmed, "--- ")
}
