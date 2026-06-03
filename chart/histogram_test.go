package chart

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/axis"
)

// --- calcBins ---

func TestCalcBins_Empty(t *testing.T) {
	edges, counts := calcBins(nil, 0, nil)
	if edges != nil || counts != nil {
		t.Fatal("expected nil for empty input")
	}
}

func TestCalcBins_AllNaNInf(t *testing.T) {
	data := []float64{math.NaN(), math.Inf(1), math.Inf(-1)}
	edges, counts := calcBins(data, 0, nil)
	if edges != nil || counts != nil {
		t.Fatal("expected nil when all values are non-finite")
	}
}

func TestCalcBins_SingleValue(t *testing.T) {
	edges, counts := calcBins([]float64{5.0}, 0, nil)
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(edges))
	}
	if len(counts) != 1 {
		t.Fatalf("expected 1 bin, got %d", len(counts))
	}
	if counts[0] != 1 {
		t.Errorf("expected count=1, got %d", counts[0])
	}
	if edges[0] >= 5.0 || edges[1] <= 5.0 {
		t.Errorf("single-value bin [%g, %g) does not contain 5.0", edges[0], edges[1])
	}
}

func TestCalcBins_AllIdentical(t *testing.T) {
	data := []float64{3.0, 3.0, 3.0, 3.0}
	_, counts := calcBins(data, 0, nil)
	if len(counts) != 1 {
		t.Fatalf("expected 1 bin for identical values, got %d", len(counts))
	}
	if counts[0] != 4 {
		t.Errorf("expected count=4, got %d", counts[0])
	}
}

func TestCalcBins_SturgesRule(t *testing.T) {
	// n=8 → ceil(log2(8)+1) = ceil(3+1) = 4 bins
	data := make([]float64, 8)
	for i := range data {
		data[i] = float64(i)
	}
	_, counts := calcBins(data, 0, nil)
	if len(counts) != 4 {
		t.Fatalf("Sturges: expected 4 bins for n=8, got %d", len(counts))
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	if total != 8 {
		t.Errorf("total count %d != len(data) %d", total, 8)
	}
}

func TestCalcBins_ExplicitBins(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	_, counts := calcBins(data, 2, nil)
	if len(counts) != 2 {
		t.Fatalf("expected 2 bins, got %d", len(counts))
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	if total != 10 {
		t.Errorf("total count %d != 10", total)
	}
}

func TestCalcBins_CustomEdges(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	customEdges := []float64{0, 2.5, 5}
	edges, counts := calcBins(data, 0, customEdges)
	if len(counts) != 2 {
		t.Fatalf("expected 2 bins from custom edges, got %d", len(counts))
	}
	// [0,2.5): 1,2 → count=2; [2.5,5]: 3,4,5 → count=3
	if counts[0] != 2 {
		t.Errorf("bin[0] count: want 2, got %d", counts[0])
	}
	if counts[1] != 3 {
		t.Errorf("bin[1] count: want 3, got %d", counts[1])
	}
	_ = edges
}

func TestCalcBins_LastEdgeClosed(t *testing.T) {
	// Value exactly at upper edge must fall in the last bin.
	data := []float64{0, 5, 10}
	edges, counts := calcBins(data, 2, nil)
	if len(counts) != 2 {
		t.Fatalf("expected 2 bins, got %d", len(counts))
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	if total != 3 {
		t.Errorf("total count %d != 3 (value at upper edge was dropped)", total)
	}
	_ = edges
}

func TestCalcBins_NaNIgnored(t *testing.T) {
	data := []float64{1, math.NaN(), 3, math.Inf(1), 5}
	_, counts := calcBins(data, 0, nil)
	total := 0
	for _, c := range counts {
		total += c
	}
	if total != 3 {
		t.Errorf("expected 3 finite values counted, got %d", total)
	}
}

func TestCalcBins_TotalCount(t *testing.T) {
	data := make([]float64, 100)
	for i := range data {
		data[i] = float64(i)
	}
	_, counts := calcBins(data, 10, nil)
	total := 0
	for _, c := range counts {
		total += c
	}
	if total != 100 {
		t.Errorf("total count %d != 100", total)
	}
}

func TestCalcBins_CustomEdges_OutOfRange(t *testing.T) {
	// Values outside [edges[0], edges[n]] must be silently dropped.
	data := []float64{-1, 0, 1, 2, 3, 4, 5, 6, 99}
	edges, counts := calcBins(data, 0, []float64{0, 2.5, 5})
	if len(counts) != 2 {
		t.Fatalf("expected 2 bins, got %d", len(counts))
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	// -1 (below lo) and 6, 99 (above hi) should be dropped; 0..5 = 6 values kept.
	if total != 6 {
		t.Errorf("expected 6 in-range values, got %d", total)
	}
	_ = edges
}

func TestCalcBins_NormalizedDensity(t *testing.T) {
	// For uniform data in [0,10) with 2 equal-width bins, density should be
	// 0.5/5 = 0.1 per bin (each bin holds half the mass over width 5).
	data := make([]float64, 100)
	for i := range data {
		data[i] = float64(i) / 10 // 0.0, 0.1, ..., 9.9 → all in [0, 10)
	}
	edges, counts := calcBins(data, 2, nil)
	if len(counts) != 2 {
		t.Fatalf("expected 2 bins, got %d", len(counts))
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	if total != 100 {
		t.Fatalf("total count %d != 100", total)
	}
	// Compute density for each bin and verify they sum to ≈1 when multiplied
	// by bin width (integral of density = 1).
	integral := 0.0
	for i, c := range counts {
		w := edges[i+1] - edges[i]
		density := float64(c) / (float64(total) * w)
		integral += density * w
	}
	if math.Abs(integral-1.0) > 1e-9 {
		t.Errorf("density integral = %g, want 1.0", integral)
	}
}

// --- HistogramCfg.Validate ---

func TestHistogramValidate_NegativeBins(t *testing.T) {
	cfg := HistogramCfg{BaseCfg: BaseCfg{ID: "h1"}, Bins: -1}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative Bins")
	}
}

func TestHistogramValidate_NegativeRadius(t *testing.T) {
	cfg := HistogramCfg{BaseCfg: BaseCfg{ID: "h1"}, Radius: -1}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative Radius")
	}
}

func TestHistogramValidate_SingleBinEdge(t *testing.T) {
	cfg := HistogramCfg{
		BaseCfg:  BaseCfg{ID: "h1"},
		BinEdges: []float64{1.0},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for BinEdges with 1 entry")
	}
}

func TestHistogramValidate_Valid(t *testing.T) {
	cfg := HistogramCfg{
		BaseCfg: BaseCfg{ID: "h1"},
		Data:    []float64{1, 2, 3},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- histogramView.hoveredBin ---

func TestHoveredBin_OutsidePlot(t *testing.T) {
	hv := &histogramView{}
	hv.cfg.applyDefaults()
	edges, counts := calcBins([]float64{1, 2, 3, 4, 5}, 4, nil)
	hv.binEdges = edges
	hv.binValues = make([]float64, len(counts))
	hv.xAxis = axis.NewLinear(axis.LinearCfg{Min: edges[0], Max: edges[len(edges)-1]})
	if got := hv.hoveredBin(0, 50, 250); got != -1 {
		t.Errorf("expected -1 for pixel left of plot, got %d", got)
	}
}

func TestHoveredBin_FirstBin(t *testing.T) {
	hv := &histogramView{}
	hv.cfg.applyDefaults()
	edges, counts := calcBins([]float64{0, 1, 2, 3, 4}, 4, nil)
	hv.binEdges = edges
	hv.binValues = make([]float64, len(counts))
	hv.xAxis = axis.NewLinear(axis.LinearCfg{Min: edges[0], Max: edges[len(edges)-1]})
	// Pixel just inside left edge → first bin.
	if got := hv.hoveredBin(51, 50, 250); got != 0 {
		t.Errorf("expected bin 0, got %d", got)
	}
}
