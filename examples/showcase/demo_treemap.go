package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoTreemapBasic(w *gui.Window) gui.View {
	data := []series.TreeNode{
		{Label: "Documents", Children: []series.TreeNode{
			{Label: "PDFs", Value: 4.2},
			{Label: "Spreadsheets", Value: 2.8},
			{Label: "Presentations", Value: 1.5},
		}},
		{Label: "Media", Children: []series.TreeNode{
			{Label: "Photos", Value: 8.1},
			{Label: "Videos", Value: 15.3},
			{Label: "Music", Value: 3.2},
		}},
		{Label: "Code", Children: []series.TreeNode{
			{Label: "Go", Value: 1.5},
			{Label: "Python", Value: 0.8},
			{Label: "JS", Value: 0.6},
		}},
	}
	return demoWithCode(w, "treemap-basic", chart.Treemap(chart.TreemapCfg{
		BaseCfg: chart.BaseCfg{
			ID:     "treemap-basic",
			Title:  "Disk Usage (GB)",
			Sizing: gui.FillFixed,
			Height: 400,
		},
		Data: data,
	}), `data := []series.TreeNode{
    {Label: "Documents", Children: []series.TreeNode{
        {Label: "PDFs", Value: 4.2},
        {Label: "Spreadsheets", Value: 2.8},
    }},
    {Label: "Media", Children: []series.TreeNode{
        {Label: "Photos", Value: 8.1},
        {Label: "Videos", Value: 15.3},
    }},
}
chart.Treemap(chart.TreemapCfg{
    BaseCfg: chart.BaseCfg{Title: "Disk Usage (GB)"},
    Data:       data,
})`)
}

func demoTreemapStyled(w *gui.Window) gui.View {
	data := []series.TreeNode{
		{Label: "Technology", Children: []series.TreeNode{
			{Label: "Apple", Value: 2900},
			{Label: "Microsoft", Value: 2800},
			{Label: "Nvidia", Value: 2200},
			{Label: "Google", Value: 1900},
		}},
		{Label: "Finance", Children: []series.TreeNode{
			{Label: "JPMorgan", Value: 580},
			{Label: "Visa", Value: 540},
			{Label: "Mastercard", Value: 420},
		}},
		{Label: "Healthcare", Children: []series.TreeNode{
			{Label: "Lilly", Value: 750},
			{Label: "UnitedHealth", Value: 480},
			{Label: "J&J", Value: 380},
		}},
		{Label: "Energy", Children: []series.TreeNode{
			{Label: "Exxon", Value: 460},
			{Label: "Chevron", Value: 280},
		}},
	}
	return demoWithCode(w, "treemap-styled", chart.Treemap(chart.TreemapCfg{
		BaseCfg: chart.BaseCfg{
			ID:     "treemap-styled",
			Title:  "Market Cap by Sector ($B)",
			Sizing: gui.FillFixed,
			Height: 400,
		},
		Data:        data,
		ShowHeaders: true,
		CellGap:     3,
		ValueFormat: "$%.0f",
	}), `data := []series.TreeNode{
    {Label: "Technology", Children: []series.TreeNode{
        {Label: "Apple", Value: 2900},
        {Label: "Microsoft", Value: 2800},
    }},
    {Label: "Finance", Children: []series.TreeNode{
        {Label: "JPMorgan", Value: 580},
    }},
}
chart.Treemap(chart.TreemapCfg{
    BaseCfg:     chart.BaseCfg{Title: "Market Cap ($B)"},
    Data:        data,
    ShowHeaders: true,
    CellGap:     3,
    ValueFormat: "$%.0f",
})`)
}
