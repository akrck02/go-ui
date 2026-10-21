# 3. Driving an interface: automatic modes, and tests on the device

> Status: building · Author's approval: proposal ☑ (2026-10-21, "sigue, hazlo") plan ☑ (2026-10-21,
> the same words: the author asked to go ahead without stopping)

## Why

`uitest` drives an interface with no display: it finds what was drawn by its name ("Add an
action", "Pad A") and gives it input, frame by frame, on a clock of its own. That idea is useful
far beyond `go test`, but today it only works in a test window. A real Go UI window, on a
computer, a phone or a 3DS, cannot be driven:

- **No automatic modes.** A program cannot play a demo of itself (an attract mode), guide a
  person through a tutorial with a pointer that shows *press here* on the real control, or replay
  what someone did to reproduce a bug they reported.
- **No tests on the device.** `uitest` proves the interface's logic on the computer, but not what
  only a device does: Go UI for Android's bridge to Java, the system keyboard with its
  suggestions, TalkBack, the screen's insets; Go Renderer drawing on the 3DS's GPU and its
  buttons. Today those are checked by hand, with captures from the emulators.
- **Two copies of one idea.** If a program wrote its own driver, it would copy what `uitest`
  already does, and the two would drift apart.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A window can be **driven**: a program gives `window.Options` a driver, which adds input to each frame before the interface runs (a pointer, keys, fingers, typed text), on every target: desktop, web, Android and the 3DS | test with a made-up driver in a window loop; a demo run on each target |
| R2 | A **script** says what to do, by the names people see: move to, click, tap, hold, drag, type, press a key, wait, *wait until* an element shows; each step finds its element in the live interface as `uitest` does | tests of each step against a made-up interface |
| R3 | `uitest` runs the same scripts and shares the code that finds elements and makes input: one implementation, two places to run it | `uitest`'s tests pass unchanged; no finding code left in two packages |
| R4 | A script can be **recorded** from what a person does in a window, and saved as a file a person can read and edit; played back, it does the same | test: record a session, play it back, the interface ends in the same state |
| R5 | A driven window can **show** what the driver does: the pointer, the fingers, the element a step found, so a demo or a tutorial is seen, not only run | capture of a tutorial step |
| R6 | A script run on a device **reports** each step's result with the error codes `uitest` uses, as lines a test runner reads (Android's log, the 3DS's console, standard output) | test of the report's format; a script run in the Android emulator and in Azahar |
| R7 | Go UI for Android gains a test that installs an app in the emulator, runs a script on it and reads the report, so its own layer (the Java bridge, the keyboard, TalkBack's tree, the insets) is proven automatically | the test passes in the emulator |
| R8 | A window with no driver costs nothing: no allocation and no work each frame; programs that do not use it are unchanged | benchmark of the window loop with and without; the samples unchanged |
| R9 | A person always wins: any real input while a script plays stops it, so an automatic mode never fights the hand on the device | test: a key pressed during a script stops it |

## Out of scope

- Automating other programs or the operating system (UIAutomator, accessibility services): this
  drives Go UI's own windows only.
- Comparing pictures pixel by pixel: it needs Go Renderer to paint with no display, which is its
  own proposal.
- Driving a window from another computer over the network.
- A script language with branches and loops: a script is a list of steps; a program that needs
  logic writes Go.

## Decided questions

1. **The script file's format: a line per step**, plain text a person reads and writes, with `#`
   comments. Each line is parsed into a step; an error names its line.
2. **The shared code lives in a new package, `uidrive`**: it finds elements by name in the live
   interface, makes input for a step, and runs a script frame by frame on a clock. `uitest` is
   rebuilt on it.
3. **R7 lives in Go UI for Android**, which owns the layer it proves. It is a later task there,
   not part of this change (see Tasks).

## Design

**A driver is a hook in the window loop.** `uidrive.Driver` is an interface of two methods:
`Drive(state, context)` runs after the backend gathered the frame's real input and before the
interface's `Begin`, and adds input to the state; `Show(drawing)` runs after `End`, and may draw
over the interface. `window.Options.Driver` takes one. Desktop, web and Android share
`interfaceApplication`'s `Update`, and the 3DS has its own `consoleRun.step`: both call the same
two lines, behind `if driver != nil`, so a window without one does what it did (R1, R8).

**`uidrive`** (new package, pure Go, so it builds for TinyGo and WebAssembly):

| File | What it holds |
|---|---|
| `errors.go` | `Code`, `Error`, `CodeOf`; the element codes moved from `uitest`, which keep their stable names (`uitest.element-not-found`…) since programs and people check them; new `uidrive.script-not-read` (with the line) and `uidrive.script-stopped` |
| `find.go` | `Element`, `Find`, `FindIn`, `FindWithin`, `FindAt`, `Texts`: what `uitest/find.go` did, over a `*ui.Context` (its screen reader's nodes and its drawing) |
| `hand.go` | `Hand`: the pointer, the fingers and typing made as input on an `input.State` (move, press, release, touch, type, keys, let go of all) |
| `script.go` | `Script`, `Step`, `Action`; `Parse` reads the line format and `Script.String` writes it |
| `player.go` | `Player`: runs a script one frame at a time on the state's clock; a `Driver` |
| `report.go` | one line per step, `uidrive: step 3 failed uitest.element-not-found: click "Remove"`, to an `io.Writer` the program gives |
| `recorder.go` | `Recorder`: a `Driver` that adds nothing and writes down what a person does, as a `Script` |
| `show.go` | the overlay: the pointer, the fingers and the element found, drawn by `Player.Show` when asked |

`uitest` keeps its whole API. `Element`, `Error` and `Code` become aliases of `uidrive`'s; its
finders call `uidrive.Find…`; its pointer, touches and keys go through a `uidrive.Hand`. It gains
a `Driver` option (the window calls it each frame, as a real one) and `Window.Play(script)`, which
runs a script with a sixtieth of a second per frame on the test clock (R3). `ui.Context` gains
`Measurer()`, which finding a text needs. On Android the window keeps the screen reader's nodes on
while driven, since a driver finds elements through them.

**The script format**, one step per line:

```text
# a lighthouse keeper writes the day's log
click "Add an entry"
type "Fog at dawn"
press Enter
wait 0.5s
until "Saved"            # waits up to 5s, or: until "Saved" 10s
tap "Light the lamp"
hold "Lens" 1s           # a long press
drag "Boat" 120 0        # by a distance in layout pixels; or: drag "Boat" to "Harbour"
scroll "Log" -3
double-click "Entry"
move-to "Help"
click-at 40 300          # a place, when what is there has no name
press Shift+Tab
```

**The person wins (R9):** a `Player` that finds input in the state before adding its own
(`input.State.HasInput`) stops, lets go of what it holds, and ends with `uidrive.script-stopped`.

**Alternatives not taken.** JSON scripts: diffable, but noisy to write by hand. Keeping the
finding code in `uitest` and importing it from windows: `uitest` imports `testing`, which has no
place in a program on a console. A driver as a `func`: an interface lets one value both drive and
show, and a nil check costs the same. A second, input-replaying recorder (raw events with times):
it would break when the layout moves; names survive a new layout.

## Tasks

- [x] T1 — move finding into `uidrive` (`Element`, `Find…`, errors); `ui.Context.Measurer`; `uitest`
  finds through it · `uidrive/errors.go`, `uidrive/find.go`, `ui.go`, `uitest/find.go`,
  `uitest/errors.go` · serves R3 · test: `TestAButtonIsFoundByItsLabelInTheLiveInterface`, and
  `uitest`'s tests unchanged
- [x] T2 — `uidrive.Hand` makes the pointer, touch and key input; `uitest` acts through it ·
  `uidrive/hand.go`, `uitest/pointer.go`, `uitest/touch.go`, `uitest/keyboard.go` · serves R3 ·
  test: `TestTheFirstFingerDownAlsoPressesThePointer`
- [x] T3 — the script format: `Parse` and `Script.String` · `uidrive/script.go` · serves R2 · test:
  `TestAScriptWrittenAndReadAgainIsTheSame`, `TestAnUnknownStepIsAnErrorWithItsLine`
- [x] T4 — `Player` runs each step frame by frame; `uitest.Driver` and `Window.Play` · `uidrive/
  player.go`, `uitest/window.go`, `uitest/play.go` · serves R2, R3 · test: one per step against a
  made-up interface (`TestAClickStepPressesTheButtonByItsName`…)
- [ ] T5 — the report, a line per step with its code · `uidrive/report.go` · serves R6 · test:
  `TestTheReportSaysEachStepAndTheCodeOfAFailure`
- [ ] T6 — a person's input stops a script · `uidrive/player.go` · serves R9 · test:
  `TestAKeyPressedDuringAScriptStopsIt`
- [ ] T7 — `window.Options.Driver` on every target, at no cost without one · `window/window.go`,
  `window/nintendo3ds.go`, `window/android.go`, `window/drive.go` · serves R1, R8 · test:
  `TestADriverAddsInputBeforeTheInterfaceRuns`, `TestAWindowWithoutADriverAllocatesNothingForIt`,
  `BenchmarkAWindowStep…`; guide page `docs/4-guides/12-driving-a-window.md`
- [ ] T8 — `Recorder` writes down a person's session as a script · `uidrive/recorder.go` · serves R4
  · test: `TestARecordedSessionPlayedBackEndsInTheSameState`
- [ ] T9 — the overlay: pointer, fingers and the element found · `uidrive/show.go`, `cmd/example` ·
  serves R5 · test: `TestAShownPlayerDrawsThePointerAndTheElementItFound`; a capture
- [ ] Later, in **go-ui-android** (not this change) — R7: a test that installs an application in
  the emulator, plays a script on it with its report going to Android's log, and reads the log
  for `uidrive: script passed`

## Result

Not built yet: waiting for the author's approval of this proposal.
