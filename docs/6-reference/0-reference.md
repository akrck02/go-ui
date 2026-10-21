# Reference

go-ui has four packages:

| Package | Import | What it is |
|---|---|---|
| [`ui`](1-ui.md) | `ui "github.com/akrck02/go-ui"` | The context, layout, every control, themes, screen readers |
| [`window`](2-window.md) | `"github.com/akrck02/go-ui/window"` | A window that runs an interface: on the desktop, in a web page, on Android; preferences, notifications, films |
| [`uitest`](3-uitest.md) | `"github.com/akrck02/go-ui/uitest"` | A pretend window for `go test`: press, type, wait and find what an interface shows, with no display |
| [`uidrive`](4-uidrive.md) | `"github.com/akrck02/go-ui/uidrive"` | Drive a real window by the names people see: find elements, play and record scripts, for demos, tutorials, replays and tests on a device |

[`ui.md`](1-ui.md) and [`window.md`](2-window.md) list every type and function with the first sentence
of its comment; they are generated from the doc comments (`python3 docs/6-reference/generate.py`).
`go doc github.com/akrck02/go-ui.Context.Table` prints any of them whole. This page is the map:
the context's methods by what they are for.

## The frame

| Method | |
|---|---|
| `ui.New() (*Context, error)` | A context that measures text with go-renderer's fonts; `NewWithMeasurer` takes another (tests) |
| `Begin(state *input.State)`, `End()` | Open and close a frame: `End` lays out, paints and claims the mouse and keyboard. `window.Run` calls both for you |
| `Drawing() *draw2d.List` | What go-renderer paints: give it to `scene.Overlay.Drawing` once |
| `Push(name)`, `Pop()` | Open and close a scope, so controls in a loop have ids of their own |
| `Area(path...) (draw2d.Rect, bool)` | Where a control was drawn in the last frame, by its scopes and label |
| `RequestFrame(delay)` | Ask for another frame within a delay, for something that moves |
| `Size() (width, height)`, `Insets()` | The window's size, and how much of each edge the system's bars cover |
| `Theme` (field) | The tokens every control reads: `ui.LightTheme()`, `ui.DarkTheme()` |

## Layout

| Method | |
|---|---|
| `Box(layout, content)` | Arranges its children |
| `Frame(layout, look, content)` | Arranges and paints: fill, border, radius, shadow, clip |
| `Row(content)`, `Column(content)`, `Grid(columns, content)` | Shortcuts with the theme's gap |
| `Spacer()`, `Separator()` | Free space; a thin line |
| `ScrollArea(name, layout, content)` | Content that scrolls; `ScrollOffset(name)`, `ScrollPull(name)` |
| `Resizable(name, side, &size, minimum, maximum, layout, content)` | A box with an edge to drag |
| `Fold(name, open, duration, layout, content)` | A box that opens to its content's height and closes, with motion |
| `SwipeBox(name, layout, axis, content)` | A box a finger swipes sideways: pages, cards |

Sizes: `ui.Fit()`, `ui.Grow()`, `ui.GrowBy(weight)`, `ui.Share(weight)`, `ui.Fixed(pixels)`,
`ui.Percent(share)`, each `.Between(minimum, maximum)`. Padding: `ui.Pad(all)`,
`ui.PadXY(horizontal, vertical)`. Alignment: `ui.Start`, `ui.Center`, `ui.End`.

## Text and pictures

| Method | |
|---|---|
| `Heading(text)`, `Label(text)`, `Caption(text)` | A title, body text, a quiet hint |
| `Text(text, ui.TextStyle{Size, Weight, Color})` | Any size, weight and colour |
| `Icon(name)`, `IconSized(name, size, colour)` | A [Lucide](https://lucide.dev/icons) icon |
| `Image(picture, layout, radius)` | A `draw2d.Image`, cropped to fill its box |
| `Paint(layout, paint func(area, drawing))` | A box you paint yourself, once laid out |
| `SetTextScale(scale)`, `TextScale()` | Every text larger or smaller (the system's font size) |

## Controls

| Method | Returns `true` when |
|---|---|
| `Button(label)`, `PrimaryButton(label)` | Clicked |
| `IconButton(icon, label)`, `PrimaryIconButton(icon, label)` | Clicked; square with an empty label |
| `FlatButton(label)`, `FlatIconButton(icon, label)` | Clicked; no fill until hovered (toolbars) |
| `Selectable(icon, label, selected)` | Clicked |
| `Clickable(name, layout, looks, content)` | Clicked: any box, with a look per state |
| `Checkbox(label, &value)`, `Toggle(label, &value)` | Changed |
| `Slider(name, &value, minimum, maximum)`, `SliderStep(…, step)` | Changed |
| `NumberField(name, &value, minimum, maximum, step, format)` | Changed |
| `TextField(name, &text)`, `TextFieldWith(name, &text, ui.TextFieldOptions{…})` | Changed |
| `Select(name, &selected, options)` | Changed |
| `Tabs(name, &selected, labels...)` | Changed |
| `ListBox(name, layout, items, &selected)` | Changed |
| `Chip(icon, label, ui.ChipLook{…})` | Clicked; `RemovableChip` returns `clicked, removed` |
| `MenuButton(icon, label, content)` with `MenuItem(icon, label, hint)` | `MenuItem`: chosen |
| `Tooltip(text)` | Explains the control before it |
| `LongPressed()` | The control before it was held, or right-clicked |

Text fields also have `FieldFocused(name)`, `FocusField(name)`, `Focus(path...)`, `Typing()`, the
clipboard (`Copy(text)`, `ClipboardText()`), and for on-screen keyboards `TypedField()` and
`EditTypedField(edit)`; the bar over a selection is worded by `SetEditingWords`.

## Lists and tables

| Method | |
|---|---|
| `List(name, layout, count, rowHeight, row func(index))` | Equal rows, only those in sight built |
| `Table(name, layout, columns, rows func(*ui.Table))` | A header and rows; `table.Group(caption)`, `table.Row(key, selected, cells...)` returns `true` when clicked |

## Layers

| Method | |
|---|---|
| `PanelWith(title, ui.PanelOptions{…}, content)`, `Panel(title, content)` | A floating card with a title bar, or docked to an edge |
| `Sheet(title, ui.SheetOptions{Width}, content) (dismissed bool)` | A modal card |
| `Modal(name, ui.LayerOptions{…}, content) (dismissed bool)` | Your content over everything, dimming the rest |
| `Overlay(name, ui.LayerOptions{…}, content)` | Your content over everything, without dimming |
| `Popover(name, &open, look, content)` | A box under the control before it |
| `Toast(text, warning)` | A short note at the bottom, fired once |

## Pointer, touch and keyboard

| Method | |
|---|---|
| `WantsMouse()`, `WantsKeyboard()` | Whether the interface claimed them in the last frame |
| `KeyboardNavigation` (`WhenGiven`, `Always`), `TakeKeyboard()`, `ReleaseKeyboard()`, `HasKeyboard()`, `FocusedID()` | When Tab, Enter and Space move between controls and press them, and when they stay with a game or a 3D view ([who has the keyboard](../5-components/10-the-keyboard.md)) |
| `LessMotion` | Held controls keep their size, for people who asked the system for less motion |
| `Pointer() (x, y)`, `Pressing()` | Where the pointer is; whether it is held |
| `Touches()` | The fingers on a touch screen |
| `Back()` | Answers the system's back (a phone's gesture); sheets and modals answer first |
| `DragSource(name, payload, layout, content)`, `DropTarget(name, layout, accepts, content)` | Drag a value of yours from a box to another |
| `Dragging()`, `DroppedOnScreen()` | A drag in progress; a drop on the open screen (the scene) |

## Motion

| Method | |
|---|---|
| `Animate(name, target, duration) float32` | Eases a value to its target, asking for frames while it moves |
| `Appear(name, duration) float32` | Eases from 0 to 1 the first time something is shown |
| `LayerOptions.Leave` | A modal or overlay leaves with motion |

## Screen readers

| Method | |
|---|---|
| `Describe(ui.Description{Role, Label, Value, Checked, …})` | What the control before it is, for a screen reader |
| `SetScreenReader(on)`, `ScreenReaderOn()`, `ScreenReaderNodes()` | The list of what can be read, made each frame while one is on |
| `ScreenReaderAct(id, action)`, `ScreenReaderSetText(id, text)`, `SetScreenReaderFocus(id)` | What the screen reader asks, done in the next frame |

The window does all of this on Android; see [the guide](../4-guides/9-screen-readers.md).

## The window

| Function | |
|---|---|
| `window.Run(options, build)` | Opens the window and calls `build` every frame |
| `window.RunWith(context, options, build)` | The same with a context you made (to set a theme or a measurer first) |
| `window.Options{…}` | `Title`, `Width`, `Height`, `MinWidth`, `MinHeight`, `Dark`, `Continuous`, `Backend`, `Canvas`, `Scene`, `MaxFramesPerSecond`, `Hidden`, `Shown`, `CapturePath`; `RegisterFlags(flags)` |
| `window.Wake()` | A frame, from another goroutine |
| `window.KeepPreference`, `Preference`, `ForgetPreference` | Settings kept between runs |
| `window.SystemPrefersDark()`, `SystemPalettes()` | The system's theme and, on Android 12+, the wallpaper's colours |
| `window.Notify`, `AskForNotifications`, `NotificationsAllowed` | Notifications on a phone |
| `window.OpenLink(link)`, `Share(title, text)`, `FeelLongPress()` | The system's browser, its share sheet, a buzz |
| `window.OpenVideo(film)` | A film the phone or the browser decodes |
| `window.Graphics()`, `LastCrash()` | What draws; the last crash's report, on a phone |

---

[← Documentation](../4-guides/1-getting-started.md)
