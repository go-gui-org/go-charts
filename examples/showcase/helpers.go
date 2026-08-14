package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
	"github.com/go-gui-org/go-gui/gui/highlight"
)

// markdownStyleWithHighlighter returns the default markdown style with
// the chroma-backed syntax highlighter enabled for fenced code blocks.
func markdownStyleWithHighlighter() gui.MarkdownStyle {
	s := gui.DefaultMarkdownStyle()
	s.CodeHighlighter = highlight.Default()
	return s
}

// posBottom is the default legend position for all showcase charts.
var posBottom = theme.LegendBottom

// chartTypeDescriptions maps the chart-group prefix (derived from
// the demo ID) to a short educational summary of the chart type.
var chartTypeDescriptions = map[string]string{
	"line": "Line charts display data points connected by straight" +
		" line segments. They are ideal for showing trends over" +
		" time or continuous data, making it easy to spot patterns," +
		" peaks, and valleys in a dataset.",

	"bar": "Bar charts represent categorical data with rectangular" +
		" bars whose lengths are proportional to the values they" +
		" represent. They excel at comparing discrete categories" +
		" side by side and support grouped, stacked, and" +
		" horizontal layouts.",

	"area": "Area charts are line charts with the region between the" +
		" line and the axis filled in. The filled region emphasizes" +
		" volume and magnitude of change over time, and stacked" +
		" area charts show how parts contribute to a whole.",

	"scatter": "Scatter plots display individual data points on a" +
		" two-dimensional plane, showing the relationship between" +
		" two variables. They are commonly used to identify" +
		" correlations, clusters, and outliers in datasets.",

	"pie": "Pie charts divide a circle into proportional slices to" +
		" show how parts make up a whole. They are best suited for" +
		" displaying a small number of categories where the" +
		" relative share of each category is the primary focus.",

	"gauge": "Gauge charts display a single value within a defined" +
		" range using a semi-circular arc, resembling a speedometer" +
		" or dial. Colored zones can indicate thresholds such as" +
		" normal, warning, and critical ranges.",

	"candlestick": "Candlestick charts visualize financial price" +
		" movements over time. Each candle shows the open, high," +
		" low, and close (OHLC) prices for a period. The body" +
		" color indicates whether the price moved up or down.",

	"histogram": "Histograms display the frequency distribution of" +
		" continuous data by grouping values into bins. The height" +
		" of each bar represents how many data points fall within" +
		" that bin's range, revealing the shape and spread of the" +
		" distribution.",

	"boxplot": "Box plots, also known as box-and-whisker plots, are" +
		" graphical representations used in statistics to display" +
		" the distribution of a dataset. They show key statistics" +
		" such as the minimum, first quartile, median, third" +
		" quartile, and maximum, allowing for a quick visual" +
		" summary of the data's spread and skewness.",

	"combo": "Combo charts overlay bar and line series on shared" +
		" category axes, making it easy to compare absolute values" +
		" (bars) with trends or rates (lines) in a single view." +
		" Bars render underneath and lines draw on top.",

	"treemap": "Treemap charts display hierarchical data as nested" +
		" rectangles whose areas are proportional to their values." +
		" A squarified layout algorithm arranges rectangles to" +
		" minimize aspect ratios, making it easy to compare sizes" +
		" across categories and subcategories at a glance.",

	"waterfall": "Waterfall charts show how an initial value is" +
		" affected by a series of intermediate positive or negative" +
		" values. Each bar floats, starting where the previous bar" +
		" ended. They are widely used for financial statements," +
		" variance analysis, and bridge charts.",

	"transform": "Data transforms are pure functions that process" +
		" series data before rendering. They include moving averages" +
		" (SMA, EMA, WMA) for smoothing, linear and polynomial" +
		" regression for trend fitting, Bollinger bands and min/max" +
		" envelopes for range analysis, LTTB downsampling for large" +
		" datasets, and binning for data grouping.",

	"sparkline": "Sparklines are compact inline mini-charts that" +
		" show trend at a glance without axes, labels, or legends." +
		" They are designed to be embedded in text, tables, or" +
		" dashboards where space is limited. Variants include line," +
		" area, and bar styles with optional markers and color bands.",
}

// demoWithCode wraps a chart view with its source code shown
// below as a markdown code block. If the demo ID maps to a
// known chart type, an educational description is inserted
// between the chart and the code.
func demoWithCode(
	w *gui.Window, id string, chartView gui.View, code string,
) gui.View {
	t := gui.CurrentTheme()
	source := "```go\n" + code + "\n```"

	views := collectExportable(id, chartView)

	if prefix, _, ok := strings.Cut(id, "-"); ok {
		if desc, found := chartTypeDescriptions[prefix]; found {
			views = append(views, gui.Text(gui.TextCfg{
				Text:      desc,
				TextStyle: t.N4,
				Mode:      gui.TextModeWrap,
			}))
		}
	}

	views = append(views,
		line(),
		gui.Text(gui.TextCfg{
			Text:      "Code",
			TextStyle: t.B3,
		}),
		w.Markdown(gui.MarkdownCfg{
			ID:      "code-" + id,
			Source:  source,
			Padding: gui.NoPadding,
			Style:   markdownStyleWithHighlighter(),
		}),
	)

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.NoPadding,
		Spacing: gui.SomeF(12),
		Content: views,
	})
}

// collectExportable returns the chart view followed by export
// buttons. If the view is a container (not a Drawer), it walks
// Content() children and adds export buttons after each chart.
func collectExportable(id string, v gui.View) []gui.View {
	if _, ok := v.(chart.Drawer); ok {
		return []gui.View{v, exportButtons(id, v)}
	}
	children := v.Content()
	if len(children) == 0 {
		return []gui.View{v}
	}
	views := make([]gui.View, 0, len(children)*2)
	idx := 0
	for _, child := range children {
		views = append(views, child)
		if _, ok := child.(chart.Drawer); ok {
			btnID := fmt.Sprintf("%s-%d", id, idx)
			views = append(views, exportButtons(btnID, child))
			idx++
		}
	}
	return views
}

func exportButtons(id string, chartView gui.View) gui.View {
	t := gui.CurrentTheme()
	btnPad := gui.NewPadding(4, 12, 4, 12)

	return gui.Row(gui.ContainerCfg{
		Sizing:     gui.FillFit,
		Padding:    gui.NoPadding,
		SizeBorder: gui.NoBorder,
		Spacing:    gui.SomeF(8),
		Content: []gui.View{
			gui.Button(gui.ButtonCfg{
				ID:      "export-svg-" + id,
				Sizing:  gui.FitFit,
				Padding: btnPad,
				Content: []gui.View{gui.Text(gui.TextCfg{
					Text:      "Show SVG",
					TextStyle: t.N4,
				})},
				OnClick: func(ctx gui.EventCtx) {
					path := filepath.Join(tempDir(), id+".svg")
					if chart.ExportSVG(chartView, 800, 600, path) == nil {
						openFile(path)
					}
				},
			}),
			gui.Button(gui.ButtonCfg{
				ID:      "export-png-" + id,
				Sizing:  gui.FitFit,
				Padding: btnPad,
				Content: []gui.View{gui.Text(gui.TextCfg{
					Text:      "Show PNG",
					TextStyle: t.N4,
				})},
				OnClick: func(ctx gui.EventCtx) {
					path := filepath.Join(tempDir(), id+".png")
					if chart.ExportPNG(chartView, 800, 600, path) == nil {
						openFile(path)
					}
				},
			}),
		},
	})
}

// tempDir returns a stable subdirectory in the system temp folder
// for chart exports.
func tempDir() string {
	dir := filepath.Join(os.TempDir(), "go-charts-export")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// openFile opens a file with the platform default viewer.
func openFile(path string) {
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", path).Start()
	case "windows":
		_ = exec.Command("cmd", "/c", "start", "", path).Start()
	default:
		_ = exec.Command("xdg-open", path).Start()
	}
}

func line() gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:     gui.FillFit,
		Padding:    gui.NewPadding(3, 0, 0, 0),
		SizeBorder: gui.NoBorder,
		Radius:     gui.NoRadius,
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:     gui.FillFit,
				Padding:    gui.NoPadding,
				SizeBorder: gui.NoBorder,
				Radius:     gui.NoRadius,
				Color:      t.ColorActive,
				Height:     1,
			}),
		},
	})
}
