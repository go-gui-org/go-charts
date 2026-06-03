package main

import (
	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoLogAxis(w *gui.Window) gui.View {
	return demoWithCode(w, "scatter-log", chart.Scatter(chart.ScatterCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "scatter-log",
			Title:          "Earthquake Magnitude vs Energy",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		YAxis: axis.NewLog(axis.LogCfg{
			Title: "Energy (joules)",
			Min:   1e3,
			Max:   1e12,
		}),
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "Events",
				Color: gui.Hex(0xE15759),
				Points: []series.Point{
					{X: 2.0, Y: 1e3},
					{X: 3.0, Y: 2e4},
					{X: 3.5, Y: 1e5},
					{X: 4.0, Y: 6e5},
					{X: 4.5, Y: 2e6},
					{X: 5.0, Y: 2e7},
					{X: 5.5, Y: 1e8},
					{X: 6.0, Y: 6e8},
					{X: 6.5, Y: 2e9},
					{X: 7.0, Y: 2e10},
					{X: 7.5, Y: 1e11},
					{X: 8.0, Y: 6e11},
				},
			}),
		},
	}), `chart.Scatter(chart.ScatterCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Earthquake Magnitude vs Energy",
    },
    YAxis: axis.NewLog(axis.LogCfg{
        Title: "Energy (joules)",
        Min:   1e3,
        Max:   1e12,
    }),
    Series: []series.XY{
        series.NewXY(series.XYCfg{
            Name:  "Events",
            Color: gui.Hex(0xE15759),
            Points: []series.Point{
                {X: 2.0, Y: 1e3},
                {X: 5.0, Y: 2e7},
                {X: 8.0, Y: 6e11},
                // ...
            },
        }),
    },
})`)
}
