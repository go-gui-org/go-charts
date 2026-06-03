package chart

import (
	"path/filepath"
	"testing"

	"github.com/go-gui-org/go-charts/series"
)

func testComboSeries() []ComboSeries {
	return []ComboSeries{
		{
			Category: series.NewCategory(series.CategoryCfg{
				Name: "Revenue",
				Values: []series.CategoryValue{
					{Label: "Jan", Value: 100},
					{Label: "Feb", Value: 150},
					{Label: "Mar", Value: 130},
					{Label: "Apr", Value: 180},
				},
			}),
			Type: ComboBar,
		},
		{
			Category: series.NewCategory(series.CategoryCfg{
				Name: "Trend",
				Values: []series.CategoryValue{
					{Label: "Jan", Value: 110},
					{Label: "Feb", Value: 130},
					{Label: "Mar", Value: 140},
					{Label: "Apr", Value: 160},
				},
			}),
			Type: ComboLine,
		},
	}
}

func TestComboValidateOK(t *testing.T) {
	cfg := ComboCfg{Series: testComboSeries()}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestComboValidateEmpty(t *testing.T) {
	cfg := ComboCfg{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty series")
	}
}

func TestComboValidateNegativeBarWidth(t *testing.T) {
	cfg := ComboCfg{
		Series:   testComboSeries(),
		BarWidth: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative BarWidth")
	}
}

func TestComboValidateNegativeBarGap(t *testing.T) {
	cfg := ComboCfg{
		Series: testComboSeries(),
		BarGap: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative BarGap")
	}
}

func TestComboValidateNegativeRadius(t *testing.T) {
	cfg := ComboCfg{
		Series: testComboSeries(),
		Radius: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative Radius")
	}
}

func TestComboValidateNegativeLineWidth(t *testing.T) {
	cfg := ComboCfg{
		Series:    testComboSeries(),
		LineWidth: -1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative LineWidth")
	}
}

func TestComboValidateSeriesLengthMismatch(t *testing.T) {
	cfg := ComboCfg{
		Series: []ComboSeries{
			{
				Category: series.NewCategory(series.CategoryCfg{
					Name: "A",
					Values: []series.CategoryValue{
						{Label: "X", Value: 1},
						{Label: "Y", Value: 2},
					},
				}),
				Type: ComboBar,
			},
			{
				Category: series.NewCategory(series.CategoryCfg{
					Name: "B",
					Values: []series.CategoryValue{
						{Label: "X", Value: 3},
					},
				}),
				Type: ComboLine,
			},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for series length mismatch")
	}
}

func TestComboYAxisComputation(t *testing.T) {
	s := []ComboSeries{
		{
			Category: series.NewCategory(series.CategoryCfg{
				Name: "Bars",
				Values: []series.CategoryValue{
					{Label: "A", Value: 10},
					{Label: "B", Value: 50},
				},
			}),
			Type: ComboBar,
		},
		{
			Category: series.NewCategory(series.CategoryCfg{
				Name: "Lines",
				Values: []series.CategoryValue{
					{Label: "A", Value: 80},
					{Label: "B", Value: 20},
				},
			}),
			Type: ComboLine,
		},
	}
	cv := &comboView{cfg: ComboCfg{Series: s}}
	cv.updateYAxis(&cv.cfg, 2)

	if cv.yAxis == nil {
		t.Fatal("yAxis not created")
	}
	// Y axis must encompass line value 80 (the global max).
	// Transform maps data→pixel; 80 should map within [top, bottom].
	py := cv.yAxis.Transform(80, 300, 0)
	if py < 0 || py > 300 {
		t.Errorf("value 80 mapped to %g, outside plot [0,300]", py)
	}
	// 0 should also be in range (bar baseline).
	py0 := cv.yAxis.Transform(0, 300, 0)
	if py0 < 0 || py0 > 300 {
		t.Errorf("value 0 mapped to %g, outside plot [0,300]", py0)
	}
}

func newComboViewForHover() *comboView {
	cv := &comboView{
		cfg: ComboCfg{
			Series: testComboSeries(),
			BarGap: DefaultBarGap,
		},
	}
	cv.updateYAxis(&cv.cfg, 4)
	return cv
}

func TestComboHoveredElement_Outside(t *testing.T) {
	cv := newComboViewForHover()
	// Left of plot.
	_, _, ok := cv.hoveredElement(-1, 150, 0, 400, 0, 300)
	if ok {
		t.Error("expected ok=false left of plot")
	}
	// Right of plot.
	_, _, ok = cv.hoveredElement(401, 150, 0, 400, 0, 300)
	if ok {
		t.Error("expected ok=false right of plot")
	}
	// Above plot.
	_, _, ok = cv.hoveredElement(200, -1, 0, 400, 0, 300)
	if ok {
		t.Error("expected ok=false above plot")
	}
	// Below plot.
	_, _, ok = cv.hoveredElement(200, 301, 0, 400, 0, 300)
	if ok {
		t.Error("expected ok=false below plot")
	}
}

func TestComboHoveredElement_BarHit(t *testing.T) {
	cv := newComboViewForHover()
	// Plot: left=0, right=400, top=0, bottom=300.
	// 4 categories → groupWidth=100. 1 bar series (si=0).
	// BarGap=4, barWidth = (100-8-0)/1 = 92, barStart = (100-92)/2 = 4.
	// Category 0 bar X range: [4, 96].
	// Value=100, baseline=0. yAxis maps 0→bottom(300), 100→somewhere
	// above. The bar spans from baseline pixel to value pixel.
	baseline := cv.yAxis.Transform(0, 300, 0)
	valPx := cv.yAxis.Transform(100, 300, 0)
	barMidY := (baseline + valPx) / 2

	ci, si, ok := cv.hoveredElement(50, barMidY, 0, 400, 0, 300)
	if !ok {
		t.Fatal("expected bar hit")
	}
	if ci != 0 {
		t.Errorf("category: got %d, want 0", ci)
	}
	if si != 0 {
		t.Errorf("series: got %d, want 0 (bar)", si)
	}

	// Category 2 bar center: groupX=200, barStart=204, midX=250.
	ci, si, ok = cv.hoveredElement(250, barMidY, 0, 400, 0, 300)
	if !ok {
		t.Fatal("expected bar hit at category 2")
	}
	if ci != 2 {
		t.Errorf("category: got %d, want 2", ci)
	}
	if si != 0 {
		t.Errorf("series: got %d, want 0 (bar)", si)
	}
}

func TestComboHoveredElement_LineSnap(t *testing.T) {
	cv := newComboViewForHover()
	// Line series is index 1. Category centers at 50, 150, 250, 350.
	// Category 0 line value=110. Compute its pixel position.
	linePx := cv.yAxis.Transform(110, 300, 0)

	// Click near line point for category 0 (center X=50).
	ci, si, ok := cv.hoveredElement(52, linePx+2, 0, 400, 0, 300)
	if !ok {
		t.Fatal("expected line snap hit")
	}
	if ci != 0 {
		t.Errorf("category: got %d, want 0", ci)
	}
	if si != 1 {
		t.Errorf("series: got %d, want 1 (line)", si)
	}
}

func TestComboHoveredElement_BarPriority(t *testing.T) {
	// When cursor is on a bar AND near a line point, bar wins.
	cv := newComboViewForHover()
	baseline := cv.yAxis.Transform(0, 300, 0)
	valPx := cv.yAxis.Transform(100, 300, 0)
	barMidY := (baseline + valPx) / 2

	// Category 0 bar center ~50, line point also at x=50.
	ci, si, ok := cv.hoveredElement(50, barMidY, 0, 400, 0, 300)
	if !ok {
		t.Fatal("expected hit")
	}
	// Bar series (si=0) should win over line series (si=1).
	if si != 0 {
		t.Errorf("expected bar series 0 priority, got si=%d", si)
	}
	if ci != 0 {
		t.Errorf("category: got %d, want 0", ci)
	}
}

func TestComboHoveredElement_HiddenBar(t *testing.T) {
	cv := newComboViewForHover()
	cv.hidden = map[int]bool{0: true} // hide bar series

	baseline := cv.yAxis.Transform(0, 300, 0)
	valPx := cv.yAxis.Transform(100, 300, 0)
	barMidY := (baseline + valPx) / 2

	// With bars hidden, clicking where bar was should not hit bar.
	// It may snap to line point instead.
	ci, si, ok := cv.hoveredElement(50, barMidY, 0, 400, 0, 300)
	if ok && si == 0 {
		t.Errorf("hidden bar should not be hit, got ci=%d si=%d", ci, si)
	}
}

func TestComboHoveredElement_HiddenLine(t *testing.T) {
	cv := newComboViewForHover()
	cv.hidden = map[int]bool{1: true} // hide line series

	linePx := cv.yAxis.Transform(110, 300, 0)
	// Click right at line point — but line is hidden.
	ci, si, ok := cv.hoveredElement(50, linePx, 0, 400, 0, 300)
	if ok && si == 1 {
		t.Errorf("hidden line should not be hit, got ci=%d si=%d", ci, si)
	}
}

func TestComboHoveredElement_NoHitEmptyArea(t *testing.T) {
	cv := newComboViewForHover()
	// Click in center of plot but far from any bar or line point.
	// Y value that maps to well above any data (~halfway up empty area).
	_, _, ok := cv.hoveredElement(200, 5, 0, 400, 0, 300)
	if ok {
		t.Error("expected no hit in empty area far from data")
	}
}

func TestExportPNG_Combo(t *testing.T) {
	v := Combo(ComboCfg{
		BaseCfg: BaseCfg{
			ID:    "test-combo",
			Width: 400, Height: 300,
		},
		Series: testComboSeries(),
	})

	path := filepath.Join(t.TempDir(), "combo.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_ComboWithMarkers(t *testing.T) {
	v := Combo(ComboCfg{
		BaseCfg: BaseCfg{
			ID:    "test-combo-markers",
			Width: 400, Height: 300,
		},
		Series: []ComboSeries{
			{
				Category: series.NewCategory(series.CategoryCfg{
					Name: "Growth",
					Values: []series.CategoryValue{
						{Label: "Q1", Value: 5},
						{Label: "Q2", Value: 12},
						{Label: "Q3", Value: 8},
					},
				}),
				Type: ComboLine,
			},
		},
		ShowMarkers: true,
		Radius:      4,
	})

	path := filepath.Join(t.TempDir(), "combo_markers.png")
	if err := ExportPNG(v, 400, 300, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 400, 300)
}

func TestExportPNG_ComboMulti(t *testing.T) {
	vals := func(vs ...float64) []series.CategoryValue {
		out := make([]series.CategoryValue, len(vs))
		labels := []string{"Jan", "Feb", "Mar", "Apr"}
		for i, v := range vs {
			out[i] = series.CategoryValue{Label: labels[i], Value: v}
		}
		return out
	}
	v := Combo(ComboCfg{
		BaseCfg: BaseCfg{
			ID:    "test-combo-multi",
			Width: 500, Height: 350,
		},
		Series: []ComboSeries{
			{Category: series.NewCategory(series.CategoryCfg{Name: "Online", Values: vals(100, 120, 90, 150)}), Type: ComboBar},
			{Category: series.NewCategory(series.CategoryCfg{Name: "In-Store", Values: vals(80, 100, 110, 95)}), Type: ComboBar},
			{Category: series.NewCategory(series.CategoryCfg{Name: "Online Trend", Values: vals(95, 105, 100, 140)}), Type: ComboLine},
			{Category: series.NewCategory(series.CategoryCfg{Name: "In-Store Trend", Values: vals(85, 95, 105, 100)}), Type: ComboLine},
		},
		ShowMarkers: true,
	})

	path := filepath.Join(t.TempDir(), "combo_multi.png")
	if err := ExportPNG(v, 500, 350, path); err != nil {
		t.Fatal(err)
	}
	assertValidPNG(t, path, 500, 350)
}
