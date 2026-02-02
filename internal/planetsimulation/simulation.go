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
	offset   []float64
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

type simulation struct {
	screen                *SimulationScreen
	gameSize              []int
	currentScale          float64
	planets               []*Planet
	toCreatePlanet        toCreatePlanet
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
		offset:   []float64{float64(gameSize[0]) / 2, float64(gameSize[1] / 2)},
	}

	return screen
}

func newSimulation(gameSize []int) *simulation {
	screen := newSimulationScreen(gameSize)

	// planet that is created by a click
	toCreatePlanet := toCreatePlanet{
		radius:   10,
		mass:     5,
		velocity: vector2{0, 0},
		color:    SetColor(255, 0, 0, 255),
		geometry: ebiten.GeoM{},
	}

	return &simulation{
		screen:                screen,
		gameSize:              gameSize,
		toCreatePlanet:        toCreatePlanet,
		gravitationalConstant: 10000.0,
		shouldReset:           false,
		running:               true,
		tps:                   120,
		planetsToRemove:       make([]int, 0),
	}
}

func (sim *simulation) returnToOrigin() {
	dx := sim.screen.offset[0] - float64(sim.gameSize[0]/2)
	dy := sim.screen.offset[1] - float64(sim.gameSize[1]/2)
	sim.screen.offset[0] -= dx
	sim.screen.offset[1] -= dy

	for _, planet := range sim.planets {
		planet.geometry.Translate(float64(-dx), float64(-dy))
	}
}

func (sim *simulation) spawnPlanet() {
	newPlanet := newPlanet(
		sim.toCreatePlanet.x,
		sim.toCreatePlanet.y,
		sim.toCreatePlanet.radius,
		sim.toCreatePlanet.mass,
		sim.toCreatePlanet.velocity,
		sim.toCreatePlanet.color,
		sim.screen.offset,
	)

	sim.planets = append(sim.planets, newPlanet)

	// show tocreateplanet
	sim.toCreatePlanet.shown = false

	// select planet if none other planet is selected
	if !sim.isPlanetSelected {
		// should be last element appended
		sim.selectedPlanetIndex = len(sim.planets) - 1
		sim.isPlanetSelected = true
		return
	}
}

func (sim *simulation) Update() error {
	ebiten.SetTPS(sim.tps)

	var err error
	for _, planet := range sim.planets {
		err = planet.Update(sim, sim.planets)
		if sim.isPlanetFocused {
			planet.focus(sim)
		}
	}

	if sim.shouldReset {
		sim.isPlanetSelected = false
		sim.isPlanetFocused = false
		sim.toCreatePlanet.shown = false
		sim.planets = slices.Delete(sim.planets, 0, len(sim.planets))
		dx := sim.screen.offset[0] - float64(sim.gameSize[0]/2)
		dy := sim.screen.offset[1] - float64(sim.gameSize[1]/2)
		sim.screen.offset[0] -= dx
		sim.screen.offset[1] -= dy

		for _, planet := range sim.planets {
			planet.geometry.Translate(float64(-dx), float64(-dy))
		}

		sim.shouldReset = false
	}

	// delete planets
	if len(sim.planetsToRemove) > 0 {
		log.Println(sim.planetsToRemove)
		for _, planetIndex := range sim.planetsToRemove {
			sim.planets = slices.Delete(sim.planets, planetIndex, planetIndex+1)
			sim.planetsToRemove = slices.DeleteFunc(sim.planetsToRemove, func(pIndex int) bool {
				if planetIndex == pIndex {
					return true
				}
				return false
			})
		}
		log.Println(sim.planetsToRemove)
	}

	return err
}

func (sim *simulation) removeSelectedPlanet(ui *ui) {
	selectedPlanetIndex := sim.selectedPlanetIndex
	for i := 0; i < len(sim.planets); i++ {
		if i == selectedPlanetIndex {
			sim.planetsToRemove = append(sim.planetsToRemove, selectedPlanetIndex)
			sim.isPlanetSelected = false
		}
	}
}

func (sim *simulation) updateToCreatePlanet(x float64, y float64) {
	// set x
	sim.toCreatePlanet.x = x
	sim.toCreatePlanet.y = y

	planet := sim.toCreatePlanet
	radius := float32(planet.radius)

	r, g, b, _ := convertColorToInt(planet.color)
	transparentColor := color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 100}
	//transparentColor := SetColor(uint8(r), uint8(g), uint8(b), 100)
	sim.toCreatePlanet.image = ebiten.NewImage(int(planet.radius*2), int(planet.radius*2))
	//strokeWidth := float32(1)
	vector.FillCircle(sim.toCreatePlanet.image, radius, radius, radius, transparentColor, true)

	planet.geometry.Reset()
	// center planet
	planet.geometry.Translate(planet.x-float64(planet.radius), planet.y-float64(planet.radius))
	// adjust for offset
	planet.geometry.Translate(float64(sim.screen.offset[0]), float64(sim.screen.offset[1]))

	sim.toCreatePlanet.geometry = planet.geometry
}

func (sim *simulation) drawToCreatePlanet(screen *ebiten.Image) {
	screen.DrawImage(sim.toCreatePlanet.image, &ebiten.DrawImageOptions{
		GeoM: sim.toCreatePlanet.geometry,
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

	// draw toCreatePlanet if needed
	if sim.toCreatePlanet.shown {
		sim.drawToCreatePlanet(sim.screen.image)
	}

	gameScreen.DrawImage(sim.screen.image, &ebiten.DrawImageOptions{
		GeoM: sim.screen.geometry,
	})
}
