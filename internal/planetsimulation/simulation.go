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

type simulationPreset struct {
	Name    string
	Planets []*Planet
}

type simulation struct {
	screen                       *SimulationScreen
	gameSize                     []int
	currentScale                 float64
	planets                      []*Planet
	planetPresets                []*Planet
	planetPresetPath             string
	planetsOffset                []float64
	simulationPresets            []*simulationPreset
	currentSimulationPresetName  string
	currentSimulationPresetIndex int
	shouldLoadSimulation         bool
	simulationPresetPath         string
	defaultOffset                []float64
	planetCounter                int
	planetCreator                *planetCreator
	gravitationalConstant        float64
	shouldReset                  bool
	running                      bool
	selectedPlanetIndex          int
	isPlanetSelected             bool
	focusedPlanetIndex           int
	isPlanetFocused              bool
	planetsToRemove              []int
	tps                          int
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
		simulationPresetPath:  "assets/data/simulation_presets.json",
		simulationPresets:     []*simulationPreset{},
		gravitationalConstant: 10000.0,
		shouldReset:           false,
		running:               true,
		tps:                   120,
		planetsToRemove:       make([]int, 0),
	}
	sim.planetsOffset = []float64{sim.defaultOffset[0], sim.defaultOffset[1]}

	sim.loadPlanetPresetsFromFile()
	sim.loadSimulationPresetsFromFile()

	return sim
}

func (sim *simulation) saveSimulationPreset() {
	planets := []*Planet{}
	for _, planet := range sim.planets {
		planets = append(planets, *&planet)
	}

	sim.simulationPresets = append(sim.simulationPresets, &simulationPreset{
		Name:    sim.currentSimulationPresetName,
		Planets: planets,
	})
	sim.saveSimulationPresets()
}

func (sim *simulation) removeSimulationPreset(i int) {
	sim.simulationPresets = slices.Delete(sim.simulationPresets, i, i+1)

	sim.saveSimulationPresets()
}

func (sim *simulation) loadSimulationPreset(i int) {
	for _, planet := range sim.simulationPresets[i].Planets {

		sim.planets = append(sim.planets, newPlanet(
			planet.Name,
			planet.X,
			planet.Y,
			planet.Radius,
			planet.Mass,
			planet.Velocity,
			planet.Color,
			sim.planetsOffset,
		))
	}

	sim.running = false
	sim.shouldLoadSimulation = false
}

func (sim *simulation) loadSimulationPresetsFromFile() {
	if _, err := os.Stat(sim.simulationPresetPath); err != nil {
		return
	}

	content, err := os.ReadFile(sim.simulationPresetPath)
	if err != nil {
		log.Printf("Failed to read file %s: %v", sim.simulationPresetPath, err)
	}

	if err := json.Unmarshal(content, &sim.simulationPresets); err != nil {
		log.Printf("Failed to unmarshal planet presets json: %v", err)
	}
}

func (sim *simulation) saveSimulationPresets() {
	if err := os.MkdirAll(filepath.Dir(sim.simulationPresetPath), os.ModePerm); err != nil {
		log.Printf("Failed to create dir for %s file: %v", sim.simulationPresetPath, err)
	}

	content, err := json.MarshalIndent(sim.simulationPresets, "", " ")

	if err != nil {
		log.Printf("Failed to marshal planet presets json: %v", err)
	}

	if err := os.WriteFile(sim.simulationPresetPath, content, os.ModePerm); err != nil {
		log.Printf("Failed to create file %s: %v", sim.simulationPresetPath, err)
	}
}

func (sim *simulation) addPlanetToPlanetPresets(planetToAdd Planet) {
	// replace if same name
	for i, planet := range sim.planetPresets {
		if planet.Name == planetToAdd.Name {
			sim.planetPresets[i] = &planetToAdd
			sim.savePlanetPresetsToFile()
			return
		}
	}

	sim.planetPresets = append(sim.planetPresets, &planetToAdd)

	sim.savePlanetPresetsToFile()
}

func (sim *simulation) savePlanetPresetsToFile() {
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

func (sim *simulation) loadPlanetPresetsFromFile() {
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
		planet.geometry.Translate(float64(-dx), float64(-dy))
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
		sim.planetCounter = 0
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
	if sim.shouldLoadSimulation {
		sim.loadSimulationPreset(sim.currentSimulationPresetIndex)
	}
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
	sim.planetCreator.planet.image = ebiten.NewImage(int(planet.Radius*2), int(planet.Radius*2))
	vector.FillCircle(sim.planetCreator.planet.image, radius, radius, radius, transparentColor, true)

	planet.geometry.Reset()
	// center planet
	planet.geometry.Translate(planet.X-float64(planet.Radius), planet.Y-float64(planet.Radius))
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
