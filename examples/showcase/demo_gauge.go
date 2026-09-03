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
		Value:     65,
		ShowValue: true,
		// No needle pivots on the centre here, so the reading sits on
		// it instead of dropping below the hub.
		ValueAnchor: chart.GaugeValueCentre,
		ValueFormat: "%.0f%%",
	}), `chart.Gauge(chart.GaugeCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Completion",
    },
    Value:       65,
    ShowValue:   true,
    ValueAnchor: chart.GaugeValueCentre,
    ValueFormat: "%.0f%%",
})`)
}

// demoGaugeValuePlacement shows the value-placement fields: no needle
// occupies the centre, so the reading sits on it, with the unit on a
// second line under it.
func demoGaugeValuePlacement(w *gui.Window) gui.View {
	return demoWithCode(w, "gauge-value-placement", chart.Gauge(chart.GaugeCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "gauge-value-placement",
			Title:          "Download Speed",
			Sizing:         gui.FillFixed,
			Height:         300,
			LegendPosition: &posBottom,
		},
		Value:       342,
		Max:         500,
		InnerRatio:  0.85,
		ShowValue:   true,
		ValueAnchor: chart.GaugeValueCentre,
		ValueLabel:  "Mbps",
	}), `chart.Gauge(chart.GaugeCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Download Speed",
    },
    Value:       342,
    Max:         500,
    InnerRatio:  0.85,
    ShowValue:   true,
    ValueAnchor: chart.GaugeValueCentre,
    ValueLabel:  "Mbps",
})`)
}
