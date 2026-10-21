package uitest

import (
	"unicode/utf8"

	"github.com/akrck02/go-renderer/draw2d"
)

// Proportions of the fixed measurer, as parts of the text size.
const (
	characterWidth = 0.5
	ascent         = 0.8
	descent        = 0.2
)

// FixedMeasurer measures text with no fonts: every character is half the text size wide. Tests
// that use it give the same layout on every machine. It has every icon.
type FixedMeasurer struct{}

// MeasureText gives each character half the size in width.
func (FixedMeasurer) MeasureText(text string, size float32, _ draw2d.Weight, _ float32) draw2d.TextMetrics {
	return draw2d.TextMetrics{
		Width:   float32(utf8.RuneCountInString(text)) * size * characterWidth,
		Ascent:  size * ascent,
		Descent: size * descent,
	}
}

// HasIcon says every icon exists.
func (FixedMeasurer) HasIcon(string) bool { return true }
