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

// demoGaugeGradient shows GradientZones: the same three zones as the
// basic dial, blended into one continuous sweep instead of stepping
// at each threshold. The legend and tooltip still use the discrete
// thresholds, so the reading is no less precise.
func demoGaugeGradient(w *gui.Window) gui.View {
	return demoWithCode(w, "gauge-gradient", chart.Gauge(chart.GaugeCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "gauge-gradient",
			Title:          "Download Speed",
			Sizing:         gui.FillFixed,
			Height:         300,
			LegendPosition: &posBottom,
		},
		Value:         342,
		Max:           500,
		ShowValue:     true,
		ShowPointer:   true,
		ValueLabel:    "Mbps",
		GradientZones: true,
		Zones: []chart.GaugeZone{
			{Label: "Slow", Threshold: 150, Color: gui.Hex(0xE15759)},
			{Label: "Fair", Threshold: 300, Color: gui.Hex(0xF28E2B)},
			{Label: "Fast", Threshold: 500, Color: gui.Hex(0x59A14F)},
		},
	}), `chart.Gauge(chart.GaugeCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Download Speed",
    },
    Value:         342,
    Max:           500,
    ShowValue:     true,
    ShowPointer:   true,
    ValueLabel:    "Mbps",
    GradientZones: true,
    Zones: []chart.GaugeZone{
        {Label: "Slow", Threshold: 150, Color: gui.Hex(0xE15759)},
        {Label: "Fair", Threshold: 300, Color: gui.Hex(0xF28E2B)},
        {Label: "Fast", Threshold: 500, Color: gui.Hex(0x59A14F)},
    },
})`)
}

// demoGaugeArcRamp shows ArcGradient: an explicit ramp for a dial
// with no zones, so the sweep is described once and directly.
func demoGaugeArcRamp(w *gui.Window) gui.View {
	return demoWithCode(w, "gauge-arc-ramp", chart.Gauge(chart.GaugeCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "gauge-arc-ramp",
			Title:          "Battery Health",
			Sizing:         gui.FillFixed,
			Height:         300,
			LegendPosition: &posBottom,
		},
		Value:       78,
		ShowValue:   true,
		ValueAnchor: chart.GaugeValueCentre,
		ValueFormat: "%.0f%%",
		InnerRatio:  0.8,
		ArcGradient: []gui.GradientStop{
			{Color: gui.Hex(0xE15759), Pos: 0},
			{Color: gui.Hex(0xF28E2B), Pos: 0.5},
			{Color: gui.Hex(0x59A14F), Pos: 1},
		},
	}), `chart.Gauge(chart.GaugeCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Battery Health",
    },
    Value:       78,
    ShowValue:   true,
    ValueAnchor: chart.GaugeValueCentre,
    ValueFormat: "%.0f%%",
    InnerRatio:  0.8,
    ArcGradient: []gui.GradientStop{
        {Color: gui.Hex(0xE15759), Pos: 0},
        {Color: gui.Hex(0xF28E2B), Pos: 0.5},
        {Color: gui.Hex(0x59A14F), Pos: 1},
    },
})`)
}
