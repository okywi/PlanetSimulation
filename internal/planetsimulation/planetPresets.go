package planetsimulation

import (
	"encoding/json"
	"log"
	"slices"
)

type planetPresets struct {
	presets  []*Planet
	filePath string
}

func newPlanetPresets() *planetPresets {
	planetPresets := &planetPresets{
		filePath: "assets/data/planet_presets.json",
	}

	planetPresets.loadFromFile()
	return planetPresets
}

func (planetPresets *planetPresets) saveToFile() {
	content, err := json.MarshalIndent(planetPresets.filePath, "", " ")
	if err != nil {
		log.Printf("Failed to marshal planet presets json: %v", err)
	}

	writeFile(planetPresets.filePath, content)
}

func (planetPresets *planetPresets) loadFromFile() {
	content := readFile(planetPresets.filePath)

	if err := json.Unmarshal(content, &planetPresets); err != nil {
		log.Printf("Failed to unmarshal planet presets json: %v", err)
	}
}

func (planetPresets *planetPresets) addPlanet(planetToAdd Planet) {
	// replace if same name
	for i, planet := range planetPresets.presets {
		if planet.Name == planetToAdd.Name {
			planetPresets.presets[i] = &planetToAdd
			planetPresets.saveToFile()
			return
		}
	}

	planetPresets.presets = append(planetPresets.presets, &planetToAdd)

	planetPresets.saveToFile()
}

func (planetPresets *planetPresets) deletePreset(index int) {
	planetPresets.presets = slices.Delete(planetPresets.presets, index, index+1)
}
