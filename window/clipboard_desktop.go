//go:build !js && !android && !nintendo3ds

package window

import (
	"bytes"
	"context"
	"encoding/hex"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// A desktop window reads only text from the clipboard: GLFW has no call for pictures. So the
// picture is read with the tool each computer already has, as OpenLink opens links: AppleScript
// on a Mac, wl-paste or xclip on Linux. A tool takes a moment to start, which is why
// ReadClipboardPicture answers on a goroutine of its own instead of in the frame.

// clipboardPatience is how long a tool may take to print the clipboard's picture before it is
// given up on: a screenshot takes a fraction of a second, a tool that hangs takes forever.
const clipboardPatience = 5 * time.Second

// pngSignature is how every PNG file begins.
var pngSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

// ReadClipboardPicture reads the picture on the system's clipboard, a screenshot most often, and
// calls read with it as PNG, on a goroutine of its own; read is not called when the clipboard
// holds no picture. It tells whether this system can read one at all: a Mac through AppleScript,
// Linux through wl-paste (Wayland) or xclip (X11), if installed; not Windows yet. In a browser the
// page's paste event brings pictures instead.
// TODO(akrck02): read pictures on Windows, and on Android through its ClipboardManager.
func ReadClipboardPicture(read func(png []byte)) bool {
	reader, found := pictureReaderFor(runtime.GOOS, os.Getenv)
	if !found {
		return false
	}
	go func() {
		if picture, found := reader.read(); found {
			read(picture)
		}
	}()
	return true
}

// pictureReader is how a computer reads its clipboard's picture: the tool that prints it, and how
// to turn what it prints into the PNG's bytes.
type pictureReader struct {
	command []string                            // the tool and its arguments
	decode  func(printed []byte) ([]byte, bool) // nil when the tool prints the PNG as it is
}

// pictureReaderFor is the reader of a system (runtime.GOOS's name), looking up its environment
// to tell a Wayland desktop from an X11 one. It tells false where there is none.
func pictureReaderFor(system string, environment func(name string) string) (pictureReader, bool) {
	switch {
	case system == "windows":
		return pictureReader{}, false
	case system == "darwin":
		return pictureReader{command: []string{"osascript", "-e", "the clipboard as «class PNGf»"}, decode: pictureFromAppleScript}, true
	case environment("WAYLAND_DISPLAY") != "":
		return pictureReader{command: []string{"wl-paste", "--no-newline", "--type", "image/png"}}, true
	}
	return pictureReader{command: []string{"xclip", "-selection", "clipboard", "-target", "image/png", "-out"}}, true
}

// read runs the tool and gives back the PNG it printed; false when it printed none.
func (reader pictureReader) read() ([]byte, bool) {
	waiting, stop := context.WithTimeout(context.Background(), clipboardPatience)
	defer stop()
	printed, err := exec.CommandContext(waiting, reader.command[0], reader.command[1:]...).Output()
	if err != nil {
		// The usual answer of every tool when the clipboard holds no picture, or holds text; a
		// tool that is not installed means the same to the person: there is no picture to paste.
		return nil, false
	}
	picture, found := printed, true
	if reader.decode != nil {
		picture, found = reader.decode(printed)
	}
	if !found || !bytes.HasPrefix(picture, pngSignature) {
		return nil, false
	}
	return picture, true
}

// pictureFromAppleScript is the PNG in what AppleScript prints for a picture: «data PNGf», its
// bytes in hexadecimal, and a closing ».
func pictureFromAppleScript(printed []byte) ([]byte, bool) {
	start := []byte("«data PNGf")
	written := bytes.TrimSpace(printed)
	if !bytes.HasPrefix(written, start) {
		return nil, false
	}
	written = bytes.TrimSuffix(bytes.TrimPrefix(written, start), []byte("»"))
	picture := make([]byte, hex.DecodedLen(len(written)))
	if _, err := hex.Decode(picture, written); err != nil {
		return nil, false // not hexadecimal: not a picture AppleScript knows how to print
	}
	return picture, true
}
