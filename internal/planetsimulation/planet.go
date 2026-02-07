package planetsimulation

import (
	"image/color"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Planet struct {
	x               float64
	y               float64
	offset          []float64
	radius          float64
	velocity        vector2
	mass            float64
	color           color.Color
	image           *ebiten.Image
	geometry        ebiten.GeoM
	traces          [][]int
	traceWidth      float64
	antialiasTraces bool
	tickCount       int
	traceEveryNTick int // every Nth tick
	drawEveryNTick  int
	isFocused       bool
}

func (p *Planet) translate(dx float64, dy float64) {
	p.x += dx
	p.y += dy

	p.geometry.Translate(dx, dy)
}

func (p *Planet) setPosition(x float64, y float64) {
	p.x = x
	p.y = y

	p.geometry.Reset()
	// center circle
	p.geometry.Translate(x-p.radius, y-p.radius)
	// adjust for offset
	p.geometry.Translate(p.offset[0], p.offset[1])
}

func (p *Planet) updateImage() {
	p.geometry.Reset()
	p.setPosition(p.x, p.y)
	radius := float32(p.radius)
	p.image = ebiten.NewImage(int(radius*2), int(radius*2))
	vector.FillCircle(p.image, radius, radius, radius, p.color, true)
}

func (p *Planet) getColor() (int, int, int) {
	r, g, b, _ := p.color.RGBA()

	r8 := int(r >> 8)
	g8 := int(g >> 8)
	b8 := int(b >> 8)

	return r8, g8, b8
}

func (p *Planet) changeColor(colorDelta ColorDelta) {
	r, g, b, a := p.color.RGBA()

	if colorDelta.R != nil {
		r = uint32(*colorDelta.R)
	}

	if colorDelta.G != nil {
		g = uint32(*colorDelta.G)
	}

	if colorDelta.B != nil {
		b = uint32(*colorDelta.B)
	}

	if colorDelta.A != nil {
		a = uint32(*colorDelta.A)
	}

	p.color = SetColor(
		uint8(r),
		uint8(g),
		uint8(b),
		uint8(a),
	)

	p.updateImage()
}

func (p *Planet) clearTraces() {
	p.traces = slices.Delete(p.traces, 0, len(p.traces))
}

func (p *Planet) focus(sim *simulation) {
	sim.returnToOrigin()

	// move to planet
	focusedPlanet := sim.planets[sim.focusedPlanetIndex]
	planetDx := focusedPlanet.x
	planetDy := focusedPlanet.y
	sim.planetsOffset[0] -= planetDx
	sim.planetsOffset[1] -= planetDy

	for _, planet := range sim.planets {
		planet.geometry.Translate(-planetDx, -planetDy)
	}

	p.updateImage()
}

func newPlanet(x float64, y float64, radius float64, mass float64, velocity vector2, color color.Color, offset []float64) *Planet {
	p := Planet{}
	p.image = ebiten.NewImage(int(radius*2), int(radius*2))
	p.color = color
	p.radius = radius
	p.velocity = velocity
	p.mass = mass
	p.offset = offset
	p.geometry = ebiten.GeoM{}

	// adjust for center and screen offset
	p.setPosition(x, y)
	p.updateImage()

	p.antialiasTraces = false
	p.traceEveryNTick = 5
	p.drawEveryNTick = 1
	p.traceWidth = 1.5

	return &p
}

func (p *Planet) handleFocusedPlanet(sim *simulation, dx float64, dy float64) {
	if sim.isPlanetFocused {
		sim.planetsOffset[0] += dx
		sim.planetsOffset[1] += dy
	}
}

func (p *Planet) Update(sim *simulation, planets []*Planet) {
	if sim.running {
		p.handleGravitation(sim)
	}
}

func mergePlanets(sim *simulation, p *Planet, otherPlanet *Planet) {
	// merge planets
	if p.mass >= otherPlanet.mass {
		sim.planetsToRemove = append(sim.planetsToRemove, slices.Index(sim.planets, otherPlanet))
		p.mass += otherPlanet.mass / 2
		if p.radius <= 1000 {
			p.radius += otherPlanet.radius / 4
		}

		p.velocity = p.velocity.add(vector2{
			((otherPlanet.velocity.x) / p.mass),
			((otherPlanet.velocity.y) / p.mass),
		})
		p.updateImage()
	}
}

func (p *Planet) handleGravitation(sim *simulation) {
	forces := make([]vector2, 0)

	for i := 0; i < len(sim.planets); i++ {
		otherPlanet := sim.planets[i]

		if slices.Contains(sim.planetsToRemove, i) || otherPlanet == p {
			continue
		}

		// calculate distance
		dx, dy, distance, overlaps := overlapsCircle(otherPlanet.x, p.x, otherPlanet.y, p.y, otherPlanet.radius, p.radius)
		if overlaps {
			mergePlanets(sim, p, otherPlanet)
		}

		force := vector2{
			x: dx,
			y: dy,
		}

		forceAmount := sim.gravitationalConstant * ((p.mass * otherPlanet.mass) / math.Pow(distance, 2))

		norForce := force.normalize()
		force.x = norForce.x * forceAmount
		force.y = norForce.y * forceAmount

		forces = append(forces, vector2{force.x, force.y})
	}

	// add all forces
	resultingForce := vector2{0, 0}

	for _, force := range forces {
		resultingForce.x += force.x
		resultingForce.y += force.y
	}

	// F = m * a
	// calculate acceleration
	acceleration := vector2{
		x: resultingForce.x / p.mass,
		y: resultingForce.y / p.mass,
	}

	// v = a * t
	// calculate new velocity
	time := 1 / ebiten.ActualFPS()

	newVelocity := vector2{
		x: acceleration.x * time,
		y: acceleration.y * time,
	}

	// add velocity
	p.velocity = p.velocity.add(newVelocity)

	// adjust for timestep
	dx := p.velocity.x * time
	dy := p.velocity.y * time

	p.translate(dx, dy)

	// trace ticks
	for p.tickCount >= p.traceEveryNTick {
		tracePosition := []int{
			int(p.x),
			int(p.y),
		}

		p.traces = append(p.traces, tracePosition)

		p.tickCount -= p.traceEveryNTick
	}

	p.tickCount++
}

func (p *Planet) Draw(screen *ebiten.Image) {
	screen.DrawImage(p.image, &ebiten.DrawImageOptions{
		GeoM: p.geometry,
	})

	for i := 0; i < len(p.traces); i++ {
		if i == len(p.traces)-1 {
			continue
		}

		if i%p.drawEveryNTick != 0 {
			continue
		}

		currentTrace := p.traces[i]
		nextTrace := p.traces[i+1]

		// adjust for offset
		vector.StrokeLine(
			screen,
			float32(float64(currentTrace[0])+float64(p.offset[0])),
			float32(float64(currentTrace[1])+float64(p.offset[1])),
			float32(float64(nextTrace[0])+float64(p.offset[0])),
			float32(float64(nextTrace[1])+float64(p.offset[1])),
			float32(p.traceWidth), p.color, p.antialiasTraces,
		)
	}
}
