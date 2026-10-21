package uitest

import (
	"errors"
	"testing"
)

func TestEachErrorCarriesItsCode(t *testing.T) {
	constructors := map[Code]func() error{
		InvalidSize: func() error { return invalidSize(0, 300) },
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
