package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-gui/gui"
)

func demoPie(w *gui.Window) gui.View {
	return demoWithCode(w, "pie-basic", chart.Pie(chart.PieCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "pie-basic",
			Title:          "Browser Market Share",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		ShowLabels:  true,
		ShowPercent: true,
		Slices: []chart.PieSlice{
			{Label: "Chrome", Value: 65, Color: gui.Hex(0x4E79A7)},
			{Label: "Safari", Value: 18, Color: gui.Hex(0xF28E2B)},
			{Label: "Firefox", Value: 8, Color: gui.Hex(0xE15759)},
			{Label: "Edge", Value: 5, Color: gui.Hex(0x76B7B2)},
			{Label: "Other", Value: 4, Color: gui.Hex(0x59A14F)},
		},
	}), `chart.Pie(chart.PieCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Browser Market Share",
    },
    ShowLabels:  true,
    ShowPercent: true,
    Slices: []chart.PieSlice{
        {Label: "Chrome", Value: 65, Color: gui.Hex(0x4E79A7)},
        {Label: "Safari", Value: 18, Color: gui.Hex(0xF28E2B)},
        {Label: "Firefox", Value: 8, Color: gui.Hex(0xE15759)},
        {Label: "Edge", Value: 5, Color: gui.Hex(0x76B7B2)},
        {Label: "Other", Value: 4, Color: gui.Hex(0x59A14F)},
    },
})`)
}

func demoDonut(w *gui.Window) gui.View {
	return demoWithCode(w, "pie-donut", chart.Pie(chart.PieCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "pie-donut",
			Title:          "Budget Allocation",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		InnerRadius: 60,
		ShowLabels:  true,
		Slices: []chart.PieSlice{
			{Label: "R&D", Value: 35, Color: gui.Hex(0x4E79A7)},
			{Label: "Marketing", Value: 25, Color: gui.Hex(0xF28E2B)},
			{Label: "Operations", Value: 22, Color: gui.Hex(0xE15759)},
			{Label: "Admin", Value: 18, Color: gui.Hex(0x76B7B2)},
		},
	}), `chart.Pie(chart.PieCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Budget Allocation",
    },
    InnerRadius: 60,
    ShowLabels:  true,
    Slices: []chart.PieSlice{
        {Label: "R&D", Value: 35, Color: gui.Hex(0x4E79A7)},
        {Label: "Marketing", Value: 25, Color: gui.Hex(0xF28E2B)},
        {Label: "Operations", Value: 22, Color: gui.Hex(0xE15759)},
        {Label: "Admin", Value: 18, Color: gui.Hex(0x76B7B2)},
    },
})`)
}
