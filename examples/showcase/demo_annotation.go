package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoLineAnnotations(w *gui.Window) gui.View {
	return demoWithCode(w, "line-annotations", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "line-annotations",
			Title:          "Server Response Time (ms)",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
			Annotations: chart.Annotations{
				Lines: []chart.LineAnnotation{
					{
						Axis:            chart.AnnotationY,
						Value:           150,
						Color:           gui.Hex(0xE15759),
						Width:           2,
						Label:           "SLA limit",
						LabelPos:        chart.LabelStart,
						LabelBackground: gui.RGBA(225, 87, 89, 200),
						LabelRadius:     3,
					},
					{
						Axis:    chart.AnnotationX,
						Value:   6,
						Color:   gui.RGBA(100, 100, 100, 180),
						DashLen: 6,
						GapLen:  4,
						Label:   "deploy",
					},
				},
				Regions: []chart.RegionAnnotation{
					{
						Axis:  chart.AnnotationY,
						Min:   0,
						Max:   100,
						Color: gui.RGBA(89, 161, 79, 30),
						Label: "healthy",
					},
					{
						Axis:  chart.AnnotationY,
						Min:   100,
						Max:   150,
						Color: gui.RGBA(255, 190, 0, 30),
						Label: "warning",
					},
				},
				Texts: []chart.TextAnnotation{
					{
						X:               9,
						Y:               170,
						Text:            "spike",
						LabelBackground: gui.RGBA(20, 20, 20, 180),
						LabelRadius:     3,
					},
				},
			},
		},
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "p95",
				Color: gui.Hex(0x4E79A7),
				Points: []series.Point{
					{X: 1, Y: 85}, {X: 2, Y: 92}, {X: 3, Y: 78},
					{X: 4, Y: 95}, {X: 5, Y: 88}, {X: 6, Y: 110},
					{X: 7, Y: 130}, {X: 8, Y: 105}, {X: 9, Y: 165},
					{X: 10, Y: 120}, {X: 11, Y: 98}, {X: 12, Y: 90},
				},
			}),
			series.NewXY(series.XYCfg{
				Name:  "p50",
				Color: gui.Hex(0x59A14F),
				Points: []series.Point{
					{X: 1, Y: 45}, {X: 2, Y: 52}, {X: 3, Y: 40},
					{X: 4, Y: 48}, {X: 5, Y: 42}, {X: 6, Y: 55},
					{X: 7, Y: 68}, {X: 8, Y: 52}, {X: 9, Y: 85},
					{X: 10, Y: 60}, {X: 11, Y: 50}, {X: 12, Y: 46},
				},
			}),
		},
	}), `chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Server Response Time (ms)",
        Annotations: chart.Annotations{
            Lines: []chart.LineAnnotation{{
                Axis:            chart.AnnotationY,
                Value:           150,
                Color:           gui.Hex(0xE15759),
                Label:           "SLA limit",
                LabelPos:        chart.LabelStart,
                LabelBackground: gui.RGBA(225, 87, 89, 200),
                LabelRadius:     3,
            }},
            Regions: []chart.RegionAnnotation{{
                Axis:  chart.AnnotationY,
                Min:   0, Max: 100,
                Color: gui.RGBA(89, 161, 79, 30),
                Label: "healthy",
            }},
            Texts: []chart.TextAnnotation{{
                X: 9, Y: 170, Text: "spike",
            }},
        },
    },
    Series: []series.XY{ ... },
})`)
}
