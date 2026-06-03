package transform

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/series"
)

func TestBinSumBasic(t *testing.T) {
	// 10 points in [0,10), 2 bins: [0,5) and [5,10).
	pts := make([]series.Point, 10)
	for i := range 10 {
		pts[i] = series.Point{X: float64(i), Y: 1}
	}
	s := series.NewXY(series.XYCfg{Name: "data", Points: pts})
	r := Bin(s, BinCfg{
		Edges:       []float64{0, 5, 10},
		Aggregation: AggSum,
	})
	if r.Len() != 2 {
		t.Fatalf("len = %d, want 2", r.Len())
	}
	assertClose(t, r.Points[0].X, 2.5, "center0")
	assertClose(t, r.Points[1].X, 7.5, "center1")
	// Bin [0,5): points 0,1,2,3,4 -> sum=5
	assertClose(t, r.Points[0].Y, 5, "sum0")
	// Bin [5,10): points 5,6,7,8,9 -> sum=5
	assertClose(t, r.Points[1].Y, 5, "sum1")
}

func TestBinMean(t *testing.T) {
	pts := make([]series.Point, 10)
	for i := range 10 {
		pts[i] = series.Point{X: float64(i), Y: float64(i)}
	}
	s := series.NewXY(series.XYCfg{Name: "data", Points: pts})
	r := Bin(s, BinCfg{
		Edges:       []float64{0, 5, 10},
		Aggregation: AggMean,
	})
	// Bin [0,5): Y=0,1,2,3,4 -> mean=2
	assertClose(t, r.Points[0].Y, 2, "mean0")
	// Bin [5,10): Y=5,6,7,8,9 -> mean=7
	assertClose(t, r.Points[1].Y, 7, "mean1")
}

func TestBinCount(t *testing.T) {
	pts := make([]series.Point, 10)
	for i := range 10 {
		pts[i] = series.Point{X: float64(i), Y: float64(i)}
	}
	s := series.NewXY(series.XYCfg{Name: "data", Points: pts})
	r := Bin(s, BinCfg{
		Edges:       []float64{0, 5, 10},
		Aggregation: AggCount,
	})
	assertClose(t, r.Points[0].Y, 5, "count0")
	assertClose(t, r.Points[1].Y, 5, "count1")
}

func TestBinMinMax(t *testing.T) {
	pts := []series.Point{
		{X: 0, Y: 5}, {X: 1, Y: 3}, {X: 2, Y: 8},
		{X: 5, Y: 1}, {X: 6, Y: 9},
	}
	s := series.NewXY(series.XYCfg{Name: "data", Points: pts})
	rMin := Bin(s, BinCfg{
		Edges:       []float64{0, 5, 10},
		Aggregation: AggMin,
	})
	rMax := Bin(s, BinCfg{
		Edges:       []float64{0, 5, 10},
		Aggregation: AggMax,
	})
	assertClose(t, rMin.Points[0].Y, 3, "min0")
	assertClose(t, rMax.Points[0].Y, 8, "max0")
	assertClose(t, rMin.Points[1].Y, 1, "min1")
	assertClose(t, rMax.Points[1].Y, 9, "max1")
}

func TestBinAutoSturges(t *testing.T) {
	pts := make([]series.Point, 100)
	for i := range 100 {
		pts[i] = series.Point{X: float64(i), Y: 1}
	}
	s := series.NewXY(series.XYCfg{Name: "data", Points: pts})
	r := Bin(s, BinCfg{Aggregation: AggCount})
	// Sturges: ceil(log2(100)+1) = ceil(7.64) = 8
	if r.Len() != 8 {
		t.Errorf("len = %d, want 8", r.Len())
	}
}

func TestBinEmpty(t *testing.T) {
	r := Bin(series.XY{}, BinCfg{Bins: 5})
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestBinAllNaN(t *testing.T) {
	s := series.XYFromYValues("t", []float64{math.NaN(), math.NaN()})
	r := Bin(s, BinCfg{Bins: 3})
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestBinExcessiveBinsCapped(t *testing.T) {
	pts := make([]series.Point, 10)
	for i := range 10 {
		pts[i] = series.Point{X: float64(i), Y: 1}
	}
	s := series.NewXY(series.XYCfg{Name: "data", Points: pts})
	r := Bin(s, BinCfg{Bins: maxBins + 1000})
	if r.Len() > maxBins {
		t.Errorf("len = %d, exceeds maxBins %d", r.Len(), maxBins)
	}
}

func TestBinName(t *testing.T) {
	s := series.XYFromYValues("Readings", []float64{1, 2, 3})
	r := Bin(s, BinCfg{Bins: 2})
	want := "Readings Binned"
	if r.Name() != want {
		t.Errorf("name = %q, want %q", r.Name(), want)
	}
}
