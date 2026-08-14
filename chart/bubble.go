package chart

import (
	"cmp"
	"log/slog"
	"math"
	"slices"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// BubbleCfg configures a bubble chart (scatter with sized markers).
type BubbleCfg struct {
	BaseCfg
	InteractionCfg

	// Data
	Series []series.XYZ

	// Axes (optional; auto-created from series bounds when nil)
	XAxis axis.Axis
	YAxis axis.Axis

	// Appearance
	MinRadius float32     // minimum marker radius; 0 means default (4)
	MaxRadius float32     // maximum marker radius; 0 means default (30)
	Marker    MarkerShape // default marker shape (circle)
	// Markers overrides the marker shape per series. When
	// Markers[i] is set it takes precedence over Marker for
	// series i. Nil or short slices fall back to Marker.
	Markers []MarkerShape
}

type bubbleView struct {
	cfg BubbleCfg
	xyBase
	lastVersion uint64
	xAxis       axis.Axis
	yAxis       axis.Axis
	xTicks      []axis.Tick
	yTicks      []axis.Tick
	zMin, zMax  float64
}

// Bubble creates a bubble chart view.
func Bubble(cfg BubbleCfg) gui.View {
	cfg.applyDefaults()
	if cfg.MinRadius == 0 {
		cfg.MinRadius = DefaultBubbleMinRadius
	}
	if cfg.MaxRadius == 0 {
		cfg.MaxRadius = DefaultBubbleMaxRadius
	}
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableXYZ(&cfg.BaseCfg, cfg.Series)
	}
	bv := &bubbleView{cfg: cfg}
	bv.base = &bv.cfg.BaseCfg
	bv.interaction = &bv.cfg.InteractionCfg
	bv.zoomX = true
	bv.zoomY = true
	bv.nearestFn = func(px, py float32) bool {
		if bv.lastPA.XAxis == nil {
			return false
		}
		_, _, _, _, ok := nearestXYZPoint(
			bv.cfg.Series, bv.lastPA, px, py, bv.cfg.MaxRadius+5)
		return ok
	}
	return bv
}

// Draw renders the chart onto dc for headless export.
func (bv *bubbleView) Draw(dc *gui.DrawContext) { bv.draw(dc) }

func (bv *bubbleView) chartTheme() *theme.Theme { return bv.cfg.Theme }

func (bv *bubbleView) GenerateLayout(w *gui.Window) gui.Layout {
	return bv.generateLayout(w, bv.draw)
}

// updateAxes recomputes axes from config or series bounds.
// Returns false if bounds are invalid.
func (bv *bubbleView) updateAxes() bool {
	cfg := &bv.cfg

	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64
	bv.zMin, bv.zMax = math.MaxFloat64, -math.MaxFloat64

	for _, s := range cfg.Series {
		if s.Len() == 0 {
			continue
		}
		sx0, sx1, sy0, sy1 := s.Bounds()
		minX = min(minX, sx0)
		maxX = max(maxX, sx1)
		minY = min(minY, sy0)
		maxY = max(maxY, sy1)
		sz0, sz1 := s.ZBounds()
		if finite(sz0) && finite(sz1) {
			bv.zMin = min(bv.zMin, sz0)
			bv.zMax = max(bv.zMax, sz1)
		}
	}

	hasBounds := minX <= maxX
	if hasBounds && (!finite(minX) || !finite(maxX) ||
		!finite(minY) || !finite(maxY)) {
		slog.Warn("non-finite bounds", "chart", cfg.ID)
		return false
	}

	var ok bool
	bv.xAxis, ok = autoLinearAxis(cfg.XAxis, minX, maxX, 0.05, cfg.ID)
	if !ok {
		return false
	}
	bv.yAxis, ok = autoLinearAxis(cfg.YAxis, minY, maxY, 0.05, cfg.ID)
	if !ok {
		return false
	}
	bv.lastVersion = cfg.Version
	return true
}

func (bv *bubbleView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &bv.cfg
	th := cfg.Theme

	if len(cfg.Series) == 0 {
		slog.Warn("no series data", "chart", cfg.ID)
		return
	}

	left := th.PaddingLeft
	right := ctx.Width() - th.PaddingRight
	top := th.PaddingTop
	bottom := ctx.Height() - th.PaddingBottom

	names := make([]string, len(cfg.Series))
	for i, s := range cfg.Series {
		names[i] = s.Name()
	}
	right -= legendRightReserve(ctx, th, cfg.LegendPosition, names)
	top += legendTopReserve(ctx, th, cfg.LegendPosition, names, left, right)

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	drawTitle(ctx, cfg.Title, th)

	if bv.xAxis == nil || cfg.Version != bv.lastVersion {
		if !bv.updateAxes() {
			return
		}
	}

	xAxis := bv.xAxis
	yAxis := bv.yAxis

	zs := loadAndApplyZoom(bv.win, bv.cfg.ID, xAxis, yAxis, true, true)

	left = resolveLeft(ctx, th, left, bottom, top, yAxis)

	bottom = ctx.Height() - resolveBottom(ctx, th,
		maxTickLabelWidth(ctx, xAxis.Ticks(left, right), th.TickStyle),
		cfg.XTickRotation, xAxis.Label())
	bottom -= legendBottomReserve(ctx, th, cfg.LegendPosition, names, left, right)

	bv.yTicks = yAxis.Ticks(bottom, top)
	bv.xTicks = xAxis.Ticks(left, right)

	for _, t := range bv.yTicks {
		ctx.Line(left, t.Position, right, t.Position,
			th.GridColor, th.GridWidth)
	}
	for _, t := range bv.xTicks {
		ctx.Line(t.Position, top, t.Position, bottom,
			th.GridColor, th.GridWidth)
	}

	ctx.Line(left, bottom, right, bottom, th.AxisColor, th.AxisWidth)
	ctx.Line(left, top, left, bottom, th.AxisColor, th.AxisWidth)

	tickLen, tickWidth, tickColor := resolvedTickMark(th)
	tickStyle := th.TickStyle
	fh := ctx.FontHeight(tickStyle)
	for _, t := range bv.xTicks {
		ctx.Line(t.Position, bottom, t.Position, bottom+tickLen,
			tickColor, tickWidth)
		lw := ctx.TextWidth(t.Label, tickStyle)
		ctx.Text(t.Position-lw/2, bottom+tickLen+2, t.Label, tickStyle)
	}
	for _, t := range bv.yTicks {
		ctx.Line(left-tickLen, t.Position, left, t.Position,
			tickColor, tickWidth)
		tw := ctx.TextWidth(t.Label, tickStyle)
		ctx.Text(left-tickLen-tw-2, t.Position-fh/2, t.Label, tickStyle)
	}

	drawXAxisLabel(ctx, xAxis.Label(), th, left, right, bottom)
	drawYAxisLabel(ctx, yAxis.Label(), th, top, bottom)

	// Annotations.
	drawAnnotations(ctx, &cfg.Annotations, th,
		plotRect{left, right, top, bottom}, xAxis, yAxis)

	// Cache plot area for cursor hit-testing in hover callback.
	bv.lastPA = plotArea{plotRect{left, right, top, bottom}, xAxis, yAxis}

	// Hover highlight: find nearest series/point.
	hovSI := -1
	hovPI := -1
	var hovPx, hovPy float32
	snapPx := cfg.MaxRadius + 5
	if bv.hovering && xAxis != nil {
		pa := bv.lastPA
		si, pi, px, py, snapOK := nearestXYZPoint(
			cfg.Series, pa, bv.hoverPx, bv.hoverPy, snapPx)
		if snapOK {
			hovSI, hovPI, hovPx, hovPy = si, pi, px, py
		}
	}

	bv.drawBubbles(ctx, cfg, xAxis, yAxis,
		left, right, top, bottom, hovSI, hovPI, hovPx, hovPy)

	entries := make([]legendEntry, len(cfg.Series))
	for i, s := range cfg.Series {
		entries[i] = legendEntry{
			Name:  s.Name(),
			Color: seriesColor(s.Color(), i, th.Palette),
			Index: i,
		}
	}
	pr := plotRect{left, right, top, bottom}
	bv.lastLB = drawLegend(ctx, entries, th, pr,
		cfg.LegendPosition, bv.hidden)
	saveLegendBounds(bv.win, cfg.ID, bv.lastLB)

	drawSelectionRectIf(ctx, zs, pr, th)

	// Crosshair and tooltip.
	if bv.hovering && bv.xAxis != nil {
		drawCrosshair(ctx, th, bv.hoverPx, bv.hoverPy, pr)
		pa := plotArea{pr, xAxis, yAxis}
		drawXYZTooltip(ctx, th, cfg.Series, pa,
			bv.hoverPx, bv.hoverPy, cfg.MaxRadius+5)
	}
}

// bubbleRadius maps a Z value to a pixel radius using sqrt scaling
// so that marker area is proportional to Z.
func bubbleRadius(z, zMin, zMax float64, minR, maxR float32) float32 {
	if zMax <= zMin {
		return (minR + maxR) / 2
	}
	frac := (z - zMin) / (zMax - zMin)
	frac = max(0, min(1, frac))
	t := math.Sqrt(frac)
	return minR + (maxR-minR)*float32(t)
}

// bubbleMarkerShape returns the marker shape for series index i,
// using Markers[i] when available, otherwise the default Marker.
func bubbleMarkerShape(cfg *BubbleCfg, i int) MarkerShape {
	if i < len(cfg.Markers) {
		return cfg.Markers[i]
	}
	return cfg.Marker
}

// bubbleDrawItem holds the data needed to draw a single bubble,
// used to sort by Z descending so smaller bubbles draw on top.
type bubbleDrawItem struct {
	px, py float32
	radius float32
	color  gui.Color
	shape  MarkerShape
}

// drawBubbles renders bubble markers sorted by Z descending and
// the hovered highlight, skipping points outside the plot area.
func (bv *bubbleView) drawBubbles(
	ctx *render.Context, cfg *BubbleCfg,
	xAxis, yAxis axis.Axis,
	left, right, top, bottom float32,
	hovSI, hovPI int, hovPx, hovPy float32,
) {
	th := cfg.Theme
	zMin, zMax := bv.zMin, bv.zMax

	// Pre-compute capacity for the items slice.
	totalPts := 0
	for i, s := range cfg.Series {
		if s.Len() > 0 && !bv.hidden[i] {
			totalPts += s.Len()
		}
	}

	progress := animProgress(bv.win, bv.cfg.ID)

	// Collect all visible bubbles for Z-sorted drawing.
	items := make([]bubbleDrawItem, 0, totalPts)
	for i, s := range cfg.Series {
		if s.Len() == 0 || bv.hidden[i] {
			continue
		}
		color := seriesColor(s.Color(), i, th.Palette)
		if hovSI >= 0 && i != hovSI {
			color = dimColor(color, HoverDimAlpha)
		}
		shape := bubbleMarkerShape(cfg, i)
		for _, p := range s.Points {
			if !finite(p.X) || !finite(p.Y) || !finite(p.Z) {
				continue
			}
			px := xAxis.Transform(p.X, left, right)
			py := yAxis.Transform(p.Y, bottom, top)
			if px < left || px > right || py < top || py > bottom {
				continue
			}
			r := bubbleRadius(p.Z, zMin, zMax,
				cfg.MinRadius, cfg.MaxRadius) * progress
			items = append(items, bubbleDrawItem{px, py, r, color, shape})
		}
	}

	// Sort Z descending (largest first) so smaller bubbles
	// draw on top.
	slices.SortFunc(items, func(a, b bubbleDrawItem) int {
		return cmp.Compare(b.radius, a.radius)
	})

	for _, it := range items {
		drawMarker(ctx, it.px, it.py, it.radius*2,
			it.shape, it.color)
	}

	// Hover highlight: draw hovered point at 1.3x radius.
	if hovSI >= 0 && hovPI >= 0 && !bv.hidden[hovSI] &&
		hovPx >= left && hovPx <= right &&
		hovPy >= top && hovPy <= bottom {
		hc := seriesColor(cfg.Series[hovSI].Color(), hovSI, th.Palette)
		hovShape := bubbleMarkerShape(cfg, hovSI)
		z := cfg.Series[hovSI].Points[hovPI].Z
		r := bubbleRadius(z, zMin, zMax,
			cfg.MinRadius, cfg.MaxRadius)
		drawMarker(ctx, hovPx, hovPy, r*2*1.3,
			hovShape, hc)
	}
}
