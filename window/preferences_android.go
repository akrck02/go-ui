//go:build android

package window

import (
	"os"
	"path/filepath"
)

// preferencesPath is in the application's own folder (go-renderer points the temporary folder
// there), which Android removes with the application.
func preferencesPath() (string, error) {
	return filepath.Join(os.TempDir(), "preferences.json"), nil
}
