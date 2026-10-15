//go:build android

package window

import (
	"hash/maphash"

	"github.com/akrck02/go-renderer/android"
	"github.com/akrck02/go-renderer/gles"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/models"
	ui "github.com/akrck02/go-ui"
)

// Wake makes the window draw a frame now: call it when the program's data changes on another
// goroutine. Safe from any goroutine.
func Wake() { android.Wake() }

// SystemPrefersDark tells whether the phone is set to its dark theme, so a program can follow it.
func SystemPrefersDark() bool { return android.SystemPrefersDark() }

// SystemPalettes returns the colours the phone made from its wallpaper, and false before Android
// 12, which has none. They change with the wallpaper: read them in every frame.
func SystemPalettes() (Palettes, bool) {
	read, found := android.SystemPalettes()
	return Palettes(read), found
}

// Notify shows a notification in the phone's shade, even while the program is in the background:
// a title and a line, and touching it opens the program. The same id replaces the one shown with
// it before. It returns false when it could not, most often because notifications are not allowed.
func Notify(id int, title, text string) bool { return android.Notify(id, title, text) }

// NotificationsAllowed tells whether the program may show notifications.
func NotificationsAllowed() bool { return android.NotificationsAllowed() }

// AskForNotifications asks the person to allow notifications (Android 13 and later, whose package
// names android.permission.POST_NOTIFICATIONS); ask when they do something that needs them.
func AskForNotifications() { android.AskForNotifications() }

// FeelLongPress gives the short buzz of a long press, as Android's own applications do, if the
// person has touch feedback on. Call it in the frame a control is long-pressed (ui.LongPressed).
func FeelLongPress() { android.FeelLongPress() }

// OpenLink opens a link in the application that shows it: a web page in the browser, mailto: in
// the mail, tel: in the phone, geo: in the maps. It returns false when nothing opens it.
func OpenLink(link string) bool { return android.OpenLink(link) }

// Share offers a text (a link, a note) to the person's other applications, through the system's
// sheet with a title over it. It returns false when it could not.
func Share(title, text string) bool { return android.Share(title, text) }

// Graphics describes the phone's OpenGL ES and GPU, or the error that stopped the window drawing,
// for a program that shows it to help find a fault on a phone nobody can reach.
func Graphics() string { return android.Graphics() }

// LastCrash returns the report of the last run's crash, or "" when it ended well (or this is the
// first run): the Go runtime writes it in the application's own folder.
func LastCrash() string { return android.LastCrash() }

// RunWith is Run with an interface context of your own (a custom theme or measurer). On Android the
// window is the application's activity, drawn through OpenGL ES by go-renderer's android package
// and its gles renderers. It returns only when the program stops; Android ends it.
func RunWith(context *ui.Context, options Options, build func(context *ui.Context)) error {
	shown := shownFrame{seed: maphash.MakeSeed()}
	view := &shownScene{}
	app := interfaceApplication(context, options, build, &shown, view)
	update := app.Update
	reader := &screenReader{}
	app.Update = func(app *models.Application, step float64) error {
		bars := android.CurrentInsets()
		context.SetInsets(ui.Insets{Top: bars.Top, Bottom: bars.Bottom, Left: bars.Left, Right: bars.Right})
		context.SetTextScale(android.TextScale()) // the person's text size, from the phone's settings
		reader.before(context, app.Input)
		takeKeyboardEdit(context)
		err := update(app, step)
		reader.after(context)
		sendTypedField(context)
		android.ShowKeyboard(context.Typing()) // the on-screen keyboard, while a field is typed in
		return err
	}
	var flat *gles.FlatRenderer
	var scene *gles.SceneRenderer
	draw := func(gl android.GL, fresh bool) error {
		if fresh {
			// a new graphics context (the first frame, or back from the background): new renderers
			flat, scene = nil, nil
			shown = shownFrame{seed: shown.seed}
		}
		width, height, scale := frameSize(app)
		if current := view.view; current != nil {
			if scene == nil {
				made, err := gles.NewSceneRenderer(gl)
				if err != nil {
					return err
				}
				scene = made
			}
			scene.OverlayScale = scale
			scene.Draw(current.World, current.Camera, width, height)
			return nil
		}
		if flat == nil {
			made, err := gles.NewFlatRenderer(gl)
			if err != nil {
				return err
			}
			flat = made
		}
		flat.Draw(context.Drawing(), context.Theme.Background, width, height, scale)
		return nil
	}
	return android.Run(app, android.Options{MaxFramesPerSecond: options.MaxFramesPerSecond, Hidden: options.Hidden, Shown: options.Shown}, draw)
}

// screenReader carries what TalkBack asks to the interface, and what the interface shows to
// TalkBack, at each step.
type screenReader struct {
	numbers readerNumbers
	nodes   []android.AccessibleNode // reused from step to step
}

// before tells the interface whether a screen reader is on, which node it is on, and what it asked.
func (reader *screenReader) before(context *ui.Context, state *input.State) {
	context.SetScreenReader(android.ScreenReaderOn())
	actions := android.TakeReaderActions()
	for _, action := range actions {
		id := reader.numbers.id(action.Node)
		switch {
		case id == 0:
		case action.Kind == android.ReaderSetText:
			context.ScreenReaderSetText(id, action.Text)
		default:
			context.ScreenReaderAct(id, readerAction(action.Kind))
		}
	}
	if len(actions) > 0 {
		// as after a touch, one more frame: what the program changed in this one (another screen)
		// shows in the next
		state.RequestFrame(0)
	}
	context.SetScreenReaderFocus(reader.numbers.id(android.ScreenReaderFocus()))
}

// after hands the step's nodes to Android.
func (reader *screenReader) after(context *ui.Context) {
	if !context.ScreenReaderOn() {
		return
	}
	reader.numbers.start()
	reader.nodes = reader.nodes[:0]
	for _, node := range context.ScreenReaderNodes() {
		reader.nodes = append(reader.nodes, reader.androidNode(node))
	}
	android.ShowAccessibleNodes(reader.nodes)
}

func (reader *screenReader) androidNode(node ui.AccessibleNode) android.AccessibleNode {
	var parent int32
	if node.Parent != 0 {
		parent = reader.numbers.number(node.Parent)
	}
	return android.AccessibleNode{
		ID: reader.numbers.number(node.ID), Parent: parent,
		X: node.Area.X, Y: node.Area.Y, Width: node.Area.Width, Height: node.Area.Height,
		Class: readerClass(node.Role), Label: node.Label, Value: node.Value,
		Pressable: node.Pressable, LongPressable: node.LongPressable,
		Checkable: isCheckable(node.Role), Checked: node.Checked, Selected: node.Selected, Disabled: node.Disabled,
		Scrollable: node.Scrollable, Editable: node.Role == ui.FieldRole, Typing: node.Typing, Secret: node.Secret,
		Heading: node.Role == ui.HeadingRole,
		Minimum: node.Range.Minimum, Maximum: node.Range.Maximum, Current: node.Range.Current,
	}
}

// readerAction is what TalkBack asked, as the interface names it.
func readerAction(kind android.ReaderActionKind) ui.ReaderAction {
	switch kind {
	case android.ReaderLongPress:
		return ui.ReaderLongPress
	case android.ReaderForward:
		return ui.ReaderForward
	case android.ReaderBackward:
		return ui.ReaderBackward
	}
	return ui.ReaderPress
}

// takeKeyboardEdit gives the field typed in what the on-screen keyboard made of its text since the
// last step: words, the word being composed, deletions.
func takeKeyboardEdit(context *ui.Context) {
	field, typing := context.TypedField()
	if !typing {
		return
	}
	edit, edited := android.TakeTextEdit(uint64(field.Field))
	if !edited {
		return
	}
	context.EditTypedField(ui.KeyboardText{Field: field.Field, Text: edit.Text,
		SelectionStart: edit.SelectionStart, SelectionEnd: edit.SelectionEnd,
		ComposingStart: edit.ComposingStart, ComposingEnd: edit.ComposingEnd})
}

// sendTypedField tells the on-screen keyboard the text of the field typed in, so it suggests from it
// and edits it.
func sendTypedField(context *ui.Context) {
	field, typing := context.TypedField()
	if !typing {
		return
	}
	android.EditText(uint64(field.Field), android.TextInput{Text: field.Text,
		SelectionStart: field.SelectionStart, SelectionEnd: field.SelectionEnd,
		ComposingStart: field.ComposingStart, ComposingEnd: field.ComposingEnd}, field.Secret)
}
