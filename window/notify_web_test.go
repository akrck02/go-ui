//go:build js && wasm

package window

import (
	"syscall/js"
	"testing"
)

// fakeNotifications stands in for the browser's Notification API, allowed, and keeps the last
// notification made, so a test can click it.
func fakeNotifications(test *testing.T) {
	test.Helper()
	js.Global().Call("eval", `
		globalThis.focus = function () {};
		globalThis.Notification = function (title, options) {
			this.close = function () {};
			globalThis.lastNotification = this;
		};
		globalThis.Notification.permission = "granted";
	`)
}

// Clicking a notification hands its target to the program once, so the page shows that thing and
// does not show it again on the next frame.
func TestAClickedNotificationsTargetIsOpenedOnce(test *testing.T) {
	fakeNotifications(test)
	if !NotifyAbout(1, "Lamp 4", "It went out", "lamps/4") {
		test.Fatal("the notification was not shown")
	}
	if _, clicked := OpenedNotification(); clicked {
		test.Fatal("a target was opened before the notification was clicked")
	}
	js.Global().Get("lastNotification").Call("onclick")
	if target, clicked := OpenedNotification(); !clicked || target != "lamps/4" {
		test.Fatalf("OpenedNotification() = %q, %v, want lamps/4, true", target, clicked)
	}
	if _, clicked := OpenedNotification(); clicked {
		test.Fatal("the same click opened its target twice")
	}
}
