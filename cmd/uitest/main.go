// Command uitest runs Go tests with recording on and shows the window tests in a Go UI window:
// which passed, which failed and with what error codes, and a replay of each test's frames with
// the pointer, the fingers and the element each action found. It needs a display.
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

	"github.com/akrck02/go-ui/uitest"
	"github.com/akrck02/go-ui/window"
)

func main() {
	options := window.Options{Title: "Window tests", MinWidth: 900, MinHeight: 560}
	options.RegisterFlags(flag.CommandLine)
	flag.Parse()
	recordings, err := runTests(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := window.Run(options, newResults(recordings).draw); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runTests runs go test with the arguments in a fresh recording folder and reads what it recorded.
// Failing tests are what the viewer is for, so go test failing is not an error here.
func runTests(arguments []string) ([]uitest.Recording, error) {
	folder, err := os.MkdirTemp("", "uitest-")
	if err != nil {
		return nil, err
	}
	// the recordings are read into memory before the window opens; a temporary folder left behind
	// is harmless, so an error removing it is not reported
	defer os.RemoveAll(folder)
	if len(arguments) == 0 {
		arguments = []string{"./..."}
	}
	command := exec.Command("go", append([]string{"test", "-count=1"}, arguments...)...)
	command.Env = append(os.Environ(), uitest.RecordVariable+"="+folder)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	runError := command.Run()
	if _, testsFailed := runError.(*exec.ExitError); runError != nil && !testsFailed {
		return nil, runError // go itself could not run
	}
	return uitest.LoadRecordings(folder)
}
