package window

import ui "github.com/akrck02/go-ui"

// readerNumbers gives the interface's nodes the numbers Android's screen reader knows them by: a
// virtual view's id is a positive 32-bit number, an interface id 64 bits. The number stays while
// the node does, so the screen reader stays on it from one frame to the next; the interface ids of
// the last frame are kept to turn the screen reader's numbers back.
type readerNumbers struct {
	ids map[int32]ui.ID
}

// number folds an interface id into a positive number, never 0 (the window) and never another
// node's in the same frame: a clash, rare, moves on to the next free number.
func (numbers *readerNumbers) number(id ui.ID) int32 {
	folded := int32((uint64(id) ^ uint64(id)>>32) & 0x7fffffff)
	for {
		if folded == 0 {
			folded = 1
		}
		known, taken := numbers.ids[folded]
		if !taken || known == id {
			numbers.ids[folded] = id
			return folded
		}
		folded = (folded + 1) & 0x7fffffff
	}
}

// start forgets the last frame's numbers, before a frame's nodes are numbered.
func (numbers *readerNumbers) start() {
	if numbers.ids == nil {
		numbers.ids = map[int32]ui.ID{}
	}
	clear(numbers.ids)
}

// id turns a number back into the interface's id; 0 when no node of the last frame had it.
func (numbers *readerNumbers) id(number int32) ui.ID { return numbers.ids[number] }

// readerClass is the Android widget a node reads as, from which a screen reader says its kind.
func readerClass(role ui.Role) string {
	switch role {
	case ui.ButtonRole, ui.TabRole:
		return "android.widget.Button"
	case ui.CheckboxRole:
		return "android.widget.CheckBox"
	case ui.SwitchRole:
		return "android.widget.Switch"
	case ui.RadioRole:
		return "android.widget.RadioButton"
	case ui.SliderRole:
		return "android.widget.SeekBar"
	case ui.FieldRole:
		return "android.widget.EditText"
	case ui.ImageRole:
		return "android.widget.ImageView"
	case ui.ScrollRole:
		return "android.widget.ScrollView"
	}
	return "android.widget.TextView"
}

// isCheckable tells whether a role is on or off, which a screen reader says.
func isCheckable(role ui.Role) bool {
	return role == ui.CheckboxRole || role == ui.SwitchRole || role == ui.RadioRole
}
