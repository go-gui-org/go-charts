package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-gui/gui"
)

func demoSparklineBasic(w *gui.Window) gui.View {
	vals := []float64{
		4, 6, 3, 8, 5, 9, 2, 7, 6, 10,
		8, 5, 11, 9, 7, 12, 10, 6, 8, 14,
		11, 9, 13, 10, 8, 15, 12, 10, 11, 16,
	}
	return demoWithCode(w, "sparkline-basic",
		chart.Sparkline(chart.SparklineCfg{
			BaseCfg: chart.BaseCfg{
				ID:     "sparkline-basic",
				Title:  "Basic Sparkline",
				Sizing: gui.FillFixed,
				Height: 40,
			},
			Values:      vals,
			ShowTooltip: true,
		}), `vals := []float64{
    4, 6, 3, 8, 5, 9, 2, 7, 6, 10,
    8, 5, 11, 9, 7, 12, 10, 6, 8, 14,
    11, 9, 13, 10, 8, 15, 12, 10, 11, 16,
}
chart.Sparkline(chart.SparklineCfg{
    BaseCfg: chart.BaseCfg{Height: 40},
    Values:  vals,
    ShowTooltip: true,
})`)
}

func demoSparklineArea(w *gui.Window) gui.View {
	vals := []float64{
		2, 5, 3, 8, 6, 4, 9, 7, 11, 8,
		6, 10, 12, 9, 7, 13, 11, 8, 14, 10,
	}
	return demoWithCode(w, "sparkline-area",
		chart.Sparkline(chart.SparklineCfg{
			BaseCfg: chart.BaseCfg{
				ID:     "sparkline-area",
				Title:  "Area Sparkline",
				Sizing: gui.FillFixed,
				Height: 50,
			},
			Values:         vals,
			Type:           chart.SparklineArea,
			ShowMinMarker:  true,
			ShowMaxMarker:  true,
			ShowLastMarker: true,
			ShowTooltip:    true,
		}), `chart.Sparkline(chart.SparklineCfg{
    BaseCfg: chart.BaseCfg{Height: 50},
    Values:  vals,
    Type:    chart.SparklineArea,
    ShowMinMarker:  true,
    ShowMaxMarker:  true,
    ShowLastMarker: true,
    ShowTooltip:    true,
})`)
}

func demoSparklineBar(w *gui.Window) gui.View {
	vals := []float64{
		3, 7, 5, 10, 4, 8, 6, 9, 2, 11,
		5, 8, 12, 6, 9, 3, 7, 10, 4, 8,
	}
	return demoWithCode(w, "sparkline-bar",
		chart.Sparkline(chart.SparklineCfg{
			BaseCfg: chart.BaseCfg{
				ID:     "sparkline-bar",
				Title:  "Bar Sparkline",
				Sizing: gui.FillFixed,
				Height: 40,
			},
			Values:      vals,
			Type:        chart.SparklineBar,
			Color:       gui.Hex(0x4E79A7),
			ShowTooltip: true,
		}), `chart.Sparkline(chart.SparklineCfg{
    BaseCfg: chart.BaseCfg{Height: 40},
    Values:  vals,
    Type:    chart.SparklineBar,
    Color:   gui.Hex(0x4E79A7),
    ShowTooltip: true,
})`)
}

func demoSparklineBand(w *gui.Window) gui.View {
	vals := []float64{
		-3, 2, -1, 5, 3, -2, 4, -4, 1, 6,
		-1, 3, -5, 2, 7, -3, 4, -2, 5, -1,
		3, -4, 6, 1, -2, 8, -1, 3, -3, 5,
	}
	return demoWithCode(w, "sparkline-band",
		chart.Sparkline(chart.SparklineCfg{
			BaseCfg: chart.BaseCfg{
				ID:     "sparkline-band",
				Title:  "Band Sparkline",
				Sizing: gui.FillFixed,
				Height: 50,
			},
			Values:            vals,
			Type:              chart.SparklineArea,
			ShowReferenceLine: true,
			ReferenceValue:    0,
			BandColoring:      true,
			BandAboveColor:    gui.RGBA(0, 180, 0, 60),
			BandBelowColor:    gui.RGBA(220, 0, 0, 60),
			ShowTooltip:       true,
			ValueFormat:       "%.1f",
		}), `chart.Sparkline(chart.SparklineCfg{
    BaseCfg: chart.BaseCfg{Height: 50},
    Values:  vals,
    Type:    chart.SparklineArea,
    ShowReferenceLine: true,
    ReferenceValue:    0,
    BandColoring:      true,
    BandAboveColor: gui.RGBA(0, 180, 0, 60),
    BandBelowColor: gui.RGBA(220, 0, 0, 60),
    ShowTooltip:    true,
    ValueFormat:    "%.1f",
})`)
}
