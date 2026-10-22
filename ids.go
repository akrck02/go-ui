package ui

import (
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
)

// ID names a control from one frame to the next, so the interface can remember that the mouse is
// pressing it or where a panel was left. It is a hash of the control's label and of the scopes it
// is in (see Push), so two "Apply" buttons in different panels do not mix.
type ID uint64

// FNV-1a, 64 bits: short, fast and spreads similar labels well.
const (
	fnvOffset uint64 = 14695981039346656037
	fnvPrime  uint64 = 1099511628211
)

func hashString(seed uint64, text string) ID {
	hash := seed
	for index := 0; index < len(text); index++ {
		hash ^= uint64(text[index])
		hash *= fnvPrime
	}
	if hash == 0 {
		hash = 1 // 0 means "no id"
	}
	return ID(hash)
}

// mix combines an id with a number: the second, third… use of the same label in a frame.
func mix(id ID, number int) ID {
	hash := uint64(id)
	for shift := 0; shift < 64; shift += 8 {
		hash ^= uint64(number>>shift) & 0xff
		hash *= fnvPrime
	}
	return ID(hash)
}

// idFor returns the id of a label in the current scope. A label used again in the same scope and
// frame gets a different id each time, in order, so two identical buttons still work separately;
// give them distinct ids with "##" (see displayText) when they can appear and disappear.
func (ui *Context) idFor(label string) ID {
	id := hashString(ui.scopeSeed(), label)
	repeats := ui.seen[id]
	ui.seen[id] = repeats + 1
	if repeats > 0 {
		id = mix(id, repeats)
	}
	return id
}

func (ui *Context) scopeSeed() uint64 {
	if len(ui.scopes) == 0 {
		return fnvOffset
	}
	return uint64(ui.scopes[len(ui.scopes)-1])
}

// Push opens a scope: ids of the controls added until Pop are made from this name too. Use it
// around controls repeated for each item of a list, with something that names the item.
func (ui *Context) Push(name string) {
	ui.scopes = append(ui.scopes, hashString(ui.scopeSeed(), name))
}

// Pop closes the scope opened last.
func (ui *Context) Pop() {
	if len(ui.scopes) > 0 {
		ui.scopes = ui.scopes[:len(ui.scopes)-1]
	}
}

// displayText is the part of a label that is shown: what comes before "##". The rest only tells
// the id apart, so "Apply##north" and "Apply##south" both read "Apply".
func displayText(label string) string {
	if shown, _, found := strings.Cut(label, "##"); found {
		return shown
	}
	return label
}

// Area tells where a control was drawn in the last frame, found by its label and the scopes
// around it, outermost first: Area("Properties", "Apply") is the Apply button of the Properties
// panel (panels open a scope with their title). Programs use it to place things next to a
// control; tests and demonstrations, to click it.
func (ui *Context) Area(path ...string) (draw2d.Rectangle, bool) {
	area, found := ui.previous[idForPath(path)]
	return area, found
}

// idForPath is the id of a control from its scopes and label, outermost first.
func idForPath(path []string) ID {
	seed := fnvOffset
	var id ID
	for _, part := range path {
		id = hashString(seed, part)
		seed = uint64(id)
	}
	return id
}
