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

	// planet that is created by a click
	planetCreator := &planetCreator{
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

	planetHandler := &planetHandler{
		planetCreator:         planetCreator,
		defaultPlanetsOffset:  []float64{float64(gameSize[0]) / 2, float64(gameSize[1] / 2)},
		presetFilePath:        "assets/data/planet_presets.json",
		planetsToRemove:       make([]int, 0),
		planetCounter:         0,
		gravitationalConstant: 10000.0,
		running:               true,
	}
	planetHandler.planetsOffset = []float64{planetHandler.defaultPlanetsOffset[0], planetHandler.defaultPlanetsOffset[1]}
	planetHandler.loadPlanetPresetsFromFile()

	simulationPresets := &simulationPresets{
		filePath: "assets/data/simulation_presets.json",
	}

	sim := &simulation{
		screen:            screen,
		gameSize:          gameSize,
		simulationPresets: simulationPresets,
		planetHandler:     planetHandler,
		shouldReset:       false,
		tps:               120,
	}

	simulationPresets.loadSimulationPresetsFromFile()

	return sim
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
