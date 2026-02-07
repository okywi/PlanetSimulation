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

func (v vector2) add(v2 vector2) vector2 {
	return vector2{
		v.x + v2.x,
		v.y + v2.y,
	}
}

type ColorDelta struct {
	R, G, B, A *uint8
}

func SetColor(r uint8, g uint8, b uint8, a uint8) color.NRGBA {
	color := color.NRGBA{
		R: r, G: g, B: b, A: a,
	}

	return color
}

func convertColorToInt(color color.Color) (int, int, int, int) {
	r, g, b, a := color.RGBA()

	r8 := int(r >> 8)
	g8 := int(g >> 8)
	b8 := int(b >> 8)
	a8 := int(a >> 8)

	return r8, g8, b8, a8
}

func overlapsXY(x int, y int, xleft int, xright, ytop int, ybottom int) bool {
	overlapsX := false
	overlapsY := false

	if x >= xleft && x <= xright {
		overlapsX = true
	}

	if y >= ytop && y <= ybottom {
		overlapsY = true
	}

	if overlapsX && overlapsY {
		return true
	}
	return false
}

func overlapsCircle(x1 float64, x2 float64, y1 float64, y2 float64, radius1 float64, radius2 float64) (float64, float64, float64, bool) {
	overlaps := false

	// calculate distance
	dx, dy := x1-x2, y1-y2
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance <= radius1+radius2 {
		overlaps = true
	}
	return dx, dy, distance, overlaps
}
