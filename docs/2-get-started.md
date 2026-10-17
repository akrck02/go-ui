# Get started

This page takes you from nothing to a window with a button that works, in about fifteen minutes.
You need to know a little Go; you need no experience with interfaces.

**What you will do**

1. Install Go and a C compiler, which the window needs.
2. Get go-ui and go-renderer, side by side, and run an example.
3. Write your first interface.

> ⚠️ go-ui is an educational project and its API still changes. Pin the commits you build against.

## 1. Install the tools

| You need | Why | How |
|---|---|---|
| **Go 1.25** or later | The language | [go.dev/dl](https://go.dev/dl/), then `go version` |
| **A C compiler** | The desktop window is opened through cgo | Linux: `gcc`; macOS: `xcode-select --install`; Windows: [MinGW-w64](https://www.mingw-w64.org/) on the `PATH` |
| **OpenGL 4.1** headers (Linux) | What draws on Linux and Windows; macOS draws through Metal | Debian, Ubuntu: `sudo apt install libgl1-mesa-dev xorg-dev` |

The library itself is pure Go: only programs that open a window need cgo. Tests do not, so
`go test ./...` runs anywhere.

## 2. Get the code

go-ui draws through [go-renderer](https://github.com/akrck02/go-renderer) and, until both publish
versions, builds against a copy of it checked out beside it:

```bash
mkdir workspace && cd workspace
git clone https://github.com/akrck02/go-renderer
git clone https://github.com/akrck02/go-ui
cd go-ui
go run ./cmd/app
```

A whole desktop application opens: a menu bar, a toolbar with a search box, a list, a table and an
inspector. Try `go run ./cmd/app -dark`, and the other examples:

| Example | Shows |
|---|---|
| `go run ./cmd/app` | A desktop application with no 3D at all |
| `go run ./cmd/example` | Panels over a 3D scene |
| `go run ./cmd/scene-app` | An application layout around a 3D view |
| `go run ./cmd/workbench` | A grouped table, chips, a menu, a sheet, drag and drop |
| `go run ./cmd/plan-editor` | Editing a 3D scene: placing huts on an island's map |

> 💡 **No display?** Add `-capture shot.png` to draw one frame into a picture; on a server, run it
> under `xvfb-run`.

## 3. Your first interface

Make a module beside the other two and point it at them:

```bash
cd .. && mkdir lighthouse && cd lighthouse
go mod init example.com/lighthouse
go mod edit -require github.com/akrck02/go-ui@v0.0.0 -replace github.com/akrck02/go-ui=../go-ui
go mod edit -require github.com/akrck02/go-renderer@v0.0.0 -replace github.com/akrck02/go-renderer=../go-renderer
```

`main.go`:

```go
package main

import (
	"log"

	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/window"
)

func main() {
	lit := false // your data lives in your variables

	err := window.Run(window.Options{Title: "Lighthouse"}, func(context *ui.Context) {
		context.Heading("North Point")
		if context.PrimaryButton("Switch") { // true on the frame it is clicked
			lit = !lit
		}
		if lit {
			context.Label("The lamp is lit.")
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

```bash
go mod tidy
go run .
```

Click the button: the line appears and goes. That is the whole idea of go-ui, called **immediate
mode**:

- The function you give `window.Run` **describes the interface**, and runs again whenever something
  happens: a click, a key, the window resized.
- A control is a function call. A button **returns `true`** on the frame it is clicked; there are
  no button objects to create, keep or wire with callbacks.
- **Your variables are the truth.** Showing something is an `if`, a list is a `for`, and the
  screen can never show a stale value.

Between frames the window sleeps: an idle go-ui program uses no processor.

## Where next

| If you want to… | Read |
|---|---|
| Learn step by step, building a register of lighthouses | [The tutorial](3-the-quest/0-the-quest.md) |
| Understand frames, layout, controls and themes in depth | [The guide](4-guides/1-getting-started.md) |
| Look up a control or an option | [The reference](6-reference/0-reference.md) |
| Put the interface over a 3D scene | [Lesson 6](3-the-quest/6-over-a-3d-scene.md) |
| Run it in a browser or on a phone | [In a web browser](7-platforms/1-web.md), [on Android](7-platforms/2-android.md), [go-ui-android](https://github.com/akrck02/go-ui-android) |

---

[← Documentation](4-guides/1-getting-started.md) · [Tutorial →](3-the-quest/0-the-quest.md)
