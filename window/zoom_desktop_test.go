//go:build !js && !android && !nintendo3ds

package window

import "testing"

func TestTheZoomIsKeptUntilChangedAndAZoomOfNothingIsRefused(test *testing.T) {
	defer SetZoom(1)
	SetZoom(1.5)
	if Zoom() != 1.5 {
		test.Fatalf("the zoom is %v after SetZoom(1.5)", Zoom())
	}
	SetZoom(0)
	if Zoom() != 1.5 {
		test.Fatalf("a zoom of nothing was taken: %v", Zoom())
	}
}
