package uitest

import (
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
)

// dragSteps is how many frames a drag takes to go from where it starts to where it ends, so
// controls that follow the pointer see it move.
const dragSteps = 10

// point is a place in the window, in layout pixels.
type point struct{ x, y float32 }

// center is the middle of a rectangle.
func center(area draw2d.Rect) point {
	return point{x: area.X + area.Width/2, y: area.Y + area.Height/2}
}

// MoveTo moves the pointer to a place in the window and runs a frame.
func (window *Window) MoveTo(x, y float32) {
	window.state.MoveCursor(float64(x), float64(y))
	window.Frame()
}

// MoveToElement moves the pointer to the middle of an element (see Element) and runs a frame.
func (window *Window) MoveToElement(label string) error {
	target, err := window.visibleElement(label)
	if err != nil {
		return err
	}
	middle := center(target.Rect)
	window.MoveTo(middle.x, middle.y)
	return nil
}

// Press holds the main button down where the pointer is and runs a frame.
func (window *Window) Press() {
	window.state.SetButton(input.MouseLeft, true)
	window.Frame()
}

// Release lets the main button go and runs a frame.
func (window *Window) Release() {
	window.state.SetButton(input.MouseLeft, false)
	window.Frame()
}

// ClickAt moves the pointer to a place, presses and releases there: three frames, as a control
// answers when the button is let go.
func (window *Window) ClickAt(x, y float32) {
	window.MoveTo(x, y)
	window.Press()
	window.Release()
}

// Click clicks the middle of an element that can be pressed.
func (window *Window) Click(label string) error {
	target, err := window.pressableElement(label)
	if err != nil {
		return err
	}
	middle := center(target.Rect)
	window.ClickAt(middle.x, middle.y)
	return nil
}

// DoubleClick clicks an element twice with no time between, which the input state counts as a
// double click.
func (window *Window) DoubleClick(label string) error {
	target, err := window.pressableElement(label)
	if err != nil {
		return err
	}
	middle := center(target.Rect)
	window.ClickAt(middle.x, middle.y)
	window.Press()
	window.Release()
	return nil
}

// Drag presses at one place, moves to another over several frames and lets go there.
func (window *Window) Drag(fromX, fromY, toX, toY float32) {
	window.MoveTo(fromX, fromY)
	window.Press()
	for step := 1; step <= dragSteps; step++ {
		part := float32(step) / dragSteps
		window.MoveTo(fromX+(toX-fromX)*part, fromY+(toY-fromY)*part)
	}
	window.Release()
}

// DragElement drags an element's middle by a distance in layout pixels.
func (window *Window) DragElement(label string, distanceX, distanceY float32) error {
	target, err := window.visibleElement(label)
	if err != nil {
		return err
	}
	middle := center(target.Rect)
	window.Drag(middle.x, middle.y, middle.x+distanceX, middle.y+distanceY)
	return nil
}

// Scroll puts the pointer over an element and turns the wheel: a negative amount goes down the
// page, as a wheel turned toward the person does.
func (window *Window) Scroll(label string, amount float32) error {
	if err := window.MoveToElement(label); err != nil {
		return err
	}
	window.state.AddScroll(float64(amount))
	window.Frame()
	return nil
}

// visibleElement finds an element that shows, or says why it cannot be used.
func (window *Window) visibleElement(label string) (Element, error) {
	target, err := window.Element(label)
	if err != nil {
		return Element{}, err
	}
	if !target.Visible {
		return Element{}, elementNotVisible(label)
	}
	return target, nil
}

// pressableElement finds an element that shows and can be pressed.
func (window *Window) pressableElement(label string) (Element, error) {
	target, err := window.visibleElement(label)
	if err != nil {
		return Element{}, err
	}
	if !target.Pressable {
		return Element{}, elementNotPressable(label)
	}
	return target, nil
}
