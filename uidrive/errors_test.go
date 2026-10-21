package uidrive

import (
	"errors"
	"testing"
)

func TestEachErrorCarriesItsCode(test *testing.T) {
	constructors := map[Code]func() error{
		ElementNotFound:     func() error { return elementNotFound("Save", []string{"Open"}) },
		ElementAmbiguous:    func() error { return elementAmbiguous("Save", 2) },
		ElementNotVisible:   func() error { return elementNotVisible("Save") },
		ElementNotPressable: func() error { return elementNotPressable("Save") },
		ScriptNotRead:       func() error { return scriptNotRead(3, "no such step") },
		ScriptStopped:       func() error { return scriptStopped(2) },
	}
	for code, constructor := range constructors {
		err := constructor()
		if CodeOf(err) != code {
			test.Errorf("constructor for %s returned code %q", code, CodeOf(err))
		}
		if !errors.Is(err, &Error{Code: code}) {
			test.Errorf("errors.Is does not match %s by its code", code)
		}
	}
}
