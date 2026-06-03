package transform

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/series"
)

const eps = 1e-9

func assertClose(t *testing.T, got, want float64, msg string) {
	t.Helper()
	if math.IsNaN(want) {
		if !math.IsNaN(got) {
			t.Errorf("%s: got %g, want NaN", msg, got)
		}
		return
	}
	if math.Abs(got-want) > eps {
		t.Errorf("%s: got %g, want %g", msg, got, want)
	}
}

func TestSMABasic(t *testing.T) {
	s := series.XYFromYValues("test", []float64{1, 2, 3, 4, 5})
	r := SMA(s, 3)
	if r.Len() != 5 {
		t.Fatalf("len = %d, want 5", r.Len())
	}
	assertClose(t, r.Points[0].Y, math.NaN(), "pts[0]")
	assertClose(t, r.Points[1].Y, math.NaN(), "pts[1]")
	assertClose(t, r.Points[2].Y, 2, "pts[2]")
	assertClose(t, r.Points[3].Y, 3, "pts[3]")
	assertClose(t, r.Points[4].Y, 4, "pts[4]")
}

func TestSMAEmpty(t *testing.T) {
	r := SMA(series.XY{}, 3)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestSMAInvalidWindow(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2})
	r := SMA(s, 0)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
	r = SMA(s, -1)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestSMAWithNaN(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, math.NaN(), 3, 4, 5})
	r := SMA(s, 3)
	// Position 2: only 2 finite values (1, 3) -> NaN
	assertClose(t, r.Points[2].Y, math.NaN(), "pts[2]")
	// Position 4: 3, 4, 5 -> finite
	assertClose(t, r.Points[4].Y, 4, "pts[4]")
}

func TestSMAWindowOne(t *testing.T) {
	s := series.XYFromYValues("t", []float64{3, 7, 2})
	r := SMA(s, 1)
	for i, p := range r.Points {
		assertClose(t, p.Y, s.Points[i].Y, "pts")
	}
}

func TestEMABasic(t *testing.T) {
	s := series.XYFromYValues("test", []float64{1, 2, 3, 4, 5})
	r := EMA(s, 3) // alpha = 0.5
	if r.Len() != 5 {
		t.Fatalf("len = %d, want 5", r.Len())
	}
	assertClose(t, r.Points[0].Y, 1, "pts[0]")
	assertClose(t, r.Points[1].Y, 1.5, "pts[1]")
	assertClose(t, r.Points[2].Y, 2.25, "pts[2]")
	assertClose(t, r.Points[3].Y, 3.125, "pts[3]")
	assertClose(t, r.Points[4].Y, 4.0625, "pts[4]")
}

func TestEMAEmpty(t *testing.T) {
	r := EMA(series.XY{}, 3)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestEMAInvalidWindow(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1})
	r := EMA(s, 0)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestEMAWithNaN(t *testing.T) {
	s := series.XYFromYValues("t", []float64{math.NaN(), 2, 3})
	r := EMA(s, 3)
	assertClose(t, r.Points[0].Y, math.NaN(), "pts[0]")
	assertClose(t, r.Points[1].Y, 2, "pts[1]") // first finite seeds
}

func TestWMABasic(t *testing.T) {
	s := series.XYFromYValues("test", []float64{1, 2, 3, 4, 5})
	r := WMA(s, 3)
	if r.Len() != 5 {
		t.Fatalf("len = %d, want 5", r.Len())
	}
	assertClose(t, r.Points[0].Y, math.NaN(), "pts[0]")
	assertClose(t, r.Points[1].Y, math.NaN(), "pts[1]")
	// (1*1 + 2*2 + 3*3) / 6 = 14/6 = 2.333...
	assertClose(t, r.Points[2].Y, 14.0/6.0, "pts[2]")
	// (1*2 + 2*3 + 3*4) / 6 = 20/6 = 3.333...
	assertClose(t, r.Points[3].Y, 20.0/6.0, "pts[3]")
	// (1*3 + 2*4 + 3*5) / 6 = 26/6 = 4.333...
	assertClose(t, r.Points[4].Y, 26.0/6.0, "pts[4]")
}

func TestWMAEmpty(t *testing.T) {
	r := WMA(series.XY{}, 3)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestWMAInvalidWindow(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2})
	r := WMA(s, 0)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestWMAWithNaN(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, math.NaN(), 3, 4, 5})
	r := WMA(s, 3)
	// Position 2: window [1, NaN, 3] -> NaN
	assertClose(t, r.Points[2].Y, math.NaN(), "pts[2]")
	// Position 3: window [NaN, 3, 4] -> NaN
	assertClose(t, r.Points[3].Y, math.NaN(), "pts[3]")
	// Position 4: window [3, 4, 5] -> (3+8+15)/6 = 26/6
	assertClose(t, r.Points[4].Y, 26.0/6.0, "pts[4]")
}

func TestSMAWindowExceedsMax(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2, 3})
	r := SMA(s, maxWindow+1)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestEMAWindowExceedsMax(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2, 3})
	r := EMA(s, maxWindow+1)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestWMAWindowExceedsMax(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2, 3})
	r := WMA(s, maxWindow+1)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestSMAName(t *testing.T) {
	s := series.XYFromYValues("Revenue", []float64{1, 2, 3})
	r := SMA(s, 2)
	want := "Revenue SMA(2)"
	if r.Name() != want {
		t.Errorf("name = %q, want %q", r.Name(), want)
	}
}
