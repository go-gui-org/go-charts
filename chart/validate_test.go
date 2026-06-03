package chart

import (
	"testing"

	"github.com/go-gui-org/go-charts/series"
)

func TestLineCfgValidateOK(t *testing.T) {
	cfg := LineCfg{
		Series: []series.XY{series.XYFromYValues("s", []float64{1})},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLineCfgValidateEmpty(t *testing.T) {
	cfg := LineCfg{}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty series")
	}
}

func TestLineCfgValidateNegativeWidth(t *testing.T) {
	cfg := LineCfg{
		BaseCfg: BaseCfg{Width: -1},
		Series:  []series.XY{series.XYFromYValues("s", []float64{1})},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative Width")
	}
}

func TestBarCfgValidateEmpty(t *testing.T) {
	cfg := BarCfg{}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty series")
	}
}

func TestBarCfgValidateNegativeBarWidth(t *testing.T) {
	cfg := BarCfg{
		Series:   []series.Category{series.CategoryFromMap("s", map[string]float64{"a": 1})},
		BarWidth: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative BarWidth")
	}
}

func TestPieCfgValidateOK(t *testing.T) {
	cfg := PieCfg{
		Slices: []PieSlice{{Label: "a", Value: 1}},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPieCfgValidateEmpty(t *testing.T) {
	cfg := PieCfg{}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty slices")
	}
}

func TestBaseCfgValidateNegativeHeight(t *testing.T) {
	cfg := BaseCfg{Height: -1}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative Height")
	}
}

func TestBarCfgValidateNegativeBarGap(t *testing.T) {
	cfg := BarCfg{
		Series: []series.Category{series.CategoryFromMap("s", map[string]float64{"a": 1})},
		BarGap: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative BarGap")
	}
}

func TestBarCfgValidateNegativeRadius(t *testing.T) {
	cfg := BarCfg{
		Series: []series.Category{series.CategoryFromMap("s", map[string]float64{"a": 1})},
		Radius: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative Radius")
	}
}

func TestBarCfgValidateSeriesLengthMismatch(t *testing.T) {
	cfg := BarCfg{
		Series: []series.Category{
			series.CategoryFromMap("s1", map[string]float64{"a": 1, "b": 2}),
			series.CategoryFromMap("s2", map[string]float64{"a": 1}),
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for series length mismatch")
	}
}

func TestAreaCfgValidateOK(t *testing.T) {
	cfg := AreaCfg{
		Series: []series.XY{series.XYFromYValues("s", []float64{1, 2})},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAreaCfgValidateEmpty(t *testing.T) {
	cfg := AreaCfg{}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty series")
	}
}

func TestAreaCfgValidateNegativeLineWidth(t *testing.T) {
	cfg := AreaCfg{
		Series:    []series.XY{series.XYFromYValues("s", []float64{1})},
		LineWidth: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative LineWidth")
	}
}

func TestAreaCfgValidateOpacityOutOfRange(t *testing.T) {
	for _, op := range []float32{-0.1, 1.1} {
		cfg := AreaCfg{
			Series:  []series.XY{series.XYFromYValues("s", []float64{1})},
			Opacity: op,
		}
		if err := cfg.Validate(); err == nil {
			t.Errorf("expected error for Opacity %v", op)
		}
	}
}

func TestScatterCfgValidateOK(t *testing.T) {
	cfg := ScatterCfg{
		Series: []series.XY{series.XYFromYValues("s", []float64{1, 2})},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestScatterCfgValidateEmpty(t *testing.T) {
	cfg := ScatterCfg{}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty series")
	}
}

func TestScatterCfgValidateNegativeMarkerSize(t *testing.T) {
	cfg := ScatterCfg{
		Series:     []series.XY{series.XYFromYValues("s", []float64{1})},
		MarkerSize: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative MarkerSize")
	}
}

func TestBubbleCfgValidateOK(t *testing.T) {
	cfg := BubbleCfg{
		Series: []series.XYZ{series.NewXYZ(series.XYZCfg{
			Name:   "s",
			Points: []series.XYZPoint{{X: 1, Y: 2, Z: 3}},
		})},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBubbleCfgValidateEmpty(t *testing.T) {
	cfg := BubbleCfg{}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty series")
	}
}

func TestBubbleCfgValidateNegativeMinRadius(t *testing.T) {
	cfg := BubbleCfg{
		Series: []series.XYZ{series.NewXYZ(series.XYZCfg{
			Name:   "s",
			Points: []series.XYZPoint{{X: 1, Y: 2, Z: 3}},
		})},
		MinRadius: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative MinRadius")
	}
}

func TestBubbleCfgValidateMinExceedsMax(t *testing.T) {
	cfg := BubbleCfg{
		Series: []series.XYZ{series.NewXYZ(series.XYZCfg{
			Name:   "s",
			Points: []series.XYZPoint{{X: 1, Y: 2, Z: 3}},
		})},
		MinRadius: 30,
		MaxRadius: 10,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for MinRadius > MaxRadius")
	}
}

func TestPieCfgValidateNegativeInnerRadius(t *testing.T) {
	cfg := PieCfg{
		Slices:      []PieSlice{{Label: "a", Value: 1}},
		InnerRadius: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative InnerRadius")
	}
}
