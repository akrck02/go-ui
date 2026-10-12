# go-ui

> **Note:** this project was written with an AI assistant and is pending thorough review. See the
> [AI disclosure](#ai-disclosure).

Interfaces over a [go-renderer](https://github.com/akrck02/go-renderer) view: panels, buttons and
labels drawn on top of a 3D scene, in **immediate mode**. Every frame your program describes the
interface again from its own variables, and each control tells you what happened to it.

```go
interfaceContext, _ := ui.New()
world.Overlay.Drawing = interfaceContext.Drawing()

update := func(world *scene.Scene, state *input.State, step float64) {
	interfaceContext.Begin(state)
	interfaceContext.Panel("Lighthouse", func() {
		interfaceContext.Heading(light.Name)
		if interfaceContext.PrimaryButton("Switch on") {
			light.On = true
		}
	})
	interfaceContext.End()
}
viewer.Run(world, options, "Lighthouse", update, 0)
```

![The example: a properties panel with tabs, an open select, a checkbox, a slider and a number field over a scene](docs/images/example-light.png)

The same pieces build a traditional desktop application with no 3D at all: package `window` opens
a plain window, and `cmd/app` lays out a menu bar, a toolbar with a search box, a navigation list,
summary cards, a table, an inspector and a status bar.

![A desktop application: menu bar, toolbar, list, cards, table, inspector and status bar](docs/images/app-light.png)

```bash
go run ./cmd/app                      # the application above (-dark for the dark theme)
go run ./cmd/scene-app                # the same kind of layout around a 3D view
go run ./cmd/example                  # the example above; -dark for the dark theme
go run ./cmd/example -capture shot.png
go test ./...
```

> ⚠️ **Warning:** an educational experiment, **not suitable for production use**. APIs will
> break. Until go-renderer publishes `draw2d` and the wider input, `go.mod` points at a local copy
> (`../go-renderer`).

Read the [documentation](docs/README.md): how a frame works, the layout rules and the controls.

## AI disclosure

I write most of my code by hand, so I want to be clear about where AI was used here.

- **Written by an AI assistant (Claude, by Anthropic) under my direction, since October 2026:**
  all of it so far: the context, the layout, the controls, the theme, the window (and its choice of
  Metal or OpenGL), the examples and the docs.

Commits made with AI carry a `Co-Authored-By: Claude` line.
