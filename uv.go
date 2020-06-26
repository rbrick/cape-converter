package main

import "math"

const (
	CapeWidth         = 10
	CapeHeight        = 16
	CapeDepth         = 1
	CapeTextureWidth  = 64
	CapeTextureHeight = 32
)

var (
	Front  = UVCoords(1, 1, CapeTextureWidth, CapeTextureHeight, CapeWidth, CapeHeight-CapeDepth)
	Back   = UVCoords(CapeWidth+CapeDepth+CapeDepth, CapeDepth, CapeTextureWidth, CapeTextureHeight, CapeWidth, CapeHeight-CapeDepth)
	Left   = UVCoords(0, CapeDepth, CapeTextureWidth, CapeTextureHeight, CapeDepth, CapeHeight-CapeHeight)
	Right  = UVCoords(CapeWidth+CapeDepth, CapeDepth, CapeTextureWidth, CapeTextureHeight, CapeDepth, CapeHeight-CapeDepth)
	Top    = UVCoords(CapeDepth, 0, CapeTextureWidth, CapeTextureHeight, CapeWidth, CapeDepth)
	Bottom = UVCoords(CapeDepth+CapeWidth, 0, CapeTextureWidth, CapeTextureHeight, CapeWidth, CapeDepth)

	Faces = map[string][]float64{
		"front":  Front,
		"back":   Back,
		"left":   Left,
		"right":  Right,
		"top":    Top,
		"bottom": Bottom}
)

func Face(face string, width, height int, f func(w, h int, a []float64)) {
	f(width, height, Faces[face])
}

func lerp(a, b, f float64) float64 {
	return a + f*(b-a)
}

func inverseLerp(min, max, value float64) float64 {
	if math.Abs(max-min) < math.SmallestNonzeroFloat64 {
		return min
	}
	return (value - min) / (max - min)
}

func UVCoords(x, y, textureWidth, textureHeight, width, height float64) []float64 {
	scaleWidth := 1 / textureWidth
	scaledHeight := 1 / textureHeight

	xmin := inverseLerp(0, scaleWidth, x/(textureWidth))
	xmax := inverseLerp(0, scaleWidth, (x+width)/textureWidth)

	ymin := inverseLerp(0, scaledHeight, y/textureHeight)
	ymax := inverseLerp(0, scaledHeight, (y+height)/textureHeight)

	return []float64{
		lerp(xmin, xmax, 0) / textureWidth, lerp(ymin, ymax, 0) / textureHeight,
		lerp(xmin, xmax, 1) / textureWidth, lerp(ymin, ymax, 0) / textureHeight,
		lerp(xmin, xmax, 1) / textureWidth, lerp(ymin, ymax, 1) / textureHeight,
		lerp(xmin, xmax, 0) / textureWidth, lerp(ymin, ymax, 1) / textureHeight,
	}
}
