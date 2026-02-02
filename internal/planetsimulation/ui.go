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
	debugui  debugui.DebugUI
	title    string
	layouts  []debugui.ContainerLayout
	windows  []image.Rectangle
	hasFocus debugui.InputCapturingState
}

func newUI() *ui {
	ui := &ui{
		debugui:  debugui.DebugUI{},
		title:    "Simulation",
		windows:  make([]image.Rectangle, 0),
		hasFocus: debugui.InputCapturingState(0),
	}

	return ui
}

func (ui *ui) formatFloat(v float64, n int) string {
	return strconv.FormatFloat(v, 'f', n, 64)
}

func (ui *ui) getCoords(game *Game) []int {
	return []int{
		-(game.simulation.screen.offset[0] - game.screenSize[0]/2),
		game.simulation.screen.offset[1] - game.screenSize[1]/2,
	}

}

func (ui *ui) Update(game *Game) error {
	slices.Delete(ui.windows, 0, len(ui.windows))

	selectedPlanet := game.simulation.selectedPlanet

	var err error
	ui.hasFocus, err = ui.debugui.Update(func(ctx *debugui.Context) error {
		ctx.Window(ui.title, image.Rect(0, 0, 250, 280), func(layout debugui.ContainerLayout) {
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

			ctx.Button("Clear all traces").On(func() {
				for _, planet := range game.simulation.planets {
					planet.clearTraces()
				}
			})

			ctx.Button("Reset Simulation").On(func() {
				game.simulation.shouldReset = true
			})
		})
		ctx.Window("Modify Planet", image.Rect(000, 300, 250, 800), func(layout debugui.ContainerLayout) {
			if selectedPlanet != nil {
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.SetGridLayout([]int{-2, -2}, []int{-1, -1})
					ctx.Text(fmt.Sprint("x: ", ui.formatFloat(selectedPlanet.x, 1)))
					ctx.Text(fmt.Sprint("y: ", ui.formatFloat(selectedPlanet.y, 1)))
				})
				radius := selectedPlanet.radius
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.SetGridLayout([]int{-2, -2}, []int{-1})
					ctx.Text("radius: ")
					ctx.NumberFieldF(&radius, 1.0, 1).On(func() {
						if radius > 0 {
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
				ctx.GridCell(func(bounds image.Rectangle) {
					ctx.SetGridLayout([]int{-2, -2}, []int{-1})
					ctx.Text("velocity y: ")
					ctx.NumberFieldF(&selectedPlanet.velocity.y, 1.0, 1)
				})
				ctx.Button("Focus Planet").On(func() {
					selectedPlanet.focus(game.simulation)
				})
				ctx.Header("Color", true, func() {
					r, g, b := selectedPlanet.getColor()
					ctx.GridCell(func(bounds image.Rectangle) {
						ctx.SetGridLayout([]int{-3, -1}, []int{59})
						ctx.GridCell(func(bounds image.Rectangle) {
							ctx.SetGridLayout([]int{-1, -6}, []int{15, 15, 15})
							ctx.Text("r: ")
							ctx.Slider(&r, 0, 255, 1).On(func() {
								selectedPlanet.changeColor(r, -1, -1)
							})
							ctx.Text("g: ")
							ctx.Slider(&g, 0, 255, 1).On(func() {
								selectedPlanet.changeColor(-1, g, -1)
							})
							ctx.Text("b: ")
							ctx.Slider(&b, 0, 255, 1).On(func() {
								selectedPlanet.changeColor(-1, -1, b)
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
					ctx.Button("Clear Traces").On(func() {
						selectedPlanet.clearTraces()
					})
				})

				ctx.Text("")
				ctx.Button("Remove Planet").On(func() {
					for i := 0; i < len(game.simulation.planets); i++ {
						planet := game.simulation.planets[i]
						if planet == selectedPlanet {
							game.simulation.planets = slices.Delete(game.simulation.planets, i, i+1)
							selectedPlanet = nil
						}
					}
				})
			}
		})
		return err
	})
	return err
}

func (ui *ui) Draw(screen *ebiten.Image) {
	ui.debugui.Draw(screen)
}
