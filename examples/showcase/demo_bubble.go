package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoBubble(w *gui.Window) gui.View {
	return demoWithCode(w, "bubble-basic", chart.Bubble(chart.BubbleCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "bubble-basic",
			Title:          "GDP vs Life Expectancy",
			Sizing:         gui.FillFixed,
			Height:         400,
			LegendPosition: &posBottom,
		},
		Series: []series.XYZ{
			series.NewXYZ(series.XYZCfg{
				Name:  "Americas",
				Color: gui.Hex(0x4E79A7),
				Points: []series.XYZPoint{
					{X: 12, Y: 75, Z: 331},
					{X: 10, Y: 77, Z: 38},
					{X: 15, Y: 75, Z: 213},
					{X: 8, Y: 72, Z: 130},
					{X: 6, Y: 74, Z: 51},
				},
			}),
			series.NewXYZ(series.XYZCfg{
				Name:  "Europe",
				Color: gui.Hex(0xE15759),
				Points: []series.XYZPoint{
					{X: 42, Y: 83, Z: 83},
					{X: 38, Y: 82, Z: 67},
					{X: 44, Y: 83, Z: 47},
					{X: 35, Y: 81, Z: 60},
					{X: 30, Y: 82, Z: 10},
				},
			}),
			series.NewXYZ(series.XYZCfg{
				Name:  "Asia",
				Color: gui.Hex(0x76B7B2),
				Points: []series.XYZPoint{
					{X: 2, Y: 69, Z: 1412},
					{X: 3, Y: 70, Z: 1408},
					{X: 30, Y: 84, Z: 126},
					{X: 32, Y: 83, Z: 52},
					{X: 10, Y: 72, Z: 274},
				},
			}),
		},
	}), `chart.Bubble(chart.BubbleCfg{
    BaseCfg: chart.BaseCfg{
        Title: "GDP vs Life Expectancy",
    },
    Series: []series.XYZ{
        series.NewXYZ(series.XYZCfg{
            Name:  "Americas",
            Color: gui.Hex(0x4E79A7),
            Points: []series.XYZPoint{
                {X: 12, Y: 75, Z: 331},
                {X: 10, Y: 77, Z: 38},
                // ... GDP(k$), life expectancy, population(M)
            },
        }),
        // ... more regions
    },
})`)
}

func demoBubbleMarkers(w *gui.Window) gui.View {
	return demoWithCode(w, "bubble-markers", chart.Bubble(chart.BubbleCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "bubble-markers",
			Title:          "Sensor Readings",
			Sizing:         gui.FillFixed,
			Height:         400,
			LegendPosition: &posBottom,
		},
		Markers: []chart.MarkerShape{
			chart.MarkerCircle,
			chart.MarkerSquare,
			chart.MarkerDiamond,
		},
		Series: []series.XYZ{
			series.NewXYZ(series.XYZCfg{
				Name:  "Temperature",
				Color: gui.Hex(0xE15759),
				Points: []series.XYZPoint{
					{X: 1, Y: 22, Z: 80},
					{X: 3, Y: 25, Z: 120},
					{X: 5, Y: 21, Z: 60},
					{X: 7, Y: 28, Z: 200},
					{X: 9, Y: 24, Z: 100},
				},
			}),
			series.NewXYZ(series.XYZCfg{
				Name:  "Humidity",
				Color: gui.Hex(0x4E79A7),
				Points: []series.XYZPoint{
					{X: 2, Y: 55, Z: 90},
					{X: 4, Y: 60, Z: 150},
					{X: 6, Y: 48, Z: 70},
					{X: 8, Y: 65, Z: 180},
					{X: 10, Y: 52, Z: 110},
				},
			}),
			series.NewXYZ(series.XYZCfg{
				Name:  "Pressure",
				Color: gui.Hex(0x76B7B2),
				Points: []series.XYZPoint{
					{X: 1, Y: 1013, Z: 50},
					{X: 4, Y: 1008, Z: 130},
					{X: 6, Y: 1015, Z: 80},
					{X: 8, Y: 1010, Z: 160},
					{X: 10, Y: 1012, Z: 100},
				},
			}),
		},
	}), `chart.Bubble(chart.BubbleCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Sensor Readings",
    },
    Markers: []chart.MarkerShape{
        chart.MarkerCircle,
        chart.MarkerSquare,
        chart.MarkerDiamond,
    },
    Series: []series.XYZ{
        series.NewXYZ(series.XYZCfg{
            Name:  "Temperature",
            Points: []series.XYZPoint{
                {X: 1, Y: 22, Z: 80},
                // ...
            },
        }),
        // ... each series gets its own marker shape
    },
})`)
}
