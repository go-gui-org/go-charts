package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoLineBasic(w *gui.Window) gui.View {
	return demoWithCode(w, "line-basic", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "line-basic",
			Title:          "Monthly Revenue",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "2025",
				Color: gui.Hex(0x4E79A7),
				Points: []series.Point{
					{X: 1, Y: 12}, {X: 2, Y: 19}, {X: 3, Y: 15},
					{X: 4, Y: 28}, {X: 5, Y: 24}, {X: 6, Y: 31},
					{X: 7, Y: 27}, {X: 8, Y: 35}, {X: 9, Y: 30},
					{X: 10, Y: 38}, {X: 11, Y: 33}, {X: 12, Y: 42},
				},
			}),
			series.NewXY(series.XYCfg{
				Name:  "2024",
				Color: gui.Hex(0xF28E2B),
				Points: []series.Point{
					{X: 1, Y: 8}, {X: 2, Y: 14}, {X: 3, Y: 11},
					{X: 4, Y: 22}, {X: 5, Y: 19}, {X: 6, Y: 26},
					{X: 7, Y: 23}, {X: 8, Y: 29}, {X: 9, Y: 25},
					{X: 10, Y: 32}, {X: 11, Y: 28}, {X: 12, Y: 36},
				},
			}),
		},
	}), `chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Monthly Revenue",
    },
    Series: []series.XY{
        series.NewXY(series.XYCfg{
            Name:   "2025",
            Color:  gui.Hex(0x4E79A7),
            Points: []series.Point{
                {X: 1, Y: 12}, {X: 2, Y: 19}, {X: 3, Y: 15},
                {X: 4, Y: 28}, {X: 5, Y: 24}, {X: 6, Y: 31},
                // ...
            },
        }),
        series.NewXY(series.XYCfg{
            Name:   "2024",
            Color:  gui.Hex(0xF28E2B),
            Points: []series.Point{ ... },
        }),
    },
})`)
}

func demoLineMarkers(w *gui.Window) gui.View {
	return demoWithCode(w, "line-markers", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "line-markers",
			Title:          "Daily Temperature (C)",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		ShowMarkers: true,
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "This Week",
				Color: gui.Hex(0xE15759),
				Points: []series.Point{
					{X: 1, Y: 18}, {X: 2, Y: 21}, {X: 3, Y: 24},
					{X: 4, Y: 22}, {X: 5, Y: 28}, {X: 6, Y: 26},
					{X: 7, Y: 20},
				},
			}),
		},
	}), `chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Daily Temperature (C)",
    },
    ShowMarkers: true,
    Series: []series.XY{
        series.NewXY(series.XYCfg{
            Name:   "This Week",
            Color:  gui.Hex(0xE15759),
            Points: []series.Point{
                {X: 1, Y: 18}, {X: 2, Y: 21}, {X: 3, Y: 24},
                {X: 4, Y: 22}, {X: 5, Y: 28}, {X: 6, Y: 26},
                {X: 7, Y: 20},
            },
        }),
    },
})`)
}

func demoLineArea(w *gui.Window) gui.View {
	return demoWithCode(w, "line-area", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "line-area",
			Title:          "Website Traffic (thousands)",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		ShowArea: true,
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "Visitors",
				Color: gui.Hex(0x59A14F),
				Points: []series.Point{
					{X: 1, Y: 15}, {X: 2, Y: 22}, {X: 3, Y: 28},
					{X: 4, Y: 35}, {X: 5, Y: 31}, {X: 6, Y: 45},
				},
			}),
		},
	}), `chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Website Traffic (thousands)",
    },
    ShowArea: true,
    Series: []series.XY{
        series.NewXY(series.XYCfg{
            Name:   "Visitors",
            Color:  gui.Hex(0x59A14F),
            Points: []series.Point{
                {X: 1, Y: 15}, {X: 2, Y: 22}, {X: 3, Y: 28},
                {X: 4, Y: 35}, {X: 5, Y: 31}, {X: 6, Y: 45},
            },
        }),
    },
})`)
}

func demoLineMulti(w *gui.Window) gui.View {
	return demoWithCode(w, "line-multi", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "line-multi",
			Title:          "Stock Index Comparison",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		LineWidth: 1.5,
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "S&P 500",
				Color: gui.Hex(0x4E79A7),
				Points: []series.Point{
					{X: 1, Y: 100}, {X: 2, Y: 103}, {X: 3, Y: 101},
					{X: 4, Y: 108}, {X: 5, Y: 112}, {X: 6, Y: 107},
					{X: 7, Y: 115}, {X: 8, Y: 118}, {X: 9, Y: 114},
					{X: 10, Y: 122},
				},
			}),
			series.NewXY(series.XYCfg{
				Name:  "FTSE 100",
				Color: gui.Hex(0xF28E2B),
				Points: []series.Point{
					{X: 1, Y: 100}, {X: 2, Y: 98}, {X: 3, Y: 102},
					{X: 4, Y: 105}, {X: 5, Y: 103}, {X: 6, Y: 109},
					{X: 7, Y: 107}, {X: 8, Y: 112}, {X: 9, Y: 110},
					{X: 10, Y: 116},
				},
			}),
			series.NewXY(series.XYCfg{
				Name:  "Nikkei 225",
				Color: gui.Hex(0xE15759),
				Points: []series.Point{
					{X: 1, Y: 100}, {X: 2, Y: 105}, {X: 3, Y: 99},
					{X: 4, Y: 104}, {X: 5, Y: 110}, {X: 6, Y: 106},
					{X: 7, Y: 113}, {X: 8, Y: 108}, {X: 9, Y: 117},
					{X: 10, Y: 120},
				},
			}),
		},
	}), `chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Stock Index Comparison",
    },
    LineWidth: 1.5,
    Series: []series.XY{
        series.NewXY(series.XYCfg{
            Name:   "S&P 500",
            Color:  gui.Hex(0x4E79A7),
            Points: []series.Point{
                {X: 1, Y: 100}, {X: 2, Y: 103}, ...
            },
        }),
        series.NewXY(series.XYCfg{
            Name:   "FTSE 100",
            Color:  gui.Hex(0xF28E2B),
            Points: []series.Point{ ... },
        }),
        series.NewXY(series.XYCfg{
            Name:   "Nikkei 225",
            Color:  gui.Hex(0xE15759),
            Points: []series.Point{ ... },
        }),
    },
})`)
}
