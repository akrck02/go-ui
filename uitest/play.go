package uitest

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-ui/uidrive"
)

// Play plays a script in the window, as a driven window would (uidrive.Player), with a sixtieth
// of a second passing on the window's clock at each frame. It returns the error of the step that
// failed, with its code, or nil when every step passed.
func (window *Window) Play(script uidrive.Script) error {
	defer window.doing(fmt.Sprintf("Play %d steps", len(script)), draw2d.Rectangle{}, false)()
	player := uidrive.Play(script, uidrive.PlayOptions{})
	ownDriver := window.driver
	window.driver = player
	defer func() { window.driver = ownDriver }()
	for !player.Done() {
		window.clock.now = window.clock.now.Add(frameDuration)
		window.Frame()
	}
	window.state.Feed().TakeFrameRequest() // the player asked for frames as it played: they have run
	return window.noted(player.Err())
}
