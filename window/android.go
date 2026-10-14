//go:build android

package window

import (
	"hash/maphash"

	"github.com/akrck02/go-renderer/android"
	"github.com/akrck02/go-renderer/gles"
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
	app.Update = func(app *models.Application, step float64) error {
		bars := android.CurrentInsets()
		context.SetInsets(ui.Insets{Top: bars.Top, Bottom: bars.Bottom, Left: bars.Left, Right: bars.Right})
		context.SetTextScale(android.TextScale()) // the person's text size, from the phone's settings
		err := update(app, step)
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
	return android.Run(app, android.Options{MaxFramesPerSecond: options.MaxFramesPerSecond, Hidden: options.Hidden}, draw)
}
