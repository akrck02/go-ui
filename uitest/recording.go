package uitest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

// RecordVariable is the environment variable that turns recording on: a folder where each test
// made with the Test option writes what its window showed, for the results viewer (cmd/uitest).
const RecordVariable = "UITEST_RECORD"

// Recording is what a test's window showed, frame by frame, and the errors it returned.
type Recording struct {
	Test          string // the test's name, as go test gives it
	Failed        bool
	Width, Height float32 // the window, in layout pixels
	Errors        []RecordedError
	Frames        []RecordedFrame
}

// RecordedError is an error a window returned, and the frame it came after.
type RecordedError struct {
	Code    Code
	Message string
	Frame   int
}

// RecordedFrame is one frame: what was drawn, where the pointer and the fingers were, and the
// action that ran it with the element it found.
type RecordedFrame struct {
	Action    string // `Click "Save"`, empty for a frame run on its own
	Target    draw2d.Rect
	HasTarget bool
	PointerX  float32
	PointerY  float32
	Pressed   bool // the main button or the first finger is down
	Touches   []RecordedTouch
	Commands  []draw2d.Command // pictures are kept by their area only
}

// RecordedTouch is a finger on the screen.
type RecordedTouch struct{ X, Y float32 }

// recorder is what a window records while the mode is on.
type recorder struct {
	recording Recording
	folder    string
	action    string
	target    draw2d.Rect
	hasTarget bool
}

// Test names the window after a test. With RecordVariable set, the window records its frames and
// writes them into that folder when the test ends; with it unset, Test changes nothing.
func Test(test testing.TB) Option {
	return func(chosen *settings) { chosen.test = test }
}

// startRecording makes the window record when a test is named and the mode is on.
func (window *Window) startRecording(test testing.TB, width, height float32) {
	folder := os.Getenv(RecordVariable)
	if test == nil || folder == "" {
		return
	}
	window.recording = &recorder{folder: folder, recording: Recording{Test: test.Name(), Width: width, Height: height}}
	test.Cleanup(func() {
		window.recording.recording.Failed = test.Failed()
		if err := window.recording.save(); err != nil {
			test.Error(err)
		}
	})
}

// doing names the action the next frames belong to, and the element it found; the function it
// returns ends the action. An action inside another (a Click's Press) keeps the outer one's name.
func (window *Window) doing(action string, target draw2d.Rect, hasTarget bool) func() {
	recording := window.recording
	if recording == nil || recording.action != "" {
		return func() {}
	}
	recording.action, recording.target, recording.hasTarget = action, target, hasTarget
	return func() { recording.action, recording.hasTarget = "", false }
}

// recordFrame keeps the frame that just ran.
func (window *Window) recordFrame() {
	recording := window.recording
	if recording == nil {
		return
	}
	commands := append([]draw2d.Command(nil), window.DrawList().Commands...)
	for index := range commands {
		commands[index].Picture = nil
	}
	frame := RecordedFrame{Action: recording.action, Target: recording.target, HasTarget: recording.hasTarget,
		PointerX: float32(window.state.CursorX), PointerY: float32(window.state.CursorY),
		Pressed: window.pointerDown, Commands: commands}
	for _, touch := range window.state.Touches() {
		frame.Touches = append(frame.Touches, RecordedTouch{X: float32(touch.X), Y: float32(touch.Y)})
	}
	recording.recording.Frames = append(recording.recording.Frames, frame)
}

// noted keeps an error the window returns, and gives it back.
func (window *Window) noted(err error) error {
	packageError, isPackageError := err.(*Error)
	if window.recording == nil || !isPackageError {
		return err
	}
	recorded := RecordedError{Code: packageError.Code, Message: packageError.Message, Frame: len(window.recording.recording.Frames)}
	window.recording.recording.Errors = append(window.recording.recording.Errors, recorded)
	return err
}

// save writes the recording as a JSON file of its own in the folder.
func (recording *recorder) save() error {
	if err := os.MkdirAll(recording.folder, 0o755); err != nil {
		return recordingNotSaved(recording.folder, err)
	}
	name := strings.NewReplacer("/", "_", " ", "_").Replace(recording.recording.Test)
	file, err := os.CreateTemp(recording.folder, name+"-*.json")
	if err != nil {
		return recordingNotSaved(recording.folder, err)
	}
	encodeError := json.NewEncoder(file).Encode(recording.recording)
	closeError := file.Close()
	if encodeError != nil {
		return recordingNotSaved(recording.folder, encodeError)
	}
	if closeError != nil {
		return recordingNotSaved(recording.folder, closeError)
	}
	return nil
}

// LoadRecordings reads every recording in a folder, sorted by test name.
func LoadRecordings(folder string) ([]Recording, error) {
	paths, err := filepath.Glob(filepath.Join(folder, "*.json"))
	if err != nil {
		return nil, recordingNotRead(folder, err)
	}
	recordings := make([]Recording, 0, len(paths))
	for _, path := range paths {
		recording, err := loadRecording(path)
		if err != nil {
			return nil, err
		}
		recordings = append(recordings, recording)
	}
	sort.SliceStable(recordings, func(first, second int) bool { return recordings[first].Test < recordings[second].Test })
	return recordings, nil
}

// loadRecording reads one recording.
func loadRecording(path string) (Recording, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Recording{}, recordingNotRead(path, err)
	}
	var recording Recording
	if err := json.Unmarshal(contents, &recording); err != nil {
		return Recording{}, recordingNotRead(path, err)
	}
	return recording, nil
}
