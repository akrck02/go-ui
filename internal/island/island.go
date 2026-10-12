// Package island builds the small scene the examples draw their interfaces over: an island with a
// lighthouse, made of coloured boxes so it needs no files.
package island

import (
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/scene"
)

// Build makes a small scene from boxes: an island, a white tower with a red band and a lamp.
func Build() *scene.Scene {
	world := scene.New()
	world.Nodes = append(world.Nodes,
		box("island", [3]float32{0, -0.5, 0}, [3]float32{24, 1, 18}, graphics.Vec4{0.45, 0.55, 0.38, 1}),
		box("rock", [3]float32{-5, 0.5, 3}, [3]float32{4, 1, 3}, graphics.Vec4{0.55, 0.53, 0.5, 1}),
		box("tower", [3]float32{0, 4, 0}, [3]float32{2.4, 8, 2.4}, graphics.Vec4{0.95, 0.95, 0.93, 1}),
		box("band", [3]float32{0, 5, 0}, [3]float32{2.5, 1.2, 2.5}, graphics.Vec4{0.8, 0.2, 0.18, 1}),
		box("lamp", [3]float32{0, 8.7, 0}, [3]float32{1.8, 1.4, 1.8}, graphics.Vec4{1, 0.85, 0.4, 1}),
		box("roof", [3]float32{0, 9.7, 0}, [3]float32{2.2, 0.6, 2.2}, graphics.Vec4{0.3, 0.34, 0.4, 1}),
	)
	return world
}

// box is a coloured box centred at a point.
func box(name string, centre, size [3]float32, color graphics.Vec4) *scene.Node {
	faces := []struct{ normal, right, up [3]float32 }{
		{[3]float32{1, 0, 0}, [3]float32{0, 0, -1}, [3]float32{0, 1, 0}},
		{[3]float32{-1, 0, 0}, [3]float32{0, 0, 1}, [3]float32{0, 1, 0}},
		{[3]float32{0, 1, 0}, [3]float32{1, 0, 0}, [3]float32{0, 0, -1}},
		{[3]float32{0, -1, 0}, [3]float32{1, 0, 0}, [3]float32{0, 0, 1}},
		{[3]float32{0, 0, 1}, [3]float32{1, 0, 0}, [3]float32{0, 1, 0}},
		{[3]float32{0, 0, -1}, [3]float32{-1, 0, 0}, [3]float32{0, 1, 0}},
	}
	mesh := &scene.Mesh{Name: name, Ground: name == "island", Material: &scene.Material{BaseColor: graphics.Vec4{1, 1, 1, 1}}}
	for _, face := range faces {
		base := uint32(len(mesh.Positions) / 3)
		for _, corner := range [4][2]float32{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}} {
			for axis := 0; axis < 3; axis++ {
				offset := face.normal[axis] + corner[0]*face.right[axis] + corner[1]*face.up[axis]
				mesh.Positions = append(mesh.Positions, centre[axis]+offset*size[axis]/2)
			}
			mesh.Normals = append(mesh.Normals, face.normal[0], face.normal[1], face.normal[2])
			mesh.Colors = append(mesh.Colors, color[0], color[1], color[2], color[3])
		}
		mesh.Indices = append(mesh.Indices, base, base+1, base+2, base, base+2, base+3)
	}
	return &scene.Node{Name: name, Transform: graphics.Identity(), Mesh: mesh}
}

// Paint colours the tower's band, so the scene shows which lighthouse is selected.
func Paint(world *scene.Scene, band graphics.Vec4) {
	for _, node := range world.Nodes {
		if node.Name != "band" || node.Mesh == nil {
			continue
		}
		for index := 0; index+3 < len(node.Mesh.Colors); index += 4 {
			node.Mesh.Colors[index], node.Mesh.Colors[index+1], node.Mesh.Colors[index+2] = band[0], band[1], band[2]
		}
		node.Mesh.Dirty = true
	}
}

// ShowLamp shows or hides the lamp room's light.
func ShowLamp(world *scene.Scene, lit bool) {
	for _, node := range world.Nodes {
		if node.Name == "lamp" {
			node.Hidden = !lit
		}
	}
}
