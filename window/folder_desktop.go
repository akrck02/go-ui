//go:build !js && !android && !nintendo3ds

package window

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/akrck02/go-ui/internal/portal"
)

// A desktop window has no folder chooser of its own in GLFW, so it opens the system's: on Linux
// through the desktop portal, which GNOME, KDE and the others answer with their own dialog and
// which needs nothing installed; on a Mac through AppleScript. zenity or kdialog stand in where a
// Linux session has no portal. The person may take as long as they like, so ChooseFolder answers
// on a goroutine of its own, never in the frame.

// ChooseFolder opens the system's own folder chooser, with a title and in a start folder ("" for
// the system's choice), and calls chosen once, on a goroutine of its own, with the folder chosen,
// or with "" when the person closes the chooser without one. It tells whether this computer has a
// chooser at all: a Mac always; Linux when its session has a message bus, or zenity or kdialog;
// not Windows yet.
// TODO(akrck02): Windows' own chooser.
func ChooseFolder(title, start string, chosen func(folder string)) bool {
	chooser, found := folderChooserFor(runtime.GOOS, os.Getenv, exec.LookPath, title, start)
	if !found {
		return false
	}
	go func() { chosen(chooser.run(title, start)) }()
	return true
}

// folderChooser is how a computer opens its folder chooser: through the desktop portal, through a
// command, or the portal first and the command where the portal does not answer.
type folderChooser struct {
	portal  bool
	command []string
}

// folderChooserFor is the chooser of a system (runtime.GOOS's name), looking up its environment
// for a message bus and a KDE desktop, and whether a tool is installed. It tells false where there
// is none.
func folderChooserFor(system string, environment func(name string) string, installed func(tool string) (string, error),
	title, start string) (folderChooser, bool) {
	switch system {
	case "darwin":
		return folderChooser{command: appleScriptChooser(title, start, environment("HOME"))}, true
	case "windows":
		return folderChooser{}, false
	}
	chooser := folderChooser{portal: portal.Available(environment("DBUS_SESSION_BUS_ADDRESS"))}
	_, zenityMissing := installed("zenity")
	_, kdialogMissing := installed("kdialog")
	onKDE := strings.Contains(strings.ToUpper(environment("XDG_CURRENT_DESKTOP")), "KDE")
	switch {
	case kdialogMissing == nil && (onKDE || zenityMissing != nil):
		chooser.command = []string{"kdialog", "--title", title, "--getexistingdirectory", start}
	case zenityMissing == nil:
		chooser.command = []string{"zenity", "--file-selection", "--directory", "--title", title, "--filename", withSlash(start)}
	}
	return chooser, chooser.portal || chooser.command != nil
}

// appleScriptChooser is the command that opens a Mac's folder chooser. The title and the folder
// are given as arguments, never written into the script.
func appleScriptChooser(title, start, home string) []string {
	script := `on run arguments
	set chosen to choose folder with prompt (item 1 of arguments) default location (POSIX file (item 2 of arguments))
	return POSIX path of chosen
end run`
	if start == "" {
		start = home
	}
	return []string{"osascript", "-e", script, title, start}
}

// run opens the chooser and waits for the person: the portal first, then the command, if the
// portal did not answer.
func (chooser folderChooser) run(title, start string) string {
	if chooser.portal {
		folder, err := portal.ChooseFolder(title, start)
		if err == nil || chooser.command == nil {
			return folder // without a command to stand in, a portal that failed chose nothing
		}
	}
	return runChooser(chooser.command)
}

// withSlash ends a folder with a slash, as zenity needs to open inside it rather than in its parent.
func withSlash(folder string) string {
	if folder == "" || strings.HasSuffix(folder, "/") {
		return folder
	}
	return folder + "/"
}

// runChooser runs a chooser's command and gives back the folder it printed, "" when the person
// closed it without one.
func runChooser(command []string) string {
	printed, err := exec.Command(command[0], command[1:]...).Output()
	if err != nil {
		return "" // every chooser ends with an error when it is closed without a folder
	}
	folder := strings.TrimRight(string(printed), "\n")
	if len(folder) > 1 {
		folder = strings.TrimSuffix(folder, "/") // AppleScript ends a folder with a slash
	}
	return folder
}
