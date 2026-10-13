# go-ui on Android

**Status:** a first step (October 2026). A go-ui program opens and draws on an Android 16
emulator, and reacts to taps and drags; the same code runs on the desktop and in a browser. This
page started as a plan (what to build, install and try, in which order); *What was built* says how
it turned out, and *Next* what is left of the plan.

![cmd/mobile on the Android emulator: a list of lighthouses, a switch, a slider and a button](images/android-emulator.jpg)

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
| The loop: lifecycle (create the renderer when visible, let it go when hidden), size, touch, keys, paint | `window/android.go` |
| go-renderer's `gles.Functions` over `golang.org/x/mobile/gl` | `window/android_gl.go` |
| Fingers as the pointer, and the screen's density as the drawing scale | `window/touch.go`, with tests |
| A phone-sized example, the same on desktop, web and Android | `cmd/mobile` |
| Building an installable package | `cmd/mobile/build-apk.sh`, `cmd/mobile/android/AndroidManifest.xml` |

**Touch.** The first finger down is the pointer: it moves the cursor and presses the left button,
lifting it releases the button, so every control works unchanged. Other fingers are ignored until
it lifts. Positions are divided by the screen's density, so layout pixels are Android's
density-independent pixels (1/160 inch) and an interface is the same size on every phone.

**Drawing on demand.** As on the desktop and in the browser, the window draws only when something
happens: a touch, a size change, a frame a control asks for (`RequestFrame`) or `window.Wake`. A
frame that would look like the one shown is not drawn. Idle on the emulator, the program used
about 1 % of a core (the Go runtime and x/mobile's threads; no frames are drawn).

**`cmd/mobile`** enlarges the theme for fingers (controls 44 pixels tall, 15-pixel text) in its
own code, until go-ui has a touch theme of its own.

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

On the emulator (Android 16, arm64, 1080 × 2400 at 420 dots per inch): the window opens and draws
with sharp text at a density of 2.625; tapping a row selects it, tapping the switch turns it off,
dragging the slider moves it; idle, nothing is drawn.

Not yet: a real phone (density, touch accuracy, how it feels), rotating the screen, going to the
background and back, and the 3D scene on Android.

## Next

| # | What | Why |
|---|---|---|
| 1 | **Insets.** The window draws under the status bar and the gesture bar (see the title in the screenshot). This x/mobile does not report their size; a few lines of JNI, or a newer x/mobile, would, and the window would pad the interface by them | Nothing important may sit under the system's bars |
| 2 | **A touch theme** (`TouchTheme()`), `Context.Width()` and `Height()` for narrow layouts, no hover looks with a finger | Comfortable on phones and tablets, in light and dark |
| 3 | **The on-screen keyboard** for text fields | Typing works only with a hardware keyboard today |
| 4 | **The 3D scene**: `gles.SceneRenderer` already draws in browsers; it needs a try on a phone | Interfaces over a scene, as on the desktop |
| 5 | **Long press, two fingers** | A right click, pinch to zoom |

## Risks

- **x/mobile on Android is OpenGL ES 2 only** (its bindings are compiled against the ES 2
  headers, and it asks for an ES 2 context). go-renderer's `gles` keeps to OpenGL ES 2 for that
  reason (see go-renderer's `docs/android.md`). If OpenGL ES 3 or instancing become necessary, the
  fallback is a small native activity of our own (cgo, EGL and the NDK's
  `android_native_app_glue`): more code, more control.
- **x/mobile changes slowly**, and pinning an older version keeps it working with Go 1.25.
- **Battery**: drawing on demand is a must on a phone, and it is in place.
