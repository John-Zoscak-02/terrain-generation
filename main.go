package main

import (
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
)

const TERRAIN_WIDTH = 50
const TERRAIN_HEIGHT = 50

// Gradient widths need to be odd numbers
const GRADIENT_WIDTH_B1 = 7
const GRADIENT_HEIGHT_B1 = 7
const GRADIENT_WIDTH_B2 = 13
const GRADIENT_HEIGHT_B2 = 13
const SEED_1 = 43
const SEED_2 = 37

func main() {
	var board1 Board
	//var board2 Board
	board1.initialize(GRADIENT_WIDTH_B1, GRADIENT_HEIGHT_B1, SEED_1)
	//board2.initialize(GRADIENT_WIDTH_B2, GRADIENT_HEIGHT_B2, SEED_2)

	//////////////////////////////////
	// ========== Test 1 ========== //
	//////////////////////////////////

	//img1 := image.NewRGBA(image.Rect(-TERRAIN_WIDTH/2, -TERRAIN_HEIGHT/2, TERRAIN_WIDTH/2, TERRAIN_HEIGHT/2))
	//stackedRenderImg(board1, board2, img1, 0.75)
	//imgFile1, err1 := os.Create("img/terrain3.png")
	//if err1 != nil {
	//	fmt.Println("Cannot create file: ", err1)
	//}
	//png.Encode(imgFile1, img1.SubImage(img1.Rect))
	//imgFile1.Close()

	//////////////////////////////////
	// ========== Test 2 ========== //
	//////////////////////////////////

	// Create application and scene
	a := app.App()
	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create perspective camera
	cam := camera.New(1)
	camPosition := float32(GRADIENT_HEIGHT_B1) * 0.4
	cam.SetPosition(-camPosition/2.0, -camPosition*2, camPosition*2.0)
	cam.LookAt(&math32.Vector3{0, 0, -0.5}, &math32.Vector3{0, 0, 1})
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
	board1.renderGl(scene, xb, yb)

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 1.0))
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(2, 0, 4)
	scene.Add(pointLight)

	// Set background color to gray
	a.Gls().ClearColor(0.3, 0.3, 0.3, 1.0)

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
	})

	/////////////////////////////////
	// ========== Test 3 ========== //
	//////////////////////////////////

	//img3 := image.NewRGBA(image.Rect(-TERRAIN_WIDTH/2, -TERRAIN_HEIGHT/2, TERRAIN_WIDTH/2, TERRAIN_HEIGHT/2))
	//board2.renderImg(img3)
	//imgFile3, err3 := os.Create("img/terrain2.png")
	//if err3 != nil {
	//	fmt.Println("Cannot create file: ", err3)
	//}
	//png.Encode(imgFile3, img3.SubImage(img3.Rect))
	//imgFile3.Close()
}
