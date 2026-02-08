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
