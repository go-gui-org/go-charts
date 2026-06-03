package chart

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func TestHeatmapColor(t *testing.T) {
	low := gui.RGBA(0, 0, 255, 255)
	high := gui.RGBA(255, 0, 0, 255)

	tests := []struct {
		name          string
		v, vMin, vMax float64
		wantR, wantB  uint8
	}{
		{"min value", 0, 0, 100, 0, 255},
		{"max value", 100, 0, 100, 255, 0},
		{"midpoint", 50, 0, 100, 127, 127},
		{"equal range", 5, 5, 5, 127, 127},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := heatmapColor(tt.v, tt.vMin, tt.vMax, low, high)
			if got.R != tt.wantR || got.B != tt.wantB {
				t.Errorf("heatmapColor(%g, %g, %g) = R:%d B:%d, "+
					"want R:%d B:%d",
					tt.v, tt.vMin, tt.vMax,
					got.R, got.B, tt.wantR, tt.wantB)
			}
		})
	}
}

func TestHeatmapHitTest(t *testing.T) {
	g, _ := series.NewGrid(series.GridCfg{
		Rows:   []string{"A", "B"},
		Cols:   []string{"X", "Y", "Z"},
		Values: [][]float64{{1, 2, 3}, {4, 5, 6}},
	})
	hv := &heatmapView{
		cfg:        HeatmapCfg{Data: g},
		lastLeft:   10,
		lastRight:  70,
		lastTop:    10,
		lastBottom: 50,
	}

	// Cell (0,0): x=[10,30), y=[10,30)
	row, col, ok := hv.hitTest(15, 15)
	if !ok || row != 0 || col != 0 {
		t.Errorf("hitTest(15,15) = (%d,%d,%v), want (0,0,true)",
			row, col, ok)
	}

	// Cell (1,2): x=[50,70), y=[30,50)
	row, col, ok = hv.hitTest(55, 35)
	if !ok || row != 1 || col != 2 {
		t.Errorf("hitTest(55,35) = (%d,%d,%v), want (1,2,true)",
			row, col, ok)
	}

	// Outside plot area.
	_, _, ok = hv.hitTest(5, 15)
	if ok {
		t.Error("hitTest outside left should return false")
	}
}

func TestHeatmapValidate(t *testing.T) {
	cfg := HeatmapCfg{
		Data: series.Grid{},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty grid")
	}

	g, _ := series.NewGrid(series.GridCfg{
		Rows:   []string{"A"},
		Cols:   []string{"X"},
		Values: [][]float64{{1}},
	})
	cfg = HeatmapCfg{Data: g, CellGap: -1}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative CellGap")
	}

	cfg = HeatmapCfg{Data: g}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHeatmapColorNaN(t *testing.T) {
	// Ensure NaN does not panic in heatmapColor.
	low := gui.Hex(0x0000FF)
	high := gui.Hex(0xFF0000)
	got := heatmapColor(math.NaN(), 0, 100, low, high)
	// Result is undefined but should not panic. Lerp clamps t.
	_ = got
}
