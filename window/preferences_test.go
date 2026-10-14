//go:build !js

package window

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreferencesAreKeptBetweenRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lighthouses", "preferences.json")
	where := func() (string, error) { return path, nil }
	first := &preferenceFile{path: where}
	if _, found := first.read("theme"); found {
		t.Fatal("a preference before any was kept")
	}
	dark := "dark"
	if err := first.keep("theme", &dark); err != nil {
		t.Fatal(err)
	}
	screen := "lists"
	if err := first.keep("screen", &screen); err != nil {
		t.Fatal(err)
	}
	second := &preferenceFile{path: where} // the next run
	if value, found := second.read("theme"); !found || value != "dark" {
		t.Fatalf("theme %q (%v) on the next run; want dark", value, found)
	}
	if err := second.keep("screen", nil); err != nil {
		t.Fatal(err)
	}
	if _, found := (&preferenceFile{path: where}).read("screen"); found {
		t.Fatal("a preference let go is still there")
	}
	if _, err := os.Stat(path + ".new"); err == nil {
		t.Fatal("the file written halfway was left beside the preferences")
	}
}

func TestADamagedFileStartsAgain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	file := &preferenceFile{path: func() (string, error) { return path, nil }}
	if _, found := file.read("theme"); found {
		t.Fatal("a damaged file gave a preference")
	}
	light := "light"
	if err := file.keep("theme", &light); err != nil {
		t.Fatalf("keeping a preference over a damaged file: %v", err)
	}
}
