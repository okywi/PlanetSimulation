package planetsimulation

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
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
	planetPresets         []*Planet
	planetPresetPath      string
	planetsOffset         []float64
	defaultOffset         []float64
	planetCounter         int
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

	sim := &simulation{
		screen:                screen,
		gameSize:              gameSize,
		planetCreator:         planetCreator,
		defaultOffset:         []float64{float64(gameSize[0]) / 2, float64(gameSize[1] / 2)},
		planetPresetPath:      "assets/data/planet_presets.json",
		gravitationalConstant: 10000.0,
		shouldReset:           false,
		running:               true,
		tps:                   120,
		planetsToRemove:       make([]int, 0),
	}
	sim.planetsOffset = []float64{sim.defaultOffset[0], sim.defaultOffset[1]}

	sim.loadPresetsFromFile()

	return sim
}

func (sim *simulation) addPlanetToPlanetPresets(planetToAdd Planet) {
	// replace if same name
	for i, planet := range sim.planetPresets {
		if planet.Name == planetToAdd.Name {
			sim.planetPresets[i] = &planetToAdd
			return
		}
	}

	sim.planetPresets = append(sim.planetPresets, &planetToAdd)

	sim.savePresetsToFile()
}

func (sim *simulation) savePresetsToFile() {
	if err := os.MkdirAll(filepath.Dir(sim.planetPresetPath), os.ModePerm); err != nil {
		log.Printf("Failed to create dir for %s file: %v", sim.planetPresetPath, err)
	}

	content, err := json.MarshalIndent(sim.planetPresets, "", " ")

	if err != nil {
		log.Printf("Failed to marshal planet presets json: %v", err)
	}

	if err := os.WriteFile(sim.planetPresetPath, content, os.ModePerm); err != nil {
		log.Printf("Failed to create file %s: %v", sim.planetPresetPath, err)
	}
}

func (sim *simulation) loadPresetsFromFile() {
	if _, err := os.Stat(sim.planetPresetPath); err != nil {
		return
	}

	content, err := os.ReadFile(sim.planetPresetPath)
	if err != nil {
		log.Printf("Failed to read file %s: %v", sim.planetPresetPath, err)
	}

	if err := json.Unmarshal(content, &sim.planetPresets); err != nil {
		log.Printf("Failed to unmarshal planet presets json: %v", err)
	}
}

func (sim *simulation) returnToOrigin() {
	// reset planetsOffset
	dx := sim.planetsOffset[0] - sim.defaultOffset[0]
	dy := sim.planetsOffset[1] - sim.defaultOffset[1]
	sim.planetsOffset[0] -= dx
	sim.planetsOffset[1] -= dy

	// move planet images as well
	for _, planet := range sim.planets {
		planet.Geometry.Translate(float64(-dx), float64(-dy))
	}
}

func (sim *simulation) spawnPlanet() {
	// check if would collide on spawn
	for _, planet := range sim.planets {
		toCreatePlanet := sim.planetCreator.planet
		if _, _, _, overlaps := overlapsCircle(planet.X, toCreatePlanet.X, planet.Y, toCreatePlanet.Y, planet.Radius, toCreatePlanet.Radius); overlaps {
			return
		}
	}

	newPlanet := newPlanet(
		sim.planetCreator.planet.Name,
		sim.planetCreator.planet.X,
		sim.planetCreator.planet.Y,
		sim.planetCreator.planet.Radius,
		sim.planetCreator.planet.Mass,
		sim.planetCreator.planet.Velocity,
		sim.planetCreator.planet.Color,
		sim.planetsOffset,
	)

	sim.planets = append(sim.planets, newPlanet)
	sim.planetCounter++

	// make planetCreator planet highlight invisible
	sim.planetCreator.showPlanet = false

	// reset name change of planetCreator
	sim.planetCreator.planet.HasNameChanged = false

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
			planet.Geometry.Translate(-dx, -dy)
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
	if !sim.planetCreator.planet.HasNameChanged {
		sim.planetCreator.planet.Name = fmt.Sprintf("Planet %d", sim.planetCounter+1)
	}
	// set x
	sim.planetCreator.planet.X = x
	sim.planetCreator.planet.Y = y

	planet := sim.planetCreator.planet
	radius := float32(planet.Radius)

	r, g, b, _ := convertColorToInt(planet.Color)

	transparentColor := SetColor(uint8(r), uint8(g), uint8(b), 100)
	sim.planetCreator.planet.Image = ebiten.NewImage(int(planet.Radius*2), int(planet.Radius*2))
	vector.FillCircle(sim.planetCreator.planet.Image, radius, radius, radius, transparentColor, true)

	planet.Geometry.Reset()
	// center planet
	planet.Geometry.Translate(planet.X-float64(planet.Radius), planet.Y-float64(planet.Radius))
	// adjust for offset
	planet.Geometry.Translate(sim.planetsOffset[0], sim.planetsOffset[1])

	sim.planetCreator.planet.Geometry = planet.Geometry
}

func (sim *simulation) drawToCreatePlanet(screen *ebiten.Image) {
	screen.DrawImage(sim.planetCreator.planet.Image, &ebiten.DrawImageOptions{
		GeoM: sim.planetCreator.planet.Geometry,
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
