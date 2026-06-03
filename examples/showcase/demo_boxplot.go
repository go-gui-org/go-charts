package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-gui/gui"
)

func demoBoxPlotBasic(w *gui.Window) gui.View {
	return demoWithCode(w, "boxplot-basic", chart.BoxPlot(chart.BoxPlotCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "boxplot-basic",
			Title:          "Quarterly Scores",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Data: []chart.BoxData{
			{Label: "Q1", Values: []float64{
				62, 70, 72, 75, 78, 80, 82, 83, 85, 86,
				87, 88, 89, 90, 91, 93, 95, 110,
			}},
			{Label: "Q2", Values: []float64{
				55, 68, 72, 74, 76, 78, 80, 81, 82, 84,
				85, 86, 88, 90, 92, 94,
			}},
			{Label: "Q3", Values: []float64{
				30, 65, 70, 73, 75, 77, 79, 80, 82, 83,
				84, 86, 87, 89, 91, 93, 96, 115,
			}},
			{Label: "Q4", Values: []float64{
				58, 71, 74, 76, 78, 80, 82, 84, 85, 87,
				88, 89, 91, 92, 94, 97,
			}},
		},
	}), `chart.BoxPlot(chart.BoxPlotCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Quarterly Scores",
    },
    Data: []chart.BoxData{
        {Label: "Q1", Values: []float64{62, 70, ...}},
        {Label: "Q2", Values: []float64{55, 68, ...}},
        ...
    },
})`)
}

func demoBoxPlotStyled(w *gui.Window) gui.View {
	return demoWithCode(w, "boxplot-styled", chart.BoxPlot(chart.BoxPlotCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "boxplot-styled",
			Title:          "Response Times by Service",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Data: []chart.BoxData{
			{Label: "Auth", Color: gui.Hex(0x4e79a7), Values: []float64{
				12, 15, 18, 20, 22, 24, 25, 26, 28, 30,
				32, 34, 36, 38, 42, 95,
			}},
			{Label: "API", Color: gui.Hex(0xf28e2b), Values: []float64{
				8, 10, 12, 14, 15, 16, 17, 18, 19, 20,
				22, 24, 26, 28, 30,
			}},
			{Label: "DB", Color: gui.Hex(0xe15759), Values: []float64{
				5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
				15, 16, 18, 20, 50, 65,
			}},
			{Label: "Cache", Color: gui.Hex(0x76b7b2), Values: []float64{
				1, 1, 2, 2, 3, 3, 3, 4, 4, 5,
				5, 6, 7, 8, 25,
			}},
		},
		BoxWidth: 40,
	}), `chart.BoxPlot(chart.BoxPlotCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Response Times by Service",
    },
    Data: []chart.BoxData{
        {Label: "Auth", Color: gui.Hex(0x4e79a7), Values: ...},
        {Label: "API",  Color: gui.Hex(0xf28e2b), Values: ...},
        ...
    },
    BoxWidth: 40,
})`)
}
