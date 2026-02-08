package planetsimulation

import (
	"fmt"
	"image"
	"slices"

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

func (ui *ui) Update(sim *simulation, planetHandler *planetHandler) error {
	ui.layouts = slices.Delete(ui.layouts, 0, len(ui.layouts))
	var err error
	ui.hasFocus, err = ui.debugui.Update(func(ctx *debugui.Context) error {
		// set global context
		ui.ctx = ctx

		ui.createSystemWindow(ctx, planetHandler, sim)
		ui.createPlanetWindow(ctx, planetHandler)
		ui.modifyPlanetWindow(ctx, planetHandler)
		ui.planetListWindow(ctx, planetHandler, sim.gameSize)
		ui.planetPresetsWindow(ctx, planetHandler, sim.gameSize)
		ui.simulationPresetsWindow(ctx, sim.simulationPresets, planetHandler, sim.gameSize)
		return err
	})
	return err
}

func (ui *ui) createSystemWindow(ctx *debugui.Context, planetHandler *planetHandler, sim *simulation) {
	ctx.Window(ui.title, image.Rect(0, 0, 250, 320), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		ctx.Header("Performance", true, func() {
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("FPS: ")
				ctx.Text(formatFloat(ebiten.ActualFPS(), 2))
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("Current TPS: ")
				ctx.Text(formatFloat(ebiten.ActualTPS(), 2))
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("Target TPS: ")
				ctx.NumberField(&sim.tps, 2.0)
			})

		})
		ctx.Header("Coordinates", true, func() {
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("x:")
				ctx.Text(formatFloat(float64(sim.getCoords(planetHandler)[0]), 1))
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("y:")
				ctx.Text(formatFloat(float64(sim.getCoords(planetHandler)[1]), 1))
			})
		})
		ctx.Header("Constants", true, func() {
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout([]int{-2, -1}, []int{-1})
				ctx.Text("Gravitational Constant:")
				ctx.NumberFieldF(&planetHandler.gravitationalConstant, 0.1, 2)
			})
		})

		ctx.Button(ui.pauseSimulationText).On(func() {
			planetHandler.running = !planetHandler.running
		})
		if planetHandler.running {
			ui.pauseSimulationText = "Pause simulation"
		} else {
			ui.pauseSimulationText = "Resume simulation"
		}

		ctx.Button("Return to origin point").On(func() {
			planetHandler.returnToOrigin()
		})

		ctx.Button("Clear all traces").On(func() {
			for _, planet := range planetHandler.planets {
				planet.clearTraces()
			}
		})

		ctx.Button("Reset Simulation").On(func() {
			sim.shouldReset = true
		})
	})
}

func (ui *ui) createPlanetWindow(ctx *debugui.Context, planetHandler *planetHandler) {
	ctx.Window("Create Planet", image.Rect(0, 325, 250, 645), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("Name: ")
			ctx.TextField(&planetHandler.planetCreator.planet.Name).On(func() {
				planetHandler.planetCreator.planet.HasNameChanged = true
			})
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("x: ")
			ctx.NumberFieldF(&planetHandler.planetCreator.planet.X, 1.0, 1)
		})
		// fake negate
		y := -planetHandler.planetCreator.planet.Y
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("y: ")
			ctx.NumberFieldF(&y, 1.0, 1).On(func() {
				planetHandler.planetCreator.planet.Y = -y
			})
		})
		radius := planetHandler.planetCreator.planet.Radius
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("radius: ")
			ctx.NumberFieldF(&radius, 1.0, 1).On(func() {
				if radius > 0 && radius < 1000 {
					planetHandler.planetCreator.planet.Radius = radius
				}
			})
		})
		mass := planetHandler.planetCreator.planet.Mass
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("mass: ")
			ctx.NumberFieldF(&mass, 1.0, 1).On(func() {
				if mass > 0 {
					planetHandler.planetCreator.planet.Mass = mass
				}
			})
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("velocity x: ")
			ctx.NumberFieldF(&planetHandler.planetCreator.planet.Velocity.X, 1.0, 1)
		})
		// fake negate
		velocityY := -planetHandler.planetCreator.planet.Velocity.Y
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{-2, -2}, []int{-1})
			ctx.Text("velocity y: ")
			ctx.NumberFieldF(&velocityY, 1.0, 1).On(func() {
				planetHandler.planetCreator.planet.Velocity.Y = -velocityY
			})
		})
		ctx.Header("Color", true, func() {
			r, g, b, _ := convertColorToInt(planetHandler.planetCreator.planet.Color)
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
					planetHandler.planetCreator.planet.Color = SetColor(uint8(r), uint8(g), uint8(b), 255)
				})
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.DrawOnlyWidget(func(screen *ebiten.Image) {
						cx := float32(bounds.Min.X) + float32(bounds.Dx())/2
						cy := float32(bounds.Min.Y) + float32(bounds.Dy())/2
						r := float32(bounds.Dx()) / 2
						vector.FillCircle(screen, cx, cy, r, planetHandler.planetCreator.planet.Color, true)
						planetHandler.planetCreator.Update(planetHandler.planetCreator.planet.X, planetHandler.planetCreator.planet.Y, planetHandler)
					})
				})
			})
		})
		ctx.Button("Save to presets").On(func() {
			planetHandler.addPlanetToPresets(*planetHandler.planetCreator.planet)
		})
		ctx.Button("Spawn").On(func() {
			planetHandler.planetCreator.spawnPlanet(planetHandler)
		})
	})
}

func (ui *ui) modifyPlanetWindow(ctx *debugui.Context, planetHandler *planetHandler) {
	if !planetHandler.selectedPlanet.isSelected || planetHandler.selectedPlanet.index >= len(planetHandler.planets) {
		return
	}

	if slices.Contains(planetHandler.planetsToRemove, planetHandler.selectedPlanet.index) {
		return
	}

	selectedPlanet := planetHandler.planets[planetHandler.selectedPlanet.index]
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
			planetHandler.focusPlanet(planetHandler.selectedPlanet.index)
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
			planetHandler.addPlanetToPresets(*selectedPlanet)
		})

		ctx.Button("Remove Planet").On(func() {
			if ui.hasRemovedPlanet {
				ui.hasRemovedPlanet = false
				return
			}
			planetHandler.removeSelectedPlanet()
			ui.hasRemovedPlanet = true
		})
	})
}

func (ui *ui) planetListWindow(ctx *debugui.Context, planetHandler *planetHandler, screenSize []int) {
	ctx.Window("Planets", image.Rect(screenSize[0]-200, 0, screenSize[0], 300), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		for i, planet := range planetHandler.planets {
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
							planetHandler.selectPlanet(i)
							planetHandler.focusPlanet(i)
						})
					})
					ctx.Button("X").On(func() {
						planetHandler.planetsToRemove = append(planetHandler.planetsToRemove, i)
					})
				})
			})
		}
	})
}

func (ui *ui) planetPresetsWindow(ctx *debugui.Context, planetHandler *planetHandler, screenSize []int) {
	ctx.Window("Planet Presets", image.Rect(screenSize[0]-200, 320, screenSize[0], 620), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		for i, planet := range planetHandler.presets {
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
							planetHandler.planetCreator.planet = newPlanet(
								planet.Name,
								planetHandler.planetCreator.planet.X,
								planetHandler.planetCreator.planet.Y,
								planet.Radius,
								planet.Mass,
								planet.Velocity,
								planet.Color,
								planet.Offset,
							)
							planetHandler.planetCreator.planet.HasNameChanged = true
						})
					})
					ctx.Button("X").On(func() {
						planetHandler.presets = slices.Delete(planetHandler.presets, i, i+1)
					})
				})
			})
		}
	})
}

func (ui *ui) simulationPresetsWindow(ctx *debugui.Context, simulationPresets *simulationPresets, planetHandler *planetHandler, screenSize []int) {
	ctx.Window("Simulation Presets", image.Rect(screenSize[0]-200, 630, screenSize[0], 940), func(layout debugui.ContainerLayout) {
		ui.layouts = append(ui.layouts, layout.BodyBounds)
		ctx.GridCell(func(bounds image.Rectangle) {

			ctx.Text("Name:")
			ctx.TextField(&simulationPresets.newPresetName)
		})
		ctx.Button("Save simulation to presets").On(func() {
			simulationPresets.saveSimulationPreset(planetHandler)
		})
		for i, simulationPreset := range simulationPresets.Presets {
			if simulationPreset == nil {
				continue
			}
			ctx.IDScope("grid "+string(i), func() {
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.SetGridLayout([]int{-3, 15}, []int{20})
					ctx.IDScope("button "+string(i), func() {
						ctx.Button(fmt.Sprintf("%s", simulationPreset.Name)).On(func() {
							simulationPresets.shouldLoadSimulation = true
							simulationPresets.presetIndex = i
						})
					})
					ctx.Button("X").On(func() {
						simulationPresets.removeSimulationPreset(i)
					})
				})
			})
		}
	})
}

func (ui *ui) Draw(screen *ebiten.Image) {
	ui.debugui.Draw(screen)
}
