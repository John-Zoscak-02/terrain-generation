package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	Northeast = 0
	Southeast = 1
	Southwest = 2
	Northwest = 3
)

// Must make the terrain width and height odd numbers
const TERRAIN_WIDTH = 100
const TERRAIN_HEIGHT = 100
const GRADIENT_WIDTH = 6
const GRADIENT_HEIGHT = 6
const SEED = 37
const C = 3517
const K = 131072

type Bounds struct {
	lower, upper int32
}

type Board struct {
	gradients [GRADIENT_HEIGHT + 2][GRADIENT_WIDTH + 2]uint8
	xBounds   Bounds
	yBounds   Bounds
}

func calculateGradient(x, y int32) uint8 {
	xu := uint(x * SEED)
	yu := uint(y * SEED)
	xu = ((xu >> 16) ^ xu) * 0x45d9f3b
	xu = ((xu >> 16) ^ xu) * 0x45d9f3b
	xu = (xu >> 16) ^ xu
	yu = ((yu >> 16) ^ yu) * 0x45d9f3b
	yu = ((yu >> 16) ^ yu) * 0x45d9f3b
	yu = (yu >> 16) ^ yu
	return uint8((31*(31+xu) + yu) % 360)
}

func toX(gradient uint8) float32 {
	//switch gradient {
	//case Northeast:
	//	return 1
	//case Southeast:
	//	return 1
	//case Southwest:
	//	return -1
	//default:
	//	return -1
	//}
	return float32(math.Cos(float64(gradient) * (math.Pi / 180)))
}

func toY(gradient uint8) float32 {
	//switch gradient {
	//case Northeast:
	//	return 1
	//case Southeast:
	//	return -1
	//case Southwest:
	//	return 1
	//default:
	//	return -1
	//}
	return float32(math.Sin(float64(gradient) * (math.Pi / 180)))
}

func (bounds *Bounds) size() int {
	return int(math.Abs(float64(bounds.upper) - float64(bounds.lower)))
}

func perlinNoise(board *Board, x, y float32) float32 {
	x0 := int32(math.Floor(float64(x))) - board.xBounds.lower
	x1 := x0 + 1
	y0 := int32(math.Floor(float64(y))) - board.yBounds.lower
	y1 := y0 + 1

	fmt.Println(x0, x1, y0, y1)

	sx := (x - float32(board.xBounds.lower)) - float32(x0)
	sy := (y - float32(board.yBounds.lower)) - float32(y0)

	n00x := toX(board.gradients[x0][y0]) * sx
	n00y := toY(board.gradients[x0][y0]) * sy
	n00 := n00x + n00y

	n10x := toX(board.gradients[x1][y0]) * (sx - 1)
	n10y := toY(board.gradients[x1][y0]) * sy
	n10 := n10x + n10y

	n01x := toX(board.gradients[x0][y1]) * sx
	n01y := toY(board.gradients[x0][y1]) * (sy - 1)
	n01 := n01x + n01y

	n11x := toX(board.gradients[x1][y1]) * (sx - 1)
	n11y := toY(board.gradients[x1][y1]) * (sy - 1)
	n11 := n11x + n11y

	fu := float32((6.0 * math.Pow(float64(sx), 5.0)) - (15.0 * math.Pow(float64(sx), 4.0)) + (10.0 * math.Pow(float64(sx), 3.0)))

	nx0 := (n00 * (1 - fu)) + (n10 * fu)
	nx1 := (n01 * (1 - fu)) + (n11 * fu)

	fv := float32((6.0 * math.Pow(float64(sy), 5.0)) - (15.0 * math.Pow(float64(sy), 4.0)) + (10.0 * math.Pow(float64(sy), 3.0)))

	nxy := (nx0 * (1 - fv)) + (nx1 * fv)
	return nxy
}

func (board *Board) initialize() {
	for y := range board.gradients {
		for x := range board.gradients[y] {
			board.gradients[y][x] = calculateGradient(int32(x)+board.xBounds.lower, int32(y)+board.yBounds.lower)
		}
	}
	board.xBounds = Bounds{-GRADIENT_WIDTH / 2, GRADIENT_WIDTH / 2}
	board.yBounds = Bounds{-GRADIENT_HEIGHT / 2, GRADIENT_HEIGHT / 2}
}

func (board *Board) render(img *image.RGBA, xRange Bounds, yRange Bounds) {
	incY := float32(yRange.size()) / float32(img.Rect.Size().Y)
	incX := float32(xRange.size()) / float32(img.Rect.Size().X)
	imgY := img.Rect.Min.Y
	imgX := img.Rect.Min.X
	for y := float32(yRange.lower); y <= float32(yRange.upper); y += incY {
		for x := float32(xRange.lower); x <= float32(xRange.upper); x += incX {
			fmt.Print(x, ", ", y, ": ")
			img.SetRGBA(imgX, imgY, color.RGBA{uint8((perlinNoise(board, x, y) + 1.5) * 85.0), 0, 0, 255})
			//fmt.Print((perlinNoise(board, x, y)+1.5)*85.0, " ")
			imgX++
		}
		imgX = img.Rect.Min.X
		imgY++
		fmt.Println()
	}
}

func main() {
	var board Board
	board.initialize()
	possibleYRange := Bounds{board.yBounds.lower + 1, board.yBounds.upper - 1}
	possibleXRange := Bounds{board.xBounds.lower + 1, board.xBounds.upper - 1}
	img := image.NewRGBA(image.Rect(-TERRAIN_WIDTH/2, -TERRAIN_HEIGHT/2, TERRAIN_WIDTH/2, TERRAIN_HEIGHT/2))
	board.render(img, possibleXRange, possibleYRange)

	imgFile, err := os.Create("img/terrain3.png")
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Cannot create file: ", err)
	}
	png.Encode(imgFile, img.SubImage(img.Rect))
}
