package chart

import (
	"math"
	"strings"
	"testing"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func newLinearAxis(lo, hi float64) axis.Axis {
	a := axis.NewLinear(axis.LinearCfg{AutoRange: true})
	a.SetRange(lo, hi)
	return a
}

func sampleErrorXY() series.ErrorXY {
	return series.NewErrorXY(series.ErrorXYCfg{
		Name:  "sensor",
		Color: gui.Hex(0x4E79A7),
		Points: []series.ErrorPoint{
			{X: 1, Y: 10, YErr: series.Symmetric(1.5)},
			{X: 2, Y: 20, YErr: series.ErrorBar{Low: 2, High: 3}},
			{X: 3, Y: 15, XErr: series.Symmetric(0.5)},
		},
	})
}

func TestScatter_WithErrorSeriesRenders(t *testing.T) {
	v := Scatter(ScatterCfg{
		BaseCfg:     BaseCfg{ID: "scatter-err", Width: 400, Height: 300},
		ErrorSeries: []series.ErrorXY{sampleErrorXY()},
	})
	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
}

func TestScatter_OnlyErrorSeriesValidates(t *testing.T) {
	cfg := ScatterCfg{
		ErrorSeries: []series.ErrorXY{sampleErrorXY()},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLine_WithErrorSeriesRenders(t *testing.T) {
	v := Line(LineCfg{
		BaseCfg:     BaseCfg{ID: "line-err", Width: 400, Height: 300},
		ErrorSeries: []series.ErrorXY{sampleErrorXY()},
	})
	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
}

func TestLine_OnlyErrorSeriesValidates(t *testing.T) {
	cfg := LineCfg{
		ErrorSeries: []series.ErrorXY{sampleErrorXY()},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestScatter_MixedSeriesAndErrorSeries(t *testing.T) {
	v := Scatter(ScatterCfg{
		BaseCfg: BaseCfg{ID: "scatter-mix", Width: 400, Height: 300},
		Series: []series.XY{
			series.NewXY(series.XYCfg{
				Name:   "plain",
				Points: []series.Point{{X: 0, Y: 5}, {X: 4, Y: 25}},
			}),
		},
		ErrorSeries: []series.ErrorXY{sampleErrorXY()},
	})
	svg, err := ExportSVGString(v, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	assertValidSVG(t, svg, 400, 300)
}

func TestErrorSeries_AxisBoundsExpand(t *testing.T) {
	// ErrorSeries with Y=10 +/-5 should drive yMin <= 5 even
	// when no plain XY series is supplied.
	sv := Scatter(ScatterCfg{
		BaseCfg: BaseCfg{ID: "scatter-bounds", Width: 400, Height: 300},
		ErrorSeries: []series.ErrorXY{
			series.NewErrorXY(series.ErrorXYCfg{
				Points: []series.ErrorPoint{
					{X: 1, Y: 10, YErr: series.Symmetric(5)},
					{X: 2, Y: 10, YErr: series.Symmetric(5)},
				},
			}),
		},
	}).(*scatterView)
	if !sv.updateAxes() {
		t.Fatal("updateAxes returned false")
	}
	yMin, yMax := sv.yAxis.Domain()
	if yMin > 5 {
		t.Errorf("yMin = %v, want <= 5", yMin)
	}
	if yMax < 15 {
		t.Errorf("yMax = %v, want >= 15", yMax)
	}
}

func TestErrorSeries_NonFiniteSurvivesRender(t *testing.T) {
	v := Scatter(ScatterCfg{
		BaseCfg: BaseCfg{ID: "scatter-nan", Width: 400, Height: 300},
		ErrorSeries: []series.ErrorXY{
			series.NewErrorXY(series.ErrorXYCfg{
				Points: []series.ErrorPoint{
					{X: 1, Y: 10, YErr: series.ErrorBar{
						Low: math.NaN(), High: math.Inf(1),
					}},
					{X: 2, Y: 20, YErr: series.Symmetric(2)},
				},
			}),
		},
	})
	if _, err := ExportSVGString(v, 400, 300); err != nil {
		t.Fatal(err)
	}
}

func TestNearestErrorXYPoint_SnapAndMiss(t *testing.T) {
	es := []series.ErrorXY{sampleErrorXY()}
	pa := plotArea{
		plotRect: plotRect{Left: 0, Right: 100, Top: 0, Bottom: 100},
		XAxis:    newLinearAxis(0, 4),
		YAxis:    newLinearAxis(0, 30),
	}
	// Point (2, 20) → x: 50px, y: 100 - (20/30)*100 = 33.3
	_, _, _, _, ok := nearestErrorXYPoint(es, pa, 50, 34, 5)
	if !ok {
		t.Error("expected snap to nearest point")
	}
	_, _, _, _, ok = nearestErrorXYPoint(es, pa, 200, 200, 5)
	if ok {
		t.Error("expected miss for far cursor")
	}
}

func TestFormatErrorPointLabel_AllCombinations(t *testing.T) {
	cases := []struct {
		name string
		p    series.ErrorPoint
		want []string
	}{
		{"named", series.ErrorPoint{X: 1, Y: 2}, []string{"X: 1", "Y: 2"}},
		{"y-err", series.ErrorPoint{X: 1, Y: 2,
			YErr: series.ErrorBar{Low: 0.5, High: 1.5}},
			[]string{"Y err: +1.5/-0.5"}},
		{"x-err", series.ErrorPoint{X: 1, Y: 2,
			XErr: series.Symmetric(0.5)},
			[]string{"X err: +0.5/-0.5"}},
	}
	for _, c := range cases {
		got := formatErrorPointLabel(c.name, c.p, "X", "Y")
		for _, sub := range c.want {
			if !strings.Contains(got, sub) {
				t.Errorf("%s: %q missing %q", c.name, got, sub)
			}
		}
	}
}

func TestSafeErr_NonFiniteAndNegative(t *testing.T) {
	cases := []struct {
		in     series.ErrorBar
		wantLo float64
		wantHi float64
	}{
		{series.ErrorBar{Low: 1, High: 2}, 1, 2},
		{series.ErrorBar{Low: -1, High: 2}, 0, 2},
		{series.ErrorBar{Low: math.NaN(), High: 2}, 0, 2},
		{series.ErrorBar{Low: 1, High: math.Inf(1)}, 1, 0},
		{series.ErrorBar{}, 0, 0},
	}
	for i, c := range cases {
		lo, hi := safeErr(c.in)
		if lo != c.wantLo || hi != c.wantHi {
			t.Errorf("case %d: got (%v,%v), want (%v,%v)",
				i, lo, hi, c.wantLo, c.wantHi)
		}
	}
}

func TestClampF32(t *testing.T) {
	if v := clampF32(5, 0, 10); v != 5 {
		t.Errorf("in-range = %v, want 5", v)
	}
	if v := clampF32(-1, 0, 10); v != 0 {
		t.Errorf("below = %v, want 0", v)
	}
	if v := clampF32(20, 0, 10); v != 10 {
		t.Errorf("above = %v, want 10", v)
	}
	if v := clampF32(float32(math.NaN()), 0, 10); v != 0 {
		t.Errorf("NaN = %v, want 0", v)
	}
}
