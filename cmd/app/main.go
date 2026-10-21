// Command app is a traditional desktop application with no 3D at all: a lighthouse register with a
// menu bar, a toolbar, a navigation list, summary cards, a table, an inspector and a status bar,
// drawn by go-ui in a plain window (package window).
//
//	go run ./cmd/app
//	go run ./cmd/app -dark
//	go run ./cmd/app -capture app.png
//	go run ./cmd/app -demo-search Rock     (types in the search box, for captures)
//	go run ./cmd/app -script tour.txt      (plays a script, showing its pointer: a demo)
//	go run ./cmd/app -record session.txt   (writes down what you do, as a script)
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uidrive"
	"github.com/akrck02/go-ui/window"
)

// lighthouse is one entry of the register: the program's data, which the interface only shows.
type lighthouse struct {
	name, place string
	built       int
	height      int // metres
	focalHeight int // metres above the sea
	character   string
	rangeNM     float32
	lit         bool
}

// register is the application's state.
type register struct {
	lighthouses []*lighthouse
	selected    int
	section     string
	filter      string
	search      string
	sort        int
	saved       bool
	steps       int
	demoSearch  string
	shown       []int // the rows the table shows this frame, in order: worked out once per frame
}

var sortings = []string{"Sort by name", "Sort by year built", "Sort by range"}

func main() {
	options := window.Options{Title: "Lighthouse register", MinWidth: 960, MinHeight: 560}
	options.RegisterFlags(flag.CommandLine)
	demoSearch := flag.String("demo-search", "", "type this in the search box on the second step (for captures)")
	scriptPath := flag.String("script", "", "play this script in the window, showing what it does")
	recordPath := flag.String("record", "", "write what you do in the window to this script file when it closes")
	flag.Parse()
	if err := drive(&options, *scriptPath, *recordPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	app := newRegister()
	app.demoSearch = *demoSearch
	if err := window.Run(options, app.draw); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// drive gives the window a driver: a player for a script file, shown and reported on standard
// output, or a recorder that writes its script to a file when the window closes.
func drive(options *window.Options, scriptPath, recordPath string) error {
	if scriptPath != "" {
		text, err := os.ReadFile(scriptPath)
		if err != nil {
			return err
		}
		script, err := uidrive.Parse(string(text))
		if err != nil {
			return err
		}
		options.Driver = uidrive.Play(script, uidrive.PlayOptions{Report: os.Stdout, Show: true})
		return nil
	}
	if recordPath != "" {
		recorder := uidrive.Record()
		options.Driver = recorder
		options.Hidden = func() {
			if err := os.WriteFile(recordPath, []byte(recorder.Script().String()), 0o644); err != nil {
				fmt.Fprintln(os.Stderr, err) // the window is closing: telling is all that is left to do
			}
		}
	}
	return nil
}

func newRegister() *register {
	return &register{
		lighthouses: []*lighthouse{
			{name: "North Point", place: "Rocky island, north shore", built: 1850, height: 30, focalHeight: 40, character: "Fl W 5s", rangeNM: 20, lit: true},
			{name: "South Cape", place: "Headland, south coast", built: 1870, height: 20, focalHeight: 60, character: "Fl(2) W 10s", rangeNM: 18, lit: true},
			{name: "Gull Rock", place: "Reef, outer bay", built: 1900, height: 15, focalHeight: 20, character: "Iso R 4s", rangeNM: 10, lit: false},
			{name: "Harbour Light", place: "Breakwater, harbour mouth", built: 1920, height: 10, focalHeight: 12, character: "Oc G 3s", rangeNM: 6, lit: true},
			{name: "West Ledge", place: "Ledge, western approach", built: 1890, height: 25, focalHeight: 30, character: "Fl(3) W 15s", rangeNM: 16, lit: true},
			{name: "Seal Island", place: "Island, northern channel", built: 1880, height: 18, focalHeight: 35, character: "LFl W 10s", rangeNM: 14, lit: false},
			{name: "East Spit", place: "Sand spit, eastern bay", built: 1910, height: 12, focalHeight: 15, character: "Q W", rangeNM: 8, lit: true},
			{name: "Black Head", place: "Cliff, western headland", built: 1860, height: 22, focalHeight: 70, character: "Fl W 15s", rangeNM: 24, lit: true},
			{name: "Pilot Point", place: "Pier, river mouth", built: 1930, height: 8, focalHeight: 10, character: "Oc R 5s", rangeNM: 5, lit: true},
			{name: "Long Shoal", place: "Shoal, outer channel", built: 1940, height: 16, focalHeight: 18, character: "VQ(3) W 5s", rangeNM: 12, lit: false},
			{name: "Castle Rock", place: "Islet, harbour approach", built: 1875, height: 14, focalHeight: 25, character: "Fl G 3s", rangeNM: 9, lit: true},
		},
		section: "Lighthouses", filter: "All", saved: true,
	}
}

func (app *register) current() *lighthouse { return app.lighthouses[app.selected] }

// draw describes the whole window, top to bottom.
func (app *register) draw(context *ui.Context) {
	app.steps++
	context.KeyboardNavigation = ui.Always // the register is only an interface: Tab walks every control
	app.shown = app.shownIndices(app.shown[:0])
	if app.demoSearch != "" && app.steps == 2 {
		app.search = app.demoSearch
		context.Focus("Search")
	}
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
		app.menuBar(context)
		context.Separator()
		app.toolbar(context)
		context.Separator()
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
			app.sidebar(context)
			context.Separator()
			app.content(context)
			context.Separator()
			app.inspector(context)
		})
		context.Separator()
		app.statusBar(context)
	})
}

// bar is a full-width strip painted with the surface colour.
func bar(context *ui.Context, height, padding float32, content func()) {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(height), Padding: ui.PadXY(padding, 0), Gap: 2, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, content)
}

func gap(context *ui.Context, pixels float32) { context.Box(ui.Layout{Width: ui.Fixed(pixels)}, nil) }

func (app *register) menuBar(context *ui.Context) {
	bar(context, 30, 6, func() {
		context.IconSized("lamp", 16, context.Theme.Primary)
		gap(context, 6)
		for _, menu := range []string{"File", "Edit", "View", "Help"} {
			context.FlatButton(menu)
		}
		context.Spacer()
		context.Caption("Lighthouse register")
		gap(context, 8)
	})
}

func (app *register) toolbar(context *ui.Context) {
	bar(context, 40, 8, func() {
		if context.FlatIconButton("save", "") {
			app.saved = true
		}
		context.Tooltip("Save the register")
		context.FlatIconButton("undo-2", "")
		context.Tooltip("Undo")
		context.FlatIconButton("redo-2", "")
		context.Tooltip("Redo")
		context.Box(ui.Layout{Height: ui.Fixed(22), Padding: ui.PadXY(6, 0)}, func() { context.Separator() })
		for _, filter := range []string{"All", "Lit", "Dark"} {
			if context.Selectable("", filter+"##filter", app.filter == filter) {
				app.filter = filter
			}
		}
		gap(context, 8)
		search := ui.TextFieldOptions{Placeholder: "Search by name or place", Width: ui.Grow().Between(140, 240)}
		context.TextFieldWith("Search", &app.search, search)
		context.Spacer()
		app.themeButton(context)
		context.FlatIconButton("upload", "Export")
		context.Tooltip("Export the register as a file")
		gap(context, 6)
		if context.PrimaryButton("New lighthouse") {
			app.addLighthouse()
		}
	})
}

func (app *register) sidebar(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(220), Height: ui.Grow(), Padding: ui.Pad(8), Gap: 2}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, func() {
		context.Box(ui.Layout{Padding: ui.PadXY(8, 6)}, func() { context.Caption("REGISTER") })
		for _, section := range []struct{ icon, name string }{
			{"lamp", "Lighthouses"}, {"book-open", "Logbook"}, {"clipboard-check", "Inspections"}, {"users", "Keepers"},
		} {
			if context.Selectable(section.icon, section.name, app.section == section.name) {
				app.section = section.name
			}
		}
		context.Spacer()
		context.Separator()
		context.Selectable("settings", "Settings", false)
	})
}

// content is the middle of the window: a title, three summary cards and the table.
func (app *register) content(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(24), Gap: 16}
	context.Box(layout, func() {
		context.Row(func() {
			context.Text(app.section, ui.TextStyle{Size: 22, Weight: draw2d.Bold})
			context.Caption(fmt.Sprintf("%d on the register", len(app.lighthouses)))
			context.Spacer()
			context.Box(ui.Layout{Width: ui.Grow().Between(150, 190)}, func() { context.Select("sort", &app.sort, sortings) })
		})
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 12}, func() {
			lit, totalRange := 0, float32(0)
			for _, entry := range app.lighthouses {
				totalRange += entry.rangeNM
				if entry.lit {
					lit++
				}
			}
			app.summaryCard(context, "sun", "Lit", fmt.Sprint(lit), "showing their light", context.Theme.Positive)
			app.summaryCard(context, "moon", "Dark", fmt.Sprint(len(app.lighthouses)-lit), "switched off", context.Theme.Warning)
			average := totalRange / float32(len(app.lighthouses))
			app.summaryCard(context, "radar", "Average range", fmt.Sprintf("%.1f nm", average), "seen from the sea", context.Theme.Primary)
		})
		app.table(context)
	})
}

func (app *register) summaryCard(context *ui.Context, icon, title, value, hint string, accent draw2d.Color) {
	theme := context.Theme
	look := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: 10}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Share(1), Padding: ui.Pad(16), Gap: 6}, look, func() {
		context.Row(func() {
			context.IconSized(icon, 16, accent)
			context.Text(title, ui.TextStyle{Weight: draw2d.Medium, Color: theme.InkSoft})
		})
		context.Text(value, ui.TextStyle{Size: 26, Weight: draw2d.Bold})
		context.Caption(hint)
	})
}

// columns are the table's column widths: name and place share what is left (Share, not Grow, so
// every row splits it the same way and the columns line up).
var columns = []ui.Sizing{ui.Share(1), ui.Share(1.4), ui.Fixed(64), ui.Fixed(72), ui.Fixed(76)}

func (app *register) table(context *ui.Context) {
	theme := context.Theme
	look := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: 10, Clip: true}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, look, func() {
		header := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(34), Padding: ui.PadXY(16, 0), Gap: 12, AlignY: ui.Center}
		context.Frame(header, ui.Look{Fill: theme.Sunken}, func() {
			for index, title := range []string{"NAME", "PLACE", "BUILT", "RANGE", "STATUS"} {
				cell(context, columns[index], func() { context.Caption(title) })
			}
		})
		context.Separator()
		if len(app.shown) == 0 {
			empty := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(24), Gap: 4, AlignX: ui.Center}
			context.Box(empty, func() {
				context.Label("No lighthouse matches")
				context.Caption("Try another name, or show All")
			})
			return
		}
		// a List builds only the rows in sight, so a register of thousands costs what twenty do
		context.List("rows", ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, len(app.shown), rowHeight, func(row int) {
			index := app.shown[row]
			app.tableRow(context, index, app.lighthouses[index])
		})
	})
}

// rowHeight is the height of a table row.
const rowHeight = 40

// shownIndices lists the lighthouses the filters and the search let through, in the chosen
// order, by their place in the register. It appends to indices, so the slice is reused.
func (app *register) shownIndices(indices []int) []int {
	search := strings.ToLower(strings.TrimSpace(app.search))
	for index, entry := range app.lighthouses {
		if app.shows(entry, search) {
			indices = append(indices, index)
		}
	}
	sort.SliceStable(indices, func(first, second int) bool {
		a, b := app.lighthouses[indices[first]], app.lighthouses[indices[second]]
		switch app.sort {
		case 1:
			return a.built < b.built
		case 2:
			return a.rangeNM > b.rangeNM
		default:
			return a.name < b.name
		}
	})
	return indices
}

// cell is one column of a table row. It clips, so text too long for a narrow column ends in "…".
func cell(context *ui.Context, width ui.Sizing, content func()) {
	context.Frame(ui.Layout{Width: width, AlignY: ui.Center}, ui.Look{Clip: true}, content)
}

func (app *register) tableRow(context *ui.Context, index int, entry *lighthouse) {
	theme := context.Theme
	normal := ui.Look{}
	if index == app.selected {
		normal = ui.Look{Fill: theme.PrimaryTint}
	}
	looks := ui.Looks{Normal: normal, Hovered: ui.Look{Fill: theme.Sunken}, Pressed: ui.Look{Fill: theme.Line}}
	if index == app.selected {
		looks.Hovered, looks.Pressed = normal, normal
	}
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(rowHeight), Padding: ui.PadXY(16, 0), Gap: 12, AlignY: ui.Center}
	if context.Clickable(entry.name+"##row", layout, looks, func() {
		cell(context, columns[0], func() { context.Text(entry.name, ui.TextStyle{Weight: draw2d.Medium}) })
		cell(context, columns[1], func() { context.Text(entry.place, ui.TextStyle{Color: theme.InkSoft}) })
		cell(context, columns[2], func() { context.Label(fmt.Sprint(entry.built)) })
		cell(context, columns[3], func() { context.Label(fmt.Sprintf("%.0f nm", entry.rangeNM)) })
		cell(context, columns[4], func() { statusChip(context, entry.lit) })
	}) {
		app.selected = index
	}
}

// statusChip is a small rounded label: green when lit, amber when dark.
func statusChip(context *ui.Context, lit bool) {
	theme := context.Theme
	text, colour := "Dark", theme.Warning
	if lit {
		text, colour = "Lit", theme.Positive
	}
	look := ui.Look{Fill: colour.WithAlpha(0.12), Radius: 10}
	context.Frame(ui.Layout{Height: ui.Fixed(20), Padding: ui.PadXY(8, 0), Gap: 4, AlignY: ui.Center}, look, func() {
		context.IconSized("circle", 8, colour)
		context.Text(text, ui.TextStyle{Size: 11, Weight: draw2d.Medium, Color: colour})
	})
}

// shows tells whether the filter and the search (already in lower case) let an entry through.
func (app *register) shows(entry *lighthouse, search string) bool {
	if search != "" &&
		!strings.Contains(strings.ToLower(entry.name), search) && !strings.Contains(strings.ToLower(entry.place), search) {
		return false
	}
	switch app.filter {
	case "Lit":
		return entry.lit
	case "Dark":
		return !entry.lit
	default:
		return true
	}
}

func (app *register) inspector(context *ui.Context) {
	entry := app.current()
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(280), Height: ui.Grow(), Padding: ui.Pad(16), Gap: 10}
	context.Frame(layout, ui.Look{Fill: context.Theme.Surface}, func() {
		context.Heading(entry.name)
		context.Caption(entry.place)
		context.Separator()
		context.Push(fmt.Sprint(app.selected)) // each lighthouse's fields keep their own history
		context.Caption("Name")
		if context.TextFieldWith("Name", &entry.name, ui.TextFieldOptions{Placeholder: "Lighthouse name", MaxLength: 40}) {
			app.saved = false
		}
		context.Caption("Place")
		if context.TextFieldWith("Place", &entry.place, ui.TextFieldOptions{Placeholder: "Where it stands"}) {
			app.saved = false
		}
		context.Pop()
		context.Separator()
		property(context, "Built", fmt.Sprint(entry.built))
		property(context, "Tower height", fmt.Sprintf("%d m", entry.height))
		property(context, "Focal height", fmt.Sprintf("%d m", entry.focalHeight))
		property(context, "Character", entry.character)
		context.Separator()
		if context.Toggle("Light shown at night", &entry.lit) {
			app.saved = false
		}
		context.Row(func() {
			context.Icon("radar")
			context.Label("Range")
			context.Spacer()
			context.Label(fmt.Sprintf("%.0f nm", entry.rangeNM))
		})
		if context.SliderStep("range", &entry.rangeNM, 0, 30, 1) {
			app.saved = false
		}
		context.Spacer()
		context.Row(func() {
			if context.IconButton("trash-2", "Remove") {
				app.removeSelected()
			}
			context.Spacer()
			label := "Switch off"
			if !entry.lit {
				label = "Switch on"
			}
			if context.PrimaryButton(label + "##switch") {
				entry.lit = !entry.lit
				app.saved = false
			}
		})
	})
}

// property is one line of the inspector: a name on the left, its value on the right.
func property(context *ui.Context, name, value string) {
	context.Row(func() {
		context.Text(name, ui.TextStyle{Color: context.Theme.InkSoft})
		context.Spacer()
		context.Label(value)
	})
}

func (app *register) statusBar(context *ui.Context) {
	bar(context, 26, 12, func() {
		state, colour := "All changes saved", context.Theme.Positive
		if !app.saved {
			state, colour = "Unsaved changes", context.Theme.Warning
		}
		context.IconSized("circle", 8, colour)
		gap(context, 4)
		context.Caption(state)
		context.Spacer()
		context.Caption(fmt.Sprintf("%s · %d of %d shown", app.section, len(app.shown), len(app.lighthouses)))
	})
}

func (app *register) addLighthouse() {
	number := len(app.lighthouses) + 1
	app.lighthouses = append(app.lighthouses, &lighthouse{
		name: fmt.Sprintf("Lighthouse %d", number), place: "Not placed yet", built: 2000,
		height: 10, focalHeight: 10, character: "Fl W 5s", rangeNM: 10,
	})
	app.selected = len(app.lighthouses) - 1
	app.saved = false
}

func (app *register) removeSelected() {
	if len(app.lighthouses) <= 1 {
		return
	}
	app.lighthouses = append(app.lighthouses[:app.selected], app.lighthouses[app.selected+1:]...)
	app.selected = min(app.selected, len(app.lighthouses)-1)
	app.saved = false
}

// themeButton switches between the light and the dark theme. The theme is a field of the
// context, so changing it here repaints everything from the next frame.
func (app *register) themeButton(context *ui.Context) {
	dark := context.Theme.Background == ui.DarkTheme().Background
	icon, hint, next := "moon", "Dark theme", ui.DarkTheme()
	if dark {
		icon, hint, next = "sun", "Light theme", ui.LightTheme()
	}
	if context.FlatIconButton(icon, "") {
		context.Theme = next
	}
	context.Tooltip(hint)
	gap(context, 4)
}
