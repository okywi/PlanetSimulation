package planetsimulation

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

type SimulationScreen struct {
	image    *ebiten.Image
	geometry ebiten.GeoM
	offset   []int
}

type simulation struct {
	screen                *SimulationScreen
	gameSize              []int
	currentScale          float64
	planets               []*Planet
	planetRadius          float32
	planetMass            float64
	gravitationalConstant float64
	shouldReset           bool
	running               bool
	selectedPlanet        *Planet
}

func newSimulationScreen(gameSize []int) *SimulationScreen {

	screen := &SimulationScreen{
		image:    ebiten.NewImage(gameSize[0], gameSize[1]),
		geometry: ebiten.GeoM{},
		offset:   []int{gameSize[0] / 2, gameSize[1] / 2},
	}

	return screen
}

func newSimulation(gameSize []int) *simulation {
	ebiten.SetTPS(120)

	screen := newSimulationScreen(gameSize)

	return &simulation{
		screen:                screen,
		gameSize:              gameSize,
		planetRadius:          10,
		planetMass:            5000,
		gravitationalConstant: 10.0,
		shouldReset:           false,
		running:               true,
	}
}

func (sim *simulation) createFirstPlanet() {
	// create first planet
	planet := createPlanet(0, 0, 10, 5000000, vector2{0, 0}, SetColor(255, 0, 0, 255), sim.screen.offset)
	sim.planets = append(sim.planets, planet)
}

func (sim *simulation) Update() error {
	var err error
	for _, planet := range sim.planets {
		err = planet.Update(sim, sim.planets)
	}

	if sim.shouldReset {
		sim.planets = slices.Delete(sim.planets, 1, len(sim.planets))
		sim.shouldReset = false
	}

	return err
}

func (sim *simulation) Draw(gameScreen *ebiten.Image) {
	sim.screen.image.Clear()

	// draw planets
	for _, planet := range sim.planets {
		planet.Draw(sim.screen.image)
	}

	gameScreen.DrawImage(sim.screen.image, &ebiten.DrawImageOptions{
		GeoM: sim.screen.geometry,
	})
}
