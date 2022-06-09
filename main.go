package main

import (
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
)

const TERRAIN_WIDTH = 100
const TERRAIN_HEIGHT = 100

// Gradient widths need to be odd numbers
const GRADIENT_WIDTH_B1 = 7
const GRADIENT_HEIGHT_B1 = 7
const GRADIENT_WIDTH_B2 = 17
const GRADIENT_HEIGHT_B2 = 17

const M = 1.2
const PROPORTION = 0.85
const SEED_1 = 43
const SEED_2 = 97

func main() {
	var board1 Board
	var board2 Board
	board1.initialize(GRADIENT_WIDTH_B1, GRADIENT_HEIGHT_B1, SEED_1)
	board2.initialize(GRADIENT_WIDTH_B2, GRADIENT_HEIGHT_B2, SEED_2)

	// Create application and scene
	a := app.App()
	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create perspective camera
	cam := camera.New(1)
	camPosition := float32(GRADIENT_HEIGHT_B1) * 0.35
	cam.SetPosition(-camPosition/2.0, -camPosition*2, camPosition*2.0)
	cam.LookAt(&math32.Vector3{0, 0, -1}, &math32.Vector3{0, 0, 1})
	scene.Add(cam)

	// Set up orbit control for the camera
	camera.NewOrbitControl(cam)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		cam.SetAspect(float32(width) / float32(height))
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	xb := Bounds{-TERRAIN_WIDTH / 2, TERRAIN_WIDTH / 2}
	yb := Bounds{-TERRAIN_HEIGHT / 2, TERRAIN_HEIGHT / 2}
	//geom := board1.GenerateSurfaceGeometry(xb, yb, 1.2)
	geom := GenerateStackedSurfaceGeometry(board1, board2, xb, yb, M, PROPORTION)
	mat := material.NewStandard(math32.NewColor("darkgrey"))
	mat.SetOpacity(1)
	mesh := graphic.NewMesh(geom, mat)
	scene.Add(mesh)

	ySlider := gui.NewVSlider(5, 585)
	ySlider.SetPosition(5, 5)
	ySlider.SetScaleFactor(585)
	ySlider.SetValue(297)
	ySlider.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		MoveUp(geom, board1, board2, xb, yb, M, PROPORTION, 1)
	})
	scene.Add(ySlider)

	xSlider := gui.NewVSlider(5, 585)
	xSlider.SetPosition(15, 5)
	xSlider.SetScaleFactor(585)
	xSlider.SetValue(297)
	xSlider.Subscribe(gui.OnChange, func(name string, ev interface{}) {

	})
	scene.Add(xSlider)

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.5))
	light := light.NewPoint(&math32.Color{1, 1, 1}, 5)
	light.SetPositionVec(math32.NewVector3(-8, 10, 8))
	light.SetLinearDecay(0.2)
	light.SetQuadraticDecay(0)
	scene.Add(light)

	// Set background color to gray
	a.Gls().ClearColor(0.3, 0.3, 0.3, 1.0)

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
	})
}
