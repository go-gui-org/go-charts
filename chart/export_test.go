package chart

import (
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

// nonChartView is a minimal gui.View that does not implement
// Drawer, used to test ExportPNG rejection.
type nonChartView struct{}

func (nonChartView) GenerateLayout(_ *gui.Window) gui.Layout { return gui.Layout{} }

func TestExportPNG_Line(t *testing.T) {
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

	path := filepath.Join(t.TempDir(), "line.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_Bar(t *testing.T) {
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

	path := filepath.Join(t.TempDir(), "bar.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_LineWithAxes(t *testing.T) {
	xa := axis.NewLinear(axis.LinearCfg{AutoRange: true})
	xa.SetRange(0, 10)
	ya := axis.NewLinear(axis.LinearCfg{AutoRange: true})
	ya.SetRange(0, 100)

	v := Line(LineCfg{
		BaseCfg: BaseCfg{
			ID:    "test-axes",
			Width: 800, Height: 600,
		},
		XAxis: xa,
		YAxis: ya,
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:   "data",
				Color:  gui.Hex(0x59A14F),
				Points: []series.Point{{X: 1, Y: 20}, {X: 5, Y: 80}, {X: 9, Y: 40}},
			}),
		},
		ShowMarkers: true,
	})

	path := filepath.Join(t.TempDir(), "axes.png")
	if err := ExportPNG(v, 800, 600, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 800, 600)
}

func TestExportPNG_LineAreaMonotonicX(t *testing.T) {
	// Monotonic X — area fill should render correctly.
	v := Line(LineCfg{
		BaseCfg: BaseCfg{
			ID:    "test-area-monotonic",
			Width: 400, Height: 300,
		},
		ShowArea: true,
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "s1",
				Color: gui.Hex(0x59A14F),
				Points: []series.Point{
					{X: 1, Y: 10}, {X: 2, Y: 30}, {X: 3, Y: 20},
					{X: 4, Y: 50}, {X: 5, Y: 40}, {X: 6, Y: 60},
				},
			}),
		},
	})

	path := filepath.Join(t.TempDir(), "area_monotonic.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_LineAreaNonMonotonicX(t *testing.T) {
	// Non-monotonic X — fill trapezoids follow source point order,
	// same as the rendered polyline. Export must succeed.
	v := Line(LineCfg{
		BaseCfg: BaseCfg{
			ID:    "test-area-nonmonotonic",
			Width: 400, Height: 300,
		},
		ShowArea: true,
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:  "s1",
				Color: gui.Hex(0x59A14F),
				Points: []series.Point{
					{X: 0, Y: 0}, {X: 10, Y: 10},
					{X: 5, Y: 5}, {X: 15, Y: 15},
				},
			}),
		},
	})

	path := filepath.Join(t.TempDir(), "area_nonmonotonic.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_Candlestick(t *testing.T) {
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

	path := filepath.Join(t.TempDir(), "candlestick.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_Area(t *testing.T) {
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

	path := filepath.Join(t.TempDir(), "area.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_Scatter(t *testing.T) {
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

	path := filepath.Join(t.TempDir(), "scatter.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_Pie(t *testing.T) {
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

	path := filepath.Join(t.TempDir(), "pie.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_Gauge(t *testing.T) {
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

	path := filepath.Join(t.TempDir(), "gauge.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_Histogram(t *testing.T) {
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

	path := filepath.Join(t.TempDir(), "histogram.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_BoxPlot(t *testing.T) {
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

	path := filepath.Join(t.TempDir(), "boxplot.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_RejectsNonChartView(t *testing.T) {
	v := nonChartView{}
	err := ExportPNG(v, 100, 100, "/dev/null")
	if err == nil {
		t.Fatal("expected error for non-chart view")
	}
}

func TestExportPNG_RejectsZeroDimensions(t *testing.T) {
	v := Line(LineCfg{
		BaseCfg: BaseCfg{Width: 100, Height: 100},
	})
	err := ExportPNG(v, 0, 100, "/dev/null")
	if err == nil {
		t.Fatal("expected error for zero width")
	}
}

func TestExportPNG_RejectsExcessiveDimensions(t *testing.T) {
	v := Line(LineCfg{
		BaseCfg: BaseCfg{Width: 100, Height: 100},
	})
	err := ExportPNG(v, 20000, 100, "/dev/null")
	if err == nil {
		t.Fatal("expected error for excessive width")
	}
	err = ExportPNG(v, 100, 20000, "/dev/null")
	if err == nil {
		t.Fatal("expected error for excessive height")
	}
}

func TestRasterizeTriBatches_SkipsNaNInfBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Fill white so we can detect any writes.
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	snapshot := make([]byte, len(img.Pix))
	copy(snapshot, img.Pix)

	nan := float32(math.NaN())
	inf := float32(math.Inf(1))
	batches := []gui.DrawCanvasTriBatch{
		{Color: gui.Hex(0xFF0000), Triangles: []float32{
			nan, 0, 5, 0, 5, 5,
		}},
		{Color: gui.Hex(0x00FF00), Triangles: []float32{
			0, 0, inf, 0, 5, 5,
		}},
	}
	rasterizeTriBatches(img, batches)

	for i, v := range img.Pix {
		if v != snapshot[i] {
			t.Fatal("NaN/Inf batch modified image pixels")
		}
	}
}

func TestRasterizeTriBatches_SharedEdgeCoverage(t *testing.T) {
	// Two triangles sharing an edge should produce full coverage
	// at the shared edge pixels (no seam).
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Black background.
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
	}

	// Two triangles forming a 10x10 quad, same batch.
	batch := gui.DrawCanvasTriBatch{
		Color: gui.RGBA(255, 0, 0, 255),
		Triangles: []float32{
			0, 0, 10, 0, 10, 10,
			0, 0, 10, 10, 0, 10,
		},
	}
	rasterizeTriBatches(img, []gui.DrawCanvasTriBatch{batch})

	// Interior pixel (5,5) must be fully red, no background
	// bleed from seam artifacts.
	off := 5*img.Stride + 5*4
	r, g, b, a := img.Pix[off], img.Pix[off+1], img.Pix[off+2], img.Pix[off+3]
	if r != 255 || g != 0 || b != 0 || a != 255 {
		t.Fatalf("interior pixel: got (%d,%d,%d,%d), want (255,0,0,255)",
			r, g, b, a)
	}
}

func TestRasterizeTriBatches_SkipsShortBatch(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Batch with fewer than 6 floats (one triangle needs 6).
	batch := gui.DrawCanvasTriBatch{
		Color:     gui.Hex(0xFF0000),
		Triangles: []float32{0, 0, 1, 1},
	}
	// Must not panic.
	rasterizeTriBatches(img, []gui.DrawCanvasTriBatch{batch})
}

func TestRasterizeTriBatches_CoverageClamp(t *testing.T) {
	// Three overlapping triangles covering the same pixel.
	// Coverage should clamp at msaaSamples, not overflow uint8.
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
	}

	tri := []float32{0, 0, 4, 0, 4, 4}
	batch := gui.DrawCanvasTriBatch{
		Color:     gui.RGBA(0, 0, 255, 255),
		Triangles: append(append(tri, tri...), tri...),
	}
	// Must not panic from uint8 overflow.
	rasterizeTriBatches(img, []gui.DrawCanvasTriBatch{batch})

	// Center pixel should be fully blue.
	off := 2*img.Stride + 2*4
	if img.Pix[off+2] != 255 {
		t.Fatalf("blue channel: got %d, want 255", img.Pix[off+2])
	}
}

func assertValidPNG(t *testing.T, path string, wantW, wantH int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatal("invalid PNG:", err)
	}
	b := img.Bounds()
	if b.Dx() != wantW || b.Dy() != wantH {
		t.Fatalf("dimensions: got %dx%d, want %dx%d",
			b.Dx(), b.Dy(), wantW, wantH)
	}
}
