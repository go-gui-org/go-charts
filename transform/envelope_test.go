package transform

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/series"
)

func TestBollingerBandsConstant(t *testing.T) {
	// Constant series: upper == middle == lower.
	s := series.XYFromYValues("c", []float64{5, 5, 5, 5, 5})
	upper, middle, lower := BollingerBands(s, 3, 2)
	for i := 2; i < 5; i++ {
		assertClose(t, upper.Points[i].Y, 5, "upper")
		assertClose(t, middle.Points[i].Y, 5, "middle")
		assertClose(t, lower.Points[i].Y, 5, "lower")
	}
}

func TestBollingerBandsNaN(t *testing.T) {
	s := series.XYFromYValues("c", []float64{5, 5, 5})
	upper, _, lower := BollingerBands(s, 3, 2)
	assertClose(t, upper.Points[0].Y, math.NaN(), "upper[0]")
	assertClose(t, lower.Points[0].Y, math.NaN(), "lower[0]")
}

func TestBollingerBandsEmpty(t *testing.T) {
	upper, middle, lower := BollingerBands(series.XY{}, 3, 2)
	if upper.Len() != 0 || middle.Len() != 0 || lower.Len() != 0 {
		t.Error("expected empty")
	}
}

func TestBollingerBandsInvalidWindow(t *testing.T) {
	s := series.XYFromYValues("c", []float64{1, 2, 3})
	upper, _, _ := BollingerBands(s, 0, 2)
	if upper.Len() != 0 {
		t.Error("expected empty for window=0")
	}
}

func TestBollingerBandsSymmetry(t *testing.T) {
	s := series.XYFromYValues("s", []float64{1, 3, 2, 4, 3})
	upper, middle, lower := BollingerBands(s, 3, 1)
	for i := 2; i < 5; i++ {
		if math.IsNaN(upper.Points[i].Y) {
			continue
		}
		uDiff := upper.Points[i].Y - middle.Points[i].Y
		lDiff := middle.Points[i].Y - lower.Points[i].Y
		assertClose(t, uDiff, lDiff, "symmetry")
	}
}

func TestBollingerBandsNegativeMult(t *testing.T) {
	s := series.XYFromYValues("c", []float64{1, 2, 3})
	upper, _, _ := BollingerBands(s, 2, -1)
	if upper.Len() != 0 {
		t.Error("expected empty for negative multiplier")
	}
}

func TestBollingerBandsName(t *testing.T) {
	s := series.XYFromYValues("Price", []float64{1, 2, 3})
	upper, middle, lower := BollingerBands(s, 2, 2)
	if upper.Name() != "Price BB Upper" {
		t.Errorf("upper name = %q", upper.Name())
	}
	if middle.Name() != "Price BB Middle" {
		t.Errorf("middle name = %q", middle.Name())
	}
	if lower.Name() != "Price BB Lower" {
		t.Errorf("lower name = %q", lower.Name())
	}
}

func TestMinMaxEnvelopeBasic(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 3, 2, 5, 4})
	mx, mn := MinMaxEnvelope(s, 3)
	if mx.Len() != 5 {
		t.Fatalf("len = %d, want 5", mx.Len())
	}
	assertClose(t, mx.Points[0].Y, math.NaN(), "max[0]")
	assertClose(t, mx.Points[1].Y, math.NaN(), "max[1]")
	assertClose(t, mx.Points[2].Y, 3, "max[2]")
	assertClose(t, mx.Points[3].Y, 5, "max[3]")
	assertClose(t, mx.Points[4].Y, 5, "max[4]")

	assertClose(t, mn.Points[2].Y, 1, "min[2]")
	assertClose(t, mn.Points[3].Y, 2, "min[3]")
	assertClose(t, mn.Points[4].Y, 2, "min[4]")
}

func TestMinMaxEnvelopeEmpty(t *testing.T) {
	mx, mn := MinMaxEnvelope(series.XY{}, 3)
	if mx.Len() != 0 || mn.Len() != 0 {
		t.Error("expected empty")
	}
}

func TestMinMaxEnvelopeInvalidWindow(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2})
	mx, _ := MinMaxEnvelope(s, 0)
	if mx.Len() != 0 {
		t.Error("expected empty for window=0")
	}
}

func TestMinMaxEnvelopeWithNaN(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, math.NaN(), 3, 4})
	mx, mn := MinMaxEnvelope(s, 3)
	// Position 2: window has NaN -> NaN
	assertClose(t, mx.Points[2].Y, math.NaN(), "max[2]")
	assertClose(t, mn.Points[2].Y, math.NaN(), "min[2]")
}

func TestBollingerBandsWindowExceedsMax(t *testing.T) {
	s := series.XYFromYValues("c", []float64{1, 2, 3})
	upper, _, _ := BollingerBands(s, maxWindow+1, 2)
	if upper.Len() != 0 {
		t.Error("expected empty for oversized window")
	}
}

func TestMinMaxEnvelopeWindowExceedsMax(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2, 3})
	mx, _ := MinMaxEnvelope(s, maxWindow+1)
	if mx.Len() != 0 {
		t.Error("expected empty for oversized window")
	}
}

func TestMinMaxEnvelopeWindowOne(t *testing.T) {
	s := series.XYFromYValues("t", []float64{3, 1, 4})
	mx, mn := MinMaxEnvelope(s, 1)
	for i, p := range s.Points {
		assertClose(t, mx.Points[i].Y, p.Y, "max")
		assertClose(t, mn.Points[i].Y, p.Y, "min")
	}
}
