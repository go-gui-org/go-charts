package main

import (
	"math"
	"time"

	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

// animRTState holds real-time demo state in a StateMap.
type animRTState struct {
	RTS     *chart.RealTimeSeries
	Counter int
	Started bool
}

const (
	nsAnimRT  = "anim-rt-demo"
	capAnimRT = 1
)

func demoAnimEntry(w *gui.Window) gui.View {
	// Three chart types with entry animation side by side.
	linePts := make([]series.Point, 30)
	for i := range 30 {
		x := float64(i)
		linePts[i] = series.Point{
			X: x,
			Y: math.Sin(x/4) * 50,
		}
	}
	linePts2 := make([]series.Point, 30)
	for i := range 30 {
		x := float64(i)
		linePts2[i] = series.Point{
			X: x,
			Y: math.Cos(x/3) * 30,
		}
	}

	lineChart := chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:      "anim-entry-line",
			Title:   "Line Entry",
			Sizing:  gui.FillFixed,
			Height:  250,
			Animate: true,
		},
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name: "Sin", Points: linePts,
			}),
			series.NewXY(series.XYCfg{
				Name: "Cos", Points: linePts2,
			}),
		},
		ShowArea: true,
	})

	barChart := chart.Bar(chart.BarCfg{
		BaseCfg: chart.BaseCfg{
			ID:      "anim-entry-bar",
			Title:   "Bar Entry",
			Sizing:  gui.FillFixed,
			Height:  250,
			Animate: true,
		},
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Name: "Q1",
				Values: []series.CategoryValue{
					{Label: "Jan", Value: 45},
					{Label: "Feb", Value: 62},
					{Label: "Mar", Value: 38},
					{Label: "Apr", Value: 71},
				},
			}),
		},
	})

	pieChart := chart.Pie(chart.PieCfg{
		BaseCfg: chart.BaseCfg{
			ID:      "anim-entry-pie",
			Title:   "Pie Entry",
			Sizing:  gui.FillFixed,
			Height:  250,
			Animate: true,
		},
		Slices: []chart.PieSlice{
			{Label: "Chrome", Value: 65},
			{Label: "Safari", Value: 18},
			{Label: "Firefox", Value: 10},
			{Label: "Edge", Value: 7},
		},
		ShowLabels:  true,
		ShowPercent: true,
	})

	replay := gui.Button(gui.ButtonCfg{
		ID:      "anim-replay",
		Sizing:  gui.FitFit,
		Padding: gui.NewPadding(6, 12, 6, 12),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Replay",
				TextStyle: gui.CurrentTheme().N4,
			}),
		},
		OnClick: func(ctx gui.EventCtx) {
			chart.ResetEntryAnimation(ctx.Window, "anim-entry-line")
			chart.ResetEntryAnimation(ctx.Window, "anim-entry-bar")
			chart.ResetEntryAnimation(ctx.Window, "anim-entry-pie")
		},
	})

	return demoWithCode(w, "anim-entry",
		gui.Column(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Spacing: gui.SomeF(12),
			Content: []gui.View{
				replay,
				lineChart,
				barChart,
				pieChart,
			},
		}),
		`chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Animate: true,
    },
    Series: []series.XY{data},
})`)
}

func demoAnimTransition(w *gui.Window) gui.View {
	// Toggle between two datasets on button click.
	type transState struct {
		Toggle  bool
		Version uint64
	}
	sm := gui.StateMap[string, transState](w, "anim-trans-demo", 1)
	ts, _ := sm.Get("state")

	var pts []series.Point
	if ts.Toggle {
		pts = make([]series.Point, 10)
		for i := range 10 {
			x := float64(i)
			pts[i] = series.Point{X: x, Y: 20 + math.Cos(x)*15}
		}
	} else {
		pts = make([]series.Point, 10)
		for i := range 10 {
			x := float64(i)
			pts[i] = series.Point{X: x, Y: 50 + math.Sin(x)*25}
		}
	}

	toggle := gui.Button(gui.ButtonCfg{
		ID:      "anim-trans-toggle",
		Sizing:  gui.FitFit,
		Padding: gui.NewPadding(6, 12, 6, 12),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Swap Data",
				TextStyle: gui.CurrentTheme().N4,
			}),
		},
		OnClick: func(ctx gui.EventCtx) {
			sm := gui.StateMap[string, transState](
				ctx.Window, "anim-trans-demo", 1)
			ts, _ := sm.Get("state")
			ts.Toggle = !ts.Toggle
			ts.Version++
			sm.Set("state", ts)
		},
	})

	return demoWithCode(w, "anim-transition",
		gui.Column(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Spacing: gui.SomeF(12),
			Content: []gui.View{
				toggle,
				chart.Line(chart.LineCfg{
					BaseCfg: chart.BaseCfg{
						ID:      "anim-transition",
						Title:   "Data Transition",
						Sizing:  gui.FillFixed,
						Height:  300,
						Version: ts.Version,
					},
					InteractionCfg: chart.InteractionCfg{
						AnimateTransitions: true,
					},
					Series: []series.XY{
						series.NewXY(series.XYCfg{
							Name: "Value", Points: pts,
						}),
					},
					ShowMarkers: true,
				}),
			},
		}),
		`chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Version: dataVersion,
    },
    InteractionCfg: chart.InteractionCfg{
        AnimateTransitions: true,
    },
    Series: []series.XY{data},
})`)
}

func demoAnimRealtime(w *gui.Window) gui.View {
	sm := gui.StateMap[string, animRTState](w, nsAnimRT, capAnimRT)
	st, _ := sm.Get("state")

	if st.RTS == nil {
		st.RTS = chart.NewRealTimeSeries(chart.RealTimeSeriesCfg{
			Name:   "Signal",
			MaxLen: 200,
		})
		sm.Set("state", st)
	}

	// Start a repeating animation to append data every ~50ms.
	// Defer via QueueCommand: view generator runs under w.mu.
	if !st.Started {
		st.Started = true
		sm.Set("state", st)
		w.QueueCommand(func(w *gui.Window) {
			w.AnimationAdd(&gui.Animate{
				AnimID:  "anim-rt-feeder",
				Delay:   50 * time.Millisecond,
				Repeat:  true,
				Refresh: gui.AnimationRefreshLayout,
				Callback: func(_ *gui.Animate, w *gui.Window) {
					sm := gui.StateMap[string, animRTState](
						w, nsAnimRT, capAnimRT)
					st, _ := sm.Get("state")
					if st.RTS == nil {
						return
					}
					st.Counter++
					x := float64(st.Counter)
					y := math.Sin(x/10)*30 + math.Cos(x/7)*15
					st.RTS.Append(series.Point{X: x, Y: y})
					sm.Set("state", st)
				},
			})
		})
	}

	snap := st.RTS.Snapshot()

	return demoWithCode(w, "anim-realtime",
		chart.Line(chart.LineCfg{
			BaseCfg: chart.BaseCfg{
				ID:      "anim-realtime",
				Title:   "Real-time Stream",
				Sizing:  gui.FillFixed,
				Height:  300,
				Version: st.RTS.Version(),
			},
			Series:     []series.XY{snap},
			AutoScroll: true,
			WindowSize: 100,
		}),
		`rts := chart.NewRealTimeSeries(
    chart.RealTimeSeriesCfg{
        Name:   "Signal",
        MaxLen: 200,
    },
)
// goroutine: rts.Append(series.Point{...})

chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Version: rts.Version(),
    },
    Series:     []series.XY{rts.Snapshot()},
    AutoScroll: true,
    WindowSize: 100,
})`)
}

func demoAnimFPS(w *gui.Window) gui.View {
	// Heavy chart: 1000 data points to demonstrate FPS tracking.
	pts := make([]series.Point, 1000)
	for i := range 1000 {
		x := float64(i) * 0.01
		pts[i] = series.Point{
			X: x,
			Y: math.Sin(x*3) * math.Cos(x) * 50,
		}
	}

	return demoWithCode(w, "anim-fps",
		chart.Line(chart.LineCfg{
			BaseCfg: chart.BaseCfg{
				ID:      "anim-fps",
				Title:   "FPS-Adaptive (1000 pts)",
				Sizing:  gui.FillFixed,
				Height:  300,
				Animate: true,
			},
			InteractionCfg: chart.InteractionCfg{
				EnableZoom: true,
				EnablePan:  true,
			},
			Series: []series.XY{
				series.NewXY(series.XYCfg{
					Name: "Heavy", Points: pts,
				}),
			},
			ShowMarkers: true,
		}),
		`chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Animate: true,
    },
    InteractionCfg: chart.InteractionCfg{
        EnableZoom: true,
        EnablePan:  true,
    },
    Series: []series.XY{heavy1000pts},
    ShowMarkers: true,
})`)
}
