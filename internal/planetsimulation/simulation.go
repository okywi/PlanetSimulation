package planetsimulation

import (
	"image/color"
	"log"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type SimulationScreen struct {
	image    *ebiten.Image
	geometry ebiten.GeoM
}

type toCreatePlanet struct {
	x        float64
	y        float64
	radius   float64
	mass     float64
	velocity vector2
	color    color.Color
	shown    bool
	image    *ebiten.Image
	geometry ebiten.GeoM
}

type planetCreator struct {
	planet     *Planet
	showPlanet bool
}

type simulation struct {
	screen                *SimulationScreen
	gameSize              []int
	currentScale          float64
	planets               []*Planet
	planetsOffset         []float64
	defaultOffset         []float64
	planetCreator         *planetCreator
	gravitationalConstant float64
	shouldReset           bool
	running               bool
	selectedPlanetIndex   int
	isPlanetSelected      bool
	focusedPlanetIndex    int
	isPlanetFocused       bool
	planetsToRemove       []int
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

	sim := &simulation{
		screen:                screen,
		gameSize:              gameSize,
		planetCreator:         planetCreator,
		defaultOffset:         []float64{float64(gameSize[0]) / 2, float64(gameSize[1] / 2)},
		gravitationalConstant: 10000.0,
		shouldReset:           false,
		running:               true,
		tps:                   120,
		planetsToRemove:       make([]int, 0),
	}
	sim.planetsOffset = []float64{sim.defaultOffset[0], sim.defaultOffset[1]}

	return sim
}

func (sim *simulation) returnToOrigin() {
	// reset planetsOffset
	dx := sim.planetsOffset[0] - sim.defaultOffset[0]
	dy := sim.planetsOffset[1] - sim.defaultOffset[1]
	sim.planetsOffset[0] -= dx
	sim.planetsOffset[1] -= dy

	// move planet images as well
	for _, planet := range sim.planets {
		planet.geometry.Translate(float64(-dx), float64(-dy))
	}
}

func (sim *simulation) spawnPlanet() {
	// check if would collide on spawn
	for _, planet := range sim.planets {
		toCreatePlanet := sim.planetCreator.planet
		if _, _, _, overlaps := overlapsCircle(planet.x, toCreatePlanet.x, planet.y, toCreatePlanet.y, planet.radius, toCreatePlanet.radius); overlaps {
			return
		}
	}

	newPlanet := newPlanet(
		sim.planetCreator.planet.x,
		sim.planetCreator.planet.y,
		sim.planetCreator.planet.radius,
		sim.planetCreator.planet.mass,
		sim.planetCreator.planet.velocity,
		sim.planetCreator.planet.color,
		sim.planetsOffset,
	)

	sim.planets = append(sim.planets, newPlanet)

	// select planet if none other planet is selected
	if !sim.isPlanetSelected {
		// should be last element appended
		sim.selectedPlanetIndex = len(sim.planets) - 1
		sim.isPlanetSelected = true
		return
	}
}

func (sim *simulation) handleReset() {
	if sim.shouldReset {
		sim.isPlanetSelected = false
		sim.isPlanetFocused = false
		sim.planetCreator.showPlanet = false
		sim.planets = slices.Delete(sim.planets, 0, len(sim.planets))
		dx := sim.planetsOffset[0] - sim.defaultOffset[0]
		dy := sim.planetsOffset[1] - sim.defaultOffset[1]
		sim.planetsOffset[0] -= dx
		sim.planetsOffset[1] -= dy

		for _, planet := range sim.planets {
			planet.geometry.Translate(-dx, -dy)
		}

		sim.shouldReset = false
	}
}

func (sim *simulation) handlePlanetDeletion() {
	if len(sim.planetsToRemove) > 0 {
		for _, planetIndex := range sim.planetsToRemove {
			// remove from planets
			if sim.selectedPlanetIndex == planetIndex {
				log.Println("deseleted", sim.selectedPlanetIndex)
				sim.isPlanetSelected = false
			}
			if sim.focusedPlanetIndex == planetIndex {
				sim.isPlanetFocused = false
			}

			sim.planets = slices.Delete(sim.planets, planetIndex, planetIndex+1)

		}

		sim.planetsToRemove = []int{}
	}
}

func (sim *simulation) updatePlanets() {
	for _, planet := range sim.planets {
		planet.Update(sim, sim.planets)
		if sim.isPlanetFocused {
			planet.focus(sim)
		}
	}
}

func (sim *simulation) Update() {
	ebiten.SetTPS(sim.tps)

	sim.handleReset()
	sim.handlePlanetDeletion()
	sim.updatePlanets()

}

func (sim *simulation) removeSelectedPlanet(ui *ui) {
	sim.planetsToRemove = append(sim.planetsToRemove, sim.selectedPlanetIndex)
	sim.isPlanetSelected = false
}

func (sim *simulation) updateToCreatePlanet(x float64, y float64) {
	// set x
	sim.planetCreator.planet.x = x
	sim.planetCreator.planet.y = y

	planet := sim.planetCreator.planet
	radius := float32(planet.radius)

	r, g, b, _ := convertColorToInt(planet.color)

	transparentColor := SetColor(uint8(r), uint8(g), uint8(b), 100)
	sim.planetCreator.planet.image = ebiten.NewImage(int(planet.radius*2), int(planet.radius*2))
	vector.FillCircle(sim.planetCreator.planet.image, radius, radius, radius, transparentColor, true)

	planet.geometry.Reset()
	// center planet
	planet.geometry.Translate(planet.x-float64(planet.radius), planet.y-float64(planet.radius))
	// adjust for offset
	planet.geometry.Translate(sim.planetsOffset[0], sim.planetsOffset[1])

	sim.planetCreator.planet.geometry = planet.geometry
}

func (sim *simulation) drawToCreatePlanet(screen *ebiten.Image) {
	screen.DrawImage(sim.planetCreator.planet.image, &ebiten.DrawImageOptions{
		GeoM: sim.planetCreator.planet.geometry,
	})
}

func (sim *simulation) Draw(gameScreen *ebiten.Image) {
	sim.screen.image.Fill(color.Black)

	// draw planets
	for _, planet := range sim.planets {
		if planet != nil {
			planet.Draw(sim.screen.image)
		}
	}

	// draw toCreatePlanet
	if sim.planetCreator.showPlanet {
		sim.drawToCreatePlanet(sim.screen.image)
	}

	gameScreen.DrawImage(sim.screen.image, &ebiten.DrawImageOptions{
		GeoM: sim.screen.geometry,
	})
}
