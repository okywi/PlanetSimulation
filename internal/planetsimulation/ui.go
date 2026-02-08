package planetsimulation

import (
	"fmt"
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
		-(game.simulation.planetHandler.planetsOffset[0] - game.simulation.planetHandler.defaultPlanetsOffset[0]),
		game.simulation.planetHandler.planetsOffset[1] - game.simulation.planetHandler.defaultPlanetsOffset[1],
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
		ui.planetListWindow(ctx, game)
		ui.planetPresetsWindow(ctx, game)
		ui.simulationPresetsWindow(ctx, game)
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
			for _, planet := range game.simulation.planetHandler.planets {
				planet.clearTraces()
			}
		})

		ctx.Button("Reset Simulation").On(func() {
			game.simulation.shouldReset = true
		})
	})
}

func (ui *ui) createPlanetWindow(ctx *debugui.Context, game *Game) {
	ctx.Window("Create Planet", image.Rect(0, 325, 250, 645), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("Name: ")
			ctx.TextField(&game.simulation.planetHandler.planetCreator.planet.Name).On(func() {
				game.simulation.planetHandler.planetCreator.planet.HasNameChanged = true
			})
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("x: ")
			ctx.NumberFieldF(&game.simulation.planetHandler.planetCreator.planet.X, 1.0, 1)
		})
		// fake negate
		y := -game.simulation.planetHandler.planetCreator.planet.Y
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("y: ")
			ctx.NumberFieldF(&y, 1.0, 1).On(func() {
				game.simulation.planetHandler.planetCreator.planet.Y = -y
			})
		})
		radius := game.simulation.planetHandler.planetCreator.planet.Radius
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("radius: ")
			ctx.NumberFieldF(&radius, 1.0, 1).On(func() {
				if radius > 0 && radius < 1000 {
					game.simulation.planetHandler.planetCreator.planet.Radius = radius
				}
			})
		})
		mass := game.simulation.planetHandler.planetCreator.planet.Mass
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("mass: ")
			ctx.NumberFieldF(&mass, 1.0, 1).On(func() {
				if mass > 0 {
					game.simulation.planetHandler.planetCreator.planet.Mass = mass
				}
			})
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("velocity x: ")
			ctx.NumberFieldF(&game.simulation.planetHandler.planetCreator.planet.Velocity.X, 1.0, 1)
		})
		// fake negate
		velocityY := -game.simulation.planetHandler.planetCreator.planet.Velocity.Y
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("velocity y: ")
			ctx.NumberFieldF(&velocityY, 1.0, 1).On(func() {
				game.simulation.planetHandler.planetCreator.planet.Velocity.Y = -velocityY
			})
		})
		ctx.Header("Color", true, func() {
			r, g, b, _ := convertColorToInt(game.simulation.planetHandler.planetCreator.planet.Color)
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
					game.simulation.planetHandler.planetCreator.planet.Color = SetColor(uint8(r), uint8(g), uint8(b), 255)
				})
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.DrawOnlyWidget(func(screen *ebiten.Image) {
						cx := float32(bounds.Min.X) + float32(bounds.Dx())/2
						cy := float32(bounds.Min.Y) + float32(bounds.Dy())/2
						r := float32(bounds.Dx()) / 2
						vector.FillCircle(screen, cx, cy, r, game.simulation.planetHandler.planetCreator.planet.Color, true)
						game.simulation.updateToCreatePlanet(game.simulation.planetHandler.planetCreator.planet.X, game.simulation.planetHandler.planetCreator.planet.Y)
					})
				})
			})
		})
		ctx.Button("Save to presets").On(func() {
			game.simulation.addPlanetToPlanetPresets(*game.simulation.planetHandler.planetCreator.planet)
		})
		ctx.Button("Spawn").On(func() {
			game.simulation.spawnPlanet()
		})
	})
}

func (ui *ui) modifyPlanetWindow(ctx *debugui.Context, game *Game) {
	if !game.simulation.planetHandler.selectedPlanet.isSelected || game.simulation.planetHandler.selectedPlanet.index >= len(game.simulation.planetHandler.planets) {
		return
	}

	if slices.Contains(game.simulation.planetHandler.planetsToRemove, game.simulation.planetHandler.selectedPlanet.index) {
		return
	}

	selectedPlanet := game.simulation.planetHandler.planets[game.simulation.planetHandler.selectedPlanet.index]
	ctx.Window("Modify Planet", image.Rect(0, 650, 250, 1015), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("Name: ")
			ctx.TextField(&selectedPlanet.Name).On(func() {
				selectedPlanet.HasNameChanged = true
			})
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2}, []int{-1, -1})
			x := selectedPlanet.X
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -2}, []int{-1})
				ctx.Text("x: ")
				ctx.NumberFieldF(&x, 1.0, 1).On(func() {
					selectedPlanet.setPosition(x, selectedPlanet.Y)
				})
			})
			// fake negate
			y := -selectedPlanet.Y
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -2}, []int{-1})
				ctx.Text("y: ")
				ctx.NumberFieldF(&y, 1.0, 1).On(func() {
					selectedPlanet.setPosition(selectedPlanet.X, -y)
				})
			})
		})
		radius := selectedPlanet.Radius
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("radius: ")
			ctx.NumberFieldF(&radius, 1.0, 1).On(func() {
				if radius > 0 && radius < 1000 {
					selectedPlanet.Radius = radius
				}
				selectedPlanet.updateImage()
			})
		})
		mass := selectedPlanet.Mass
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("mass: ")
			ctx.NumberFieldF(&mass, 1.0, 1).On(func() {
				if mass > 0 {
					selectedPlanet.Mass = mass
				}
			})
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("velocity x: ")
			ctx.NumberFieldF(&selectedPlanet.Velocity.X, 1.0, 1)
		})
		// fake negate
		velocityY := -selectedPlanet.Velocity.Y
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("velocity y: ")
			ctx.NumberFieldF(&velocityY, 1.0, 1).On(func() {
				selectedPlanet.Velocity.Y = -velocityY
			})
		})
		ctx.Button("Focus Planet").On(func() {
			game.simulation.planetHandler.focusedPlanet.index = game.simulation.planetHandler.selectedPlanet.index
			game.simulation.planetHandler.focusedPlanet.isFocused = true
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
						vector.FillCircle(screen, cx, cy, r, selectedPlanet.Color, true)
					})
				})
			})
		})
		ctx.Header("Traces", false, func() {
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-3, -2}, []int{-1})
				ctx.Text("trace every Nth tick:")
				ctx.Slider(&selectedPlanet.TraceEveryNTick, 1, 15, 1)
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-3, -2}, []int{-1})
				ctx.Text("draw every Nth tick:")
				ctx.Slider(&selectedPlanet.DrawEveryNTick, 1, 15, 1)
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-3, -2}, []int{-1})
				ctx.Text("trace width:")
				ctx.SliderF(&selectedPlanet.TraceWidth, 1, 15, 1, 0)
			})
			ctx.Button("Clear Traces").On(func() {
				selectedPlanet.clearTraces()
			})
		})

		ctx.Button("Save to presets").On(func() {
			game.simulation.addPlanetToPlanetPresets(*selectedPlanet)
		})

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

func (ui *ui) planetListWindow(ctx *debugui.Context, game *Game) {
	ctx.Window("Planets", image.Rect(game.screenSize[0]-200, 0, game.screenSize[0], 300), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		for i, planet := range game.simulation.planetHandler.planets {
			ctx.IDScope("grid "+string(i), func() {
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.SetGridLayout([]int{15, -4, 15}, []int{20})
					ctx.DrawOnlyWidget(func(screen *ebiten.Image) {
						cx := float32(bounds.Min.X) + 7
						cy := float32(bounds.Min.Y) + float32(bounds.Dy())/2
						r := float32(8)
						vector.FillCircle(screen, cx, cy, r, planet.Color, true)
					})
					ctx.IDScope("button "+string(i), func() {
						ctx.Button(fmt.Sprintf("%s: %.1f, %.1f", planet.Name, planet.X, planet.Y)).On(func() {
							game.simulation.planetHandler.selectedPlanet.index = i
							game.simulation.planetHandler.selectedPlanet.isSelected = true
							game.simulation.planetHandler.focusedPlanet.index = i
							game.simulation.planetHandler.focusedPlanet.isFocused = true
							game.simulation.planetHandler.planets[i].focus(game.simulation)
						})
					})
					ctx.Button("X").On(func() {
						game.simulation.planetHandler.planetsToRemove = append(game.simulation.planetHandler.planetsToRemove, i)
					})
				})
			})
		}
	})
}

func (ui *ui) planetPresetsWindow(ctx *debugui.Context, game *Game) {
	ctx.Window("Planet Presets", image.Rect(game.screenSize[0]-200, 320, game.screenSize[0], 620), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		for i, planet := range game.simulation.planetHandler.presets {
			if planet == nil {
				continue
			}
			ctx.IDScope("grid "+string(i), func() {
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.SetGridLayout([]int{15, -3, 15}, []int{20})
					ctx.DrawOnlyWidget(func(screen *ebiten.Image) {
						cx := float32(bounds.Min.X) + 7
						cy := float32(bounds.Min.Y) + float32(bounds.Dy())/2
						r := float32(8)
						vector.FillCircle(screen, cx, cy, r, planet.Color, true)
					})
					ctx.IDScope("button "+string(i), func() {
						ctx.Button(fmt.Sprintf("%s", planet.Name)).On(func() {
							game.simulation.planetHandler.planetCreator.planet = newPlanet(
								planet.Name,
								game.simulation.planetHandler.planetCreator.planet.X,
								game.simulation.planetHandler.planetCreator.planet.Y,
								planet.Radius,
								planet.Mass,
								planet.Velocity,
								planet.Color,
								planet.Offset,
							)
							game.simulation.planetHandler.planetCreator.planet.HasNameChanged = true
						})
					})
					ctx.Button("X").On(func() {
						game.simulation.planetHandler.presets = slices.Delete(game.simulation.planetHandler.presets, i, i+1)
					})
				})
			})
		}
	})
}

func (ui *ui) simulationPresetsWindow(ctx *debugui.Context, game *Game) {
	ctx.Window("Simulation Presets", image.Rect(game.screenSize[0]-200, 630, game.screenSize[0], 940), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		ctx.GridCell(func(bounds image.Rectangle) {

			ctx.Text("Name:")
			ctx.TextField(&game.simulation.simulationPresets.newPresetName)
		})
		ctx.Button("Save simulation to presets").On(func() {
			game.simulation.saveSimulationPreset()
		})
		for i, simulationPreset := range game.simulation.simulationPresets.Presets {
			if simulationPreset == nil {
				continue
			}
			ctx.IDScope("grid "+string(i), func() {
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.SetGridLayout([]int{-3, 15}, []int{20})
					ctx.IDScope("button "+string(i), func() {
						ctx.Button(fmt.Sprintf("%s", simulationPreset.Name)).On(func() {
							game.simulation.shouldReset = true
							game.simulation.simulationPresets.shouldLoadSimulation = true
							game.simulation.simulationPresets.presetIndex = i
						})
					})
					ctx.Button("X").On(func() {
						game.simulation.removeSimulationPreset(i)
					})
				})
			})
		}
	})
}

func (ui *ui) Draw(screen *ebiten.Image) {
	ui.debugui.Draw(screen)
}
