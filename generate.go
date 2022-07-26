package main

import (
	"fmt"
	"math"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/math32"
)

////////////////////////////////////////////////////////////////////////////////////////////////
//===========================================Bounds===========================================//
////////////////////////////////////////////////////////////////////////////////////////////////
type Bounds struct {
	lower, upper int32
}

func (bounds *Bounds) size() int {
	return int(math.Abs(float64(bounds.upper) - float64(bounds.lower)))
}

///////////////////////////////////////////////////////////////////////////////////////////////
//=======================================GradientBoard=======================================//
///////////////////////////////////////////////////////////////////////////////////////////////
type GradientBoard struct {
	xBounds Bounds
	yBounds Bounds
	seed    int32
}

/*
 * Calculate gradient is a Board method that will take in integer coordinate points and determine the gradient angle of the board at that location.
 * It performs a hash on the x and y coordinates using the seeds and produces a gradient angle from it. This function is deterministic and well-distributed.
 * @param x the x coordinate to calculate the gradient for
 * @param y the y coordinate to calculate the gradient for
 */
func (board GradientBoard) calculateGradient(x, y int32) uint16 {
	xu := uint(x * board.seed)
	yu := uint(y * board.seed)
	xu = ((xu >> 16) ^ xu) * 0x45d9f3b
	xu = ((xu >> 16) ^ xu) * 0x45d9f3b
	xu = (xu >> 16) ^ xu
	yu = ((yu >> 16) ^ yu) * 0x45d9f3b
	yu = ((yu >> 16) ^ yu) * 0x45d9f3b
	yu = (yu >> 16) ^ yu
	return uint16((31*(31+xu) + yu) % 360)
}

/*
 * This method will take uses an empty *Board and will initialize its bounds about the origin (0, 0) and the seed
 * @param gradientWidth The number of gradients in the X-direction that the Board has centered about the origin
 * @param gradientHeight The number of gradients in the Y-direction that the Board has centered about the origin
 * @seed seed The seed that the board will use to generate it's gradients.
 *
 */
func (board *GradientBoard) initialize(gradientWidth, gradientHeight uint32, seed int32) {
	board.xBounds = Bounds{-int32(gradientWidth) / 2, int32(gradientWidth) / 2}
	board.yBounds = Bounds{-int32(gradientHeight) / 2, int32(gradientHeight) / 2}
	board.seed = seed
}

/*
 * This method uses a *Board and will produce a geometry with a unique vertex buffer object (VBO) with surface triangles rendered within calculated terrain heights using the perlin noise algorithm
 * @param terrainWidth The number of terrain heights in the x-direction to generate
 * @param terrainHeight The number of terrain heights in the y-direction to generate
 * @param m The magnitude or multiplier of the terrain produced
 */
//func (board *GradientBoard) GenerateSurfaceGeometry(terrainWidth, terrainHeight uint16, m float32) *geometry.Geometry {
//	incY := float32(board.yBounds.size()) / float32(terrainHeight)
//	incX := float32(board.xBounds.size()) / float32(terrainWidth)
//	positions := math32.NewArrayF32(0, int(terrainHeight*terrainWidth))
//	indices := math32.NewArrayU32(0, int((terrainHeight-1)*(terrainWidth-1))*2)
//	index := uint32(0)
//	for y := float32(board.yBounds.lower); y <= float32(board.yBounds.upper); y += incY {
//		for x := float32(board.xBounds.lower); x <= float32(board.xBounds.upper); x += incX {
//			if (index)%uint32(terrainWidth) != 0 && (index-1)+uint32(terrainWidth)+1 < uint32(terrainHeight)*uint32(terrainWidth) {
//				indices.Append((index - 1), (index-1)+1, (index-1)+uint32(terrainWidth)+1)
//				//indices.Append((index - 1), (index-1)+uint32(terrainWidth)+2, (index-1)+uint32(terrainWidth+1))
//			}
//			height := float32(perlinNoise(*board, x, y)) * m
//			fmt.Print(x, ", ", y, " ")
//			positions.Append(x, y, height, 0, 0, 1)
//			index++
//		}
//		fmt.Println()
//	}
//	geom := geometry.NewGeometry()
//	geom.SetIndices(indices)
//	geom.AddVBO(gls.NewVBO(positions).
//		AddAttrib(gls.VertexPosition).
//		AddAttrib(gls.VertexNormal),
//	)

//	return geom
//}

func (board *GradientBoard) GenerateSurfaceGeometry(terrainWidth, terrainHeight uint16, m float32) *geometry.Geometry {
	incY := float32(board.yBounds.size()) / float32(terrainHeight)
	incX := float32(board.xBounds.size()) / float32(terrainWidth)
	positions := math32.NewArrayF32(0, int(terrainHeight*terrainWidth))
	indices := math32.NewArrayU32(0, int((terrainHeight-1)*(terrainWidth-1))*2)
	index := uint32(0)
	for y := float32(board.yBounds.lower); y <= float32(board.yBounds.upper); y += incY {
		for x := float32(board.xBounds.lower); x <= float32(board.xBounds.upper); x += incX {
			height := float32(perlinNoise(*board, x, y)) * m
			positions.Append(x, y, height, 0, 0, 1)
			if x+incX <= float32(board.xBounds.upper) && y+incY <= float32(board.yBounds.upper) {
				indices.Append(index, index+uint32(terrainWidth)+2, index+uint32(terrainHeight)+1)
			}
			index++
		}
		fmt.Println()
	}
	geom := geometry.NewGeometry()
	geom.SetIndices(indices)
	geom.AddVBO(gls.NewVBO(positions).
		AddAttrib(gls.VertexPosition).
		AddAttrib(gls.VertexNormal),
	)
	return geom
}

////////////////////////////////////////////////////////////////////////////////////////////////
//======================================BipartiteTerrain======================================//
////////////////////////////////////////////////////////////////////////////////////////////////
/*
 * @var geom The geometry of the bipartite terrain
 * @var board1 The macro board to caculate macro texture gradients
 * @var board2 The micro board to calculate micro texture gradients
 * @var terrainWidth The number of terrain heights in the x-direction to generate
 * @var terrainHeight The number of terrain heights in the y-direction to generate
 * @var m The magnitude or multiplier of the terrain produced
 * @var prop The proportion that describes macro terrain over micro terrain
 */
type BipartiteTerrain struct {
	geom        *geometry.Geometry
	macro       GradientBoard
	micro       GradientBoard
	jointWidth  uint16
	jointHeight uint16
	xDisp       int
	yDisp       int
	m           float32
	prop        float32
}

func (terrain *BipartiteTerrain) initialize(macro, micro GradientBoard, terrainWidth, terrainHeight uint16, m, prop float32) {
	terrain.geom = geometry.NewGeometry()
	terrain.macro = macro
	terrain.micro = micro
	terrain.jointWidth = terrainWidth
	terrain.jointHeight = terrainHeight
	terrain.xDisp = 0
	terrain.yDisp = 0
	terrain.m = m
	terrain.prop = prop
	terrain.GenerateStackedSurfaceGeometry()
}

/*
 * This method takes two boards and will produce a geometry with a unique vertex buffer object (VBO) with surface triangles rendered within calculated net terrain heights.
 * Uses the perlin noise algorithm and the magnitude and proportion metrics to calculate the terrain heights
 */
func (terrain *BipartiteTerrain) GenerateStackedSurfaceGeometry() {
	incY1 := float32(terrain.macro.yBounds.size()) / float32(terrain.jointHeight-1)
	incX1 := float32(terrain.macro.xBounds.size()) / float32(terrain.jointWidth-1)
	incY2 := float32(terrain.micro.yBounds.size()) / float32(terrain.jointHeight-1)
	incX2 := float32(terrain.micro.xBounds.size()) / float32(terrain.jointWidth-1)

	positions := math32.NewArrayF32(0, int(terrain.jointHeight*terrain.jointWidth))
	indices := math32.NewArrayU32(0, int((terrain.jointHeight-1)*(terrain.jointWidth-1)>>1))
	index := uint32(0)

	//for y1, y2 := float32(terrain.macro.yBounds.lower), float32(terrain.micro.yBounds.lower); index < uint32(terrain.jointHeight+1)*uint32(terrain.jointWidth+1); y1, y2 = y1+incY1, y2+incY2 {
	//	for x1, x2 := float32(terrain.macro.xBounds.lower), float32(terrain.micro.xBounds.lower); index%uint32(terrain.jointWidth+1) <= uint32(terrain.jointWidth); x1, x2 = x1+incX1, x2+incX2 {
	//		if (index)%uint32(terrain.jointWidth+1) != 0 {
	//			// && (index-1)+uint32(terrain.jointWidth)+1 < uint32(terrain.jointHeight)*uint32(terrain.jointWidth)
	//			indices.Append((index - 1), (index-1)+1, (index-1)+uint32(terrain.jointWidth)+1)
	//			//indices.Append((index - 1), (index-1)+uint32(terrain.jointWidth)+2, (index-1)+uint32(terrain.jointWidth+1))
	//		}
	//		height1 := float32((perlinNoise(terrain.macro, x1, y1)) * terrain.prop)
	//		height2 := float32((perlinNoise(terrain.micro, x2, y2)) * (1 - terrain.prop))
	//		positions.Append(x1, y1, (height1+height2)*terrain.m, 0, 0, 1)
	//		fmt.Print(index, " ")
	//		index++
	//	}
	//	fmt.Println()
	//}
	for y1, y2 := float32(terrain.macro.yBounds.lower), float32(terrain.micro.yBounds.lower); y1 <= float32(terrain.macro.yBounds.upper); y1, y2 = y1+incY1, y2+incY2 {
		for x1, x2 := float32(terrain.macro.xBounds.lower), float32(terrain.micro.xBounds.lower); x1 <= float32(terrain.macro.xBounds.upper); x1, x2 = x1+incX1, x2+incX2 {
			height1 := float32((perlinNoise(terrain.macro, x1, y1)) * terrain.prop)
			height2 := float32((perlinNoise(terrain.micro, x2, y2)) * (1 - terrain.prop))
			positions.Append(x1, y1, (height1+height2)*terrain.m, 0, 0, 1)
			//if x1+incX1 < float32(terrain.macro.xBounds.upper) && y1+incY1 < float32(terrain.macro.yBounds.upper) {
			//	indices.Append(index, index+uint32(terrain.jointWidth)+1, index+uint32(terrain.jointWidth))
			//}
			if index < (uint32(terrain.jointWidth*terrain.jointHeight)-1)-uint32(terrain.jointWidth) && (index+1)%uint32(terrain.jointWidth) != 0 {
				//indices.Append(index, index+uint32(terrain.jointWidth)+1, index+uint32(terrain.jointWidth))
				indices.Append(index, index+1, index+uint32(terrain.jointWidth))
			}
			index++
		}
		fmt.Println()
	}

	terrain.geom.SetIndices(indices)
	terrain.geom.AddVBO(gls.NewVBO(positions).
		AddAttrib(gls.VertexPosition).
		AddAttrib(gls.VertexNormal),
	)
}

func (terrain *BipartiteTerrain) MoveLeft(amount int) {
	terrain.xDisp = terrain.xDisp + amount
	incX1 := float32(terrain.macro.xBounds.size()) / float32(terrain.jointWidth)
	incX2 := float32(terrain.micro.xBounds.size()) / float32(terrain.jointWidth)
	incY1 := float32(terrain.macro.yBounds.size()) / float32(terrain.jointHeight)
	incY2 := float32(terrain.micro.yBounds.size()) / float32(terrain.jointHeight)
	i := 0
	queue := make([]float32, -amount+1)
	front := 0
	back := 0
	//fmt.Println(-amount)
	terrain.geom.OperateOnVertices(func(vertex *math32.Vector3) bool {
		//fmt.Print(i % int(terrain.jointWidth+1))
		if i%int(terrain.jointWidth) < -amount {
			if i%int(terrain.jointWidth) == 0 {
				back = 0
				//fmt.Println()
			}
			front = 0
			x1 := vertex.X + (float32(terrain.xDisp) * incX1)
			x2 := (x1 / incX1) * incX2
			y1 := vertex.Y + (float32(terrain.yDisp) * incY1)
			y2 := (y1 / incY1) * incY2
			height1 := float32((perlinNoise(terrain.macro, x1, y1)) * terrain.prop)
			height2 := float32((perlinNoise(terrain.micro, x2, y2)) * (1 - terrain.prop))
			queue[back] = vertex.Z
			vertex.Z = (height1 + height2) * terrain.m
			back++
			//fmt.Print(" Generat ")
		} else if back-1 != front {
			temp := vertex.Z
			vertex.Z = queue[front]
			queue[front] = temp
			front += 1
			//fmt.Print(" Passing ")
		} else {
			temp := vertex.Z
			vertex.Z = queue[front]
			queue[front] = temp
			front = 0
			//fmt.Print(" Returns ")
		}
		i++
		return false
	})
}

func (terrain *BipartiteTerrain) MoveRight(amount int) {
	terrain.xDisp = terrain.xDisp + amount
	vbo := terrain.geom.GetGeometry().VBO(gls.VertexPosition).Buffer().ToFloat32()
	incX1 := float32(terrain.macro.xBounds.size()) / float32(terrain.jointWidth)
	incX2 := float32(terrain.micro.xBounds.size()) / float32(terrain.jointWidth)
	incY1 := float32(terrain.macro.yBounds.size()) / float32(terrain.jointHeight)
	incY2 := float32(terrain.micro.yBounds.size()) / float32(terrain.jointHeight)
	i := 0
	terrain.geom.OperateOnVertices(func(vertex *math32.Vector3) bool {
		if (i+amount)%int(terrain.jointWidth) < amount {
			x1 := vertex.X + (float32(terrain.xDisp) * incX1)
			x2 := (x1 / incX1) * incX2
			y1 := vertex.Y + (float32(terrain.yDisp) * incY1)
			y2 := (y1 / incY1) * incY2
			height1 := float32((perlinNoise(terrain.macro, x1, y1)) * terrain.prop)
			height2 := float32((perlinNoise(terrain.micro, x2, y2)) * (1 - terrain.prop))
			vertex.Z = (height1 + height2) * terrain.m
		} else {
			vertex.Z = vbo[((i+amount)*6)+2]
		}
		i++
		return false
	})
}

func (terrain *BipartiteTerrain) MoveDown(amount int) {
	terrain.yDisp = terrain.yDisp + amount
	incX1 := float32(terrain.macro.xBounds.size()) / float32(terrain.jointWidth)
	incX2 := float32(terrain.micro.xBounds.size()) / float32(terrain.jointWidth)
	incY1 := float32(terrain.macro.yBounds.size()) / float32(terrain.jointHeight)
	incY2 := float32(terrain.micro.yBounds.size()) / float32(terrain.jointHeight)
	i := 0
	pivot := 0
	line := make([]float32, int(terrain.jointWidth+1)*(-amount+1), int(terrain.jointWidth+1)*(-amount+1))
	//fmt.Println(-amount)
	terrain.geom.OperateOnVertices(func(vertex *math32.Vector3) bool {
		//fmt.Print(i % int(terrain.jointWidth+1))
		//if i%int(terrain.jointWidth+1) == 0 {
		//	fmt.Println()
		//}
		if i < int(terrain.jointWidth)*(-amount) {
			line[i] = vertex.Z
			x1 := vertex.X + (float32(terrain.xDisp) * incX1)
			x2 := (x1 / incX1) * incX2
			y1 := vertex.Y + (float32(terrain.yDisp) * incY1)
			y2 := (y1 / incY1) * incY2
			height1 := float32((perlinNoise(terrain.macro, x1, y1)) * terrain.prop)
			height2 := float32((perlinNoise(terrain.micro, x2, y2)) * (1 - terrain.prop))
			vertex.Z = (height1 + height2) * terrain.m
			//fmt.Print(" Generat ")
		} else if pivot < int(terrain.jointWidth)*(-amount)-1 {
			temp := vertex.Z
			vertex.Z = line[pivot]
			line[pivot] = temp
			pivot++
			//fmt.Print(" Passing ")
		} else {
			temp := vertex.Z
			vertex.Z = line[pivot]
			line[pivot] = temp
			pivot = 0
			//fmt.Print(" Returns ")
		}
		i++
		return false
	})
}

func (terrain *BipartiteTerrain) MoveUp(amount int) {
	terrain.yDisp = terrain.yDisp + amount
	vbo := terrain.geom.GetGeometry().VBO(gls.VertexPosition).Buffer().ToFloat32()
	incX1 := float32(terrain.macro.xBounds.size()) / float32(terrain.jointWidth)
	incX2 := float32(terrain.micro.xBounds.size()) / float32(terrain.jointWidth)
	incY1 := float32(terrain.macro.yBounds.size()) / float32(terrain.jointHeight)
	incY2 := float32(terrain.micro.yBounds.size()) / float32(terrain.jointHeight)
	i := 0
	terrain.geom.OperateOnVertices(func(vertex *math32.Vector3) bool {
		if i >= int(terrain.jointWidth)*(int(terrain.jointHeight)-amount) {
			x1 := vertex.X + (float32(terrain.xDisp) * incX1)
			x2 := (x1 / incX1) * incX2
			y1 := vertex.Y + (float32(terrain.yDisp) * incY1)
			y2 := (y1 / incY1) * incY2
			height1 := float32((perlinNoise(terrain.macro, x1, y1)) * terrain.prop)
			height2 := float32((perlinNoise(terrain.micro, x2, y2)) * (1 - terrain.prop))
			vertex.Z = (height1 + height2) * terrain.m
		} else {
			vertex.Z = vbo[i*6+(int(terrain.jointWidth)*6*amount)+2]
		}
		i++
		return false
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////
//============================================Math============================================//
////////////////////////////////////////////////////////////////////////////////////////////////

/*
 * Takes in an angle and determines the unit circle x and y displacements given sx and sy
 *
 */
func toXY(gradient uint16, sx, sy float32) (float32, float32) {
	return float32(math.Cos(float64(gradient)*(math.Pi/180))) * sx, float32(math.Sin(float64(gradient)*(math.Pi/180))) * sy
}

/*
 * Perlin noise algorithm will return some float32 (-1.5, 1.5) given some board at an x and y position
 * @param board The board that we are calculating a noise for
 * @param x The x position at which we would like to have a height (Note: x must be within board's xBounds)
 * @param y The y position at which we would like to have a height (Note: y must be within board's yBounds)
 */
func perlinNoise(board GradientBoard, x, y float32) float32 {
	x0 := int32(math.Floor(float64(x))) - board.xBounds.lower
	x1 := x0 + 1
	y0 := int32(math.Floor(float64(y))) - board.yBounds.lower
	y1 := y0 + 1

	sx := (x - float32(board.xBounds.lower)) - float32(x0)
	sy := (y - float32(board.yBounds.lower)) - float32(y0)

	n00x, n00y := toXY(board.calculateGradient(x0, y0), sx, sy)
	n00 := n00x + n00y

	n10x, n10y := toXY(board.calculateGradient(x1, y0), sx-1, sy)
	n10 := n10x + n10y

	n01x, n01y := toXY(board.calculateGradient(x0, y1), sx, sy-1)
	n01 := n01x + n01y

	n11x, n11y := toXY(board.calculateGradient(x1, y1), sx-1, sy-1)
	n11 := n11x + n11y

	fu := float32((6.0 * math.Pow(float64(sx), 5.0)) - (15.0 * math.Pow(float64(sx), 4.0)) + (10.0 * math.Pow(float64(sx), 3.0)))

	nx0 := (n00 * (1 - fu)) + (n10 * fu)
	nx1 := (n01 * (1 - fu)) + (n11 * fu)

	fv := float32((6.0 * math.Pow(float64(sy), 5.0)) - (15.0 * math.Pow(float64(sy), 4.0)) + (10.0 * math.Pow(float64(sy), 3.0)))

	nxy := (nx0 * (1 - fv)) + (nx1 * fv)
	return nxy
}
