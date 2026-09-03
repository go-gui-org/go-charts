package main

import (
	"sort"
	"strings"
)

const (
	themeDefault      = "default"
	themePastel       = "pastel"
	themeVivid        = "vivid"
	themeHighContrast = "high-contrast"
)

const (
	groupAll         = "all"
	groupTypes       = "types"
	groupLine        = "line"
	groupBar         = "bar"
	groupPie         = "pie"
	groupGauge       = "gauge"
	groupArea        = "area"
	groupScatter     = "scatter"
	groupCandlestick = "candlestick"
	groupStyles      = "styles"
	groupHistogram   = "histogram"
	groupBoxPlot     = "boxplot"
	groupWaterfall   = "waterfall"
	groupCombo       = "combo"
	groupRadar       = "radar"
	groupBubble      = "bubble"
	groupHeatmap     = "heatmap"
	groupTreemap     = "treemap"
	groupFunnel      = "funnel"
	groupSankey      = "sankey"
	groupSparkline   = "sparkline"
	groupTransform   = "transform"
	groupAnimation   = "animation"
)

// ShowcaseApp holds all state for the charts showcase.
type ShowcaseApp struct {
	NavQuery          string
	SelectedGroup     string
	SelectedComponent string
	SelectedTheme     string
}

func newShowcaseApp() *ShowcaseApp {
	return &ShowcaseApp{
		SelectedGroup:     groupAll,
		SelectedComponent: "line_basic",
		SelectedTheme:     themeDefault,
	}
}

// DemoEntry describes one demo in the catalog.
type DemoEntry struct {
	ID      string
	Label   string
	Group   string
	Summary string
	Tags    []string

	idLower    string
	labelLower string
}

// DemoGroup describes a catalog group in the left pane.
type DemoGroup struct {
	Key   string
	Label string
}

var demoGroups = []DemoGroup{
	{groupAll, "All"},
	{groupTypes, "Types"},
	{groupLine, "Line"},
	{groupBar, "Bar"},
	{groupPie, "Pie"},
	{groupGauge, "Gauge"},
	{groupArea, "Area"},
	{groupScatter, "Scatter"},
	{groupCandlestick, "Candlestick"},
	{groupHistogram, "Histogram"},
	{groupBoxPlot, "Box Plot"},
	{groupWaterfall, "Waterfall"},
	{groupCombo, "Combo"},
	{groupRadar, "Radar"},
	{groupBubble, "Bubble"},
	{groupHeatmap, "Heatmap"},
	{groupTreemap, "Treemap"},
	{groupFunnel, "Funnel"},
	{groupSankey, "Sankey"},
	{groupSparkline, "Sparkline"},
	{groupTransform, "Transform"},
	{groupAnimation, "Animation"},
	{groupStyles, "Styles"},
}

var demoEntries = []DemoEntry{
	// Types
	{ID: "type_basecfg", Label: "BaseCfg", Group: groupTypes, Summary: "Common configuration fields embedded by all chart types.", Tags: []string{"base", "config", "id", "title", "sizing", "theme", "version"}},
	{ID: "type_series_xy", Label: "series.XY", Group: groupTypes, Summary: "XY data series for Line, Area, and Scatter charts.", Tags: []string{"series", "xy", "point", "data", "slices"}},
	{ID: "type_series_cat", Label: "series.Category", Group: groupTypes, Summary: "Categorical data series for Bar charts.", Tags: []string{"series", "category", "label", "value", "map"}},
	{ID: "type_theme", Label: "theme.Theme", Group: groupTypes, Summary: "Visual style: colors, fonts, padding, palette.", Tags: []string{"theme", "palette", "color", "style", "tableau", "pastel", "vivid"}},
	{ID: "type_series_xyz", Label: "series.XYZ", Group: groupTypes, Summary: "XYZ data series for Bubble charts.", Tags: []string{"series", "xyz", "point", "data", "bubble", "size"}},
	{ID: "type_axis", Label: "axis.Linear", Group: groupTypes, Summary: "Linear numeric axis with auto-tick generation.", Tags: []string{"axis", "linear", "tick", "range", "format", "auto"}},
	{ID: "type_axis_log", Label: "axis.Log", Group: groupTypes, Summary: "Logarithmic axis with major/minor tick generation.", Tags: []string{"axis", "log", "logarithmic", "tick", "base", "decade"}},
	{ID: "type_axis_time", Label: "axis.Time", Group: groupTypes, Summary: "Time axis with adaptive tick alignment.", Tags: []string{"axis", "time", "date", "tick", "format", "auto"}},
	{ID: "type_data_csv", Label: "Data: CSV", Group: groupTypes, Summary: "Parse CSV data into typed series.", Tags: []string{"csv", "data", "parse", "import", "file", "reader"}},
	{ID: "type_data_json", Label: "Data: JSON", Group: groupTypes, Summary: "Parse JSON data into typed series.", Tags: []string{"json", "data", "parse", "import", "file", "reader"}},

	// Line
	{ID: "line_basic", Label: "Basic Line", Group: groupLine, Summary: "Monthly revenue comparison across two years.", Tags: []string{"line", "multi", "revenue"}},
	{ID: "line_markers", Label: "Line with Markers", Group: groupLine, Summary: "Temperature readings with visible data points.", Tags: []string{"line", "markers", "temperature"}},
	{ID: "line_area", Label: "Line with Area Fill", Group: groupLine, Summary: "Website traffic with shaded area under the curve.", Tags: []string{"line", "area", "fill", "traffic"}},
	{ID: "line_multi", Label: "Multi-Series Line", Group: groupLine, Summary: "Three series with custom line width.", Tags: []string{"line", "multi", "series"}},
	{ID: "line_annotations", Label: "Line with Annotations", Group: groupLine, Summary: "Reference lines, shaded regions, and text labels.", Tags: []string{"line", "annotation", "region", "reference", "label"}},

	// Bar
	{ID: "bar_basic", Label: "Basic Bar", Group: groupBar, Summary: "Regional sales comparison across quarters.", Tags: []string{"bar", "grouped", "sales"}},
	{ID: "bar_single", Label: "Single Series Bar", Group: groupBar, Summary: "Monthly rainfall totals.", Tags: []string{"bar", "single", "rainfall"}},
	{ID: "bar_wide", Label: "Fixed Width Bars", Group: groupBar, Summary: "Department headcount with fixed bar width.", Tags: []string{"bar", "fixed", "width", "headcount"}},
	{ID: "bar_rounded", Label: "Rounded Bars", Group: groupBar, Summary: "Product category revenue with rounded corners.", Tags: []string{"bar", "rounded", "radius"}},
	{ID: "bar_horizontal", Label: "Horizontal Bar", Group: groupBar, Summary: "Survey results with bars drawn left-to-right.", Tags: []string{"bar", "horizontal", "survey"}},
	{ID: "bar_stacked", Label: "Stacked Bar", Group: groupBar, Summary: "Traffic by channel stacked per month.", Tags: []string{"bar", "stacked", "channel", "traffic"}},

	// Pie
	{ID: "pie_basic", Label: "Basic Pie", Group: groupPie, Summary: "Browser market share distribution.", Tags: []string{"pie", "share", "percent"}},
	{ID: "pie_donut", Label: "Donut Chart", Group: groupPie, Summary: "Budget allocation with inner radius.", Tags: []string{"pie", "donut", "budget"}},

	// Gauge
	{ID: "gauge_basic", Label: "Gauge with Zones", Group: groupGauge, Summary: "CPU usage gauge with colored warning zones.", Tags: []string{"gauge", "arc", "zone", "value"}},
	{ID: "gauge_simple", Label: "Simple Gauge", Group: groupGauge, Summary: "Completion percentage with no zones.", Tags: []string{"gauge", "simple", "percent"}},
	{ID: "gauge_value_placement", Label: "Gauge Value Placement", Group: groupGauge, Summary: "Value centred in a thin ring with a unit line under it.", Tags: []string{"gauge", "value", "label", "anchor"}},

	// Area
	{ID: "area_basic", Label: "Basic Area", Group: groupArea, Summary: "User signups over time.", Tags: []string{"area", "signups", "time"}},
	{ID: "area_stacked", Label: "Stacked Area", Group: groupArea, Summary: "Revenue breakdown by product line.", Tags: []string{"area", "stacked", "revenue"}},

	// Scatter
	{ID: "scatter_basic", Label: "Basic Scatter", Group: groupScatter, Summary: "Height versus weight correlation.", Tags: []string{"scatter", "correlation", "points"}},
	{ID: "scatter_error_bars", Label: "Error Bars", Group: groupScatter, Summary: "Measurement uncertainty with asymmetric error bars.", Tags: []string{"scatter", "error", "bars", "uncertainty"}},
	{ID: "scatter_markers", Label: "Marker Shapes", Group: groupScatter, Summary: "Wind speed versus temperature with different marker shapes.", Tags: []string{"scatter", "markers", "shapes"}},

	// Candlestick
	{ID: "candlestick_basic", Label: "Basic Candlestick", Group: groupCandlestick, Summary: "AAPL price action with up/down candle colors.", Tags: []string{"candlestick", "ohlc", "price", "stock", "finance"}},

	// Histogram
	{ID: "histogram_basic", Label: "Basic Histogram", Group: groupHistogram, Summary: "Score distribution with auto bin count.", Tags: []string{"histogram", "bin", "frequency", "distribution", "count"}},
	{ID: "histogram_density", Label: "Density Histogram", Group: groupHistogram, Summary: "Normalized frequency density with 20 bins.", Tags: []string{"histogram", "density", "normalized", "bins"}},

	// Box Plot
	{ID: "boxplot_basic", Label: "Basic Box Plot", Group: groupBoxPlot, Summary: "Quarterly scores with quartiles, whiskers, and outliers.", Tags: []string{"boxplot", "quartile", "whisker", "outlier", "median", "statistics"}},
	{ID: "boxplot_styled", Label: "Styled Box Plot", Group: groupBoxPlot, Summary: "Response times with custom colors and fixed box width.", Tags: []string{"boxplot", "color", "style", "custom", "width"}},

	// Waterfall
	{ID: "waterfall_basic", Label: "Basic Waterfall", Group: groupWaterfall, Summary: "Profit & Loss breakdown from revenue to net income.", Tags: []string{"waterfall", "pnl", "profit", "loss", "running", "total"}},
	{ID: "waterfall_styled", Label: "Styled Waterfall", Group: groupWaterfall, Summary: "Monthly cash flow with custom colors and rounded bars.", Tags: []string{"waterfall", "cashflow", "custom", "color", "rounded"}},

	// Radar
	{ID: "radar_basic", Label: "Basic Radar", Group: groupRadar, Summary: "Character stat comparison with circular grid.", Tags: []string{"radar", "spider", "polygon", "stats"}},
	{ID: "radar_polygon", Label: "Polygon Grid", Group: groupRadar, Summary: "Radar chart with polygon grid lines.", Tags: []string{"radar", "spider", "polygon", "grid"}},

	// Bubble
	{ID: "bubble_basic", Label: "Basic Bubble", Group: groupBubble, Summary: "GDP vs life expectancy with population as bubble size.", Tags: []string{"bubble", "scatter", "size", "xyz", "population"}},
	{ID: "bubble_markers", Label: "Marker Shapes", Group: groupBubble, Summary: "Sensor readings with different marker shapes per series.", Tags: []string{"bubble", "markers", "shapes", "per-series"}},

	// Heatmap
	{ID: "heatmap_basic", Label: "Basic Heatmap", Group: groupHeatmap, Summary: "Correlation matrix with values displayed in cells.", Tags: []string{"heatmap", "grid", "correlation", "matrix", "color"}},
	{ID: "heatmap_activity", Label: "Activity Heatmap", Group: groupHeatmap, Summary: "Weekly activity by month with green gradient.", Tags: []string{"heatmap", "grid", "activity", "calendar", "gradient"}},

	// Treemap
	{ID: "treemap_basic", Label: "Basic Treemap", Group: groupTreemap, Summary: "Disk usage by category with nested rectangles.", Tags: []string{"treemap", "nested", "hierarchy", "squarified", "rectangles"}},
	{ID: "treemap_styled", Label: "Styled Treemap", Group: groupTreemap, Summary: "Market cap by sector with group headers and custom gap.", Tags: []string{"treemap", "headers", "market", "sector", "styled"}},

	// Funnel
	{ID: "funnel_basic", Label: "Basic Funnel", Group: groupFunnel, Summary: "Sales funnel from prospects to closed deals.", Tags: []string{"funnel", "stages", "conversion", "sales"}},
	{ID: "funnel_styled", Label: "Styled Funnel", Group: groupFunnel, Summary: "Hiring pipeline with custom colors and percentages.", Tags: []string{"funnel", "styled", "percent", "hiring", "pipeline"}},

	// Sankey
	{ID: "sankey_basic", Label: "Basic Sankey", Group: groupSankey, Summary: "Energy flow from sources through intermediates to end uses.", Tags: []string{"sankey", "flow", "energy", "ribbon", "node", "link"}},
	{ID: "sankey_styled", Label: "Styled Sankey", Group: groupSankey, Summary: "Website traffic flow with custom colors and node sizing.", Tags: []string{"sankey", "flow", "traffic", "styled", "custom", "color"}},

	// Sparkline
	{ID: "sparkline_basic", Label: "Basic Sparkline", Group: groupSparkline, Summary: "Compact inline line sparkline showing trend.", Tags: []string{"sparkline", "inline", "trend", "mini"}},
	{ID: "sparkline_area", Label: "Area Sparkline", Group: groupSparkline, Summary: "Filled area sparkline with min/max markers.", Tags: []string{"sparkline", "area", "fill", "markers"}},
	{ID: "sparkline_bar", Label: "Bar Sparkline", Group: groupSparkline, Summary: "Bar-style sparkline showing distribution.", Tags: []string{"sparkline", "bar", "distribution"}},
	{ID: "sparkline_band", Label: "Band Sparkline", Group: groupSparkline, Summary: "Sparkline with positive/negative color bands and reference line.", Tags: []string{"sparkline", "band", "reference", "positive", "negative"}},

	// Transform
	{ID: "transform_ma", Label: "Moving Averages", Group: groupTransform, Summary: "Noisy signal smoothed with SMA, EMA, and WMA overlays.", Tags: []string{"transform", "sma", "ema", "wma", "moving", "average", "smooth"}},
	{ID: "transform_regression", Label: "Regression", Group: groupTransform, Summary: "Linear trend and polynomial curve fit on scattered data.", Tags: []string{"transform", "regression", "linear", "polynomial", "trend", "fit"}},
	{ID: "transform_bands", Label: "Bollinger Bands", Group: groupTransform, Summary: "Price series with upper/middle/lower Bollinger bands.", Tags: []string{"transform", "bollinger", "bands", "envelope", "stddev"}},
	{ID: "transform_downsample", Label: "LTTB Downsampling", Group: groupTransform, Summary: "1000-point series downsampled to 50 points via LTTB.", Tags: []string{"transform", "lttb", "downsample", "large", "dataset"}},

	// Animation
	{ID: "anim_entry", Label: "Entry Animation", Group: groupAnimation, Summary: "Line, bar, and pie charts with animated entry.", Tags: []string{"animation", "entry", "draw-in", "grow", "sweep"}},
	{ID: "anim_transition", Label: "Data Transition", Group: groupAnimation, Summary: "Smooth interpolation when data changes.", Tags: []string{"animation", "transition", "interpolate", "morph"}},
	{ID: "anim_realtime", Label: "Real-time Stream", Group: groupAnimation, Summary: "Streaming data with rolling window and auto-scroll.", Tags: []string{"animation", "realtime", "stream", "live", "append"}},
	{ID: "anim_fps", Label: "FPS Adaptive", Group: groupAnimation, Summary: "Heavy chart with FPS-aware detail reduction during animation.", Tags: []string{"animation", "fps", "adaptive", "performance", "detail"}},

	// Combo
	{ID: "combo_basic", Label: "Basic Combo", Group: groupCombo, Summary: "Monthly revenue as bars with trend line overlay.", Tags: []string{"combo", "bar", "line", "mixed", "trend"}},
	{ID: "combo_multi", Label: "Multi-Series Combo", Group: groupCombo, Summary: "Multiple bar and line series on shared axes.", Tags: []string{"combo", "bar", "line", "multi", "series"}},

	// Styles
	{ID: "style_palette", Label: "Palette Swap", Group: groupStyles, Summary: "Same data rendered with Tableau 10, Pastel, and Vivid palettes.", Tags: []string{"style", "palette", "theme", "color", "tableau", "pastel", "vivid"}},
	{ID: "style_tick_marks", Label: "Tick Marks", Group: groupStyles, Summary: "Custom tick mark length, color, and width.", Tags: []string{"style", "tick", "marks", "axis"}},
	{ID: "style_legend_pos", Label: "Legend Positions", Group: groupStyles, Summary: "Legend placed in corners, outside top/bottom/right, or hidden.", Tags: []string{"style", "legend", "position", "corner", "top", "bottom", "right", "none", "hidden"}},
	{ID: "style_legend_cfg", Label: "Legend Styling", Group: groupStyles, Summary: "Custom swatch size, padding, gaps, and background.", Tags: []string{"style", "legend", "swatch", "padding", "gap"}},
	{ID: "style_rotation", Label: "Rotated Labels", Group: groupStyles, Summary: "X-axis tick labels rotated for long category names.", Tags: []string{"style", "rotation", "labels", "tick", "angle"}},
	{ID: "style_padding", Label: "Custom Padding", Group: groupStyles, Summary: "Tight versus spacious chart padding.", Tags: []string{"style", "padding", "spacing", "inset"}},
	{ID: "style_kitchen", Label: "Kitchen Sink", Group: groupStyles, Summary: "All style knobs combined on a single chart.", Tags: []string{"style", "combined", "kitchen", "sink", "all"}},
	{ID: "style_high_contrast", Label: "High Contrast", Group: groupStyles, Summary: "Accessibility theme with bold axes, larger text, WCAG-compliant palette.", Tags: []string{"style", "high", "contrast", "accessibility", "wcag", "bold"}},
	{ID: "style_data_table", Label: "Data Table", Group: groupStyles, Summary: "Accessible table fallback showing chart data in tabular form.", Tags: []string{"style", "table", "data", "accessibility", "fallback", "a11y"}},
	{ID: "style_zoom", Label: "Zoom & Pan", Group: groupStyles, Summary: "Ctrl+scroll to zoom, drag to pan, shift+drag to select a region, double-click to reset.", Tags: []string{"style", "zoom", "pan", "scroll", "drag", "select", "brush"}},
}

func init() {
	for i := range demoEntries {
		demoEntries[i].idLower = strings.ToLower(demoEntries[i].ID)
		demoEntries[i].labelLower = strings.ToLower(demoEntries[i].Label)
	}
	sort.SliceStable(demoEntries, func(i, j int) bool {
		return entrySortBefore(demoEntries[i], demoEntries[j])
	})
}

func entryMatchesQuery(entry DemoEntry, query string) bool {
	if query == "" {
		return true
	}
	q := strings.ToLower(query)
	if strings.Contains(entry.idLower, q) ||
		strings.Contains(entry.labelLower, q) ||
		strings.Contains(entry.Group, q) {
		return true
	}
	for _, tag := range entry.Tags {
		if strings.Contains(tag, q) {
			return true
		}
	}
	return strings.Contains(strings.ToLower(entry.Summary), q)
}

func filteredEntries(app *ShowcaseApp) []DemoEntry {
	out := make([]DemoEntry, 0, len(demoEntries))
	for _, entry := range demoEntries {
		if app.SelectedGroup != groupAll && entry.Group != app.SelectedGroup {
			continue
		}
		if !entryMatchesQuery(entry, app.NavQuery) {
			continue
		}
		out = append(out, entry)
	}
	return out
}

func hasEntry(entries []DemoEntry, selected string) bool {
	for _, entry := range entries {
		if entry.ID == selected {
			return true
		}
	}
	return false
}

func selectedEntry(entries []DemoEntry, selected string) DemoEntry {
	for _, entry := range entries {
		if entry.ID == selected {
			return entry
		}
	}
	if len(entries) == 0 {
		return DemoEntry{}
	}
	return entries[0]
}

func preferredComponentForGroup(entries []DemoEntry) string {
	if len(entries) == 0 {
		return ""
	}
	best := entries[0]
	for _, entry := range entries[1:] {
		if entrySortBefore(entry, best) {
			best = entry
		}
	}
	return best.ID
}

func entrySortBefore(a, b DemoEntry) bool {
	if a.labelLower != b.labelLower {
		return a.labelLower < b.labelLower
	}
	return a.idLower < b.idLower
}
