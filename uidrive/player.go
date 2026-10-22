package uidrive

import (
	"io"

	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// PlayOptions choose how a Player plays.
type PlayOptions struct {
	// Report, when set, gets a line for each step and one at the end (report.go): standard
	// output, Android's log, the 3DS's console.
	Report io.Writer
	// Show draws the pointer, the fingers and the element each step found over the interface, so
	// a demo or a tutorial is seen, not only run (show.go).
	Show bool
}

// Player plays a script in a window, one frame at a time: it is a Driver. Each step finds its
// element in what the last frame showed, then gives its input over the frames that follow, as a
// person would, and lets one more frame show what it changed. A step that fails ends the script.
type Player struct {
	script  Script
	options PlayOptions
	hand    Hand
	index   int    // the step playing now
	moves   []move // what the step playing now still has to do, a frame each
	planned bool   // the step playing now has found its element and made its moves
	started bool   // a frame has run since the first Drive, so its nodes can be searched
	found   Element
	done    bool
	err     error
}

// move is one frame of a step: it gives its input and tells whether it is finished; a move that
// waits gives none and answers false until its time has come.
type move func(state *input.State, context *ui.Context) (finished bool, err error)

// Play makes a player for a script; give it to window.Options.Driver, or to uitest's Play.
func Play(script Script, options PlayOptions) *Player {
	return &Player{script: script, options: options}
}

// Done tells whether the script has ended: every step played, or one failed.
func (player *Player) Done() bool { return player.done }

// Err is why the script ended early: the failed step's error, with its code. It is nil while the
// script plays and when every step passed.
func (player *Player) Err() error { return player.err }

// Drive plays a frame of the script. Elements are found through the screen reader's nodes in what
// the last frame showed, so the first frame turns the screen reader on and only lets the interface
// run, to make them. Any input already in the state when Drive is called came from a person (a
// key, a click, the pointer moved): it stops the script with ScriptStopped, so an automatic mode
// never fights the hand on the device.
func (player *Player) Drive(state *input.State, context *ui.Context) {
	if player.done {
		return
	}
	if len(player.script) == 0 {
		reportScriptPassed(player.options.Report, 0)
		player.done = true
		return
	}
	if player.started && state.Feed().HasInput() {
		player.fail(state, scriptStopped(player.index+1))
		return
	}
	state.RequestFrame(0)         // a window that draws on demand keeps drawing while a script plays
	context.SetScreenReader(true) // on a phone the window may turn it off before each frame
	if !player.started {
		player.started = true
		return
	}
	if err := player.playFrame(state, context); err != nil {
		player.fail(state, err)
	}
}

// playFrame plans the step playing now if it is new, and plays its next move.
func (player *Player) playFrame(state *input.State, context *ui.Context) error {
	step := player.script[player.index]
	if !player.planned {
		moves, found, err := plan(&player.hand, step, context)
		if err != nil {
			return err
		}
		player.moves, player.planned = moves, true
		if found.Label != "" { // a wait keeps showing what the step before it found
			player.found = found
		}
	}
	finished, err := player.moves[0](state, context)
	if err != nil || !finished {
		return err
	}
	player.moves = player.moves[1:]
	if len(player.moves) == 0 {
		reportPassed(player.options.Report, player.index+1, step)
		player.index++
		player.planned = false
	}
	if player.index == len(player.script) {
		reportScriptPassed(player.options.Report, len(player.script))
		player.done = true
	}
	return nil
}

// fail ends the script at the step playing now: it lets go of what it holds, so nothing stays
// pressed, and reports why.
func (player *Player) fail(state *input.State, err error) {
	player.hand.LetGo(state)
	player.done, player.err = true, err
	reportFailed(player.options.Report, player.index+1, player.script[player.index], err)
}
