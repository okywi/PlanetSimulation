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
	game.controls.Update(game.simulation, game.ui)
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
		screenSize: make([]int, 2),
		controls:   newControls(),
		ui:         newUI(),
	}

	// Window Setup
	ebiten.SetWindowTitle("Planet Simulation")
	ebiten.SetFullscreen(true)
	game.screenSize[0], game.screenSize[1] = ebiten.Monitor().Size()

	// new simulation
	game.simulation = newSimulation(game.screenSize)

	return &game
}

func Start() {
	if err := ebiten.RunGame(createGame()); err != nil {
		log.Fatal(err)
	}
}
