package window

import (
	"hash/maphash"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

func TestAFrameThatLooksTheSameIsNotShownAgain(t *testing.T) {
	shown := shownFrame{seed: maphash.MakeSeed()}
	var list draw2d.List
	list.Rectangle(draw2d.Rectangle{Width: 10, Height: 10}, 2, draw2d.MustHex("#ffffff"), draw2d.Color{}, 0)
	background := draw2d.MustHex("#f6f8fb")
	if shown.same(&list, background, 800, 600, 2) {
		t.Fatal("the first frame is always shown")
	}
	if !shown.same(&list, background, 800, 600, 2) {
		t.Fatal("the same list at the same size looks the same")
	}
	if shown.same(&list, background, 900, 600, 2) {
		t.Fatal("a new size is shown")
	}
	if shown.same(&list, draw2d.MustHex("#121a23"), 900, 600, 2) {
		t.Fatal("a new background (another theme) is shown")
	}
	list.Text("Faro", 1, 8, 12, draw2d.Regular, draw2d.MustHex("#000000"))
	if shown.same(&list, draw2d.MustHex("#121a23"), 900, 600, 2) {
		t.Fatal("a new command is shown")
	}
}
