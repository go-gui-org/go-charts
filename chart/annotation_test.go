package chart

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-gui/gui"
)

func testPlotRect() plotRect {
	return plotRect{Left: 60, Right: 380, Top: 40, Bottom: 260}
}

func testLinearAxes() (axis.Axis, axis.Axis) {
	xAxis := axis.NewLinear(axis.LinearCfg{Min: 0, Max: 100})
	yAxis := axis.NewLinear(axis.LinearCfg{Min: 0, Max: 100})
	return xAxis, yAxis
}

// --- empty annotations ---

func TestDrawAnnotationsEmpty(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, yAxis := testLinearAxes()
	ann := Annotations{}
	drawAnnotations(ctx, &ann, th, pr, xAxis, yAxis)
	if len(dc.Batches()) != 0 {
		t.Errorf("empty annotations should produce no batches, got %d",
			len(dc.Batches()))
	}
	if len(dc.Texts()) != 0 {
		t.Errorf("empty annotations should produce no texts, got %d",
			len(dc.Texts()))
	}
}

// --- line annotations ---

func TestDrawLineAnnotationHorizontal(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:  AnnotationY,
			Value: 50,
			Color: gui.Hex(0xFF0000),
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if len(dc.Batches()) == 0 {
		t.Fatal("horizontal line annotation should produce batches")
	}
}

func TestDrawLineAnnotationVertical(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, _ := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:  AnnotationX,
			Value: 50,
			Color: gui.Hex(0x0000FF),
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, nil)
	if len(dc.Batches()) == 0 {
		t.Fatal("vertical line annotation should produce batches")
	}
}

func TestDrawLineAnnotationDashed(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:    AnnotationY,
			Value:   75,
			DashLen: 6,
			GapLen:  4,
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if len(dc.Batches()) == 0 {
		t.Fatal("dashed line annotation should produce batches")
	}
}

func TestDrawLineAnnotationWithLabel(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:  AnnotationY,
			Value: 50,
			Label: "target",
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if _, ok := findText(dc, "target"); !ok {
		t.Error("line annotation label not rendered")
	}
}

func TestDrawLineAnnotationOutOfRange(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:  AnnotationY,
			Value: 200, // outside 0-100 range
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if len(dc.Batches()) != 0 {
		t.Error("out-of-range line annotation should not draw")
	}
}

func TestDrawLineAnnotationNilAxis(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	ann := Annotations{
		Lines: []LineAnnotation{
			{Axis: AnnotationX, Value: 50},
			{Axis: AnnotationY, Value: 50},
		},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, nil)
	if len(dc.Batches()) != 0 {
		t.Error("nil axes should skip all line annotations")
	}
}

// --- region annotations ---

func TestDrawRegionAnnotationY(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Regions: []RegionAnnotation{{
			Axis:  AnnotationY,
			Min:   20,
			Max:   80,
			Color: gui.RGBA(0, 255, 0, 40),
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if len(dc.Batches()) == 0 {
		t.Fatal("Y region annotation should produce batches")
	}
}

func TestDrawRegionAnnotationX(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, _ := testLinearAxes()
	ann := Annotations{
		Regions: []RegionAnnotation{{
			Axis: AnnotationX,
			Min:  10,
			Max:  90,
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, nil)
	if len(dc.Batches()) == 0 {
		t.Fatal("X region annotation should produce batches")
	}
}

func TestDrawRegionAnnotationWithLabel(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Regions: []RegionAnnotation{{
			Axis:  AnnotationY,
			Min:   30,
			Max:   70,
			Label: "zone",
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if _, ok := findText(dc, "zone"); !ok {
		t.Error("region annotation label not rendered")
	}
}

func TestDrawRegionAnnotationNilAxis(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	ann := Annotations{
		Regions: []RegionAnnotation{
			{Axis: AnnotationX, Min: 10, Max: 90},
		},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, nil)
	if len(dc.Batches()) != 0 {
		t.Error("nil X axis should skip X region annotation")
	}
}

// --- text annotations ---

func TestDrawTextAnnotation(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, yAxis := testLinearAxes()
	ann := Annotations{
		Texts: []TextAnnotation{{
			X:    50,
			Y:    50,
			Text: "hello",
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, yAxis)
	if _, ok := findText(dc, "hello"); !ok {
		t.Error("text annotation not rendered")
	}
}

func TestDrawTextAnnotationOutOfBounds(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, yAxis := testLinearAxes()
	ann := Annotations{
		Texts: []TextAnnotation{{
			X:    200, // outside 0-100 range
			Y:    50,
			Text: "offscreen",
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, yAxis)
	for _, te := range dc.Texts() {
		if te.Text == "offscreen" {
			t.Error("out-of-bounds text should not render")
		}
	}
}

func TestDrawTextAnnotationNilAxis(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	ann := Annotations{
		Texts: []TextAnnotation{{
			X: 50, Y: 50, Text: "skip",
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, nil)
	if len(dc.Texts()) != 0 {
		t.Error("nil axes should skip text annotation")
	}
}

// --- dash fallback ---

func TestDrawLineAnnotationDashNoGapFallsBackToSolid(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:    AnnotationY,
			Value:   50,
			DashLen: 6,
			GapLen:  0, // no gap → solid
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if len(dc.Batches()) == 0 {
		t.Fatal("DashLen with zero GapLen should still draw (solid)")
	}
}

func TestDrawLineAnnotationNegativeDashFallsBackToSolid(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:    AnnotationY,
			Value:   50,
			DashLen: -5,
			GapLen:  -3,
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if len(dc.Batches()) == 0 {
		t.Fatal("negative dash/gap should fall back to solid line")
	}
}

// --- empty text guard ---

func TestDrawTextAnnotationEmptyText(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, yAxis := testLinearAxes()
	ann := Annotations{
		Texts: []TextAnnotation{{
			X:               50,
			Y:               50,
			Text:            "",
			LabelBackground: gui.RGBA(0, 0, 0, 100),
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, yAxis)
	if len(dc.Batches()) != 0 {
		t.Error("empty text should not draw background rect")
	}
}

// --- text annotation styling ---

func TestDrawTextAnnotationCustomStyle(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, yAxis := testLinearAxes()
	ann := Annotations{
		Texts: []TextAnnotation{{
			X:        50,
			Y:        50,
			Text:     "styled",
			Color:    gui.Hex(0xFF0000),
			FontSize: 18,
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, yAxis)
	found := false
	for _, te := range dc.Texts() {
		if te.Text == "styled" {
			found = true
			if te.Style.Color != gui.Hex(0xFF0000) {
				t.Errorf("color = %v, want red", te.Style.Color)
			}
			if te.Style.Size != 18 {
				t.Errorf("size = %v, want 18", te.Style.Size)
			}
		}
	}
	if !found {
		t.Error("styled text annotation not rendered")
	}
}

// --- vertical line label ---

func TestDrawLineAnnotationVerticalWithLabel(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, _ := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:  AnnotationX,
			Value: 50,
			Label: "marker",
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, nil)
	found := false
	for _, te := range dc.Texts() {
		if te.Text == "marker" {
			found = true
			break
		}
	}
	if !found {
		t.Error("vertical line annotation label not rendered")
	}
}

// --- region Y swapped ---

func TestDrawRegionAnnotationYSwapped(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Regions: []RegionAnnotation{{
			Axis: AnnotationY,
			Min:  80,
			Max:  20, // swapped
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if len(dc.Batches()) == 0 {
		t.Fatal("Y region with swapped min/max should still draw")
	}
}

// --- category axis ---

func TestDrawAnnotationsWithCategoryAxis(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis := axis.NewCategory(axis.CategoryCfg{
		Categories: []string{"A", "B", "C", "D"},
	})
	yAxis := axis.NewLinear(axis.LinearCfg{Min: 0, Max: 100})
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:  AnnotationX,
			Value: 1, // category index
		}},
		Regions: []RegionAnnotation{{
			Axis: AnnotationX,
			Min:  0,
			Max:  2,
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, yAxis)
	if len(dc.Batches()) == 0 {
		t.Fatal("category axis annotations should produce batches")
	}
}

// --- multiple annotations ---

func TestDrawMultipleAnnotations(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, yAxis := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{
			{Axis: AnnotationY, Value: 25},
			{Axis: AnnotationY, Value: 75},
			{Axis: AnnotationX, Value: 50},
		},
		Regions: []RegionAnnotation{
			{Axis: AnnotationY, Min: 10, Max: 30},
			{Axis: AnnotationX, Min: 60, Max: 90},
		},
		Texts: []TextAnnotation{
			{X: 20, Y: 20, Text: "a"},
			{X: 80, Y: 80, Text: "b"},
		},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, yAxis)
	if len(dc.Batches()) == 0 {
		t.Errorf("multiple annotations should produce batches, got %d",
			len(dc.Batches()))
	}
	if len(dc.Texts()) < 2 {
		t.Errorf("should have >= 2 text entries, got %d",
			len(dc.Texts()))
	}
}

// --- default color paths ---

func TestDrawAnnotationsDefaultColors(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Lines:   []LineAnnotation{{Axis: AnnotationY, Value: 50}},
		Regions: []RegionAnnotation{{Axis: AnnotationY, Min: 20, Max: 40}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if len(dc.Batches()) == 0 {
		t.Fatal("default-color annotations should still draw")
	}
}

// --- label positions ---

func findText(dc *gui.DrawContext, text string) (gui.DrawCanvasTextEntry, bool) {
	for _, te := range dc.Texts() {
		if te.Text == text {
			return te, true
		}
	}
	return gui.DrawCanvasTextEntry{}, false
}

func TestDrawLineAnnotationHorizontalLabelPositions(t *testing.T) {
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	th := testTheme()

	xs := make(map[LabelPosition]float32)
	for _, pos := range []LabelPosition{LabelStart, LabelCenter, LabelEnd} {
		ctx, dc := testCtx(400, 300)
		ann := Annotations{
			Lines: []LineAnnotation{{
				Axis:     AnnotationY,
				Value:    50,
				Label:    "pos",
				LabelPos: pos,
			}},
		}
		drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
		te, ok := findText(dc, "pos")
		if !ok {
			t.Fatalf("LabelPosition %d: label not rendered", pos)
		}
		xs[pos] = te.X
	}
	if xs[LabelStart] >= xs[LabelCenter] {
		t.Errorf("Start X (%g) should be < Center X (%g)",
			xs[LabelStart], xs[LabelCenter])
	}
	if xs[LabelCenter] >= xs[LabelEnd] {
		t.Errorf("Center X (%g) should be < End X (%g)",
			xs[LabelCenter], xs[LabelEnd])
	}
}

func TestDrawLineAnnotationVerticalLabelPositions(t *testing.T) {
	pr := testPlotRect()
	xAxis, _ := testLinearAxes()
	th := testTheme()

	ys := make(map[LabelPosition]float32)
	for _, pos := range []LabelPosition{LabelStart, LabelCenter, LabelEnd} {
		ctx, dc := testCtx(400, 300)
		ann := Annotations{
			Lines: []LineAnnotation{{
				Axis:     AnnotationX,
				Value:    50,
				Label:    "vpos",
				LabelPos: pos,
			}},
		}
		drawAnnotations(ctx, &ann, th, pr, xAxis, nil)
		te, ok := findText(dc, "vpos")
		if !ok {
			t.Fatalf("vertical LabelPosition %d: label not rendered", pos)
		}
		ys[pos] = te.Y
	}
	// End = top (small Y), Start = bottom (large Y)
	if ys[LabelEnd] >= ys[LabelCenter] {
		t.Errorf("End Y (%g) should be < Center Y (%g)",
			ys[LabelEnd], ys[LabelCenter])
	}
	if ys[LabelCenter] >= ys[LabelStart] {
		t.Errorf("Center Y (%g) should be < Start Y (%g)",
			ys[LabelCenter], ys[LabelStart])
	}
}

func TestDrawLineAnnotationVerticalDefaultIsTop(t *testing.T) {
	pr := testPlotRect()
	xAxis, _ := testLinearAxes()
	th := testTheme()
	ctx, dc := testCtx(400, 300)
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:  AnnotationX,
			Value: 50,
			Label: "def",
			// LabelPos zero = LabelEnd = top for vertical
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, nil)
	te, ok := findText(dc, "def")
	if !ok {
		t.Fatal("default label not rendered")
	}
	// Default (LabelEnd) should be near top of plot.
	mid := (pr.Top + pr.Bottom) / 2
	if te.Y > mid {
		t.Errorf("default vertical label Y (%g) should be above midpoint (%g)",
			te.Y, mid)
	}
}

// --- label backgrounds ---

func TestDrawTextAnnotationBackground(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, yAxis := testLinearAxes()
	ann := Annotations{
		Texts: []TextAnnotation{{
			X:               50,
			Y:               50,
			Text:            "bg",
			LabelBackground: gui.RGBA(20, 20, 20, 180),
			LabelRadius:     4,
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, yAxis)
	if len(dc.Batches()) == 0 {
		t.Fatal("text with background should produce batches")
	}
	if _, ok := findText(dc, "bg"); !ok {
		t.Error("text annotation with background not rendered")
	}
}

func TestDrawLineAnnotationLabelBackground(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{{
			Axis:            AnnotationY,
			Value:           50,
			Label:           "ref",
			LabelBackground: gui.RGBA(200, 0, 0, 180),
			LabelRadius:     3,
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if _, ok := findText(dc, "ref"); !ok {
		t.Error("line annotation label with background not rendered")
	}
	// Background rect + line = batches.
	if len(dc.Batches()) == 0 {
		t.Fatal("label background should produce batches")
	}
}

func TestDrawRegionAnnotationLabelBackground(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	_, yAxis := testLinearAxes()
	ann := Annotations{
		Regions: []RegionAnnotation{{
			Axis:            AnnotationY,
			Min:             20,
			Max:             80,
			Label:           "zone",
			LabelBackground: gui.RGBA(0, 0, 0, 120),
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, nil, yAxis)
	if _, ok := findText(dc, "zone"); !ok {
		t.Error("region label with background not rendered")
	}
}

func TestDrawLabelBackgroundNoRadius(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, yAxis := testLinearAxes()
	ann := Annotations{
		Texts: []TextAnnotation{{
			X:               50,
			Y:               50,
			Text:            "flat",
			LabelBackground: gui.RGBA(0, 0, 0, 100),
			LabelRadius:     0, // no rounding
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, yAxis)
	if len(dc.Batches()) == 0 {
		t.Fatal("label background with radius=0 should use FilledRect")
	}
}

// --- NaN/Inf guards ---

func TestDrawAnnotationsNaNSkipped(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, yAxis := testLinearAxes()
	ann := Annotations{
		Lines: []LineAnnotation{
			{Axis: AnnotationY, Value: math.NaN()},
			{Axis: AnnotationX, Value: math.Inf(1)},
		},
		Regions: []RegionAnnotation{
			{Axis: AnnotationY, Min: math.NaN(), Max: 50},
		},
		Texts: []TextAnnotation{
			{X: math.NaN(), Y: 50, Text: "nan"},
			{X: 50, Y: math.Inf(-1), Text: "inf"},
		},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, yAxis)
	if len(dc.Batches()) != 0 {
		t.Errorf("NaN/Inf annotations should produce no batches, got %d",
			len(dc.Batches()))
	}
	if len(dc.Texts()) != 0 {
		t.Errorf("NaN/Inf annotations should produce no texts, got %d",
			len(dc.Texts()))
	}
}

// --- X region min/max swap ---

func TestDrawRegionAnnotationXSwapped(t *testing.T) {
	ctx, dc := testCtx(400, 300)
	th := testTheme()
	pr := testPlotRect()
	xAxis, _ := testLinearAxes()
	ann := Annotations{
		Regions: []RegionAnnotation{{
			Axis: AnnotationX,
			Min:  90,
			Max:  10, // swapped
		}},
	}
	drawAnnotations(ctx, &ann, th, pr, xAxis, nil)
	if len(dc.Batches()) == 0 {
		t.Fatal("X region with swapped min/max should still draw")
	}
}

// --- Annotations.empty ---

func TestAnnotationsEmpty(t *testing.T) {
	a := Annotations{}
	if !a.empty() {
		t.Error("zero-value Annotations should be empty")
	}
	a.Lines = []LineAnnotation{{Value: 1}}
	if a.empty() {
		t.Error("Annotations with lines should not be empty")
	}
}
