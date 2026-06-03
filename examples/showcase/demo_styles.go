package main

import (
	"math"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// styleSeries returns a reusable two-series dataset for style demos.
func styleSeries() []series.XY {
	return []series.XY{
		series.NewXY(series.XYCfg{
			Name: "Product A",
			Points: []series.Point{
				{X: 1, Y: 14}, {X: 2, Y: 22}, {X: 3, Y: 18},
				{X: 4, Y: 30}, {X: 5, Y: 26}, {X: 6, Y: 35},
			},
		}),
		series.NewXY(series.XYCfg{
			Name: "Product B",
			Points: []series.Point{
				{X: 1, Y: 8}, {X: 2, Y: 16}, {X: 3, Y: 24},
				{X: 4, Y: 20}, {X: 5, Y: 32}, {X: 6, Y: 28},
			},
		}),
	}
}

func demoPaletteSwap(w *gui.Window) gui.View {
	makeChart := func(id, title string, palette []gui.Color) gui.View {
		t := theme.Default()
		t.Palette = palette
		return chart.Line(chart.LineCfg{
			BaseCfg: chart.BaseCfg{
				ID:             id,
				Title:          title,
				Sizing:         gui.FillFixed,
				Height:         200,
				Theme:          t,
				LegendPosition: &posBottom,
			},
			ShowMarkers: true,
			Series:      styleSeries(),
		})
	}

	return demoWithCode(w, "style-palette", gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.NoPadding,
		Spacing: gui.SomeF(16),
		Content: []gui.View{
			makeChart("palette-tableau", "Tableau 10 (Default)", theme.Tableau10()),
			makeChart("palette-pastel", "Pastel", theme.Pastel()),
			makeChart("palette-vivid", "Vivid", theme.Vivid()),
		},
	}), `t := theme.Default()
t.Palette = theme.Pastel() // or Vivid(), Tableau10()

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Pastel",
        Theme: t,
    },
    ShowMarkers: true,
    Series:      data,
})`)
}

func demoTickMarks(w *gui.Window) gui.View {
	t := theme.Default()
	t.TickMark = theme.TickMarkStyle{
		Length: 10,
		Color:  gui.Hex(0xE15759),
		Width:  2,
	}

	return demoWithCode(w, "style-tick-marks", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "tick-marks",
			Title:          "Custom Tick Marks",
			Sizing:         gui.FillFixed,
			Height:         350,
			Theme:          t,
			LegendPosition: &posBottom,
		},
		ShowMarkers: true,
		Series:      styleSeries(),
	}), `t := theme.Default()
t.TickMark = theme.TickMarkStyle{
    Length: 10,                 // pixels (default 5)
    Color:  gui.Hex(0xE15759), // red (default AxisColor)
    Width:  2,                 // pixels (default AxisWidth)
}

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Custom Tick Marks",
        Theme: t,
    },
    Series: data,
})`)
}

func demoLegendPositions(w *gui.Window) gui.View {
	positions := []struct {
		id    string
		title string
		pos   theme.LegendPosition
	}{
		{"legend-tl", "TopLeft", theme.LegendTopLeft},
		{"legend-tr", "TopRight", theme.LegendTopRight},
		{"legend-bl", "BottomLeft", theme.LegendBottomLeft},
		{"legend-br", "BottomRight", theme.LegendBottomRight},
		{"legend-top", "Top (Outside)", theme.LegendTop},
		{"legend-bottom", "Bottom (Outside)", theme.LegendBottom},
		{"legend-right", "Right (Outside)", theme.LegendRight},
		{"legend-none", "None (Hidden)", theme.LegendNone},
	}

	charts := make([]gui.View, len(positions))
	for i, p := range positions {
		pos := p.pos
		charts[i] = chart.Line(chart.LineCfg{
			BaseCfg: chart.BaseCfg{
				ID:             p.id,
				Title:          p.title,
				Sizing:         gui.FillFixed,
				Height:         220,
				LegendPosition: &pos,
			},
			Series: styleSeries(),
		})
	}

	return demoWithCode(w, "style-legend-pos", gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.NoPadding,
		Spacing: gui.SomeF(16),
		Content: charts,
	}), `pos := theme.LegendTop // or LegendBottom, LegendRight, LegendNone

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title:          "Right (Outside)",
        LegendPosition: &pos,
    },
    Series: data,
})`)
}

func demoLegendStyling(w *gui.Window) gui.View {
	t := theme.Default()
	t.Legend = theme.LegendStyle{
		Position:   theme.LegendBottom,
		Background: gui.RGBA(40, 40, 80, 180),
		SwatchSize: 16,
		Padding:    10,
		ItemGap:    8,
		RowGap:     4,
	}

	return demoWithCode(w, "style-legend-cfg", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "legend-styled",
			Title:          "Custom Legend Style",
			Sizing:         gui.FillFixed,
			Height:         350,
			Theme:          t,
			LegendPosition: &posBottom,
		},
		ShowMarkers: true,
		Series:      styleSeries(),
	}), `t := theme.Default()
t.Legend = theme.LegendStyle{
    Position:   theme.LegendBottom,
    Background: gui.RGBA(40, 40, 80, 180),
    SwatchSize: 16,  // default 12
    Padding:    10,  // default 6
    ItemGap:    8,   // default 4
    RowGap:     4,   // default 2
}

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Custom Legend Style",
        Theme: t,
    },
    Series: data,
})`)
}

func demoRotatedLabels(w *gui.Window) gui.View {
	data := []series.Category{
		series.NewCategory(series.CategoryCfg{
			Name: "Revenue",
			Values: []series.CategoryValue{
				{Label: "North America", Value: 85},
				{Label: "Western Europe", Value: 72},
				{Label: "East Asia Pacific", Value: 68},
				{Label: "Latin America", Value: 45},
				{Label: "Middle East & Africa", Value: 32},
				{Label: "Southeast Asia", Value: 28},
			},
		}),
	}

	return demoWithCode(w, "style-rotation", chart.Bar(chart.BarCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "rotated-labels",
			Title:          "Revenue by Region",
			Sizing:         gui.FillFixed,
			Height:         380,
			XTickRotation:  -math.Pi / 6,
			LegendPosition: &posBottom,
		},
		Series: data,
	}), `chart.Bar(chart.BarCfg{
    BaseCfg: chart.BaseCfg{
        Title:         "Revenue by Region",
        XTickRotation: -math.Pi / 6, // -30 degrees
    },
    Series: []series.Category{
        series.NewCategory(series.CategoryCfg{
            Name: "Revenue",
            Values: []series.CategoryValue{
                {Label: "North America", Value: 85},
                {Label: "Western Europe", Value: 72},
                {Label: "East Asia Pacific", Value: 68},
                // ...
            },
        }),
    },
})`)
}

func demoCustomPadding(w *gui.Window) gui.View {
	tight := theme.Default()
	tight.PaddingTop = 20
	tight.PaddingRight = 15
	tight.PaddingBottom = 30
	tight.PaddingLeft = 35

	spacious := theme.Default()
	spacious.PaddingTop = 60
	spacious.PaddingRight = 60
	spacious.PaddingBottom = 80
	spacious.PaddingLeft = 80

	return demoWithCode(w, "style-padding", gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.NoPadding,
		Spacing: gui.SomeF(16),
		Content: []gui.View{
			chart.Line(chart.LineCfg{
				BaseCfg: chart.BaseCfg{
					ID:             "padding-tight",
					Title:          "Tight Padding",
					Sizing:         gui.FillFixed,
					Height:         250,
					Theme:          spacious,
					LegendPosition: &posBottom,
				},
				Series: styleSeries(),
			}),
			chart.Line(chart.LineCfg{
				BaseCfg: chart.BaseCfg{
					ID:             "padding-spacious",
					Title:          "Spacious Padding",
					Sizing:         gui.FillFixed,
					Height:         250,
					Theme:          tight,
					LegendPosition: &posBottom,
				},
				Series: styleSeries(),
			}),
		},
	}), `tight := theme.Default()
tight.PaddingTop = 20
tight.PaddingRight = 15
tight.PaddingBottom = 30
tight.PaddingLeft = 35

spacious := theme.Default()
spacious.PaddingTop = 60
spacious.PaddingRight = 60
spacious.PaddingBottom = 80
spacious.PaddingLeft = 80

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Tight Padding",
        Theme: tight,
    },
    Series: data,
})`)
}

func demoKitchenSink(w *gui.Window) gui.View {
	t := theme.Default()
	t.Palette = theme.Vivid()
	t.Background = gui.Hex(0x1A1A2E)
	t.AxisColor = gui.Hex(0x7F8C8D)
	t.AxisWidth = 1.5
	t.GridColor = gui.RGBA(127, 140, 141, 50)
	t.GridWidth = 0.5
	t.TickMark = theme.TickMarkStyle{
		Length: 8,
		Color:  gui.Hex(0xE15759),
		Width:  1.5,
	}
	t.Legend = theme.LegendStyle{
		Position:   theme.LegendBottom,
		Background: gui.RGBA(26, 26, 46, 200),
		SwatchSize: 14,
		Padding:    8,
		ItemGap:    6,
		RowGap:     3,
	}
	t.PaddingTop = 50
	t.PaddingRight = 30
	t.PaddingBottom = 70
	t.PaddingLeft = 70

	yAxis := axis.NewLinear(axis.LinearCfg{
		Title:     "Revenue ($k)",
		AutoRange: true,
	})

	xAxis := axis.NewLinear(axis.LinearCfg{
		Title:     "Month",
		AutoRange: true,
	})

	pos := theme.LegendBottom

	return demoWithCode(w, "style-kitchen", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "kitchen-sink",
			Title:          "All Style Knobs",
			Sizing:         gui.FillFixed,
			Height:         400,
			Theme:          t,
			XTickRotation:  -math.Pi / 8,
			LegendPosition: &pos,
		},
		XAxis:       xAxis,
		YAxis:       yAxis,
		LineWidth:   2.5,
		ShowMarkers: true,
		ShowArea:    true,
		Series:      styleSeries(),
	}), `t := theme.Default()
t.Palette = theme.Vivid()
t.Background = gui.Hex(0x1A1A2E)
t.AxisColor = gui.Hex(0x7F8C8D)
t.AxisWidth = 1.5
t.GridColor = gui.RGBA(127, 140, 141, 50)
t.TickMark = theme.TickMarkStyle{
    Length: 8,
    Color:  gui.Hex(0xE15759),
    Width:  1.5,
}
t.Legend = theme.LegendStyle{
    Position:   theme.LegendBottom,
    Background: gui.RGBA(26, 26, 46, 200),
    SwatchSize: 14,
    Padding:    8,
    ItemGap:    6,
}
t.PaddingTop = 50
t.PaddingBottom = 70
t.PaddingLeft = 70

pos := theme.LegendBottom

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title:          "All Style Knobs",
        Theme:          t,
        XTickRotation:  -math.Pi / 8,
        LegendPosition: &pos,
    },
    XAxis: axis.NewLinear(axis.LinearCfg{
        Title: "Month", AutoRange: true,
    }),
    YAxis: axis.NewLinear(axis.LinearCfg{
        Title: "Revenue ($k)", AutoRange: true,
    }),
    LineWidth:   2.5,
    ShowMarkers: true,
    ShowArea:    true,
    Series:      data,
})`)
}

func demoHighContrast(w *gui.Window) gui.View {
	t := theme.HighContrastTheme()
	barData := []series.Category{
		series.NewCategory(series.CategoryCfg{
			Name: "Q1",
			Values: []series.CategoryValue{
				{Label: "North", Value: 45},
				{Label: "South", Value: 32},
				{Label: "East", Value: 58},
				{Label: "West", Value: 41},
			},
		}),
		series.NewCategory(series.CategoryCfg{
			Name: "Q2",
			Values: []series.CategoryValue{
				{Label: "North", Value: 52},
				{Label: "South", Value: 38},
				{Label: "East", Value: 49},
				{Label: "West", Value: 55},
			},
		}),
	}

	return demoWithCode(w, "style-high-contrast", gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.NoPadding,
		Spacing: gui.SomeF(16),
		Content: []gui.View{
			chart.Line(chart.LineCfg{
				BaseCfg: chart.BaseCfg{
					ID:             "hc-line",
					Title:          "High Contrast Line",
					Sizing:         gui.FillFixed,
					Height:         250,
					Theme:          t,
					LegendPosition: &posBottom,
				},
				ShowMarkers: true,
				LineWidth:   2.5,
				Series:      styleSeries(),
			}),
			chart.Bar(chart.BarCfg{
				BaseCfg: chart.BaseCfg{
					ID:             "hc-bar",
					Title:          "High Contrast Bar",
					Sizing:         gui.FillFixed,
					Height:         250,
					Theme:          t,
					LegendPosition: &posBottom,
				},
				Series: barData,
			}),
		},
	}), `t := theme.HighContrastTheme()

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "High Contrast Line",
        Theme: t,
    },
    ShowMarkers: true,
    LineWidth:   2.5,
    Series:      data,
})`)
}

func demoZoomPan(w *gui.Window) gui.View {
	zoomBase := func(id, title string) chart.BaseCfg {
		return chart.BaseCfg{
			ID:             id,
			Title:          title,
			Sizing:         gui.FillFixed,
			Height:         300,
			LegendPosition: &posBottom,
		}
	}
	zoomInteraction := chart.InteractionCfg{
		EnableZoom:        true,
		EnablePan:         true,
		EnableRangeSelect: true,
	}

	return demoWithCode(w, "style-zoom", gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.NoPadding,
		Spacing: gui.SomeF(16),
		Content: []gui.View{
			chart.Line(chart.LineCfg{
				BaseCfg:        zoomBase("zoom-line", "Line — Zoom & Pan"),
				InteractionCfg: zoomInteraction,
				ShowMarkers:    true,
				Series:         styleSeries(),
			}),
			chart.Area(chart.AreaCfg{
				BaseCfg:        zoomBase("zoom-area", "Area — Zoom & Pan"),
				InteractionCfg: zoomInteraction,
				Series:         styleSeries(),
			}),
			chart.Bar(chart.BarCfg{
				BaseCfg:        zoomBase("zoom-bar", "Bar — Zoom & Pan"),
				InteractionCfg: zoomInteraction,
				Series: []series.Category{
					series.NewCategory(series.CategoryCfg{
						Name: "Q1",
						Values: []series.CategoryValue{
							{Label: "North", Value: 45},
							{Label: "South", Value: 32},
							{Label: "East", Value: 58},
							{Label: "West", Value: 41},
						},
					}),
					series.NewCategory(series.CategoryCfg{
						Name: "Q2",
						Values: []series.CategoryValue{
							{Label: "North", Value: 52},
							{Label: "South", Value: 38},
							{Label: "East", Value: 49},
							{Label: "West", Value: 55},
						},
					}),
				},
			}),
			chart.Scatter(chart.ScatterCfg{
				BaseCfg:        zoomBase("zoom-scatter", "Scatter — Zoom & Pan"),
				InteractionCfg: zoomInteraction,
				Series: []series.XY{
					series.NewXY(series.XYCfg{
						Name: "Subjects",
						Points: []series.Point{
							{X: 155, Y: 52}, {X: 160, Y: 58},
							{X: 162, Y: 55}, {X: 165, Y: 62},
							{X: 167, Y: 60}, {X: 168, Y: 65},
							{X: 170, Y: 68}, {X: 172, Y: 70},
							{X: 173, Y: 66}, {X: 175, Y: 75},
							{X: 176, Y: 72}, {X: 178, Y: 78},
							{X: 180, Y: 80}, {X: 181, Y: 76},
							{X: 183, Y: 85}, {X: 185, Y: 82},
							{X: 187, Y: 88}, {X: 190, Y: 92},
						},
					}),
				},
			}),
			chart.Histogram(chart.HistogramCfg{
				BaseCfg:        zoomBase("zoom-hist", "Histogram — Zoom & Pan"),
				InteractionCfg: zoomInteraction,
				Data:           histData,
			}),
		},
	}), `// All XY chart types support zoom, pan, and range-select.
// Set EnableZoom, EnablePan, EnableRangeSelect in InteractionCfg.

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{...},
    InteractionCfg: chart.InteractionCfg{
        EnableZoom:        true,
        EnablePan:         true,
        EnableRangeSelect: true,
    },
    Series: data,
})

// Double-click to reset zoom`)
}
