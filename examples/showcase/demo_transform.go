package main

import (
	"math"

	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/transform"
	"github.com/go-gui-org/go-gui/gui"
)

func demoTransformMA(w *gui.Window) gui.View {
	// Noisy sine wave.
	pts := make([]series.Point, 50)
	for i := range 50 {
		x := float64(i)
		pts[i] = series.Point{
			X: x,
			Y: math.Sin(x/5)*10 + noisyOffset(i),
		}
	}
	raw := series.NewXY(series.XYCfg{
		Name:   "Signal",
		Color:  gui.Hex(0xBAB0AC),
		Points: pts,
	})
	sma := transform.SMA(raw, 5)
	ema := transform.EMA(raw, 5)
	wma := transform.WMA(raw, 5)

	return demoWithCode(w, "transform-ma", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "transform-ma",
			Title:          "Moving Averages",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Series: []series.XY{raw, sma, ema, wma},
	}), `raw := series.NewXY(series.XYCfg{
    Name:   "Signal",
    Points: noisySineWave,
})
sma := transform.SMA(raw, 5)
ema := transform.EMA(raw, 5)
wma := transform.WMA(raw, 5)

chart.Line(chart.LineCfg{
    Series: []series.XY{raw, sma, ema, wma},
})`)
}

func demoTransformRegression(w *gui.Window) gui.View {
	// Quadratic data with noise.
	pts := make([]series.Point, 30)
	for i := range 30 {
		x := float64(i)
		pts[i] = series.Point{
			X: x,
			Y: 0.05*x*x - 0.5*x + noisyOffset(i)*0.5,
		}
	}
	raw := series.NewXY(series.XYCfg{
		Name:   "Data",
		Color:  gui.Hex(0xBAB0AC),
		Points: pts,
	})
	linear := transform.LinearTrend(raw)
	poly := transform.PolynomialRegression(raw, 3, 60)

	return demoWithCode(w, "transform-regression",
		chart.Line(chart.LineCfg{
			BaseCfg: chart.BaseCfg{
				ID:             "transform-regression",
				Title:          "Regression",
				Sizing:         gui.FillFixed,
				Height:         350,
				LegendPosition: &posBottom,
			},
			ShowMarkers: true,
			Series:      []series.XY{raw, linear, poly},
		}), `raw := series.NewXY(series.XYCfg{
    Name:   "Data",
    Points: scatteredQuadratic,
})
linear := transform.LinearTrend(raw)
poly := transform.PolynomialRegression(raw, 3, 60)

chart.Line(chart.LineCfg{
    ShowMarkers: true,
    Series:      []series.XY{raw, linear, poly},
})`)
}

func demoTransformBands(w *gui.Window) gui.View {
	// Random walk price series.
	pts := make([]series.Point, 100)
	price := 100.0
	for i := range 100 {
		price += noisyOffset(i) * 0.8
		pts[i] = series.Point{X: float64(i), Y: price}
	}
	raw := series.NewXY(series.XYCfg{
		Name:   "Price",
		Color:  gui.Hex(0x4E79A7),
		Points: pts,
	})
	upper, middle, lower := transform.BollingerBands(raw, 20, 2)

	return demoWithCode(w, "transform-bands", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "transform-bands",
			Title:          "Bollinger Bands",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Series: []series.XY{raw, upper, middle, lower},
	}), `price := randomWalkSeries("Price", 100)
upper, middle, lower := transform.BollingerBands(price, 20, 2)

chart.Line(chart.LineCfg{
    Series: []series.XY{price, upper, middle, lower},
})`)
}

func demoTransformDownsample(w *gui.Window) gui.View {
	// Large dataset.
	pts := make([]series.Point, 1000)
	for i := range 1000 {
		x := float64(i)
		pts[i] = series.Point{
			X: x,
			Y: math.Sin(x/50) * math.Cos(x/13) * 10,
		}
	}
	raw := series.NewXY(series.XYCfg{
		Name:   "Original (1000 pts)",
		Color:  gui.Hex(0xBAB0AC),
		Points: pts,
	})
	down := transform.LTTB(raw, 50)

	return demoWithCode(w, "transform-downsample",
		chart.Line(chart.LineCfg{
			BaseCfg: chart.BaseCfg{
				ID:             "transform-downsample",
				Title:          "LTTB Downsampling",
				Sizing:         gui.FillFixed,
				Height:         350,
				LegendPosition: &posBottom,
			},
			ShowMarkers: true,
			Series:      []series.XY{raw, down},
		}), `pts := make([]series.Point, 1000)
for i := range 1000 {
    x := float64(i)
    pts[i] = series.Point{
        X: x,
        Y: math.Sin(x/50) * math.Cos(x/13) * 10,
    }
}
raw := series.NewXY(series.XYCfg{
    Name:   "Original (1000 pts)",
    Points: pts,
})
down := transform.LTTB(raw, 50)

chart.Line(chart.LineCfg{
    ShowMarkers: true,
    Series:      []series.XY{raw, down},
})`)
}

// noisyOffset returns a deterministic pseudo-random offset for
// repeatable demo data. Range roughly [-2, +2].
func noisyOffset(i int) float64 {
	// Simple hash-based noise; avoids math/rand import.
	x := uint32(i*2654435761) >> 16
	return (float64(x%1000)/250 - 2)
}
