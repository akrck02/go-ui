package portal

import "github.com/akrck02/go-renderer/failure"

// The codes of the errors this package returns. Check them with failure.Is.
var (
	// NoBus is a session with no D-Bus to reach: no address, or one that does not answer.
	NoBus = failure.Define("portal.no-bus", "the session's message bus cannot be reached")
	// Refused is a bus that did not let this program in.
	Refused = failure.Define("portal.refused", "the message bus refused this program")
	// BadMessage is a message from the bus that is not one: cut short, or of a form not known.
	BadMessage = failure.Define("portal.bad-message", "the message bus sent something that is not a message")
	// CallFailed is a call the bus or the portal answered with an error; the details hold its name.
	CallFailed = failure.Define("portal.call-failed", "the desktop portal answered the call with an error")
)
