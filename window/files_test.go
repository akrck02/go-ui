//go:build !js && !android && !nintendo3ds

package window

import "testing"

// A program offers to attach files only where the system can choose them: where it cannot,
// ChooseFiles must say so at once and never call back, so the program hides its button.
func TestChoosingFilesWhereThereIsNoChooserSaysSo(test *testing.T) {
	if CanChooseFiles() {
		test.Fatal("CanChooseFiles says the desktop has a chooser")
	}
	called := false
	if ChooseFiles(1<<20, func([]ChosenFile) { called = true }) {
		test.Fatal("ChooseFiles says the desktop has a chooser")
	}
	if called {
		test.Fatal("ChooseFiles called back where nothing was chosen")
	}
}
