//go:build js && wasm

package window

import (
	"hash/maphash"
	"strconv"
	"syscall/js"

	"github.com/akrck02/go-renderer/browser"
	"github.com/akrck02/go-renderer/gles"
	ui "github.com/akrck02/go-ui"
)

// Wake makes the page draw a frame now: call it when the program's data changes on another
// goroutine (a fetch finished). Safe from any goroutine.
func Wake() { browser.Wake() }

// Graphics describes a phone's OpenGL ES and GPU; in a browser it is "".
func Graphics() string { return "" }

// LastCrash returns the report of the last run's crash; in a browser the console shows it, and this
// is always "".
func LastCrash() string { return "" }

// Stop ends the window in a web page: Run returns after one more frame and the canvas is let go.
// For a page whose own application shows the program on one of its screens, and takes the screen
// away. Safe from any goroutine.
func Stop() { browser.Stop() }

// SystemPrefersDark tells whether the browser is set to a dark theme (prefers-color-scheme).
func SystemPrefersDark() bool {
	query := js.Global().Call("matchMedia", "(prefers-color-scheme: dark)")
	return query.Truthy() && query.Get("matches").Bool()
}

// Notify shows a notification through the browser (the Notification API), once the person has
// allowed the page's; the id replaces the one shown with it before. It returns false when it could
// not.
func Notify(id int, title, text string) bool {
	if !NotificationsAllowed() {
		return false
	}
	options := map[string]any{"body": text, "tag": "notification-" + strconv.Itoa(id)}
	js.Global().Get("Notification").New(title, options)
	return true
}

// NotificationsAllowed tells whether the page may show notifications.
func NotificationsAllowed() bool {
	notification := js.Global().Get("Notification")
	return notification.Truthy() && notification.Get("permission").String() == "granted"
}

// AskForNotifications shows the browser's question to allow the page's notifications. Browsers
// ask only just after the person touched or clicked something: call it from such a step.
func AskForNotifications() {
	notification := js.Global().Get("Notification")
	if notification.Truthy() && notification.Get("permission").String() == "default" {
		notification.Call("requestPermission")
	}
}

// FeelLongPress gives a short buzz for a long press, where the browser can (the Vibration API, on
// phones); elsewhere nothing.
func FeelLongPress() {
	navigator := js.Global().Get("navigator")
	if navigator.Get("vibrate").Truthy() {
		navigator.Call("vibrate", 25)
	}
}

// OpenLink opens a link in a new tab, or in the application that shows it (mailto:, tel:).
func OpenLink(link string) bool {
	return js.Global().Call("open", link, "_blank", "noopener").Truthy()
}

// Share offers a text to the person's other applications, through the browser's sheet where it
// has one (phones' browsers, mostly): false where it has none.
func Share(title, text string) bool {
	navigator := js.Global().Get("navigator")
	if !navigator.Get("share").Truthy() {
		return false
	}
	navigator.Call("share", map[string]any{"title": title, "text": text})
	return true
}

// whenHidden calls the program's Hidden when the page is hidden: another tab, the phone locked,
// the page closing.
func whenHidden(hidden func()) {
	if hidden == nil {
		return
	}
	document := js.Global().Get("document")
	document.Call("addEventListener", "visibilitychange", js.FuncOf(func(js.Value, []js.Value) any {
		if document.Get("visibilityState").String() == "hidden" {
			hidden()
		}
		return nil
	}))
}

// SystemPalettes returns the colours a phone made from its wallpaper; a browser does not tell them.
func SystemPalettes() (Palettes, bool) { return Palettes{}, false }

// RunWith is Run with an interface context of your own (a custom theme or measurer). In a web page
// the window is the page's canvas (Options.Canvas); it draws through WebGL2 and returns only when
// the program stops, which keeps the WebAssembly program alive.
func RunWith(context *ui.Context, options Options, build func(context *ui.Context)) error {
	whenHidden(options.Hidden)
	canvas, err := browser.OpenCanvas(options.Canvas)
	if err != nil {
		return err
	}
	// the atlas already smooths every edge of an interface; a scene's geometry needs antialiasing
	functions, err := canvas.WebGL(options.Scene != nil)
	if err != nil {
		return err
	}
	renderer, err := gles.NewFlatRenderer(functions)
	if err != nil {
		return err
	}
	var sceneRenderer *gles.SceneRenderer
	if options.Title != "" {
		js.Global().Get("document").Set("title", options.Title)
	}
	shown := shownFrame{seed: maphash.MakeSeed()}
	view := &shownScene{}
	app := interfaceApplication(context, options, build, &shown, view)
	draw := func() error {
		width, height, scale := frameSize(app)
		if view.view == nil {
			renderer.Draw(context.Drawing(), context.Theme.Background, width, height, scale)
			return nil
		}
		if sceneRenderer == nil {
			created, err := gles.NewSceneRenderer(functions)
			if err != nil {
				return err
			}
			sceneRenderer = created
		}
		sceneRenderer.OverlayScale = scale
		sceneRenderer.Draw(view.view.World, view.view.Camera, width, height)
		return nil
	}
	return browser.Run(canvas, app, draw)
}
