# 🚀 Getting started

Go UI draws interfaces over a Go Renderer view. This page explains how a frame works, shows the
examples that come with it, and how what you describe ends up on the screen.

## A frame

<p align="center">
  <img alt="Your program describes the interface every frame → Go UI: Begin and the controls add boxes, End lays them out and makes a draw2d list → Go Renderer paints it in one draw call; the mouse, keys and touches come back for the next frame" src="../images/how-it-works.svg" width="880">
</p>

In **immediate mode** there are no button objects to create, keep and update. Each update step,
the program calls the controls again, with its own variables:

```go
interfaceContext.Begin(state) // the step's mouse and keyboard
interfaceContext.Panel("Lighthouse", func() {
	interfaceContext.Heading(light.Name)
	if interfaceContext.PrimaryButton("Switch on") { // true on the step it is clicked
		light.On = true
	}
})
interfaceContext.End() // lay out, paint, claim the mouse
```

Showing or hiding something is an `if`; a list is a `for`. The interface holds none of your data,
so it can never show a stale value. What it does remember between frames is small and keyed by
each control's **id**: which control is being pressed, where a panel was dragged, whether it is
folded.

Call `Begin` and `End` in every update step (the function given to `viewer.Run`), and set
`world.Overlay.Drawing = interfaceContext.Drawing()` once, so Go Renderer paints the result.

### Ids and scopes

An id is a hash of the control's label and of the **scopes** around it. Panels open a scope with
their title, so two "Apply" buttons in different panels do not mix. Inside a loop, open a scope per
item with `Push(name)` and `Pop()`. Anything after `##` in a label is part of the id but not shown:
`"Switch off##switch"` and `"Switch on##switch"` are the same button with different words.

| Call | What it does |
|---|---|
| `Push(name)`, `Pop()` | Open and close a scope, so controls with the same label in a loop stay apart |
| `"Label##id"` | Words after `##` are part of the id but not shown |
| `Area(path…)` | Where a control was drawn in the last frame, by its scopes and label (`Area("Properties", "Apply")`): to place something next to it, or to click it in a test |

## The examples

Every example accepts the window's flags: `-width`, `-height`, `-dark` for the dark theme, and
`-capture file.png` to draw one frame to a picture and exit (`go run ./cmd/<name> -h` lists the rest).

<table>
<tr>
<th width="50%"><code>cmd/example</code>: a panel over a scene</th>
<th width="50%"><code>cmd/app</code>: an application with no 3D</th>
</tr>
<tr>
<td>
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../images/example-dark.png">
  <img alt="A properties panel with tabs, a field and an open select floating over a lighthouse, and a tools panel on the right" src="../images/example-light.png">
</picture>
</td>
<td>
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../images/hero-dark.png">
  <img alt="A register of lighthouses: menu bar, toolbar, navigation list, cards, table, inspector and status bar" src="../images/hero-light.png">
</picture>
</td>
</tr>
<tr>
<th><code>cmd/scene-app</code>: an application around a 3D view</th>
<th><code>cmd/plan-editor</code>: editing the scene it shows</th>
</tr>
<tr>
<td>
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../images/scene-app-dark.png">
  <img alt="A menu bar, a toolbar, a list of lighthouses on the left and an inspector on the right, with the 3D view showing through in the middle" src="../images/scene-app-light.png">
</picture>
</td>
<td>
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../images/plan-editor-dark.png">
  <img alt="An island seen from above with a hut placed on it, and a panel that lists the huts" src="../images/plan-editor-light.png">
</picture>
</td>
</tr>
<tr>
<th><code>cmd/workbench</code>: the pieces of an editor</th>
<th><code>cmd/mobile</code>: phone-sized, on Android too</th>
</tr>
<tr>
<td>
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../images/workbench-dark.png">
  <img alt="A grouped table of parts, the island in the middle, tags as chips on the right" src="../images/workbench-light.png">
</picture>
</td>
<td align="center">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../images/mobile-dark.png">
  <img alt="A list of lighthouses with switches, filter chips and a navigation bar at the bottom" src="../images/mobile-light.png" width="160">
</picture>
</td>
</tr>
</table>

```bash
go run ./cmd/app                      # the application above (-dark for the dark theme)
go run ./cmd/scene-app                # the same kind of layout around a 3D view
go run ./cmd/plan-editor              # place huts on an island's map
go run ./cmd/workbench                # a grouped table, chips, a menu, a sheet, drag and drop
go run ./cmd/example                  # a panel over a scene
go run ./cmd/example -capture shot.png
go run ./cmd/mobile -width 400 -height 860
```

`cmd/web` is the scene example put together for a browser ([web](../7-platforms/1-web.md)); `cmd/mobile` builds into
an Android app ([Android](../7-platforms/2-android.md)).

The library itself needs no cgo; the examples do, as they open a Go Renderer window. The tests need
no display: they measure text with a fake measurer.

```bash
go vet ./...
go test ./...
go test -run xxx -bench .             # a frame of three hundred controls
```

## How it is drawn

At `End` the laid-out boxes become commands of a `draw2d.List` (rounded rectangles, shadows,
lines of text, icons), each panel on its own layer and its content clipped to it. Go Renderer
paints that list after the scene, in the same pass and atlas as its labels, so the whole interface
is one draw call. Text is measured with Go Renderer's own fonts (`overlay.NewMeasurer`) at the
screen's density, so what is laid out is exactly what is painted. A program with fonts of its own
(a brand's) sets `window.Options.Fonts`: the window measures and paints in them (and a program that
makes its interface itself uses `ui.NewInFonts` with the same fonts).

A frame of three panels with three hundred controls builds, lays out and paints in about 0.1 ms
(`go test -bench .`). More in [performance and battery](../8-project/1-performance-audit.md).

Next: [layout](2-layout.md), then the [controls](3-controls.md).
