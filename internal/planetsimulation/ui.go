package planetsimulation

import (
	"image"
	"slices"
	"strconv"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type ui struct {
	debugui             debugui.DebugUI
	title               string
	hasFocus            debugui.InputCapturingState
	ctx                 *debugui.Context
	layouts             []image.Rectangle
	hasRemovedPlanet    bool
	pauseSimulationText string
}

func newUI() *ui {
	ui := &ui{
		debugui:             debugui.DebugUI{},
		title:               "Simulation",
		hasFocus:            debugui.InputCapturingState(0),
		hasRemovedPlanet:    false,
		pauseSimulationText: "Pause simulation",
	}

	return ui
}

func (ui *ui) formatFloat(v float64, n int) string {
	return strconv.FormatFloat(v, 'f', n, 64)
}

func (ui *ui) getCoords(game *Game) []float64 {
	return []float64{
		-(game.simulation.planetsOffset[0] - game.simulation.defaultOffset[0]),
		game.simulation.planetsOffset[1] - game.simulation.defaultOffset[1],
	}

}

func (ui *ui) Update(game *Game) error {
	ui.layouts = slices.Delete(ui.layouts, 0, len(ui.layouts))
	var err error
	ui.hasFocus, err = ui.debugui.Update(func(ctx *debugui.Context) error {
		// set global context
		ui.ctx = ctx

		ui.createSystemWindow(ctx, game)
		ui.createPlanetWindow(ctx, game)
		ui.modifyPlanetWindow(ctx, game)
		return err
	})
	return err
}

func (ui *ui) createSystemWindow(ctx *debugui.Context, game *Game) {
	ctx.Window(ui.title, image.Rect(0, 0, 250, 320), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		ctx.Header("Performance", true, func() {
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("FPS: ")
				ctx.Text(ui.formatFloat(ebiten.ActualFPS(), 2))
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("Current TPS: ")
				ctx.Text(ui.formatFloat(ebiten.ActualTPS(), 2))
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("Target TPS: ")
				ctx.NumberField(&game.simulation.tps, 2.0)
			})

		})
		ctx.Header("Coordinates", true, func() {
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("x:")
				ctx.Text(ui.formatFloat(float64(ui.getCoords(game)[0]), 1))
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("y:")
				ctx.Text(ui.formatFloat(float64(ui.getCoords(game)[1]), 1))
			})
		})
		ctx.Header("Constants", true, func() {
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("Gravitational Constant:")
				ctx.NumberFieldF(&game.simulation.gravitationalConstant, 0.1, 2)
			})
		})

		ctx.Button(ui.pauseSimulationText).On(func() {
			game.simulation.running = !game.simulation.running
		})
		if game.simulation.running {
			ui.pauseSimulationText = "Pause simulation"
		} else {
			ui.pauseSimulationText = "Resume simulation"
		}

		ctx.Button("Return to origin point").On(func() {
			game.simulation.returnToOrigin()
		})

		ctx.Button("Clear all traces").On(func() {
			for _, planet := range game.simulation.planets {
				planet.clearTraces()
			}
		})

		ctx.Button("Reset Simulation").On(func() {
			game.simulation.shouldReset = true
		})
	})
}

func (ui *ui) createPlanetWindow(ctx *debugui.Context, game *Game) {
	ctx.Window("Create Planet", image.Rect(0, 340, 250, 620), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("x: ")
			ctx.NumberFieldF(&game.simulation.planetCreator.planet.x, 1.0, 1)
		})
		// fake negate
		y := -game.simulation.planetCreator.planet.y
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("y: ")
			ctx.NumberFieldF(&y, 1.0, 1).On(func() {
				game.simulation.planetCreator.planet.y = -y
			})
		})
		radius := game.simulation.planetCreator.planet.radius
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("radius: ")
			ctx.NumberFieldF(&radius, 1.0, 1).On(func() {
				if radius > 0 && radius < 1000 {
					game.simulation.planetCreator.planet.radius = radius
				}
			})
		})
		mass := game.simulation.planetCreator.planet.mass
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("mass: ")
			ctx.NumberFieldF(&mass, 1.0, 1).On(func() {
				if mass > 0 {
					game.simulation.planetCreator.planet.mass = mass
				}
			})
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("velocity x: ")
			ctx.NumberFieldF(&game.simulation.planetCreator.planet.velocity.x, 1.0, 1)
		})
		// fake negate
		velocityY := -game.simulation.planetCreator.planet.velocity.y
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("velocity y: ")
			ctx.NumberFieldF(&velocityY, 1.0, 1).On(func() {
				game.simulation.planetCreator.planet.velocity.y = -velocityY
			})
		})
		ctx.Header("Color", true, func() {
			r, g, b, _ := convertColorToInt(game.simulation.planetCreator.planet.color)
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-3, -1}, []int{59})
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.SetGridLayout([]int{-1, -6}, []int{15, 15, 15})
					ctx.Text("r: ")
					ctx.Slider(&r, 0, 255, 1)
					ctx.Text("g: ")
					ctx.Slider(&g, 0, 255, 1)
					ctx.Text("b: ")
					ctx.Slider(&b, 0, 255, 1)
					game.simulation.planetCreator.planet.color = SetColor(uint8(r), uint8(g), uint8(b), 255)
				})
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.DrawOnlyWidget(func(screen *ebiten.Image) {
						cx := float32(bounds.Min.X) + float32(bounds.Dx())/2
						cy := float32(bounds.Min.Y) + float32(bounds.Dy())/2
						r := float32(bounds.Dx()) / 2
						vector.FillCircle(screen, cx, cy, r, game.simulation.planetCreator.planet.color, true)
						game.simulation.updateToCreatePlanet(game.simulation.planetCreator.planet.x, game.simulation.planetCreator.planet.y)
					})
				})
			})
		})
		ctx.Button("Spawn").On(func() {
			game.simulation.spawnPlanet()
		})
	})
}

func (ui *ui) modifyPlanetWindow(ctx *debugui.Context, game *Game) {
	if !game.simulation.isPlanetSelected {
		return
	}
	if slices.Contains(game.simulation.planetsToRemove, game.simulation.selectedPlanetIndex) {
		return
	}

	selectedPlanet := game.simulation.planets[game.simulation.selectedPlanetIndex]
	ctx.Window("Modify Planet", image.Rect(000, 640, 250, 1075), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)

		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2}, []int{-1, -1})
			x := selectedPlanet.x
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -2}, []int{-1})
				ctx.Text("x: ")
				ctx.NumberFieldF(&x, 1.0, 1).On(func() {
					selectedPlanet.setPosition(x, selectedPlanet.y)
				})
			})
			// fake negate
			y := -selectedPlanet.y
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -2}, []int{-1})
				ctx.Text("y: ")
				ctx.NumberFieldF(&y, 1.0, 1).On(func() {
					selectedPlanet.setPosition(selectedPlanet.x, -y)
				})
			})
		})
		radius := selectedPlanet.radius
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("radius: ")
			ctx.NumberFieldF(&radius, 1.0, 1).On(func() {
				if radius > 0 && radius < 1000 {
					selectedPlanet.radius = radius
				}
				selectedPlanet.updateImage()
			})
		})
		mass := selectedPlanet.mass
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("mass: ")
			ctx.NumberFieldF(&mass, 1.0, 1).On(func() {
				if mass > 0 {
					selectedPlanet.mass = mass
				}
			})
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("velocity x: ")
			ctx.NumberFieldF(&selectedPlanet.velocity.x, 1.0, 1)
		})
		// fake negate
		velocityY := -selectedPlanet.velocity.y
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("velocity y: ")
			ctx.NumberFieldF(&velocityY, 1.0, 1).On(func() {
				selectedPlanet.velocity.y = -velocityY
			})
		})
		ctx.Button("Focus Planet").On(func() {
			game.simulation.focusedPlanetIndex = game.simulation.selectedPlanetIndex
			game.simulation.isPlanetFocused = true
		})
		ctx.Header("Color", true, func() {
			r, g, b := selectedPlanet.getColor()
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-3, -1}, []int{59})
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.SetGridLayout([]int{-1, -6}, []int{15, 15, 15})
					ctx.Text("r: ")
					ctx.Slider(&r, 0, 255, 1).On(func() {
						red := uint8(r)
						selectedPlanet.changeColor(ColorDelta{R: &red})
					})
					ctx.Text("g: ")
					ctx.Slider(&g, 0, 255, 1).On(func() {
						green := uint8(g)
						selectedPlanet.changeColor(ColorDelta{G: &green})
					})
					ctx.Text("b: ")
					ctx.Slider(&b, 0, 255, 1).On(func() {
						blue := uint8(b)
						selectedPlanet.changeColor(ColorDelta{B: &blue})
					})
				})
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.DrawOnlyWidget(func(screen *ebiten.Image) {
						cx := float32(bounds.Min.X) + float32(bounds.Dx())/2
						cy := float32(bounds.Min.Y) + float32(bounds.Dy())/2
						r := float32(bounds.Dx()) / 2
						vector.FillCircle(screen, cx, cy, r, selectedPlanet.color, true)
					})
				})
			})
		})
		ctx.Header("Traces", true, func() {
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-3, -2}, []int{-1})
				ctx.Text("trace every Nth tick:")
				ctx.Slider(&selectedPlanet.traceEveryNTick, 1, 15, 1)
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-3, -2}, []int{-1})
				ctx.Text("draw every Nth tick:")
				ctx.Slider(&selectedPlanet.drawEveryNTick, 1, 15, 1)
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-3, -2}, []int{-1})
				ctx.Text("trace width:")
				ctx.SliderF(&selectedPlanet.traceWidth, 1, 15, 1, 0)
			})
			ctx.Button("Clear Traces").On(func() {
				selectedPlanet.clearTraces()
			})
		})

		ctx.Text("")
		ctx.Button("Remove Planet").On(func() {
			if ui.hasRemovedPlanet {
				ui.hasRemovedPlanet = false
				return
			}
			game.simulation.removeSelectedPlanet(game.ui)
			ui.hasRemovedPlanet = true
		})
	})
}

func (ui *ui) Draw(screen *ebiten.Image) {
	ui.debugui.Draw(screen)
}
