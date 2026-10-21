package window

import (
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uidrive"
)

// driveBefore lets a driver add its input to a frame, after the backend gathered the real input
// and before the interface runs. Without a driver it does nothing.
func driveBefore(driver uidrive.Driver, state *input.State, context *ui.Context) {
	if driver != nil {
		driver.Drive(state, context)
	}
}

// showAfter lets a driver draw over the interface, once the interface has drawn. Without a driver
// it does nothing.
func showAfter(driver uidrive.Driver, context *ui.Context) {
	if driver != nil {
		driver.Show(context.Drawing())
	}
}
