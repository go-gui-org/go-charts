package transform

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/series"
)

func TestLinearRegressionPerfect(t *testing.T) {
	// y = 2x + 1
	s := series.NewXY(series.XYCfg{
		Name: "test",
		Points: []series.Point{
			{X: 0, Y: 1}, {X: 1, Y: 3}, {X: 2, Y: 5},
		},
	})
	r := LinearRegression(s)
	if r.Len() != 2 {
		t.Fatalf("len = %d, want 2", r.Len())
	}
	assertClose(t, r.Points[0].X, 0, "x0")
	assertClose(t, r.Points[0].Y, 1, "y0")
	assertClose(t, r.Points[1].X, 2, "x1")
	assertClose(t, r.Points[1].Y, 5, "y1")
}

func TestLinearRegressionIdentity(t *testing.T) {
	s := series.NewXY(series.XYCfg{
		Name: "id",
		Points: []series.Point{
			{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2},
		},
	})
	r := LinearRegression(s)
	assertClose(t, r.Points[0].Y, 0, "y0")
	assertClose(t, r.Points[1].Y, 2, "y1")
}

func TestLinearRegressionTooFew(t *testing.T) {
	s := series.XYFromYValues("t", []float64{5})
	r := LinearRegression(s)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestLinearRegressionEmpty(t *testing.T) {
	r := LinearRegression(series.XY{})
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestLinearRegressionAllNaN(t *testing.T) {
	s := series.XYFromYValues("t", []float64{
		math.NaN(), math.NaN(),
	})
	r := LinearRegression(s)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestLinearTrendBasic(t *testing.T) {
	s := series.NewXY(series.XYCfg{
		Name: "test",
		Points: []series.Point{
			{X: 0, Y: 1}, {X: 1, Y: 3}, {X: 2, Y: 5},
		},
	})
	r := LinearTrend(s)
	if r.Len() != 3 {
		t.Fatalf("len = %d, want 3", r.Len())
	}
	assertClose(t, r.Points[0].Y, 1, "y0")
	assertClose(t, r.Points[1].Y, 3, "y1")
	assertClose(t, r.Points[2].Y, 5, "y2")
}

func TestLinearTrendName(t *testing.T) {
	s := series.XYFromYValues("Sales", []float64{1, 2, 3})
	r := LinearTrend(s)
	want := "Sales Linear Trend"
	if r.Name() != want {
		t.Errorf("name = %q, want %q", r.Name(), want)
	}
}

func TestPolynomialRegressionQuadratic(t *testing.T) {
	// y = x^2
	pts := make([]series.Point, 11)
	for i := range 11 {
		x := float64(i)
		pts[i] = series.Point{X: x, Y: x * x}
	}
	s := series.NewXY(series.XYCfg{Name: "quad", Points: pts})
	r := PolynomialRegression(s, 2, 11)
	if r.Len() != 11 {
		t.Fatalf("len = %d, want 11", r.Len())
	}
	for i, p := range r.Points {
		want := p.X * p.X
		if math.Abs(p.Y-want) > 1e-6 {
			t.Errorf("pts[%d]: got %g, want %g", i, p.Y, want)
		}
	}
}

func TestPolynomialRegressionCubic(t *testing.T) {
	// y = x^3 - 2x
	pts := make([]series.Point, 21)
	for i := range 21 {
		x := float64(i) - 10
		pts[i] = series.Point{X: x, Y: x*x*x - 2*x}
	}
	s := series.NewXY(series.XYCfg{Name: "cubic", Points: pts})
	r := PolynomialRegression(s, 3, 21)
	for i, p := range r.Points {
		want := p.X*p.X*p.X - 2*p.X
		if math.Abs(p.Y-want) > 1e-4 {
			t.Errorf("pts[%d]: got %g, want %g", i, p.Y, want)
		}
	}
}

func TestPolynomialRegressionInvalidDegree(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 2})
	r := PolynomialRegression(s, 0, 10)
	if r.Len() != 0 {
		t.Errorf("degree=0: len = %d, want 0", r.Len())
	}
	r = PolynomialRegression(s, 3, 10) // degree >= n
	if r.Len() != 0 {
		t.Errorf("degree>=n: len = %d, want 0", r.Len())
	}
}

func TestPolynomialRegressionEmpty(t *testing.T) {
	r := PolynomialRegression(series.XY{}, 2, 10)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestPolynomialRegressionDegreeExceedsMax(t *testing.T) {
	s := series.XYFromYValues("t", make([]float64, 100))
	r := PolynomialRegression(s, maxDegree+1, 10)
	if r.Len() != 0 {
		t.Errorf("len = %d, want 0", r.Len())
	}
}

func TestPolynomialRegressionNPointsCapped(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 4, 9, 16, 25})
	r := PolynomialRegression(s, 2, maxOutputPoints+100)
	if r.Len() != maxOutputPoints {
		t.Errorf("len = %d, want %d", r.Len(), maxOutputPoints)
	}
}

func TestPolynomialRegressionDefaultNPoints(t *testing.T) {
	s := series.XYFromYValues("t", []float64{1, 4, 9, 16, 25})
	r := PolynomialRegression(s, 2, 0)
	if r.Len() != 5 {
		t.Errorf("len = %d, want 5", r.Len())
	}
}
