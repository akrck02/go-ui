# bubble

```go
import "github.com/akrck02/go-ui/bubble"
```

Package bubble draws go-ui in Bubble 4, the design system of akrck02's projects
(https://github.com/akrck02/bubble). Bubble makes a theme in Go, with its
contrast tested; this package turns it into go-ui's own theme, so every control
follows it, and paints glass's ground.

    theme, err := design.NewTheme(design.Options{Style: design.StylePaper, Mode: design.ModeLight,
    	Density: design.DensityComfortable, Accent: accent}) // design is github.com/akrck02/bubble
    context.Theme = bubble.Theme(theme)

### Functions

| Function | |
|---|---|
| `func Ground(theme design.Theme, width, height int) *image.RGBA` | Ground paints glass's ground the way the browser does, layer by layer: the background, the second light over it, the first light on top. |
| `func Theme(theme design.Theme) ui.Theme` | Theme turns a Bubble theme into go-ui's. |

---

[← Reference](0-reference.md) · generated from the doc comments by `generate.py`
