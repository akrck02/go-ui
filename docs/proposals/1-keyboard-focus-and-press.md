# 1. Keyboard focus on every control, and a press that gives

> Status: planned · Author's approval: proposal ☑ plan ☑

## Why

Today the keyboard only reaches text fields: Tab goes from one field to the next, and buttons, check
boxes and choices can only be used with a pointer. That breaks the first rule of an accessible
interface, that the keyboard reaches everything. Bubble 4 has also settled how focus and pressing
should look (focus lights the control's own edge; a control shrinks a little while pressed), and
go-ui should let a design system ask for that through its theme, as it already does for colours
and corners.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | While the interface has the keyboard, Tab and Shift+Tab move the focus through every control that can be used, fields included, in the order they are drawn | test `TestTabVisitsEveryControlInTheOrderItIsDrawn` |
| R2 | Enter or Space presses the focused button, ticks the focused check box, picks the focused choice | test `TestTheKeyboardUsesTheFocusedControl` |
| R3 | The focused control shows a ring only when the focus came from the keyboard, never after a click or a tap | test `TestAClickGivesFocusWithoutARing` |
| R4 | The ring lies on the control's edge; on a control filled with the primary colour it is drawn inside, in the colour of its text | tests `TestTheRingLiesOnTheEdge`, `TestTheRingOfAPrimaryButtonIsInside` |
| R5 | A theme can ask a field to light its own border instead of drawing a ring around it | test `TestAFieldCanLightItsOwnBorder` |
| R6 | A theme can ask controls to shrink a little while held; with less motion asked by the system, they don't | tests `TestAHeldButtonShrinks`, `TestLessMotionKeepsAHeldButtonStill` |
| R7 | A theme made before these tokens draws exactly as before, apart from the ring that keyboard focus now shows | `cmd/app` captures before and after, compared pixel by pixel with no keyboard focus |
| R8 | Bubble's `goui` adapter sets the new tokens, so a Bubble theme in go-ui behaves like Bubble's CSS | test in `bubble/goui`; the showcase compared with the browser |
| R9 | go-ui only takes Tab, Enter, Space and the arrows while the interface has the keyboard: a field is being typed in, a dialog, sheet or menu is open, the program gave it the keyboard, or the program says the interface is all there is. Otherwise every key reaches the program: a 3D view or a game keeps them | tests `TestAGameKeepsTabAndSpaceWhileTheInterfaceDoesNotHaveTheKeyboard`, `TestAClickOnAButtonDoesNotTakeTheKeyboard`, `TestEscapeGivesTheKeyboardBack` |
| R10 | A frame costs the same as today: no allocation added per frame, the 300-control benchmark within 3 % of today, and a still interface still draws nothing new | `testing.AllocsPerRun` on a frame with focus; `BenchmarkThreeHundredControls` before and after; the existing idle test |
| R11 | The docs explain who has the keyboard so a newcomer picks the right mode: a guide with a game or 3D view (`WhenGiven`) and a form-only app (`Always`), what a click does and does not do, how Escape and `TakeKeyboard` work, and a table of which keys each mode reads | The guide's examples run as shown; the author reads it |

## Out of scope

- Arrow keys inside lists, menus and grids: Tab reaches the list, and moving inside it is a later
  proposal.
- Changes for screen readers: they already have their own focus (`SetScreenReaderFocus`).
- Animating the press: it shrinks while held and comes back, with no tween, until go-ui has one.

## Design

- **Who has the keyboard**: the program decides, through one field on the context:
  - `KeyboardNavigation: WhenGiven` (the default, chosen by the author: if a program forgets to
    choose, the worst case is that Tab does not walk the buttons, not that a game stops answering): the interface has the keyboard only while a
    field is typed in, a modal layer (dialog, sheet, menu) is open, or the program calls
    `TakeKeyboard()`. Escape, or `ReleaseKeyboard()`, gives it back. A click on a button does not
    take the keyboard, so Space in a game never presses the last button clicked.
  - `KeyboardNavigation: Always`: for programs that are only an interface (a register, a settings
    app). Tab always moves the focus.
  - While the interface has the keyboard, go-ui claims it in the input state, as fields do today,
    so go-renderer's viewer and a game's own controls know to stay quiet. Otherwise it claims
    nothing and reads no key.
- **Cost**: the focus order is a slice of IDs kept on the context and reset to length zero each
  frame, so it allocates nothing once warm; nothing is added to it while the interface does not
  have the keyboard. The ring is one stroke, only on the focused control. The press scales one
  element while it is held. Moving the focus changes the drawing, so it redraws; a still interface
  with no keys pressed draws nothing new, as today.
- **Focus order**: every `pressable` control and every field joins a list in drawing order, each
  frame. The `tabFrom` and `tabBack` that fields already use move over that list instead of the
  fields alone. The focus stays an `ID`, so it survives from frame to frame as it does today.
- **Keyboard or pointer**: a new `focusFromKeyboard` flag is set by Tab and cleared by any press of
  the pointer. The ring is only painted while it is set, which is what CSS calls `:focus-visible`.
- **Theme tokens**, all zero for go-ui's own look:
  - `FocusWidth`: the ring's width. Zero keeps today's 3 pixels.
  - `FieldFocusEdge`: when true, a field colours its border with `FocusRing` and draws no ring.
  - `PressScale`: how big a held control is drawn (0.97 in Bubble). Zero means it doesn't change.
  - The ring of a control filled with `Primary` is drawn inside it in `OnPrimary`, always: a ring
    in `FocusRing` could not be told from the fill.
- **Less motion**: go-ui reads it from go-renderer's input state if the platform gives it. If not,
  a `ReducedMotion` field on the context lets the program say so.
- **Painting**: the ring is one more `Rectangle` stroke in `paint.go`, from the element's rectangle
  and corner. The press scales the element's rectangle around its centre before painting it and
  its children, so text and icons shrink with it.
- Not taken: rings that follow the browser's look with a gap. Bubble chose an edge with no gap, and
  a gap only looks right on the page's own background.

## Tasks

- [ ] T1 — Who has the keyboard: `KeyboardNavigation`, `TakeKeyboard`, `ReleaseKeyboard`, Escape · `ui.go`, `interaction.go` · R9 · tests: `TestAGameKeepsTabAndSpaceWhileTheInterfaceDoesNotHaveTheKeyboard`, `TestAClickOnAButtonDoesNotTakeTheKeyboard`, `TestEscapeGivesTheKeyboardBack`
- [ ] T2 — The focus order and Tab over every control · `interaction.go`, `textfield.go`, `memory.go` · R1 R10 · tests: `TestTabVisitsEveryControlInTheOrderItIsDrawn`, `TestAFrameWithFocusAllocatesNothing`
- [ ] T3 — Enter and Space use the focused control · `controls.go`, `widgets.go`, `choices.go` · R2 · test: `TestTheKeyboardUsesTheFocusedControl`
- [ ] T4 — The ring, only for keyboard focus, on the edge or inside a filled control · `paint.go`, `theme.go` · R3 R4 R7 · tests: `TestAClickGivesFocusWithoutARing`, `TestTheRingLiesOnTheEdge`, `TestTheRingOfAPrimaryButtonIsInside`
- [ ] T5 — `FieldFocusEdge` · `textfield.go`, `theme.go` · R5 · test: `TestAFieldCanLightItsOwnBorder`
- [ ] T6 — `PressScale`, and less motion · `paint.go`, `theme.go`, `context.go` · R6 · tests: `TestAHeldButtonShrinks`, `TestLessMotionKeepsAHeldButtonStill`
- [ ] T7 — Docs: a guide «Who has the keyboard» (`docs/5-components/`), with a game over a 3D view and a settings app side by side, what a click does, Escape and `TakeKeyboard`, and a table of keys per mode; the components' keyboard sections and the design-system tokens table; a line in the README's example · R1–R6 R11
- [ ] T8 — Bubble's adapter sets the tokens; compare the showcase with the browser · `bubble/goui/theme.go` · R8 · test: `TestABubbleThemeLightsFieldsAndShrinksWhenPressed`
- [ ] T9 — Measure: the 300-control benchmark and the register's audit benchmarks before and after, written into `docs/8-project/1-performance-audit.md` · R10

## Result

Not built yet.
