package chart

import (
	"log/slog"
	"math"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// MarkerShape controls the shape of scatter plot markers.
type MarkerShape uint8

// MarkerShape constants.
const (
	MarkerCircle MarkerShape = iota
	MarkerSquare
	MarkerTriangle
	MarkerDiamond
	MarkerCross
)

// ScatterCfg configures a scatter plot.
type ScatterCfg struct {
	BaseCfg
	InteractionCfg

	// Data
	Series      []series.XY
	ErrorSeries []series.ErrorXY

	// Axes (optional; auto-created from series bounds when nil)
	XAxis axis.Axis
	YAxis axis.Axis

	// Appearance
	MarkerSize float32 // 0 means default (6)
	Marker     MarkerShape
}

type scatterView struct {
	cfg ScatterCfg
	xyBase
	lastVersion uint64
	xAxis       axis.Axis
	yAxis       axis.Axis
	xTicks      []axis.Tick
	yTicks      []axis.Tick
}

// Scatter creates a scatter plot view.
func Scatter(cfg ScatterCfg) gui.View {
	cfg.applyDefaults()
	if cfg.MarkerSize == 0 {
		cfg.MarkerSize = DefaultMarkerSize
	}
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableXY(&cfg.BaseCfg, cfg.Series)
	}
	sv := &scatterView{cfg: cfg}
	sv.base = &sv.cfg.BaseCfg
	sv.interaction = &sv.cfg.InteractionCfg
	sv.zoomX = true
	sv.zoomY = true
	sv.nearestFn = func(px, py float32) bool {
		if sv.lastPA.XAxis == nil {
			return false
		}
		_, _, _, _, ok := nearestXYPoint(
			sv.cfg.Series, sv.lastPA, px, py, 20)
		if !ok {
			_, _, _, _, ok = nearestErrorXYPoint(
				sv.cfg.ErrorSeries, sv.lastPA, px, py, 20)
		}
		return ok
	}
	return sv
}

// Draw renders the chart onto dc for headless export.
func (sv *scatterView) Draw(dc *gui.DrawContext) { sv.draw(dc) }

func (sv *scatterView) chartTheme() *theme.Theme { return sv.cfg.Theme }

func (sv *scatterView) Content() []gui.View { return nil }

func (sv *scatterView) GenerateLayout(w *gui.Window) gui.Layout {
	return sv.generateLayout(w, sv.draw)
}

// updateAxes recomputes axes from config or series bounds.
// Returns false if bounds are invalid.
func (sv *scatterView) updateAxes() bool {
	cfg := &sv.cfg

	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64

	for _, s := range cfg.Series {
		if s.Len() == 0 {
			continue
		}
		sx0, sx1, sy0, sy1 := s.Bounds()
		minX = min(minX, sx0)
		maxX = max(maxX, sx1)
		minY = min(minY, sy0)
		maxY = max(maxY, sy1)
	}
	for _, s := range cfg.ErrorSeries {
		if s.Len() == 0 {
			continue
		}
		sx0, sx1, sy0, sy1 := s.Bounds()
		minX = min(minX, sx0)
		maxX = max(maxX, sx1)
		minY = min(minY, sy0)
		maxY = max(maxY, sy1)
	}

	hasBounds := minX <= maxX
	if hasBounds && (!finite(minX) || !finite(maxX) ||
		!finite(minY) || !finite(maxY)) {
		slog.Warn("non-finite bounds", "chart", cfg.ID)
		return false
	}

	var ok bool
	sv.xAxis, ok = autoLinearAxis(cfg.XAxis, minX, maxX, 0.05, cfg.ID)
	if !ok {
		return false
	}
	sv.yAxis, ok = autoLinearAxis(cfg.YAxis, minY, maxY, 0.05, cfg.ID)
	if !ok {
		return false
	}
	sv.lastVersion = cfg.Version
	return true
}

func (sv *scatterView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &sv.cfg
	th := cfg.Theme

	if len(cfg.Series) == 0 && len(cfg.ErrorSeries) == 0 {
		slog.Warn("no series data", "chart", cfg.ID)
		return
	}

	left := th.PaddingLeft
	right := ctx.Width() - th.PaddingRight
	top := th.PaddingTop
	bottom := ctx.Height() - th.PaddingBottom

	names := make([]string, len(cfg.Series)+len(cfg.ErrorSeries))
	for i, s := range cfg.Series {
		names[i] = s.Name()
	}
	for i, s := range cfg.ErrorSeries {
		names[len(cfg.Series)+i] = s.Name()
	}
	right -= legendRightReserve(ctx, th, cfg.LegendPosition, names)
	top += legendTopReserve(ctx, th, cfg.LegendPosition, names, left, right)

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	drawTitle(ctx, cfg.Title, th)

	if sv.xAxis == nil || cfg.Version != sv.lastVersion {
		if !sv.updateAxes() {
			return
		}
	}

	xAxis := sv.xAxis
	yAxis := sv.yAxis

	zs := loadAndApplyZoom(sv.win, sv.cfg.ID, xAxis, yAxis, true, true)

	left = resolveLeft(ctx, th, left, bottom, top, yAxis)

	bottom = ctx.Height() - resolveBottom(ctx, th,
		maxTickLabelWidth(ctx, xAxis.Ticks(left, right), th.TickStyle),
		cfg.XTickRotation, xAxis.Label())
	bottom -= legendBottomReserve(ctx, th, cfg.LegendPosition, names, left, right)

	sv.yTicks = yAxis.Ticks(bottom, top)
	sv.xTicks = xAxis.Ticks(left, right)

	for _, t := range sv.yTicks {
		ctx.Line(left, t.Position, right, t.Position,
			th.GridColor, th.GridWidth)
	}
	for _, t := range sv.xTicks {
		ctx.Line(t.Position, top, t.Position, bottom,
			th.GridColor, th.GridWidth)
	}

	ctx.Line(left, bottom, right, bottom, th.AxisColor, th.AxisWidth)
	ctx.Line(left, top, left, bottom, th.AxisColor, th.AxisWidth)

	tickLen, tickWidth, tickColor := resolvedTickMark(th)
	tickStyle := th.TickStyle
	fh := ctx.FontHeight(tickStyle)
	for _, t := range sv.xTicks {
		ctx.Line(t.Position, bottom, t.Position, bottom+tickLen,
			tickColor, tickWidth)
		lw := ctx.TextWidth(t.Label, tickStyle)
		ctx.Text(t.Position-lw/2, bottom+tickLen+2, t.Label, tickStyle)
	}
	for _, t := range sv.yTicks {
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
	sv.lastPA = plotArea{plotRect{left, right, top, bottom}, xAxis, yAxis}

	// Hover highlight: find nearest series/point across both
	// XY and ErrorXY series. Indices >= len(cfg.Series) refer
	// to ErrorSeries.
	hovSI := -1
	var hovPx, hovPy float32
	if sv.hovering && xAxis != nil {
		pa := sv.lastPA
		si, _, px, py, snapOK := nearestXYPoint(
			cfg.Series, pa, sv.hoverPx, sv.hoverPy, 20)
		if snapOK {
			hovSI, hovPx, hovPy = si, px, py
		}
		esi, _, epx, epy, eOK := nearestErrorXYPoint(
			cfg.ErrorSeries, pa, sv.hoverPx, sv.hoverPy, 20)
		if eOK {
			dx0 := hovPx - sv.hoverPx
			dy0 := hovPy - sv.hoverPy
			dx1 := epx - sv.hoverPx
			dy1 := epy - sv.hoverPy
			if !snapOK || dx1*dx1+dy1*dy1 < dx0*dx0+dy0*dy0 {
				hovSI = len(cfg.Series) + esi
				hovPx, hovPy = epx, epy
			}
		}
	}

	progress := animProgress(sv.win, cfg.ID)

	sv.drawMarkers(ctx, cfg, xAxis, yAxis,
		left, right, top, bottom, hovSI, hovPx, hovPy, progress)

	nBase := len(cfg.Series)
	entries := make([]legendEntry, nBase+len(cfg.ErrorSeries))
	for i, s := range cfg.Series {
		entries[i] = legendEntry{
			Name:  s.Name(),
			Color: seriesColor(s.Color(), i, th.Palette),
			Index: i,
		}
	}
	for i, s := range cfg.ErrorSeries {
		idx := nBase + i
		entries[idx] = legendEntry{
			Name:  s.Name(),
			Color: seriesColor(s.Color(), idx, th.Palette),
			Index: idx,
		}
	}
	pr := plotRect{left, right, top, bottom}
	sv.lastLB = drawLegend(ctx, entries, th, pr,
		cfg.LegendPosition, sv.hidden)
	saveLegendBounds(sv.win, cfg.ID, sv.lastLB)

	drawSelectionRectIf(ctx, zs, pr, th)

	// Crosshair and tooltip.
	if sv.hovering && sv.xAxis != nil {
		drawCrosshair(ctx, th, sv.hoverPx, sv.hoverPy, pr)
		pa := plotArea{pr, xAxis, yAxis}
		drawXYTooltip(ctx, th, cfg.Series, pa,
			sv.hoverPx, sv.hoverPy)
		drawErrorXYTooltip(ctx, th, cfg.ErrorSeries, pa,
			sv.hoverPx, sv.hoverPy)
	}
}

// drawMarker renders a single marker at (cx, cy) with the given size and shape.
func drawMarker(
	ctx *render.Context, cx, cy, size float32,
	shape MarkerShape, color gui.Color,
) {
	h := size / 2
	switch shape {
	case MarkerSquare:
		ctx.FilledRect(cx-h, cy-h, size, size, color)
	case MarkerTriangle:
		// Equilateral triangle pointing up.
		pts := [6]float32{
			cx, cy - h,
			cx + h, cy + h,
			cx - h, cy + h,
		}
		ctx.FilledPolygon(pts[:], color)
	case MarkerDiamond:
		pts := [8]float32{
			cx, cy - h,
			cx + h, cy,
			cx, cy + h,
			cx - h, cy,
		}
		ctx.FilledPolygon(pts[:], color)
	case MarkerCross:
		w := size / 4
		ctx.Line(cx-h, cy, cx+h, cy, color, w)
		ctx.Line(cx, cy-h, cx, cy+h, color, w)
	default: // MarkerCircle
		ctx.FilledCircle(cx, cy, h, color)
	}
}

// errorBarCapSize is the half-width of error bar caps in pixels.
const errorBarCapSize float32 = 4

// drawErrorBars renders X and Y whiskers with caps for a single
// ErrorPoint. Zero-valued ErrorBar fields are skipped.
func drawErrorBars(
	ctx *render.Context,
	px, py float32,
	p series.ErrorPoint,
	xAxis, yAxis axis.Axis,
	left, right, top, bottom float32,
	color gui.Color,
) {
	w := 1 / ctx.Scale()
	yLo, yHi := safeErr(p.YErr)
	if yLo > 0 || yHi > 0 {
		pyLo := yAxis.Transform(p.Y-yLo, bottom, top)
		pyHi := yAxis.Transform(p.Y+yHi, bottom, top)
		pyLo = clampF32(pyLo, top, bottom)
		pyHi = clampF32(pyHi, top, bottom)
		ctx.Line(px, pyLo, px, pyHi, color, w)
		ctx.Line(px-errorBarCapSize, pyLo, px+errorBarCapSize, pyLo, color, w)
		ctx.Line(px-errorBarCapSize, pyHi, px+errorBarCapSize, pyHi, color, w)
	}
	xLo, xHi := safeErr(p.XErr)
	if xLo > 0 || xHi > 0 {
		pxLo := xAxis.Transform(p.X-xLo, left, right)
		pxHi := xAxis.Transform(p.X+xHi, left, right)
		pxLo = clampF32(pxLo, left, right)
		pxHi = clampF32(pxHi, left, right)
		ctx.Line(pxLo, py, pxHi, py, color, w)
		ctx.Line(pxLo, py-errorBarCapSize, pxLo, py+errorBarCapSize, color, w)
		ctx.Line(pxHi, py-errorBarCapSize, pxHi, py+errorBarCapSize, color, w)
	}
}

// safeErr returns finite, non-negative low/high error widths.
// NaN, Inf, and negatives become 0 so downstream rendering
// stays bounded.
func safeErr(e series.ErrorBar) (lo, hi float64) {
	if finite(e.Low) && e.Low > 0 {
		lo = e.Low
	}
	if finite(e.High) && e.High > 0 {
		hi = e.High
	}
	return
}

// clampF32 clamps v to [lo, hi]. Non-finite v becomes lo.
func clampF32(v, lo, hi float32) float32 {
	if v != v || v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// drawMarkers renders scatter markers and the hovered highlight,
// skipping points outside the plot area.
func (sv *scatterView) drawMarkers(
	ctx *render.Context, cfg *ScatterCfg,
	xAxis, yAxis axis.Axis,
	left, right, top, bottom float32,
	hovSI int, hovPx, hovPy float32, progress float32,
) {
	th := cfg.Theme
	markerR := cfg.MarkerSize * progress
	for i, s := range cfg.Series {
		if s.Len() == 0 || sv.hidden[i] {
			continue
		}
		color := seriesColor(s.Color(), i, th.Palette)
		if hovSI >= 0 && i != hovSI {
			color = dimColor(color, HoverDimAlpha)
		}
		for _, p := range s.Points {
			if !finite(p.X) || !finite(p.Y) {
				continue
			}
			px := xAxis.Transform(p.X, left, right)
			py := yAxis.Transform(p.Y, bottom, top)
			if px < left || px > right || py < top || py > bottom {
				continue
			}
			drawMarker(ctx, px, py, markerR, cfg.Marker, color)
		}
	}
	// Error bar series: whiskers + caps + center marker.
	nBase := len(cfg.Series)
	for i, s := range cfg.ErrorSeries {
		if s.Len() == 0 {
			continue
		}
		idx := nBase + i
		color := seriesColor(s.Color(), idx, th.Palette)
		if hovSI >= 0 && idx != hovSI {
			color = dimColor(color, HoverDimAlpha)
		}
		for _, p := range s.Points {
			if !finite(p.X) || !finite(p.Y) {
				continue
			}
			px := xAxis.Transform(p.X, left, right)
			py := yAxis.Transform(p.Y, bottom, top)
			if px < left || px > right || py < top || py > bottom {
				continue
			}
			drawErrorBars(ctx, px, py, p, xAxis, yAxis,
				left, right, top, bottom, color)
			drawMarker(ctx, px, py, markerR, cfg.Marker, color)
		}
	}

	if hovSI >= 0 && !sv.hidden[hovSI] &&
		hovPx >= left && hovPx <= right &&
		hovPy >= top && hovPy <= bottom {
		hc := hoverSeriesColor(cfg, hovSI, th.Palette)
		drawMarker(ctx, hovPx, hovPy, cfg.MarkerSize*2, cfg.Marker, hc)
	}
}

// hoverSeriesColor returns the palette color for the hovered
// series index, mapping indices >= len(Series) to ErrorSeries.
func hoverSeriesColor(
	cfg *ScatterCfg, idx int, palette []gui.Color,
) gui.Color {
	nBase := len(cfg.Series)
	if idx < nBase {
		return seriesColor(cfg.Series[idx].Color(), idx, palette)
	}
	ei := idx - nBase
	return seriesColor(cfg.ErrorSeries[ei].Color(), idx, palette)
}
