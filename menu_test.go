package ui

import (
	"testing"

	"github.com/akrck02/go-renderer/input"
)

// menuScene is a menu button with three items, counting what is chosen.
type menuScene struct{ chosen []string }

func (scene *menuScene) build(ui *Context) {
	ui.Box(Layout{Padding: Pad(20)}, func() {
		ui.MenuButton("plus", "Add", func() {
			for _, item := range []string{"Hut", "Boat", "Pier"} {
				if ui.MenuItem("", item, "") {
					scene.chosen = append(scene.chosen, item)
				}
			}
			ui.Separator()
			if ui.MenuItem("trash", "Clear", "Del") {
				scene.chosen = append(scene.chosen, "Clear")
			}
		})
	})
}

func menuList() ID { return childID(idOf("Add"), "##menu") }

func menuItem(label string) ID { return hashString(uint64(menuList()), label) }

func (h *harness) isShown(id ID) bool {
	_, shown := h.ui.previous[id]
	return shown
}

func TestAMenuOpensUnderItsButtonAndClosesWhenChosen(t *testing.T) {
	h := newHarness(t)
	scene := &menuScene{}
	h.frame(scene.build)
	h.clickInside(idOf("Add"), scene.build)
	h.frame(scene.build)
	if !h.isShown(menuList()) {
		t.Fatal("clicking the button opens its menu")
	}
	button, list := h.rect(idOf("Add")), h.rect(menuList())
	if list.Y < button.Bottom() || list.Width < menuMinimumWidth {
		t.Fatalf("the menu opens under the button, wide enough for its items: %+v", list)
	}
	if !h.ui.WantsKeyboard() {
		t.Fatal("an open menu holds the keyboard, so Escape is its own")
	}
	h.clickInside(menuItem("Boat"), scene.build)
	h.frame(scene.build)
	if len(scene.chosen) != 1 || scene.chosen[0] != "Boat" || h.isShown(menuList()) {
		t.Fatalf("clicking an item chooses it and closes the menu: %v", scene.chosen)
	}
}

func TestAMenuFollowsTheKeyboard(t *testing.T) {
	h := newHarness(t)
	scene := &menuScene{}
	h.frame(scene.build)
	h.clickInside(idOf("Add"), scene.build)
	h.frame(scene.build)
	h.moveTo(700, 500) // the pointer away, so it does not move the highlight
	h.key(input.KeyDown, scene.build)
	h.key(input.KeyDown, scene.build)
	h.key(input.KeyEnter, scene.build)
	h.frame(scene.build)
	if len(scene.chosen) != 1 || scene.chosen[0] != "Boat" || h.isShown(menuList()) {
		t.Fatalf("two steps down and Enter choose the second item: %v", scene.chosen)
	}
}

func TestEscapeOrAClickElsewhereClosesAMenu(t *testing.T) {
	h := newHarness(t)
	scene := &menuScene{}
	h.frame(scene.build)
	h.clickInside(idOf("Add"), scene.build)
	h.key(input.KeyEscape, scene.build)
	h.frame(scene.build)
	if h.isShown(menuList()) || h.ui.WantsKeyboard() {
		t.Fatal("Escape closes the menu and gives the keyboard back")
	}
	h.clickInside(idOf("Add"), scene.build)
	h.clickAt(700, 500, scene.build)
	h.frame(scene.build)
	if h.isShown(menuList()) || len(scene.chosen) != 0 {
		t.Fatal("a click elsewhere closes the menu without choosing")
	}
}

func TestAMenuInASheetTakesEscapeFirst(t *testing.T) {
	h := newHarness(t)
	scene, open := &menuScene{}, true
	build := func(ui *Context) {
		if open && ui.Sheet("Edit", SheetOptions{}, func() { scene.build(ui) }) {
			open = false
		}
	}
	h.frame(build)
	h.frame(build)
	h.clickInside(hashString(uint64(idOf("Edit")), "Add"), build)
	h.key(input.KeyEscape, build)
	if !open {
		t.Fatal("Escape closes the open menu, not the sheet around it")
	}
	h.frame(build)
	h.key(input.KeyEscape, build)
	if open {
		t.Fatal("the next Escape closes the sheet")
	}
}
