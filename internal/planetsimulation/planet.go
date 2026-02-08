package planetsimulation

import (
	"image/color"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Planet struct {
	Name            string
	HasNameChanged  bool
	X               float64
	Y               float64
	Offset          []float64
	Radius          float64
	Velocity        vector2 `json:"velocity"`
	Mass            float64
	Color           color.NRGBA
	image           *ebiten.Image
	geometry        ebiten.GeoM
	traces          [][]int
	TraceWidth      float64
	AntialiasTraces bool
	TickCount       int
	TraceEveryNTick int // every Nth tick
	DrawEveryNTick  int
	isFocused       bool
}

func (p *Planet) translate(dx float64, dy float64) {
	p.X += dx
	p.Y += dy

	p.geometry.Translate(dx, dy)
}

func (p *Planet) setPosition(x float64, y float64) {
	p.X = x
	p.Y = y

	p.geometry.Reset()
	// center circle
	p.geometry.Translate(x-p.Radius, y-p.Radius)
	// adjust for offset
	p.geometry.Translate(p.Offset[0], p.Offset[1])
}

func (p *Planet) updateImage() {
	p.geometry.Reset()
	p.setPosition(p.X, p.Y)
	radius := float32(p.Radius)
	p.image = ebiten.NewImage(int(radius*2), int(radius*2))
	vector.FillCircle(p.image, radius, radius, radius, p.Color, true)
}

func (p *Planet) getColor() (int, int, int) {
	r, g, b, _ := p.Color.RGBA()

	r8 := int(r >> 8)
	g8 := int(g >> 8)
	b8 := int(b >> 8)

	return r8, g8, b8
}

func (p *Planet) changeColor(colorDelta ColorDelta) {
	r, g, b, a := p.Color.RGBA()

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

	p.Color = SetColor(
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

	if sim.focusedPlanetIndex >= len(sim.planets) {
		return
	}

	// move to planet
	focusedPlanet := sim.planets[sim.focusedPlanetIndex]
	planetDx := focusedPlanet.X
	planetDy := focusedPlanet.Y
	sim.planetsOffset[0] -= planetDx
	sim.planetsOffset[1] -= planetDy

	for _, planet := range sim.planets {
		planet.geometry.Translate(-planetDx, -planetDy)
	}

	p.updateImage()
}

func newPlanet(name string, x float64, y float64, radius float64, mass float64, velocity vector2, color color.NRGBA, offset []float64) *Planet {
	p := Planet{}
	p.Name = name
	p.image = ebiten.NewImage(int(radius*2), int(radius*2))
	p.Color = color
	p.Radius = radius
	p.Velocity = velocity
	p.Mass = mass
	p.Offset = offset
	p.geometry = ebiten.GeoM{}

	// adjust for center and screen offset
	p.setPosition(x, y)
	p.updateImage()

	p.AntialiasTraces = false
	p.TraceEveryNTick = 5
	p.DrawEveryNTick = 1
	p.TraceWidth = 1.5

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
	if p.Mass >= otherPlanet.Mass {
		sim.planetsToRemove = append(sim.planetsToRemove, slices.Index(sim.planets, otherPlanet))
		p.Mass += otherPlanet.Mass / 2
		if p.Radius <= 1000 {
			p.Radius += otherPlanet.Radius / 4
		}

		p.Velocity = p.Velocity.add(vector2{
			((otherPlanet.Velocity.X) / p.Mass),
			((otherPlanet.Velocity.Y) / p.Mass),
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
		dx, dy, distance, overlaps := overlapsCircle(otherPlanet.X, p.X, otherPlanet.Y, p.Y, otherPlanet.Radius, p.Radius)
		if overlaps {
			mergePlanets(sim, p, otherPlanet)
		}

		force := vector2{
			X: dx,
			Y: dy,
		}

		forceAmount := sim.gravitationalConstant * ((p.Mass * otherPlanet.Mass) / math.Pow(distance, 2))

		norForce := force.normalize()
		force.X = norForce.X * forceAmount
		force.Y = norForce.Y * forceAmount

		forces = append(forces, vector2{force.X, force.Y})
	}

	// add all forces
	resultingForce := vector2{0, 0}

	for _, force := range forces {
		resultingForce.X += force.X
		resultingForce.Y += force.Y
	}

	// F = m * a
	// calculate acceleration
	acceleration := vector2{
		X: resultingForce.X / p.Mass,
		Y: resultingForce.Y / p.Mass,
	}

	// v = a * t
	// calculate new velocity
	time := 1 / ebiten.ActualFPS()

	newVelocity := vector2{
		X: acceleration.X * time,
		Y: acceleration.Y * time,
	}

	// add velocity
	p.Velocity = p.Velocity.add(newVelocity)

	// adjust for timestep
	dx := p.Velocity.X * time
	dy := p.Velocity.Y * time

	p.translate(dx, dy)

	// trace ticks
	for p.TickCount >= p.TraceEveryNTick {
		tracePosition := []int{
			int(p.X),
			int(p.Y),
		}

		p.traces = append(p.traces, tracePosition)

		p.TickCount -= p.TraceEveryNTick
	}

	p.TickCount++
}

func (p *Planet) Draw(screen *ebiten.Image) {
	screen.DrawImage(p.image, &ebiten.DrawImageOptions{
		GeoM: p.geometry,
	})

	for i := 0; i < len(p.traces); i++ {
		if i == len(p.traces)-1 {
			continue
		}

		if i%p.DrawEveryNTick != 0 {
			continue
		}

		currentTrace := p.traces[i]
		nextTrace := p.traces[i+1]

		// adjust for offset
		vector.StrokeLine(
			screen,
			float32(float64(currentTrace[0])+float64(p.Offset[0])),
			float32(float64(currentTrace[1])+float64(p.Offset[1])),
			float32(float64(nextTrace[0])+float64(p.Offset[0])),
			float32(float64(nextTrace[1])+float64(p.Offset[1])),
			float32(p.TraceWidth), p.Color, p.AntialiasTraces,
		)
	}
}
