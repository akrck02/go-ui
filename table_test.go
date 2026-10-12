package ui

import (
	"strconv"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

// tableScene is a table of parts in groups of a hundred, with a selection.
type tableScene struct {
	selected string
	clicks   int
	built    int
}

var partColumns = []Column{
	{Title: "Part", Width: Share(2)},
	{Title: "Count", Width: Fixed(80), Align: End},
}

func (scene *tableScene) build(ui *Context) {
	ui.Table("parts", Layout{Width: Fixed(400), Height: Fixed(300)}, partColumns, func(table *Table) {
		for group := range 3 {
			table.Group("Group " + strconv.Itoa(group))
			for index := range 100 {
				key := "part-" + strconv.Itoa(group*100+index)
				if table.Row(key, key == scene.selected, key, strconv.Itoa(index*10)) {
					scene.selected = key
					scene.clicks++
				}
			}
		}
	})
	scene.built = 0
	for index := range ui.elements {
		if ui.elements[index].kind == textElement {
			scene.built++
		}
	}
}

func rowID(key string) ID { return scoped("parts", key) }

func TestATableBuildsOnlyTheRowsInSight(t *testing.T) {
	h := newHarness(t)
	scene := &tableScene{}
	h.frame(scene.build)
	h.frame(scene.build)
	if scene.built > 40 {
		t.Fatalf("three hundred rows build only the dozen in sight: %d texts", scene.built)
	}
	table := h.rect(idOf("parts"))
	bodyTop := table.Y + groupRowHeight + 1 // under the header and its line
	if row := h.rect(rowID("part-5")); row.Y != bodyTop+groupRowHeight+5*28 {
		t.Fatalf("rows land under their group caption: %+v", row)
	}
	if _, built := h.ui.previous[rowID("part-50")]; built {
		t.Fatal("a row out of sight is not built")
	}
	h.moveTo(float64(table.X+100), float64(table.Y+150))
	h.state.AddScroll(-1000) // to the end
	h.frame(scene.build)
	h.frame(scene.build)
	last := h.rect(rowID("part-299"))
	if last.Bottom() != table.Bottom() {
		t.Fatalf("scrolled to the end, the last row sits on the table's bottom: %+v in %+v", last, table)
	}
	if scene.built > 40 {
		t.Fatalf("scrolled, still only the rows in sight: %d texts", scene.built)
	}
}

func TestTableRowsSelectAndAlignNumbers(t *testing.T) {
	h := newHarness(t)
	scene := &tableScene{}
	h.frame(scene.build)
	h.clickInside(rowID("part-3"), scene.build)
	h.frame(scene.build)
	if scene.selected != "part-3" || scene.clicks != 1 {
		t.Fatalf("clicking a row reports it: %q", scene.selected)
	}
	theme := h.ui.Theme
	for index := range h.ui.elements {
		element := &h.ui.elements[index]
		if element.id == rowID("part-3") && element.look.Fill != theme.PrimaryTint {
			t.Fatal("the selected row is tinted")
		}
		if element.kind == boxElement && element.parent >= 0 && h.ui.elements[element.parent].id == rowID("part-3") &&
			element.firstChild < 0 && element.look.Fill != theme.Primary {
			t.Fatal("with a strip of the primary colour on its left")
		}
		if element.kind == textElement && element.text == "30" {
			cell := h.ui.elements[element.parent].rect
			if right := element.rect.Right(); right != cell.Right()-8 {
				t.Fatalf("numbers sit on the right of their column: %v in %+v", right, cell)
			}
		}
	}
}

func TestTableHeaderAndRowsLineUp(t *testing.T) {
	h := newHarness(t)
	scene := &tableScene{}
	h.frame(scene.build)
	h.frame(scene.build)
	headerCount, rowCount := draw2d.Rect{}, draw2d.Rect{}
	for index := range h.ui.elements {
		element := &h.ui.elements[index]
		if element.kind != textElement {
			continue
		}
		switch element.text {
		case "Count":
			headerCount = h.ui.elements[element.parent].rect
		case "0":
			rowCount = h.ui.elements[element.parent].rect
		}
	}
	if headerCount.X != rowCount.X || headerCount.Width != rowCount.Width {
		t.Fatalf("the header's columns are the rows' columns: %+v and %+v", headerCount, rowCount)
	}
}

func TestGroupCaptionsAreShownInCapitals(t *testing.T) {
	h := newHarness(t)
	scene := &tableScene{}
	h.frame(scene.build)
	for index := range h.ui.elements {
		if h.ui.elements[index].text == "GROUP 0" {
			return
		}
	}
	t.Fatal("a group caption is written in capitals")
}
