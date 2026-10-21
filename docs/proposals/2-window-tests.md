# 2. Window tests: drive an interface in `go test`, with no display

> Status: building · Author's approval: proposal ☑ (2026-10-21) plan ☑ (2026-10-21)

## Why

Go UI's own tests already drive interfaces frame by frame: a private `harness` in `ui_test.go`
makes an `input.State` of 800×600, moves a fake mouse, runs `Begin`, the interface and `End`, and
reads where an element was laid out. Programs built on Go UI (an editor, a settings app, a game's
menus) cannot use it: it is private, it uses a fake measurer of the test file, and it finds
elements by hashing a label by hand. Each program ends up writing its own copy, or not testing
its interface at all.

Like Go's `net/http/httptest`, Go UI should ship a small public package for that: open a pretend
window, press, type and wait in it, and ask what it shows. No display, no OpenGL, no cgo, so it
runs in any `go test` and in CI.

## What

The package is called `uitest` (chosen by the author: tests of a Go UI interface, as `httptest` tests a server). Its tests are
written against an interface made only for them: a lighthouse keeper's log with a button, a
check box, a text field and a dialog.

| Id | Requirement | Check |
|---|---|---|
| R1 | `uitest.New` makes a test window of a size in layout pixels, with a scale, a theme and a text measurer to choose (a measurer with fixed widths by default, so results are the same on every machine), and runs an interface function for one frame (`Frame`) or several (`Frames(count)`) exactly as a backend does: `Begin`, the function, `End`, the end of the input step | test `TestAFrameRunsTheInterfaceOnceLikeABackend` |
| R2 | The pointer: move to a point or to an element, press, release, click, double click, drag from one point or element to another over several frames, and scroll | tests `TestAClickOnAButtonPressesIt`, `TestADragMovesASlider`, `TestScrollingShowsWhatWasBelow` |
| R3 | The keyboard: press, hold and release a key, Tab and Shift+Tab, and type a text, one character per call or all in one frame | tests `TestTypedTextReachesTheFocusedField`, `TestTabMovesTheFocusToTheNextControl` |
| R4 | Touch: put a finger down, move it and lift it, a tap and a long press | tests `TestATapPressesAButton`, `TestALongPressOpensTheMenu` |
| R5 | Time: the window has its own clock, which only moves when a test says so (`Wait(duration)`), so double clicks, long presses and animations give the same result on every run | test `TestALongPressNeedsItsTime` |
| R6 | Finding what was drawn: an element by its label or name (the one a screen reader gets, so labels that repeat are told apart by their scope), its rectangle, whether it is visible (on screen, not hidden by a scroll area or a layer above), whether it has the focus, and every text drawn in the frame in order | tests `TestAnElementIsFoundByItsLabel`, `TestAnElementScrolledAwayIsNotVisible`, `TestTheTextsOfAFrameComeInDrawingOrder` |
| R7 | Failures come back as errors with a code (`uitest.element-not-found`, `uitest.element-not-visible`, `uitest.element-ambiguous`, …), declared once in the package's `errors.go`; nothing in the package panics or calls `t.Fatal` for the test. The message says what was asked and lists the labels that were on screen | test `TestEachErrorCarriesItsCode`; a test that a missing label gives `element-not-found` |
| R8 | The draw list of the last frame can be read as it is (`DrawList`) for tests that check colours or shapes | test `TestTheDrawListOfTheLastFrameCanBeRead` |
| R9 | Go UI's own tests move to `uitest` where it does the same as today's private harness, so the package is used every day and stays honest | `go test ./...` passes; the private harness is replaced wherever `uitest` does the same (the author's decision) |
| R10 | An example test that runs (`ExampleWindow`, checked by `go test`) and a guide page in `docs/4-guides/` for someone testing their own program: open a window, click, type, find, and read an error | the example's output is checked; the author reads the guide |
| R11 | Using the package costs nothing to programs that don't: it lives in its own package, the `ui` package does not import it, and no field of `ui.Context` is added only for it | `go list -deps` of `ui` does not name `uitest` |
| R12 | *Added by the author, 2026-10-21.* A mode that shows the results in a Go UI window, visual and pretty: `go run ./cmd/uitest ./...` runs the tests with recording on and opens a window listing every test that used `uitest`, passed or failed, with its error codes and messages; for each test a replay of its frames, drawn from the recorded draw lists, stepped frame by frame, with the pointer, the clicks and the element each action found highlighted. Built with Go UI's own controls and theme (cards, icons, chips, the theme's colours), nothing hand-made where Go UI has a control. Recording happens only while the mode is on, so plain `go test` pays nothing | tests with `uitest` of the viewer's logic (loading recordings, the list, stepping frames, the highlight); a test that a window with recording off keeps nothing; the window itself needs a display, so its look is the author's check

## Out of scope

- A picture of a frame without a display: the viewer of R12 paints recorded draw lists in a real
  window, which needs a display.
- A PNG of a frame and golden images (the author's decision): Go Renderer has no software painter
  today (its `software/` package is a stub), so there is no way to paint a frame without a display.
  Once Go Renderer can paint a draw list into an `image.RGBA`, a `Picture` method and a golden-file
  helper are a separate proposal. `DrawList` (R8) is what tests read until then.
- Testing through a real backend window (`window/`): this package drives `ui.Context` and
  `input.State` directly; the backends keep their own tests.
- Recording a session by hand and playing it back.
- Game pads, and the screen-reader actions beyond what R6 reads (they already have
  `ReaderPress` and friends).

## Design

### The pieces

`uitest/` is a new package at the root, beside `window/`. It imports `ui`, `draw2d` and `input`
only; `ui` never imports it. Its files, one part each:

| File | What it holds |
|---|---|
| `window.go` | `Window`: a `*ui.Context`, an `*input.State`, the interface function, the clock. `New(build, options...)`, `Frame`, `Frames(count)`, `Resize`, `Context`, `State` (an escape hatch for what the package does not cover yet) |
| `options.go` | `Size(width, height)`, `Scale(factor)`, `Theme(theme)`, `Measurer(measurer)`: functional options, so `New(build)` works with nothing else |
| `measurer.go` | `FixedMeasurer`: each character half the text size wide, as the private `fakeMeasurer` does today; the default |
| `clock.go` | the window's clock, set into `input.State.Now`; `Wait(duration)` moves it and runs the frames that fall in that time at a fixed step (`frameDuration`, one sixtieth of a second) |
| `pointer.go` | `MoveTo`, `MoveToElement`, `Press`, `Release`, `Click`, `DoubleClick`, `Drag`, `Scroll` |
| `keyboard.go` | `PressKey`, `HoldKey`, `ReleaseKey`, `Tab`, `ShiftTab`, `Type` |
| `touch.go` | `TouchDown`, `TouchMove`, `TouchUp`, `Tap`, `LongPress` |
| `find.go` | `Element(label)` and `ElementIn(scope, label)` return an `Element` (label, `Rect`, `Visible`, `Focused`, `Pressable`); `Texts()`; `DrawList()` |
| `errors.go` | the codes, at the top, and one constructor each |
| `example_test.go`, `*_test.go` | the tests named in What, beside the code |

Every action is one or more whole frames: an action changes the input state, runs `Frame`, and
returns an error when its target is missing. `Click` is press, frame, release, frame, because a
control in Go UI answers on release and needs the frame between.

### Finding elements

The package reads the screen-reader tree (`ScreenReaderNodes`), which already gives each control
its ID, label, rectangle (already clipped by scroll areas) and whether it can be pressed. The test
window switches the screen reader on when it is made. Two small public additions to `ui` are
needed, both generic and useful to any program:

- `(*Context).FocusedID() ID`: which control has the focus. Today it is private.
- `AccessibleNode.Hidden`, or an equivalent rule: a node under a modal layer is not visible. If
  the tree already leaves those nodes out, nothing is added; T2 checks which.

`Texts()` walks the draw list for `TextKind` commands, so it also finds text that is not a control
(a heading, a status line).

Errors are a small type: `Error{Code, Message}` with `Code` a string constant, compared with
`errors.Is` against a sentinel per code (`uitest.ErrorElementNotFound`), so returning one only
allocates when a test fails. Codes: `uitest.element-not-found`, `uitest.element-ambiguous`,
`uitest.element-not-visible`, `uitest.element-not-pressable`, `uitest.invalid-size`.

### Go UI's own tests

Tests inside `package ui` cannot import `uitest`: `uitest` imports `ui`, and that would be a cycle.
So the tests that only use public behaviour (click a button, type in a field, Tab through
controls) move to `package ui_test` files and use `uitest`; the tests that read private state
(`ui.previous`, the focus order, memory) keep a much smaller private harness. T6 lists which is
which before moving anything.

### The results viewer (R12)

- **Recording** (`uitest/recording.go`): a window made with the `Test(t)` option and the
  environment variable `UITEST_RECORD` set to a folder keeps, for each frame, a copy of the draw
  list's commands, the pointer, the fingers, the action that ran it (`Click "Save"`) and the
  rectangle of the element that action found. Errors returned by the package are kept too. At the
  end of the test (`t.Cleanup`) it writes one JSON file per test into that folder, with the test's
  name and whether it failed. With the variable unset, `Test(t)` only remembers the name: no copy,
  no file. Pictures in a draw list are kept by size only (a grey box in the replay).
- **The program** (`cmd/uitest/`): `go run ./cmd/uitest ./...` makes a fresh folder, runs
  `go test` with `UITEST_RECORD` set and the arguments given, then opens a `window` with the
  results. Its logic (`results.go`: load the folder, sort failed tests first, step through frames)
  is apart from its look (`view.go`), so the logic is tested with `uitest` itself, headless.
- **The look**: a side list of tests as `Selectable` rows with a check or cross icon and a `Chip`
  for each error code; the chosen test in a card with its name, its messages and the replay: the
  recorded draw list painted with `Paint`, scaled to fit, the pointer as a ring, a press as a
  filled ring, and the found element outlined in the theme's `FocusRing`. Buttons and a `Slider`
  step through the frames; the left and right arrows too.
- **Not taken**: an HTML report (the author asked for a Go UI window); recording always (it would
  cost every `go test`).

### Not taken

- **Finding elements through a new private-state accessor on `Context`** (a public map of laid-out
  rectangles): it would add a field read only by tests (against R11). The screen-reader tree is
  already public, already clipped, and already labelled.
- **Methods that take `*testing.T` and fail the test themselves**: shorter tests, but the author
  asked for errors with codes and no panics, and a returned error also works outside `go test`.
- **Real time with `time.Sleep`**: slow and flaky in CI.
- **A PNG of a frame**: no software painter yet (see Out of scope).

## Tasks

- [x] T1 — The window, its options, the fixed measurer and the clock: `New`, `Frame`, `Frames`, `Wait` · `uitest/window.go`, `options.go`, `measurer.go`, `clock.go`, `errors.go` · R1 R5 R11 · tests: `TestAFrameRunsTheInterfaceOnceLikeABackend`, `TestWaitMovesTheClockAndRunsTheFramesInIt`, `TestEachErrorCarriesItsCode` (for the codes so far), a `go list -deps` test that `ui` does not import `uitest`
- [x] T2 — Finding what was drawn: `Element`, `ElementIn`, `Texts`, `DrawList`; `FocusedID` on `ui.Context`; check that switching the screen reader on does not change the layout · `uitest/find.go`, `ui/focus.go` (no `Hidden` was needed: the tree already leaves out what is behind a dialog) · R6 R7 R8 · tests: `TestAnElementIsFoundByItsLabel`, `TestAnElementScrolledAwayIsNotVisible`, `TestTheTextsOfAFrameComeInDrawingOrder`, `TestTheDrawListOfTheLastFrameCanBeRead`, `TestAMissingLabelGivesElementNotFound`, `TestTheScreenReaderTreeDoesNotChangeTheLayout`
- [x] T3 — The pointer · `uitest/pointer.go` · R2 R7 · tests: `TestAClickOnAButtonPressesIt`, `TestADragMovesASlider`, `TestScrollingShowsWhatWasBelow`
- [x] T4 — The keyboard and typed text · `uitest/keyboard.go` · R3 · tests: `TestTypedTextReachesTheFocusedField`, `TestTabMovesTheFocusToTheNextControl`
- [x] T5 — Touch, with the clock for long presses · `uitest/touch.go` · R4 R5 · tests: `TestATapPressesAButton`, `TestALongPressOpensTheMenu`, `TestALongPressNeedsItsTime`
- [ ] T6 — Go UI's own tests move to `uitest` where it does the same; the private harness keeps only what reads private state · `*_test.go` at the root · R9 · `go vet ./...` and `go test ./...` pass with the same number of tests or more
- [ ] T7 — The example and the guide: `ExampleWindow` (a lighthouse keeper's log: click a button, type a line, find it), `docs/4-guides/11-testing-an-interface.md`, a reference page `docs/6-reference/3-uitest.md` made by `generate.py`, links from the overview · `uitest/example_test.go`, `docs/` · R10 · the example's `// Output:` is checked by `go test`; the author reads the guide
- [ ] T8 — Recording: the `Test(t)` option, frames and errors kept only while `UITEST_RECORD` is set, one JSON file per test · `uitest/recording.go`, `options.go` · R12 · tests: `TestAWindowWithoutRecordingKeepsNothing`, `TestARecordedTestWritesItsFramesAndErrors`
- [ ] T9 — The results viewer: `cmd/uitest` runs the tests and shows them, logic apart from look; a guide section on it · `cmd/uitest/main.go`, `results.go`, `view.go`, `docs/4-guides/11-testing-an-interface.md` · R12 · tests (with `uitest`): `TestFailedTestsComeFirst`, `TestTheArrowsStepThroughTheFrames`, `TestTheFoundElementIsHighlighted`; the window's look is checked by the author on a display
