package planetsimulation

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

type planetHandler struct {
	planets               []*Planet
	planetPresets         *planetPresets
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
		planetPresets:         newPlanetPresets(),
		planetsToRemove:       make([]int, 0),
		planetCounter:         0,
		gravitationalConstant: 10000.0,
		running:               true,
	}
	planetHandler.planetsOffset = []float64{planetHandler.defaultPlanetsOffset[0], planetHandler.defaultPlanetsOffset[1]}

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

func (handler *planetHandler) deletePlanet(index int) {
	handler.planetsToRemove = append(handler.planetsToRemove, index)
}

func (handler *planetHandler) mergePlanets(p *Planet, otherPlanet *Planet) {
	// merge planets
	if p.Mass >= otherPlanet.Mass {
		handler.deletePlanet(slices.Index(handler.planets, otherPlanet))
		p.Mass += otherPlanet.Mass / 2
		if p.Radius <= 1000 {
			p.Radius += otherPlanet.Radius / 4
		}

		p.Velocity = p.Velocity.add(vector2{
			((otherPlanet.Velocity.X) / p.Mass),
			((otherPlanet.Velocity.Y) / p.Mass),
		})
		p.updateImage()
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

func (handler *planetHandler) deleteSelectedPlanet() {
	handler.deletePlanet(handler.selectedPlanet.index)
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
