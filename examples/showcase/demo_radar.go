package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-gui/gui"
)

func demoRadar(w *gui.Window) gui.View {
	return demoWithCode(w, "radar-basic", chart.Radar(chart.RadarCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "radar-basic",
			Title:          "Character Stats",
			Sizing:         gui.FillFixed,
			Height:         400,
			LegendPosition: &posBottom,
		},
		Axes: []chart.RadarAxis{
			{Label: "Attack", Max: 100},
			{Label: "Defense", Max: 100},
			{Label: "Speed", Max: 100},
			{Label: "HP", Max: 100},
			{Label: "Magic", Max: 100},
		},
		Series: []chart.RadarSeries{
			{Name: "Warrior", Values: []float64{90, 80, 40, 70, 20},
				Color: gui.Hex(0x4E79A7)},
			{Name: "Mage", Values: []float64{30, 40, 50, 60, 95},
				Color: gui.Hex(0xF28E2B)},
		},
	}), `chart.Radar(chart.RadarCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Character Stats",
    },
    Axes: []chart.RadarAxis{
        {Label: "Attack", Max: 100},
        {Label: "Defense", Max: 100},
        {Label: "Speed", Max: 100},
        {Label: "HP", Max: 100},
        {Label: "Magic", Max: 100},
    },
    Series: []chart.RadarSeries{
        {Name: "Warrior", Values: []float64{90, 80, 40, 70, 20}},
        {Name: "Mage", Values: []float64{30, 40, 50, 60, 95}},
    },
})`)
}

func demoRadarPolygon(w *gui.Window) gui.View {
	return demoWithCode(w, "radar-polygon", chart.Radar(chart.RadarCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "radar-polygon",
			Title:          "Team Skills",
			Sizing:         gui.FillFixed,
			Height:         400,
			LegendPosition: &posBottom,
		},
		PolygonGrid: true,
		Axes: []chart.RadarAxis{
			{Label: "Frontend", Max: 10},
			{Label: "Backend", Max: 10},
			{Label: "DevOps", Max: 10},
			{Label: "Design", Max: 10},
			{Label: "Testing", Max: 10},
			{Label: "Security", Max: 10},
		},
		Series: []chart.RadarSeries{
			{Name: "Team A", Values: []float64{8, 6, 7, 4, 5, 9},
				Color: gui.Hex(0xE15759)},
			{Name: "Team B", Values: []float64{5, 9, 4, 8, 7, 3},
				Color: gui.Hex(0x76B7B2)},
		},
	}), `chart.Radar(chart.RadarCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Team Skills",
    },
    PolygonGrid: true,
    Axes: []chart.RadarAxis{
        {Label: "Frontend", Max: 10},
        {Label: "Backend", Max: 10},
        {Label: "DevOps", Max: 10},
        {Label: "Design", Max: 10},
        {Label: "Testing", Max: 10},
        {Label: "Security", Max: 10},
    },
    Series: []chart.RadarSeries{
        {Name: "Team A", Values: []float64{8, 6, 7, 4, 5, 9}},
        {Name: "Team B", Values: []float64{5, 9, 4, 8, 7, 3}},
    },
})`)
}
