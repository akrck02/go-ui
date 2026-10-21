package uidrive

import "github.com/akrck02/go-renderer/input"

// keyNames are the names a script gives keys, as their labels on a keyboard read.
var keyNames = map[string]input.Key{
	"A": input.KeyA, "B": input.KeyB, "C": input.KeyC, "D": input.KeyD, "E": input.KeyE,
	"F": input.KeyF, "G": input.KeyG, "H": input.KeyH, "I": input.KeyI, "J": input.KeyJ,
	"K": input.KeyK, "L": input.KeyL, "M": input.KeyM, "N": input.KeyN, "O": input.KeyO,
	"P": input.KeyP, "Q": input.KeyQ, "R": input.KeyR, "S": input.KeyS, "T": input.KeyT,
	"U": input.KeyU, "V": input.KeyV, "W": input.KeyW, "X": input.KeyX, "Y": input.KeyY,
	"Z": input.KeyZ,
	"0": input.Key0, "1": input.Key1, "2": input.Key2, "3": input.Key3, "4": input.Key4,
	"5": input.Key5, "6": input.Key6, "7": input.Key7, "8": input.Key8, "9": input.Key9,
	"Up": input.KeyUp, "Down": input.KeyDown, "Left": input.KeyLeft, "Right": input.KeyRight,
	"Space": input.KeySpace, "Tab": input.KeyTab, "Enter": input.KeyEnter, "Escape": input.KeyEscape,
	"Shift": input.KeyShift, "Control": input.KeyControl, "Alt": input.KeyAlt, "Super": input.KeySuper,
	"Plus": input.KeyPlus, "Minus": input.KeyMinus, "PageUp": input.KeyPageUp, "PageDown": input.KeyPageDown,
	"LeftBracket": input.KeyLeftBracket, "RightBracket": input.KeyRightBracket,
	"Comma": input.KeyComma, "Period": input.KeyPeriod, "Backspace": input.KeyBackspace,
	"Delete": input.KeyDelete, "Home": input.KeyHome, "End": input.KeyEnd,
}

// KeyName is the name a script gives a key ("Enter", "A"), or "" for a key it has no name for.
func KeyName(key input.Key) string {
	for name, named := range keyNames {
		if named == key {
			return name
		}
	}
	return ""
}

// modifierKeys are the keys a person holds while pressing another: Shift+Tab, Control+S.
var modifierKeys = []input.Key{input.KeyShift, input.KeyControl, input.KeyAlt, input.KeySuper}
