package planetsimulation

import (
	"fmt"
	"image"
	"strconv"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
)

type ui struct {
	debugui debugui.DebugUI
	title   string
}

func newUI() *ui {
	ui := &ui{
		debugui: debugui.DebugUI{},
		title:   "Simulation",
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

func (ui *ui) Update() {
	ui.debugui.Update(func(ctx *debugui.Context) error {
		ctx.Window(ui.title, image.Rect(0, 400, 200, 200), func(layout debugui.ContainerLayout) {
			ctx.TreeNode("Performance", func() {
				ctx.Text(fmt.Sprint("FPS: ", ui.formatFloat(ebiten.ActualFPS(), 2)))
				ctx.Text(fmt.Sprint("TPS: ", ui.formatFloat(ebiten.ActualTPS(), 2)))
			})
			ctx.TreeNode("Coordinates", func() {
				ctx.Text(fmt.Sprint("FPS: ", ui.formatFloat(ebiten.ActualFPS(), 2)))
				ctx.Text(fmt.Sprint("TPS: ", ui.formatFloat(ebiten.ActualTPS(), 2)))
			})
		})
		ctx.Window(ui.title, image.Rect(400, 800, 200, 200), func(layout debugui.ContainerLayout) {
			ctx.Text("Some text")
		})
		return nil
	})
}

func (ui *ui) Draw(screen *ebiten.Image) {
	ui.debugui.Draw(screen)
}
