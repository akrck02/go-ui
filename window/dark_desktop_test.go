//go:build !js && !android && !nintendo3ds

package window

import "testing"

func TestEachDesktopSaysItIsDarkInItsOwnWords(test *testing.T) {
	cases := []struct {
		system, output string
		dark           bool
	}{
		{"darwin", "Dark\n", true},
		{"darwin", "", false}, // macOS has no style set when it is light
		{"linux", "'prefer-dark'\n", true},
		{"linux", "'default'\n", false},
		{"linux", "", false}, // no gsettings
		{"windows", "    AppsUseLightTheme    REG_DWORD    0x0\n", true},
		{"windows", "    AppsUseLightTheme    REG_DWORD    0x1\n", false},
	}
	for _, each := range cases {
		if dark := darkAnswer(each.system, each.output); dark != each.dark {
			test.Errorf("%s answering %q: dark is %v, not %v", each.system, each.output, dark, each.dark)
		}
	}
}
