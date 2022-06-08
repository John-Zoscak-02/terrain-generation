package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/math32"
)

////////////////////
//=====Bounds=====//
////////////////////
type Bounds struct {
	lower, upper int32
}

func (bounds *Bounds) size() int {
	return int(math.Abs(float64(bounds.upper) - float64(bounds.lower)))
}

///////////////////
//=====Board=====//
///////////////////
type Board struct {
	gradients [][]uint8
	xBounds   Bounds
	yBounds   Bounds
	seed      int32
}

func (board Board) calculateGradient(x, y int32) uint8 {
	xu := uint(x * board.seed)
	yu := uint(y * board.seed)
	xu = ((xu >> 16) ^ xu) * 0x45d9f3b
	xu = ((xu >> 16) ^ xu) * 0x45d9f3b
	xu = (xu >> 16) ^ xu
	yu = ((yu >> 16) ^ yu) * 0x45d9f3b
	yu = ((yu >> 16) ^ yu) * 0x45d9f3b
	yu = (yu >> 16) ^ yu
	return uint8((31*(31+xu) + yu) % 360)
}

func (board *Board) initialize(gradientWidth, gradientHeight uint32, seed int32) {
	board.gradients = make([][]uint8, gradientHeight)
	board.xBounds = Bounds{-int32(gradientWidth) / 2, int32(gradientWidth) / 2}
	board.yBounds = Bounds{-int32(gradientHeight) / 2, int32(gradientHeight) / 2}
	board.seed = seed
	for y := range board.gradients {
		board.gradients[y] = make([]uint8, gradientWidth)
		for x := range board.gradients[y] {
			board.gradients[y][x] = board.calculateGradient(int32(x)+board.xBounds.lower, int32(y)+board.yBounds.lower)
		}
	}
}

func stackedRenderImg(board1, board2 Board, img *image.RGBA, prop float64) {
	incY1 := float64(board1.yBounds.size()) / float64(img.Rect.Size().Y)
	incX1 := float64(board1.xBounds.size()) / float64(img.Rect.Size().X)
	incY2 := float64(board2.yBounds.size()) / float64(img.Rect.Size().Y)
	incX2 := float64(board2.xBounds.size()) / float64(img.Rect.Size().X)
	imgY := img.Rect.Min.Y
	imgX := img.Rect.Min.X
	for y1, y2 := float64(board1.yBounds.lower), float64(board2.yBounds.lower); y1 <= float64(board1.yBounds.upper); y1, y2 = y1+incY1, y2+incY2 {
		for x1, x2 := float64(board1.xBounds.lower), float64(board2.xBounds.lower); x1 <= float64(board1.xBounds.upper); x1, x2 = x1+incX1, x2+incX2 {
			h1 := uint8(((perlinNoise(board1, x1, y1) + 1.5) * 85.0) * prop)
			h2 := uint8(((perlinNoise(board2, x2, y2) + 1.5) * 85.0) * (1 - prop))
			img.SetRGBA(imgX, imgY, color.RGBA{0, h1 + h2, 255 - (h1 + h2), 255})
			//fmt.Println((perlinNoise(board1, x1, y1)+1.5)*85.0, " ", (perlinNoise(board2, x2, y2)+1.5)*85.0)
			fmt.Println()
			imgX++
		}
		imgX = img.Rect.Min.X
		imgY++
		//fmt.Println()
	}
}

func (board Board) renderImg(img *image.RGBA) {
	incY := float64(board.yBounds.size()) / float64(img.Rect.Size().Y)
	incX := float64(board.xBounds.size()) / float64(img.Rect.Size().X)
	imgY := img.Rect.Min.Y
	imgX := img.Rect.Min.X
	for y := float64(board.yBounds.lower); y <= float64(board.yBounds.upper); y += incY {
		for x := float64(board.xBounds.lower); x <= float64(board.xBounds.upper); x += incX {
			img.SetRGBA(imgX, imgY, color.RGBA{uint8((perlinNoise(board, x, y) + 1.5) * 85.0), 0, 0, 255})
			//fmt.Print((perlinNoise(board, x, y)+1.5)*85.0, " ")
			imgX++
		}
		imgX = img.Rect.Min.X
		imgY++
		//fmt.Println()
	}
}

func (board *Board) renderArr(heights *[][]uint8) {
	incY := float64(board.yBounds.size()) / float64(len(*heights))
	incX := float64(board.xBounds.size()) / float64(len((*heights)[0]))
	scaledY := 0
	scaledX := 0
	for y := float64(board.yBounds.lower); y <= float64(board.yBounds.upper); y += incY {
		for x := float64(board.xBounds.lower); x <= float64(board.xBounds.upper); x += incX {
			(*heights)[scaledY][scaledX] = uint8((perlinNoise(*board, x, y) + 1.5) * 85.0)
			fmt.Print(uint8((perlinNoise(*board, x, y)+1.5)*85.0), " ")
			scaledX++
		}
		scaledX = 0
		scaledY++
		//fmt.Println()
	}
}

func (board *Board) GenerateSurfaceGeometry(xBounds, yBounds Bounds) *geometry.Geometry {
	incY := float32(board.yBounds.size()) / float32(yBounds.size())
	incX := float32(board.xBounds.size()) / float32(xBounds.size())
	positions := math32.NewArrayF32(0, yBounds.size()*xBounds.size())
	indices := math32.NewArrayU32(0, (yBounds.size()-1)*(xBounds.size()-1)*2)
	index := uint32(0)
	for y := float32(board.yBounds.lower); y <= float32(board.yBounds.upper); y += incY {
		for x := float32(board.xBounds.lower); x <= float32(board.xBounds.upper); x += incX {
			height := float32((perlinNoise(*board, float64(x), float64(y)) + 1.5))
			//geom := geometry.NewBox(incY, incY, height)
			//mat := material.NewStandard(math32.NewColor("darkgreen"))
			//mesh := graphic.NewMesh(geom, mat)
			//mesh.SetPositionVec(math32.NewVector3(x, y, height/2))
			//scene.Add(mesh)
			positions.Append(x, y, height, 0, 0, 1)
			if x+incX <= float32(board.xBounds.upper) && y+incY <= float32(board.yBounds.upper) {
				indices.Append(index, index+1+uint32(float32(board.xBounds.size())/incX)+1, index+uint32(float32(board.xBounds.size())/incX)+1)
				//indices.Append(index, index+1, index+1+uint32(float32(board.xBounds.size())/incX)+1)
			}
			index++
		}
	}
	geom := geometry.NewGeometry()
	geom.SetIndices(indices)
	geom.AddVBO(gls.NewVBO(positions).
		AddAttrib(gls.VertexPosition).
		AddAttrib(gls.VertexNormal),
	)

	return geom
	//return scene
}

func GenerateStackedSurfaceGeometry(board1, board2 Board, xBounds, yBounds Bounds, prop float32) *geometry.Geometry {
	incY1 := float32(board1.yBounds.size()) / float32(yBounds.size())
	incX1 := float32(board1.xBounds.size()) / float32(xBounds.size())
	incY2 := float32(board2.yBounds.size()) / float32(yBounds.size())
	incX2 := float32(board2.xBounds.size()) / float32(xBounds.size())

	positions := math32.NewArrayF32(0, yBounds.size()*xBounds.size())
	indices := math32.NewArrayU32(0, (yBounds.size()-1)*(xBounds.size()-1)*2)
	index := uint32(0)

	for y1, y2 := float32(board1.yBounds.lower), float32(board2.yBounds.lower); y1 <= float32(board1.yBounds.upper); y1, y2 = y1+incY1, y2+incY2 {
		for x1, x2 := float32(board1.xBounds.lower), float32(board2.xBounds.lower); x1 <= float32(board1.xBounds.upper); x1, x2 = x1+incX1, x2+incX2 {
			height1 := float32((perlinNoise(board1, float64(x1), float64(y1)) + 1.5) * prop)
			height2 := float32((perlinNoise(board2, float64(x2), float64(y2)) + 1.5) * (1 - prop))
			positions.Append(x1, y1, height1+height2, 0, 0, 1)
			if x1+incX1 <= float32(board1.xBounds.upper) && y1+incY1 <= float32(board1.yBounds.upper) {
				indices.Append(index, index+1+uint32(float32(board1.xBounds.size())/incX1)+1, index+uint32(float32(board1.xBounds.size())/incX1)+1)
				//indices.Append(index, index+1, index+1+uint32(float32(board.xBounds.size())/incX)+1)
			}
			index++
		}
	}
}

////////////////////
//======Math======//
////////////////////
func toX(gradient uint8) float64 {
	return float64(math.Cos(float64(gradient) * (math.Pi / 180)))
}

func toY(gradient uint8) float64 {
	return float64(math.Sin(float64(gradient) * (math.Pi / 180)))
}

func perlinNoise(board Board, x, y float64) float64 {
	x0 := int32(math.Floor(float64(x))) - board.xBounds.lower
	x1 := x0 + 1
	y0 := int32(math.Floor(float64(y))) - board.yBounds.lower
	y1 := y0 + 1

	//fmt.Print(x0, y0, ", ", x1, y1)

	sx := (x - float64(board.xBounds.lower)) - float64(x0)
	sy := (y - float64(board.yBounds.lower)) - float64(y0)

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

	fu := (6.0 * math.Pow(float64(sx), 5.0)) - (15.0 * math.Pow(float64(sx), 4.0)) + (10.0 * math.Pow(float64(sx), 3.0))

	nx0 := (n00 * (1 - fu)) + (n10 * fu)
	nx1 := (n01 * (1 - fu)) + (n11 * fu)

	fv := (6.0 * math.Pow(float64(sy), 5.0)) - (15.0 * math.Pow(float64(sy), 4.0)) + (10.0 * math.Pow(float64(sy), 3.0))

	nxy := (nx0 * (1 - fv)) + (nx1 * fv)
	return nxy
}
