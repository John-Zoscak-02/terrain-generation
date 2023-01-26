package main

import (
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
 * Calculate gradient is a GradientBoard method that will take in integer coordinate points and determine the gradient angle of the board at that location.
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
 * Perlin noise algorithm will return some float32 (-1.5, 1.5) given some board at an x and y position
 * @param x The x position at which we would like to have a height (Note: x must be within board's xBounds)
 * @param y The y position at which we would like to have a height (Note: y must be within board's yBounds)
 */
func (board GradientBoard) perlinNoise(x, y float32) float32 {
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

///////////////////////////////////////////////////////////////////////////////////////////////////////
//=============================================Terrains==============================================//
///////////////////////////////////////////////////////////////////////////////////////////////////////
// Interface for terrains that can be progressivley generated
type Terrain interface {
	GenerateSurfaceGeometry()
	MoveUp(int)
	MoveDown(int)
	MoveLeft(int)
	MoveRight(int)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
//==========================================SimpleTerrain===========================================//
//////////////////////////////////////////////////////////////////////////////////////////////////////
type SimpleTerrain struct {
	// The surface geometry of this terrain
	geoms []*geometry.Geometry
	// The shared vbo
	vbo *gls.VBO
	// The gradient board of this terrain used by the perlin noise generation
	board GradientBoard
	// The number of triangles to be rendered in the x direction of the terrain's geometry
	width uint32
	// The number of triangles to be rendered in the y direction of the terrain's geometry
	height uint32
	// The current displacement from x=0 and y=0 of the rendered terrain
	xDisp int
	yDisp int
	// The magnitude of this terrain
	m float32
}

/*
 * Sets the fields of the simple terrain object to thier default for initial terrain generation
 * @param board The gradient board used to generate the perlin noise textures
 * @param terrainWidth The number of triangles to be rendered in the x direction of the terrain
 * @param terrainHeight The number of triangles to be rendered in the y direction of the terrain
 * @param m The magnitude of the terrain
 */
func (terrain *SimpleTerrain) initialize(board GradientBoard, terrainWidth, terrainHeight uint32, m float32) {
	terrain.geoms = make([]*geometry.Geometry, (terrainWidth-1)*(terrainHeight-1))
	terrain.board = board
	terrain.width = terrainWidth
	terrain.height = terrainHeight
	terrain.xDisp = 0
	terrain.yDisp = 0
	terrain.m = m
	terrain.GenerateSurfaceGeometry()
}

/*
 * Uses a *SimpleTerrain and its fields to:
 *  - produce a geometry with a unique vertex buffer object (VBO) and surface triangles rendered within calculated terrain heights using the perlin noise algorithm.
 *  - set that terrain's geometry's VBO to the generated VBO
 *  - set that terrain's geometry's rendered triangle indicies to the generated triangle indicies list
 */
func (terrain *SimpleTerrain) GenerateSurfaceGeometry() {
	incY := float32(terrain.board.yBounds.size()) / float32(terrain.height-1)
	incX := float32(terrain.board.xBounds.size()) / float32(terrain.width-1)
	terrain.vbo = gls.NewVBO(math32.NewArrayF32(0, int(terrain.height*terrain.width))).
		AddAttrib(gls.VertexPosition)
	index := uint32(0)
	for y := float32(terrain.board.yBounds.lower); y <= float32(terrain.board.yBounds.upper); y += incY {
		for x := float32(terrain.board.xBounds.lower); x <= float32(terrain.board.xBounds.upper); x += incX {
			height := float32(terrain.board.perlinNoise(x, y)) * terrain.m
			terrain.vbo.Buffer().AppendVector3(math32.NewVector3(x, y, height))
			if x+incX <= float32(terrain.board.xBounds.upper) && y+incY <= float32(terrain.board.yBounds.upper) {
				terrain.geoms[index] = geometry.NewGeometry()
				indices := math32.NewArrayU32(0, 3)
				indices.Append(index, index+uint32(terrain.width)+1, index+uint32(terrain.height))
				terrain.geoms[index].SetIndices(indices)
				terrain.geoms[index].AddVBO(terrain.vbo)
				//fmt.Println(indices)
				index++
			}
			//fmt.Print(fmt.Sprintf("i=%d (%2.3f, %2.3f, %2.3f)", index, x, y, height))
		}
		//fmt.Println()
	}

	//fmt.Println(fmt.Sprintf("(%2.3f, %2.3f, %2.3f)", terrain.vbo.Buffer().ToFloat32()[0], terrain.vbo.Buffer().ToFloat32()[1], terrain.vbo.Buffer().ToFloat32()[2]))
	//fmt.Println(fmt.Sprintf("(%2.3f, %2.3f, %2.3f)", terrain.geoms[0].VBO(gls.VertexPosition).Buffer().ToFloat32()[0], terrain.geoms[0].VBO(gls.VertexPosition).Buffer().ToFloat32()[1], terrain.geoms[0].VBO(gls.VertexPosition).Buffer().ToFloat32()[2]))
	//fmt.Println(fmt.Sprintf("(%2.3f, %2.3f, %2.3f)", terrain.geoms[1].VBO(gls.VertexPosition).Buffer().ToFloat32()[0], terrain.geoms[1].VBO(gls.VertexPosition).Buffer().ToFloat32()[1], terrain.geoms[1].VBO(gls.VertexPosition).Buffer().ToFloat32()[2]))

	//terrain.vbo.Buffer().SetVector3(0, math32.NewVector3(1.0, 1.0, 1.0))

	//fmt.Println(fmt.Sprintf("(%2.3f, %2.3f, %2.3f)", terrain.vbo.Buffer().ToFloat32()[0], terrain.vbo.Buffer().ToFloat32()[1], terrain.vbo.Buffer().ToFloat32()[2]))
	//fmt.Println(fmt.Sprintf("(%2.3f, %2.3f, %2.3f)", terrain.geoms[0].VBO(gls.VertexPosition).Buffer().ToFloat32()[0], terrain.geoms[0].VBO(gls.VertexPosition).Buffer().ToFloat32()[1], terrain.geoms[0].VBO(gls.VertexPosition).Buffer().ToFloat32()[2]))
	//fmt.Println(fmt.Sprintf("(%2.3f, %2.3f, %2.3f)", terrain.geoms[1].VBO(gls.VertexPosition).Buffer().ToFloat32()[0], terrain.geoms[1].VBO(gls.VertexPosition).Buffer().ToFloat32()[1], terrain.geoms[1].VBO(gls.VertexPosition).Buffer().ToFloat32()[2]))
}

/*
 * Iterates over each vertex in the VBO of the Terrain's geometry and will determine new terrain surface heights when displaced from the current configuration by the amount parameter in the -x direction
 */
func (terrain *SimpleTerrain) MoveLeft(amount int) {
	terrain.xDisp = terrain.xDisp + amount
	incX := float32(terrain.board.xBounds.size()) / float32(terrain.width)
	incY := float32(terrain.board.yBounds.size()) / float32(terrain.height)
	i := 0
	queue := make([]float32, -amount+1)
	front := 0
	back := 0
	terrain.vbo.OperateOnVectors3(gls.VertexPosition, func(vertex *math32.Vector3) bool {
		if i%int(terrain.width) < -amount {
			front = 0
			x := vertex.X + (float32(terrain.xDisp) * incX)
			y := vertex.Y + (float32(terrain.yDisp) * incY)
			height := float32(terrain.board.perlinNoise(x, y))
			queue[back] = vertex.Z
			vertex.Z = height * terrain.m
			back++
		} else if -amount-1 != front {
			temp := vertex.Z
			vertex.Z = queue[front]
			queue[front] = temp
			front += 1
			back = 0
		} else {
			temp := vertex.Z
			vertex.Z = queue[front]
			queue[front] = temp
			front = 0
			back = 0
		}
		i++
		return false
	})
}

/*
 * Iterates over each vertex in the VBO of the Terrain's geometry and will determine new terrain surface heights when displaced from the current configuration by the amount parameter in the +x direction
 */
func (terrain *SimpleTerrain) MoveRight(amount int) {
	vbo := terrain.vbo.Buffer().ToFloat32()
	terrain.xDisp = terrain.xDisp + amount
	incX := float32(terrain.board.xBounds.size()) / float32(terrain.width)
	incY := float32(terrain.board.yBounds.size()) / float32(terrain.height)
	i := 0
	terrain.vbo.OperateOnVectors3(gls.VertexPosition, func(vertex *math32.Vector3) bool {
		if (i+amount)%int(terrain.width) < amount {
			x := vertex.X + (float32(terrain.xDisp) * incX)
			y := vertex.Y + (float32(terrain.yDisp) * incY)
			height := float32(terrain.board.perlinNoise(x, y))
			vertex.Z = height * terrain.m
		} else {
			vertex.Z = vbo[((i+amount)*6)+2]
		}
		i++
		return false
	})
}

/*
 * Iterates over each vertex in the VBO of the Terrain's geometry and will determine new terrain surface heights when displaced from the current configuration by the amount parameter in the -y direction
 */
func (terrain *SimpleTerrain) MoveDown(amount int) {
	terrain.yDisp = terrain.yDisp + amount
	incX := float32(terrain.board.xBounds.size()) / float32(terrain.width)
	incY := float32(terrain.board.yBounds.size()) / float32(terrain.height)
	i := 0
	pivot := 0
	queue := make([]float32, int(terrain.width+1)*(-amount+1))
	terrain.vbo.OperateOnVectors3(gls.VertexPosition, func(vertex *math32.Vector3) bool {
		if i < int(terrain.width)*(-amount) {
			queue[i] = vertex.Z
			x := vertex.X + (float32(terrain.xDisp) * incX)
			y := vertex.Y + (float32(terrain.yDisp) * incY)
			height := float32(terrain.board.perlinNoise(x, y))
			vertex.Z = height * terrain.m
		} else if pivot < int(terrain.width)*(-amount)-1 {
			temp := vertex.Z
			vertex.Z = queue[pivot]
			queue[pivot] = temp
			pivot++
		} else {
			temp := vertex.Z
			vertex.Z = queue[pivot]
			queue[pivot] = temp
			pivot = 0
		}
		i++
		return false
	})
}

/*
 * Iterates over each vertex in the VBO of the Terrain's geometry and will determine new terrain surface heights when displaced from the current configuration by the amount parameter in the +y direction
 */
func (terrain *SimpleTerrain) MoveUp(amount int) {
	terrain.yDisp = terrain.yDisp + amount
	vbo := terrain.vbo.Buffer().ToFloat32()
	incX := float32(terrain.board.xBounds.size()) / float32(terrain.width)
	incY := float32(terrain.board.yBounds.size()) / float32(terrain.height)
	i := 0
	terrain.vbo.OperateOnVectors3(gls.VertexPosition, func(vertex *math32.Vector3) bool {
		if i >= int(terrain.width)*(int(terrain.height)-amount) {
			x := vertex.X + (float32(terrain.xDisp) * incX)
			y := vertex.Y + (float32(terrain.yDisp) * incY)
			height := float32(terrain.board.perlinNoise(x, y))
			vertex.Z = height * terrain.m
		} else {
			vertex.Z = vbo[(i+(int(terrain.width)*amount))*6+2]
		}
		i++
		return false
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////
//======================================BipartiteTerrain======================================//
////////////////////////////////////////////////////////////////////////////////////////////////
type BipartiteTerrain struct {
	// The surface geometry of this terrain
	geom *geometry.Geometry
	// The macro textures gradient board of this terrain used by the perlin noise generation
	macro GradientBoard
	// The micro textures gradient board of this terrain used by the perlin noise generation
	micro GradientBoard
	// The number of triangles rendered in the x direction of this terrain
	jointWidth uint32
	// The number of triangles rendered inthe y direction of this terrain
	jointHeight uint32
	// The current displacement from x=0 and y=0 of the rendered terrain
	xDisp int
	yDisp int
	// The magnitude of this terrain
	m float32
	// The effect of the macro texture generation on the surface geometry opposed to the effect of the micro texture generation
	prop float32
}

/*
 * Sets the fields of the bipartite terrain object to thier default for initial terrain generation
 * @param macro The gradient board used to generate the macro perlin noise textures
 * @param micro The gradient board used to generate the micro perlin noise textures
 * @param terrainWidth The number of triangles to be rendered in the x direction of the terrain
 * @param terrainHeight The number of triangles to be rendered in the y direction of the terrain
 * @param m The magnitude of the terrain
 * @param prop The effect of the macro texture generation on the surface geometry opposed to the effect of the micro texture generation
 */
func (terrain *BipartiteTerrain) initialize(macro, micro GradientBoard, terrainWidth, terrainHeight uint32, m, prop float32) {
	terrain.geom = geometry.NewGeometry()
	terrain.macro = macro
	terrain.micro = micro
	terrain.jointWidth = terrainWidth
	terrain.jointHeight = terrainHeight
	terrain.xDisp = 0
	terrain.yDisp = 0
	terrain.m = m
	terrain.prop = prop
	terrain.GenerateSurfaceGeometry()
}

/*
 * Uses a *BipartiteTerrain and its fields to:
 *  - produce a geometry with a unique vertex buffer object (VBO) and surface triangles rendered within calculated terrain heights using the perlin noise algorithm.
 *  - set that terrain's geometry's VBO to the generated VBO
 *  - set that terrain's geometry's rendered triangle indicies to the generated triangle indicies list
 */
func (terrain *BipartiteTerrain) GenerateSurfaceGeometry() {
	incY1 := float32(terrain.macro.yBounds.size()) / float32(terrain.jointHeight-1)
	incX1 := float32(terrain.macro.xBounds.size()) / float32(terrain.jointWidth-1)
	incY2 := float32(terrain.micro.yBounds.size()) / float32(terrain.jointHeight-1)
	incX2 := float32(terrain.micro.xBounds.size()) / float32(terrain.jointWidth-1)

	positions := math32.NewArrayF32(0, int(terrain.jointHeight*terrain.jointWidth))
	indices := math32.NewArrayU32(0, int((terrain.jointHeight-1)*(terrain.jointWidth-1)>>1))
	index := uint32(0)

	for y1, y2 := float32(terrain.macro.yBounds.lower), float32(terrain.micro.yBounds.lower); y1 <= float32(terrain.macro.yBounds.upper); y1, y2 = y1+incY1, y2+incY2 {
		for x1, x2 := float32(terrain.macro.xBounds.lower), float32(terrain.micro.xBounds.lower); x1 <= float32(terrain.macro.xBounds.upper); x1, x2 = x1+incX1, x2+incX2 {
			height1 := float32((terrain.macro.perlinNoise(x1, y1)) * terrain.prop)
			height2 := float32((terrain.micro.perlinNoise(x2, y2)) * (1 - terrain.prop))
			positions.Append(x1, y1, (height1+height2)*terrain.m, 0, 0, 1)
			if index < (uint32(terrain.jointWidth*terrain.jointHeight)-1)-uint32(terrain.jointWidth) && (index+1)%uint32(terrain.jointWidth) != 0 {
				indices.Append(index, index+1, index+uint32(terrain.jointWidth)+1)
			}
			index++
		}
	}

	terrain.geom.SetIndices(indices)
	terrain.geom.AddVBO(gls.NewVBO(positions).
		AddAttrib(gls.VertexPosition).
		AddAttrib(gls.VertexNormal),
	)
}

/*
 * Iterates over each vertex in the VBO of the Terrain's geometry and will determine new terrain surface heights when displaced from the current configuration by the amount parameter in the -x direction
 */
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
	terrain.geom.OperateOnVertices(func(vertex *math32.Vector3) bool {
		if i%int(terrain.jointWidth) < -amount {
			front = 0
			x1 := vertex.X + (float32(terrain.xDisp) * incX1)
			x2 := (x1 / incX1) * incX2
			y1 := vertex.Y + (float32(terrain.yDisp) * incY1)
			y2 := (y1 / incY1) * incY2
			height1 := float32((terrain.macro.perlinNoise(x1, y1)) * terrain.prop)
			height2 := float32((terrain.micro.perlinNoise(x2, y2)) * (1 - terrain.prop))
			queue[back] = vertex.Z
			vertex.Z = (height1 + height2) * terrain.m
			back++
		} else if -amount-1 != front {
			temp := vertex.Z
			vertex.Z = queue[front]
			queue[front] = temp
			front += 1
			back = 0
		} else {
			temp := vertex.Z
			vertex.Z = queue[front]
			queue[front] = temp
			front = 0
			back = 0
		}
		i++
		return false
	})
}

/*
 * Iterates over each vertex in the VBO of the Terrain's geometry and will determine new terrain surface heights when displaced from the current configuration by the amount parameter in the +x direction
 */
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
			height1 := float32((terrain.macro.perlinNoise(x1, y1)) * terrain.prop)
			height2 := float32((terrain.micro.perlinNoise(x2, y2)) * (1 - terrain.prop))
			vertex.Z = (height1 + height2) * terrain.m
		} else {
			vertex.Z = vbo[((i+amount)*6)+2]
		}
		i++
		return false
	})
}

/*
 * Iterates over each vertex in the VBO of the Terrain's geometry and will determine new terrain surface heights when displaced from the current configuration by the amount parameter in the -y direction
 */
func (terrain *BipartiteTerrain) MoveDown(amount int) {
	terrain.yDisp = terrain.yDisp + amount
	incX1 := float32(terrain.macro.xBounds.size()) / float32(terrain.jointWidth)
	incX2 := float32(terrain.micro.xBounds.size()) / float32(terrain.jointWidth)
	incY1 := float32(terrain.macro.yBounds.size()) / float32(terrain.jointHeight)
	incY2 := float32(terrain.micro.yBounds.size()) / float32(terrain.jointHeight)
	i := 0
	pivot := 0
	queue := make([]float32, int(terrain.jointWidth+1)*(-amount+1))
	terrain.geom.OperateOnVertices(func(vertex *math32.Vector3) bool {
		if i < int(terrain.jointWidth)*(-amount) {
			queue[i] = vertex.Z
			x1 := vertex.X + (float32(terrain.xDisp) * incX1)
			x2 := (x1 / incX1) * incX2
			y1 := vertex.Y + (float32(terrain.yDisp) * incY1)
			y2 := (y1 / incY1) * incY2
			height1 := float32((terrain.macro.perlinNoise(x1, y1)) * terrain.prop)
			height2 := float32((terrain.micro.perlinNoise(x2, y2)) * (1 - terrain.prop))
			vertex.Z = (height1 + height2) * terrain.m
		} else if pivot < int(terrain.jointWidth)*(-amount)-1 {
			temp := vertex.Z
			vertex.Z = queue[pivot]
			queue[pivot] = temp
			pivot++
		} else {
			temp := vertex.Z
			vertex.Z = queue[pivot]
			queue[pivot] = temp
			pivot = 0
		}
		i++
		return false
	})
}

/*
 * Iterates over each vertex in the VBO of the Terrain's geometry and will determine new terrain surface heights when displaced from the current configuration by the amount parameter in the +y direction
 */
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
			height1 := float32((terrain.macro.perlinNoise(x1, y1)) * terrain.prop)
			height2 := float32((terrain.micro.perlinNoise(x2, y2)) * (1 - terrain.prop))
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
