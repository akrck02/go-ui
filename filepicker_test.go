package ui

import (
	"errors"
	"testing"
)

// islandFolders are the folders of a made-up disk, for pickers to read.
var islandFolders = map[string][]FileEntry{
	"/island":            {{Name: "notes.txt"}, {Name: "lighthouse", Folder: true}, {Name: "map.glb"}, {Name: ".cache", Folder: true}},
	"/island/lighthouse": {{Name: "lamp.glb"}, {Name: "keeper", Folder: true}},
}

func readIsland(folder string) ([]FileEntry, error) {
	entries, found := islandFolders[folder]
	if !found {
		return nil, errors.New("no such folder")
	}
	return append([]FileEntry(nil), entries...), nil
}

// pickerScene is a file picker shown until it closes, and what it ended with.
type pickerScene struct {
	options FilePickerOptions
	open    bool
	chosen  string
	closes  int
}

func (scene *pickerScene) build(ui *Context) {
	if !scene.open {
		return
	}
	if path, closed := ui.FilePicker(scene.options); closed {
		scene.chosen, scene.open = path, false
		scene.closes++
	}
}

// clickNode clicks in the middle of what the screen reader reads with a label.
func (h *harness) clickNode(label string, build func(ui *Context)) {
	h.t.Helper()
	area := h.node(label).Area
	h.clickAt(area.X+area.Width/2, area.Y+area.Height/2, build)
}

// hasNode tells whether the screen reader reads something with a label.
func (h *harness) hasNode(label string) bool {
	for _, node := range h.ui.ScreenReaderNodes() {
		if node.Label == label {
			return true
		}
	}
	return false
}

func TestAFilePickerGoesIntoAFolderAndEndsWithAFileClickedTwice(t *testing.T) {
	h := readerHarness(t)
	scene := &pickerScene{open: true, options: FilePickerOptions{Title: "Choose a model", Start: "/island", Suffixes: []string{".glb"}, ReadFolder: readIsland}}
	h.frame(scene.build)
	h.frame(scene.build)
	if !h.hasNode("map.glb") || h.hasNode("notes.txt") || h.hasNode(".cache") {
		t.Fatalf("the picker shows the folders and the .glb files, not the others nor the hidden: %v", h.labels())
	}
	if h.node("lighthouse").Area.Y > h.node("map.glb").Area.Y {
		t.Fatal("folders come before files")
	}
	h.clickNode("lighthouse", scene.build)
	h.frame(scene.build)
	h.clickNode("lamp.glb", scene.build)
	if scene.closes != 0 {
		t.Fatal("one click on a file chooses it without ending the picker")
	}
	h.clickNode("lamp.glb", scene.build)
	if scene.closes != 1 || scene.chosen != "/island/lighthouse/lamp.glb" {
		t.Fatalf("a file clicked twice ends the picker with it: %q, %d", scene.chosen, scene.closes)
	}
}

func TestChooseEndsWithTheFileClickedAndCancelWithNothing(t *testing.T) {
	h := readerHarness(t)
	scene := &pickerScene{open: true, options: FilePickerOptions{Start: "/island/map.glb", ReadFolder: readIsland}}
	h.frame(scene.build)
	h.frame(scene.build)
	h.clickNode("Choose", scene.build)
	if scene.chosen != "/island/map.glb" {
		t.Fatalf("Choose ends with the file the picker started at: %q", scene.chosen)
	}
	scene.open = true
	h.frame(scene.build)
	h.frame(scene.build)
	h.clickNode("Cancel", scene.build)
	if scene.closes != 2 || scene.chosen != "" {
		t.Fatalf("Cancel ends with nothing chosen: %q", scene.chosen)
	}
}

func TestAFolderPickerChoosesTheFolderOpenAndShowsNoFiles(t *testing.T) {
	h := readerHarness(t)
	scene := &pickerScene{open: true, options: FilePickerOptions{Folder: true, Start: "/island", ReadFolder: readIsland}}
	h.frame(scene.build)
	h.frame(scene.build)
	if h.hasNode("map.glb") || !h.hasNode("lighthouse") {
		t.Fatalf("a folder picker shows only folders: %v", h.labels())
	}
	h.clickNode("lighthouse", scene.build)
	h.frame(scene.build)
	h.clickNode("Up", scene.build)
	h.frame(scene.build)
	h.clickNode("lighthouse", scene.build)
	h.frame(scene.build)
	h.clickNode("Choose", scene.build)
	if scene.chosen != "/island/lighthouse" {
		t.Fatalf("Choose ends with the folder open: %q", scene.chosen)
	}
}

func TestAFolderThatCannotBeReadSaysWhy(t *testing.T) {
	h := readerHarness(t)
	scene := &pickerScene{open: true, options: FilePickerOptions{Start: "/nowhere/", ReadFolder: readIsland}}
	h.frame(scene.build)
	h.frame(scene.build)
	if !h.hasNode("no such folder") {
		t.Fatalf("a folder that cannot be read says why: %v", h.labels())
	}
}
