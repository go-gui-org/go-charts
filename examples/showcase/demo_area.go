package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoArea(w *gui.Window) gui.View {
	return demoWithCode(w, "area-basic", chart.Area(chart.AreaCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "area-basic",
			Title:          "User Signups",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "Organic",
				Color: gui.Hex(0x4E79A7),
				Points: []series.Point{
					{X: 1, Y: 120}, {X: 2, Y: 180}, {X: 3, Y: 250},
					{X: 4, Y: 310}, {X: 5, Y: 400}, {X: 6, Y: 480},
				},
			}),
			series.NewXY(series.XYCfg{
				Name:  "Paid",
				Color: gui.Hex(0xF28E2B),
				Points: []series.Point{
					{X: 1, Y: 80}, {X: 2, Y: 110}, {X: 3, Y: 160},
					{X: 4, Y: 200}, {X: 5, Y: 260}, {X: 6, Y: 300},
				},
			}),
		},
	}), `chart.Area(chart.AreaCfg{
    BaseCfg: chart.BaseCfg{
        Title: "User Signups",
    },
    Series: []series.XY{
        series.NewXY(series.XYCfg{
            Name:   "Organic",
            Color:  gui.Hex(0x4E79A7),
            Points: []series.Point{
                {X: 1, Y: 120}, {X: 2, Y: 180}, {X: 3, Y: 250},
                {X: 4, Y: 310}, {X: 5, Y: 400}, {X: 6, Y: 480},
            },
        }),
        series.NewXY(series.XYCfg{
            Name:   "Paid",
            Color:  gui.Hex(0xF28E2B),
            Points: []series.Point{ ... },
        }),
    },
})`)
}

func demoAreaStacked(w *gui.Window) gui.View {
	return demoWithCode(w, "area-stacked", chart.Area(chart.AreaCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "area-stacked",
			Title:          "Revenue by Product",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Stacked: true,
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "Widgets",
				Color: gui.Hex(0x4E79A7),
				Points: []series.Point{
					{X: 1, Y: 50}, {X: 2, Y: 55}, {X: 3, Y: 62},
					{X: 4, Y: 58}, {X: 5, Y: 70}, {X: 6, Y: 75},
				},
			}),
			series.NewXY(series.XYCfg{
				Name:  "Gadgets",
				Color: gui.Hex(0xF28E2B),
				Points: []series.Point{
					{X: 1, Y: 30}, {X: 2, Y: 38}, {X: 3, Y: 42},
					{X: 4, Y: 45}, {X: 5, Y: 50}, {X: 6, Y: 55},
				},
			}),
			series.NewXY(series.XYCfg{
				Name:  "Gizmos",
				Color: gui.Hex(0xE15759),
				Points: []series.Point{
					{X: 1, Y: 20}, {X: 2, Y: 22}, {X: 3, Y: 25},
					{X: 4, Y: 28}, {X: 5, Y: 32}, {X: 6, Y: 35},
				},
			}),
		},
	}), `chart.Area(chart.AreaCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Revenue by Product",
    },
    Stacked: true,
    Series: []series.XY{
        series.NewXY(series.XYCfg{
            Name:   "Widgets",
            Color:  gui.Hex(0x4E79A7),
            Points: []series.Point{
                {X: 1, Y: 50}, {X: 2, Y: 55}, ...
            },
        }),
        series.NewXY(series.XYCfg{
            Name:   "Gadgets",
            Color:  gui.Hex(0xF28E2B),
            Points: []series.Point{ ... },
        }),
        series.NewXY(series.XYCfg{
            Name:   "Gizmos",
            Color:  gui.Hex(0xE15759),
            Points: []series.Point{ ... },
        }),
    },
})`)
}
