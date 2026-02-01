package planetsimulation

import (
	"image/color"
	"math"
)

type vector2 struct {
	x float64
	y float64
}

func (v vector2) normalize() vector2 {
	distance := math.Sqrt(v.x*v.x + v.y*v.y)

	norX := v.x / distance
	norY := v.y / distance

	return vector2{
		norX,
		norY,
	}
}

func (v vector2) add(v1 vector2, v2 vector2) vector2 {
	return vector2{
		v1.x + v2.x,
		v2.y + v2.y,
	}
}

func SetColor(r uint8, g uint8, b uint8, a uint8) color.RGBA {
	color := color.RGBA{}

	color.R = r
	color.G = g
	color.B = b
	color.A = a

	return color
}
