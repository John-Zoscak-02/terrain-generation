package main

import (
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
	xBounds Bounds
	yBounds Bounds
	seed    int32
}

//type BoardGroup struct {

//}

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
	board.xBounds = Bounds{-int32(gradientWidth) / 2, int32(gradientWidth) / 2}
	board.yBounds = Bounds{-int32(gradientHeight) / 2, int32(gradientHeight) / 2}
	board.seed = seed
}

func (board *Board) GenerateSurfaceGeometry(xBounds, yBounds Bounds, m float32) *geometry.Geometry {
	incY := float32(board.yBounds.size()) / float32(yBounds.size())
	incX := float32(board.xBounds.size()) / float32(xBounds.size())
	positions := math32.NewArrayF32(0, yBounds.size()*xBounds.size())
	indices := math32.NewArrayU32(0, (yBounds.size()-1)*(xBounds.size()-1)*2)
	index := uint32(0)
	for y := float32(board.yBounds.lower); y <= float32(board.yBounds.upper); y += incY {
		for x := float32(board.xBounds.lower); x <= float32(board.xBounds.upper); x += incX {
			height := float32(perlinNoise(*board, x, y)) * m
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
}

func GenerateStackedSurfaceGeometry(board1, board2 Board, xBounds, yBounds Bounds, m, prop float32) *geometry.Geometry {
	incY1 := float32(board1.yBounds.size()) / float32(yBounds.size())
	incX1 := float32(board1.xBounds.size()) / float32(xBounds.size())
	incY2 := float32(board2.yBounds.size()) / float32(yBounds.size())
	incX2 := float32(board2.xBounds.size()) / float32(xBounds.size())

	positions := math32.NewArrayF32(0, yBounds.size()*xBounds.size())
	indices := math32.NewArrayU32(0, (yBounds.size()-1)*(xBounds.size()-1)>>1)
	index := uint32(0)

	for y1, y2 := float32(board1.yBounds.lower), float32(board2.yBounds.lower); y1 < float32(board1.yBounds.upper); y1, y2 = y1+incY1, y2+incY2 {
		for x1, x2 := float32(board1.xBounds.lower), float32(board2.xBounds.lower); x1 < float32(board1.xBounds.upper); x1, x2 = x1+incX1, x2+incX2 {
			height1 := float32((perlinNoise(board1, x1, y1)) * prop)
			height2 := float32((perlinNoise(board2, x2, y2)) * (1 - prop))
			positions.Append(x1, y1, (height1+height2)*m, 0, 0, 1)
			if x1+incX1 < float32(board1.xBounds.upper) && y1+incY1 < float32(board1.yBounds.upper) {
				indices.Append(index, index+1+uint32(float32(board1.xBounds.size())/incX1)+1, index+uint32(float32(board1.xBounds.size())/incX1)+1)
				//indices.Append(index, index+1, index+1+uint32(float32(board1.xBounds.size())/incX1)+1)
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
}

func Move(geom *geometry.Geometry, board1, board2 Board, xBounds, yBounds Bounds, m, prop float32, xDisp, yDisp int) {
	//vbo := geom.GetGeometry().VBO(gls.VertexPosition).Buffer()
	//fmt.Println(vbo.Len())
	incX1 := float32(board1.xBounds.size()) / float32(xBounds.size())
	incX2 := float32(board2.xBounds.size()) / float32(xBounds.size())
	incY1 := float32(board1.yBounds.size()) / float32(yBounds.size())
	incY2 := float32(board2.yBounds.size()) / float32(yBounds.size())
	geom.OperateOnVertices(func(vertex *math32.Vector3) bool {
		x1 := vertex.X + (float32(xDisp) * incX1)
		x2 := (x1 / incX1) * incX2
		y1 := vertex.Y + (float32(yDisp) * incY1)
		y2 := (y1 / incY1) * incY2
		height1 := float32((perlinNoise(board1, x1, y1)) * prop)
		height2 := float32((perlinNoise(board2, x2, y2)) * (1 - prop))
		vertex.Z = (height1 + height2) * m
		return false
	})
}

////////////////////
//======Math======//
////////////////////
func toX(gradient uint8) float32 {
	return float32(math.Cos(float64(gradient) * (math.Pi / 180)))
}

func toY(gradient uint8) float32 {
	return float32(math.Sin(float64(gradient) * (math.Pi / 180)))
}

func perlinNoise(board Board, x, y float32) float32 {
	x0 := int32(math.Floor(float64(x))) - board.xBounds.lower
	x1 := x0 + 1
	y0 := int32(math.Floor(float64(y))) - board.yBounds.lower
	y1 := y0 + 1

	//fmt.Print(x0, y0, ", ", x1, y1)

	sx := (x - float32(board.xBounds.lower)) - float32(x0)
	sy := (y - float32(board.yBounds.lower)) - float32(y0)

	n00x := toX(board.calculateGradient(x0, y0)) * sx
	n00y := toY(board.calculateGradient(x0, y0)) * sy
	n00 := n00x + n00y

	n10x := toX(board.calculateGradient(x1, y0)) * (sx - 1)
	n10y := toY(board.calculateGradient(x1, y0)) * sy
	n10 := n10x + n10y

	n01x := toX(board.calculateGradient(x0, y1)) * sx
	n01y := toY(board.calculateGradient(x0, y1)) * (sy - 1)
	n01 := n01x + n01y

	n11x := toX(board.calculateGradient(x1, y1)) * (sx - 1)
	n11y := toY(board.calculateGradient(x1, y1)) * (sy - 1)
	n11 := n11x + n11y

	fu := float32((6.0 * math.Pow(float64(sx), 5.0)) - (15.0 * math.Pow(float64(sx), 4.0)) + (10.0 * math.Pow(float64(sx), 3.0)))

	nx0 := (n00 * (1 - fu)) + (n10 * fu)
	nx1 := (n01 * (1 - fu)) + (n11 * fu)

	fv := float32((6.0 * math.Pow(float64(sy), 5.0)) - (15.0 * math.Pow(float64(sy), 4.0)) + (10.0 * math.Pow(float64(sy), 3.0)))

	nxy := (nx0 * (1 - fv)) + (nx1 * fv)
	return nxy
}
