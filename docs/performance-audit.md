# Performance and battery audit

**Date:** October 2026. **Scope:** go-ui and the 2D path of go-renderer it draws through (draw list,
painter, atlas, flat renderer, window loop), measured with the register application (`cmd/app`).
**Machine:** a laptop with an Apple M4 and a 60 Hz built-in screen, macOS 26, Go 1.25.

The short answer: **an idle application now costs nothing**, and a frame of a normal interface is
cheap. What is left to win is work done in frames that change nothing, and work that grows with
content nobody sees (rows scrolled out of view). A Metal backend is not needed for battery; it is
worth designing as insurance against OpenGL leaving macOS ([design](../../go-renderer/docs/metal.md)).

## How it was measured

- **The whole program**, in a real window at 1280 × 720: processor time (`ps`), idle wakeups and
  energy (`top -stats cpu,idlew,power`), GPU use (`ioreg`, "Device Utilization %"), memory
  (`footprint`), frames per second (a probe program counting frames).
- **One frame**, without a window: Go benchmarks with the real fonts at a 2560 × 1440 framebuffer,
  CPU and allocation profiles (`go test -bench -cpuprofile -memprofile`, `go tool pprof`).

The benchmark files stay in the repositories so the numbers can be measured again:

| File | What it measures |
|---|---|
| `go-ui/cmd/app/audit_bench_test.go` | A frame of the register: idle, pointer moving, a field focused, typing, a search, 11 to 3000 rows; the painter on that frame |
| `go-ui/audit_frame_test.go` | Struct sizes, `ScrollArea` against `List` from 11 to 10 000 rows, the phases of a frame, the text field, cache misses |
| `go-ui/audit_render_test.go` | The update step, painting, the whole frame's CPU, hashing the draw list, whether an idle pointer move changes it |
| `go-renderer/opengl/overlay/audit_painter_test.go` | The painter on 2000 commands, vertex filling (now and packed), sorting, clipping, glyphs, shadows, atlas growth |
| `go-renderer/opengl/overlay/audit_register_paint_test.go` | The painter on the register's draw list, per text line and per character |

```bash
go test ./cmd/app -run Audit -bench Audit -benchmem
go test . -run Audit -bench Audit -benchmem
```

## The whole program

| Register at 1280 × 720 | Draws on demand (default) | Draws every frame (`-continuous`) |
|---|---|---|
| Processor, idle | **0.14 %** (0.10 s in 60 s) | 3.9 % (2.45 s in 60 s) |
| GPU, idle | **0 %** | 16 % |
| Idle wakeups | 0 | 0 |
| Frames while idle | **0** (6 at start-up) | 60 to 120 per second, for 60 updates |
| Memory (system footprint) | 116 MB | 166 MB |

Memory was 186 MB before this audit: the window asked for a depth and stencil buffer that flat
drawing never uses. `FlatOnly` now leaves it out. Of what remains, about 70 MB is the window's own
buffers that macOS keeps (they grow with the window) and about 25 MB the program itself.

## One frame

| Frame | go-ui (build, layout, draw list) | Painter (draw list to quads) | Vertex filling | OpenGL calls |
|---|---|---|---|---|
| Register, 11 rows | 44 µs, 114 allocations | 85 µs | 13 µs | about 0.1–0.2 ms |
| Register, 1000 rows | 1.9 ms, 4077 allocations | 3.65 ms | 13 µs | the same |
| Register, 1000 rows in a `List` | 77 µs, 65 allocations | about 0.1 ms | 13 µs | the same |

Hovering, typing and a focused field cost about the same as an idle frame. At 120 pointer events
a second, the 11-row register spends about 1.7 % of a core; with 1000 rows in a `ScrollArea`, about
67 %, and from about 2000 rows it no longer fits in a frame.

## Findings

Ranked by what they cost the battery. Effort: S under half a day, M one or two days, L more.

| # | Finding | Where | Impact | Fix | Effort |
|---|---|---|---|---|---|
| 1 | **Frames that change nothing are painted and shown.** A pointer moving over still areas, the frame that follows input, and in continuous mode up to half of all frames (60 to 120 shown for 60 updates) produce a draw list identical to the last | `opengl/opengl.go` loop, `waitForWork` | High: each is a full paint, upload, draw and swap | Hash the draw list (6 µs) with the background, size and scale; when unchanged, skip painting, drawing and the swap. In continuous mode, draw only after an update step | S–M |
| 2 | **Rows out of sight are built, laid out and painted.** A table in a `ScrollArea` costs 1.9 µs per row whether seen or not | `go-ui/scroll.go`, `paint.go`; `cmd/app` table | High for long lists: 25 to 47 times less with `List` | Use `List` for long tables (the example should); in `paintBox`, skip children outside the clip | S (app), M (library) |
| 3 | **Text cut away by a clip is laid out glyph by glyph, then thrown away.** 29 % of the quads of a scrolled panel, 98 % at 1000 rows | `draw2d/list.go` (texts are never culled), `overlay/drawing.go` | High for long lists: 3.65 ms to about 0.1 ms at 1000 rows | Reject a text whose line (baseline minus size to baseline plus a third) misses the clip, before laying it out | S |
| 4 | **The text measure cache empties itself when full**, in the middle of a frame, past 4096 texts | `go-ui/paint.go` | Medium: +37 % frame time from 1900 to 2100 rows | Keep two generations, so texts used this frame are never dropped | S |
| 5 | **Each character is looked up twice**, with a `[]rune` conversion per text | `overlay/layout.go` (`drawable`, `hasGlyph`), `overlay/drawing.go` | Medium: 64 ns per character, 62 % of painting | One atlas lookup that also knows a missing glyph; a direct table for ASCII | M |
| 6 | **192 bytes per quad, in a buffer reallocated every frame** (0.5–1.2 MB a frame) | `opengl/overlay.go` | Medium in continuous mode | Instanced quads of about 32 bytes; keep the buffer and use `BufferSubData` | M |
| 7 | **Rendering at the display rate is not guaranteed by vsync** on macOS: 60, 90 and 120 fps were measured on a 60 Hz screen, even with `SwapInterval(1)` (now set explicitly) | `opengl/window.go` | Medium in continuous mode | Finding 1 caps it: no update, no frame | S |
| 8 | **Large structs copied by value**: an element is 344 bytes, a draw command 136; copying is about 20 % of a frame | `go-ui/element.go`, `layout.go`, `draw2d/list.go` | Low to medium, grows with the tree | Build elements in place; smaller enums; cold fields apart | M |
| 9 | **The whole atlas is uploaded again** when one glyph or corner is added (1 MB, 16 MB at its largest), and nothing is ever evicted | `opengl/overlay.go`, `overlay/atlas.go` | Low: one-off pauses, not steady cost | Upload only the changed rectangle (`TexSubImage2D`) | S |
| 10 | **The flat path does 3D work**: a second full-screen clear, the 3D program bound, the 3D shaders compiled at start | `opengl/opengl.go` | Low | When `FlatOnly`, let the flat renderer own the frame | S |
| 11 | **A dual-GPU Intel Mac stays on the discrete GPU** while an OpenGL window is open: GLFW does not opt into graphics switching by default | `opengl/window.go` | High on those Macs, none on Apple silicon | `glfw.WindowHint(glfw.CocoaGraphicsSwitching, glfw.True)` for flat windows | S |
| 12 | **The caret wakes the window twice a second** for as long as a field has the focus | `go-ui/textfield.go` | Low | Stop blinking after a while without typing, as macOS does | S |
| 13 | **A focused field measures prefixes every frame**, needed only while clicking or dragging | `go-ui/textfield.go` | Low (16 allocations a frame with a long text) | Measure only on press or drag; keep widths until the text or caret changes | S |
| 14 | **The example application allocates every frame**: the dark theme built twice to compare it, `fmt` per cell, the order sorted and the search lower-cased three times per row | `cmd/app` | Low, but it is the code people copy | Keep themes in variables; work out the shown rows once; make `ParseHex` allocation-free | S |
| 15 | **Small per-frame bookkeeping**: the painter sorts commands that are already in order; `rootElementOf` walks up for every element; memory maps are scanned every frame | several | Very low | Check before sorting; record the root while building; sweep every 60 frames | S |
| 16 | **`StartLoop` crashes when not called on the main thread** (macOS needs it), and says nothing about it | `opengl/opengl.go` | Robustness, not speed | Document it and check it at start | S |

Measured and found fine: one draw call and one program per frame, uniform locations cached, no
`glGetError`, cursor shape and lock changed only when they change, the window title updated only in
continuous mode, no busy waiting when idle, theme changes and window resizes do not rasterise
glyphs again. A partial redraw (repainting only what changed) is not worth it: after a swap the
back buffer is undefined on macOS, so it would need a full copy each frame.

## Done: steps 1 to 4

Applied in October 2026, with tests for each; the documentation captures came out identical, pixel
for pixel.

| Measured again | Before | After |
|---|---|---|
| Register frame, 11 rows (go-ui) | 44 µs, 114 allocations | 37 µs, 109 allocations |
| Register frame, 1000 rows (go-ui) | 1.9 ms | **55 µs** |
| Register frame, 3000 rows (go-ui) | 7.5 ms | **109 µs** |
| Painter, register with 1000 rows | 3.65 ms | **69 µs** |
| Continuous mode, frames shown | 60 to 120 a second | 60, one per update, and none when nothing changed |
| Continuous mode, GPU | 16 % | **0 %** while nothing changes |
| Memory (system footprint) | 116 MB | **93 MB** |

What changed:

- **Frames that would look the same are skipped** (1, 7). `draw2d.List.Fingerprint` sums up a draw
  list; go-ui's window answers go-renderer's new `Application.Unchanged` with it, and the loop then
  neither draws nor presents. In continuous mode a frame is drawn only after an update step, and
  the loop sleeps until the next one.
- **What is out of sight costs nothing** (2, 3). The draw list drops texts whose line misses the
  clip; go-ui does not paint elements wholly outside a clipping box; the example's table is a
  `List`, and works out its rows once a frame.
- **The text cache keeps two generations** (4), and a focused field measures only when something
  changed (13), keeps its text as a string, and **stops blinking after ten seconds** without a key
  or a click (12).
- **The flat window is flat** (10, 11, 16): no 3D shaders, matrices or second clear; automatic
  graphics switching on Macs with two GPUs; `StartLoop` returns an error off the main thread
  instead of crashing.

## What to do, in order

1. **Skip unchanged frames** (finding 1), with continuous mode drawing only after an update (7).
2. **Stop paying for what is out of sight**: cull texts by the clip in the painter (3), skip
   children outside the clip in go-ui (2), and move the example's table to `List`.
3. **The text cache's generations** (4) and the **field's measuring** (13).
4. **Graphics switching** for flat windows (11) and **the flat path's own frame** (10).
5. **The painter's per-character work** (5), **packed instanced quads** (6) and **partial atlas
   uploads** (9).
6. Smaller items: copies (8), the caret (12), the example (14), bookkeeping (15), the main thread (16).

Steps 1 to 4 are done (above); step 5 and the smaller items are left.

## Metal

Not needed for battery: with drawing on demand and one draw call a frame, Apple's layer that runs
OpenGL on Metal costs about 0.1–0.2 ms per frame that is drawn, and nothing while idle. It is worth
designing because OpenGL has been deprecated on macOS since 2018 and may be removed; the
[design](../../go-renderer/docs/metal.md) ports only the flat path first, reusing the painter, in
about three or four days.
