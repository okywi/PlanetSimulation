package planetsimulation

import (
	"log"
	"math/rand"
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

func (controls *controls) selectPlanetIfPossible(sim *simulation, x int, y int) bool {
	for _, planet := range sim.planets {
		isSelectedX := false
		isSelectedY := false

		if x >= int(planet.x)-int(planet.radius) && x <= int(planet.x)+int(planet.radius) {
			isSelectedX = true
		}

		if y >= int(planet.y)-int(planet.radius) && y <= int(planet.y)+int(planet.radius) {
			isSelectedY = true
		}

		if isSelectedX && isSelectedY {
			log.Println("Selected planet with color: ", planet.color)
			sim.selectedPlanet = planet
			return true
		}
	}

	return false
}

func (controls *controls) handlePlanetCreation(sim *simulation) {
	mouseX, mouseY := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) && !controls.mouseButtonsPressed[0] {
		controls.mouseButtonsPressed[ebiten.MouseButton0] = true

		selectedX := mouseX - sim.screen.offset[0]
		selectedY := mouseY - sim.screen.offset[1]
		log.Println(selectedX, selectedY)
		hasSelected := controls.selectPlanetIfPossible(sim, selectedX, selectedY)
		if hasSelected {
			return
		}

		newPlanet := createPlanet(
			float64(selectedX),
			float64(selectedY),
			sim.planetRadius,
			sim.planetMass,
			vector2{1, 200},
			SetColor(uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255),
			sim.screen.offset,
		)
		sim.planets = append(sim.planets, newPlanet)
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

		// update offsets
		sim.screen.offset[0] += dx
		sim.screen.offset[1] += dy

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
