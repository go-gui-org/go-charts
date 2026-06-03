package chart

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func TestExportSVG_Line(t *testing.T) {
	v := Line(LineCfg{
		BaseCfg: BaseCfg{
			ID:    "test-line",
			Width: 400, Height: 300,
		},
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:   "s1",
				Color:  gui.Hex(0x4E79A7),
				Points: []series.Point{{X: 0, Y: 1}, {X: 1, Y: 3}, {X: 2, Y: 2}},
			}),
		},
	})

	path := filepath.Join(t.TempDir(), "line.svg")
	if err := ExportSVG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	svg := readFile(t, path)
	assertValidSVG(t, svg, 400, 300)
	assertContains(t, svg, "<polyline")
}

func TestExportSVGString_Line(t *testing.T) {
	v := Line(LineCfg{
		BaseCfg: BaseCfg{
			ID:    "test-line",
			Width: 400, Height: 300,
		},
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:   "s1",
				Color:  gui.Hex(0x4E79A7),
				Points: []series.Point{{X: 0, Y: 1}, {X: 1, Y: 3}, {X: 2, Y: 2}},
			}),
		},
	})

	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
}

func TestExportSVG_Bar(t *testing.T) {
	v := Bar(BarCfg{
		BaseCfg: BaseCfg{
			ID:    "test-bar",
			Width: 400, Height: 300,
		},
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Name:  "s1",
				Color: gui.Hex(0xE15759),
				Values: []series.CategoryValue{
					{Label: "A", Value: 10},
					{Label: "B", Value: 20},
					{Label: "C", Value: 15},
				},
			}),
		},
	})

	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
	assertContains(t, svg, "<rect")
}

func TestExportSVG_Area(t *testing.T) {
	v := Area(AreaCfg{
		BaseCfg: BaseCfg{
			ID:    "test-area",
			Width: 400, Height: 300,
		},
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "signups",
				Color: gui.Hex(0x4E79A7),
				Points: []series.Point{
					{X: 0, Y: 10}, {X: 1, Y: 25},
					{X: 2, Y: 18}, {X: 3, Y: 30},
				},
			}),
		},
	})

	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
	assertContains(t, svg, "<polygon")
}

func TestExportSVG_Scatter(t *testing.T) {
	v := Scatter(ScatterCfg{
		BaseCfg: BaseCfg{
			ID:    "test-scatter",
			Width: 400, Height: 300,
		},
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "data",
				Color: gui.Hex(0xE15759),
				Points: []series.Point{
					{X: 1, Y: 2}, {X: 3, Y: 7},
					{X: 5, Y: 4}, {X: 8, Y: 9},
				},
			}),
		},
	})

	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
	assertContains(t, svg, "<circle")
}

func TestExportSVG_Pie(t *testing.T) {
	v := Pie(PieCfg{
		BaseCfg: BaseCfg{
			ID:    "test-pie",
			Width: 400, Height: 300,
		},
		Slices: []PieSlice{
			{Label: "A", Value: 40},
			{Label: "B", Value: 30},
			{Label: "C", Value: 20},
			{Label: "D", Value: 10},
		},
	})

	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
	assertContains(t, svg, "<path")
}

func TestExportSVG_Gauge(t *testing.T) {
	v := Gauge(GaugeCfg{
		BaseCfg: BaseCfg{
			ID:    "test-gauge",
			Width: 400, Height: 300,
		},
		Value: 72,
		Min:   0,
		Max:   100,
		Zones: []GaugeZone{
			{Threshold: 50, Color: gui.Hex(0x59A14F)},
			{Threshold: 80, Color: gui.Hex(0xF28E2B)},
			{Threshold: 100, Color: gui.Hex(0xE15759)},
		},
	})

	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
	assertContains(t, svg, "<path")
}

func TestExportSVG_Histogram(t *testing.T) {
	v := Histogram(HistogramCfg{
		BaseCfg: BaseCfg{
			ID:    "test-histogram",
			Width: 400, Height: 300,
		},
		Data: []float64{
			1, 2, 2, 3, 3, 3, 4, 4, 5, 6,
			7, 7, 8, 8, 8, 9, 9, 10,
		},
	})

	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
	assertContains(t, svg, "<rect")
}

func TestExportSVG_BoxPlot(t *testing.T) {
	v := BoxPlot(BoxPlotCfg{
		BaseCfg: BaseCfg{
			ID:    "test-boxplot",
			Width: 400, Height: 300,
		},
		Data: []BoxData{
			{Label: "A", Values: []float64{
				10, 15, 20, 25, 30, 35, 40, 45, 50, 90,
			}},
			{Label: "B", Values: []float64{
				5, 12, 18, 22, 28, 33, 38, 42, 48,
			}},
		},
	})

	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
}

func TestExportSVG_Candlestick(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	pts := make([]series.OHLC, 5)
	opens := []float64{100, 105, 103, 108, 106}
	highs := []float64{110, 115, 112, 118, 114}
	lows := []float64{98, 102, 100, 105, 103}
	closes := []float64{106, 103, 109, 107, 111}
	for i := range pts {
		pts[i] = series.OHLC{
			Time:  base.AddDate(0, 0, i),
			Open:  opens[i],
			High:  highs[i],
			Low:   lows[i],
			Close: closes[i],
		}
	}
	v := Candlestick(CandlestickCfg{
		BaseCfg: BaseCfg{
			ID:    "test-candlestick",
			Width: 400, Height: 300,
		},
		Series: []series.OHLCSeries{
			series.NewOHLC(series.OHLCCfg{
				Name:      "AAPL",
				ColorUp:   gui.Hex(0x26a69a),
				ColorDown: gui.Hex(0xef5350),
				Points:    pts,
			}),
		},
	})

	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
}

func TestExportSVG_RejectsNonChartView(t *testing.T) {
	_, err := ExportSVGString(nonChartView{}, 100, 100)
	if err == nil {
		t.Fatal("expected error for non-chart view")
	}
}

func TestExportSVG_RejectsZeroDimensions(t *testing.T) {
	v := Line(LineCfg{
		BaseCfg: BaseCfg{Width: 100, Height: 100},
	})
	_, err := ExportSVGString(v, 0, 100)
	if err == nil {
		t.Fatal("expected error for zero width")
	}
}

func TestExportSVG_TextEscaping(t *testing.T) {
	v := Line(LineCfg{
		BaseCfg: BaseCfg{
			ID:    "test-escape",
			Title: "A & B < C",
			Width: 400, Height: 300,
		},
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:   "s1",
				Color:  gui.Hex(0x4E79A7),
				Points: []series.Point{{X: 0, Y: 1}},
			}),
		},
	})

	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, svg, "A &amp; B &lt; C")
}

func TestExportSVG_RejectsExcessiveDimensions(t *testing.T) {
	v := Line(LineCfg{
		BaseCfg: BaseCfg{Width: 100, Height: 100},
	})
	_, err := ExportSVGString(v, 20000, 100)
	if err == nil {
		t.Fatal("expected error for excessive width")
	}
	_, err = ExportSVGString(v, 100, 20000)
	if err == nil {
		t.Fatal("expected error for excessive height")
	}
}

func TestWriteArcPath_DegenerateAndNonFinite(t *testing.T) {
	tests := []struct {
		name                         string
		cx, cy, rx, ry, start, sweep float32
	}{
		{"zero rx", 50, 50, 0, 10, 0, 1},
		{"zero ry", 50, 50, 10, 0, 0, 1},
		{"zero sweep", 50, 50, 10, 10, 0, 0},
		{"NaN cx", float32(math.NaN()), 50, 10, 10, 0, 1},
		{"Inf ry", 50, 50, 10, float32(math.Inf(1)), 0, 1},
		{"NaN sweep", 50, 50, 10, 10, 0, float32(math.NaN())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			writeArcPath(&b, tt.cx, tt.cy, tt.rx, tt.ry,
				tt.start, tt.sweep, true)
			if b.Len() != 0 {
				t.Fatalf("expected empty output, got %q", b.String())
			}
		})
	}
}

func TestWriteArcPath_LargeAndNegativeSweep(t *testing.T) {
	// Large sweep (> pi) should set largeArc=1.
	var b strings.Builder
	writeArcPath(&b, 50, 50, 40, 40, 0, 4, false)
	s := b.String()
	// "A40.0 40.0 0 1 1" — large-arc=1, sweep=1
	assertContains(t, s, " 1 1 ")

	// Negative sweep should set sweepFlag=0.
	b.Reset()
	writeArcPath(&b, 50, 50, 40, 40, 0, -1, false)
	s = b.String()
	// "A40.0 40.0 0 0 0" — large-arc=0, sweep=0
	assertContains(t, s, " 0 0 ")
}

func TestIsNonFinite32(t *testing.T) {
	if isNonFinite32(0) {
		t.Fatal("0 should be finite")
	}
	if isNonFinite32(1.5) {
		t.Fatal("1.5 should be finite")
	}
	if isNonFinite32(-100) {
		t.Fatal("-100 should be finite")
	}
	if !isNonFinite32(float32(math.NaN())) {
		t.Fatal("NaN should be non-finite")
	}
	if !isNonFinite32(float32(math.Inf(1))) {
		t.Fatal("+Inf should be non-finite")
	}
	if !isNonFinite32(float32(math.Inf(-1))) {
		t.Fatal("-Inf should be non-finite")
	}
}

func TestSVGOpacityAttributes(t *testing.T) {
	// Full alpha: no opacity attribute.
	var b strings.Builder
	writeOpacity(&b, gui.RGBA(255, 0, 0, 255))
	if b.Len() != 0 {
		t.Fatalf("full alpha should not emit opacity, got %q", b.String())
	}

	// Half alpha: stroke-opacity.
	b.Reset()
	writeOpacity(&b, gui.RGBA(255, 0, 0, 128))
	s := b.String()
	if !strings.Contains(s, "stroke-opacity") {
		t.Fatalf("expected stroke-opacity, got %q", s)
	}

	// Fill opacity.
	b.Reset()
	writeFillOpacity(&b, gui.RGBA(0, 0, 255, 64))
	s = b.String()
	if !strings.Contains(s, "fill-opacity") {
		t.Fatalf("expected fill-opacity, got %q", s)
	}

	// Full alpha fill: no attribute.
	b.Reset()
	writeFillOpacity(&b, gui.RGBA(0, 0, 255, 255))
	if b.Len() != 0 {
		t.Fatalf("full alpha should not emit fill-opacity, got %q", b.String())
	}
}

func TestSVGText_Rotation(t *testing.T) {
	txt := &svgText{
		x: 100, y: 50,
		text:   "rotated",
		style:  gui.TextStyle{Size: 14, RotationRadians: 1.5708},
		ascent: 11.2,
	}
	var b strings.Builder
	txt.writeSVG(&b)
	s := b.String()
	if !strings.Contains(s, "transform=\"rotate(") {
		t.Fatalf("expected rotate transform, got %q", s)
	}
	if !strings.Contains(s, "rotated") {
		t.Fatalf("expected text content, got %q", s)
	}
}

func TestCopyPoints_Truncation(t *testing.T) {
	// Empty input.
	cp := copyPoints(nil)
	if len(cp) != 0 {
		t.Fatalf("nil input: got len %d", len(cp))
	}

	// Normal input.
	pts := []float32{1, 2, 3, 4}
	cp = copyPoints(pts)
	if len(cp) != 4 {
		t.Fatalf("normal input: got len %d", len(cp))
	}

	// Verify it's a copy, not the same slice.
	pts[0] = 99
	if cp[0] == 99 {
		t.Fatal("copyPoints returned same backing array")
	}
}

func TestColorToCSS(t *testing.T) {
	tests := []struct {
		c    gui.Color
		want string
	}{
		{gui.RGBA(0, 0, 0, 255), "#000000"},
		{gui.RGBA(255, 255, 255, 255), "#ffffff"},
		{gui.RGBA(78, 121, 167, 255), "#4e79a7"},
		{gui.RGBA(255, 0, 128, 128), "#ff0080"}, // alpha ignored
	}
	for _, tt := range tests {
		got := colorToCSS(tt.c)
		if got != tt.want {
			t.Errorf("colorToCSS(%v) = %q, want %q", tt.c, got, tt.want)
		}
	}
}

// -----------------------------------------------------------
// Helpers
// -----------------------------------------------------------

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func assertValidSVG(t *testing.T, svg string, wantW, wantH int) {
	t.Helper()
	if !strings.HasPrefix(svg, "<svg") {
		t.Fatal("SVG does not start with <svg")
	}
	if !strings.HasSuffix(strings.TrimSpace(svg), "</svg>") {
		t.Fatal("SVG does not end with </svg>")
	}
	if !strings.Contains(svg, "viewBox=") {
		t.Fatal("SVG missing viewBox")
	}
	if !strings.Contains(svg, "<text") {
		t.Log("warning: SVG contains no <text> elements")
	}
}

func assertContains(t *testing.T, svg, substr string) {
	t.Helper()
	if !strings.Contains(svg, substr) {
		t.Fatalf("SVG missing expected element %q", substr)
	}
}
