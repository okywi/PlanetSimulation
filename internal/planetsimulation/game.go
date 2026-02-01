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
	game.controls.handlePlanetCreation(game.simulation)
	game.ui.Update()
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	// draw Simulation
	game.simulation.Draw(screen)

	// draw ui
	game.ui.Draw(screen)

	// draw debug
	/*ebitenutil.DebugPrint(screen, "FPS: "+strconv.FormatFloat(ebiten.ActualFPS(), 'f', 2, 64))
	ebitenutil.DebugPrintAt(screen, "TPS: "+strconv.FormatFloat(ebiten.ActualTPS(), 'f', 2, 64), 0, 30)
	ebitenutil.DebugPrintAt(screen, "Planet Radius: "+strconv.FormatFloat(float64(game.simulation.planetRadius), 'f', 2, 64), 0, 60)
	ebitenutil.DebugPrintAt(screen, "offset X: "+strconv.FormatFloat(float64(-(game.simulation.screen.offset[0]-game.screenSize[0]/2)), 'f', 2, 64), 0, 120)
	ebitenutil.DebugPrintAt(screen, "offset Y: "+strconv.FormatFloat(float64(game.simulation.screen.offset[1]-game.screenSize[1]/2), 'f', 2, 64), 0, 150)

	ebitenutil.DebugPrintAt(screen, "Selected Planet:", 0, 200)

	if game.simulation.selectedPlanet != nil {
		ebitenutil.DebugPrintAt(screen, "X: "+strconv.FormatFloat(game.simulation.selectedPlanet.x, 'f', 2, 64), 0, 220)
		ebitenutil.DebugPrintAt(screen, "Y: "+strconv.FormatFloat(game.simulation.selectedPlanet.y, 'f', 2, 64), 0, 230)
		ebitenutil.DebugPrintAt(screen, "Radius: "+strconv.FormatFloat(float64(game.simulation.selectedPlanet.radius), 'f', 2, 64), 0, 240)
		ebitenutil.DebugPrintAt(screen, "Mass: "+strconv.FormatFloat(float64(game.simulation.selectedPlanet.mass), 'f', 2, 64), 0, 250)
		ebitenutil.DebugPrintAt(screen, "Velocity X: "+strconv.FormatFloat(float64(game.simulation.selectedPlanet.velocity.x), 'f', 2, 64), 0, 260)
		ebitenutil.DebugPrintAt(screen, "Velocity Y: "+strconv.FormatFloat(float64(game.simulation.selectedPlanet.velocity.y), 'f', 2, 64), 0, 270)
		ebitenutil.DebugPrintAt(screen, "Color: "+fmt.Sprint(game.simulation.selectedPlanet.color), 0, 280)
	}*/

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

	game.simulation.createFirstPlanet()

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
