package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/akrck02/go-ui/uitest"
)

// someEvents are go test -json's events for a made-up package: a test that passed writing a line,
// one that failed, and one skipped.
const someEvents = `{"Action":"start","Package":"example/lamp"}
{"Action":"run","Package":"example/lamp","Test":"TestTheLampLights"}
{"Action":"output","Package":"example/lamp","Test":"TestTheLampLights","Output":"=== RUN   TestTheLampLights\n"}
{"Action":"output","Package":"example/lamp","Test":"TestTheLampLights","Output":"    lamp_test.go:9: lit at dusk\n"}
{"Action":"pass","Package":"example/lamp","Test":"TestTheLampLights","Elapsed":0.25}
{"Action":"run","Package":"example/lamp","Test":"TestTheBellRings"}
{"Action":"output","Package":"example/lamp","Test":"TestTheBellRings","Output":"    bell_test.go:12: no bell\n"}
{"Action":"fail","Package":"example/lamp","Test":"TestTheBellRings","Elapsed":0.5}
{"Action":"skip","Package":"example/lamp","Test":"TestTheFogHorn","Elapsed":0}
{"Action":"fail","Package":"example/lamp","Elapsed":0.75}
`

func TestEveryTestGoTestRanIsReadWithItsOutput(t *testing.T) {
	var echo strings.Builder
	tests, err := readTestEvents(strings.NewReader(someEvents), &echo)
	if err != nil {
		t.Fatal(err)
	}
	if len(tests) != 3 {
		t.Fatalf("read %d tests, want 3", len(tests))
	}
	lamp, bell, fog := tests[0], tests[1], tests[2]
	if lamp.State != passedTest || lamp.Seconds != 0.25 || !slices.Equal(lamp.Output, []string{"    lamp_test.go:9: lit at dusk"}) {
		t.Fatalf("the passed test reads as %+v", lamp)
	}
	if bell.State != failedTest || fog.State != skippedTest {
		t.Fatalf("the failed and skipped tests read as %v and %v", bell.State, fog.State)
	}
	if !strings.Contains(echo.String(), "no bell") {
		t.Fatal("the output is not shown as go test shows it")
	}
}

func TestATestWithAWindowGetsItsRecording(t *testing.T) {
	tests, _ := readTestEvents(strings.NewReader(someEvents), &strings.Builder{})
	recordings := []uitest.Recording{{Test: "TestTheBellRings", Failed: true}}
	results := newResults(tests, recordings)
	if recording, found := results.current(); !found || recording.Test != "TestTheBellRings" {
		t.Fatal("the failed test, shown first, does not show its window")
	}
	window, err := uitest.New(results.draw, uitest.Size(1200, 800))
	if err != nil {
		t.Fatal(err)
	}
	window.Frames(2)
	if err := window.Click("With a window 1"); err != nil {
		t.Fatal(err)
	}
	if _, err := window.Element("TestTheLampLights"); err == nil {
		t.Fatal("the window tests' chip still lists a test with no window")
	}
	if _, err := window.ElementWithin("Tests", "TestTheBellRings"); err != nil {
		t.Fatalf("the window tests' chip does not list the test with a window: %v", err)
	}
}

func TestATestWithNoWindowShowsWhatItWrote(t *testing.T) {
	tests, _ := readTestEvents(strings.NewReader(someEvents), &strings.Builder{})
	results := newResults(tests, nil)
	window, err := uitest.New(results.draw, uitest.Size(1200, 800))
	if err != nil {
		t.Fatal(err)
	}
	window.Frames(2)
	if _, err := window.Element("    bell_test.go:12: no bell"); err != nil {
		t.Fatalf("the failed test's output is not shown: %v", err)
	}
	if err := window.Click("TestTheLampLights"); err != nil {
		t.Fatal(err)
	}
	if _, err := window.Element("    lamp_test.go:9: lit at dusk"); err != nil {
		t.Fatalf("a passed test's output is not shown: %v", err)
	}
}

func TestPackagesAreNamedWithoutTheFoldersTheyShare(t *testing.T) {
	results := newResults([]testResult{{Package: "example.com/game/engine", Name: "TestA"}, {Package: "example.com/game/editor", Name: "TestB"}}, nil)
	if short := results.shortPackage("example.com/game/engine"); short != "engine" {
		t.Fatalf("the package is named %q", short)
	}
	alone := newResults([]testResult{{Package: "example.com/game/engine", Name: "TestA"}}, nil)
	if short := alone.shortPackage("example.com/game/engine"); short != "engine" {
		t.Fatalf("a package alone is named %q", short)
	}
}
