# 📚 go-ui documentation

| | Page | What is in it |
|---|---|---|
| 🚀 | [Getting started](getting-started.md) | A frame in immediate mode, ids and scopes, the examples and their flags, how it is drawn. |
| 📐 | [Layout](layout.md) | Boxes, `Fit`, `Grow`, `Share` and the rest, and how sizes and positions are worked out. |
| 🎛️ | [Controls](controls.md) | Text and icons, buttons, choices, lists, rows and grids, chips, tooltips, painting and motion. |
| 🗂️ | [Panels and layers](panels-and-layers.md) | Panels, the screen, editing a scene, popups, menus, sheets, modals and overlays. |
| 📋 | [Tables, dragging and resizing](tables-and-dragging.md) | Grouped tables, drag and drop onto the scene, boxes resized by their edge. |
| ⌨️ | [Text fields](text-fields.md) | Typing, selecting, the clipboard, undo, Tab, focus, and a phone's keyboard and handles. |
| 🖱️ | [Mouse, keyboard and touch](input.md) | Which control gets the pointer, keeping the camera out of the way, fingers and gestures. |
| 🎨 | [Themes](themes.md) | The tokens every control reads, light and dark. |
| 🗣️ | [Screen readers](screen-readers.md) | How the boxes are read, describing a control of your own, what the screen reader asks. |
| 🖥️ | [A window of its own](desktop.md) | Interfaces without a scene, drawing only when something happens, Metal or OpenGL. |
| 🌐 | [In a web browser](web.md) | The same programs built for WebAssembly and drawn through WebGL2. |
| 📱 | [On Android](android.md) | The window, touch, building a package, what was tried and what is next. |
| ⚡ | [Performance and battery audit](performance-audit.md) | What an idle and a busy interface cost, measured, and what to improve in order. |

New here? Start with the [project README](../README.md).

## 🚧 Not done yet

- Text does not wrap, except in a box with `Wrap`; there is one typeface, so no monospaced table
  column ([layout](layout.md), [tables](tables-and-dragging.md#tables)).
- On the desktop: notifications, films (`ErrVideoUnsupported`) and the system's dark theme
  ([Android](android.md)).
- In a browser: only Chromium on macOS was tried; other browsers, phones and tablets, and the
  on-screen keyboard are not ([web](web.md)).
- On Android: rotating the screen, other phones and tablets, the bars' icons following the
  program's theme, two fingers, and shadows, sky and waves in the scene ([Android](android.md#next)).
