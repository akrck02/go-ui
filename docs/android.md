# Plan: go-ui on Android

**Status:** a plan, not started. It explains what stops go-ui applications from running on Android
today, what has to be built, what has to be installed, and in which order, so the work can be
picked up later.

## Where things stand

| Piece | Runs on Android? | Why |
|---|---|---|
| go-ui (context, layout, controls, theme) | **Yes** | Plain Go with no cgo: it only reads an `input.State` and fills a `draw2d.List` |
| go-renderer `draw2d`, `input`, the overlay painter and measurer | **Yes** | Plain Go: shapes, fonts and the atlas need no graphics context |
| go-renderer `opengl` (window, flat renderer, scene renderer) | **No** | GLFW and desktop OpenGL 4.1 core; Android has neither: it has OpenGL ES 3 through EGL, inside an activity |
| go-ui `window` | **No** | Built on go-renderer's GLFW loop |

So the interface itself is portable; what is missing is **a window and a renderer for Android**,
and **touch** as an input.

## Scope

**First version:** applications that are only an interface (like `cmd/app`), on Android 8.0
(API 26) or newer, on arm64 devices and the arm64 emulator.

**Not in the first version:**

- The 3D scene renderer (shadows, water, levels of detail) on OpenGL ES: much larger, done later.
- The on-screen keyboard and text input: it arrives with the text field.
- iOS: the same route works there (see *Later*), but it is not tried first.

## How it fits together

```text
your program            build(context) every frame, the same code as on the desktop
go-ui                   unchanged: context, layout, controls, theme
go-ui/window (mobile)   NEW: the Android app loop: lifecycle, surface, density, touch → input.State
go-renderer/gles        NEW: FlatRenderer for OpenGL ES 3 (the overlay's atlas and quads)
go-renderer/input       touch as a pointer, and the kind of pointer (mouse or finger)
golang.org/x/mobile     the activity, the EGL surface and the events, built into an APK by gomobile
```

`golang.org/x/mobile/app` gives a Go program an Android activity with an OpenGL ES context and a
stream of events (lifecycle, size, paint, touch, key). `gomobile build` turns the program into an
APK. No Java or Kotlin is written.

## The pieces

### 1. An OpenGL ES flat renderer (go-renderer, `feature/` branch)

The overlay painter already produces everything as rectangles of one atlas (`overlay.Quad`). On
the desktop, `opengl.FlatRenderer` uploads the atlas and draws the rectangles in one call. The
Android version does the same with OpenGL ES 3:

- A new package (for example `gles`) with `FlatRenderer.Draw(list, background, width, height,
  scale)`, the same signature as the desktop one.
- The same two shaders with `#version 300 es` and a precision line; the atlas as a single-channel
  `R8` texture, as today.
- It talks to OpenGL ES through `golang.org/x/mobile/gl` (a `gl.Context` from the app's lifecycle)
  rather than `go-gl/gl`. About 150 lines; the painter, the atlas and the fonts are reused as they
  are.

### 2. Touch as a pointer (go-renderer `input`, go-ui)

- The first finger down becomes the **pointer**: it moves the cursor and presses and releases the
  left button, so every control works unchanged. A second finger is ignored for now.
- `input.State` learns the **kind of pointer** (`PointerMouse` or `PointerTouch`). With a finger
  there is no hover: go-ui does not show hovered looks or tooltips, and the pressed look shows
  while the finger is down.
- Dragging a finger on a list will scroll it once `ScrollArea` exists; until then a drag only
  moves sliders and panels.
- Long press as a right click, and pinch, are left for later.

### 3. The Android window (go-ui `window`, built only for Android)

`window.Run` keeps its signature; on Android (`//go:build android`) it runs the x/mobile loop:

| Event | What happens |
|---|---|
| `lifecycle` (visible / not visible) | create the GL context and the renderer, or let them go; the app stops drawing in the background |
| `size` | window size in layout pixels and the screen's density (`PixelsPerPt`), which becomes the drawing scale; the insets of the status bar and the gesture bar |
| `touch` | the pointer, as above |
| `paint` | one interface step (`Begin`, `build`, `End`), draw, publish |

Like the desktop window, it **draws only when something happened** (a touch, a resize, or a
control asking for another frame with `RequestFrame`), which is what a phone's battery needs.

### 4. A touch theme and narrow layouts (go-ui)

- `TouchTheme()`: the same colours with controls 44–48 px tall, 15 px text and wider gaps, so a
  finger hits what it aims at.
- `Context.Width()` and `Context.Height()`: the window's size, so `build` can choose a layout. In
  immediate mode this is an `if`: on a narrow screen the register shows the list, and the
  inspector opens over it when a row is touched; on a tablet it keeps three columns.
- The insets from the window are padding for the top and bottom bars.

### 5. Building and installing

```bash
gomobile build -target android/arm64 -androidapi 26 -o register.apk ./cmd/app
adb install -r register.apk
```

The application's name and icon come from an `AndroidManifest.xml` and an `assets/` folder next to
the program, which gomobile picks up.

## What has to be installed

None of this is on the development machine yet (only `adb`). About 6–8 GB:

| Tool | Why | How |
|---|---|---|
| A JDK (17) | The Android build tools need it | Homebrew `openjdk@17` |
| Android command-line tools | `sdkmanager` and `avdmanager` | Homebrew cask `android-commandlinetools`, or Android Studio |
| SDK platform 34, platform tools, build tools | Compile and install APKs | `sdkmanager "platforms;android-34" "platform-tools" "build-tools;34.0.0"` |
| NDK | gomobile compiles Go for Android with it | `sdkmanager "ndk;26.3.11579264"` (or the current long-term one) |
| Emulator and an arm64 system image | Run without a phone | `sdkmanager "emulator" "system-images;android-34;google_apis;arm64-v8a"` then `avdmanager create avd` |
| gomobile | Builds the APK | `go install golang.org/x/mobile/cmd/gomobile@latest`, then `gomobile init` |

A phone with USB debugging enabled can replace the emulator (and is the better test of touch and
density).

## Iterations

| # | What | Done when | Effort |
|---|---|---|---|
| 0 | Install the tools; build and run x/mobile's own GL example | It shows on the emulator | ½ day |
| 1 | `gles.FlatRenderer`; the Android `window` with lifecycle, size and paint | A one-panel go-ui program shows crisp text on the emulator, captured with `adb exec-out screencap` | 1 day |
| 2 | Touch as a pointer, no hover with a finger, drawing on demand (already done on the desktop: frame requests and waking are in place, the Android loop only has to honour them) | Buttons, rows and filters of `cmd/app` react to taps; the app is idle (no frames) when untouched | 1 day |
| 3 | `TouchTheme`, `Width`/`Height`, insets, a narrow layout for `cmd/app`, name and icon | The register is comfortable on a phone and a tablet, in light and dark | 1 day |

About **3–4 days** in all. Each step keeps the desktop working, with tests and captures.

## Testing

- **Without a device:** unit tests for the touch mapping (a tap is a click; no hover with a
  finger), for drawing on demand, and for the narrow layout (which boxes exist below a width).
- **On the emulator:** captures through `adb exec-out screencap -p` at each iteration, taps through
  `adb shell input tap x y`, and checks that the app stops drawing when idle.
- **On a phone:** density, touch accuracy, the gesture bar, and how it feels.

## Later

- **The on-screen keyboard**, with the text field: recent versions of x/mobile can show and hide
  the virtual keyboard; if not, a few lines of JNI do it. Characters then arrive as key events.
- **The 3D scene on OpenGL ES 3**: shaders ported to ES, shadow maps as depth textures (ES 3 has
  them), smaller defaults for phones. A separate, larger piece of work.
- **iOS**: the same x/mobile route builds for iOS with OpenGL ES (deprecated there, but working);
  Metal would be the long-term answer.

## Risks

- **x/mobile changes slowly** and some of it is experimental. If it gets in the way, the fallback
  is a small native activity of our own (cgo, EGL and the NDK's `android_native_app_glue`): more
  code, more control.
- **Text sharpness at phone densities** (2.6× to 3.5×, not whole numbers): the painter rounds to
  pixels and rasterises at the real size, which should hold; checked on the first captures.
- **Battery**: drawing on demand is a must on a phone, not an extra.

## Decisions to take before starting

1. Emulator, a real phone, or both.
2. The lowest Android version (API 26, Android 8.0, is proposed).
3. ~~Whether drawing on demand also goes to the desktop window~~: done, the desktop window draws
   on demand by default.
