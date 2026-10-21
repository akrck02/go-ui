package uitest

import "fmt"

// Code is the stable name of an error of this package: programs compare it, people search for it.
type Code string

// The codes this package returns.
const (
	InvalidSize         Code = "uitest.invalid-size"
	ElementNotFound     Code = "uitest.element-not-found"
	ElementAmbiguous    Code = "uitest.element-ambiguous"
	ElementNotVisible   Code = "uitest.element-not-visible"
	ElementNotPressable Code = "uitest.element-not-pressable"
)

// Error is an error of this package: its code, and a message for people.
type Error struct {
	Code    Code
	Message string
}

// The errors to compare with errors.Is; comparing them allocates nothing.
var (
	ErrorInvalidSize         = &Error{Code: InvalidSize}
	ErrorElementNotFound     = &Error{Code: ElementNotFound}
	ErrorElementAmbiguous    = &Error{Code: ElementAmbiguous}
	ErrorElementNotVisible   = &Error{Code: ElementNotVisible}
	ErrorElementNotPressable = &Error{Code: ElementNotPressable}
)

// Error gives the code and the message.
func (err *Error) Error() string {
	if err.Message == "" {
		return string(err.Code)
	}
	return string(err.Code) + ": " + err.Message
}

// Is tells whether target is an error of this package with the same code, so errors.Is compares
// codes, never messages.
func (err *Error) Is(target error) bool {
	other, isError := target.(*Error)
	return isError && other.Code == err.Code
}

// CodeOf is the code of an error of this package, or "" for any other error (and for nil).
func CodeOf(err error) Code {
	packageError, isError := err.(*Error)
	if !isError {
		return ""
	}
	return packageError.Code
}

// invalidSize is the error of a window asked with no width or no height.
func invalidSize(width, height float32) error {
	return &Error{Code: InvalidSize, Message: fmt.Sprintf("a window of %vx%v layout pixels cannot show anything", width, height)}
}
