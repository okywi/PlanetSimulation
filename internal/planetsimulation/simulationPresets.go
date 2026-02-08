package planetsimulation

import (
	"encoding/json"
	"log"
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

func newSimulationPresets() *simulationPresets {
	simulationPresets := &simulationPresets{
		filePath: "assets/data/simulation_presets.json",
	}
	simulationPresets.loadSimulationPresetsFromFile()

	return simulationPresets
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
	content := readFile(presets.filePath)

	if err := json.Unmarshal(content, &presets); err != nil {
		log.Printf("Failed to unmarshal planet presets json: %v", err)
	}
}

func (presets *simulationPresets) saveSimulationPresetsToFile() {
	content, err := json.MarshalIndent(presets, "", " ")

	if err != nil {
		log.Printf("Failed to marshal planet presets json: %v", err)
	}

	writeFile(presets.filePath, content)
}
