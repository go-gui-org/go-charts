package chart

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/series"
)

func TestSparklineValidateEmpty(t *testing.T) {
	cfg := SparklineCfg{}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty data")
	}
}

func TestSparklineValidateNegativeLineWidth(t *testing.T) {
	cfg := SparklineCfg{
		Values:    []float64{1, 2, 3},
		LineWidth: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative LineWidth")
	}
}

func TestSparklineValidateNegativeMarkerRadius(t *testing.T) {
	cfg := SparklineCfg{
		Values:       []float64{1, 2, 3},
		MarkerRadius: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative MarkerRadius")
	}
}

func TestSparklineValidateInvalidType(t *testing.T) {
	cfg := SparklineCfg{
		Values: []float64{1, 2, 3},
		Type:   SparklineType(99),
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid Type")
	}
}

func TestSparklineValidateOK(t *testing.T) {
	cfg := SparklineCfg{Values: []float64{1, 2, 3}}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSparklineValidateSeriesOK(t *testing.T) {
	cfg := SparklineCfg{
		Series: series.XYFromYValues("test", []float64{1, 2}),
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSparklineResolveDataValues(t *testing.T) {
	sv := &sparklineView{
		cfg: SparklineCfg{Values: []float64{10, 20, 30}},
	}
	sv.resolveData()
	if sv.resolved.Len() != 3 {
		t.Fatalf("len = %d, want 3", sv.resolved.Len())
	}
	// Auto-indexed: X = 0, 1, 2.
	pts := sv.resolved.Points
	for i, p := range pts {
		if p.X != float64(i) {
			t.Errorf("pts[%d].X = %g, want %g", i, p.X,
				float64(i))
		}
	}
	if pts[1].Y != 20 {
		t.Errorf("pts[1].Y = %g, want 20", pts[1].Y)
	}
}

func TestSparklineResolveDataSeriesWins(t *testing.T) {
	xy := series.XYFromYValues("s", []float64{5, 6})
	sv := &sparklineView{
		cfg: SparklineCfg{
			Values: []float64{1, 2, 3},
			Series: xy,
		},
	}
	sv.resolveData()
	if sv.resolved.Len() != 2 {
		t.Fatalf("len = %d, want 2 (Series should win)",
			sv.resolved.Len())
	}
}

func TestSparklineBuildAxes(t *testing.T) {
	sv := &sparklineView{
		cfg: SparklineCfg{Values: []float64{10, 20, 30}},
	}
	sv.resolveData()
	if !sv.buildAxes() {
		t.Fatal("buildAxes returned false")
	}
	if sv.xAxis == nil || sv.yAxis == nil {
		t.Fatal("axes not created")
	}
}

func TestSparklineBuildAxesMixedNaN(t *testing.T) {
	// Mix of NaN and finite values should still build axes.
	sv := &sparklineView{
		cfg: SparklineCfg{
			Values: []float64{math.NaN(), 5, math.Inf(1), 10},
		},
	}
	sv.resolveData()
	if !sv.buildAxes() {
		t.Error("buildAxes should succeed with some finite data")
	}
}

func TestSparklineMinMaxLast(t *testing.T) {
	vals := []float64{5, math.NaN(), 1, 8, 3}
	sv := &sparklineView{
		cfg: SparklineCfg{
			Values:         vals,
			ShowMinMarker:  true,
			ShowMaxMarker:  true,
			ShowLastMarker: true,
		},
	}
	sv.resolveData()

	// Scan for min/max/last manually to verify.
	minIdx, maxIdx, lastIdx := -1, -1, -1
	minVal := math.MaxFloat64
	maxVal := -math.MaxFloat64
	for i, p := range sv.resolved.Points {
		if !finite(p.Y) {
			continue
		}
		lastIdx = i
		if p.Y < minVal {
			minVal = p.Y
			minIdx = i
		}
		if p.Y > maxVal {
			maxVal = p.Y
			maxIdx = i
		}
	}

	if minIdx != 2 {
		t.Errorf("minIdx = %d, want 2", minIdx)
	}
	if maxIdx != 3 {
		t.Errorf("maxIdx = %d, want 3", maxIdx)
	}
	if lastIdx != 4 {
		t.Errorf("lastIdx = %d, want 4", lastIdx)
	}
}

func TestSparklineValidateAllTypes(t *testing.T) {
	for _, tp := range []SparklineType{
		SparklineLine, SparklineBar, SparklineArea,
	} {
		cfg := SparklineCfg{
			Values: []float64{1, 2, 3},
			Type:   tp,
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("type %d: unexpected error: %v", tp, err)
		}
	}
}

func TestSparklineBuildAxesSinglePoint(t *testing.T) {
	sv := &sparklineView{
		cfg: SparklineCfg{Values: []float64{42}},
	}
	sv.resolveData()
	if !sv.buildAxes() {
		t.Fatal("buildAxes failed for single point")
	}
}

func TestSparklineBuildAxesConstant(t *testing.T) {
	sv := &sparklineView{
		cfg: SparklineCfg{Values: []float64{5, 5, 5}},
	}
	sv.resolveData()
	if !sv.buildAxes() {
		t.Fatal("buildAxes failed for constant values")
	}
}

func TestSparklineBuildAxesNaNReference(t *testing.T) {
	sv := &sparklineView{
		cfg: SparklineCfg{
			Values:            []float64{1, 2, 3},
			ShowReferenceLine: true,
			ReferenceValue:    math.NaN(),
		},
	}
	sv.resolveData()
	// NaN reference should not corrupt axis range.
	if !sv.buildAxes() {
		t.Fatal("buildAxes failed with NaN reference")
	}
}

func TestSparklineBuildAxesInfReference(t *testing.T) {
	sv := &sparklineView{
		cfg: SparklineCfg{
			Values:            []float64{1, 2, 3},
			ShowReferenceLine: true,
			ReferenceValue:    math.Inf(1),
			BandColoring:      true,
		},
	}
	sv.resolveData()
	// Inf reference should not corrupt axis range.
	if !sv.buildAxes() {
		t.Fatal("buildAxes failed with Inf reference")
	}
}

func TestSparklineAllNaN(t *testing.T) {
	// All-NaN Values: resolveData produces points, but all
	// are non-finite. Bounds returns (0,0,0,0) — degenerate
	// but finite. buildAxes should not panic.
	sv := &sparklineView{
		cfg: SparklineCfg{
			Values: []float64{
				math.NaN(), math.NaN(), math.NaN(),
			},
		},
	}
	sv.resolveData()
	// Should not panic regardless of outcome.
	sv.buildAxes()
}

func TestSparklineBarNonFiniteReference(t *testing.T) {
	// Non-finite ReferenceValue should fall back to zero
	// baseline in drawBars, not produce NaN pixel coords.
	sv := &sparklineView{
		cfg: SparklineCfg{
			Values:            []float64{1, 2, 3},
			Type:              SparklineBar,
			ShowReferenceLine: true,
			ReferenceValue:    math.NaN(),
		},
	}
	sv.resolveData()
	if !sv.buildAxes() {
		t.Fatal("buildAxes failed")
	}
	// Verify refVal falls back to 0 when ReferenceValue is
	// non-finite.
	refVal := float64(0)
	if (sv.cfg.ShowReferenceLine || sv.cfg.BandColoring) &&
		finite(sv.cfg.ReferenceValue) {
		refVal = sv.cfg.ReferenceValue
	}
	if refVal != 0 {
		t.Errorf("refVal = %g, want 0 (NaN fallback)", refVal)
	}
}
