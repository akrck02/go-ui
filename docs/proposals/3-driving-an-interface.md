# 3. Driving an interface: automatic modes, and tests on the device

> Status: proposed · Author's approval: proposal ☐ plan ☐

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

## Open questions for the author

1. **The script file's format.** Plain JSON, readable and diffable, or a line per step
   (`click "Add an action"`), shorter to write by hand?
2. **Where the shared code lives.** A new package (for example `uidrive`), used by `uitest` and by
   windows; or inside `uitest`, renamed for both uses?
3. **R7's place.** In Go UI for Android, which owns that layer, or in Go UI, which owns the
   driver?

## Result

Not built yet: waiting for the author's approval of this proposal.
