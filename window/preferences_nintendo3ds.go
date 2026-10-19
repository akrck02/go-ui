//go:build nintendo3ds

package window

import "errors"

// preferencesPath: TinyGo's os cannot reach the 3DS's SD card yet, so preferences last only the
// run (reading answers none, keeping fails, and programs go on without them).
func preferencesPath() (string, error) {
	return "", errors.New("window: no folder for preferences on the 3DS yet")
}
