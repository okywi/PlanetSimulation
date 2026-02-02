package planetsimulation

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	screenSize []int
	simulation *simulation
	controls   *controls
	ui         *ui
}

func (game *Game) Update() error {
	game.simulation.Update()
	game.controls.handlePlanetCreation(game.simulation, game.ui)
	if err := game.ui.Update(game); err != nil {
		return err
	}
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	// draw Simulation
	game.simulation.Draw(screen)

	// draw ui
	game.ui.Draw(screen)
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return game.screenSize[0], game.screenSize[1]
}

func createGame() *Game {
	game := Game{
		screenSize: []int{1920, 1080},
		controls:   newControls(),
		ui:         newUI(),
	}

	game.simulation = newSimulation(game.screenSize)

	//game.simulation.createFirstPlanet()

	// Window Setup
	ebiten.SetWindowSize(game.screenSize[0], game.screenSize[1])
	ebiten.SetWindowTitle("Planet Simulation")

	return &game
}

func Start() {
	if err := ebiten.RunGame(createGame()); err != nil {
		log.Fatal(err)
	}
}
