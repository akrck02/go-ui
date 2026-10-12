package main

import (
	"image"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestMetalDrawsLikeOpenGL captures the register through both backends, light and dark, and
// compares them pixel by pixel: they share the painter, so they may differ only by rounding (the
// Metal path sends colours as bytes). It opens hidden windows, so it needs macOS with a display;
// -short skips it.
func TestMetalDrawsLikeOpenGL(t *testing.T) {
	if runtime.GOOS != "darwin" || testing.Short() {
		t.Skip("needs macOS and a display")
	}
	directory := t.TempDir()
	program := filepath.Join(directory, "register")
	if output, err := exec.Command("go", "build", "-o", program, ".").CombinedOutput(); err != nil {
		t.Fatalf("building the register: %v\n%s", err, output)
	}
	for _, theme := range []string{"light", "dark"} {
		metal := capture(t, program, directory, "metal", theme)
		opengl := capture(t, program, directory, "opengl", theme)
		worst, differing := compare(t, metal, opengl)
		if worst > 1 {
			t.Errorf("%s: a channel differs by %d/255; at most 1 is rounding", theme, worst)
		}
		if limit := len(metal.Pix) / 4 / 100; differing > limit {
			t.Errorf("%s: %d pixels differ, more than 1 %%", theme, differing)
		}
	}
}

func capture(t *testing.T, program, directory, backend, theme string) *image.NRGBA {
	t.Helper()
	path := filepath.Join(directory, backend+"-"+theme+".png")
	arguments := []string{"-backend", backend, "-capture", path}
	if theme == "dark" {
		arguments = append(arguments, "-dark")
	}
	if output, err := exec.Command(program, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("capturing through %s: %v\n%s", backend, err, output)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	picture, err := png.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	// an opaque PNG decodes as RGBA, which is the same bytes as NRGBA when every pixel is opaque
	converted := image.NewNRGBA(picture.Bounds())
	draw.Draw(converted, converted.Bounds(), picture, picture.Bounds().Min, draw.Src)
	return converted
}

// compare returns the largest difference in a colour channel and how many pixels differ at all.
func compare(t *testing.T, first, second *image.NRGBA) (worst, differing int) {
	t.Helper()
	if first.Bounds() != second.Bounds() {
		t.Fatalf("sizes differ: %v and %v", first.Bounds(), second.Bounds())
	}
	for index := 0; index < len(first.Pix); index += 4 {
		pixelDiffers := false
		for channel := 0; channel < 3; channel++ {
			difference := int(first.Pix[index+channel]) - int(second.Pix[index+channel])
			difference = max(difference, -difference)
			worst = max(worst, difference)
			pixelDiffers = pixelDiffers || difference > 0
		}
		if pixelDiffers {
			differing++
		}
	}
	return worst, differing
}
