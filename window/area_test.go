//go:build !android && !js && !nintendo3ds

package window

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/opengl"
)

func TestASceneAreaIsCountedFromTheFramebuffersBottom(t *testing.T) {
	area := framebufferArea(draw2d.Rect{X: 100, Y: 50, Width: 400, Height: 240}, 1600, 2)
	if area != (opengl.Area{X: 200, Y: 1600 - 100 - 480, Width: 800, Height: 480}) {
		t.Fatalf("the area at 100, 50 of a window at twice its pixels is %+v", area)
	}
}

func TestASceneBelowWithAPlaceIsDrawnOverAWholeWindowScene(t *testing.T) {
	corner := draw2d.Rect{X: 900, Y: 500, Width: 300, Height: 200}
	whole := &SceneView{Below: &BelowView{Area: &corner}}
	if !drawsInAreas(whole) {
		t.Fatal("a scene below with a corner of the window is not drawn")
	}
	if drawsInAreas(&SceneView{Below: &BelowView{}}) {
		t.Fatal("a scene below with no place is drawn on a computer")
	}
}
