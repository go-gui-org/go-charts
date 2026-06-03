package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-gui/gui"
)

func demoGauge(w *gui.Window) gui.View {
	return demoWithCode(w, "gauge-basic", chart.Gauge(chart.GaugeCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "gauge-basic",
			Title:          "CPU Usage",
			Sizing:         gui.FillFixed,
			Height:         300,
			LegendPosition: &posBottom,
		},
		Value:       72,
		ShowValue:   true,
		ShowMinMax:  true,
		ShowPointer: true,
		Zones: []chart.GaugeZone{
			{Label: "Normal", Threshold: 60, Color: gui.Hex(0x59A14F)},
			{Label: "Warning", Threshold: 80, Color: gui.Hex(0xF28E2B)},
			{Label: "Critical", Threshold: 100, Color: gui.Hex(0xE15759)},
		},
	}), `chart.Gauge(chart.GaugeCfg{
    BaseCfg: chart.BaseCfg{
        Title: "CPU Usage",
    },
    Value:       72,
    ShowValue:   true,
    ShowMinMax:  true,
    ShowPointer: true,
    Zones: []chart.GaugeZone{
        {Label: "Normal", Threshold: 60, Color: gui.Hex(0x59A14F)},
        {Label: "Warning", Threshold: 80, Color: gui.Hex(0xF28E2B)},
        {Label: "Critical", Threshold: 100, Color: gui.Hex(0xE15759)},
    },
})`)
}

func demoGaugeSimple(w *gui.Window) gui.View {
	return demoWithCode(w, "gauge-simple", chart.Gauge(chart.GaugeCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "gauge-simple",
			Title:          "Completion",
			Sizing:         gui.FillFixed,
			Height:         300,
			LegendPosition: &posBottom,
		},
		Value:       65,
		ShowValue:   true,
		ValueFormat: "%.0f%%",
	}), `chart.Gauge(chart.GaugeCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Completion",
    },
    Value:       65,
    ShowValue:   true,
    ValueFormat: "%.0f%%",
})`)
}
