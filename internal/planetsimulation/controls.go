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

func (controls *controls) selectPlanetIfPossible(ui *ui, sim *simulation, x int, y int) bool {
	for i, planet := range sim.planets {
		// deselect when already selected
		if sim.isPlanetSelected {
			selectedPlanet := sim.planets[sim.selectedPlanetIndex]
			isSelectedAgain := overlapsXY(
				x, y,
				int(selectedPlanet.x)-int(selectedPlanet.radius), int(selectedPlanet.x)+int(selectedPlanet.radius),
				int(selectedPlanet.y)-int(selectedPlanet.radius), int(selectedPlanet.y)+int(selectedPlanet.radius),
			)
			if isSelectedAgain {
				sim.isPlanetSelected = false
				return true
			}
		}

		isSelected := overlapsXY(
			x, y,
			int(planet.x)-int(planet.radius), int(planet.x)+int(planet.radius),
			int(planet.y)-int(planet.radius), int(planet.y)+int(planet.radius),
		)

		if isSelected {
			sim.selectedPlanetIndex = i
			sim.isPlanetSelected = true
			return true
		}
	}

	return false
}

func (controls *controls) Update(sim *simulation, ui *ui) {
	controls.handlePlanetCreation(sim, ui)
	controls.handleMovement(sim)
	controls.handlePausing(sim)
}

func (controls *controls) checkUIFocusLayouts(ui *ui, mouseX int, mouseY int) bool {
	for _, rect := range ui.layouts {
		if mouseX >= rect.Min.X && mouseX <= rect.Max.X && mouseY >= rect.Min.Y && mouseY <= rect.Max.Y {
			return true
		}
	}

	return false
}

func (controls *controls) handlePlanetCreation(sim *simulation, ui *ui) {
	mouseX, mouseY := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) && !controls.mouseButtonsPressed[0] {
		controls.mouseButtonsPressed[ebiten.MouseButton0] = true

		// check if ui focused/hovered
		if ui.hasFocus == 1 || controls.checkUIFocusLayouts(ui, mouseX, mouseY) {
			return
		}

		selectedX := float64(mouseX) - sim.planetsOffset[0]
		selectedY := float64(mouseY) - sim.planetsOffset[1]

		if controls.selectPlanetIfPossible(ui, sim, int(selectedX), int(selectedY)) {
			return
		}

		sim.updateToCreatePlanet(float64(selectedX), float64(selectedY))
		sim.planetCreator.showPlanet = true
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		controls.mouseButtonsPressed[ebiten.MouseButton0] = false
	}
}

func (controls *controls) handlePausing(sim *simulation) {
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		if slices.Contains(controls.keysPressed, ebiten.KeySpace) {
			return
		}

		sim.running = !sim.running
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

func (controls *controls) handleMovement(sim *simulation) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton1) {
		ebiten.SetCursorShape(ebiten.CursorShapeMove)
		// clear focused planet
		sim.isPlanetFocused = false

		// get difference of mouse positions
		currentMousePosition := make([]int, 2)
		currentMousePosition[0], currentMousePosition[1] = ebiten.CursorPosition()

		dx := currentMousePosition[0] - controls.previousMousePosition[0]
		dy := currentMousePosition[1] - controls.previousMousePosition[1]
		// update offsets
		sim.planetsOffset[0] += float64(dx)
		sim.planetsOffset[1] += float64(dy)

		// move planet images
		for _, planet := range sim.planets {
			planet.geometry.Translate(float64(dx), float64(dy))
		}

		controls.previousMousePosition = currentMousePosition
	} else {
		controls.previousMousePosition[0], controls.previousMousePosition[1] = ebiten.CursorPosition()
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}
}
