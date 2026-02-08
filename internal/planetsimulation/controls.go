package planetsimulation

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

type controls struct {
	mouseButtonsPressed   []bool
	wheelMultiplier       float64
	previousMousePosition []int
	keysPressed           []ebiten.Key
}

func newControls() *controls {
	return &controls{
		mouseButtonsPressed:   make([]bool, 3),
		wheelMultiplier:       2.5,
		previousMousePosition: []int{0, 0},
	}
}

func (controls *controls) selectPlanetIfPossible(ui *ui, planetHandler *planetHandler, x int, y int) bool {
	for i, planet := range planetHandler.planets {
		// deselect when already selected
		if planetHandler.selectedPlanet.isSelected {
			selectedPlanet := planetHandler.planets[planetHandler.selectedPlanet.index]
			isSelectedAgain := overlapsXY(
				x, y,
				int(selectedPlanet.X)-int(selectedPlanet.Radius), int(selectedPlanet.X)+int(selectedPlanet.Radius),
				int(selectedPlanet.Y)-int(selectedPlanet.Radius), int(selectedPlanet.Y)+int(selectedPlanet.Radius),
			)
			if isSelectedAgain {
				planetHandler.selectedPlanet.isSelected = false
				return true
			}
		}

		isSelected := overlapsXY(
			x, y,
			int(planet.X)-int(planet.Radius), int(planet.X)+int(planet.Radius),
			int(planet.Y)-int(planet.Radius), int(planet.Y)+int(planet.Radius),
		)

		if isSelected {
			planetHandler.selectedPlanet.index = i
			planetHandler.selectedPlanet.isSelected = true
			return true
		}
	}

	return false
}

func (controls *controls) isUiFocused(ui *ui) bool {
	mouseX, mouseY := ebiten.CursorPosition()
	// check if ui focused/hovered
	if ui.hasFocus != 0 || controls.checkUIFocusLayouts(ui, mouseX, mouseY) {
		return true
	}

	return false
}

func (controls *controls) Update(planetHandler *planetHandler, ui *ui) {
	if !controls.isUiFocused(ui) {
		controls.handlePlanetCreation(planetHandler, ui)
		controls.handleMovement(planetHandler, ui)
		controls.handlePausing(planetHandler, ui)
	}
}

func (controls *controls) checkUIFocusLayouts(ui *ui, mouseX int, mouseY int) bool {
	for _, rect := range ui.layouts {
		if mouseX >= rect.Min.X && mouseX <= rect.Max.X && mouseY >= rect.Min.Y && mouseY <= rect.Max.Y {
			return true
		}
	}

	return false
}

func (controls *controls) handlePlanetCreation(planetHandler *planetHandler, ui *ui) {
	mouseX, mouseY := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) && !controls.mouseButtonsPressed[0] {
		controls.mouseButtonsPressed[ebiten.MouseButton0] = true

		selectedX := float64(mouseX) - planetHandler.planetsOffset[0]
		selectedY := float64(mouseY) - planetHandler.planetsOffset[1]

		if controls.selectPlanetIfPossible(ui, planetHandler, int(selectedX), int(selectedY)) {
			return
		}

		planetHandler.updateToCreatePlanet(float64(selectedX), float64(selectedY))
		planetHandler.planetCreator.showPlanet = true
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		controls.mouseButtonsPressed[ebiten.MouseButton0] = false
	}
}

func (controls *controls) handlePausing(planetHandler *planetHandler, ui *ui) {
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		if slices.Contains(controls.keysPressed, ebiten.KeySpace) {
			return
		}

		planetHandler.running = !planetHandler.running
		controls.keysPressed = append(controls.keysPressed, ebiten.KeySpace)
	}

	if !ebiten.IsKeyPressed(ebiten.KeySpace) {
		controls.keysPressed = slices.DeleteFunc(controls.keysPressed, func(key ebiten.Key) bool {
			if key == ebiten.KeySpace {
				return true
			}
			return false
		})
	}
}

func (controls *controls) handleMovement(planetHandler *planetHandler, ui *ui) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton1) {
		ebiten.SetCursorShape(ebiten.CursorShapeMove)
		// clear focused planet
		planetHandler.focusedPlanet.isFocused = false

		// get difference of mouse positions
		currentMousePosition := make([]int, 2)
		currentMousePosition[0], currentMousePosition[1] = ebiten.CursorPosition()

		dx := currentMousePosition[0] - controls.previousMousePosition[0]
		dy := currentMousePosition[1] - controls.previousMousePosition[1]
		// update offsets
		planetHandler.planetsOffset[0] += float64(dx)
		planetHandler.planetsOffset[1] += float64(dy)

		// move planet images
		for _, planet := range planetHandler.planets {
			planet.geometry.Translate(float64(dx), float64(dy))
		}

		controls.previousMousePosition = currentMousePosition
	} else {
		controls.previousMousePosition[0], controls.previousMousePosition[1] = ebiten.CursorPosition()
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}
}
