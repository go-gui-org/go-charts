package main

import "github.com/go-gui-org/go-gui/gui"

func detailPanel(w *gui.Window, entries []DemoEntry) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ShowcaseApp](w)

	if len(entries) == 0 {
		return gui.Column(gui.ContainerCfg{
			ID:         scrollDetail,
			Scrollable: true,
			Sizing:     gui.FillFill,
			SizeBorder: gui.NoBorder,
			Padding:    detailPadding(),
			ScrollbarCfgY: &gui.ScrollbarCfg{
				GapEdge: 4,
			},
			Content: []gui.View{
				gui.Text(gui.TextCfg{
					Text:      "No chart matches filter",
					TextStyle: t.B2,
				}),
			},
		})
	}

	entry := selectedEntry(entries, app.SelectedComponent)

	content := componentDemo(w, entry.ID)

	return gui.Column(gui.ContainerCfg{
		ID:         scrollDetail,
		Scrollable: true,
		Sizing:     gui.FillFill,
		Color:      t.ColorBackground,
		SizeBorder: gui.NoBorder,
		Padding:    detailPadding(),
		Spacing:    gui.Some(t.SpacingLarge),
		ScrollbarCfgY: &gui.ScrollbarCfg{
			GapEdge: 4,
		},
		Content: []gui.View{
			viewTitleBar(entry),
			gui.Text(gui.TextCfg{
				Text:      entry.Summary,
				TextStyle: t.N3,
				Mode:      gui.TextModeWrap,
			}),
			content,
		},
	})
}

func detailPadding() gui.Padding {
	base := gui.CurrentTheme().PaddingLarge
	base.Right += gui.CurrentTheme().ScrollbarStyle.Size + 4
	return base
}

func viewTitleBar(entry DemoEntry) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:     gui.FillFit,
		Spacing:    gui.NoSpacing,
		Padding:    gui.NoPadding,
		SizeBorder: gui.NoBorder,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      entry.Label,
				TextStyle: gui.CurrentTheme().B1,
			}),
			line(),
		},
	})
}

var componentDemos = map[string]func(*gui.Window) gui.View{
	"type_basecfg":          demoBaseCfg,
	"type_series_xy":        demoSeriesXY,
	"type_series_cat":       demoSeriesCategory,
	"type_theme":            demoTheme,
	"type_axis":             demoAxisLinear,
	"type_axis_log":         demoLogAxis,
	"type_axis_time":        demoTimeAxis,
	"line_basic":            demoLineBasic,
	"line_markers":          demoLineMarkers,
	"line_area":             demoLineArea,
	"line_multi":            demoLineMulti,
	"line_annotations":      demoLineAnnotations,
	"bar_basic":             demoBarBasic,
	"bar_single":            demoBarSingle,
	"bar_wide":              demoBarWide,
	"bar_rounded":           demoBarRounded,
	"bar_horizontal":        demoBarHorizontal,
	"bar_stacked":           demoBarStacked,
	"pie_basic":             demoPie,
	"pie_donut":             demoDonut,
	"gauge_basic":           demoGauge,
	"gauge_simple":          demoGaugeSimple,
	"gauge_value_placement": demoGaugeValuePlacement,
	"area_basic":            demoArea,
	"area_stacked":          demoAreaStacked,
	"scatter_basic":         demoScatter,
	"scatter_error_bars":    demoScatterErrorBars,
	"scatter_markers":       demoScatterMarkers,
	"candlestick_basic":     demoCandlestickBasic,
	"histogram_basic":       demoHistogramBasic,
	"histogram_density":     demoHistogramDensity,
	"boxplot_basic":         demoBoxPlotBasic,
	"boxplot_styled":        demoBoxPlotStyled,
	"waterfall_basic":       demoWaterfallBasic,
	"waterfall_styled":      demoWaterfallStyled,
	"combo_basic":           demoComboBasic,
	"combo_multi":           demoComboMulti,
	"radar_basic":           demoRadar,
	"radar_polygon":         demoRadarPolygon,
	"bubble_basic":          demoBubble,
	"bubble_markers":        demoBubbleMarkers,
	"heatmap_basic":         demoHeatmapBasic,
	"heatmap_activity":      demoHeatmapActivity,
	"treemap_basic":         demoTreemapBasic,
	"treemap_styled":        demoTreemapStyled,
	"funnel_basic":          demoFunnelBasic,
	"funnel_styled":         demoFunnelStyled,
	"sankey_basic":          demoSankeyBasic,
	"sankey_styled":         demoSankeyStyled,
	"sparkline_basic":       demoSparklineBasic,
	"sparkline_area":        demoSparklineArea,
	"sparkline_bar":         demoSparklineBar,
	"sparkline_band":        demoSparklineBand,
	"transform_ma":          demoTransformMA,
	"transform_regression":  demoTransformRegression,
	"transform_bands":       demoTransformBands,
	"transform_downsample":  demoTransformDownsample,
	"type_series_xyz":       demoSeriesXYZ,
	"type_data_csv":         demoDataCSV,
	"type_data_json":        demoDataJSON,
	"style_palette":         demoPaletteSwap,
	"style_tick_marks":      demoTickMarks,
	"style_legend_pos":      demoLegendPositions,
	"style_legend_cfg":      demoLegendStyling,
	"style_rotation":        demoRotatedLabels,
	"style_padding":         demoCustomPadding,
	"style_kitchen":         demoKitchenSink,
	"style_high_contrast":   demoHighContrast,
	"style_data_table":      demoDataTable,
	"style_zoom":            demoZoomPan,
	"anim_entry":            demoAnimEntry,
	"anim_transition":       demoAnimTransition,
	"anim_realtime":         demoAnimRealtime,
	"anim_fps":              demoAnimFPS,
}

func componentDemo(w *gui.Window, id string) gui.View {
	if fn, ok := componentDemos[id]; ok {
		return fn(w)
	}
	return demoPlaceholder("Demo: " + id)
}

func demoPlaceholder(text string) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Color:   t.ColorPanel,
		Padding: gui.NewPadding(24, 24, 24, 24),
		Radius:  gui.SomeF(8),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      text,
				TextStyle: t.N3,
				Mode:      gui.TextModeWrap,
			}),
		},
	})
}
