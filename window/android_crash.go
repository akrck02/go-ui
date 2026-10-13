//go:build android

package window

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync/atomic"
)

// lastCrash is the report of how the program ended last time, if it crashed.
var lastCrash string

// keepCrashReports makes the Go runtime write the report of a crash (an unrecovered panic, a fatal
// error or signal) to a file in the application's own folder, and reads the one the last run left.
// A phone shows a person nothing when an application closes: the program can show it next time.
func keepCrashReports() {
	path := filepath.Join(os.TempDir(), "last-crash.txt") // x/mobile points TMPDIR at the app's cache
	if previous, err := os.ReadFile(path); err == nil {
		lastCrash = strings.TrimSpace(string(previous))
	}
	file, err := os.Create(path) // empty: a run that ends well leaves nothing
	if err != nil {
		return
	}
	// built as a shared library, the program starts with no traceback: ask for the failing
	// goroutine's, which says where it broke
	debug.SetTraceback("single")
	_ = debug.SetCrashOutput(file, debug.CrashOptions{})
	_ = file.Close() // SetCrashOutput keeps its own copy of the file
}

// graphics describes the phone's OpenGL ES (version, GPU) once the window has drawn, or what went
// wrong making its renderers.
var graphics atomic.Value // string

// Graphics describes the phone's OpenGL ES and GPU, or the error that stopped the window drawing,
// for a program that shows it to help find a fault on a phone nobody can reach.
func Graphics() string {
	text, _ := graphics.Load().(string)
	return text
}

// LastCrash returns the report of the last run's crash, or "" when it ended well (or this is the
// first run). On Android it is kept in the application's folder; elsewhere it is always "".
func LastCrash() string { return lastCrash }
