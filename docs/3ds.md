# 🎮 On the Nintendo 3DS

**Status:** a first step (October 2026). Built with a TinyGo that has a `nintendo3ds` target, a
go-ui program runs on the console as homebrew, from the Homebrew Launcher: `window.Run` opens the
console's two screens, drawn through go-renderer's `ctr` (citro3d). Checked in the Azahar emulator
with a large Material program (go-ui-android's pieces); not yet on a console. The same drawing,
with a loop of its own, runs on a New 3DS XL in go-3ds's explorer.

<p align="center">
  <img alt="The emulator's two screens: an island village at nine in the evening above, and below a dark go-ui panel with tabs and the time of day" src="images/3ds-explorer.png" width="320">
</p>

## Two screens

| | Usual | After SELECT |
|---|---|---|
| Top screen (400 × 240, in 3D with the slider) | The scene (`Options.Scene`) | The interface, to be read |
| Bottom screen (320 × 240, touch) | The interface: the touch screen is the pointer | The scene, which the finger turns and picks in |

A program written for a phone works unchanged: the interface is laid out for the screen it is on
(320 or 400 pixels wide), and the scene gets the input the interface did not claim. With the
screens swapped the interface only gets the keys, since the top screen cannot be touched. START
ends the program.

| Button | Is |
|---|---|
| Cross | The arrow keys |
| L, R | Minus and plus (closer and farther, for most cameras) |
| A, B | Enter and Escape; B is also the system's back |
| Touch screen | The pointer: dragging is dragging with a mouse |

## What the console has not

`Notify`, `Share`, `OpenLink`, `Feel` and the rest do nothing, and `SystemPrefersDark` is false.
Preferences last only the run: TinyGo's `os` cannot reach the SD card yet. Films do not play.

## How it runs

`window/nintendo3ds.go` reads the controls into the input state each frame (`ctr.Controls`),
builds the interface, asks the program for its scene, and draws both in one frame:
`SceneRenderer.PaintStereo` (or `PaintOn` the bottom screen) and `FlatRenderer.PaintOn`. It draws
every frame, as fast as the screen shows them, and yields to the program's other goroutines after
each: TinyGo's scheduler there is cooperative, and waiting for the screen happens in C.

## Building

```bash
tinygo build -target=nintendo3ds -o program.3dsx ./cmd/program
```

Copy the `.3dsx` to `/3ds/` on the SD card. A program that reads much data at start should keep it
small: the console has about 86 MB for Go, and TinyGo's collector neither moves what it keeps nor
knows which words on a stack are pointers, so garbage costs more there than on a phone.
