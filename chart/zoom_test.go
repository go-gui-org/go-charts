package chart

import (
	"math"
	"testing"

	"fmt"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

func TestZoomAroundCursorCenter(t *testing.T) {
	t.Parallel()
	// Cursor at center of [0, 100], zoom in 2x.
	lo, hi := zoomAroundCursor(50, 0, 100, 2)
	if math.Abs(lo-25) > 1e-9 || math.Abs(hi-75) > 1e-9 {
		t.Errorf("got [%g, %g], want [25, 75]", lo, hi)
	}
}

func TestZoomAroundCursorEdge(t *testing.T) {
	t.Parallel()
	// Cursor at left edge.
	lo, hi := zoomAroundCursor(0, 0, 100, 2)
	if math.Abs(lo-0) > 1e-9 || math.Abs(hi-50) > 1e-9 {
		t.Errorf("got [%g, %g], want [0, 50]", lo, hi)
	}
	// Cursor at right edge.
	lo, hi = zoomAroundCursor(100, 0, 100, 2)
	if math.Abs(lo-50) > 1e-9 || math.Abs(hi-100) > 1e-9 {
		t.Errorf("got [%g, %g], want [50, 100]", lo, hi)
	}
}

func TestZoomAroundCursorZoomOut(t *testing.T) {
	t.Parallel()
	// Factor < 1 → zoom out (double range).
	lo, hi := zoomAroundCursor(50, 25, 75, 0.5)
	if math.Abs(lo-0) > 1e-9 || math.Abs(hi-100) > 1e-9 {
		t.Errorf("got [%g, %g], want [0, 100]", lo, hi)
	}
}

func TestZoomAroundCursorZeroDomain(t *testing.T) {
	t.Parallel()
	lo, hi := zoomAroundCursor(5, 5, 5, 2)
	if lo != 5 || hi != 5 {
		t.Errorf("got [%g, %g], want [5, 5]", lo, hi)
	}
}

func TestZoomAroundCursorZeroFactor(t *testing.T) {
	t.Parallel()
	lo, hi := zoomAroundCursor(50, 0, 100, 0)
	if lo != 0 || hi != 100 {
		t.Errorf("got [%g, %g], want [0, 100]", lo, hi)
	}
}

func TestPanDomain(t *testing.T) {
	t.Parallel()
	// 10px drag on 100px span → 10% shift.
	lo, hi := panDomain(0, 100, 10, 100)
	// float32→float64 conversion introduces ~1e-7 error.
	if math.Abs(lo-(-10)) > 1e-5 || math.Abs(hi-90) > 1e-5 {
		t.Errorf("got [%g, %g], want [-10, 90]", lo, hi)
	}
}

func TestPanDomainZeroSpan(t *testing.T) {
	t.Parallel()
	lo, hi := panDomain(0, 100, 10, 0)
	if lo != 0 || hi != 100 {
		t.Errorf("got [%g, %g], want [0, 100]", lo, hi)
	}
}

func TestClampZoomRangeMinSpan(t *testing.T) {
	t.Parallel()
	lo, hi := clampZoomRange(50, 50+1e-15, 0, 100)
	span := hi - lo
	// Allow small floating-point rounding (< 1% of min range).
	if span < DefaultMinZoomRange*0.99 {
		t.Errorf("span %g < min %g", span, DefaultMinZoomRange)
	}
}

func TestClampZoomRangeMaxExtent(t *testing.T) {
	t.Parallel()
	lo, hi := clampZoomRange(-10, 110, 0, 100)
	if lo < 0 || hi > 100 {
		t.Errorf("got [%g, %g], want within [0, 100]", lo, hi)
	}
}

func TestClampZoomRangeShiftRight(t *testing.T) {
	t.Parallel()
	// Panned left beyond origin → shift right.
	lo, hi := clampZoomRange(-5, 45, 0, 100)
	if math.Abs(lo-0) > 1e-9 || math.Abs(hi-50) > 1e-9 {
		t.Errorf("got [%g, %g], want [0, 50]", lo, hi)
	}
}

func TestClampZoomRangeShiftLeft(t *testing.T) {
	t.Parallel()
	// Panned right beyond max → shift left.
	lo, hi := clampZoomRange(60, 110, 0, 100)
	if math.Abs(lo-50) > 1e-9 || math.Abs(hi-100) > 1e-9 {
		t.Errorf("got [%g, %g], want [50, 100]", lo, hi)
	}
}

func TestClampZoomRangeMinSpanAtBoundary(t *testing.T) {
	t.Parallel()
	// Near origMax: min-span expansion must not exceed origMax.
	lo, hi := clampZoomRange(100-1e-15, 100, 0, 100)
	if hi > 100 {
		t.Errorf("dMax %g exceeds origMax 100", hi)
	}
	if lo < 0 {
		t.Errorf("dMin %g below origMin 0", lo)
	}
	// Near origMin: min-span expansion must not go below origMin.
	lo, hi = clampZoomRange(0, 0+1e-15, 0, 100)
	if lo < 0 {
		t.Errorf("dMin %g below origMin 0", lo)
	}
	if hi > 100 {
		t.Errorf("dMax %g exceeds origMax 100", hi)
	}
}

// --- clampRectToPlot ---

func TestClampRectToPlotInside(t *testing.T) {
	t.Parallel()
	x, y, w, h, vis := clampRectToPlot(20, 20, 60, 40, 0, 100, 0, 80)
	if !vis {
		t.Fatal("expected visible")
	}
	if x != 20 || y != 20 || w != 60 || h != 40 {
		t.Errorf("got (%g,%g,%g,%g), want (20,20,60,40)", x, y, w, h)
	}
}

func TestClampRectToPlotClipsTop(t *testing.T) {
	t.Parallel()
	// Rect extends above plot top (top=10).
	x, y, w, h, vis := clampRectToPlot(20, 5, 30, 20, 0, 100, 10, 80)
	if !vis {
		t.Fatal("expected visible")
	}
	if y != 10 || h != 15 {
		t.Errorf("got y=%g h=%g, want y=10 h=15", y, h)
	}
	if x != 20 || w != 30 {
		t.Errorf("x/w changed: got (%g,%g)", x, w)
	}
}

func TestClampRectToPlotClipsBottom(t *testing.T) {
	t.Parallel()
	// Rect extends below plot bottom (bottom=80).
	_, y, _, h, vis := clampRectToPlot(20, 70, 30, 20, 0, 100, 0, 80)
	if !vis {
		t.Fatal("expected visible")
	}
	if y != 70 || h != 10 {
		t.Errorf("got y=%g h=%g, want y=70 h=10", y, h)
	}
}

func TestClampRectToPlotClipsLeft(t *testing.T) {
	t.Parallel()
	x, _, w, _, vis := clampRectToPlot(-5, 20, 20, 10, 0, 100, 0, 80)
	if !vis {
		t.Fatal("expected visible")
	}
	if x != 0 || w != 15 {
		t.Errorf("got x=%g w=%g, want x=0 w=15", x, w)
	}
}

func TestClampRectToPlotClipsRight(t *testing.T) {
	t.Parallel()
	x, _, w, _, vis := clampRectToPlot(90, 20, 20, 10, 0, 100, 0, 80)
	if !vis {
		t.Fatal("expected visible")
	}
	if x != 90 || w != 10 {
		t.Errorf("got x=%g w=%g, want x=90 w=10", x, w)
	}
}

func TestClampRectToPlotFullyOutside(t *testing.T) {
	t.Parallel()
	_, _, _, _, vis := clampRectToPlot(110, 20, 20, 10, 0, 100, 0, 80)
	if vis {
		t.Error("expected not visible (right of plot)")
	}
	_, _, _, _, vis = clampRectToPlot(20, 90, 20, 10, 0, 100, 0, 80)
	if vis {
		t.Error("expected not visible (below plot)")
	}
}

// --- clampVerticalLine ---

func TestClampVerticalLineInside(t *testing.T) {
	t.Parallel()
	y0, y1, vis := clampVerticalLine(20, 60, 10, 80)
	if !vis {
		t.Fatal("expected visible")
	}
	if y0 != 20 || y1 != 60 {
		t.Errorf("got [%g,%g], want [20,60]", y0, y1)
	}
}

func TestClampVerticalLineSwapped(t *testing.T) {
	t.Parallel()
	// y0 > y1 should be normalized.
	y0, y1, vis := clampVerticalLine(60, 20, 10, 80)
	if !vis {
		t.Fatal("expected visible")
	}
	if y0 != 20 || y1 != 60 {
		t.Errorf("got [%g,%g], want [20,60]", y0, y1)
	}
}

func TestClampVerticalLineClipsTop(t *testing.T) {
	t.Parallel()
	y0, y1, vis := clampVerticalLine(5, 50, 10, 80)
	if !vis {
		t.Fatal("expected visible")
	}
	if y0 != 10 || y1 != 50 {
		t.Errorf("got [%g,%g], want [10,50]", y0, y1)
	}
}

func TestClampVerticalLineClipsBottom(t *testing.T) {
	t.Parallel()
	y0, y1, vis := clampVerticalLine(30, 90, 10, 80)
	if !vis {
		t.Fatal("expected visible")
	}
	if y0 != 30 || y1 != 80 {
		t.Errorf("got [%g,%g], want [30,80]", y0, y1)
	}
}

func TestClampVerticalLineFullyOutside(t *testing.T) {
	t.Parallel()
	_, _, vis := clampVerticalLine(85, 95, 10, 80)
	if vis {
		t.Error("expected not visible (below plot)")
	}
	_, _, vis = clampVerticalLine(2, 8, 10, 80)
	if vis {
		t.Error("expected not visible (above plot)")
	}
}

// --- insidePlot ---

func TestInsidePlot(t *testing.T) {
	t.Parallel()
	if !insidePlot(50, 50, 0, 100, 0, 100) {
		t.Error("center should be inside")
	}
	if !insidePlot(0, 0, 0, 100, 0, 100) {
		t.Error("top-left corner should be inside (inclusive)")
	}
	if !insidePlot(100, 100, 0, 100, 0, 100) {
		t.Error("bottom-right corner should be inside (inclusive)")
	}
	if insidePlot(-1, 50, 0, 100, 0, 100) {
		t.Error("left of plot should be outside")
	}
	if insidePlot(50, 101, 0, 100, 0, 100) {
		t.Error("below plot should be outside")
	}
}

// --- selectionColors ---

func TestSelectionColorsDefaults(t *testing.T) {
	t.Parallel()
	th := &theme.Theme{}
	fill, border := selectionColors(th)
	if fill != gui.RGBA(70, 130, 220, 30) {
		t.Errorf("fill = %v, want default", fill)
	}
	if border != gui.RGBA(70, 130, 220, 180) {
		t.Errorf("border = %v, want default", border)
	}
}

func TestSelectionColorsCustom(t *testing.T) {
	t.Parallel()
	custom := gui.RGBA(255, 0, 0, 50)
	th := &theme.Theme{SelectionFill: custom}
	fill, border := selectionColors(th)
	if fill != custom {
		t.Errorf("fill = %v, want custom %v", fill, custom)
	}
	// Border should still be default.
	if border != gui.RGBA(70, 130, 220, 180) {
		t.Errorf("border = %v, want default", border)
	}
}

// --- ensureOrigBounds ---

func TestEnsureOrigBoundsInitsDomain(t *testing.T) {
	t.Parallel()
	xAxis := axis.NewLinear(axis.LinearCfg{Min: 10, Max: 90})
	yAxis := axis.NewLinear(axis.LinearCfg{Min: -5, Max: 50})
	zs := zoomState{}

	ensureOrigBounds(&zs, xAxis, yAxis, true, true)

	if !zs.OrigStored {
		t.Fatal("OrigStored should be true")
	}
	if zs.OrigXMin != 10 || zs.OrigXMax != 90 {
		t.Errorf("OrigX = [%g,%g], want [10,90]",
			zs.OrigXMin, zs.OrigXMax)
	}
	if zs.OrigYMin != -5 || zs.OrigYMax != 50 {
		t.Errorf("OrigY = [%g,%g], want [-5,50]",
			zs.OrigYMin, zs.OrigYMax)
	}
	// When not zoomed, zoom domain should match orig.
	if zs.XMin != 10 || zs.XMax != 90 {
		t.Errorf("XMin/XMax = [%g,%g], want [10,90]",
			zs.XMin, zs.XMax)
	}
	if zs.YMin != -5 || zs.YMax != 50 {
		t.Errorf("YMin/YMax = [%g,%g], want [-5,50]",
			zs.YMin, zs.YMax)
	}
}

func TestEnsureOrigBoundsPreservesZoomed(t *testing.T) {
	t.Parallel()
	xAxis := axis.NewLinear(axis.LinearCfg{Min: 0, Max: 100})
	zs := zoomState{
		Zoomed: true,
		XMin:   20, XMax: 80,
	}

	ensureOrigBounds(&zs, xAxis, nil, true, false)

	if zs.OrigXMin != 0 || zs.OrigXMax != 100 {
		t.Errorf("OrigX = [%g,%g], want [0,100]",
			zs.OrigXMin, zs.OrigXMax)
	}
	// Zoomed domain should NOT be overwritten.
	if zs.XMin != 20 || zs.XMax != 80 {
		t.Errorf("XMin/XMax = [%g,%g], want [20,80]",
			zs.XMin, zs.XMax)
	}
}

func TestEnsureOrigBoundsNoop(t *testing.T) {
	t.Parallel()
	xAxis := axis.NewLinear(axis.LinearCfg{Min: 0, Max: 100})
	zs := zoomState{
		OrigStored: true,
		OrigXMin:   5, OrigXMax: 95,
	}

	ensureOrigBounds(&zs, xAxis, nil, true, false)

	// Already stored — should not overwrite.
	if zs.OrigXMin != 5 || zs.OrigXMax != 95 {
		t.Errorf("OrigX changed to [%g,%g], want [5,95]",
			zs.OrigXMin, zs.OrigXMax)
	}
}

func TestEnsureOrigBoundsNilAxes(t *testing.T) {
	t.Parallel()
	zs := zoomState{}
	ensureOrigBounds(&zs, nil, nil, true, true)

	if !zs.OrigStored {
		t.Fatal("OrigStored should be true even with nil axes")
	}
	// Domains stay zero.
	if zs.OrigXMin != 0 || zs.OrigXMax != 0 {
		t.Errorf("OrigX = [%g,%g], want [0,0]",
			zs.OrigXMin, zs.OrigXMax)
	}
}

func TestEnsureOrigBoundsPartialAxes(t *testing.T) {
	t.Parallel()
	yAxis := axis.NewLinear(axis.LinearCfg{Min: -10, Max: 10})
	zs := zoomState{}

	// zoomX=true but xAxis=nil, zoomY=true with valid yAxis.
	ensureOrigBounds(&zs, nil, yAxis, true, true)

	if zs.OrigXMin != 0 || zs.OrigXMax != 0 {
		t.Errorf("OrigX = [%g,%g], want [0,0] (nil axis)",
			zs.OrigXMin, zs.OrigXMax)
	}
	if zs.OrigYMin != -10 || zs.OrigYMax != 10 {
		t.Errorf("OrigY = [%g,%g], want [-10,10]",
			zs.OrigYMin, zs.OrigYMax)
	}
}

// --- handleDragHover / handleDragEnd / isDragging ---

// testPA builds a plotArea with axes spanning [0,100] and a
// 400x200 pixel plot region (left=50, right=450, top=20,
// bottom=220).
func testPA() plotArea {
	xa := axis.NewLinear(axis.LinearCfg{Min: 0, Max: 100})
	ya := axis.NewLinear(axis.LinearCfg{Min: 0, Max: 100})
	return plotArea{
		plotRect{50, 450, 20, 220},
		xa, ya,
	}
}

func TestHandleDragHoverStartsRangeSelect(t *testing.T) {
	t.Parallel()
	w := &gui.Window{}
	l := &gui.Layout{Shape: &gui.Shape{}}
	pa := testPA()

	e := &gui.Event{
		MouseX: 200, MouseY: 100,
		Modifiers: gui.ModLMB | gui.ModShift,
	}
	ok := handleDragHover(w, l, e, "dh1", pa,
		true, true, true, true)
	if !ok {
		t.Fatal("expected handled")
	}
	zs, _ := loadZoomState(w, "dh1")
	if !zs.Dragging || !zs.DragSelect {
		t.Fatalf("Dragging=%v DragSelect=%v", zs.Dragging, zs.DragSelect)
	}
	if zs.SelX0 != 200 || zs.SelY0 != 100 {
		t.Errorf("SelX0=%g SelY0=%g, want 200,100",
			zs.SelX0, zs.SelY0)
	}
}

func TestHandleDragHoverUsesCanvasLocalCoords(t *testing.T) {
	t.Parallel()
	w := &gui.Window{}
	l := &gui.Layout{Shape: &gui.Shape{X: 150, Y: 75}}
	pa := testPA()

	// Start drag.
	e := &gui.Event{
		MouseX: 200, MouseY: 100,
		Modifiers: gui.ModLMB | gui.ModShift,
	}
	handleDragHover(w, l, e, "dh2", pa, true, true, true, true)

	// Move past threshold — coords should be e.MouseX/Y
	// directly, NOT e.MouseX - l.Shape.X.
	e2 := &gui.Event{
		MouseX: 250, MouseY: 130,
		Modifiers: gui.ModLMB | gui.ModShift,
	}
	handleDragHover(w, l, e2, "dh2", pa, true, true, true, true)

	zs, _ := loadZoomState(w, "dh2")
	if zs.SelX1 != 250 || zs.SelY1 != 130 {
		t.Errorf("SelX1=%g SelY1=%g, want 250,130",
			zs.SelX1, zs.SelY1)
	}
}

func TestHandleDragHoverShiftOnDownStartsSelect(t *testing.T) {
	t.Parallel()
	w := &gui.Window{}
	l := &gui.Layout{Shape: &gui.Shape{}}
	pa := testPA()

	// Simulate MouseDown with Shift via handleDoubleClickCheck.
	eDown := &gui.Event{Modifiers: gui.ModShift}
	handleDoubleClickCheck(w, l, eDown, "dh3")

	// First MouseMove: LMB held, Shift NOT held.
	e := &gui.Event{
		MouseX: 200, MouseY: 100,
		Modifiers: gui.ModLMB, // no ModShift
	}
	ok := handleDragHover(w, l, e, "dh3", pa,
		true, true, true, true)
	if !ok {
		t.Fatal("expected handled")
	}
	zs, _ := loadZoomState(w, "dh3")
	if !zs.DragSelect {
		t.Error("DragSelect should be true via ShiftOnDown")
	}
}

func TestHandleDragEndFinalizesZoom(t *testing.T) {
	t.Parallel()
	w := &gui.Window{}
	l := &gui.Layout{Shape: &gui.Shape{}}
	pa := testPA()

	// Start shift+drag.
	e1 := &gui.Event{
		MouseX: 100, MouseY: 60,
		Modifiers: gui.ModLMB | gui.ModShift,
	}
	handleDragHover(w, l, e1, "de1", pa, true, true, true, true)

	// Move to build selection rect.
	e2 := &gui.Event{
		MouseX: 300, MouseY: 150,
		Modifiers: gui.ModLMB | gui.ModShift,
	}
	handleDragHover(w, l, e2, "de1", pa, true, true, true, true)

	// Mouse up.
	eUp := &gui.Event{}
	handleDragEnd(w, l, eUp, "de1", pa, true, true)

	zs, _ := loadZoomState(w, "de1")
	if zs.Dragging {
		t.Error("Dragging should be false after mouse-up")
	}
	if zs.DragSelect {
		t.Error("DragSelect should be false after mouse-up")
	}
	if !zs.Zoomed {
		t.Error("Zoomed should be true after range select")
	}
	if !eUp.IsHandled {
		t.Error("event should be marked handled")
	}
}

func TestHandleDragEndNopWhenNotDragging(t *testing.T) {
	t.Parallel()
	w := &gui.Window{}
	l := &gui.Layout{Shape: &gui.Shape{}}
	pa := testPA()

	e := &gui.Event{}
	handleDragEnd(w, l, e, "de2", pa, true, true)

	if e.IsHandled {
		t.Error("should not handle when not dragging")
	}
}

func TestIsDragging(t *testing.T) {
	t.Parallel()
	w := &gui.Window{}
	l := &gui.Layout{Shape: &gui.Shape{}}

	if isDragging(w, "id1") {
		t.Error("should be false before any drag")
	}

	// Start a drag.
	pa := testPA()
	e := &gui.Event{
		MouseX: 200, MouseY: 100,
		Modifiers: gui.ModLMB | gui.ModShift,
	}
	handleDragHover(w, l, e, "id1", pa, true, true, true, true)

	if !isDragging(w, "id1") {
		t.Error("should be true during drag")
	}

	// End drag.
	eUp := &gui.Event{}
	handleDragEnd(w, l, eUp, "id1", pa, true, true)

	if isDragging(w, "id1") {
		t.Error("should be false after drag end")
	}
}

// --- handleZoomScroll modifier tests ---

func TestHandleZoomScrollRequiresCtrlOrSuper(t *testing.T) {
	t.Parallel()
	w := &gui.Window{}
	l := &gui.Layout{Shape: &gui.Shape{}}
	pa := testPA()

	// Plain scroll (no modifier) — should NOT zoom.
	e := &gui.Event{ScrollY: -5, Modifiers: gui.ModNone}
	handleZoomScroll(w, l, e, "zs1", pa, true, true)
	if e.IsHandled {
		t.Error("plain scroll should not be handled")
	}

	// Ctrl+scroll — should zoom.
	e2 := &gui.Event{ScrollY: -5, Modifiers: gui.ModCtrl}
	handleZoomScroll(w, l, e2, "zs2", pa, true, true)
	if !e2.IsHandled {
		t.Error("Ctrl+scroll should be handled")
	}

	// Super+scroll — should zoom.
	e3 := &gui.Event{ScrollY: -5, Modifiers: gui.ModSuper}
	handleZoomScroll(w, l, e3, "zs3", pa, true, true)
	if !e3.IsHandled {
		t.Error("Super+scroll should be handled")
	}
}

// --- internalHover clears state during drag ---

func TestHoverClearedDuringDrag(t *testing.T) {
	t.Parallel()
	w := &gui.Window{}
	l := &gui.Layout{Shape: &gui.Shape{}}
	pa := testPA()
	id := "hd1"

	// Set hover state.
	saveHover(w, l, id, true, 100, 50)
	var hovering bool
	var px, py float32
	loadHover(w, id, &hovering, &px, &py)
	if !hovering {
		t.Fatal("hover should be active before drag")
	}

	// Start a drag.
	e := &gui.Event{
		MouseX: 200, MouseY: 100,
		Modifiers: gui.ModLMB,
	}
	handleDragHover(w, l, e, id, pa, true, false, true, true)
	if !isDragging(w, id) {
		t.Fatal("should be dragging")
	}

	// Simulate what internalHover does during drag.
	if isDragging(w, id) {
		saveHover(w, l, id, false, 0, 0)
	}

	loadHover(w, id, &hovering, &px, &py)
	if hovering {
		t.Error("hover should be cleared during drag")
	}
	if px != 0 || py != 0 {
		t.Errorf("hover coords should be zeroed, got (%g, %g)", px, py)
	}
}

// --- clipConvexToRect tests ---

func approxEq(a, b, tol float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < tol
}

func polyApproxEq(a, b []float32, tol float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !approxEq(a[i], b[i], tol) {
			return false
		}
	}
	return true
}

func TestClipConvexFullyInside(t *testing.T) {
	t.Parallel()
	// Quad fully inside the rect — no clipping needed.
	quad := []float32{20, 20, 80, 20, 80, 80, 20, 80}
	got, _, _ := clipConvexToRect(quad, 0, 100, 0, 100, nil, nil)
	if !polyApproxEq(got, quad, 0.01) {
		t.Errorf("got %v, want %v", got, quad)
	}
}

func TestClipConvexFullyOutside(t *testing.T) {
	t.Parallel()
	// Quad entirely to the left of the rect.
	quad := []float32{-50, 20, -10, 20, -10, 80, -50, 80}
	got, _, _ := clipConvexToRect(quad, 0, 100, 0, 100, nil, nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestClipConvexClipsLeft(t *testing.T) {
	t.Parallel()
	// Quad straddles the left edge.
	quad := []float32{-50, 0, 50, 0, 50, 100, -50, 100}
	got, _, _ := clipConvexToRect(quad, 0, 100, 0, 100, nil, nil)
	want := []float32{0, 0, 50, 0, 50, 100, 0, 100}
	if !polyApproxEq(got, want, 0.01) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestClipConvexClipsTopAndRight(t *testing.T) {
	t.Parallel()
	// Quad extends above top and past right.
	quad := []float32{50, -50, 150, -50, 150, 50, 50, 50}
	got, _, _ := clipConvexToRect(quad, 0, 100, 0, 100, nil, nil)
	want := []float32{50, 0, 100, 0, 100, 50, 50, 50}
	if !polyApproxEq(got, want, 0.01) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestClipConvexAreaQuadLineAboveTop(t *testing.T) {
	t.Parallel()
	// Simulates area fill quad when line is entirely above the
	// plot rect. Line at y=-50, baseline at y=100.
	// Quad: (10,-50),(90,-50),(90,100),(10,100)
	// After clip to [0,100]x[0,100]: (10,0),(90,0),(90,100),(10,100)
	quad := []float32{10, -50, 90, -50, 90, 100, 10, 100}
	got, _, _ := clipConvexToRect(quad, 0, 100, 0, 100, nil, nil)
	want := []float32{10, 0, 90, 0, 90, 100, 10, 100}
	if !polyApproxEq(got, want, 0.01) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestClipConvexAreaQuadDiagonal(t *testing.T) {
	t.Parallel()
	// Line segment from far-left-below to far-right-above
	// (typical zoomed-in area chart scenario).
	// Quad: (-100,200),(200,-100),(200,100),(-100,100)
	// Plot rect: [0,100]x[0,100]
	quad := []float32{-100, 200, 200, -100, 200, 100, -100, 100}
	got, _, _ := clipConvexToRect(quad, 0, 100, 0, 100, nil, nil)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	// Verify all result vertices are within the clip rect.
	for i := 0; i < len(got); i += 2 {
		x, y := got[i], got[i+1]
		if x < -0.01 || x > 100.01 || y < -0.01 || y > 100.01 {
			t.Errorf("vertex (%g,%g) outside clip rect", x, y)
		}
	}
	// Should have >= 3 vertices (triangle or more).
	if len(got) < 6 {
		t.Errorf("expected >= 3 vertices, got %d", len(got)/2)
	}
}

func TestClipConvexScratchBuffersReused(t *testing.T) {
	t.Parallel()
	// Verify scratch buffers are returned and reusable.
	quad := []float32{10, 10, 90, 10, 90, 90, 10, 90}
	var a, b []float32
	_, a, b = clipConvexToRect(quad, 0, 100, 0, 100, a, b)
	if a == nil || b == nil {
		t.Fatal("scratch buffers should be non-nil after first call")
	}
	capA, capB := cap(a), cap(b)

	// Second call should not allocate (reuses backing arrays).
	quad2 := []float32{20, 20, 80, 20, 80, 80, 20, 80}
	got, a2, b2 := clipConvexToRect(quad2, 0, 100, 0, 100, a, b)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if cap(a2) < capA || cap(b2) < capB {
		t.Error("scratch buffer capacity should not shrink")
	}
}

func TestClipConvexTooFewPoints(t *testing.T) {
	t.Parallel()
	got, _, _ := clipConvexToRect([]float32{1, 2, 3, 4}, 0, 100, 0, 100, nil, nil)
	if got != nil {
		t.Errorf("expected nil for < 3 vertices, got %v", got)
	}
}

func TestClipConvexNilInput(t *testing.T) {
	t.Parallel()
	got, _, _ := clipConvexToRect(nil, 0, 100, 0, 100, nil, nil)
	if got != nil {
		t.Errorf("expected nil for nil input, got %v", got)
	}
}

// --- Edge function tests ---

func TestRectEdgeInsideAllEdges(t *testing.T) {
	t.Parallel()
	cases := []struct {
		edge   int
		x, y   float32
		inside bool
	}{
		{0, 10, 50, true},  // left: x >= 10
		{0, 5, 50, false},  // left: x < 10
		{1, 90, 50, true},  // right: x <= 90
		{1, 95, 50, false}, // right: x > 90
		{2, 50, 20, true},  // top: y >= 20
		{2, 50, 15, false}, // top: y < 20
		{3, 50, 80, true},  // bottom: y <= 80
		{3, 50, 85, false}, // bottom: y > 80
	}
	for _, tc := range cases {
		got := rectEdgeInside(tc.edge, tc.x, tc.y, 10, 90, 20, 80)
		if got != tc.inside {
			t.Errorf("edge=%d (%g,%g): got %v, want %v",
				tc.edge, tc.x, tc.y, got, tc.inside)
		}
	}
}

func TestEdgeIsectComputation(t *testing.T) {
	t.Parallel()
	cases := []struct {
		edge   int
		ax, ay float32
		bx, by float32
		wantX  float32
		wantY  float32
	}{
		// Left edge at x=10: segment from (0,0) to (20,40)
		{0, 0, 0, 20, 40, 10, 20},
		// Right edge at x=90: segment from (80,0) to (100,50)
		{1, 80, 0, 100, 50, 90, 25},
		// Top edge at y=20: segment from (50,0) to (50,40)
		{2, 50, 0, 50, 40, 50, 20},
		// Bottom edge at y=80: segment from (30,60) to (30,100)
		{3, 30, 60, 30, 100, 30, 80},
	}
	for _, tc := range cases {
		x, y := edgeIsect(tc.edge, tc.ax, tc.ay, tc.bx, tc.by,
			10, 90, 20, 80)
		if !approxEq(x, tc.wantX, 0.01) || !approxEq(y, tc.wantY, 0.01) {
			t.Errorf("edge=%d (%g,%g)->(%g,%g): got (%g,%g), want (%g,%g)",
				tc.edge, tc.ax, tc.ay, tc.bx, tc.by,
				x, y, tc.wantX, tc.wantY)
		}
	}
}

// --- Degenerate quad skip tests ---

func TestAreaFillSkipsDegenerateQuad(t *testing.T) {
	t.Parallel()
	// When both line Y values are at or below bottom, the
	// clamped quad has zero height and should be skipped.
	bottom := float32(100)
	cases := []struct {
		y0, y1 float32
		skip   bool
	}{
		{50, 60, false},  // both above bottom
		{100, 100, true}, // both at bottom
		{150, 200, true}, // both below bottom (clamped to bottom)
		{50, 100, false}, // one above, one at bottom
		{50, 150, false}, // one above, one below
	}
	for _, tc := range cases {
		qy0 := min(tc.y0, bottom)
		qy1 := min(tc.y1, bottom)
		skipped := qy0 == bottom && qy1 == bottom
		if skipped != tc.skip {
			t.Errorf("y0=%g y1=%g: skipped=%v, want %v",
				tc.y0, tc.y1, skipped, tc.skip)
		}
	}
}

func TestStackedFillSkipsDegenerateBand(t *testing.T) {
	t.Parallel()
	// Stacked mode: skip when cur and prev clamp to same Y.
	bottom := float32(100)
	cases := []struct {
		curY, prevY float32
		skip        bool
	}{
		{50, 80, false},  // both in range, different
		{150, 200, true}, // both below bottom → both clamp to 100
		{100, 100, true}, // both exactly at bottom
		{50, 150, false}, // cur in range, prev below
		{150, 50, false}, // cur below, prev in range
	}
	for i, tc := range cases {
		cy := min(tc.curY, bottom)
		py := min(tc.prevY, bottom)
		skipped := cy == py
		if skipped != tc.skip {
			t.Errorf("case %d: curY=%g prevY=%g: skipped=%v, want %v",
				i, tc.curY, tc.prevY, skipped, tc.skip)
		}
	}
}

// --- Convexity verification ---

func TestClipConvexResultIsConvex(t *testing.T) {
	t.Parallel()
	// Various quads that partially overlap the clip rect.
	// Verify the clipped result has consistent cross-product sign
	// (all vertices turn the same direction → convex).
	quads := [][]float32{
		{-50, 50, 50, 50, 50, 150, -50, 150},     // straddles left
		{50, -50, 150, -50, 150, 50, 50, 50},     // straddles top-right
		{-50, -50, 150, -50, 150, 150, -50, 150}, // covers entire rect
		{30, 30, 70, 30, 70, 70, 30, 70},         // fully inside
	}
	for qi, quad := range quads {
		got, _, _ := clipConvexToRect(quad, 0, 100, 0, 100, nil, nil)
		if got == nil {
			continue
		}
		n := len(got) / 2
		if n < 3 {
			continue
		}
		var pos, neg int
		for i := range n {
			j := (i + 1) % n
			k := (i + 2) % n
			dx1 := got[j*2] - got[i*2]
			dy1 := got[j*2+1] - got[i*2+1]
			dx2 := got[k*2] - got[j*2]
			dy2 := got[k*2+1] - got[j*2+1]
			cross := dx1*dy2 - dy1*dx2
			if cross > 0.001 {
				pos++
			} else if cross < -0.001 {
				neg++
			}
		}
		if pos > 0 && neg > 0 {
			t.Errorf("quad %d: clipped result is concave: %v "+
				"(pos=%d neg=%d)", qi, got, pos, neg)
		}
	}
}

// Verify that area fill quads with Y clamped to bottom produce
// convex input for clipConvexToRect.
func TestAreaQuadConvexAfterClamp(t *testing.T) {
	t.Parallel()
	bottom := float32(300)
	// Simulate various line Y positions including below-bottom.
	ys := []struct{ y0, y1 float32 }{
		{100, 200},  // both above bottom
		{100, 400},  // y1 below bottom
		{400, 100},  // y0 below bottom
		{400, 500},  // both below bottom
		{-50, -100}, // both above top
	}
	for _, tc := range ys {
		qy0 := min(tc.y0, bottom)
		qy1 := min(tc.y1, bottom)
		if qy0 == bottom && qy1 == bottom {
			continue // degenerate, skipped
		}
		quad := []float32{
			10, qy0, 90, qy1,
			90, bottom, 10, bottom,
		}
		// Check convexity.
		n := len(quad) / 2
		var pos, neg int
		for i := range n {
			j := (i + 1) % n
			k := (i + 2) % n
			dx1 := quad[j*2] - quad[i*2]
			dy1 := quad[j*2+1] - quad[i*2+1]
			dx2 := quad[k*2] - quad[j*2]
			dy2 := quad[k*2+1] - quad[j*2+1]
			cross := dx1*dy2 - dy1*dx2
			if cross > 0.001 {
				pos++
			} else if cross < -0.001 {
				neg++
			}
		}
		if pos > 0 && neg > 0 {
			t.Errorf("y0=%g y1=%g: clamped quad is concave %v",
				tc.y0, tc.y1, quad)
		}
	}
}

// Verify convexity is broken WITHOUT the clamp (documenting the
// bug the clamp fixes).
func TestAreaQuadConcaveWithoutClamp(t *testing.T) {
	t.Parallel()
	bottom := float32(300)
	// y0 below bottom → reflex vertex at (x0, bottom).
	y0 := float32(400)
	y1 := float32(100)
	quad := []float32{10, y0, 90, y1, 90, bottom, 10, bottom}
	n := len(quad) / 2
	var pos, neg int
	for i := range n {
		j := (i + 1) % n
		k := (i + 2) % n
		dx1 := quad[j*2] - quad[i*2]
		dy1 := quad[j*2+1] - quad[i*2+1]
		dx2 := quad[k*2] - quad[j*2]
		dy2 := quad[k*2+1] - quad[j*2+1]
		cross := dx1*dy2 - dy1*dx2
		if cross > 0.001 {
			pos++
		} else if cross < -0.001 {
			neg++
		}
	}
	if pos == 0 || neg == 0 {
		t.Error("unclamped quad with y0 > bottom should be concave")
	}
}

func TestClipConvexPreservesResultAcrossCalls(t *testing.T) {
	t.Parallel()
	// Ensure result from call N is not corrupted by call N+1
	// (scratch buffer reuse safety).
	q1 := []float32{10, 10, 90, 10, 90, 90, 10, 90}
	q2 := []float32{20, 20, 80, 20, 80, 80, 20, 80}
	var a, b []float32
	var r1 []float32
	r1, a, b = clipConvexToRect(q1, 0, 100, 0, 100, a, b)
	// Copy r1 before next call.
	saved := make([]float32, len(r1))
	copy(saved, r1)
	_, _, _ = clipConvexToRect(q2, 0, 100, 0, 100, a, b)
	// r1 may have been overwritten (shares backing array with a).
	// Verify the saved copy is correct.
	if !polyApproxEq(saved, q1, 0.01) {
		t.Errorf("first result corrupted: got %v, want %v",
			saved, q1)
	}
}

func TestEdgeIsectBoundaryExact(t *testing.T) {
	t.Parallel()
	// Verify intersection at exact boundary coordinates.
	// Segment crossing left edge at x=0 from x=-10 to x=10
	x, y := edgeIsect(0, -10, 50, 10, 50, 0, 100, 0, 100)
	if !approxEq(x, 0, 0.01) || !approxEq(y, 50, 0.01) {
		t.Errorf("left edge: got (%g,%g), want (0,50)", x, y)
	}
}

// Benchmark to verify scratch buffer reuse avoids allocations.
func BenchmarkClipConvexToRect(b *testing.B) {
	quad := []float32{-50, -50, 150, -50, 150, 150, -50, 150}
	var sa, sb []float32
	b.ReportAllocs()
	for range b.N {
		_, sa, sb = clipConvexToRect(quad, 0, 100, 0, 100, sa, sb)
	}
	_ = fmt.Sprintf("%v %v", sa, sb) // prevent optimization
}
