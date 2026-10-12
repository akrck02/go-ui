// Command workbench shows the controls an editor around a 3D view needs: a side list of a
// lighthouse's parts in a table with groups, columns that can be resized, an inspector with tags
// as chips, a menu to add a part, a sheet that asks before removing one, and tags that can be
// dragged onto a part or dropped on the island.
//
//	go run ./cmd/workbench
//	go run ./cmd/workbench -capture workbench.png
//	go run ./cmd/workbench -demo menu|sheet|drag -capture demo.png   (opens it on the second step)
//	go run ./cmd/workbench -dark
package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/scene"
	"github.com/akrck02/go-renderer/viewer"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/internal/island"
)

// part is one entry of the list: the program's data, which the interface only shows and changes.
type part struct {
	key, group, name, code string
	count                  int
	tags                   []string
}

// tag is a label that can be put on parts, with its colours.
type tag struct {
	name string
	look ui.ChipLook
}

// workbench is the program's state.
type workbench struct {
	parts         []*part
	selected      int
	sideWidth     float32
	inspectorSize float32
	confirming    bool   // the sheet asking before removing a part is open
	status        string // what happened last, in the status bar
	added         int
}

var tags = []tag{
	{"Painted", ui.ChipLook{Fill: draw2d.Hex("#e3effd"), Ink: draw2d.Hex("#1f5fc4")}},
	{"Checked", ui.ChipLook{Fill: draw2d.Hex("#dff3e8"), Ink: draw2d.Hex("#1e7d4f")}},
	{"Worn", ui.ChipLook{Fill: draw2d.Hex("#fdf0dc"), Ink: draw2d.Hex("#a86400")}},
	{"Spare", ui.ChipLook{}},
}

var columns = []ui.Column{
	{Title: "Part", Width: ui.Share(1)},
	{Title: "Code", Width: ui.Fixed(64)},
	{Title: "Count", Width: ui.Fixed(56), Align: ui.End},
}

func main() {
	options := viewer.DefaultOptions()
	options.RegisterFlags(flag.CommandLine)
	dark := flag.Bool("dark", false, "use the dark theme")
	demo := flag.String("demo", "", "open the menu, the sheet or a drag on the second step: menu, sheet or drag (for captures)")
	flag.Parse()

	interfaceContext, err := ui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *dark {
		interfaceContext.Theme = ui.DarkTheme()
	}
	world := island.Build()
	world.Overlay.Drawing = interfaceContext.Drawing()
	app := newWorkbench()
	steps := 0

	update := func(world *scene.Scene, state *input.State, step float64) {
		steps++
		playDemo(interfaceContext, state, *demo, steps)
		interfaceContext.Begin(state)
		app.draw(interfaceContext)
		interfaceContext.End()
		app.followDrop(interfaceContext)
		island.ShowLamp(world, true)
	}
	if err := viewer.Run(world, options, "Lighthouse workbench", update, 0); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newWorkbench() *workbench {
	app := &workbench{sideWidth: 320, inspectorSize: 280, status: "Drag a tag onto the part, or onto the island"}
	add := func(group, name, code string, count int, tags ...string) {
		app.parts = append(app.parts, &part{key: code, group: group, name: name, code: code, count: count, tags: tags})
	}
	add("Lantern", "Lens", "LN-01", 1, "Checked")
	add("Lantern", "Lamp", "LN-02", 2, "Spare")
	add("Lantern", "Gallery rail", "LN-03", 24, "Painted", "Worn")
	add("Lantern", "Vent cap", "LN-04", 1)
	add("Tower", "Door", "TW-01", 1, "Painted")
	add("Tower", "Window", "TW-02", 8)
	for course := 1; course <= 120; course++ { // a long group, to scroll
		add("Tower", "Stone course "+strconv.Itoa(course), fmt.Sprintf("SC-%03d", course), 40)
	}
	add("Grounds", "Keeper's hut", "GR-01", 1, "Painted")
	add("Grounds", "Boat shed", "GR-02", 1)
	add("Grounds", "Landing", "GR-03", 1, "Worn")
	return app
}

func (app *workbench) current() *part { return app.parts[app.selected] }

// draw describes the whole window: a bar, the side list, the open middle where the island shows,
// the inspector and the status bar. The sheet, when open, goes over all of it.
func (app *workbench) draw(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		app.topBar(context)
		context.Separator()
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
			app.sideList(context)
			context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, nil) // the island shows through
			app.inspector(context)
		})
		context.Separator()
		app.statusBar(context)
	})
	if app.confirming {
		app.confirmRemoval(context)
	}
}

// bar is a full-width strip painted with the surface colour.
func bar(context *ui.Context, height float32, content func()) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(height), Padding: ui.PadXY(8, 0), Gap: 4, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, content)
}

func (app *workbench) topBar(context *ui.Context) {
	bar(context, 40, func() {
		context.IconSized("lamp", 16, context.Theme.Primary)
		context.Box(ui.Layout{Width: ui.Fixed(4)}, nil)
		context.MenuButton("plus", "Add", func() {
			for _, group := range []string{"Lantern", "Tower", "Grounds"} {
				if context.MenuItem("box", "Part in "+group, "") {
					app.addPart(group)
				}
			}
			context.Separator()
			if context.MenuItem("copy", "Copy of the selected part", "Ctrl+D") {
				app.addCopy()
			}
		})
		context.MenuButton("ellipsis", "", func() {
			if context.MenuItem("trash-2", "Remove the selected part", "") {
				app.confirming = true
			}
		})
		context.Spacer()
		context.Caption("Lighthouse workbench")
	})
}

// sideList is the column of parts, as wide as the person drags it.
func (app *workbench) sideList(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Height: ui.Grow()}
	context.Resizable("parts", ui.RightSide, &app.sideWidth, 240, 520, layout, func() {
		inner := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(8), Gap: 8}
		context.Frame(inner, ui.Look{Fill: context.Theme.Surface}, func() {
			context.Caption(fmt.Sprintf("%d parts", len(app.parts)))
			context.Table("parts", ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, columns, func(table *ui.Table) {
				group := ""
				for index, entry := range app.parts {
					if entry.group != group {
						group = entry.group
						table.Group(group)
					}
					if table.Row(entry.key, index == app.selected, entry.name, entry.code, strconv.Itoa(entry.count)) {
						app.selected = index
					}
				}
			})
		})
	})
}

// inspector shows the selected part: its tags as chips, which take tags dropped on them, and the
// tags to drag.
func (app *workbench) inspector(context *ui.Context) {
	entry := app.current()
	layout := ui.Layout{Direction: ui.Vertical, Height: ui.Grow()}
	context.Resizable("inspector", ui.LeftSide, &app.inspectorSize, 220, 420, layout, func() {
		inner := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 10}
		context.Frame(inner, ui.Look{Fill: context.Theme.Surface}, func() {
			context.Heading(entry.name)
			context.Caption(entry.group + " · " + entry.code)
			context.Separator()
			context.Caption("Tags (drop more here)")
			app.tagsOfPart(context, entry)
			context.Caption("Tags to drag")
			context.Push("palette") // the same names as the part's tags, kept apart
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true}, func() {
				for _, available := range tags {
					context.DragSource(available.name, available.name, ui.Layout{}, func() {
						context.Chip("tag", available.name, available.look)
					})
				}
			})
			context.Pop()
			context.Spacer()
			context.Row(func() {
				if context.IconButton("trash-2", "Remove…") {
					app.confirming = true
				}
			})
		})
	})
}

// tagsOfPart is the part's tags, in a box that takes the tags dropped on it.
func (app *workbench) tagsOfPart(context *ui.Context, entry *part) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fit().Between(40, 0), Padding: ui.Pad(6), Gap: 6, Wrap: true, AlignY: ui.Center}
	isTag := func(payload any) bool { _, isName := payload.(string); return isName }
	payload, dropped := context.DropTarget("part tags", layout, isTag, func() {
		for _, name := range entry.tags {
			if _, removed := context.RemovableChip("", name, lookOf(name)); removed {
				entry.tags = slices.DeleteFunc(entry.tags, func(other string) bool { return other == name })
				app.status = "Removed the tag " + name
			}
		}
		if len(entry.tags) == 0 {
			context.Caption("No tags")
		}
	})
	if name, isName := payload.(string); dropped && isName && !slices.Contains(entry.tags, name) {
		entry.tags = append(entry.tags, name)
		app.status = "Tagged " + entry.name + " as " + name
	}
}

func lookOf(name string) ui.ChipLook {
	for _, known := range tags {
		if known.name == name {
			return known.look
		}
	}
	return ui.ChipLook{}
}

// followDrop is the scene's side of a drag, after End: while a tag is over the open island it says
// where it would land, and on the release it takes it there.
func (app *workbench) followDrop(context *ui.Context) {
	if drag, dragging := context.Dragging(); dragging && drag.OverScreen {
		app.status = fmt.Sprintf("Drop %v on the island at %.0f, %.0f", drag.Payload, drag.X, drag.Y)
	}
	if drop, dropped := context.DroppedOnScreen(); dropped {
		app.status = fmt.Sprintf("Dropped %v on the island at %.0f, %.0f", drop.Payload, drop.X, drop.Y)
	}
}

// confirmRemoval is the sheet that asks before a part is removed.
func (app *workbench) confirmRemoval(context *ui.Context) {
	entry := app.current()
	dismissed := context.Sheet("Remove part", ui.SheetOptions{}, func() {
		context.Label("Remove " + entry.name + " from the list?")
		context.Caption("Its tags go with it. This cannot be undone.")
		context.Row(func() {
			context.Spacer()
			if context.Button("Cancel") {
				app.confirming = false
			}
			if context.PrimaryButton("Remove") {
				app.removeSelected()
				app.confirming = false
			}
		})
	})
	if dismissed {
		app.confirming = false
	}
}

func (app *workbench) statusBar(context *ui.Context) {
	bar(context, 26, func() {
		context.Caption(app.status)
		context.Spacer()
		context.Caption(fmt.Sprintf("%d tags on %s", len(app.current().tags), app.current().name))
	})
}

func (app *workbench) addPart(group string) {
	app.added++
	code := fmt.Sprintf("NW-%02d", app.added)
	entry := &part{key: code, group: group, name: "New part " + strconv.Itoa(app.added), code: code, count: 1}
	last := 0 // after the last part of its group, so the groups stay together
	for index, other := range app.parts {
		if other.group == group {
			last = index + 1
		}
	}
	app.parts = slices.Insert(app.parts, last, entry)
	app.selected = last
	app.status = "Added " + entry.name
}

func (app *workbench) addCopy() {
	app.added++
	copied := *app.current()
	copied.key = fmt.Sprintf("%s-%d", copied.code, app.added)
	copied.name += " (copy)"
	copied.tags = slices.Clone(copied.tags)
	app.parts = slices.Insert(app.parts, app.selected+1, &copied)
	app.selected++
	app.status = "Copied " + app.parts[app.selected-1].name
}

func (app *workbench) removeSelected() {
	if len(app.parts) <= 1 {
		return
	}
	app.status = "Removed " + app.current().name
	app.parts = slices.Delete(app.parts, app.selected, app.selected+1)
	app.selected = min(app.selected, len(app.parts)-1)
}

// playDemo acts as a person would on the second and third steps, so a capture shows the menu
// open, the sheet open or a tag being dragged over the island. It needs where things were drawn,
// known after the first step.
func playDemo(context *ui.Context, state *input.State, demo string, steps int) {
	switch {
	case demo == "menu" && steps == 2:
		clickOn(context, state, "Add")
	case demo == "sheet" && steps == 2:
		clickOn(context, state, "Remove…")
	case demo == "drag" && steps == 2:
		if area, found := context.Area("palette", "Checked"); found {
			state.MoveCursor(float64(area.X+area.Width/2), float64(area.Y+area.Height/2))
			state.SetButton(input.MouseLeft, true)
		}
	case demo == "drag" && steps == 3:
		state.MoveCursor(float64(state.WindowWidth)/2, float64(state.WindowHeight)/2)
	}
}

func clickOn(context *ui.Context, state *input.State, path ...string) {
	area, found := context.Area(path...)
	if !found {
		return
	}
	state.MoveCursor(float64(area.X+area.Width/2), float64(area.Y+area.Height/2))
	state.SetButton(input.MouseLeft, true)
	state.SetButton(input.MouseLeft, false)
}
