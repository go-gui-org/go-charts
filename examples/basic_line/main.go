package main

import (
	"fmt"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
	"github.com/go-gui-org/go-gui/gui/backend"
)

type App struct {
	Status string
}

func main() {
	gui.SetTheme(gui.ThemeDark.WithBorders(true))

	w := gui.NewWindow(gui.WindowCfg{
		State:  &App{},
		Title:  "Line Chart",
		Width:  800,
		Height: 600,
		OnInit: func(w *gui.Window) {
			w.UpdateView(view)
		},
	})

	backend.Run(w)
}

func lineChart() gui.View {
	return chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:     "line-demo",
			Title:  "Monthly Revenue",
			Sizing: gui.FillFixed,
			Height: 520,
		},
		XAxis: axis.NewLinear(axis.LinearCfg{
			Title: "Month", AutoRange: true,
		}),
		YAxis: axis.NewLinear(axis.LinearCfg{
			Title: "Revenue ($K)", AutoRange: true,
		}),
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "2025",
				Color: gui.Hex(0x4E79A7),
				Points: []series.Point{
					{X: 1, Y: 12},
					{X: 2, Y: 19},
					{X: 3, Y: 15},
					{X: 4, Y: 28},
					{X: 5, Y: 24},
					{X: 6, Y: 31},
				},
			}),
			series.NewXY(series.XYCfg{
				Name:  "2024",
				Color: gui.Hex(0xF28E2B),
				Points: []series.Point{
					{X: 1, Y: 8},
					{X: 2, Y: 14},
					{X: 3, Y: 11},
					{X: 4, Y: 22},
					{X: 5, Y: 19},
					{X: 6, Y: 26},
				},
			}),
		},
	})
}

func view(w *gui.Window) gui.View {
	app := gui.State[App](w)
	return gui.Column(gui.ContainerCfg{
		Content: []gui.View{
			lineChart(),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Padding: gui.SomeP(8, 8, 8, 8),
				Spacing: gui.SomeF(8),
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						ID:      "export_png",
						Content: []gui.View{gui.Text(gui.TextCfg{Text: "Export PNG"})},
						OnClick: func(ctx gui.EventCtx) {
							v := lineChart()
							err := chart.ExportPNG(v, 800, 600, "line-chart.png")
							a := gui.State[App](ctx.Window)
							if err != nil {
								a.Status = fmt.Sprintf("Export failed: %v", err)
							} else {
								a.Status = "Exported to line-chart.png"
							}
						},
					}),
					gui.Text(gui.TextCfg{Text: app.Status}),
				},
			}),
		},
	})
}
