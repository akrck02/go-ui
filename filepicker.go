package ui

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/akrck02/go-renderer/input"
)

// A file picker is a sheet to choose a file or a folder on the disk: the places to start from on
// the left (the home folder, the disk's root, and those the program names), what the folder holds
// on the right, and the folder's path above, to type one straight in. A click on a folder goes
// into it; a click on a file chooses it, and a second click, or Choose, ends with it. It is drawn
// by go-ui like any other sheet, so it looks the same on every system and needs no program of
// the system's.

// FilePickerOptions say what a file picker chooses and where it starts.
type FilePickerOptions struct {
	// Title is the sheet's title, which also tells pickers apart; empty is "Choose a file" or
	// "Choose a folder".
	Title string
	// Folder makes it choose a folder (the one open, or the one clicked) instead of a file.
	Folder bool
	// Start is the file or folder chosen before: the picker opens in it, or in the folder it is in.
	// Empty, or nowhere, opens the home folder.
	Start string
	// Suffixes, when there are any, are the endings of the files shown (".glb", ".png"); the
	// folders are always shown, to go through them.
	Suffixes []string
	// Places are folders offered on the left besides the home folder and the root: where the
	// programs the person looks for are usually installed.
	Places []string
	// ReadFolder reads what a folder holds; empty reads the disk. Tests give a folder of their own.
	ReadFolder func(folder string) ([]FileEntry, error)
}

// FileEntry is a file or a folder in a folder.
type FileEntry struct {
	Name   string
	Folder bool
}

// The file picker's measures, in pixels.
const (
	filePickerWidth       = 680
	filePickerListHeight  = 320
	filePickerPlacesWidth = 170
)

// filePickerMemory is what a file picker keeps while it is shown.
type filePickerMemory struct {
	lastFrame   int    // the last frame it was shown, to tell when it opens again
	folder      string // the folder open
	typed       string // the path in the field above the list
	read        string // the folder entries holds, read once and not on every frame
	entries     []FileEntry
	failure     error  // why the folder open could not be read
	chosen      string // the name of the entry clicked in the folder open
	chosenAt    time.Time
	hiddenShown bool
}

// FilePicker shows a sheet to choose a file or a folder. On the step the person chooses one it
// returns its path, and closed; on the step they close it without choosing (Cancel, Escape, the
// backdrop) it returns closed with no path. The program shows it with an if, as a Sheet.
func (ui *Context) FilePicker(options FilePickerOptions) (path string, closed bool) {
	title := options.Title
	if title == "" {
		title = "Choose a file"
		if options.Folder {
			title = "Choose a folder"
		}
	}
	memory := remember(ui, childID(ui.idFor(title), "##picker"), func() *filePickerMemory { return &filePickerMemory{lastFrame: -1} })
	if memory.lastFrame != ui.frame-1 {
		memory.openAt(options)
	}
	memory.lastFrame = ui.frame
	memory.readIfMoved(options)
	dismissed := ui.Sheet(title, SheetOptions{Width: Fixed(filePickerWidth)}, func() {
		path, closed = ui.filePickerContent(memory, options)
	})
	return path, closed || dismissed
}

// filePickerContent is the sheet's content: the path, the places and the folder, and the buttons.
func (ui *Context) filePickerContent(memory *filePickerMemory, options FilePickerOptions) (path string, closed bool) {
	ui.filePickerPath(memory)
	ui.Box(Layout{Width: Grow(), Height: Fixed(filePickerListHeight), Gap: 8}, func() {
		ui.filePickerPlaces(memory, options)
		ui.ScrollArea("Folder", Layout{Direction: Vertical, Width: Grow(), Height: Grow()}, func() {
			path = ui.filePickerEntries(memory, options)
		})
	})
	if memory.failure != nil {
		ui.Text(memory.failure.Error(), TextStyle{Color: ui.Theme.Danger, Size: ui.Theme.SmallTextSize})
	}
	if path != "" {
		return path, true
	}
	return ui.filePickerButtons(memory, options)
}

// filePickerPath is the row above the list: up to the folder that holds this one, and the path,
// to type one in and go there.
func (ui *Context) filePickerPath(memory *filePickerMemory) {
	ui.Box(Layout{Width: Grow(), Gap: 6, AlignY: Center}, func() {
		if ui.IconButton("arrow-up", "Up") {
			memory.goTo(filepath.Dir(memory.folder))
		}
		ui.Tooltip("The folder that holds this one")
		ui.Box(Layout{Width: Grow()}, func() {
			ui.TextField("Path", &memory.typed)
		})
		if ui.IconButton("arrow-right", "Go") {
			memory.goTo(filepath.Clean(memory.typed))
		}
		ui.Tooltip("Open the folder typed")
	})
}

// filePickerPlaces are the folders to start from.
func (ui *Context) filePickerPlaces(memory *filePickerMemory, options FilePickerOptions) {
	ui.Box(Layout{Direction: Vertical, Width: Fixed(filePickerPlacesWidth), Height: Grow()}, func() {
		ui.Push("Places")
		defer ui.Pop()
		for _, place := range placesFor(options) {
			if ui.Selectable(place.icon, place.name, memory.folder == place.folder) {
				memory.goTo(place.folder)
			}
		}
	})
}

// filePickerEntries are what the folder open holds, folders first; it returns the path of a file
// clicked twice.
func (ui *Context) filePickerEntries(memory *filePickerMemory, options FilePickerOptions) string {
	ui.Push(memory.folder)
	defer ui.Pop()
	shown := 0
	for _, entry := range memory.entries {
		if !memory.shows(entry, options) {
			continue
		}
		shown++
		icon := "file"
		if entry.Folder {
			icon = "folder"
		}
		if !ui.Selectable(icon, entry.Name, memory.chosen == entry.Name) {
			continue
		}
		if path, done := memory.click(entry, options, ui.now); done {
			return path
		}
	}
	if shown == 0 && memory.failure == nil {
		ui.Text("Nothing to choose here", TextStyle{Color: ui.Theme.InkSoft})
	}
	return ""
}

// filePickerButtons are the hidden files' switch, Cancel and Choose.
func (ui *Context) filePickerButtons(memory *filePickerMemory, options FilePickerOptions) (path string, closed bool) {
	ui.Box(Layout{Width: Grow(), Gap: 8, AlignY: Center}, func() {
		ui.Checkbox("Hidden files", &memory.hiddenShown)
		ui.Box(Layout{Width: Grow()}, nil)
		if ui.Button("Cancel") {
			closed = true
		}
		if ui.PrimaryButton("Choose") {
			path, closed = memory.choice(options)
		}
	})
	return path, closed
}

// openAt opens the picker where the options start it.
func (memory *filePickerMemory) openAt(options FilePickerOptions) {
	*memory = filePickerMemory{lastFrame: memory.lastFrame}
	folder, file := startOf(options)
	memory.goTo(folder)
	memory.chosen = file
}

// startOf is the folder a picker opens in, and the name chosen in it: the start's, or the home
// folder.
func startOf(options FilePickerOptions) (folder, name string) {
	if options.Start == "" {
		home, _ := os.UserHomeDir() // with no home the root is the start: filepath.Dir("") is "."
		return home, ""
	}
	start := filepath.Clean(options.Start)
	if _, err := readerOf(options)(start); err == nil {
		return start, "" // it reads as a folder
	}
	return filepath.Dir(start), filepath.Base(start)
}

// readerOf is how a picker reads a folder: the options', or the disk.
func readerOf(options FilePickerOptions) func(folder string) ([]FileEntry, error) {
	if options.ReadFolder != nil {
		return options.ReadFolder
	}
	return readDiskFolder
}

// goTo opens a folder; what it holds is read on the next step.
func (memory *filePickerMemory) goTo(folder string) {
	memory.folder, memory.typed, memory.chosen = folder, folder, ""
}

// readIfMoved reads the folder open, when it is not the one read.
func (memory *filePickerMemory) readIfMoved(options FilePickerOptions) {
	if memory.read == memory.folder {
		return
	}
	memory.read = memory.folder
	memory.entries, memory.failure = readerOf(options)(memory.folder)
	slices.SortFunc(memory.entries, compareEntries)
}

// compareEntries puts folders first, then the names as a person reads them, ignoring case.
func compareEntries(first, second FileEntry) int {
	if first.Folder != second.Folder {
		if first.Folder {
			return -1
		}
		return 1
	}
	return strings.Compare(strings.ToLower(first.Name), strings.ToLower(second.Name))
}

// shows tells whether an entry is shown: hidden ones (a dot first) only when asked, and in a
// folder picker only folders, in a file picker only the files of the suffixes.
func (memory *filePickerMemory) shows(entry FileEntry, options FilePickerOptions) bool {
	if strings.HasPrefix(entry.Name, ".") && !memory.hiddenShown {
		return false
	}
	if entry.Folder {
		return true
	}
	if options.Folder {
		return false
	}
	if len(options.Suffixes) == 0 {
		return true
	}
	return slices.ContainsFunc(options.Suffixes, func(suffix string) bool {
		return strings.HasSuffix(strings.ToLower(entry.Name), strings.ToLower(suffix))
	})
}

// click answers a click on an entry: a folder opens, a file is chosen, and a file clicked again
// soon after ends the picker with it.
func (memory *filePickerMemory) click(entry FileEntry, options FilePickerOptions, now time.Time) (string, bool) {
	if entry.Folder {
		memory.goTo(filepath.Join(memory.folder, entry.Name))
		return "", false
	}
	again := memory.chosen == entry.Name && now.Sub(memory.chosenAt) <= input.MultiClickInterval
	memory.chosen, memory.chosenAt = entry.Name, now
	if again {
		return filepath.Join(memory.folder, entry.Name), true
	}
	return "", false
}

// choice is what Choose chooses: the file clicked, or for a folder picker the folder open. A file
// picker with no file clicked chooses nothing and stays open.
func (memory *filePickerMemory) choice(options FilePickerOptions) (string, bool) {
	if options.Folder {
		return memory.folder, true
	}
	if memory.chosen == "" {
		return "", false
	}
	return filepath.Join(memory.folder, memory.chosen), true
}

// filePlace is a folder offered on the left.
type filePlace struct{ icon, name, folder string }

// placesFor are the home folder, the root of its disk, and the program's places.
func placesFor(options FilePickerOptions) []filePlace {
	home, _ := os.UserHomeDir() // with no home there is no home place
	var places []filePlace
	if home != "" {
		places = append(places, filePlace{"house", "Home", home})
	}
	root := filepath.VolumeName(home) + string(filepath.Separator)
	places = append(places, filePlace{"hard-drive", root, root})
	for _, folder := range options.Places {
		places = append(places, filePlace{"folder", filepath.Base(folder), folder})
	}
	return places
}

// readDiskFolder reads a folder of the disk; a link to a folder is a folder.
func readDiskFolder(folder string) ([]FileEntry, error) {
	found, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}
	entries := make([]FileEntry, 0, len(found))
	for _, entry := range found {
		isFolder := entry.IsDir()
		if entry.Type()&fs.ModeSymlink != 0 {
			details, statErr := os.Stat(filepath.Join(folder, entry.Name()))
			isFolder = statErr == nil && details.IsDir() // a broken link is a file that opens nothing
		}
		entries = append(entries, FileEntry{Name: entry.Name(), Folder: isFolder})
	}
	return entries, nil
}
