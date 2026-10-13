package window

import "github.com/akrck02/go-renderer/draw2d"

// Palettes are the colours a phone made from its wallpaper (Android 12 and later, Material You):
// three accent palettes and two neutral ones, each in thirteen shades from white to black, as
// Android names them (system_accent1_0 … system_accent1_1000). A program that builds its colours
// from them matches the phone, as the system's own applications do.
type Palettes struct {
	Accent1, Accent2, Accent3, Neutral1, Neutral2 [13]draw2d.Color
}

// PaletteTones are the shades' numbers in Android's names, in the order of each palette: 0 is
// white, 500 the middle, 1000 black.
var PaletteTones = [13]int{0, 10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}
