# Who has the keyboard

This page is for whoever builds a program with go-ui and wants the keyboard to work for both the
interface and whatever is behind it. At the end you will know which of the two modes to choose, and
what a click, Tab, Enter, Space and Escape do in each.

## The problem

Tab, Enter, Space and the arrows mean two things at once:

- **To an interface**, they move between buttons and press them, so someone who can't use a
  mouse can still use every control.
- **To a game or a 3D view**, they are its own controls: Space jumps, the arrows turn the
  camera, Tab shows the scores.

Both can't have them. If the interface always took Space, a game would stop jumping and press
whichever button was last clicked. So go-ui asks the program one question, through
`context.KeyboardNavigation`: **is there anything else that wants these keys?**

## The two modes

| | `ui.WhenGiven` (the default) | `ui.Always` |
|---|---|---|
| For | a game, a 3D viewer, an editor over a scene | a program that is only an interface: a register, a settings app |
| Tab, Shift+Tab | move between controls only while the interface **has** the keyboard | always move between controls |
| Enter, Space | press the control Tab reached, while the interface has the keyboard | press the control Tab reached |
| A click on a button | presses it, and **leaves the keyboard where it was**; no ring | presses it; no ring; the next Tab goes on from it |
| Escape | gives the keyboard back to the program (an open menu or dialog closes first) | closes an open menu or dialog, or takes the ring off |
| What the program reads | every key while the interface does not have the keyboard | none: the interface claims the keyboard all the time |

The interface **has the keyboard** while:

- a text field is typed in,
- a dialog or a sheet is open, or a menu or a select is open,
- Tab moved the focus while it had the keyboard (from a field to a button, say): it keeps it
  until Escape or a click,
- or the program gave it with `context.TakeKeyboard()`.

With `Always`, it has the keyboard all the time.

`WhenGiven` is the default because forgetting to choose then costs little: Tab doesn't walk the
buttons. The other way round, a game that forgot would stop answering its own keys.

## A game over a 3D view

Leave the mode alone. The game keeps its keys, and its pause menu takes the keyboard while it is
open:

```go
if !paused && state.Pressed(input.KeyP) {
	paused = true
	context.TakeKeyboard() // Tab, Enter and Space now move through the menu
}
context.Begin(state)
if paused {
	context.Panel("Paused", func() {
		if context.PrimaryButton("Resume") {
			paused = false
			context.ReleaseKeyboard() // the keys go back to the game
		}
		context.Button("Settings")
	})
}
context.End()
if paused && !context.HasKeyboard() {
	paused = false // Escape gave the keys back: resume, as Resume does
}

if !context.HasKeyboard() && state.Pressed(input.KeySpace) {
	ship.Jump() // Space is the game's whenever the menu doesn't have it
}
```

Pressing Escape in the menu gives the keyboard back too, and the game resumes on the next frame.
`HasKeyboard()` stays true until the end of any frame in which the interface used Enter, Space or
Escape, so the Space that pressed **Resume** is not also a jump. A click on Resume with the mouse
presses it but takes nothing, so the Space that follows is a jump again.

While the interface has the keyboard, go-ui **claims** it in the input state, as a text field
always has, so go-renderer's viewer leaves its camera keys alone. `context.WantsKeyboard()` tells
your own code the same thing.

## An application that is only an interface

Set the mode once, before the first frame:

```go
context.KeyboardNavigation = ui.Always
```

Now Tab goes through every button, check box, toggle, tab, select, chip and field in the order they
are drawn, and comes round at the end. Shift+Tab goes back. Enter or Space presses the control Tab
reached: it clicks a button, ticks a check box, picks a tab, opens a select. When the control with
the focus goes away (a row deleted), the next Tab goes on from where it stood. A field reached by
Tab selects its text, ready to be typed over.

## What Tab reaches

- **Reached:** `Button` and its variants, `Selectable`, `Clickable`, `Checkbox`, `Toggle`, `Tabs`
  (each tab), `Select`, `Chip`, `TextField`, `NumberField` (its number; the − and + buttons are
  skipped, since the number can be typed).
- **A panel's fold button**, in its title bar, comes before the panel's content.
- **Inside a dialog or a sheet:** Tab goes round its own controls only.
- **Inside menus and lists:** the arrows move, not Tab; that is unchanged.

## How the focus looks

The ring only shows when the focus came from the keyboard. A click never shows it.

The ring lies on the control's edge, with no gap, in `Theme.FocusRing`, `Theme.FocusWidth` wide
(3 pixels by default). On a primary button it is drawn **inside**, in `Theme.OnPrimary`: a ring in
the primary colour around a button of that colour could not be seen.

A theme can also ask a field to light its own edge instead of drawing a ring around it
(`Theme.FieldFocusEdge`), and a held control to shrink a little (`Theme.PressScale`, 0.97 in
Bubble). If the person asked the system for less motion, set `context.LessMotion` and held controls
keep their size. The tokens are in [a design system's look](9-design-systems.md).

## What it costs

While the interface does not have the keyboard, no control joins the Tab order, and a frame of
three hundred controls takes about 1 % longer than before, from one comparison per control. While
it has the keyboard, it takes up to 2.5 % longer. Neither allocates anything more. Moving the
focus asks for one frame, so the ring shows at once even in an application that only draws when
something changes. The measurements are in the
[performance audit](../8-project/1-performance-audit.md#keyboard-focus).

---

[← Components](0-components.md)
