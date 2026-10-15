//go:build !js

package window

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// Preferences are small values a program keeps between runs (the theme chosen, the last screen,
// a draft): Preference reads one by its name, KeepPreference keeps one. On a phone they are in the
// application's own folder, which Android removes with the application; on the desktop, in the
// user's configuration folder, under the program's name; in a browser, in the page's local
// storage. They are written at once, so a program the system ends keeps what it kept.

var preferences = preferenceFile{path: preferencesPath}

// Preference is the value kept under a name, and whether there is one.
func Preference(name string) (string, bool) { return preferences.read(name) }

// KeepPreference keeps a value under a name, replacing the one there was.
func KeepPreference(name, value string) error { return preferences.keep(name, &value) }

// ForgetPreference lets go of the value kept under a name.
func ForgetPreference(name string) error { return preferences.keep(name, nil) }

// preferenceFile keeps the values in one JSON file, read the first time one is asked for.
type preferenceFile struct {
	path   func() (string, error)
	lock   sync.Mutex
	values map[string]string
}

func (file *preferenceFile) read(name string) (string, bool) {
	file.lock.Lock()
	defer file.lock.Unlock()
	if err := file.load(); err != nil {
		return "", false
	}
	value, found := file.values[name]
	return value, found
}

// keep sets a value (nil lets it go) and writes the file again.
func (file *preferenceFile) keep(name string, value *string) error {
	file.lock.Lock()
	defer file.lock.Unlock()
	if err := file.load(); err != nil {
		return err
	}
	if value == nil {
		delete(file.values, name)
	} else {
		file.values[name] = *value
	}
	return file.save()
}

func (file *preferenceFile) load() error {
	if file.values != nil {
		return nil
	}
	path, err := file.path()
	if err != nil {
		return err
	}
	file.values = map[string]string{}
	contents, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("window: reading the preferences: %w", err)
	}
	if err := json.Unmarshal(contents, &file.values); err != nil {
		file.values = map[string]string{} // a damaged file starts again rather than stopping the program
	}
	return nil
}

// save writes a new file beside the old one and puts it in its place, so the system ending the
// program halfway leaves the old values whole.
func (file *preferenceFile) save() error {
	path, err := file.path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("window: keeping the preferences: %w", err)
	}
	contents, err := json.MarshalIndent(file.values, "", "  ")
	if err != nil {
		return err
	}
	written := path + ".new"
	if err := os.WriteFile(written, contents, 0o600); err != nil {
		return fmt.Errorf("window: keeping the preferences: %w", err)
	}
	return os.Rename(written, path)
}
