# 🖱️ Mouse, keyboard and touch

Go UI shares the mouse and the keyboard with what is behind it: a 3D view's camera, the viewer's
shortcuts. This page says who gets them, and how fingers are followed on a touch screen.

## Who gets the mouse

Controls react to the mouse using where they were laid out in the **previous** frame (sizes are
only known at `End`). The pointer is a frame behind what is drawn, which nobody notices at sixty
frames a second.

- **Hover**: the front-most panel under the pointer is the only one whose controls can be hovered.
  Within it, the first control to find the pointer over it takes it; containers that react to the
  mouse (a title bar) look after their content, so the buttons inside them win.
- **Press and click**: a pressed control keeps the mouse until the button is let go, even if the
  pointer strays. A click is a press and a release over the same control.
- **The scene keeps its drags**: a press on empty screen belongs to the camera until it is let go;
  panels it passes over do not take it.
- **Claims**: at `End`, when the pointer is over a panel or a control is pressed, the interface
  calls `state.ClaimMouse()`. Go Renderer's viewer then gives its cameras and shortcuts the input
  without the mouse, so dragging inside a panel does not turn the view. A text field with the focus
  claims the keyboard the same way. `WantsMouse()` and `WantsKeyboard()` tell your own code.
- **Pointer shape**: a hand over buttons, four arrows over a title bar that can be dragged and
  while dragging a payload, an I-beam over a text field, a two-headed arrow over a resize handle.
- **Modal and dragging**: while a sheet is open or a payload is dragged, the interface claims the
  mouse and the keyboard wherever the pointer is.

Which layer is in front, and so gets the mouse first, is drawn in
[panels and layers](4-panels-and-layers.md).

## Fingers

On a touch screen, one finger is followed from where it comes down: the way it goes first decides
who takes it, the innermost scroll area for up and down, a `SwipeBox` for sideways (pages, cards in
a row, a row swiped away), and the control it came down on is then not touched. A finger that stays
still on a control for half a second long-presses it (`LongPressed`).

| Call | What it does |
|---|---|
| `SwipeBox(name, layout, axis, content)` | Takes sideways swipes: pages, cards in a row, a row swiped away |
| `LongPressed()` | Whether the control just added was long-pressed: a finger held still for half a second, or the right mouse button |
| `ScrollPull(name)` | Measures a pull past a list's top, for a list that refreshes |
| `ScrollOffset(name)` | How far an area is scrolled, for a bar that shrinks over it |
| `Context.Touches()` | The fingers on a touch screen, for a control of a program's own that follows them (an on-screen stick); the first finger is also the pointer that clicks |
| `Context.Back` | Answers a phone's back gesture; modals and sheets answer it first |

Scroll areas glide on when let go. How touches reach Go UI on a phone is in [Android](../7-platforms/2-android.md).
