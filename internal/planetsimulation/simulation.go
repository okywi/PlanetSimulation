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

type focusedPlanet struct {
	isFocused bool
	index     int
}

type selectedPlanet struct {
	index      int
	isSelected bool
}

type simulation struct {
	screen                *SimulationScreen
	gameSize              []int
	simulationPresets     *simulationPresets
	planetHandler         *planetHandler
	gravitationalConstant float64
	shouldReset           bool
	running               bool
	tps                   int
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
		planetCreator:        planetCreator,
		defaultPlanetsOffset: []float64{float64(gameSize[0]) / 2, float64(gameSize[1] / 2)},
		presetFilePath:       "assets/data/planet_presets.json",
		planetsToRemove:      make([]int, 0),
		planetCounter:        0,
	}
	planetHandler.planetsOffset = []float64{planetHandler.defaultPlanetsOffset[0], planetHandler.defaultPlanetsOffset[1]}

	simulationPresets := &simulationPresets{
		filePath: "assets/data/simulation_presets.json",
	}

	sim := &simulation{
		screen:                screen,
		gameSize:              gameSize,
		simulationPresets:     simulationPresets,
		planetHandler:         planetHandler,
		gravitationalConstant: 10000.0,
		shouldReset:           false,
		running:               true,
		tps:                   120,
	}

	sim.loadPlanetPresetsFromFile()
	sim.loadSimulationPresetsFromFile()

	return sim
}

func (sim *simulation) returnToOrigin() {
	// reset planetsOffset
	dx := sim.planetHandler.planetsOffset[0] - sim.planetHandler.defaultPlanetsOffset[0]
	dy := sim.planetHandler.planetsOffset[1] - sim.planetHandler.defaultPlanetsOffset[1]
	sim.planetHandler.planetsOffset[0] -= dx
	sim.planetHandler.planetsOffset[1] -= dy

	// move planet images as well
	for _, planet := range sim.planetHandler.planets {
		planet.geometry.Translate(float64(-dx), float64(-dy))
	}
}

func (sim *simulation) handleReset() {
	if sim.shouldReset {
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
	sim.handlePlanetDeletion()
	sim.updatePlanets()
	sim.handleLoadSimulationPreset(sim.simulationPresets.presetIndex)
}

func (sim *simulation) Draw(gameScreen *ebiten.Image) {
	sim.screen.image.Fill(color.Black)

	// draw planets
	for _, planet := range sim.planetHandler.planets {
		if planet != nil {
			planet.Draw(sim.screen.image)
		}
	}

	// draw toCreatePlanet
	if sim.planetHandler.planetCreator.showPlanet {
		sim.drawToCreatePlanet(sim.screen.image)
	}

	gameScreen.DrawImage(sim.screen.image, &ebiten.DrawImageOptions{
		GeoM: sim.screen.geometry,
	})
}
