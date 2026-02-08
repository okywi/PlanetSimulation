package planetsimulation

import (
	"encoding/json"
	"log"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

type planetHandler struct {
	planets               []*Planet
	presets               []*Planet
	presetFilePath        string
	planetsOffset         []float64
	planetCounter         int
	planetCreator         *planetCreator
	planetsToRemove       []int
	defaultPlanetsOffset  []float64
	selectedPlanet        selectedPlanet
	focusedPlanet         focusedPlanet
	gravitationalConstant float64
	running               bool
}

type focusedPlanet struct {
	isFocused bool
	index     int
}

type selectedPlanet struct {
	index      int
	isSelected bool
}

func newPlanetHandler(gameSize []int) *planetHandler {
	// planet that is created by a click
	planetHandler := &planetHandler{
		planetCreator:         newPlanetCreator(),
		defaultPlanetsOffset:  []float64{float64(gameSize[0]) / 2, float64(gameSize[1] / 2)},
		presetFilePath:        "assets/data/planet_presets.json",
		planetsToRemove:       make([]int, 0),
		planetCounter:         0,
		gravitationalConstant: 10000.0,
		running:               true,
	}
	planetHandler.planetsOffset = []float64{planetHandler.defaultPlanetsOffset[0], planetHandler.defaultPlanetsOffset[1]}
	planetHandler.loadPlanetPresetsFromFile()

	return planetHandler
}

func (handler *planetHandler) handlePlanetDeletion() {
	if len(handler.planetsToRemove) > 0 {
		for _, planetIndex := range handler.planetsToRemove {
			// remove from planets
			if handler.selectedPlanet.index == planetIndex {
				handler.selectedPlanet.isSelected = false
			}
			if handler.focusedPlanet.index == planetIndex {
				handler.focusedPlanet.isFocused = false
			}

			handler.planets = slices.Delete(handler.planets, planetIndex, planetIndex+1)

		}

		handler.planetsToRemove = []int{}
	}
}

func (handler *planetHandler) updatePlanets() {
	if !handler.running {
		return
	}
	for _, planet := range handler.planets {
		planet.Update(handler)
		if handler.focusedPlanet.isFocused {
			planet.focus(handler)
		}
	}
}

func (handler *planetHandler) selectPlanet(planetIndex int) {
	handler.selectedPlanet.index = planetIndex
	handler.selectedPlanet.isSelected = true
}

func (handler *planetHandler) focusPlanet(planetIndex int) {
	handler.focusedPlanet.index = planetIndex
	handler.focusedPlanet.isFocused = true
}

func (handler *planetHandler) removeSelectedPlanet() {
	handler.planetsToRemove = append(handler.planetsToRemove, handler.selectedPlanet.index)
	handler.selectedPlanet.isSelected = false
}

func (handler *planetHandler) returnToOrigin() {
	// reset planetsOffset
	dx := handler.planetsOffset[0] - handler.defaultPlanetsOffset[0]
	dy := handler.planetsOffset[1] - handler.defaultPlanetsOffset[1]
	handler.planetsOffset[0] -= dx
	handler.planetsOffset[1] -= dy

	// move planet images as well
	for _, planet := range handler.planets {
		planet.geometry.Translate(float64(-dx), float64(-dy))
	}
}

func (handler *planetHandler) savePlanetPresetsToFile() {
	content, err := json.MarshalIndent(handler.presetFilePath, "", " ")
	if err != nil {
		log.Printf("Failed to marshal planet presets json: %v", err)
	}

	writeFile(handler.presetFilePath, content)
}

func (handler *planetHandler) loadPlanetPresetsFromFile() {
	content := readFile(handler.presetFilePath)

	if err := json.Unmarshal(content, &handler.presets); err != nil {
		log.Printf("Failed to unmarshal planet presets json: %v", err)
	}
}

func (handler *planetHandler) addPlanetToPresets(planetToAdd Planet) {
	// replace if same name
	for i, planet := range handler.presets {
		if planet.Name == planetToAdd.Name {
			handler.presets[i] = &planetToAdd
			handler.savePlanetPresetsToFile()
			return
		}
	}

	handler.presets = append(handler.presets, &planetToAdd)

	handler.savePlanetPresetsToFile()
}

func (handler *planetHandler) Update() {
	handler.handlePlanetDeletion()
	handler.updatePlanets()
}

func (handler *planetHandler) Draw(simScreen *ebiten.Image) {
	// draw planets
	for _, planet := range handler.planets {
		if planet != nil {
			planet.Draw(simScreen)
		}
	}

	// draw planetCreator
	if handler.planetCreator.showPlanet {
		simScreen.DrawImage(handler.planetCreator.planet.image, &ebiten.DrawImageOptions{
			GeoM: handler.planetCreator.planet.geometry,
		})
	}
}
