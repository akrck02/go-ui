package uitest

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	ui "github.com/akrck02/go-ui"
)

func TestAFrameRunsTheInterfaceOnceLikeABackend(t *testing.T) {
	calls := 0
	var width, height float32
	window, err := New(func(context *ui.Context) {
		calls++
		width, height = context.Size()
		context.Label("Lighthouse log")
	}, Size(400, 300))
	if err != nil {
		t.Fatal(err)
	}
	window.Frame()
	if calls != 1 {
		t.Fatalf("the interface ran %d times in one frame, want 1", calls)
	}
	window.Frames(3)
	if calls != 4 {
		t.Fatalf("the interface ran %d times after four frames, want 4", calls)
	}
	if width != 400 || height != 300 {
		t.Fatalf("the interface saw a window of %vx%v, want 400x300", width, height)
	}
	if window.DrawList().IsEmpty() {
		t.Fatal("a frame with a label drew nothing")
	}
}

func TestAScaleMakesTheFramebufferBigger(t *testing.T) {
	window, err := New(func(*ui.Context) {}, Size(400, 300), Scale(2))
	if err != nil {
		t.Fatal(err)
	}
	state := window.State()
	if state.FramebufferWidth != 800 || state.FramebufferHeight != 600 {
		t.Fatalf("framebuffer %dx%d, want 800x600", state.FramebufferWidth, state.FramebufferHeight)
	}
}

func TestAThemeIsGivenToTheInterface(t *testing.T) {
	window, err := New(func(*ui.Context) {}, Theme(ui.DarkTheme()))
	if err != nil {
		t.Fatal(err)
	}
	if window.Context().Theme != ui.DarkTheme() {
		t.Fatal("the window did not use the theme it was given")
	}
}

func TestANewWindowRejectsAnEmptySize(t *testing.T) {
	_, err := New(func(*ui.Context) {}, Size(0, 300))
	if !errors.Is(err, ErrorInvalidSize) {
		t.Fatalf("got %v, want %s", err, InvalidSize)
	}
}

func TestWaitMovesTheClockAndRunsTheFramesInIt(t *testing.T) {
	var seen []time.Time
	window, err := New(func(context *ui.Context) {})
	if err != nil {
		t.Fatal(err)
	}
	window.build = func(*ui.Context) { seen = append(seen, window.State().Time()) }
	start := window.Now()
	window.Frames(2)
	if !window.Now().Equal(start) {
		t.Fatal("frames moved the clock; only Wait should")
	}
	seen = seen[:0]
	window.Wait(100 * time.Millisecond)
	if !window.Now().Equal(start.Add(100 * time.Millisecond)) {
		t.Fatalf("the clock is at %v after waiting 100 ms from %v", window.Now(), start)
	}
	if len(seen) != 6 {
		t.Fatalf("waiting 100 ms ran %d frames, want 6 at sixty a second", len(seen))
	}
	if !seen[len(seen)-1].Equal(window.Now()) {
		t.Fatal("the last frame of a wait did not see the time it ends at")
	}
}

func TestTheUIPackageDoesNotImportUITest(t *testing.T) {
	output, err := exec.Command("go", "list", "-deps", "github.com/akrck02/go-ui").Output()
	if err != nil {
		t.Skipf("go list could not run here: %v", err)
	}
	if strings.Contains(string(output), "go-ui/uitest") {
		t.Fatal("package ui depends on uitest")
	}
}
