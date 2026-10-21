package uidrive

import (
	"fmt"
	"strconv"
)

// Code is the stable name of an error: programs compare it, people search for it.
type Code string

// The codes this package returns. The element codes were born in uitest and keep its names, since
// tests and the reports of scripts on devices check them.
const (
	ElementNotFound     Code = "uitest.element-not-found"
	ElementAmbiguous    Code = "uitest.element-ambiguous"
	ElementNotVisible   Code = "uitest.element-not-visible"
	ElementNotPressable Code = "uitest.element-not-pressable"
	ScriptNotRead       Code = "uidrive.script-not-read"
	ScriptStopped       Code = "uidrive.script-stopped"
)

// Error is an error with a code, and a message for people.
type Error struct {
	Code    Code
	Message string
	Cause   error // what went wrong underneath, when something outside this package failed
}

// The errors to compare with errors.Is; comparing them allocates nothing.
var (
	ErrorElementNotFound     = &Error{Code: ElementNotFound}
	ErrorElementAmbiguous    = &Error{Code: ElementAmbiguous}
	ErrorElementNotVisible   = &Error{Code: ElementNotVisible}
	ErrorElementNotPressable = &Error{Code: ElementNotPressable}
	ErrorScriptNotRead       = &Error{Code: ScriptNotRead}
	ErrorScriptStopped       = &Error{Code: ScriptStopped}
)

// Error gives the code and the message.
func (err *Error) Error() string {
	if err.Message == "" {
		return string(err.Code)
	}
	return string(err.Code) + ": " + err.Message
}

// Unwrap is the cause underneath, for errors.Is and errors.As.
func (err *Error) Unwrap() error { return err.Cause }

// Is tells whether target is an Error with the same code, so errors.Is compares codes, never
// messages.
func (err *Error) Is(target error) bool {
	other, isError := target.(*Error)
	return isError && other.Code == err.Code
}

// CodeOf is the code of an Error, or "" for any other error (and for nil).
func CodeOf(err error) Code {
	codedError, isError := err.(*Error)
	if !isError {
		return ""
	}
	return codedError.Code
}

// elementNotFound is the error of a label the last frame did not show, with what it did show.
func elementNotFound(label string, shown []string) error {
	return &Error{Code: ElementNotFound, Message: fmt.Sprintf("no element %s in the last frame; it showed %q", describe(label), shown)}
}

// elementAmbiguous is the error of a label shown more than once.
func elementAmbiguous(label string, count int) error {
	return &Error{Code: ElementAmbiguous, Message: fmt.Sprintf("%d elements are called %s; find one by its scope (FindIn)", count, describe(label))}
}

// elementNotVisible is the error of an element laid out where it cannot be reached.
func elementNotVisible(label string) error {
	return &Error{Code: ElementNotVisible, Message: describe(label) + " is scrolled away or behind a dialog"}
}

// elementNotPressable is the error of a click on something that does nothing when pressed.
func elementNotPressable(label string) error {
	return &Error{Code: ElementNotPressable, Message: describe(label) + " is not a control: a click on it does nothing"}
}

// scriptNotRead is the error of a script's line that is not a step.
func scriptNotRead(line int, problem string) error {
	return &Error{Code: ScriptNotRead, Message: "line " + strconv.Itoa(line) + ": " + problem}
}

// scriptStopped is the error of a script a person's own input stopped.
func scriptStopped(step int) error {
	return &Error{Code: ScriptStopped, Message: "a person's input stopped the script at step " + strconv.Itoa(step)}
}

// describe is a short form of a label for messages.
func describe(label string) string { return strconv.Quote(label) }
