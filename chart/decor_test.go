package chart

import (
	"testing"

	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

func testCtx(w, h float32) (*render.Context, *gui.DrawContext) {
	dc := &gui.DrawContext{Width: w, Height: h}
	return render.NewContext(dc), dc
}

func testTheme() *theme.Theme {
	return &theme.Theme{
		TitleStyle:    gui.TextStyle{Size: 14, Color: gui.White},
		LabelStyle:    gui.TextStyle{Size: 11, Color: gui.White},
		TickStyle:     gui.TextStyle{Size: 11, Color: gui.White},
		AxisColor:     gui.White,
		AxisWidth:     1,
		GridColor:     gui.RGBA(128, 128, 128, 40),
		GridWidth:     0.5,
		Palette:       theme.DefaultPalette(),
		PaddingTop:    40,
		PaddingRight:  20,
		PaddingBottom: 40,
		PaddingLeft:   60,
	}
}

// --- drawTitle ---

func TestDrawTitleRendersText(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	drawTitle(ctx, "My Chart", th)
	if len(dc.Texts()) != 1 {
		t.Fatalf("texts = %d, want 1", len(dc.Texts()))
	}
	if dc.Texts()[0].Text != "My Chart" {
		t.Errorf("text = %q, want %q",
			dc.Texts()[0].Text, "My Chart")
	}
}

func TestDrawTitleEmptySkipped(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	drawTitle(ctx, "", th)
	if len(dc.Texts()) != 0 {
		t.Errorf("empty title should produce no text, got %d",
			len(dc.Texts()))
	}
}

// --- drawLegend ---

func TestDrawLegendRendersEntries(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	entries := []legendEntry{
		{Name: "Series A", Color: gui.Hex(0xFF0000)},
		{Name: "Series B", Color: gui.Hex(0x00FF00)},
	}
	drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, nil, nil)
	// Background rect + 2 swatches = 3 rounded rects → batches.
	if len(dc.Batches()) == 0 {
		t.Error("expected batches for legend background/swatches")
	}
	// 2 text labels.
	if len(dc.Texts()) != 2 {
		t.Errorf("texts = %d, want 2", len(dc.Texts()))
	}
}

func TestDrawLegendSkipsUnnamedEntries(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	entries := []legendEntry{
		{Name: "", Color: gui.Hex(0xFF0000)},
		{Name: "Visible", Color: gui.Hex(0x00FF00)},
	}
	drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, nil, nil)
	// Only 1 named entry → 1 text label.
	if len(dc.Texts()) != 1 {
		t.Errorf("texts = %d, want 1", len(dc.Texts()))
	}
}

func TestDrawLegendAllUnnamedSkipped(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	entries := []legendEntry{
		{Name: "", Color: gui.Hex(0xFF0000)},
		{Name: "", Color: gui.Hex(0x00FF00)},
	}
	drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, nil, nil)
	if len(dc.Texts()) != 0 {
		t.Errorf("all-unnamed should produce no text, got %d",
			len(dc.Texts()))
	}
	if len(dc.Batches()) != 0 {
		t.Errorf("all-unnamed should produce no batches, got %d",
			len(dc.Batches()))
	}
}

func TestDrawLegendEmptyEntries(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	drawLegend(ctx, nil, th, plotRect{60, 380, 40, 260}, nil, nil)
	if len(dc.Texts()) != 0 || len(dc.Batches()) != 0 {
		t.Error("nil entries should produce no output")
	}
}

// --- Integration: bar chart title + legend ---

func TestBarDrawTitleAndLegend(t *testing.T) {
	bv := Bar(BarCfg{
		BaseCfg: BaseCfg{
			ID:    "test-bar",
			Title: "Bar Title",
		},
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Name: "S1",
				Values: []series.CategoryValue{
					{Label: "A", Value: 10},
					{Label: "B", Value: 20},
				},
			}),
			series.NewCategory(series.CategoryCfg{
				Name: "S2",
				Values: []series.CategoryValue{
					{Label: "A", Value: 15},
					{Label: "B", Value: 25},
				},
			}),
		},
	}).(*barView)

	dc := gui.DrawContext{Width: 400, Height: 300}
	bv.draw(&dc)

	// Should have title text + Y tick labels + X tick labels +
	// legend labels.
	hasTitle := false
	hasLegendS1 := false
	hasLegendS2 := false
	for _, txt := range dc.Texts() {
		switch txt.Text {
		case "Bar Title":
			hasTitle = true
		case "S1":
			hasLegendS1 = true
		case "S2":
			hasLegendS2 = true
		}
	}
	if !hasTitle {
		t.Error("title not rendered")
	}
	if !hasLegendS1 {
		t.Error("legend entry S1 not rendered")
	}
	if !hasLegendS2 {
		t.Error("legend entry S2 not rendered")
	}
}

// --- Integration: line chart title + legend ---

func TestLineDrawTitleAndLegend(t *testing.T) {
	lv := Line(LineCfg{
		BaseCfg: BaseCfg{
			ID:    "test-line",
			Title: "Line Title",
		},
		Series: []series.XY{
			series.XYFromYValues("Alpha", []float64{1, 2, 3}),
			series.XYFromYValues("Beta", []float64{3, 2, 1}),
		},
	}).(*lineView)

	dc := gui.DrawContext{Width: 400, Height: 300}
	lv.draw(&dc)

	hasTitle := false
	hasAlpha := false
	hasBeta := false
	for _, txt := range dc.Texts() {
		switch txt.Text {
		case "Line Title":
			hasTitle = true
		case "Alpha":
			hasAlpha = true
		case "Beta":
			hasBeta = true
		}
	}
	if !hasTitle {
		t.Error("title not rendered")
	}
	if !hasAlpha {
		t.Error("legend entry Alpha not rendered")
	}
	if !hasBeta {
		t.Error("legend entry Beta not rendered")
	}
}

// --- No title/legend when absent ---

func TestBarDrawNoTitleNoLegend(t *testing.T) {
	bv := Bar(BarCfg{
		BaseCfg: BaseCfg{ID: "test-bar-bare"},
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Values: []series.CategoryValue{
					{Label: "A", Value: 10},
				},
			}),
		},
	}).(*barView)

	dc := gui.DrawContext{Width: 400, Height: 300}
	bv.draw(&dc)

	for _, txt := range dc.Texts() {
		// Should only have tick labels, no title or legend text.
		if txt.Text == "" {
			t.Error("empty text rendered")
		}
	}
	// No legend entries since series has no name.
	for _, txt := range dc.Texts() {
		if txt.Text == "test-bar-bare" {
			t.Error("ID should not appear as title")
		}
	}
}

// --- Legend position ---

func TestDrawLegendTopLeft(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	th.Legend.Position = theme.LegendTopLeft
	entries := []legendEntry{
		{Name: "A", Color: gui.Hex(0xFF0000)},
	}
	drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, nil, nil)
	if len(dc.Texts()) != 1 {
		t.Fatalf("texts = %d, want 1", len(dc.Texts()))
	}
}

func TestDrawLegendPerChartOverride(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	// Theme says TopRight, override to BottomLeft.
	pos := theme.LegendBottomLeft
	entries := []legendEntry{
		{Name: "A", Color: gui.Hex(0xFF0000)},
	}
	drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, &pos, nil)
	if len(dc.Texts()) != 1 {
		t.Fatalf("texts = %d, want 1", len(dc.Texts()))
	}
}

func TestDrawLegendCustomBackground(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	th.Legend.Background = gui.RGBA(255, 0, 0, 200)
	entries := []legendEntry{
		{Name: "A", Color: gui.Hex(0x00FF00)},
	}
	drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, nil, nil)
	if len(dc.Batches()) == 0 {
		t.Error("expected batches for legend background")
	}
}

func TestDrawLegendCustomTextStyle(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	th.Legend.TextStyle = gui.TextStyle{
		Size: 20, Color: gui.Hex(0xFF0000),
	}
	entries := []legendEntry{
		{Name: "Big", Color: gui.Hex(0x0000FF)},
	}
	drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, nil, nil)
	if len(dc.Texts()) != 1 {
		t.Fatalf("texts = %d, want 1", len(dc.Texts()))
	}
}

// --- LegendNone ---

func TestDrawLegendNone(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pos := theme.LegendNone
	entries := []legendEntry{
		{Name: "A", Color: gui.Hex(0xFF0000)},
		{Name: "B", Color: gui.Hex(0x00FF00)},
	}
	drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, &pos, nil)
	if len(dc.Texts()) != 0 {
		t.Errorf("LegendNone should produce no text, got %d",
			len(dc.Texts()))
	}
	if len(dc.Batches()) != 0 {
		t.Errorf("LegendNone should produce no batches, got %d",
			len(dc.Batches()))
	}
}

func TestDrawLegendNoneViaTheme(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	th.Legend.Position = theme.LegendNone
	entries := []legendEntry{
		{Name: "A", Color: gui.Hex(0xFF0000)},
	}
	drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, nil, nil)
	if len(dc.Texts()) != 0 {
		t.Errorf("LegendNone via theme should produce no text, got %d",
			len(dc.Texts()))
	}
}

// --- LegendBottom ---

func TestDrawLegendBottom(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pos := theme.LegendBottom
	entries := []legendEntry{
		{Name: "Alpha", Color: gui.Hex(0xFF0000), Index: 0},
		{Name: "Beta", Color: gui.Hex(0x00FF00), Index: 1},
	}
	lb := drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, &pos, nil)
	// Should render 2 text labels.
	if len(dc.Texts()) != 2 {
		t.Fatalf("texts = %d, want 2", len(dc.Texts()))
	}
	// Should have background + 2 swatches.
	if len(dc.Batches()) == 0 {
		t.Error("expected batches for legend background/swatches")
	}
	// Hit-test bounds should have 2 entries.
	if len(lb.EntryRects) != 2 {
		t.Fatalf("entry rects = %d, want 2", len(lb.EntryRects))
	}
	// Legend should be near bottom of canvas.
	for _, r := range lb.EntryRects {
		if r.Y < 200 {
			t.Errorf("legend entry Y=%v, expected near bottom", r.Y)
		}
	}
}

func TestDrawLegendBottomSingleEntry(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pos := theme.LegendBottom
	entries := []legendEntry{
		{Name: "Solo", Color: gui.Hex(0xFF0000), Index: 0},
	}
	drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, &pos, nil)
	if len(dc.Texts()) != 1 {
		t.Fatalf("texts = %d, want 1", len(dc.Texts()))
	}
}

// --- LegendRight ---

func TestDrawLegendRight(t *testing.T) {
	ctx, dc := testCtx(500, 300)
	th := testTheme()
	pos := theme.LegendRight
	entries := []legendEntry{
		{Name: "Alpha", Color: gui.Hex(0xFF0000), Index: 0},
		{Name: "Beta", Color: gui.Hex(0x00FF00), Index: 1},
	}
	lb := drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, &pos, nil)
	if len(dc.Texts()) != 2 {
		t.Fatalf("texts = %d, want 2", len(dc.Texts()))
	}
	if len(dc.Batches()) == 0 {
		t.Error("expected batches for legend background/swatches")
	}
	if len(lb.EntryRects) != 2 {
		t.Fatalf("entry rects = %d, want 2", len(lb.EntryRects))
	}
	// Legend should be to the right of the plot area.
	for _, r := range lb.EntryRects {
		if r.X < 380 {
			t.Errorf("legend entry X=%v, expected right of plot (380)",
				r.X)
		}
	}
}

func TestDrawLegendRightTopAligned(t *testing.T) {
	ctx, dc := testCtx(500, 300)
	th := testTheme()
	pos := theme.LegendRight
	entries := []legendEntry{
		{Name: "Solo", Color: gui.Hex(0xFF0000), Index: 0},
	}
	lb := drawLegend(ctx, entries, th, plotRect{60, 380, 40, 260}, &pos, nil)
	if len(dc.Texts()) != 1 {
		t.Fatalf("texts = %d, want 1", len(dc.Texts()))
	}
	// Top of legend box should be at plot top (40).
	if lb.EntryRects[0].Y < 40 || lb.EntryRects[0].Y > 60 {
		t.Errorf("legend Y=%v, expected near top (40)",
			lb.EntryRects[0].Y)
	}
}

// --- LegendTop ---

func TestDrawLegendTop(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pos := theme.LegendTop
	entries := []legendEntry{
		{Name: "Alpha", Color: gui.Hex(0xFF0000), Index: 0},
		{Name: "Beta", Color: gui.Hex(0x00FF00), Index: 1},
	}
	// top=80 simulates reserve having pushed it down from 40.
	lb := drawLegend(ctx, entries, th, plotRect{60, 380, 80, 260}, &pos, nil)
	if len(dc.Texts()) != 2 {
		t.Fatalf("texts = %d, want 2", len(dc.Texts()))
	}
	if len(dc.Batches()) == 0 {
		t.Error("expected batches for legend background/swatches")
	}
	if len(lb.EntryRects) != 2 {
		t.Fatalf("entry rects = %d, want 2", len(lb.EntryRects))
	}
	// Legend should be above the plot top (80).
	for _, r := range lb.EntryRects {
		if r.Y >= 80 {
			t.Errorf("legend entry Y=%v, expected above plot top (80)",
				r.Y)
		}
	}
}

func TestLegendTopReserve(t *testing.T) {
	ctx, _ := testCtx(400, 300)
	th := testTheme()
	names := []string{"Alpha", "Beta"}
	r := legendTopReserve(ctx, th, nil, names, 60, 380)
	if r != 0 {
		t.Errorf("default position should reserve 0, got %v", r)
	}
	pos := theme.LegendTop
	r = legendTopReserve(ctx, th, &pos, names, 60, 380)
	if r <= 0 {
		t.Error("LegendTop should reserve positive space")
	}
}

// --- Tick mark style ---

func TestResolvedTickMarkDefaults(t *testing.T) {
	th := testTheme()
	length, width, color := resolvedTickMark(th)
	if length != DefaultTickLength {
		t.Errorf("length = %v, want %v", length, DefaultTickLength)
	}
	if width != th.AxisWidth {
		t.Errorf("width = %v, want %v", width, th.AxisWidth)
	}
	if color != th.AxisColor {
		t.Errorf("color = %v, want %v", color, th.AxisColor)
	}
}

func TestResolvedTickMarkCustom(t *testing.T) {
	th := testTheme()
	th.TickMark = theme.TickMarkStyle{
		Length: 10,
		Color:  gui.Hex(0xFF0000),
		Width:  3,
	}
	length, width, color := resolvedTickMark(th)
	if length != 10 {
		t.Errorf("length = %v, want 10", length)
	}
	if width != 3 {
		t.Errorf("width = %v, want 3", width)
	}
	if color != gui.Hex(0xFF0000) {
		t.Errorf("color = %v, want red", color)
	}
}

// --- X tick rotation ---

func TestLineXTickRotation(t *testing.T) {
	lv := Line(LineCfg{
		BaseCfg: BaseCfg{
			ID:            "rot-test",
			XTickRotation: -0.5,
		},
		Series: []series.XY{
			series.XYFromYValues("S", []float64{1, 2, 3}),
		},
	}).(*lineView)

	dc := gui.DrawContext{Width: 400, Height: 300}
	lv.draw(&dc)

	// At least one X tick label should have rotation set.
	found := false
	for _, txt := range dc.Texts() {
		if txt.Style.RotationRadians == -0.5 {
			found = true
			break
		}
	}
	if !found {
		t.Error("no X tick label with expected rotation")
	}
}

// --- legendRightReserve ---

func TestLegendRightReserveDefault(t *testing.T) {
	ctx, _ := testCtx(400, 300)
	th := testTheme()
	names := []string{"Alpha", "Beta"}
	r := legendRightReserve(ctx, th, nil, names)
	if r != 0 {
		t.Errorf("default position should reserve 0, got %v", r)
	}
}

func TestLegendRightReserveActive(t *testing.T) {
	ctx, _ := testCtx(400, 300)
	th := testTheme()
	pos := theme.LegendRight
	names := []string{"Alpha", "Beta"}
	r := legendRightReserve(ctx, th, &pos, names)
	if r <= 0 {
		t.Error("LegendRight should reserve positive space")
	}
}

func TestLegendRightReserveAllEmpty(t *testing.T) {
	ctx, _ := testCtx(400, 300)
	th := testTheme()
	pos := theme.LegendRight
	names := []string{"", ""}
	r := legendRightReserve(ctx, th, &pos, names)
	if r != 0 {
		t.Errorf("all-empty names should reserve 0, got %v", r)
	}
}

// --- legendBottomReserve ---

func TestLegendBottomReserveDefault(t *testing.T) {
	ctx, _ := testCtx(400, 300)
	th := testTheme()
	names := []string{"Alpha", "Beta"}
	r := legendBottomReserve(ctx, th, nil, names, 60, 380)
	if r != 0 {
		t.Errorf("default position should reserve 0, got %v", r)
	}
}

func TestLegendBottomReserveActive(t *testing.T) {
	ctx, _ := testCtx(400, 300)
	th := testTheme()
	pos := theme.LegendBottom
	names := []string{"Alpha", "Beta"}
	r := legendBottomReserve(ctx, th, &pos, names, 60, 380)
	if r <= 0 {
		t.Error("LegendBottom should reserve positive space")
	}
}

func TestLegendBottomReserveAllEmpty(t *testing.T) {
	ctx, _ := testCtx(400, 300)
	th := testTheme()
	pos := theme.LegendBottom
	names := []string{"", ""}
	r := legendBottomReserve(ctx, th, &pos, names, 60, 380)
	if r != 0 {
		t.Errorf("all-empty names should reserve 0, got %v", r)
	}
}

// --- legendHitTest ---

func TestLegendHitTestHit(t *testing.T) {
	lb := legendBounds{
		EntryRects: []legendEntryRect{
			{Index: 0, X: 10, Y: 10, Width: 80, Height: 20},
			{Index: 1, X: 10, Y: 35, Width: 80, Height: 20},
		},
	}
	if idx := legendHitTest(lb, 50, 15); idx != 0 {
		t.Errorf("expected 0, got %d", idx)
	}
	if idx := legendHitTest(lb, 50, 40); idx != 1 {
		t.Errorf("expected 1, got %d", idx)
	}
}

func TestLegendHitTestMiss(t *testing.T) {
	lb := legendBounds{
		EntryRects: []legendEntryRect{
			{Index: 0, X: 10, Y: 10, Width: 80, Height: 20},
		},
	}
	if idx := legendHitTest(lb, 200, 200); idx != -1 {
		t.Errorf("expected -1, got %d", idx)
	}
}

func TestLegendHitTestEmpty(t *testing.T) {
	lb := legendBounds{}
	if idx := legendHitTest(lb, 50, 50); idx != -1 {
		t.Errorf("expected -1, got %d", idx)
	}
}
