# Components

Every go-ui control, one by one: what it is for, concrete cases with their code, how it behaves
with the mouse, the keyboard, a finger and a screen reader, and its options. Each case is a real
program's code; the pictures are captures of it.

All the cases assume the usual frame: `context` is the `*ui.Context` your build function gets
(`window.Run`) or that you call `Begin` and `End` on. State such as `tab` or `foghorn` is a
variable of your program that lives between frames.

| Page | Components |
|---|---|
| [Text, icons and pictures](1-text.md) | `Heading`, `Label`, `Caption`, `Text`, `Icon`, `IconSized`, `Image`, `Tooltip` |
| [Buttons](2-buttons.md) | `Button`, `PrimaryButton`, `IconButton`, `PrimaryIconButton`, `FlatButton`, `FlatIconButton`, `Selectable`, `Chip`, `RemovableChip`, `Clickable`, `LongPressed` |
| [Choices](3-choices.md) | `Checkbox`, `Toggle`, `Slider`, `SliderStep`, `NumberField`, `Select`, `Tabs`, `ListBox` |
| [Text fields](4-text-fields.md) | `TextField`, `TextFieldWith`, focus, validation, search |
| [Containers](5-containers.md) | `Box`, `Frame`, `Row`, `Column`, `Grid`, `Spacer`, `Separator`, `ScrollArea`, `Fold`, `Paint`, `SwipeBox` |
| [Lists and tables](6-lists-and-tables.md) | `List`, `Table` |
| [Layers](7-layers.md) | `Panel`, `PanelWith`, `Sheet`, `Modal`, `Overlay`, `Popover`, `MenuButton`, `MenuItem`, `Toast` |
| [Dragging and resizing](8-dragging-and-resizing.md) | `DragSource`, `DropTarget`, `Dragging`, `DroppedOnScreen`, `Resizable` |
| [A design system's look](9-design-systems.md) | The theme's design tokens: buttons, the chosen item, line heights, check boxes |

## Rules every control follows

- **A control is a call.** It is drawn in the frame it is called in, and gone in the frame it is
  not. Show it with an `if`; repeat it with a `for`.
- **It answers what happened in this frame.** Buttons return `true` on the frame they are clicked;
  controls that edit a value take a pointer to it and return `true` on the frame they change it.
- **It is told apart by its label and its scopes** (its *id*). In a loop, open a scope per item
  with `context.Push(name)` … `context.Pop()`. Text after `##` is part of the id but not shown.
- **Colours and sizes come from `context.Theme`.** Change a token and every control follows.
- **The pointer works from the last frame's layout**, so a control reacts where it was drawn.

---

[← Documentation](../4-guide.md)
