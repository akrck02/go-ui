# A design system's look

A design system is more than a palette: it decides how a chosen item looks, whether the main button
stands out, how tall a line of text is. go-ui's theme has a token for each of those, so a design
system can dress **every control** without the program drawing any of them by hand.

You don't need any of this to use go-ui: left at zero, these tokens keep go-ui's own look.

## The shortest way

Start from a theme of go-ui's, then set what your design system asks for:

```go
theme := ui.LightTheme()
theme.Primary, theme.OnPrimary = draw2d.Hex("#2f6f8f"), draw2d.Hex("#ffffff")

theme.ButtonRadius, theme.ButtonPadding, theme.ButtonWeight = 10, 18, draw2d.Bold
theme.PrimaryEdge, theme.EdgeWidth = draw2d.Hex("#214e64"), 2 // the main button stands on a block
theme.Chosen, theme.ChosenInk = theme.Sunken, theme.Ink       // what is chosen is brighter, not coloured
theme.LineHeight, theme.HeadingLineHeight = 1.5, 1.2

context.Theme = theme // before the first control of the frame
```

Every button, list and check box of the frame now follows it.

## The tokens

| Token | What it changes | go-ui's own look (when zero) |
|---|---|---|
| `ButtonRadius`, `ButtonPadding`, `ButtonGap`, `ButtonWeight` | A button's corners, the room on each side of its label, the gap after its icon, its label's weight | The controls' corners, 12 pixels a side, 6 of gap |
| `PrimaryEdge`, `EdgeWidth` | A solid band under the main button, as if it stood on a block; it takes no room, like a shadow | No edge |
| `FlatHeight`, `FlatPadding`, `FlatInk` | A flat button's height, room and ink at rest | A little shorter than the others, in soft ink, for toolbars |
| `Chosen`, `ChosenInk`, `ChosenWeight` | How a selected item of a list looks | The primary colour's tint |
| `ItemPadding`, `ItemGap`, `ItemIconInk` | The room in an item of a list, the gap after its icon, the icon's ink | 8 and 8, soft icons |
| `LineHeight`, `HeadingLineHeight` | A line as tall as its size × this, and the same for text at `HeadingSize` and above | As tall as the letters |
| `CheckSize`, `CheckRadius`, `CheckBorder`, `CheckMark`, `CheckLine`, `CheckedBorder` | A check box's side, corners, edge, tick, empty edge and ticked edge | go-ui's small box |
| `Backdrop` | Blurs what is behind every translucent panel, as frosted glass, by this many pixels | Nothing blurred, at no cost |
| `FocusWidth` | The width of the ring around the control Tab reached, lying on its edge; inside a primary button, in `OnPrimary` | 3 pixels |
| `FieldFocusEdge` | A field being typed in lights its own edge in `FocusRing` instead of a ring around it | A ring around the field |
| `PressScale` | How big a held control is drawn (0.97 gives a little); with `context.LessMotion` set it stays still | It keeps its size |

`Backdrop` frosts a translucent panel that is not inside another one: the rows hovered or chosen
inside a frosted panel are translucent too, but blurring again under them would cost a blur for
nothing. go-renderer takes one snapshot of the frame per layer, however many panels it has, and
blurs it small; its guide to the overlay says what that costs.

Two rules keep old themes looking the same:

- **A zero size, or an invisible colour, means go-ui's own look**, so a theme made before these
  tokens existed draws as it always did.
- **The chosen item follows the primary colour** while `Chosen` and `ChosenInk` are left alone, so
  a theme that only changes `Primary` and `PrimaryTint` still marks it with its own tint.

## A design system that uses them

[Bubble 4](https://github.com/akrck02/bubble), akrck02's design system, makes its themes in Go with
their contrast tested, and its `goui` module turns one into a go-ui theme. Its guide shows the
same screen in a browser and in go-ui, side by side, in every style and finish. go-ui itself knows
nothing of Bubble: any design system can do the same with these tokens.

---

[← Components](0-components.md)
