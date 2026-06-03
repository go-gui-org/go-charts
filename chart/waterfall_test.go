package chart

import (
	"path/filepath"
	"testing"

	"github.com/go-gui-org/go-gui/gui"
)

func TestWaterfallBuildBars(t *testing.T) {
	wv := &waterfallView{cfg: WaterfallCfg{
		Values: []WaterfallValue{
			{Label: "Revenue", Value: 5000},
			{Label: "COGS", Value: -2000},
			{Label: "Gross", IsTotal: true},
			{Label: "OpEx", Value: -800},
			{Label: "Net", IsTotal: true},
		},
	}}
	wv.buildBars(&wv.cfg)

	if len(wv.bars) != 5 {
		t.Fatalf("expected 5 bars, got %d", len(wv.bars))
	}

	// Revenue: 0 → 5000, up
	assertBar(t, wv.bars[0], 0, 5000, 5000, waterfallUp)
	// COGS: 5000 → 3000, down
	assertBar(t, wv.bars[1], 3000, 5000, 3000, waterfallDown)
	// Gross Profit: 0 → 3000, total
	assertBar(t, wv.bars[2], 0, 3000, 3000, waterfallTotal)
	// OpEx: 3000 → 2200, down
	assertBar(t, wv.bars[3], 2200, 3000, 2200, waterfallDown)
	// Net: 0 → 2200, total
	assertBar(t, wv.bars[4], 0, 2200, 2200, waterfallTotal)
}

func TestWaterfallBuildBarsAllNegative(t *testing.T) {
	wv := &waterfallView{cfg: WaterfallCfg{
		Values: []WaterfallValue{
			{Label: "Loss A", Value: -100},
			{Label: "Loss B", Value: -200},
			{Label: "Total", IsTotal: true},
		},
	}}
	wv.buildBars(&wv.cfg)

	assertBar(t, wv.bars[0], -100, 0, -100, waterfallDown)
	assertBar(t, wv.bars[1], -300, -100, -300, waterfallDown)
	assertBar(t, wv.bars[2], 0, -300, -300, waterfallTotal)

	if wv.yAxis == nil {
		t.Fatal("yAxis not created")
	}
}

func TestWaterfallBuildBarsEmpty(t *testing.T) {
	wv := &waterfallView{cfg: WaterfallCfg{}}
	wv.buildBars(&wv.cfg)
	if len(wv.bars) != 0 {
		t.Fatalf("expected 0 bars, got %d", len(wv.bars))
	}
}

func TestWaterfallBuildBarsTotalOverride(t *testing.T) {
	wv := &waterfallView{cfg: WaterfallCfg{
		Values: []WaterfallValue{
			{Label: "Opening", Value: 10000, IsTotal: true},
			{Label: "Sales", Value: 3000},
			{Label: "Closing", IsTotal: true},
		},
	}}
	wv.buildBars(&wv.cfg)

	// Opening: override running total to 10000.
	assertBar(t, wv.bars[0], 0, 10000, 10000, waterfallTotal)
	// Sales: 10000 → 13000, up.
	assertBar(t, wv.bars[1], 10000, 13000, 13000, waterfallUp)
	// Closing: 0 → 13000, total.
	assertBar(t, wv.bars[2], 0, 13000, 13000, waterfallTotal)
}

func TestWaterfallHoveredBar(t *testing.T) {
	wv := &waterfallView{cfg: WaterfallCfg{
		Values: []WaterfallValue{
			{Label: "A", Value: 10},
			{Label: "B", Value: 20},
			{Label: "C", Value: 30},
		},
	}}

	// 3 bars, plot area 0–300.
	if got := wv.hoveredBar(50, 0, 300); got != 0 {
		t.Errorf("expected slot 0, got %d", got)
	}
	if got := wv.hoveredBar(150, 0, 300); got != 1 {
		t.Errorf("expected slot 1, got %d", got)
	}
	if got := wv.hoveredBar(250, 0, 300); got != 2 {
		t.Errorf("expected slot 2, got %d", got)
	}
	// Outside.
	if got := wv.hoveredBar(-1, 0, 300); got != -1 {
		t.Errorf("expected -1, got %d", got)
	}
	if got := wv.hoveredBar(301, 0, 300); got != -1 {
		t.Errorf("expected -1, got %d", got)
	}
}

func TestWaterfallHoveredBarEmpty(t *testing.T) {
	wv := &waterfallView{cfg: WaterfallCfg{}}
	if got := wv.hoveredBar(50, 0, 300); got != -1 {
		t.Errorf("expected -1, got %d", got)
	}
}

func TestWaterfallSingleTotal(t *testing.T) {
	wv := &waterfallView{cfg: WaterfallCfg{
		Values: []WaterfallValue{
			{Label: "Start", IsTotal: true},
		},
	}}
	wv.buildBars(&wv.cfg)
	// Running total is 0, so total bar spans 0–0.
	assertBar(t, wv.bars[0], 0, 0, 0, waterfallTotal)
}

func TestWaterfallResolveColor(t *testing.T) {
	// Default colors.
	wv := &waterfallView{cfg: WaterfallCfg{}}
	if got := wv.resolveColor(waterfallUp); got != defaultUpColor {
		t.Errorf("up: got %v, want %v", got, defaultUpColor)
	}
	if got := wv.resolveColor(waterfallDown); got != defaultDownColor {
		t.Errorf("down: got %v, want %v", got, defaultDownColor)
	}
	if got := wv.resolveColor(waterfallTotal); got != defaultTotalColor {
		t.Errorf("total: got %v, want %v", got, defaultTotalColor)
	}

	// Custom colors.
	custom := gui.Hex(0xFF00FF)
	wv2 := &waterfallView{cfg: WaterfallCfg{
		UpColor:    custom,
		DownColor:  custom,
		TotalColor: custom,
	}}
	if got := wv2.resolveColor(waterfallUp); got != custom {
		t.Errorf("custom up: got %v, want %v", got, custom)
	}
	if got := wv2.resolveColor(waterfallDown); got != custom {
		t.Errorf("custom down: got %v, want %v", got, custom)
	}
	if got := wv2.resolveColor(waterfallTotal); got != custom {
		t.Errorf("custom total: got %v, want %v", got, custom)
	}
}

func TestWaterfallValidate(t *testing.T) {
	// No values.
	cfg := WaterfallCfg{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for no values")
	}

	// Negative BarWidth.
	cfg = WaterfallCfg{
		Values:   []WaterfallValue{{Label: "a", Value: 1}},
		BarWidth: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative BarWidth")
	}

	// Negative Radius.
	cfg = WaterfallCfg{
		Values: []WaterfallValue{{Label: "a", Value: 1}},
		Radius: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative Radius")
	}

	// Valid.
	cfg = WaterfallCfg{
		Values: []WaterfallValue{{Label: "a", Value: 1}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExportPNG_Waterfall(t *testing.T) {
	v := Waterfall(WaterfallCfg{
		BaseCfg: BaseCfg{
			ID:    "test-waterfall",
			Width: 400, Height: 300,
		},
		Values: []WaterfallValue{
			{Label: "Revenue", Value: 5000},
			{Label: "COGS", Value: -2000},
			{Label: "Gross", IsTotal: true},
			{Label: "OpEx", Value: -800},
			{Label: "Net", IsTotal: true},
		},
	})

	path := filepath.Join(t.TempDir(), "waterfall.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_WaterfallWithConnectors(t *testing.T) {
	f := false
	v := Waterfall(WaterfallCfg{
		BaseCfg: BaseCfg{
			ID:    "test-waterfall-noconn",
			Width: 400, Height: 300,
		},
		Values: []WaterfallValue{
			{Label: "Opening", Value: 10000, IsTotal: true},
			{Label: "Sales", Value: 3200},
			{Label: "Closing", IsTotal: true},
		},
		ShowConnectors: &f,
		Radius:         4,
		UpColor:        gui.Hex(0x2ca02c),
		DownColor:      gui.Hex(0xd62728),
		TotalColor:     gui.Hex(0x1f77b4),
	})

	path := filepath.Join(t.TempDir(), "waterfall_styled.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func assertBar(
	t *testing.T,
	bb waterfallBar,
	wantBottom, wantTop, wantTotal float64,
	wantKind int,
) {
	t.Helper()
	if bb.Bottom != wantBottom {
		t.Errorf("Bottom: got %.4g, want %.4g", bb.Bottom, wantBottom)
	}
	if bb.Top != wantTop {
		t.Errorf("Top: got %.4g, want %.4g", bb.Top, wantTop)
	}
	if bb.RunningTotal != wantTotal {
		t.Errorf("RunningTotal: got %.4g, want %.4g",
			bb.RunningTotal, wantTotal)
	}
	if bb.Kind != wantKind {
		t.Errorf("Kind: got %d, want %d", bb.Kind, wantKind)
	}
}
