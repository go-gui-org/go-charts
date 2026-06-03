package chart

import (
	"math"
	"testing"
	"time"

	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

// testOHLC builds a slice of OHLC points starting at the given base time,
// advancing by one day per point.
func testOHLC(base time.Time, points []series.OHLC) []series.OHLC {
	for i := range points {
		points[i].Time = base.AddDate(0, 0, i)
	}
	return points
}

func TestCandlestickValidate_NoSeries(t *testing.T) {
	cfg := CandlestickCfg{BaseCfg: BaseCfg{ID: "c1"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty Series")
	}
}

func TestCandlestickValidate_NegativeCandleWidth(t *testing.T) {
	cfg := CandlestickCfg{
		BaseCfg: BaseCfg{ID: "c1"},
		Series: []series.OHLCSeries{
			series.NewOHLC(series.OHLCCfg{
				Points: []series.OHLC{{Open: 1, High: 2, Low: 0.5, Close: 1.5}},
			}),
		},
		CandleWidth: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative CandleWidth")
	}
}

func TestCandlestickYAxisRange(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cv := &candlestickView{
		cfg: CandlestickCfg{
			BaseCfg: BaseCfg{ID: "r1"},
			Series: []series.OHLCSeries{
				series.NewOHLC(series.OHLCCfg{
					Points: testOHLC(base, []series.OHLC{
						{Open: 100, High: 110, Low: 90, Close: 105},
						{Open: 105, High: 120, Low: 100, Close: 115},
						{Open: 115, High: 125, Low: 108, Close: 110},
					}),
				}),
			},
		},
	}
	cv.cfg.applyDefaults()
	cv.cfg.XTimeFormat = "01/02"
	cv.buildAxes(&cv.cfg)

	// min(Low)=90, max(High)=125; with 5% padding the domain should include
	// values slightly below 90 and slightly above 125.
	// Transform maps domain min → pixelMin and domain max → pixelMax.
	// Use a 400px canvas: if 90 maps to a pixel > 0, the domain extends below 90.
	pixelMin := float32(400)
	pixelMax := float32(0)
	px90 := cv.yAxis.Transform(90, pixelMin, pixelMax)
	px125 := cv.yAxis.Transform(125, pixelMin, pixelMax)
	// 90 should be below the top of the canvas (not mapped to pixel 400)
	if px90 >= pixelMin {
		t.Errorf("Transform(90) = %.2f; domain min should be below 90 (padded)", px90)
	}
	// 125 should be above the bottom of the canvas (not mapped to pixel 0)
	if px125 <= pixelMax {
		t.Errorf("Transform(125) = %.2f; domain max should be above 125 (padded)", px125)
	}
}

func TestCandlestickXAxisLabels(t *testing.T) {
	base := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC) // March
	cv := &candlestickView{
		cfg: CandlestickCfg{
			BaseCfg:     BaseCfg{ID: "xl"},
			XTimeFormat: "Jan",
			Series: []series.OHLCSeries{
				series.NewOHLC(series.OHLCCfg{
					Points: testOHLC(base, []series.OHLC{
						{Open: 1, High: 2, Low: 0.5, Close: 1.5},
						{Open: 2, High: 3, Low: 1.5, Close: 2.5},
					}),
				}),
			},
		},
	}
	cv.cfg.applyDefaults()
	cv.buildAxes(&cv.cfg)

	ticks := cv.xAxis.Ticks(0, 200)
	if len(ticks) != 2 {
		t.Fatalf("expected 2 ticks, got %d", len(ticks))
	}
	for _, tk := range ticks {
		if tk.Label != "Mar" {
			t.Errorf("expected label %q, got %q", "Mar", tk.Label)
		}
	}
}

func TestCandlestickDefaultColors(t *testing.T) {
	s := series.NewOHLC(series.OHLCCfg{}) // zero-value colors
	if got := candleColor(s, true); got != defaultCandleUp {
		t.Errorf("up color: got %v, want %v", got, defaultCandleUp)
	}
	if got := candleColor(s, false); got != defaultCandleDown {
		t.Errorf("down color: got %v, want %v", got, defaultCandleDown)
	}
}

func TestCandlestickCustomColors(t *testing.T) {
	up := gui.Hex(0xffffff)
	down := gui.Hex(0x000000)
	s := series.NewOHLC(series.OHLCCfg{ColorUp: up, ColorDown: down})
	if got := candleColor(s, true); got != up {
		t.Errorf("custom up color: got %v, want %v", got, up)
	}
	if got := candleColor(s, false); got != down {
		t.Errorf("custom down color: got %v, want %v", got, down)
	}
}

func TestCandlestickHiddenSeriesExcludedFromRange(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cv := &candlestickView{
		cfg: CandlestickCfg{
			BaseCfg: BaseCfg{ID: "h1"},
			Series: []series.OHLCSeries{
				series.NewOHLC(series.OHLCCfg{
					// series 0: extreme values — should be ignored
					Points: testOHLC(base, []series.OHLC{
						{Open: 1000, High: 2000, Low: 500, Close: 1500},
					}),
				}),
				series.NewOHLC(series.OHLCCfg{
					Points: testOHLC(base, []series.OHLC{
						{Open: 100, High: 110, Low: 90, Close: 105},
					}),
				}),
			},
		},
		xyBase: xyBase{hidden: map[int]bool{0: true, 1: true}}, // hide series 0 (up=0, down=1)
	}
	cv.cfg.applyDefaults()
	cv.cfg.XTimeFormat = "01/02"
	cv.buildAxes(&cv.cfg)

	// Transform with pixelMin=400, pixelMax=0 (Y inverted):
	// values within [90,110] should map between 0 and 400 (exclusive of extremes).
	// Hidden series 0 has High=2000; if included, Transform(2000) would be
	// near 0, meaning Transform(110) would be nearly at 400 (floor).
	pxMin := float32(400)
	pxMax := float32(0)
	px90 := cv.yAxis.Transform(90, pxMin, pxMax)
	px110 := cv.yAxis.Transform(110, pxMin, pxMax)
	if px90 >= pxMin {
		t.Errorf("Transform(90) = %.2f; expected below pixelMin %.2f", px90, pxMin)
	}
	if px110 <= pxMax {
		t.Errorf("Transform(110) = %.2f; expected above pixelMax %.2f", px110, pxMax)
	}
	// Sanity: pixelMin=400 is bottom, pixelMax=0 is top, so higher values
	// map to smaller pixel positions (px110 < px90). If hidden series were
	// included, domain would be ~[500,2100] making px90 and px110 nearly equal.
	if px90-px110 < 10 {
		t.Errorf("series 0 high likely included: px90=%.2f px110=%.2f too close", px90, px110)
	}
}

func TestCandlestickNaNInfExcludedFromRange(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cv := &candlestickView{
		cfg: CandlestickCfg{
			BaseCfg: BaseCfg{ID: "nan1"},
			Series: []series.OHLCSeries{
				series.NewOHLC(series.OHLCCfg{
					Points: testOHLC(base, []series.OHLC{
						{Open: math.NaN(), High: math.Inf(1), Low: math.Inf(-1), Close: math.NaN()},
						{Open: 100, High: 110, Low: 90, Close: 105},
					}),
				}),
			},
		},
	}
	cv.cfg.applyDefaults()
	cv.cfg.XTimeFormat = "01/02"
	// buildAxes must not panic and must ignore the NaN/Inf point.
	cv.buildAxes(&cv.cfg)

	pxMin := float32(400)
	pxMax := float32(0)
	px90 := cv.yAxis.Transform(90, pxMin, pxMax)
	px110 := cv.yAxis.Transform(110, pxMin, pxMax)
	if px90 >= pxMin {
		t.Errorf("NaN/Inf point affected axis min: Transform(90)=%.2f", px90)
	}
	if px110 <= pxMax {
		t.Errorf("NaN/Inf point affected axis max: Transform(110)=%.2f", px110)
	}
}

func TestCandlestickEmptyPoints(t *testing.T) {
	cv := &candlestickView{
		cfg: CandlestickCfg{
			BaseCfg: BaseCfg{ID: "empty1"},
			Series: []series.OHLCSeries{
				series.NewOHLC(series.OHLCCfg{Points: []series.OHLC{}}),
			},
		},
	}
	cv.cfg.applyDefaults()
	cv.cfg.XTimeFormat = "01/02"
	// Must not panic with empty Points slice.
	cv.buildAxes(&cv.cfg)
}
