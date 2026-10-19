//go:build !js && !android && !nintendo3ds

package window

import (
	"os"
	"path/filepath"
	"strings"
)

// preferencesPath is in the user's configuration folder, under the program's name.
func preferencesPath() (string, error) {
	folder, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	program := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
	return filepath.Join(folder, program, "preferences.json"), nil
}
