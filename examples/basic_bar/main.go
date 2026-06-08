package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
	"github.com/go-gui-org/go-gui/gui/backend"
)

type App struct{}

func main() {
	gui.SetTheme(gui.ThemeDark.WithBorders(true))

	w := gui.NewWindow(gui.WindowCfg{
		State:  &App{},
		Title:  "Bar Chart",
		Width:  800,
		Height: 600,
		OnInit: func(w *gui.Window) {
			w.UpdateView(view)
		},
	})

	backend.Run(w)
}

func view(w *gui.Window) gui.View {
	return chart.Bar(chart.BarCfg{
		BaseCfg: chart.BaseCfg{
			ID:    "bar-demo",
			Title: "Sales by Region",
		},
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Name:  "Q1",
				Color: gui.Hex(0x4E79A7),
				Values: []series.CategoryValue{
					{Label: "North", Value: 45},
					{Label: "South", Value: 32},
					{Label: "East", Value: 58},
					{Label: "West", Value: 41},
				},
			}),
			series.NewCategory(series.CategoryCfg{
				Name:  "Q2",
				Color: gui.Hex(0xF28E2B),
				Values: []series.CategoryValue{
					{Label: "North", Value: 52},
					{Label: "South", Value: 38},
					{Label: "East", Value: 49},
					{Label: "West", Value: 55},
				},
			}),
		},
	})
}
