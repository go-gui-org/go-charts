package chart

import (
	"math"
	"time"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// zoomState persists zoom/pan/selection across frames via
// gui.StateMap. Keyed by chart ID.
type zoomState struct {
	Zoomed                     bool
	XMin, XMax, YMin, YMax     float64
	OrigXMin, OrigXMax         float64
	OrigYMin, OrigYMax         float64
	OrigStored                 bool
	Dragging                   bool
	DragStartPx, DragStartPy   float32
	DragSelect                 bool // true = range-select, false = pan
	ShiftOnDown                bool // Shift held at MouseDown
	SelX0, SelY0, SelX1, SelY1 float32
	LastClickTime              int64
	Version                    uint64
}

const (
	nsChartZoom  = "chart-zoom"
	capChartZoom = 64
)

// isDragging returns true when a drag operation (pan or
// range-select) is active. Used to suppress hover effects.
// Callers return without setting e.IsHandled; layoutHover
// returns true regardless so event propagation is unaffected.
func isDragging(w *gui.Window, id string) bool {
	zs, _ := loadZoomState(w, id)
	return zs.Dragging
}

// loadZoomState reads persisted zoom state for a chart.
func loadZoomState(w *gui.Window, id string) (zoomState, uint64) {
	if w == nil || id == "" {
		return zoomState{}, 0
	}
	sm := gui.StateMapRead[string, zoomState](w, nsChartZoom)
	if sm == nil {
		return zoomState{}, 0
	}
	zs, ok := sm.Get(id)
	if !ok {
		return zoomState{}, 0
	}
	return zs, zs.Version
}

// loadZoomVersion returns only the version for cache-key use
// in GenerateLayout.
func loadZoomVersion(w *gui.Window, id string) uint64 {
	_, v := loadZoomState(w, id)
	return v
}

// saveZoomState writes zoom state and bumps the layout version
// to invalidate the draw cache. Pass nil layout when called
// from draw (version already bumped via GenerateLayout).
func saveZoomState(
	w *gui.Window, l *gui.Layout, id string, zs zoomState,
) {
	if w == nil || id == "" {
		return
	}
	sm := gui.StateMap[string, zoomState](w, nsChartZoom, capChartZoom)
	zs.Version++
	sm.Set(id, zs)
	if l != nil {
		l.Shape.Version = zs.Version
	}
}

// --- Pure math (no gui dependency) ---

// zoomAroundCursor zooms the domain [dMin, dMax] by factor,
// keeping cursorData at the same relative position. factor > 1
// zooms in (smaller range), < 1 zooms out.
func zoomAroundCursor(
	cursorData, dMin, dMax, factor float64,
) (float64, float64) {
	span := dMax - dMin
	if span == 0 || factor == 0 {
		return dMin, dMax
	}
	ratio := (cursorData - dMin) / span
	newRange := span / factor
	return cursorData - ratio*newRange,
		cursorData + (1-ratio)*newRange
}

// panDomain shifts [dMin, dMax] by a pixel delta converted to
// data space via the pixel span.
func panDomain(
	dMin, dMax float64, deltaPx, pixelSpan float32,
) (float64, float64) {
	if pixelSpan == 0 {
		return dMin, dMax
	}
	dataSpan := dMax - dMin
	shift := dataSpan * float64(deltaPx/pixelSpan)
	return dMin - shift, dMax - shift
}

// clampZoomRange enforces minimum domain span and prevents
// zooming beyond original bounds.
func clampZoomRange(
	dMin, dMax, origMin, origMax float64,
) (float64, float64) {
	// Clamp to original extent.
	if dMin < origMin {
		dMax += origMin - dMin
		dMin = origMin
	}
	if dMax > origMax {
		dMin -= dMax - origMax
		dMax = origMax
	}
	dMin = max(dMin, origMin)
	dMax = min(dMax, origMax)
	// Enforce minimum span after extent clamping.
	if dMax-dMin < DefaultMinZoomRange {
		mid := (dMin + dMax) / 2
		half := DefaultMinZoomRange / 2
		dMin = mid - half
		dMax = mid + half
	}
	// Re-clamp: min-span expansion may overshoot bounds.
	dMin = max(dMin, origMin)
	dMax = min(dMax, origMax)
	return dMin, dMax
}

// --- Axis helpers ---

// applyZoomToAxes sets the zoomed domain on the given axes
// and enables overrideDomain to prevent Ticks() expansion.
func applyZoomToAxes(
	zs zoomState, xAxis, yAxis axis.Axis,
	zoomX, zoomY bool,
) {
	if zoomX && xAxis != nil {
		xAxis.SetRange(zs.XMin, zs.XMax)
		xAxis.SetOverrideDomain(true)
	}
	if zoomY && yAxis != nil {
		yAxis.SetRange(zs.YMin, zs.YMax)
		yAxis.SetOverrideDomain(true)
	}
}

// --- Event handlers ---

// handleZoomScroll processes a mouse scroll event for zoom.
func handleZoomScroll(
	w *gui.Window, l *gui.Layout, e *gui.Event,
	id string, pa plotArea, zoomX, zoomY bool,
) {
	if pa.XAxis == nil && pa.YAxis == nil {
		return
	}
	if !e.Modifiers.HasAny(gui.ModCtrl, gui.ModSuper) {
		return
	}
	dy := e.ScrollY
	if dy == 0 {
		return
	}
	e.IsHandled = true

	zs, _ := loadZoomState(w, id)
	ensureOrigBounds(&zs, pa.XAxis, pa.YAxis, zoomX, zoomY)

	factor := math.Pow(DefaultZoomFactor, float64(dy))

	// Mouse position relative to canvas.
	mx := e.MouseX - l.Shape.X
	my := e.MouseY - l.Shape.Y

	if zoomX && pa.XAxis != nil {
		curX := pa.XAxis.Invert(mx, pa.Left, pa.Right)
		zs.XMin, zs.XMax = zoomAroundCursor(
			curX, zs.XMin, zs.XMax, factor)
		zs.XMin, zs.XMax = clampZoomRange(
			zs.XMin, zs.XMax, zs.OrigXMin, zs.OrigXMax)
	}
	if zoomY && pa.YAxis != nil {
		curY := pa.YAxis.Invert(my, pa.Bottom, pa.Top)
		zs.YMin, zs.YMax = zoomAroundCursor(
			curY, zs.YMin, zs.YMax, factor)
		zs.YMin, zs.YMax = clampZoomRange(
			zs.YMin, zs.YMax, zs.OrigYMin, zs.OrigYMax)
	}
	zs.Zoomed = true
	saveZoomState(w, l, id, zs)
}

// handleZoomGesture processes touch gestures for zoom/pan.
func handleZoomGesture(
	w *gui.Window, l *gui.Layout, e *gui.Event,
	id string, pa plotArea, zoomX, zoomY bool,
) {
	switch e.GestureType {
	case gui.GestureDoubleTap:
		e.IsHandled = true
		handleZoomReset(w, l, id)
		return

	case gui.GesturePinch:
		if e.GesturePhase == gui.GesturePhaseChanged {
			e.IsHandled = true
			zs, _ := loadZoomState(w, id)
			ensureOrigBounds(&zs, pa.XAxis, pa.YAxis, zoomX, zoomY)

			factor := float64(e.PinchScale)
			if factor <= 0 {
				return
			}
			cx := e.CentroidX - l.Shape.X
			cy := e.CentroidY - l.Shape.Y

			if zoomX && pa.XAxis != nil {
				curX := pa.XAxis.Invert(cx, pa.Left, pa.Right)
				zs.XMin, zs.XMax = zoomAroundCursor(
					curX, zs.XMin, zs.XMax, factor)
				zs.XMin, zs.XMax = clampZoomRange(
					zs.XMin, zs.XMax, zs.OrigXMin, zs.OrigXMax)
			}
			if zoomY && pa.YAxis != nil {
				curY := pa.YAxis.Invert(cy, pa.Bottom, pa.Top)
				zs.YMin, zs.YMax = zoomAroundCursor(
					curY, zs.YMin, zs.YMax, factor)
				zs.YMin, zs.YMax = clampZoomRange(
					zs.YMin, zs.YMax, zs.OrigYMin, zs.OrigYMax)
			}
			zs.Zoomed = true
			saveZoomState(w, l, id, zs)
		}
		return

	case gui.GesturePan:
		if e.GesturePhase == gui.GesturePhaseChanged {
			e.IsHandled = true
			zs, _ := loadZoomState(w, id)
			if !zs.Zoomed {
				return
			}
			if zoomX && pa.XAxis != nil {
				zs.XMin, zs.XMax = panDomain(
					zs.XMin, zs.XMax,
					e.GestureDX, pa.Right-pa.Left)
				zs.XMin, zs.XMax = clampZoomRange(
					zs.XMin, zs.XMax, zs.OrigXMin, zs.OrigXMax)
			}
			if zoomY && pa.YAxis != nil {
				zs.YMin, zs.YMax = panDomain(
					zs.YMin, zs.YMax,
					-e.GestureDY, pa.Bottom-pa.Top)
				zs.YMin, zs.YMax = clampZoomRange(
					zs.YMin, zs.YMax, zs.OrigYMin, zs.OrigYMax)
			}
			saveZoomState(w, l, id, zs)
		}
		return
	}
}

// handleDragHover processes mouse hover with LMB held for
// pan or range-select. Returns true if the event was consumed.
func handleDragHover(
	w *gui.Window, l *gui.Layout, e *gui.Event,
	id string, pa plotArea,
	panOk, selectOk, zoomX, zoomY bool,
) bool {
	lmbHeld := e.Modifiers&gui.ModLMB != 0
	zs, _ := loadZoomState(w, id)

	// LMB released while dragging → end drag.
	if zs.Dragging && !lmbHeld {
		if zs.DragSelect {
			finishRangeSelect(&zs, pa, zoomX, zoomY)
		}
		zs.Dragging = false
		zs.DragSelect = false
		saveZoomState(w, l, id, zs)
		return true
	}

	if !lmbHeld {
		return false
	}

	// OnMouseMove delivers canvas-local coordinates
	// (executeMouseCallback subtracts l.Shape.X/Y). Using raw
	// e.MouseX/Y avoids double subtraction.
	mx := e.MouseX
	my := e.MouseY

	// Start drag.
	if !zs.Dragging {
		zs.Dragging = true
		zs.DragStartPx = mx
		zs.DragStartPy = my
		shift := zs.ShiftOnDown || e.Modifiers&gui.ModShift != 0
		zs.DragSelect = shift && selectOk
		if !zs.DragSelect && !panOk {
			zs.Dragging = false
			return false
		}
		ensureOrigBounds(&zs, pa.XAxis, pa.YAxis, zoomX, zoomY)
		if zs.DragSelect {
			zs.SelX0 = mx
			zs.SelY0 = my
			zs.SelX1 = mx
			zs.SelY1 = my
		}
		saveZoomState(w, l, id, zs)
		e.IsHandled = true
		return true
	}

	// Check minimum drag distance.
	dx := mx - zs.DragStartPx
	dy := my - zs.DragStartPy
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if dist < DefaultMinDragPx {
		return true
	}

	e.IsHandled = true

	if zs.DragSelect {
		// Update selection rectangle.
		zs.SelX1 = mx
		zs.SelY1 = my
		saveZoomState(w, l, id, zs)
		w.SetMouseCursorCrosshair()
		return true
	}

	// Pan mode.
	if !zs.Zoomed {
		return true
	}
	if zoomX && pa.XAxis != nil {
		zs.XMin, zs.XMax = panDomain(
			zs.XMin, zs.XMax,
			e.MouseDX, pa.Right-pa.Left)
		zs.XMin, zs.XMax = clampZoomRange(
			zs.XMin, zs.XMax, zs.OrigXMin, zs.OrigXMax)
	}
	if zoomY && pa.YAxis != nil {
		zs.YMin, zs.YMax = panDomain(
			zs.YMin, zs.YMax,
			-e.MouseDY, pa.Bottom-pa.Top)
		zs.YMin, zs.YMax = clampZoomRange(
			zs.YMin, zs.YMax, zs.OrigYMin, zs.OrigYMax)
	}
	saveZoomState(w, l, id, zs)
	w.SetMouseCursorAll()
	return true
}

// handleDragEnd completes a drag operation on mouse-up. Called
// from OnMouseUp so the zoom applies immediately on button
// release without waiting for the next mouse move.
func handleDragEnd(
	w *gui.Window, l *gui.Layout, e *gui.Event,
	id string, pa plotArea, zoomX, zoomY bool,
) {
	zs, _ := loadZoomState(w, id)
	if !zs.Dragging {
		return
	}
	if zs.DragSelect {
		finishRangeSelect(&zs, pa, zoomX, zoomY)
	}
	zs.Dragging = false
	zs.DragSelect = false
	e.IsHandled = true
	saveZoomState(w, l, id, zs)
}

// finishRangeSelect converts the pixel selection rectangle to
// a data-space domain and sets it as the zoomed view.
func finishRangeSelect(
	zs *zoomState, pa plotArea, zoomX, zoomY bool,
) {
	// Normalize rect.
	x0, x1 := zs.SelX0, zs.SelX1
	y0, y1 := zs.SelY0, zs.SelY1
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}

	// Minimum selection size.
	if x1-x0 < DefaultMinDragPx && y1-y0 < DefaultMinDragPx {
		return
	}

	if zoomX && pa.XAxis != nil {
		dMin := pa.XAxis.Invert(x0, pa.Left, pa.Right)
		dMax := pa.XAxis.Invert(x1, pa.Left, pa.Right)
		if dMin > dMax {
			dMin, dMax = dMax, dMin
		}
		zs.XMin, zs.XMax = clampZoomRange(
			dMin, dMax, zs.OrigXMin, zs.OrigXMax)
	}
	if zoomY && pa.YAxis != nil {
		// Y pixel axis is inverted (top=min pixel, bottom=max pixel).
		dMin := pa.YAxis.Invert(y1, pa.Bottom, pa.Top)
		dMax := pa.YAxis.Invert(y0, pa.Bottom, pa.Top)
		if dMin > dMax {
			dMin, dMax = dMax, dMin
		}
		zs.YMin, zs.YMax = clampZoomRange(
			dMin, dMax, zs.OrigYMin, zs.OrigYMax)
	}
	zs.Zoomed = true
}

// handleDoubleClickCheck detects mouse double-clicks via
// wall-clock time and resets zoom. Returns true if a
// double-click was detected.
func handleDoubleClickCheck(
	w *gui.Window, l *gui.Layout, e *gui.Event, id string,
) bool {
	zs, _ := loadZoomState(w, id)
	zs.ShiftOnDown = e.Modifiers&gui.ModShift != 0
	now := time.Now().UnixMilli()
	prev := zs.LastClickTime
	zs.LastClickTime = now

	if prev > 0 && now-prev <= zoomDoubleClickMs {
		zs.LastClickTime = 0 // prevent triple-click
		if zs.Zoomed {
			zs.Zoomed = false
			zs.OrigStored = false
			saveZoomState(w, l, id, zs)
			return true
		}
	}
	saveZoomState(w, l, id, zs)
	return false
}

// handleZoomReset restores the original (pre-zoom) domain.
func handleZoomReset(w *gui.Window, l *gui.Layout, id string) {
	zs, _ := loadZoomState(w, id)
	if !zs.Zoomed {
		return
	}
	zs.Zoomed = false
	zs.OrigStored = false
	saveZoomState(w, l, id, zs)
}

// --- Draw helpers ---

// drawSelectionRectIf renders the brush-to-zoom selection
// rectangle if a drag-select is active. No-op otherwise.
func drawSelectionRectIf(
	ctx *render.Context, zs zoomState, pr plotRect,
	th *theme.Theme,
) {
	if !zs.Dragging || !zs.DragSelect {
		return
	}
	drawSelectionRect(ctx, zs, pr, th)
}

// selectionColors returns the fill and border colors for the
// selection rectangle, falling back to defaults when the theme
// fields are zero-valued.
func selectionColors(th *theme.Theme) (gui.Color, gui.Color) {
	fill := th.SelectionFill
	if fill == (gui.Color{}) {
		fill = gui.RGBA(70, 130, 220, 30)
	}
	border := th.SelectionBorder
	if border == (gui.Color{}) {
		border = gui.RGBA(70, 130, 220, 180)
	}
	return fill, border
}

// drawSelectionRect renders the brush-to-zoom selection
// rectangle during a drag-select operation.
func drawSelectionRect(
	ctx *render.Context, zs zoomState, pr plotRect,
	th *theme.Theme,
) {
	x0, x1 := zs.SelX0, zs.SelX1
	y0, y1 := zs.SelY0, zs.SelY1
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	// Clamp to plot area.
	x0 = max(x0, pr.Left)
	x1 = min(x1, pr.Right)
	y0 = max(y0, pr.Top)
	y1 = min(y1, pr.Bottom)

	w := x1 - x0
	h := y1 - y0
	if w <= 0 || h <= 0 {
		return
	}
	fill, border := selectionColors(th)
	ctx.FilledRect(x0, y0, w, h, fill)
	ctx.Rect(x0, y0, w, h, border, 1)
}

// loadAndApplyZoom loads zoom state and applies it to the given
// axes. Returns the zoom state for selection-rect drawing.
// Safe to call with nil window (headless export).
func loadAndApplyZoom(
	w *gui.Window, id string,
	xAxis, yAxis axis.Axis, zoomX, zoomY bool,
) zoomState {
	zs, _ := loadZoomState(w, id)
	if !zs.Zoomed {
		return zs
	}
	ensureOrigBounds(&zs, xAxis, yAxis, zoomX, zoomY)
	applyZoomToAxes(zs, xAxis, yAxis, zoomX, zoomY)
	// nil layout: draw-time save, version tracked via GenerateLayout
	saveZoomState(w, nil, id, zs)
	return zs
}

// clipPolylineToRect clips a flat (x,y) polyline to the plot
// rectangle using Cohen-Sutherland segment clipping. Returns a
// new polyline with correct boundary intersections.
func clipPolylineToRect(
	pts []float32, left, right, top, bottom float32,
) []float32 {
	if len(pts) < 4 {
		return pts
	}
	out := make([]float32, 0, len(pts))
	for i := 0; i < len(pts)-2; i += 2 {
		cx0, cy0, cx1, cy1, vis := clipSegment(
			pts[i], pts[i+1], pts[i+2], pts[i+3],
			left, right, top, bottom)
		if !vis {
			continue
		}
		// Avoid duplicate when consecutive segments share an
		// endpoint.
		n := len(out)
		if n == 0 || out[n-2] != cx0 || out[n-1] != cy0 {
			out = append(out, cx0, cy0)
		}
		out = append(out, cx1, cy1)
	}
	return out
}

// clipSegment clips a line segment to a rectangle using the
// Cohen-Sutherland algorithm. Returns the clipped endpoints
// and whether any part is visible.
func clipSegment(
	x0, y0, x1, y1, left, right, top, bottom float32,
) (float32, float32, float32, float32, bool) {
	const (
		cInside = 0
		cLeft   = 1
		cRight  = 2
		cBottom = 4
		cTop    = 8
	)
	outcode := func(x, y float32) int {
		c := cInside
		if x < left {
			c |= cLeft
		} else if x > right {
			c |= cRight
		}
		if y < top {
			c |= cTop
		} else if y > bottom {
			c |= cBottom
		}
		return c
	}

	c0 := outcode(x0, y0)
	c1 := outcode(x1, y1)
	for {
		if c0|c1 == cInside {
			return x0, y0, x1, y1, true
		}
		if c0&c1 != cInside {
			return 0, 0, 0, 0, false
		}
		c := c0
		if c == cInside {
			c = c1
		}
		var x, y float32
		dx := x1 - x0
		dy := y1 - y0
		switch {
		case c&cTop != 0:
			x = x0 + dx*(top-y0)/dy
			y = top
		case c&cBottom != 0:
			x = x0 + dx*(bottom-y0)/dy
			y = bottom
		case c&cRight != 0:
			y = y0 + dy*(right-x0)/dx
			x = right
		default: // cLeft
			y = y0 + dy*(left-x0)/dx
			x = left
		}
		if c == c0 {
			x0, y0 = x, y
			c0 = outcode(x0, y0)
		} else {
			x1, y1 = x, y
			c1 = outcode(x1, y1)
		}
	}
}

// clipConvexToRect clips a convex polygon to the rectangle
// [left,right] x [top,bottom] using Sutherland-Hodgman.
// Returns the clipped polygon as flat [x0,y0,x1,y1,...] pairs,
// or nil if fully outside. The result is convex and safe for
// fan-triangulation (FilledPolygon).
//
// a and b are caller-owned scratch buffers reused across calls to
// avoid per-call heap allocations in draw loops. Both are grown as
// needed and returned so the caller can keep the larger backing
// arrays for the next call.
func clipConvexToRect(
	pts []float32, left, right, top, bottom float32,
	a, b []float32,
) ([]float32, []float32, []float32) {
	if len(pts) < 6 {
		return nil, a, b
	}
	a = append(a[:0], pts...)
	b = b[:0]

	for edge := range 4 {
		na := len(a) / 2
		if na < 3 {
			return nil, a, b
		}
		b = b[:0]
		px, py := a[(na-1)*2], a[(na-1)*2+1]
		pIn := rectEdgeInside(edge, px, py, left, right, top, bottom)

		for i := range na {
			cx, cy := a[i*2], a[i*2+1]
			cIn := rectEdgeInside(edge, cx, cy, left, right, top, bottom)
			if cIn {
				if !pIn {
					ix, iy := edgeIsect(edge, px, py, cx, cy,
						left, right, top, bottom)
					b = append(b, ix, iy)
				}
				b = append(b, cx, cy)
			} else if pIn {
				ix, iy := edgeIsect(edge, px, py, cx, cy,
					left, right, top, bottom)
				b = append(b, ix, iy)
			}
			px, py = cx, cy
			pIn = cIn
		}
		a, b = b, a
	}
	if len(a) < 6 {
		return nil, a, b
	}
	return a, a, b
}

func rectEdgeInside(
	edge int, x, y, left, right, top, bottom float32,
) bool {
	switch edge {
	case 0:
		return x >= left
	case 1:
		return x <= right
	case 2:
		return y >= top
	default:
		return y <= bottom
	}
}

func edgeIsect(
	edge int, ax, ay, bx, by, left, right, top, bottom float32,
) (float32, float32) {
	switch edge {
	case 0:
		t := (left - ax) / (bx - ax)
		return left, ay + t*(by-ay)
	case 1:
		t := (right - ax) / (bx - ax)
		return right, ay + t*(by-ay)
	case 2:
		t := (top - ay) / (by - ay)
		return ax + t*(bx-ax), top
	default:
		t := (bottom - ay) / (by - ay)
		return ax + t*(bx-ax), bottom
	}
}

// --- Internal helpers ---

// ensureOrigBounds stores the original axis domain if not
// already captured and initializes the zoom domain.
func ensureOrigBounds(
	zs *zoomState, xAxis, yAxis axis.Axis,
	zoomX, zoomY bool,
) {
	if zs.OrigStored {
		return
	}
	if zoomX && xAxis != nil {
		zs.OrigXMin, zs.OrigXMax = xAxis.Domain()
	}
	if zoomY && yAxis != nil {
		zs.OrigYMin, zs.OrigYMax = yAxis.Domain()
	}
	zs.OrigStored = true
	if !zs.Zoomed {
		zs.XMin, zs.XMax = zs.OrigXMin, zs.OrigXMax
		zs.YMin, zs.YMax = zs.OrigYMin, zs.OrigYMax
	}
}

// clampRectToPlot clamps a filled rectangle to the plot area.
// Returns the adjusted (x, y, w, h) and whether any visible
// area remains.
func clampRectToPlot(
	x, y, w, h, pLeft, pRight, pTop, pBottom float32,
) (float32, float32, float32, float32, bool) {
	x1 := x + w
	y1 := y + h
	x = max(x, pLeft)
	y = max(y, pTop)
	x1 = min(x1, pRight)
	y1 = min(y1, pBottom)
	w = x1 - x
	h = y1 - y
	if w <= 0 || h <= 0 {
		return 0, 0, 0, 0, false
	}
	return x, y, w, h, true
}

// clampVerticalLine clamps a vertical line segment to the plot
// area Y bounds. Returns the clamped (y0, y1) and whether any
// visible portion remains.
func clampVerticalLine(
	y0, y1, pTop, pBottom float32,
) (float32, float32, bool) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	y0 = max(y0, pTop)
	y1 = min(y1, pBottom)
	return y0, y1, y0 < y1
}

// insidePlot returns whether a point is inside the plot area.
func insidePlot(
	x, y, pLeft, pRight, pTop, pBottom float32,
) bool {
	return x >= pLeft && x <= pRight && y >= pTop && y <= pBottom
}
