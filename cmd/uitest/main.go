// Command uitest runs Go tests with recording on and shows them in a Go UI window: every test by
// package, which passed, which failed and what it wrote; for a window test, its error codes and a
// replay of its frames with the pointer, the fingers and the element each action found. It needs a
// display.
//
//	go run github.com/akrck02/go-ui/cmd/uitest ./...
//	go run github.com/akrck02/go-ui/cmd/uitest -dark ./editor/... -run TestSaving
//
// The arguments after the window's flags go to go test as they are.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/akrck02/go-renderer/opengl/overlay"
	"github.com/akrck02/go-ui/uitest"
	"github.com/akrck02/go-ui/window"
)

func main() {
	options := window.Options{Title: "Tests", MinWidth: 900, MinHeight: 560}
	options.RegisterFlags(flag.CommandLine)
	flag.Parse()
	tests, recordings, err := runTests(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	shown := newResults(tests, recordings)
	shown.measurer, err = overlay.NewMeasurer() // the fonts the window paints with
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := window.Run(options, shown.draw); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runTests runs go test -json with the arguments in a fresh recording folder, showing its output
// as go test would, and reads which tests ran and what their windows recorded. Failing tests are
// what the viewer is for, so go test failing is not an error here.
func runTests(arguments []string) ([]testResult, []uitest.Recording, error) {
	folder, err := os.MkdirTemp("", "uitest-")
	if err != nil {
		return nil, nil, err
	}
	// the recordings are read into memory before the window opens; a temporary folder left behind
	// is harmless, so an error removing it is not reported
	defer os.RemoveAll(folder)
	if len(arguments) == 0 {
		arguments = []string{"./..."}
	}
	command := exec.Command("go", append([]string{"test", "-json", "-count=1"}, arguments...)...)
	command.Env = append(os.Environ(), uitest.RecordVariable+"="+folder)
	command.Stderr = os.Stderr
	events, err := command.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := command.Start(); err != nil {
		return nil, nil, err // go itself could not run
	}
	tests, readError := readTestEvents(events, os.Stdout)
	runError := command.Wait()
	if _, testsFailed := runError.(*exec.ExitError); runError != nil && !testsFailed {
		return nil, nil, runError
	}
	if readError != nil {
		return nil, nil, readError
	}
	recordings, err := uitest.LoadRecordings(folder)
	return tests, recordings, err
}
