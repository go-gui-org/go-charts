package main

import (
	"time"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoTimeAxis(w *gui.Window) gui.View {
	t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	pts := make([]series.Point, 12)
	values := []float64{
		42, 47, 53, 61, 58, 72, 79, 76, 68, 55, 48, 44,
	}
	for i := range pts {
		t := t0.AddDate(0, i, 0)
		pts[i] = series.Point{
			X: float64(t.Unix()) + float64(t.Nanosecond())/1e9,
			Y: values[i],
		}
	}

	xAxis := axis.NewTime(axis.TimeCfg{
		Title: "Month",
		Min:   t0,
		Max:   t0.AddDate(0, 11, 0),
	})

	return demoWithCode(w, "line-time", chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "line-time",
			Title:          "Monthly Average Temperature",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		XAxis: xAxis,
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:   "Austin, TX",
				Color:  gui.Hex(0x4E79A7),
				Points: pts,
			}),
		},
	}), `t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

xAxis := axis.NewTime(axis.TimeCfg{
    Title: "Month",
    Min:   t0,
    Max:   t0.AddDate(0, 11, 0),
})

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Monthly Average Temperature",
    },
    XAxis: xAxis,
    Series: []series.XY{
        series.NewXY(series.XYCfg{
            Name:   "Austin, TX",
            Color:  gui.Hex(0x4E79A7),
            Points: pts, // X = timeToSeconds, Y = temp
        }),
    },
})`)
}
