package transform

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/series"
)

func TestLTTBBasic(t *testing.T) {
	pts := make([]series.Point, 100)
	for i := range 100 {
		pts[i] = series.Point{X: float64(i), Y: float64(i * i)}
	}
	s := series.NewXY(series.XYCfg{Name: "data", Points: pts})
	r := LTTB(s, 20)
	if r.Len() != 20 {
		t.Fatalf("len = %d, want 20", r.Len())
	}
	// First and last always preserved.
	assertClose(t, r.Points[0].X, 0, "first.X")
	assertClose(t, r.Points[19].X, 99, "last.X")
}

func TestLTTBThresholdExceedsLen(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2, 3})
	r := LTTB(s, 10)
	if r.Len() != 3 {
		t.Fatalf("len = %d, want 3", r.Len())
	}
}

func TestLTTBThresholdLessThan3(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2, 3, 4, 5})
	r := LTTB(s, 2)
	if r.Len() != 5 {
		t.Fatalf("len = %d, want 5 (returns copy)", r.Len())
	}
}

func TestLTTBEmpty(t *testing.T) {
	r := LTTB(series.XY{}, 10)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestLTTBExactThreshold(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2, 3, 4, 5})
	r := LTTB(s, 5)
	if r.Len() != 5 {
		t.Fatalf("len = %d, want 5", r.Len())
	}
}

func TestLTTBPreservesExtremes(t *testing.T) {
	// Create data with a clear spike; LTTB should keep it.
	pts := make([]series.Point, 50)
	for i := range 50 {
		pts[i] = series.Point{X: float64(i), Y: 0}
	}
	pts[25].Y = 100 // spike
	s := series.NewXY(series.XYCfg{Name: "spike", Points: pts})
	r := LTTB(s, 10)

	found := false
	for _, p := range r.Points {
		if p.Y == 100 {
			found = true
			break
		}
	}
	if !found {
		t.Error("LTTB did not preserve spike at index 25")
	}
}

func TestLTTBWithNaN(t *testing.T) {
	pts := make([]series.Point, 20)
	for i := range 20 {
		pts[i] = series.Point{X: float64(i), Y: float64(i)}
	}
	pts[5].Y = math.NaN()
	pts[10].Y = math.Inf(1)
	s := series.NewXY(series.XYCfg{Name: "dirty", Points: pts})
	r := LTTB(s, 10)
	// Non-finite points filtered: 18 finite points, downsampled to 10.
	if r.Len() != 10 {
		t.Fatalf("len = %d, want 10", r.Len())
	}
	for i, p := range r.Points {
		if math.IsNaN(p.Y) || math.IsInf(p.Y, 0) {
			t.Errorf("pts[%d] is non-finite: %g", i, p.Y)
		}
	}
}

func TestLTTBName(t *testing.T) {
	s := series.XYFromYValues("Sensor", []float64{1, 2, 3})
	r := LTTB(s, 10)
	want := "Sensor (downsampled)"
	if r.Name() != want {
		t.Errorf("name = %q, want %q", r.Name(), want)
	}
}
