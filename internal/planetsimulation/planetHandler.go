package planetsimulation

import (
	"fmt"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type planetCreator struct {
	planet     *Planet
	showPlanet bool
}

type planetHandler struct {
	planets              []*Planet
	presets              []*Planet
	presetFilePath       string
	planetsOffset        []float64
	planetCounter        int
	planetCreator        *planetCreator
	planetsToRemove      []int
	defaultPlanetsOffset []float64
	selectedPlanet       selectedPlanet
	focusedPlanet        focusedPlanet
}

func (sim *simulation) spawnPlanet() {
	// check if would collide on spawn
	for _, planet := range sim.planetHandler.planets {
		toCreatePlanet := sim.planetHandler.planetCreator.planet
		if _, _, _, overlaps := overlapsCircle(planet.X, toCreatePlanet.X, planet.Y, toCreatePlanet.Y, planet.Radius, toCreatePlanet.Radius); overlaps {
			return
		}
	}

	planetInCreator := sim.planetHandler.planetCreator.planet
	newPlanet := newPlanet(
		planetInCreator.Name,
		planetInCreator.X,
		planetInCreator.Y,
		planetInCreator.Radius,
		planetInCreator.Mass,
		planetInCreator.Velocity,
		planetInCreator.Color,
		sim.planetHandler.planetsOffset,
	)

	sim.planetHandler.planets = append(sim.planetHandler.planets, newPlanet)
	sim.planetHandler.planetCounter++

	// make planetCreator planet highlight invisible
	sim.planetHandler.planetCreator.showPlanet = false

	// reset name change of planetCreator
	sim.planetHandler.planetCreator.planet.HasNameChanged = false

	// select planet if none other planet is selected
	if !sim.planetHandler.selectedPlanet.isSelected {
		// should be last element appended
		sim.planetHandler.selectedPlanet.index = len(sim.planetHandler.planets) - 1
		sim.planetHandler.selectedPlanet.isSelected = true
		return
	}
}

func (sim *simulation) handlePlanetDeletion() {
	if len(sim.planetHandler.planetsToRemove) > 0 {
		for _, planetIndex := range sim.planetHandler.planetsToRemove {
			// remove from planets
			if sim.planetHandler.selectedPlanet.index == planetIndex {
				sim.planetHandler.selectedPlanet.isSelected = false
			}
			if sim.planetHandler.focusedPlanet.index == planetIndex {
				sim.planetHandler.focusedPlanet.isFocused = false
			}

			sim.planetHandler.planets = slices.Delete(sim.planetHandler.planets, planetIndex, planetIndex+1)

		}

		sim.planetHandler.planetsToRemove = []int{}
	}
}

func (sim *simulation) updatePlanets() {
	for _, planet := range sim.planetHandler.planets {
		planet.Update(sim, sim.planetHandler.planets)
		if sim.planetHandler.focusedPlanet.isFocused {
			planet.focus(sim)
		}
	}
}

func (sim *simulation) removeSelectedPlanet(ui *ui) {
	sim.planetHandler.planetsToRemove = append(sim.planetHandler.planetsToRemove, sim.planetHandler.selectedPlanet.index)
	sim.planetHandler.selectedPlanet.isSelected = false
}

func (sim *simulation) updateToCreatePlanet(x float64, y float64) {
	if !sim.planetHandler.planetCreator.planet.HasNameChanged {
		sim.planetHandler.planetCreator.planet.Name = fmt.Sprintf("Planet %d", sim.planetHandler.planetCounter+1)
	}
	// set x
	sim.planetHandler.planetCreator.planet.X = x
	sim.planetHandler.planetCreator.planet.Y = y

	planet := sim.planetHandler.planetCreator.planet
	radius := float32(planet.Radius)

	r, g, b, _ := convertColorToInt(planet.Color)

	transparentColor := SetColor(uint8(r), uint8(g), uint8(b), 100)
	sim.planetHandler.planetCreator.planet.image = ebiten.NewImage(int(planet.Radius*2), int(planet.Radius*2))
	vector.FillCircle(sim.planetHandler.planetCreator.planet.image, radius, radius, radius, transparentColor, true)

	planet.geometry.Reset()
	// center planet
	planet.geometry.Translate(planet.X-float64(planet.Radius), planet.Y-float64(planet.Radius))
	// adjust for offset
	planet.geometry.Translate(sim.planetHandler.planetsOffset[0], sim.planetHandler.planetsOffset[1])

	sim.planetHandler.planetCreator.planet.geometry = planet.geometry
}

func (sim *simulation) drawToCreatePlanet(screen *ebiten.Image) {
	screen.DrawImage(sim.planetHandler.planetCreator.planet.image, &ebiten.DrawImageOptions{
		GeoM: sim.planetHandler.planetCreator.planet.geometry,
	})
}
