package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-gui/gui"
)

// histData is a fixed dataset approximating a normal distribution
// centered around 50 for use in histogram demos.
var histData = []float64{
	22, 28, 31, 33, 35, 37, 38, 39, 40, 41,
	42, 43, 44, 44, 45, 45, 46, 46, 47, 47,
	48, 48, 49, 49, 50, 50, 50, 51, 51, 52,
	52, 53, 53, 54, 54, 55, 55, 56, 57, 58,
	59, 60, 61, 62, 63, 65, 67, 70, 74, 80,
}

func demoHistogramBasic(w *gui.Window) gui.View {
	return demoWithCode(w, "histogram-basic", chart.Histogram(chart.HistogramCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "histogram-basic",
			Title:          "Score Distribution",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Data: histData,
	}), `chart.Histogram(chart.HistogramCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Score Distribution",
    },
    Data: data, // []float64 of raw values
})`)
}

func demoHistogramDensity(w *gui.Window) gui.View {
	return demoWithCode(w, "histogram-density", chart.Histogram(chart.HistogramCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "histogram-density",
			Title:          "Score Density (20 bins)",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Data:       histData,
		Bins:       20,
		Normalized: true,
	}), `chart.Histogram(chart.HistogramCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Score Density (20 bins)",
    },
    Data:       data,
    Bins:       20,
    Normalized: true,
})`)
}
