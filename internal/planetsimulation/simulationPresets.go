package planetsimulation

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"
)

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

func (presets *simulationPresets) saveSimulationPreset(planetHandler *planetHandler) {
	planets := []*Planet{}
	for _, planet := range planetHandler.planets {
		planets = append(planets, *&planet)
	}

	presets.Presets = append(presets.Presets, &simulationPreset{
		Name:    presets.newPresetName,
		Planets: planets,
	})
	presets.saveSimulationPresetsToFile()
}

func (presets *simulationPresets) removeSimulationPreset(i int) {
	presets.Presets = slices.Delete(presets.Presets, i, i+1)

	presets.saveSimulationPresetsToFile()
}

func (presets *simulationPresets) handleLoadSimulationPreset(planetHandler *planetHandler, i int) {
	if presets.shouldLoadSimulation {
		for _, planet := range presets.Presets[i].Planets {

			planetHandler.planets = append(planetHandler.planets, newPlanet(
				planet.Name,
				planet.X,
				planet.Y,
				planet.Radius,
				planet.Mass,
				planet.Velocity,
				planet.Color,
				planetHandler.planetsOffset,
			))
		}

		planetHandler.running = false
		presets.shouldLoadSimulation = false
	}
}

func (presets *simulationPresets) loadSimulationPresetsFromFile() {
	filePath := presets.filePath
	if _, err := os.Stat(filePath); err != nil {
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read file %s: %v", filePath, err)
	}

	if err := json.Unmarshal(content, &presets); err != nil {
		log.Printf("Failed to unmarshal planet presets json: %v", err)
	}
}

func (presets *simulationPresets) saveSimulationPresetsToFile() {
	filePath := presets.filePath

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		log.Printf("Failed to create dir for %s file: %v", filePath, err)
	}

	content, err := json.MarshalIndent(presets, "", " ")

	if err != nil {
		log.Printf("Failed to marshal planet presets json: %v", err)
	}

	if err := os.WriteFile(filePath, content, os.ModePerm); err != nil {
		log.Printf("Failed to create file %s: %v", filePath, err)
	}
}
