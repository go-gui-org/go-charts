package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoHeatmapBasic(w *gui.Window) gui.View {
	rows := []string{"Math", "Sci", "Eng", "Hist", "Art"}
	cols := []string{"Math", "Sci", "Eng", "Hist", "Art"}
	values := [][]float64{
		{1.00, 0.85, 0.72, 0.30, 0.15},
		{0.85, 1.00, 0.68, 0.35, 0.20},
		{0.72, 0.68, 1.00, 0.25, 0.18},
		{0.30, 0.35, 0.25, 1.00, 0.60},
		{0.15, 0.20, 0.18, 0.60, 1.00},
	}
	g, _ := series.NewGrid(series.GridCfg{
		Name: "Correlation", Rows: rows, Cols: cols, Values: values,
	})
	return demoWithCode(w, "heatmap-basic", chart.Heatmap(chart.HeatmapCfg{
		BaseCfg: chart.BaseCfg{
			ID:     "heatmap-basic",
			Title:  "Subject Correlation Matrix",
			Sizing: gui.FillFixed,
			Height: 400,
		},
		Data:       g,
		ShowValues: true,
	}), `g, _ := series.NewGrid(series.GridCfg{
    Rows:   []string{"Math", "Sci", "Eng", "Hist", "Art"},
    Cols:   []string{"Math", "Sci", "Eng", "Hist", "Art"},
    Values: [][]float64{
        {1.00, 0.85, 0.72, 0.30, 0.15},
        {0.85, 1.00, 0.68, 0.35, 0.20},
        // ...
    },
})
chart.Heatmap(chart.HeatmapCfg{
    BaseCfg: chart.BaseCfg{Title: "Subject Correlation"},
    Data:       g,
    ShowValues: true,
})`)
}

func demoHeatmapActivity(w *gui.Window) gui.View {
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	g := series.GridFromFunc("Activity", days, months,
		func(r, c int) float64 {
			// Synthetic activity pattern: busier mid-week, summer.
			dayFactor := 1.0 - 0.3*float64((r-3)*(r-3))/9
			monthFactor := 1.0 - 0.4*float64((c-6)*(c-6))/36
			return float64(int((dayFactor*monthFactor*80+10)*10)) / 10
		})
	return demoWithCode(w, "heatmap-activity", chart.Heatmap(chart.HeatmapCfg{
		BaseCfg: chart.BaseCfg{
			ID:     "heatmap-activity",
			Title:  "Weekly Activity by Month",
			Sizing: gui.FillFixed,
			Height: 350,
		},
		Data:      g,
		ColorLow:  gui.Hex(0xEBF5E3),
		ColorHigh: gui.Hex(0x2D6A2E),
	}), `g := series.GridFromFunc("Activity", days, months,
    func(r, c int) float64 {
        // activity score per day/month
        return computeActivity(r, c)
    })
chart.Heatmap(chart.HeatmapCfg{
    BaseCfg:   chart.BaseCfg{Title: "Weekly Activity"},
    Data:      g,
    ColorLow:  gui.Hex(0xEBF5E3),
    ColorHigh: gui.Hex(0x2D6A2E),
})`)
}
