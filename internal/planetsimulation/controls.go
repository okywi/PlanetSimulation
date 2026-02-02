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
			isSelectedAgain := overlaps(
				x, y,
				int(selectedPlanet.x)-int(selectedPlanet.radius), int(selectedPlanet.x)+int(selectedPlanet.radius),
				int(selectedPlanet.y)-int(selectedPlanet.radius), int(selectedPlanet.y)+int(selectedPlanet.radius),
			)
			if isSelectedAgain {
				sim.isPlanetSelected = false
				return true
			}
		}

		isSelected := overlaps(
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

func (controls *controls) handlePlanetCreation(sim *simulation, ui *ui) {
	mouseX, mouseY := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) && !controls.mouseButtonsPressed[0] {
		controls.mouseButtonsPressed[ebiten.MouseButton0] = true

		// check if ui focused/hovered
		if ui.hasFocus == 1 {
			return
		}

		selectedX := float64(mouseX) - sim.screen.offset[0]
		selectedY := float64(mouseY) - sim.screen.offset[1]

		if controls.selectPlanetIfPossible(ui, sim, int(selectedX), int(selectedY)) {
			return
		}

		sim.updateToCreatePlanet(float64(selectedX), float64(selectedY))
		sim.toCreatePlanet.shown = true
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		controls.mouseButtonsPressed[ebiten.MouseButton0] = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		sim.shouldReset = true
	}

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

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton1) {
		ebiten.SetCursorShape(ebiten.CursorShapeMove)
		currentMousePosition := []int{0, 0}
		currentMousePosition[0], currentMousePosition[1] = ebiten.CursorPosition()

		dx := currentMousePosition[0] - controls.previousMousePosition[0]
		dy := currentMousePosition[1] - controls.previousMousePosition[1]

		// clear focused planet
		sim.isPlanetFocused = false

		// update offsets
		sim.screen.offset[0] += float64(dx)
		sim.screen.offset[1] += float64(dy)

		// move planets
		for _, planet := range sim.planets {
			planet.geometry.Translate(float64(dx), float64(dy))
		}

		controls.previousMousePosition = currentMousePosition
	} else {
		controls.previousMousePosition[0], controls.previousMousePosition[1] = ebiten.CursorPosition()
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}
}
