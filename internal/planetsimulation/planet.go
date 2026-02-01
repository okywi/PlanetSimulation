package planetsimulation

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Planet struct {
	x               float64
	y               float64
	offset          []int
	radius          float32
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
}

func (p *Planet) translate(dx float64, dy float64) {
	p.x += dx
	p.y += dy

	p.geometry.Translate(dx, dy)
}

func (p *Planet) setPositionOnStart(x float64, y float64, offsetX int, offsetY int) {
	p.x = x
	p.y = y

	// center circle
	p.geometry.Translate(x-float64(p.radius), y-float64(p.radius))
	// adjust for offset
	p.geometry.Translate(float64(offsetX), float64(offsetY))
}

func (p *Planet) updateImage() {
	vector.FillCircle(p.image, p.radius, p.radius, p.radius, p.color, true)
}

func createPlanet(x float64, y float64, radius float32, mass float64, velocity vector2, color color.Color, offset []int) *Planet {
	p := Planet{}

	p.image = ebiten.NewImage(int(radius*2), int(radius*2))
	p.color = color
	p.radius = radius
	p.velocity = velocity
	p.mass = mass

	vector.FillCircle(p.image, radius, radius, radius, p.color, true)

	p.geometry = ebiten.GeoM{}
	// adjust for center and offset
	p.offset = offset
	p.setPositionOnStart(x, y, p.offset[0], p.offset[1])

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
