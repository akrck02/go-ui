//go:build js && wasm

package window

import "syscall/js"

// In a browser the preferences are in the page's local storage, under a prefix of their own.
const preferencePrefix = "go-ui."

// Preference is the value kept under a name, and whether there is one.
func Preference(name string) (string, bool) {
	value := js.Global().Get("localStorage").Call("getItem", preferencePrefix+name)
	if value.IsNull() {
		return "", false
	}
	return value.String(), true
}

// KeepPreference keeps a value under a name, replacing the one there was.
func KeepPreference(name, value string) error {
	js.Global().Get("localStorage").Call("setItem", preferencePrefix+name, value)
	return nil
}

// ForgetPreference lets go of the value kept under a name.
func ForgetPreference(name string) error {
	js.Global().Get("localStorage").Call("removeItem", preferencePrefix+name)
	return nil
}
