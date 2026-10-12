# go-ui

Immediate-mode interfaces (panels, buttons, labels) drawn over a go-renderer view. It produces a
`draw2d.List` that go-renderer paints, and reads go-renderer's `input.State`. Start with
`docs/README.md`.

## Commands

```bash
go vet ./...
go test ./...                                   # no display needed: tests use a fake text measurer
go test -run xxx -bench .                       # a frame of three hundred controls
go run ./cmd/example                            # needs a display and OpenGL 4.1
go run ./cmd/example -capture frame.png         # one frame to PNG
```

The library itself needs no cgo; the example does (it opens a go-renderer window).

## Rules

- **Generic only.** This project knows nothing about any particular world, game or story. Never put
  names, places, dates, numbers or settings from a user's own project in code, comments, tests,
  examples, docs or commit messages. Examples use made-up places (a lighthouse on an island) and
  round, neutral numbers.
- **Readable code.** Full-word names (no acronyms, no one-letter variables), small functions that do
  one thing, early returns instead of nested branches, comments that say *why*. Match the code
  around you.
- **Tests with every change.** `go vet ./...` and `go test ./...` must pass before committing.
- **Docs live in `docs/`**, written for people: explanatory, friendly, with 2D images where they
  help. Keep the README short and point to `docs/`. Update the docs in the same commit as the
  behaviour they describe.
- **AI disclosure.** Commits written with an AI assistant end with a `Co-Authored-By: Claude …`
  line, and the *AI disclosure* section of the README stays true. The code is pending human review:
  say so, don't overclaim.
- **Git.** Commit only when asked; never push, force-push, rewrite history or delete branches or
  the repository without the author's explicit permission. Do not change the global git config.
- **Renderer changes go to go-renderer**, generic and on a `feature/` branch there; this repo only
  uses its public packages (`draw2d`, `input`, `opengl/overlay`'s measurer, `scene`, `viewer`).

## Human and AI code

- `main` holds hand-written code only. AI work goes on feature branches (`feature/…`) and is merged
  by the author after review.
- Visual changes can't be checked without a display: say what was tested (build, unit tests,
  `-capture` frames) and what still needs a look on screen.
