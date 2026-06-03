package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-gui/gui"
)

func demoFunnelBasic(w *gui.Window) gui.View {
	slices := []chart.PieSlice{
		{Label: "Prospects", Value: 1000},
		{Label: "Qualified", Value: 600},
		{Label: "Proposals", Value: 400},
		{Label: "Negotiations", Value: 200},
		{Label: "Closed", Value: 100},
	}
	return demoWithCode(w, "funnel-basic", chart.Funnel(chart.FunnelCfg{
		BaseCfg: chart.BaseCfg{
			ID:     "funnel-basic",
			Title:  "Sales Funnel",
			Sizing: gui.FillFixed,
			Height: 400,
		},
		Slices:     slices,
		ShowLabels: true,
	}), `slices := []chart.PieSlice{
    {Label: "Prospects", Value: 1000},
    {Label: "Qualified", Value: 600},
    {Label: "Proposals", Value: 400},
    {Label: "Negotiations", Value: 200},
    {Label: "Closed", Value: 100},
}
chart.Funnel(chart.FunnelCfg{
    BaseCfg:    chart.BaseCfg{Title: "Sales Funnel"},
    Slices:     slices,
    ShowLabels: true,
})`)
}

func demoFunnelStyled(w *gui.Window) gui.View {
	slices := []chart.PieSlice{
		{Label: "Applications", Value: 500, Color: gui.Hex(0x4E79A7)},
		{Label: "Phone Screen", Value: 300, Color: gui.Hex(0x59A14F)},
		{Label: "On-Site", Value: 120, Color: gui.Hex(0xF28E2B)},
		{Label: "Offer", Value: 45, Color: gui.Hex(0xE15759)},
		{Label: "Hired", Value: 30, Color: gui.Hex(0x76B7B2)},
	}
	return demoWithCode(w, "funnel-styled", chart.Funnel(chart.FunnelCfg{
		BaseCfg: chart.BaseCfg{
			ID:     "funnel-styled",
			Title:  "Hiring Pipeline",
			Sizing: gui.FillFixed,
			Height: 400,
		},
		Slices:      slices,
		ShowLabels:  true,
		ShowPercent: true,
		SegmentGap:  6,
		ValueFormat: "%.0f",
	}), `slices := []chart.PieSlice{
    {Label: "Applications", Value: 500, Color: gui.Hex(0x4E79A7)},
    {Label: "Phone Screen", Value: 300, Color: gui.Hex(0x59A14F)},
    {Label: "On-Site", Value: 120, Color: gui.Hex(0xF28E2B)},
    {Label: "Offer", Value: 45, Color: gui.Hex(0xE15759)},
    {Label: "Hired", Value: 30, Color: gui.Hex(0x76B7B2)},
}
chart.Funnel(chart.FunnelCfg{
    BaseCfg:     chart.BaseCfg{Title: "Hiring Pipeline"},
    Slices:      slices,
    ShowLabels:  true,
    ShowPercent: true,
    SegmentGap:  6,
})`)
}
