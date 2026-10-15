package window

import (
	"testing"

	ui "github.com/akrck02/go-ui"
)

func TestReaderNumbersArePositiveStableAndTurnBack(t *testing.T) {
	var numbers readerNumbers
	numbers.start()
	ids := []ui.ID{1, 0xffffffff_00000000, 0x7fffffff, 0x80000000_80000000, 42}
	seen := map[int32]bool{}
	for _, id := range ids {
		number := numbers.number(id)
		if number <= 0 || seen[number] {
			t.Fatalf("id %x got %d: positive, and not another's", id, number)
		}
		seen[number] = true
		if numbers.id(number) != id {
			t.Fatalf("%d turns back into %x, not %x", number, numbers.id(number), id)
		}
	}
	first := numbers.number(42)
	numbers.start()
	if numbers.number(42) != first {
		t.Fatal("the same id gets the same number in the next frame")
	}
	if numbers.id(12345) != 0 {
		t.Fatal("a number no node has is no id")
	}
}

func TestRolesReadAsAndroidsWidgets(t *testing.T) {
	if readerClass(ui.SwitchRole) != "android.widget.Switch" || readerClass(ui.NoRole) != "android.widget.TextView" {
		t.Fatal("roles map to Android's widgets")
	}
	if !isCheckable(ui.RadioRole) || isCheckable(ui.ButtonRole) {
		t.Fatal("only checkboxes, switches and radio buttons are on or off")
	}
}
