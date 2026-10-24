package uitest

import (
	"testing"

	ui "github.com/akrck02/go-ui"
)

// A button that shows only its icon must still say what it does: a screen reader, and a test,
// find it by the name its label hides, not by the icon's name.
func TestAnIconOnlyButtonIsReadByItsHiddenName(t *testing.T) {
	window, err := New(func(context *ui.Context) {
		context.IconButton("trash-2", "##Delete the note")
		context.FlatIconButton("undo-2", "##Undo")
		context.MenuButton("ellipsis", "##More", func() {})
	}, Size(400, 200), Test(t))
	if err != nil {
		t.Fatal(err)
	}
	window.Frames(2)
	for _, name := range []string{"Delete the note", "Undo", "More"} {
		if _, err := window.Element(name); err != nil {
			t.Errorf("no control is read as %q: %v", name, err)
		}
	}
}

// A drag source holds controls of its own, so it has no words to be read by: it is read by its
// name, and a person can find what to drag.
func TestADragSourceIsReadByItsName(t *testing.T) {
	window, err := New(func(context *ui.Context) {
		context.DragSource("Lighthouse card", "lighthouse", ui.Layout{}, func() {
			context.Button("Open")
		})
	}, Size(400, 200), Test(t))
	if err != nil {
		t.Fatal(err)
	}
	window.Frames(2)
	if _, err := window.Element("Lighthouse card"); err != nil {
		t.Errorf("the drag source is not read by its name: %v", err)
	}
}
