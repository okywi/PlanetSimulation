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

type planetCreator struct {
	planet     *Planet
	showPlanet bool
}

type planetHandler struct {
	planets              []*Planet
	presets              []*Planet
	presetFilePath       string
	planetsOffset        []float64
	planetCounter        int
	planetCreator        *planetCreator
	planetsToRemove      []int
	defaultPlanetsOffset []float64
	selectedPlanet       selectedPlanet
	focusedPlanet        focusedPlanet
}

type focusedPlanet struct {
	isFocused bool
	index     int
}

type selectedPlanet struct {
	index      int
	isSelected bool
}

type simulationPresets struct {
	Presets              []*simulationPreset
	newPresetName        string
	presetIndex          int
	shouldLoadSimulation bool
	filePath             string
}

type simulationPreset struct {
	Name    string
	Planets []*Planet
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

func (sim *simulation) saveSimulationPreset() {
	planets := []*Planet{}
	for _, planet := range sim.planetHandler.planets {
		planets = append(planets, *&planet)
	}

	sim.simulationPresets.Presets = append(sim.simulationPresets.Presets, &simulationPreset{
		Name:    sim.simulationPresets.newPresetName,
		Planets: planets,
	})
	sim.saveSimulationPresets()
}

func (sim *simulation) removeSimulationPreset(i int) {
	sim.simulationPresets.Presets = slices.Delete(sim.simulationPresets.Presets, i, i+1)

	sim.saveSimulationPresets()
}

func (sim *simulation) handleLoadSimulationPreset(i int) {
	if sim.simulationPresets.shouldLoadSimulation {
		for _, planet := range sim.simulationPresets.Presets[i].Planets {

			sim.planetHandler.planets = append(sim.planetHandler.planets, newPlanet(
				planet.Name,
				planet.X,
				planet.Y,
				planet.Radius,
				planet.Mass,
				planet.Velocity,
				planet.Color,
				sim.planetHandler.planetsOffset,
			))
		}

		sim.running = false
		sim.simulationPresets.shouldLoadSimulation = false
	}
}

func (sim *simulation) loadSimulationPresetsFromFile() {
	if _, err := os.Stat(sim.simulationPresets.filePath); err != nil {
		return
	}

	content, err := os.ReadFile(sim.simulationPresets.filePath)
	if err != nil {
		log.Printf("Failed to read file %s: %v", sim.simulationPresets.filePath, err)
	}

	if err := json.Unmarshal(content, &sim.simulationPresets); err != nil {
		log.Printf("Failed to unmarshal planet presets json: %v", err)
	}
}

func (sim *simulation) saveSimulationPresets() {
	if err := os.MkdirAll(filepath.Dir(sim.simulationPresets.filePath), os.ModePerm); err != nil {
		log.Printf("Failed to create dir for %s file: %v", sim.simulationPresets.filePath, err)
	}

	content, err := json.MarshalIndent(sim.simulationPresets, "", " ")

	if err != nil {
		log.Printf("Failed to marshal planet presets json: %v", err)
	}

	if err := os.WriteFile(sim.simulationPresets.filePath, content, os.ModePerm); err != nil {
		log.Printf("Failed to create file %s: %v", sim.simulationPresets.filePath, err)
	}
}

func (sim *simulation) addPlanetToPlanetPresets(planetToAdd Planet) {
	// replace if same name
	for i, planet := range sim.planetHandler.presets {
		if planet.Name == planetToAdd.Name {
			sim.planetHandler.presets[i] = &planetToAdd
			sim.savePlanetPresetsToFile()
			return
		}
	}

	sim.planetHandler.presets = append(sim.planetHandler.presets, &planetToAdd)

	sim.savePlanetPresetsToFile()
}

func (sim *simulation) savePlanetPresetsToFile() {
	if err := os.MkdirAll(filepath.Dir(sim.planetHandler.presetFilePath), os.ModePerm); err != nil {
		log.Printf("Failed to create dir for %s file: %v", sim.planetHandler.presetFilePath, err)
	}

	content, err := json.MarshalIndent(sim.planetHandler.presetFilePath, "", " ")
	if err != nil {
		log.Printf("Failed to marshal planet presets json: %v", err)
	}

	if err := os.WriteFile(sim.planetHandler.presetFilePath, content, os.ModePerm); err != nil {
		log.Printf("Failed to create file %s: %v", sim.planetHandler.presetFilePath, err)
	}
}

func (sim *simulation) loadPlanetPresetsFromFile() {
	if _, err := os.Stat(sim.planetHandler.presetFilePath); err != nil {
		return
	}

	content, err := os.ReadFile(sim.planetHandler.presetFilePath)
	if err != nil {
		log.Printf("Failed to read file %s: %v", sim.planetHandler.presetFilePath, err)
	}

	if err := json.Unmarshal(content, &sim.planetHandler.presets); err != nil {
		log.Printf("Failed to unmarshal planet presets json: %v", err)
	}
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

func (sim *simulation) spawnPlanet() {
	// check if would collide on spawn
	for _, planet := range sim.planetHandler.planets {
		toCreatePlanet := sim.planetHandler.planetCreator.planet
		if _, _, _, overlaps := overlapsCircle(planet.X, toCreatePlanet.X, planet.Y, toCreatePlanet.Y, planet.Radius, toCreatePlanet.Radius); overlaps {
			return
		}
	}

	planetInCreator := sim.planetHandler.planetCreator.planet
	newPlanet := newPlanet(
		planetInCreator.Name,
		planetInCreator.X,
		planetInCreator.Y,
		planetInCreator.Radius,
		planetInCreator.Mass,
		planetInCreator.Velocity,
		planetInCreator.Color,
		sim.planetHandler.planetsOffset,
	)

	sim.planetHandler.planets = append(sim.planetHandler.planets, newPlanet)
	sim.planetHandler.planetCounter++

	// make planetCreator planet highlight invisible
	sim.planetHandler.planetCreator.showPlanet = false

	// reset name change of planetCreator
	sim.planetHandler.planetCreator.planet.HasNameChanged = false

	// select planet if none other planet is selected
	if !sim.planetHandler.selectedPlanet.isSelected {
		// should be last element appended
		sim.planetHandler.selectedPlanet.index = len(sim.planetHandler.planets) - 1
		sim.planetHandler.selectedPlanet.isSelected = true
		return
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

func (sim *simulation) handlePlanetDeletion() {
	if len(sim.planetHandler.planetsToRemove) > 0 {
		for _, planetIndex := range sim.planetHandler.planetsToRemove {
			// remove from planets
			if sim.planetHandler.selectedPlanet.index == planetIndex {
				sim.planetHandler.selectedPlanet.isSelected = false
			}
			if sim.planetHandler.focusedPlanet.index == planetIndex {
				sim.planetHandler.focusedPlanet.isFocused = false
			}

			sim.planetHandler.planets = slices.Delete(sim.planetHandler.planets, planetIndex, planetIndex+1)

		}

		sim.planetHandler.planetsToRemove = []int{}
	}
}

func (sim *simulation) updatePlanets() {
	for _, planet := range sim.planetHandler.planets {
		planet.Update(sim, sim.planetHandler.planets)
		if sim.planetHandler.focusedPlanet.isFocused {
			planet.focus(sim)
		}
	}
}

func (sim *simulation) Update() {
	ebiten.SetTPS(sim.tps)

	sim.handleReset()
	sim.handlePlanetDeletion()
	sim.updatePlanets()
	sim.handleLoadSimulationPreset(sim.simulationPresets.presetIndex)
}

func (sim *simulation) removeSelectedPlanet(ui *ui) {
	sim.planetHandler.planetsToRemove = append(sim.planetHandler.planetsToRemove, sim.planetHandler.selectedPlanet.index)
	sim.planetHandler.selectedPlanet.isSelected = false
}

func (sim *simulation) updateToCreatePlanet(x float64, y float64) {
	if !sim.planetHandler.planetCreator.planet.HasNameChanged {
		sim.planetHandler.planetCreator.planet.Name = fmt.Sprintf("Planet %d", sim.planetHandler.planetCounter+1)
	}
	// set x
	sim.planetHandler.planetCreator.planet.X = x
	sim.planetHandler.planetCreator.planet.Y = y

	planet := sim.planetHandler.planetCreator.planet
	radius := float32(planet.Radius)

	r, g, b, _ := convertColorToInt(planet.Color)

	transparentColor := SetColor(uint8(r), uint8(g), uint8(b), 100)
	sim.planetHandler.planetCreator.planet.image = ebiten.NewImage(int(planet.Radius*2), int(planet.Radius*2))
	vector.FillCircle(sim.planetHandler.planetCreator.planet.image, radius, radius, radius, transparentColor, true)

	planet.geometry.Reset()
	// center planet
	planet.geometry.Translate(planet.X-float64(planet.Radius), planet.Y-float64(planet.Radius))
	// adjust for offset
	planet.geometry.Translate(sim.planetHandler.planetsOffset[0], sim.planetHandler.planetsOffset[1])

	sim.planetHandler.planetCreator.planet.geometry = planet.geometry
}

func (sim *simulation) drawToCreatePlanet(screen *ebiten.Image) {
	screen.DrawImage(sim.planetHandler.planetCreator.planet.image, &ebiten.DrawImageOptions{
		GeoM: sim.planetHandler.planetCreator.planet.geometry,
	})
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
