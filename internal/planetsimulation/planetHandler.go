package planetsimulation

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type planetCreator struct {
	planet     *Planet
	showPlanet bool
}

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

func (handler *planetHandler) spawnPlanetAtMouse() {
	// check if would collide on spawn
	for _, planet := range handler.planets {
		toCreatePlanet := handler.planetCreator.planet
		if _, _, _, overlaps := overlapsCircle(planet.X, toCreatePlanet.X, planet.Y, toCreatePlanet.Y, planet.Radius, toCreatePlanet.Radius); overlaps {
			return
		}
	}

	planetInCreator := handler.planetCreator.planet
	newPlanet := newPlanet(
		planetInCreator.Name,
		planetInCreator.X,
		planetInCreator.Y,
		planetInCreator.Radius,
		planetInCreator.Mass,
		planetInCreator.Velocity,
		planetInCreator.Color,
		handler.planetsOffset,
	)

	handler.planets = append(handler.planets, newPlanet)
	handler.planetCounter++

	// make planetCreator planet highlight invisible
	handler.planetCreator.showPlanet = false

	// reset name change of planetCreator
	handler.planetCreator.planet.HasNameChanged = false

	// select planet if none other planet is selected
	if !handler.selectedPlanet.isSelected {
		// should be last element appended
		handler.selectedPlanet.index = len(handler.planets) - 1
		handler.selectedPlanet.isSelected = true
		return
	}
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

func (handler *planetHandler) updateToCreatePlanet(x float64, y float64) {
	planetCreator := handler.planetCreator
	if !planetCreator.planet.HasNameChanged {
		planetCreator.planet.Name = fmt.Sprintf("Planet %d", handler.planetCounter+1)
	}
	// set x
	planetCreator.planet.X = x
	planetCreator.planet.Y = y

	planet := planetCreator.planet
	radius := float32(planet.Radius)

	r, g, b, _ := convertColorToInt(planet.Color)

	transparentColor := SetColor(uint8(r), uint8(g), uint8(b), 100)
	planetCreator.planet.image = ebiten.NewImage(int(planet.Radius*2), int(planet.Radius*2))
	vector.FillCircle(planetCreator.planet.image, radius, radius, radius, transparentColor, true)

	planet.geometry.Reset()
	// center planet
	planet.geometry.Translate(planet.X-float64(planet.Radius), planet.Y-float64(planet.Radius))
	// adjust for offset
	planet.geometry.Translate(handler.planetsOffset[0], handler.planetsOffset[1])

	handler.planetCreator.planet.geometry = planet.geometry
}

func (handler *planetHandler) savePlanetPresetsToFile() {
	if err := os.MkdirAll(filepath.Dir(handler.presetFilePath), os.ModePerm); err != nil {
		log.Printf("Failed to create dir for %s file: %v", handler.presetFilePath, err)
	}

	content, err := json.MarshalIndent(handler.presetFilePath, "", " ")
	if err != nil {
		log.Printf("Failed to marshal planet presets json: %v", err)
	}

	if err := os.WriteFile(handler.presetFilePath, content, os.ModePerm); err != nil {
		log.Printf("Failed to create file %s: %v", handler.presetFilePath, err)
	}
}

func (handler *planetHandler) loadPlanetPresetsFromFile() {
	if _, err := os.Stat(handler.presetFilePath); err != nil {
		return
	}

	content, err := os.ReadFile(handler.presetFilePath)
	if err != nil {
		log.Printf("Failed to read file %s: %v", handler.presetFilePath, err)
	}

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
