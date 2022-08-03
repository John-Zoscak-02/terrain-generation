package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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

type TerrainMap struct {
	typ uint8
	// Gradient widths need to be odd numbers
	gradient_width_b1  uint32
	gradient_height_b1 uint32
	gradient_width_b2  uint32
	gradient_height_b2 uint32
	// Seed for the macro gradient board
	seed1 int32
	// Seed for the micro gradient board
	seed2 int32
	// Magnitude / Amplitude of the terrain
	m float32
	// The significiance of macro and micro componenets of the bipartite terrain
	prop float32
}

func prepareScene(cam_multipliler uint32) (*app.Application, *core.Node, *camera.Camera) {
	// Create application and scene
	a := app.App()
	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create perspective camera
	cam := camera.New(1)
	camPosition := float32(3.0)
	cam.SetPosition(-camPosition/2, -camPosition*2, camPosition*2)
	cam.LookAt(&math32.Vector3{0, 0, -1.2}, &math32.Vector3{0, 0, 1})
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

	return a, scene, cam
}

func completeScene(a *app.Application, scene *core.Node, terrain Terrain, cam *camera.Camera) {
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
			terrain.MoveDown(yDisp - (int(ySlider.Value()*570) - 285))
		} else if int(ySlider.Value()*570)-285 < yDisp {
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
		} else if int(xSlider.Value()*570)-285 < xDisp {
			terrain.MoveRight(xDisp - (int(xSlider.Value()*570) - 285))
		}
		xDisp = int(xSlider.Value()*570) - 285
	})
	scene.Add(xSlider)

	// water plane
	//waterGeometry := geometry.NewPlane(GRADIENT_WIDTH_B1-1, GRADIENT_HEIGHT_B1-1)
	//waterColor := material.NewStandard(math32.NewColor("darkblue"))
	//water := graphic.NewMesh(waterGeometry, waterColor)
	//waterGeometry.OperateOnVertices(func(vertex *math32.Vector3) bool {
	//	vertex.Z = -0.2 * M
	//	return false
	//})
	//scene.Add(water)

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

func renderSimpleTerrain(terrainMap TerrainMap, terrainWidth, terrainHeight uint32) {
	var board GradientBoard
	board.initialize(terrainMap.gradient_width_b1, terrainMap.gradient_height_b1, terrainMap.seed1)

	a, scene, cam := prepareScene(terrainMap.gradient_height_b1)

	terrain := new(SimpleTerrain)
	terrain.initialize(board, terrainWidth, terrainHeight, terrainMap.m)
	mat := material.NewStandard(math32.NewColor("darkgrey"))
	mesh := graphic.NewMesh(terrain.geom, mat)
	scene.Add(mesh)

	completeScene(a, scene, terrain, cam)
}

func renderBipartiteTerrain(terrainMap TerrainMap, terrainWidth, terrainHeight uint32) {
	var macro GradientBoard
	var micro GradientBoard
	macro.initialize(terrainMap.gradient_width_b1, terrainMap.gradient_height_b1, terrainMap.seed1)
	micro.initialize(terrainMap.gradient_width_b2, terrainMap.gradient_height_b2, terrainMap.seed2)

	a, scene, cam := prepareScene(terrainMap.gradient_height_b1)

	terrain := new(BipartiteTerrain)
	terrain.initialize(macro, micro, terrainWidth, terrainHeight, terrainMap.m, terrainMap.prop)
	mat := material.NewStandard(math32.NewColor("darkgrey"))
	mesh := graphic.NewMesh(terrain.geom, mat)
	scene.Add(mesh)

	completeScene(a, scene, terrain, cam)
}

// Make the terrain widths (passed as command line arguements) odd numbers
// Some terrain sizes do not work because the gradients will not divide rationally into them(within the specificity of float32). It will be clear after running if the terrain/gradient sizes failed:
//    - A diagonal section of the terrain will not be rendered
// 	  and/or
//    - An extra, large gray triangle will be rendered in the topside of the terrain.
func main() {
	if len(os.Args[1:]) == 3 {
		var i interface{}
		file, err1 := ioutil.ReadFile(fmt.Sprintf("maps/%s.json", os.Args[1]))
		if err1 != nil {
			fmt.Println("Error! Could not read that file")
			return
		}
		err2 := json.Unmarshal(file, &i)
		if err2 != nil {
			fmt.Println("Error! That json file could not be deconstructed into a terrain map")
			return
		}

		terrainMap := TerrainMap{}
		m := i.(map[string]interface{})
		for k, v := range m {
			switch k {
			case "typ":
				terrainMap.typ = uint8(v.(float64))
			case "gradient_width_b1":
				terrainMap.gradient_width_b1 = uint32(v.(float64))
			case "gradient_height_b1":
				terrainMap.gradient_height_b1 = uint32(v.(float64))
			case "gradient_width_b2":
				terrainMap.gradient_width_b2 = uint32(v.(float64))
			case "gradient_height_b2":
				terrainMap.gradient_height_b2 = uint32(v.(float64))
			case "seed1":
				terrainMap.seed1 = int32(v.(float64))
			case "seed2":
				terrainMap.seed2 = int32(v.(float64))
			case "m":
				terrainMap.m = float32(v.(float64))
			case "prop":
				terrainMap.prop = float32(v.(float64))
			}
		}

		terrainWidth, _ := strconv.ParseUint(os.Args[2], 10, 32)
		terrainHeight, _ := strconv.ParseUint(os.Args[3], 10, 32)

		if terrainMap.typ == 1 {
			renderSimpleTerrain(terrainMap, uint32(terrainWidth), uint32(terrainHeight))
		} else if terrainMap.typ == 2 {
			renderBipartiteTerrain(terrainMap, uint32(terrainWidth), uint32(terrainHeight))
		} else {
			fmt.Println("Had problems reading json or the type of map is not valid")
		}
	}
}
