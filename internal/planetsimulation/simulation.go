package planetsimulation

import (
	"image/color"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

type SimulationScreen struct {
	image    *ebiten.Image
	geometry ebiten.GeoM
}

type simulation struct {
	screen            *SimulationScreen
	gameSize          []int
	simulationPresets *simulationPresets
	planetHandler     *planetHandler
	shouldReset       bool
	tps               int
}

func newSimulationScreen(gameSize []int) *SimulationScreen {
	screen := &SimulationScreen{
		image:    ebiten.NewImage(gameSize[0], gameSize[1]),
		geometry: ebiten.GeoM{},
	}

	return screen
}

func newSimulation(gameSize []int) *simulation {
	screen := newSimulationScreen(gameSize)

	sim := &simulation{
		screen:            screen,
		gameSize:          gameSize,
		simulationPresets: newSimulationPresets(),
		planetHandler:     newPlanetHandler(gameSize),
		shouldReset:       false,
		tps:               120,
	}

	return sim
}

func (sim *simulation) getCoords(planetHandler *planetHandler) []float64 {
	return []float64{
		-(planetHandler.planetsOffset[0] - planetHandler.defaultPlanetsOffset[0]),
		planetHandler.planetsOffset[1] - planetHandler.defaultPlanetsOffset[1],
	}
}

func (sim *simulation) handleReset() {
	if sim.shouldReset || sim.simulationPresets.shouldLoadSimulation {
		sim.planetHandler.selectedPlanet.isSelected = false
		sim.planetHandler.focusedPlanet.isFocused = false
		sim.planetHandler.planetCreator.showPlanet = false
		sim.planetHandler.planetCounter = 0
		sim.planetHandler.planets = slices.Delete(sim.planetHandler.planets, 0, len(sim.planetHandler.planets))
		dx := sim.planetHandler.planetsOffset[0] - sim.planetHandler.defaultPlanetsOffset[0]
		dy := sim.planetHandler.planetsOffset[1] - sim.planetHandler.defaultPlanetsOffset[1]
		sim.planetHandler.planetsOffset[0] -= dx
		sim.planetHandler.planetsOffset[1] -= dy

		for _, planet := range sim.planetHandler.planets {
			planet.geometry.Translate(-dx, -dy)
		}

		sim.shouldReset = false
	}
}

func (sim *simulation) Update() {
	ebiten.SetTPS(sim.tps)

	sim.handleReset()
	sim.planetHandler.Update()
	sim.simulationPresets.handleLoadSimulationPreset(sim.planetHandler, sim.simulationPresets.presetIndex)
}

func (sim *simulation) Draw(gameScreen *ebiten.Image) {
	sim.screen.image.Fill(color.Black)

	sim.planetHandler.Draw(sim.screen.image)

	gameScreen.DrawImage(sim.screen.image, &ebiten.DrawImageOptions{
		GeoM: sim.screen.geometry,
	})
}
