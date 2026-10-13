# go-ui on Android

**Status:** a first step (October 2026). A go-ui program opens and draws on an Android 16
emulator, with a 3D scene under its interface, and reacts to taps and drags; the same code runs on
the desktop and in a browser. The example looks like Android's own apps (Material Design 3). This
page started as a plan (what to build, install and try, in which order); *What was built* says how
it turned out, and *Next* what is left.

![cmd/mobile on the Android emulator: the list of lighthouses, a lighthouse's details, and the island at three in the afternoon](images/android-emulator.jpg)

## What was built

```text
your program            build(context) every frame: the same code as on the desktop
go-ui                   unchanged: context, layout, controls, theme
go-ui/window (android)  the activity's loop: lifecycle, size and density, touch → input.State, paint
                        and gles.Functions over golang.org/x/mobile/gl
go-renderer/gles        FlatRenderer: the overlay's atlas and quads, shared with the browser
golang.org/x/mobile     the activity, the EGL surface and the events; gomobile compiles
```

| Piece | Where |
|---|---|
| The loop: lifecycle (create the renderers when visible, let them go when hidden), size, touch, keys, paint | `window/android.go` |
| go-renderer's `gles.Functions` over `golang.org/x/mobile/gl` | `window/android_gl.go` |
| What the system's bars cover, and whether the phone is in its dark theme, asked through JNI | `window/android_system.go` |
| Fingers as the pointer, and the screen's density as the drawing scale | `window/touch.go`, with tests |
| A 3D scene under the interface (`Options.Scene`), on every platform | `window/window.go` and each platform's file |
| An example that looks like an Android app, the same on desktop, web and Android | `cmd/mobile` |
| Building an installable package, with a window theme | `cmd/mobile/build-apk.sh`, `cmd/mobile/android/` |

**Touch.** The first finger down is the pointer: it moves the cursor and presses the left button,
lifting it releases the button, so every control works unchanged. Other fingers are ignored until
it lifts. Positions are divided by the screen's density, so layout pixels are Android's
density-independent pixels (1/160 inch) and an interface is the same size on every phone.

**Drawing on demand.** As on the desktop and in the browser, the window draws only when something
happens: a touch, a size change, a frame a control asks for (`RequestFrame`) or `window.Wake`. A
frame that would look like the one shown is not drawn. Idle on the emulator, the program used
about 0.5–1 % of a core (the Go runtime and x/mobile's threads; no frames are drawn).

**The system's bars.** The window covers the whole screen, under see-through bars, as Android's
own apps do. `Context.Insets()` says how much of each edge the status bar and the gesture bar
cover, in layout pixels, so a program paints its own bars' backgrounds under them and keeps its
controls clear. x/mobile does not report them, so the window asks Android through JNI
(`getRootWindowInsets`, through x/mobile's `RunOnJVM`); before the window is shown it falls back on
the bars' standard heights. The window theme in `cmd/mobile/android/res` makes the bars
see-through, with dark icons in the system's light theme and light ones in its dark theme.

**The dark theme.** `window.SystemPrefersDark()` tells whether the phone (or the browser) is set to
a dark theme, so a program can follow it; `cmd/mobile` asks at every frame until a theme is chosen
in its settings.

**A scene under the interface.** `Options.Scene` returns, at every step, a scene and a camera to
draw behind the interface, or nil for none. It gets the input the interface did not claim, so a
drag on an empty part of the screen turns the camera while a drag on a slider moves the slider.
The scene is drawn by go-renderer's `gles.SceneRenderer` on Android and in browsers, and its
OpenGL scene renderer on the desktop.

**`cmd/mobile`** is built the way Android's own apps look, out of go-ui boxes, texts and
clickables (`material.go`): a top app bar and a navigation bar painted under the system's bars,
list items with a round icon and a Material switch, filter chips, pill-shaped buttons, cards and
floating buttons, in Material Design 3's baseline colours, light and dark, and in Roboto, Android's
typeface (embedded under the SIL Open Font License, in `cmd/mobile/fonts`). Its three destinations
are a list of lighthouses with their details, a map with the 3D island (drag to turn, buttons to
come closer, a slider for the time of day) and settings (the dark theme, the screen's size and
insets).

## Building and trying it

```bash
go install golang.org/x/mobile/cmd/gomobile@v0.0.0-20250711185624-d5bb5ecc55c0   # the version in go.mod
export ANDROID_HOME=…/android-sdk ANDROID_NDK_HOME=$ANDROID_HOME/ndk/<version>
cmd/mobile/build-apk.sh mobile.apk
adb install -r mobile.apk
adb shell am start -n io.github.akrck02.goui.mobile/org.golang.app.GoNativeActivity
adb exec-out screencap -p > screen.png          # what the screen shows
adb shell input tap 230 330                     # a tap, in screen pixels
```

**Why a script and not just `gomobile build`.** gomobile writes a manifest that targets Android 4.1
(SDK 16), and Android 14 and later refuse to install such a package; it also refuses a manifest
that says otherwise. `build-apk.sh` lets gomobile compile, takes the native library and the
activity's code out of its package and packs them again, with the SDK's own tools (`aapt2`,
`zipalign`, `apksigner`), under `cmd/mobile/android/AndroidManifest.xml`. It signs with the debug
key: for trying, not for publishing.

**Which gomobile.** The newest `golang.org/x/mobile` needs Go 1.26; these repositories are on
Go 1.25, so `go.mod` holds a version from July 2025 and gomobile must be the same one.

**Installed for this:** the NDK (r29) through `sdkmanager`, and gomobile. The SDK, its build tools,
the emulator and an arm64 Android 16 image were already there.

## What was checked, and what was not

On the emulator (Android 16, arm64, 1080 × 2400 at 420 dots per inch): the window draws with
sharp text at a density of 2.625, clear of the status bar and the gesture bar (49 and 24
layout pixels); a tap on a row opens its details and a tap on its switch flips it without
opening it; the back arrow, the chips, the buttons, the navigation bar and the sliders work; the
island draws at any time of day, turns with a drag and comes closer with the buttons; the dark
theme switches every screen. Idle, the program uses about 0.5 % of a core; the map takes 16 draw
calls a frame.

On a phone (Pixel 6 Pro, Android 16, Mali-G78, 1080 × 2340): the same, in the system's dark theme
at start, and after going to the background and back. That phone gives an OpenGL ES 3.2 context
although x/mobile asks for version 2. It found two faults the emulator hid, now fixed: the 3D map
lost its ground after coming back from the background (meshes kept the buffers of the destroyed
graphics context), and the program read the system's theme before Android gave it.

Not yet: rotating the screen, and how it feels on other phones and tablets.

## Next

| # | What | Why |
|---|---|---|
| 1 | **The bars' icons follow the program's theme**, not only the system's: it needs a call on Android's interface thread, which x/mobile does not offer | Dark icons on a dark app (when the app's theme differs from the phone's) |
| 2 | **Material pieces in go-ui itself**: a touch theme, the switch, list items, the navigation bar, now in the example | Every program gets them, not only the example |
| 3 | **The on-screen keyboard** for text fields | Typing works only with a hardware keyboard today |
| 4 | **Long press, two fingers, the back gesture** | A right click, pinch to zoom the map, back from a lighthouse's details |
| 5 | **Shadows, sky and waves** on phones, as on the desktop | The scene is go-renderer's simple one on phones and in browsers |

## Risks

- **x/mobile on Android is OpenGL ES 2 only** (its bindings are compiled against the ES 2
  headers, and it asks for an ES 2 context). go-renderer's `gles` keeps to OpenGL ES 2 for that
  reason (see go-renderer's `docs/android.md`). If OpenGL ES 3 or instancing become necessary, the
  fallback is a small native activity of our own (cgo, EGL and the NDK's
  `android_native_app_glue`): more code, more control.
- **x/mobile changes slowly**, and pinning an older version keeps it working with Go 1.25.
- **Battery**: drawing on demand is a must on a phone, and it is in place.
