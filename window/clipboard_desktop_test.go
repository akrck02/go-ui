//go:build !js && !android && !nintendo3ds

package window

import (
	"bytes"
	"testing"
)

// Each computer reads its clipboard's picture with the tool it has: a Mac with AppleScript, a
// Wayland desktop with wl-paste, an X11 one with xclip. Windows has no reader yet, and says so.
func TestEachDesktopReadsItsClipboardsPictureWithItsOwnTool(test *testing.T) {
	cases := []struct {
		system, wayland, wantProgram string
		wantFound                    bool
	}{
		{"darwin", "", "osascript", true},
		{"linux", "wayland-0", "wl-paste", true},
		{"linux", "", "xclip", true},
		{"freebsd", "", "xclip", true},
		{"windows", "", "", false},
	}
	for _, each := range cases {
		environment := func(name string) string {
			if name == "WAYLAND_DISPLAY" {
				return each.wayland
			}
			return ""
		}
		reader, found := pictureReaderFor(each.system, environment)
		if found != each.wantFound {
			test.Errorf("%s (wayland %q): found a reader: %v", each.system, each.wayland, found)
			continue
		}
		if found && reader.command[0] != each.wantProgram {
			test.Errorf("%s (wayland %q): reads with %q, not %q", each.system, each.wayland, reader.command[0], each.wantProgram)
		}
	}
}

// AppleScript prints the clipboard's PNG as «data PNGf» and the bytes in hexadecimal; the picture
// is those bytes, and anything else is no picture.
func TestAMacsClipboardPictureIsReadFromAppleScriptsHexadecimal(test *testing.T) {
	picture, found := pictureFromAppleScript([]byte("«data PNGf89504E470D0A1A0A»\n"))
	if !found || !bytes.Equal(picture, pngSignature) {
		test.Fatalf("read %x (found: %v), not %x", picture, found, pngSignature)
	}
	for _, output := range []string{"", "«data TIFF4D4D002A»\n", "«data PNGfnot hexadecimal»\n"} {
		if picture, found := pictureFromAppleScript([]byte(output)); found {
			test.Errorf("%q gave a picture: %x", output, picture)
		}
	}
}

// What the tool prints is a picture only when it is a PNG: a tool that fails, or prints text,
// gives none, so a paste of words never attaches a file.
func TestOnlyAPNGTheToolPrintsIsAPicture(test *testing.T) {
	printsPNG := pictureReader{command: []string{"printf", `\211PNG\r\n\032\n`}}
	if picture, found := printsPNG.read(); !found || !bytes.Equal(picture, pngSignature) {
		test.Errorf("a tool printing a PNG gave %x (found: %v)", picture, found)
	}
	for _, reader := range []pictureReader{
		{command: []string{"printf", "north"}},
		{command: []string{"false"}},
		{command: []string{"a-tool-this-computer-does-not-have"}},
	} {
		if picture, found := reader.read(); found {
			test.Errorf("%v gave a picture: %x", reader.command, picture)
		}
	}
}
