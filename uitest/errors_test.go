package uitest

import (
	"errors"
	"testing"
)

func TestEachErrorCarriesItsCode(t *testing.T) {
	constructors := map[Code]func() error{
		InvalidSize:      func() error { return invalidSize(0, 300) },
		ElementNotFound:  func() error { return elementNotFound("Save", []string{"Open"}) },
		ElementAmbiguous: func() error { return elementAmbiguous("Save", 2) },
	}
	for code, constructor := range constructors {
		err := constructor()
		if CodeOf(err) != code {
			t.Errorf("constructor for %s returned code %q", code, CodeOf(err))
		}
		if !errors.Is(err, &Error{Code: code}) {
			t.Errorf("errors.Is does not match %s by its code", code)
		}
	}
}
