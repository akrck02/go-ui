package uitest

import (
	"fmt"

	"github.com/akrck02/go-ui/uidrive"
)

// Code is the stable name of an error: programs compare it, people search for it. It is
// uidrive's, which finds elements for this package.
type Code = uidrive.Code

// The codes this package returns; the element codes are uidrive's, under the same names.
const (
	InvalidSize         Code = "uitest.invalid-size"
	ElementNotFound          = uidrive.ElementNotFound
	ElementAmbiguous         = uidrive.ElementAmbiguous
	ElementNotVisible        = uidrive.ElementNotVisible
	ElementNotPressable      = uidrive.ElementNotPressable
	RecordingNotSaved   Code = "uitest.recording-not-saved"
	RecordingNotRead    Code = "uitest.recording-not-read"
)

// Error is an error with a code, and a message for people (uidrive.Error).
type Error = uidrive.Error

// The errors to compare with errors.Is; comparing them allocates nothing.
var (
	ErrorInvalidSize         = &Error{Code: InvalidSize}
	ErrorElementNotFound     = uidrive.ErrorElementNotFound
	ErrorElementAmbiguous    = uidrive.ErrorElementAmbiguous
	ErrorElementNotVisible   = uidrive.ErrorElementNotVisible
	ErrorElementNotPressable = uidrive.ErrorElementNotPressable
	ErrorRecordingNotSaved   = &Error{Code: RecordingNotSaved}
	ErrorRecordingNotRead    = &Error{Code: RecordingNotRead}
)

// CodeOf is the code of an error of this package, or "" for any other error (and for nil).
func CodeOf(err error) Code { return uidrive.CodeOf(err) }

// invalidSize is the error of a window asked with no width or no height.
func invalidSize(width, height float32) error {
	return &Error{Code: InvalidSize, Message: fmt.Sprintf("a window of %vx%v layout pixels cannot show anything", width, height)}
}

// recordingNotSaved is the error of a recording that could not be written.
func recordingNotSaved(folder string, cause error) error {
	return &Error{Code: RecordingNotSaved, Message: "the recording could not be written in " + folder + ": " + cause.Error(), Cause: cause}
}

// recordingNotRead is the error of a recording that could not be read.
func recordingNotRead(path string, cause error) error {
	return &Error{Code: RecordingNotRead, Message: path + " could not be read as a recording: " + cause.Error(), Cause: cause}
}
