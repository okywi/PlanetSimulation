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
	offset          []int
	radius          float64
	velocity        vector2
	mass            float64
	color           color.Color
	image           *ebiten.Image
	geometry        ebiten.GeoM
	traces          [][]int
	traceWidth      float32
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

func (p *Planet) setPositionOnStart(x float64, y float64, offset []int) {
	p.x = x
	p.y = y

	// center circle
	p.geometry.Translate(x-float64(p.radius), y-float64(p.radius))
	// adjust for offset
	p.geometry.Translate(float64(offset[0]), float64(offset[1]))
}

func (p *Planet) updateImage() {
	p.geometry.Reset()
	p.setPositionOnStart(p.x, p.y, p.offset)
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

func (p *Planet) changeColor(nR int, nG int, nB int) {
	r, g, b, a := p.color.RGBA()

	if nR != -1 {
		r = uint32(nR)
	}

	if nG != -1 {
		g = uint32(nG)
	}

	if nB != -1 {
		b = uint32(nB)
	}

	p.color = SetColor(uint8(r), uint8(g), uint8(b), uint8(a))
	p.updateImage()
}

func (p *Planet) clearTraces() {
	p.traces = slices.Delete(p.traces, 0, len(p.traces))
}

func (p *Planet) focus(sim *simulation) {
	sim.focusedPlanet = sim.selectedPlanet

	dx := int(sim.focusedPlanet.x) - p.offset[0] + sim.screen.offset[0]
	dy := int(sim.focusedPlanet.y) - p.offset[1] + sim.screen.offset[1]

	p.offset[0] += dx
	p.offset[1] += dy
}

func createPlanet(x float64, y float64, radius float64, mass float64, velocity vector2, color color.Color, offset []int) *Planet {
	p := Planet{}

	p.image = ebiten.NewImage(int(radius*2), int(radius*2))
	p.color = color
	p.radius = radius
	p.velocity = velocity
	p.mass = mass

	p.geometry = ebiten.GeoM{}
	// adjust for center and offset
	p.offset = offset
	p.setPositionOnStart(x, y, p.offset)
	p.updateImage()

	p.antialiasTraces = false
	p.traceEveryNTick = 5
	p.drawEveryNTick = 1
	p.traceWidth = 1.5

	return &p
}

func (p *Planet) Update(sim *simulation, planets []*Planet) error {
	if sim.running {
		p.handleGravitation(sim, planets)
	}

	return nil
}

func (p *Planet) handleGravitation(sim *simulation, planets []*Planet) {
	// dont calculate if only one planet
	if len(planets) <= 1 {
		return
	}

	forces := make([]vector2, 0)

	for i := 0; i < len(planets); i++ {
		otherPlanet := planets[i]

		if otherPlanet == p {
			continue
		}

		// calculate distance
		dx, dy := otherPlanet.x-p.x, otherPlanet.y-p.y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance <= float64(p.radius)+float64(otherPlanet.radius) {
			// merge planets
			//p.mass += otherPlanet.mass / 2
			//p.radius += otherPlanet.radius / 2
			//planets = append(planets[:i], planets[i+1:]...)
			//return
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
	time := 1 / ebiten.ActualTPS()

	newVelocity := vector2{
		x: acceleration.x * time,
		y: acceleration.y * time,
	}

	// add velocity
	p.velocity.x = p.velocity.x + newVelocity.x
	p.velocity.y = p.velocity.y + newVelocity.y

	// adjust for timestep
	dx := p.velocity.x * time
	dy := p.velocity.y * time

	p.translate(dx, dy)

	// trace ticks
	if p.tickCount >= p.traceEveryNTick {
		tracePosition := []int{
			int(p.x),
			int(p.y),
		}

		p.traces = append(p.traces, tracePosition)
		p.tickCount = 0
	}

	p.tickCount += 1
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
			p.traceWidth, p.color, p.antialiasTraces,
		)

	}
}
