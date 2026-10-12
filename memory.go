package ui

// remembered is something a control keeps between frames (a scroll position, whether a select is
// open), with the last frame it was used in so it can be forgotten when the control goes away.
type remembered struct {
	value     any
	lastFrame int
}

// memoryForgottenAfter is how many frames a control may go unseen before its memory is dropped.
const memoryForgottenAfter = 3600

// remember returns what the control with an id keeps, creating it the first time.
func remember[T any](ui *Context, id ID, create func() *T) *T {
	entry, found := ui.memory[id]
	if !found {
		entry = &remembered{value: create()}
		ui.memory[id] = entry
	}
	entry.lastFrame = ui.frame
	value, isType := entry.value.(*T)
	if !isType {
		// the same id used by a control of another kind: start again for this one
		value = create()
		entry.value = value
	}
	return value
}

// forgetUnusedMemory drops what controls unseen for a long time kept.
func (ui *Context) forgetUnusedMemory() {
	for id, entry := range ui.memory {
		if ui.frame-entry.lastFrame > memoryForgottenAfter {
			delete(ui.memory, id)
		}
	}
}

// pressable adds a box that reacts to the mouse and tells what the mouse did. It looks at the
// mouse after its content, so controls inside it get the pointer first; callers change the box's
// look afterwards through the returned index.
func (ui *Context) pressable(id ID, layout Layout, look Look, content func()) (int, interaction) {
	index := ui.frameWithID(id, layout, look, content)
	return index, ui.interact(id)
}

// childID derives the id of a part of a control (a scroll bar's thumb, a select's list).
func childID(parent ID, part string) ID { return hashString(uint64(parent), part) }
