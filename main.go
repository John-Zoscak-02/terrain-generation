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
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

const TERRAIN_WIDTH = 127
const TERRAIN_HEIGHT = 127

// Gradient widths need to be odd numbers
const GRADIENT_WIDTH_B1 = 5
const GRADIENT_HEIGHT_B1 = 5
const GRADIENT_WIDTH_B2 = 27
const GRADIENT_HEIGHT_B2 = 27

// Magnitude / Amplitude of the terrain
const M = 1.0

// The significiance of macro and micro componenets of the bipartite terrain
const PROPORTION = 0.89

// Seed for the macro gradient board
const SEED_1 = 43

// Seed for the micro gradient board
const SEED_2 = 97

func main() {
	//Initializing the gradient boards to add to the bipartite terrain
	var macro GradientBoard
	var micro GradientBoard
	macro.initialize(GRADIENT_WIDTH_B1, GRADIENT_HEIGHT_B1, SEED_1)
	micro.initialize(GRADIENT_WIDTH_B2, GRADIENT_HEIGHT_B2, SEED_2)

	// Create application and scene
	a := app.App()
	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create perspective camera
	cam := camera.New(1)
	camPosition := float32(GRADIENT_HEIGHT_B1) * 0.30
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

	//geom := board1.GenerateSurfaceGeometry(xb, yb, 1.2)
	terrain := new(BipartiteTerrain)
	terrain.initialize(macro, micro, TERRAIN_WIDTH, TERRAIN_HEIGHT, M, PROPORTION)
	mat := material.NewStandard(math32.NewColor("darkgrey"))
	mesh := graphic.NewMesh(terrain.geom, mat)
	scene.Add(mesh)

	// water plane
	//waterGeometry := geometry.NewPlane(GRADIENT_WIDTH_B1-1, GRADIENT_HEIGHT_B1-1)
	//waterColor := material.NewStandard(math32.NewColor("darkblue"))
	//water := graphic.NewMesh(waterGeometry, waterColor)
	//waterGeometry.OperateOnVertices(func(vertex *math32.Vector3) bool {
	//	vertex.Z = -0.2 * M
	//	return false
	//})
	//scene.Add(water)

	// Variables to keep track of the current dispacement from the terrain origin
	xDisp := 0
	yDisp := 0

	// Label for Y slider
	sliderYTitle := gui.NewLabel("Y")
	sliderYTitle.SetPosition(5, 3)
	//sliderYTitle.SetSize(5.0, 5.0)
	scene.Add(sliderYTitle)

	// Label for X slider
	sliderXTitle := gui.NewLabel("X")
	sliderXTitle.SetPosition(15, 3)
	//sliderXTitle.SetSize(5.0, 5.0)
	scene.Add(sliderXTitle)

	// Y Slider for changing the rendered terrain in the Y direction
	ySlider := gui.NewVScrollBar(5, 570)
	ySlider.SetPosition(8, 20)
	ySlider.SetScale(0, 570, 0)
	ySlider.SetValue(0.5)
	ySlider.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		if int(ySlider.Value()*570)-285 > yDisp {
			//for i := 0; i < (int(ySlider.Value()*570)-285)-yDisp; i++ {
			//	terrain.MoveDown()
			//}
			terrain.MoveDown(yDisp - (int(ySlider.Value()*570) - 285))
		} else if int(ySlider.Value()*570)-285 < yDisp {
			//for i := 0; i < yDisp-(int(ySlider.Value()*570)-285); i++ {
			//	terrain.MoveUp()
			//}
			terrain.MoveUp(yDisp - (int(ySlider.Value()*570) - 285))
		}
		yDisp = int(ySlider.Value()*570) - 285
	})
	scene.Add(ySlider)

	// X Slider for changing the rendered terrain in the X direction
	xSlider := gui.NewVScrollBar(5, 570)
	xSlider.SetPosition(18, 20)
	xSlider.SetScale(0, 570, 0)
	xSlider.SetValue(0.5)
	xSlider.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		if int(xSlider.Value()*570)-285 > xDisp {
			terrain.MoveLeft(xDisp - (int(xSlider.Value()*570) - 285))
			//for i := 0; i < (int(xSlider.Value()*570)-285)-xDisp; i++ {
			//	terrain.MoveLeft()
			//}
		} else if int(xSlider.Value()*570)-285 < xDisp {
			terrain.MoveRight(xDisp - (int(xSlider.Value()*570) - 285))
			//for i := 0; i < xDisp-(int(xSlider.Value()*570)-285); i++ {
			//	terrain.MoveRight()
			//}
		}
		xDisp = int(xSlider.Value()*570) - 285
	})
	scene.Add(xSlider)

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.5))
	light := light.NewPoint(&math32.Color{1, 1, 1}, 5)
	light.SetPositionVec(math32.NewVector3(-8, 10, 8))
	light.SetLinearDecay(0.2)
	light.SetQuadraticDecay(0)
	scene.Add(light)

	// Create and add an axis helper to the scene
	scene.Add(helper.NewAxes(0.5))

	// Set background color to gray
	a.Gls().ClearColor(0.3, 0.3, 0.3, 1.0)

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
	})
}
