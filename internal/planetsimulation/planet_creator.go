package planetsimulation

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type planetCreator struct {
	planet     *Planet
	showPlanet bool
}

func newPlanetCreator() *planetCreator {
	return &planetCreator{
		planet: newPlanet(
			"Planet 1",
			0,
			0,
			10,
			5,
			vector2{0, 0},
			SetColor(255, 0, 0, 255),
			[]float64{0, 0},
		),
		showPlanet: false,
	}
}

func (planetCreator *planetCreator) Update(x float64, y float64, planetHandler *planetHandler) {
	if !planetCreator.planet.HasNameChanged {
		planetCreator.planet.Name = fmt.Sprintf("Planet %d", planetHandler.planetCounter+1)
	}
	// set x
	planetCreator.planet.X = x
	planetCreator.planet.Y = y

	planet := planetCreator.planet
	planet.geometry.Reset()
	// center planet
	planet.geometry.Translate(planet.X-float64(planet.Radius), planet.Y-float64(planet.Radius))
	// adjust for offset
	planet.geometry.Translate(planetHandler.planetsOffset[0], planetHandler.planetsOffset[1])

	// update image
	radius := float32(planet.Radius)
	transparentColor := color.NRGBA{planet.Color.R, planet.Color.G, planet.Color.B, 100}
	planetCreator.planet.image = ebiten.NewImage(int(planet.Radius*2), int(planet.Radius*2))
	vector.FillCircle(planetCreator.planet.image, radius, radius, radius, transparentColor, true)

}

func (planetCreator *planetCreator) spawnPlanet(planetHandler *planetHandler) {
	// check if would collide on spawn
	for _, planet := range planetHandler.planets {
		toCreatePlanet := planetHandler.planetCreator.planet
		if _, _, _, overlaps := overlapsCircle(planet.X, toCreatePlanet.X, planet.Y, toCreatePlanet.Y, planet.Radius, toCreatePlanet.Radius); overlaps {
			return
		}
	}

	newPlanet := newPlanet(
		planetCreator.planet.Name,
		planetCreator.planet.X,
		planetCreator.planet.Y,
		planetCreator.planet.Radius,
		planetCreator.planet.Mass,
		planetCreator.planet.Velocity,
		planetCreator.planet.Color,
		planetHandler.planetsOffset,
	)

	planetHandler.planets = append(planetHandler.planets, newPlanet)
	planetHandler.planetCounter++

	// make planetCreator planet highlight invisible
	planetCreator.showPlanet = false

	// reset name change of planetCreator
	planetCreator.planet.HasNameChanged = false

	// select planet if none other planet is selected
	if !planetHandler.selectedPlanet.isSelected {
		// should be last element appended
		planetHandler.selectPlanet(len(planetHandler.planets) - 1)
		return
	}
}
