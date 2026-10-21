package uitest

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// The window's size when no Size option is given, in layout pixels.
const (
	defaultWidth  = 800
	defaultHeight = 600
)

// settings are what the options choose.
type settings struct {
	width, height float32
	scale         float32
	theme         *ui.Theme
	measurer      draw2d.Measurer
}

// Option chooses something about a test window; give them to New.
type Option func(*settings)

// Size sets the window's size in layout pixels: what the interface's Size reports.
func Size(width, height float32) Option {
	return func(chosen *settings) { chosen.width, chosen.height = width, height }
}

// Scale sets how many framebuffer pixels a layout pixel takes (2 on most phones).
func Scale(factor float32) Option {
	return func(chosen *settings) { chosen.scale = factor }
}

// Theme sets the interface's theme; without it the window uses ui's default.
func Theme(theme ui.Theme) Option {
	return func(chosen *settings) { chosen.theme = &theme }
}

// Measurer sets how text is measured; without it the window uses FixedMeasurer.
func Measurer(measurer draw2d.Measurer) Option {
	return func(chosen *settings) { chosen.measurer = measurer }
}
