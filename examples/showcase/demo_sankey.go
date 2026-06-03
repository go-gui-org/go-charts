package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-gui/gui"
)

func demoSankeyBasic(w *gui.Window) gui.View {
	nodes := []chart.SankeyNode{
		{Label: "Coal"},        // 0
		{Label: "Gas"},         // 1
		{Label: "Solar"},       // 2
		{Label: "Electricity"}, // 3
		{Label: "Heat"},        // 4
		{Label: "Residential"}, // 5
		{Label: "Industrial"},  // 6
		{Label: "Transport"},   // 7
	}
	links := []chart.SankeyLink{
		{Source: 0, Target: 3, Value: 40},
		{Source: 0, Target: 4, Value: 15},
		{Source: 1, Target: 3, Value: 30},
		{Source: 1, Target: 4, Value: 20},
		{Source: 2, Target: 3, Value: 25},
		{Source: 3, Target: 5, Value: 45},
		{Source: 3, Target: 6, Value: 30},
		{Source: 3, Target: 7, Value: 20},
		{Source: 4, Target: 5, Value: 20},
		{Source: 4, Target: 6, Value: 15},
	}
	return demoWithCode(w, "sankey-basic", chart.Sankey(chart.SankeyCfg{
		BaseCfg: chart.BaseCfg{
			ID:     "sankey-basic",
			Title:  "Energy Flow",
			Sizing: gui.FillFixed,
			Height: 400,
		},
		Nodes:      nodes,
		Links:      links,
		ShowLabels: true,
	}), `nodes := []chart.SankeyNode{
    {Label: "Coal"}, {Label: "Gas"}, {Label: "Solar"},
    {Label: "Electricity"}, {Label: "Heat"},
    {Label: "Residential"}, {Label: "Industrial"},
    {Label: "Transport"},
}
links := []chart.SankeyLink{
    {Source: 0, Target: 3, Value: 40},
    {Source: 0, Target: 4, Value: 15},
    {Source: 1, Target: 3, Value: 30},
    {Source: 1, Target: 4, Value: 20},
    {Source: 2, Target: 3, Value: 25},
    {Source: 3, Target: 5, Value: 45},
    {Source: 3, Target: 6, Value: 30},
    {Source: 3, Target: 7, Value: 20},
    {Source: 4, Target: 5, Value: 20},
    {Source: 4, Target: 6, Value: 15},
}
chart.Sankey(chart.SankeyCfg{
    BaseCfg:    chart.BaseCfg{Title: "Energy Flow"},
    Nodes:      nodes,
    Links:      links,
    ShowLabels: true,
})`)
}

func demoSankeyStyled(w *gui.Window) gui.View {
	nodes := []chart.SankeyNode{
		{Label: "Search", Color: gui.Hex(0x4E79A7)},   // 0
		{Label: "Social", Color: gui.Hex(0x59A14F)},   // 1
		{Label: "Direct", Color: gui.Hex(0xF28E2B)},   // 2
		{Label: "Landing", Color: gui.Hex(0xE15759)},  // 3
		{Label: "Blog", Color: gui.Hex(0x76B7B2)},     // 4
		{Label: "Pricing", Color: gui.Hex(0xEDC948)},  // 5
		{Label: "Signup", Color: gui.Hex(0xB07AA1)},   // 6
		{Label: "Purchase", Color: gui.Hex(0xFF9DA7)}, // 7
		{Label: "Bounce", Color: gui.Hex(0x9C755F)},   // 8
	}
	links := []chart.SankeyLink{
		{Source: 0, Target: 3, Value: 50},
		{Source: 0, Target: 4, Value: 20},
		{Source: 1, Target: 3, Value: 30},
		{Source: 1, Target: 4, Value: 15},
		{Source: 2, Target: 5, Value: 25},
		{Source: 3, Target: 6, Value: 40},
		{Source: 3, Target: 8, Value: 40},
		{Source: 4, Target: 5, Value: 20},
		{Source: 4, Target: 8, Value: 15},
		{Source: 5, Target: 6, Value: 30},
		{Source: 5, Target: 8, Value: 15},
		{Source: 6, Target: 7, Value: 45},
		{Source: 6, Target: 8, Value: 25},
	}
	return demoWithCode(w, "sankey-styled", chart.Sankey(chart.SankeyCfg{
		BaseCfg: chart.BaseCfg{
			ID:     "sankey-styled",
			Title:  "Website Traffic",
			Sizing: gui.FillFixed,
			Height: 400,
		},
		Nodes:       nodes,
		Links:       links,
		ShowLabels:  true,
		NodeWidth:   24,
		NodeGap:     12,
		ValueFormat: "%.0f visits",
	}), `nodes := []chart.SankeyNode{
    {Label: "Search", Color: gui.Hex(0x4E79A7)},
    {Label: "Social", Color: gui.Hex(0x59A14F)},
    {Label: "Direct", Color: gui.Hex(0xF28E2B)},
    {Label: "Landing", Color: gui.Hex(0xE15759)},
    {Label: "Blog", Color: gui.Hex(0x76B7B2)},
    {Label: "Pricing", Color: gui.Hex(0xEDC948)},
    {Label: "Signup", Color: gui.Hex(0xB07AA1)},
    {Label: "Purchase", Color: gui.Hex(0xFF9DA7)},
    {Label: "Bounce", Color: gui.Hex(0x9C755F)},
}
links := []chart.SankeyLink{
    {Source: 0, Target: 3, Value: 50},
    // ... (see full demo)
}
chart.Sankey(chart.SankeyCfg{
    BaseCfg:     chart.BaseCfg{Title: "Website Traffic"},
    Nodes:       nodes,
    Links:       links,
    ShowLabels:  true,
    NodeWidth:   24,
    NodeGap:     12,
    ValueFormat: "%.0f visits",
})`)
}
