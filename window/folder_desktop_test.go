//go:build !js && !android && !nintendo3ds

package window

import (
	"errors"
	"slices"
	"testing"
)

// Each computer opens its own folder chooser: Linux asks the desktop portal through its message
// bus, with kdialog (on KDE) or zenity standing in where the portal does not answer, or alone
// where the session has no bus; a Mac uses AppleScript. Windows has none for now.
func TestEachDesktopOpensItsOwnFolderChooser(test *testing.T) {
	const bus = "unix:path=/run/user/1000/bus"
	cases := []struct {
		name, system, desktop, bus string
		tools                      []string
		wantPortal                 bool
		wantProgram                string
		wantFound                  bool
	}{
		{"a Mac", "darwin", "", "", nil, false, "osascript", true},
		{"GNOME, nothing installed", "linux", "GNOME", bus, nil, true, "", true},
		{"GNOME with zenity", "linux", "GNOME", bus, []string{"zenity", "kdialog"}, true, "zenity", true},
		{"KDE with kdialog", "linux", "KDE", bus, []string{"zenity", "kdialog"}, true, "kdialog", true},
		{"KDE with zenity only", "linux", "KDE", bus, []string{"zenity"}, true, "zenity", true},
		{"no bus, zenity", "linux", "GNOME", "", []string{"zenity"}, false, "zenity", true},
		{"no bus, nothing installed", "linux", "GNOME", "", nil, false, "", false},
		{"Windows", "windows", "", bus, []string{"zenity"}, false, "", false},
	}
	for _, each := range cases {
		environment := func(name string) string {
			switch name {
			case "XDG_CURRENT_DESKTOP":
				return each.desktop
			case "DBUS_SESSION_BUS_ADDRESS":
				return each.bus
			}
			return ""
		}
		installed := func(tool string) (string, error) {
			if slices.Contains(each.tools, tool) {
				return "/usr/bin/" + tool, nil
			}
			return "", errors.New("not installed")
		}
		chooser, found := folderChooserFor(each.system, environment, installed, "Open a folder", "/home/someone")
		program := ""
		if chooser.command != nil {
			program = chooser.command[0]
		}
		if found != each.wantFound || chooser.portal != each.wantPortal || program != each.wantProgram {
			test.Errorf("%s: found %v, portal %v, program %q; want %v, %v, %q", each.name, found, chooser.portal, program,
				each.wantFound, each.wantPortal, each.wantProgram)
		}
	}
}

// zenity opens inside a folder only when it ends with a slash; otherwise it opens its parent.
func TestZenityStartsInsideTheFolderGiven(test *testing.T) {
	installed := func(tool string) (string, error) { return "/usr/bin/" + tool, nil }
	chooser, _ := folderChooserFor("linux", func(string) string { return "" }, func(tool string) (string, error) {
		if tool == "kdialog" {
			return "", errors.New("not installed")
		}
		return installed(tool)
	}, "Open a folder", "/home/someone")
	if last := chooser.command[len(chooser.command)-1]; last != "/home/someone/" {
		test.Errorf("zenity starts in %q, want /home/someone/", last)
	}
}

// The chooser's folder comes without the newline the tools print, nor AppleScript's last slash.
func TestTheFolderChosenComesAsAPlainPath(test *testing.T) {
	if folder := runChooser([]string{"printf", "/home/someone/harbour/\n"}); folder != "/home/someone/harbour" {
		test.Errorf("chosen %q, want /home/someone/harbour", folder)
	}
	if folder := runChooser([]string{"false"}); folder != "" {
		test.Errorf("a chooser closed without a folder gave %q, want nothing", folder)
	}
}
