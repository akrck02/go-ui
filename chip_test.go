package ui

import (
	"strconv"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

func TestAWrappingRowBreaksIntoLines(t *testing.T) {
	h := newHarness(t)
	widths := []float32{60, 60, 60, 100}
	h.frame(func(ui *Context) {
		ui.frameWithID(idOf("tags"), Layout{Width: Fixed(200), Padding: Pad(10), Gap: 10, Wrap: true}, Look{}, func() {
			for index, width := range widths {
				ui.frameWithID(idOf("tag"+strconv.Itoa(index)), Layout{Width: Fixed(width), Height: Fixed(20)}, Look{}, nil)
			}
		})
	})
	// inner width 180: 60 + 10 + 60 fit, a third 60 would reach 200; the next line takes 60 and 100
	if third := h.rect(idOf("tag2")); third.X != 10 || third.Y != 40 {
		t.Fatalf("the third tag starts the second line: %+v", third)
	}
	if fourth := h.rect(idOf("tag3")); fourth.X != 80 || fourth.Y != 40 {
		t.Fatalf("the fourth follows it on that line: %+v", fourth)
	}
	if box := h.rect(idOf("tags")); box.Height != 10+20+10+20+10 {
		t.Fatalf("the box is as tall as its two lines: %v", box.Height)
	}
}

func TestAWrappingRowFollowsTheWidthItIsGiven(t *testing.T) {
	h := newHarness(t)
	build := func(width float32) func(ui *Context) {
		return func(ui *Context) {
			ui.Box(Layout{Direction: Vertical, Width: Fixed(width)}, func() {
				ui.frameWithID(idOf("tags"), Layout{Width: Grow(), Gap: 4, Wrap: true, AlignY: Center}, Look{}, func() {
					for index := range 6 {
						height := float32(20)
						if index == 1 {
							height = 30
						}
						ui.frameWithID(idOf("tag"+strconv.Itoa(index)), Layout{Width: Fixed(50), Height: Fixed(height)}, Look{}, nil)
					}
				})
			})
		}
	}
	h.frame(build(400))
	if tags := h.rect(idOf("tags")); tags.Height != 30 {
		t.Fatalf("six tags of 50 fit one line of 400: %v", tags.Height)
	}
	if first := h.rect(idOf("tag0")); first.Y != 5 {
		t.Fatalf("a shorter tag is centred in its line: %+v", first)
	}
	h.frame(build(110))
	if tags := h.rect(idOf("tags")); tags.Height != 30+4+20+4+20 {
		t.Fatalf("in 110 they make three lines of two: %v", tags.Height)
	}
}

func TestChipsReportClicksAndRemovals(t *testing.T) {
	h := newHarness(t)
	clicks, removals := 0, 0
	look := ChipLook{Fill: draw2d.MustHex("#dff3e8"), Ink: draw2d.MustHex("#1e7d4f")}
	build := func(ui *Context) {
		ui.Row(func() {
			if ui.Chip("anchor", "Harbour", look) {
				clicks++
			}
			clicked, removed := ui.RemovableChip("", "Beacon", ChipLook{})
			if clicked {
				clicks++
			}
			if removed {
				removals++
			}
		})
	}
	h.frame(build)
	h.clickInside(idOf("Harbour"), build)
	if clicks != 1 || removals != 0 {
		t.Fatalf("clicking a chip reports a click: %d clicks, %d removals", clicks, removals)
	}
	chip := h.rect(idOf("Beacon"))
	h.clickAt(chip.X+4, chip.Y+chip.Height/2, build)
	h.frame(build)
	h.clickInside(childID(idOf("Beacon"), "##remove"), build)
	if clicks != 2 || removals != 1 {
		t.Fatalf("the × reports a removal and not a click: %d clicks, %d removals", clicks, removals)
	}
	if chip.Height != chipHeight {
		t.Fatalf("chips are smaller than buttons: %v", chip.Height)
	}
}

func TestAChipLeftAtZeroTakesTheThemeColours(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) { ui.Chip("", "Tag", ChipLook{}) })
	for _, element := range h.ui.elements {
		if element.id == idOf("Tag") && element.look.Fill != h.ui.Theme.Sunken {
			t.Fatalf("a chip with no colours is filled with Sunken: %v", element.look.Fill)
		}
	}
}

func TestTheLinesOfAWrappingRowCanTouchWhileItsChildrenKeepTheirGap(t *testing.T) {
	h := newHarness(t)
	h.frame(func(ui *Context) {
		ui.frameWithID(idOf("words"), Layout{Width: Fixed(130), Gap: 10, LineGap: NoLineGap, Wrap: true}, Look{}, func() {
			for index := range 3 {
				ui.frameWithID(idOf("word"+strconv.Itoa(index)), Layout{Width: Fixed(60), Height: Fixed(20)}, Look{}, nil)
			}
		})
	})
	if second := h.rect(idOf("word1")); second.X != 70 {
		t.Fatalf("words on a line keep the gap: %+v", second)
	}
	if third := h.rect(idOf("word2")); third.Y != 20 {
		t.Fatalf("the second line starts right under the first: %+v", third)
	}
	if box := h.rect(idOf("words")); box.Height != 40 {
		t.Fatalf("the box is as tall as its two lines and nothing between: %v", box.Height)
	}
}
