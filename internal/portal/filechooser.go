// Package portal opens the desktop's own dialogs on Linux through its portal, xdg-desktop-portal,
// the service GNOME, KDE and the other desktops answer, asked over the session's message bus
// (D-Bus). It speaks just enough D-Bus for that, so programs need no library and no tool
// installed: https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.FileChooser.html
package portal

import (
	"net/url"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/akrck02/go-renderer/failure"
)

// The portal's names on the bus.
const (
	portalService     = "org.freedesktop.portal.Desktop"
	portalPath        = "/org/freedesktop/portal/desktop"
	fileChooser       = "org.freedesktop.portal.FileChooser"
	requestInterface  = "org.freedesktop.portal.Request"
	requestPathPrefix = "/org/freedesktop/portal/desktop/request/"
)

// responseChosen is the portal's answer when the person chose; 1 is cancelled, 2 anything else.
const responseChosen = 0

// tokens makes each request's token apart, within this program.
var tokens atomic.Uint32

// Available tells whether a session's message bus address, its DBUS_SESSION_BUS_ADDRESS, names a
// bus this package can ask the portal through.
func Available(address string) bool {
	_, _, found := socketOf(address)
	return found
}

// ChooseFolder opens the desktop's folder chooser, with a title and in a start folder ("" for
// the portal's own choice), and waits for the person: it gives back the folder chosen, or ""
// when they close it without one. It blocks until then, so call it on a goroutine of its own.
func ChooseFolder(title, start string) (string, error) {
	opened, err := dial(os.Getenv("DBUS_SESSION_BUS_ADDRESS"), os.Getuid())
	if err != nil {
		return "", err
	}
	defer opened.close()
	match := "type='signal',interface='" + requestInterface + "',member='Response'"
	body := &encoder{}
	body.text(match)
	if _, err := opened.call("org.freedesktop.DBus", "/org/freedesktop/DBus", "org.freedesktop.DBus", "AddMatch", "s", body.data); err != nil {
		return "", err
	}
	token := "goui" + strconv.Itoa(os.Getpid()) + "_" + strconv.Itoa(int(tokens.Add(1)))
	reply, err := opened.call(portalService, portalPath, fileChooser, "OpenFile", "ssa{sv}", openFolderBody(title, start, token))
	if err != nil {
		return "", err
	}
	handle := requestPathPrefix + escapedName(opened.name) + "/" + token
	if len(reply.body) > 0 {
		handle, _ = reply.body[0].(string) // older portals answer with a path of their own
	}
	response, err := opened.nextSignal(handle, "Response")
	if err != nil {
		return "", err
	}
	return chosenFolder(response)
}

// openFolderBody is the body of OpenFile asking for one folder: no parent window, the title, and
// the options (its token, folders only, modal, where to start).
func openFolderBody(title, start, token string) []byte {
	body := &encoder{}
	body.text("") // no parent window: the portal centres the dialog on the screen
	body.text(title)
	body.array(8, func() {
		body.option("handle_token", "s", func() { body.text(token) })
		body.option("directory", "b", func() { body.boolean(true) })
		body.option("modal", "b", func() { body.boolean(true) })
		if start != "" {
			body.option("current_folder", "ay", func() {
				body.array(1, func() { body.data = append(append(body.data, start...), 0) })
			})
		}
	})
	return body.data
}

// chosenFolder reads the folder chosen from the portal's Response signal: its response code and
// its results, whose "uris" holds file:// addresses.
func chosenFolder(response message) (string, error) {
	if len(response.body) < 2 {
		return "", failure.Wrap(BadMessage, "Response", nil)
	}
	if code, _ := response.body[0].(uint32); code != responseChosen {
		return "", nil
	}
	for _, entry := range asList(response.body[1]) {
		pair, _ := entry.([2]any)
		if key, _ := pair[0].(string); key != "uris" {
			continue
		}
		for _, address := range asList(pair[1]) {
			text, _ := address.(string)
			parsed, err := url.Parse(text)
			if err == nil && parsed.Scheme == "file" {
				return parsed.Path, nil
			}
		}
	}
	return "", nil
}
