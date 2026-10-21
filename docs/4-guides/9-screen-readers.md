# 🗣️ Screen readers

A screen reader speaks what is on the screen for a person who cannot see it, and lets them act on
it: on a phone, TalkBack reads what a finger moves over, and a double tap presses it. It needs the
screen as *things with names*, not pixels. While one is on (`SetScreenReader(true)`, which the
window does on Android), every frame also makes a list of what can be read, worked out from the
boxes themselves, at no cost while none is on.

## How the boxes are read

- **A control is a node**, read by the texts inside it: a button reads "Light the lamp, button". A
  control inside another (a switch in a row) is a node of its own, inside the row.
- **Texts beside each other in one box are read together**: a row's title and the line under it are
  one node, and a paragraph is one, not a node per word.
- **A scroll area holds its rows**, and only those that show are read; the screen reader can scroll
  it on (`ReaderForward`) or back.
- **What is behind a modal or a sheet is left out**, as it cannot be touched.
- **The node the screen reader is on is ringed** (`SetScreenReaderFocus`), dark outside and light
  inside, over everything, as a phone rings what TalkBack reads.

## Describing a control of your own

The controls say what they are and their state: a checkbox ticked, a toggle on (a switch), the tab
shown, a slider's range, a field's text (not a secret one's). `Describe` says it for a control of
your own, after it, in every frame:

```go
if interfaceContext.Clickable("beacon", layout, looks, drawSwitch) {
	beacon = !beacon
}
interfaceContext.Describe(ui.Description{Role: ui.SwitchRole, Label: "Beacon", Checked: beacon})
```

What `Describe` leaves empty stays as the boxes tell it. `Hidden` leaves out a decoration, or a label
read already with the field it names. A control with an icon alone and no label is read by the
`Tooltip` that follows it ("Go back"), as the pointer's note already says what it does. With no
tooltip it is read by the icon's name ("arrow left") for want of a better one: describe it with
what it does. After a menu button
or a select, `Describe` describes the button, even while its list is open over it.

## What the screen reader asks

| Call | What it does |
|---|---|
| `SetScreenReader(on)` | Makes the list of nodes every frame while a screen reader is on |
| `SetScreenReaderFocus(id)` | Rings the node the screen reader is on |
| `ScreenReaderAct(id, action)` | Done in the next frame: a press clicks the control as a tap would (a field takes the keyboard), a long press is read by `LongPressed`, forward and backward scroll an area or move a slider a step |
| `ScreenReaderSetText(id, text)` | Text the screen reader types, pastes or dictates into a field: the field takes it whole, as typing (one line, cut at its `MaxLength`), and the keyboard stays down |

Nodes keep their ids from frame to frame, so the screen reader stays on what it was reading. On
Android, the window hands the list to Go Renderer, which gives it to TalkBack
([Android](../7-platforms/2-android.md#screen-readers)).
