package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-gui/gui"
)

func demoWaterfallBasic(w *gui.Window) gui.View {
	return demoWithCode(w, "waterfall-basic", chart.Waterfall(chart.WaterfallCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "waterfall-basic",
			Title:          "Profit & Loss",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
			XTickRotation:  0.5,
		},
		Values: []chart.WaterfallValue{
			{Label: "Revenue", Value: 5000},
			{Label: "COGS", Value: -2000},
			{Label: "Gross Profit", IsTotal: true},
			{Label: "OpEx", Value: -800},
			{Label: "Marketing", Value: -400},
			{Label: "EBIT", IsTotal: true},
			{Label: "Tax", Value: -360},
			{Label: "Net Income", IsTotal: true},
		},
	}), `chart.Waterfall(chart.WaterfallCfg{
    BaseCfg: chart.BaseCfg{
        Title:         "Profit & Loss",
        XTickRotation: 0.5,
    },
    Values: []chart.WaterfallValue{
        {Label: "Revenue", Value: 5000},
        {Label: "COGS", Value: -2000},
        {Label: "Gross Profit", IsTotal: true},
        {Label: "OpEx", Value: -800},
        {Label: "Marketing", Value: -400},
        {Label: "EBIT", IsTotal: true},
        {Label: "Tax", Value: -360},
        {Label: "Net Income", IsTotal: true},
    },
})`)
}

func demoWaterfallStyled(w *gui.Window) gui.View {
	return demoWithCode(w, "waterfall-styled", chart.Waterfall(chart.WaterfallCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "waterfall-styled",
			Title:          "Monthly Cash Flow",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
			XTickRotation:  0.5,
		},
		Values: []chart.WaterfallValue{
			{Label: "Opening", Value: 10000, IsTotal: true},
			{Label: "Sales", Value: 3200},
			{Label: "Services", Value: 1800},
			{Label: "Rent", Value: -2500},
			{Label: "Salaries", Value: -4000},
			{Label: "Utilities", Value: -600},
			{Label: "Marketing", Value: -800},
			{Label: "Closing", IsTotal: true},
		},
		UpColor:    gui.Hex(0x2ca02c),
		DownColor:  gui.Hex(0xd62728),
		TotalColor: gui.Hex(0x1f77b4),
		Radius:     4,
	}), `chart.Waterfall(chart.WaterfallCfg{
    BaseCfg: chart.BaseCfg{
        Title:         "Monthly Cash Flow",
        XTickRotation: 0.5,
    },
    Values: []chart.WaterfallValue{
        {Label: "Opening", Value: 10000, IsTotal: true},
        {Label: "Sales", Value: 3200},
        {Label: "Services", Value: 1800},
        {Label: "Rent", Value: -2500},
        ...
        {Label: "Closing", IsTotal: true},
    },
    UpColor:    gui.Hex(0x2ca02c),
    DownColor:  gui.Hex(0xd62728),
    TotalColor: gui.Hex(0x1f77b4),
    Radius:     4,
})`)
}
