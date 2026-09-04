package chart

import (
	"cmp"
	"log/slog"
	"math"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// AreaCfg configures an area chart.
type AreaCfg struct {
	BaseCfg
	InteractionCfg

	// Data
	Series []series.XY

	// Axes (optional; auto-created from series bounds when nil)
	XAxis axis.Axis
	YAxis axis.Axis

	// Appearance
	Stacked   bool
	LineWidth float32 // 0 means default (2)
	Opacity   float32 // fill opacity 0-1; 0 means default (0.3)

	// AutoScroll enables smooth scrolling to follow latest
	// data. Typically used with RealTimeSeries.
	AutoScroll bool
	// WindowSize is the visible X-axis range when AutoScroll
	// is enabled. Zero shows all data.
	WindowSize float64
}

type areaView struct {
	cfg AreaCfg
	xyBase
	lastVersion uint64
	xAxis       axis.Axis
	yAxis       axis.Axis
	xTicks      []axis.Tick
	yTicks      []axis.Tick
	ptsBuf      []float32
	prevPtsBuf  []float32
	curPtsBuf   []float32
	clipA       []float32 // scratch for clipConvexToRect
	clipB       []float32
}

// Area creates an area chart view.
func Area(cfg AreaCfg) gui.View {
	cfg.applyDefaults()
	cfg.LineWidth = cmp.Or(cfg.LineWidth, DefaultLineWidth)
	cfg.Opacity = cmp.Or(cfg.Opacity, DefaultAreaOpacity)
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableXY(&cfg.BaseCfg, cfg.Series)
	}
	av := &areaView{cfg: cfg}
	av.base = &av.cfg.BaseCfg
	av.interaction = &av.cfg.InteractionCfg
	av.zoomX = true
	av.zoomY = true
	av.extraVersionFn = func(w *gui.Window) uint64 {
		return loadScrollVersion(w, av.cfg.ID)
	}
	av.nearestFn = func(px, py float32) bool {
		if av.lastPA.XAxis == nil {
			return false
		}
		if av.cfg.Stacked {
			_, _, _, _, ok := nearestStackedPoint(
				av.cfg.Series, av.lastPA, px, py, 20)
			return ok
		}
		_, _, _, _, ok := nearestXYPoint(
			av.cfg.Series, av.lastPA, px, py, 20)
		return ok
	}
	return av
}

// Draw renders the chart onto dc for headless export.
func (av *areaView) Draw(dc *gui.DrawContext) { av.draw(dc) }

func (av *areaView) chartTheme() *theme.Theme { return av.cfg.Theme }

func (av *areaView) GenerateLayout(w *gui.Window) gui.Layout {
	return av.generateLayout(w, av.draw)
}

// updateAxes recomputes axes from config or series bounds.
// Returns false if bounds are invalid (empty or non-finite).
func (av *areaView) updateAxes() bool {
	cfg := &av.cfg

	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64

	for _, s := range cfg.Series {
		if s.Len() == 0 {
			continue
		}
		sx0, sx1, sy0, sy1 := s.Bounds()
		minX = min(minX, sx0)
		maxX = max(maxX, sx1)
		if !cfg.Stacked {
			minY = min(minY, sy0)
			maxY = max(maxY, sy1)
		}
	}

	// For stacked mode, Y range is the cumulative sum envelope.
	if cfg.Stacked {
		refLen := 0
		for _, s := range cfg.Series {
			if s.Len() > 0 {
				refLen = s.Len()
				break
			}
		}
		if refLen > 0 {
			sums := make([]float64, refLen)
			for _, s := range cfg.Series {
				n := min(s.Len(), refLen)
				for j := range n {
					sums[j] += s.Points[j].Y
					maxY = max(maxY, sums[j])
					minY = min(minY, sums[j])
				}
			}
			minY = min(minY, 0)
		}
	}

	hasBounds := minX <= maxX
	if hasBounds && (!finite(minX) || !finite(maxX) ||
		!finite(minY) || !finite(maxY)) {
		slog.Warn("non-finite bounds", "chart", cfg.ID)
		return false
	}

	var ok bool
	av.xAxis, ok = autoLinearAxis(cfg.XAxis, minX, maxX, 0, cfg.ID)
	if !ok {
		return false
	}
	// Stacked Y: only pad the top (minY is already ≥ 0).
	// Non-stacked: symmetric 5% padding.
	if cfg.Stacked && hasBounds {
		r := maxY - minY
		if r == 0 || math.IsInf(r, 0) {
			r = 1
		}
		av.yAxis, ok = autoLinearAxis(cfg.YAxis, minY, maxY+r*0.05, 0, cfg.ID)
	} else {
		av.yAxis, ok = autoLinearAxis(cfg.YAxis, minY, maxY, 0.05, cfg.ID)
	}
	if !ok {
		return false
	}
	av.lastVersion = cfg.Version
	return true
}

func (av *areaView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &av.cfg
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

	if av.xAxis == nil || cfg.Version != av.lastVersion {
		if !av.updateAxes() {
			return
		}
	}

	xAxis := av.xAxis
	yAxis := av.yAxis

	zs := loadAndApplyZoom(av.win, av.cfg.ID, xAxis, yAxis, true, true)
	applyAutoScroll(av.win, cfg.ID, cfg.AutoScroll,
		cfg.WindowSize, zs.Zoomed, cfg.Series, xAxis)

	left = resolveLeft(ctx, th, left, bottom, top, yAxis)

	bottom = ctx.Height() - resolveBottom(ctx, th,
		maxTickLabelWidth(ctx, xAxis.Ticks(left, right), th.TickStyle),
		cfg.XTickRotation, xAxis.Label())
	bottom -= legendBottomReserve(ctx, th, cfg.LegendPosition, names, left, right)

	av.yTicks = yAxis.Ticks(bottom, top)
	av.xTicks = xAxis.Ticks(left, right)

	for _, t := range av.yTicks {
		ctx.Line(left, t.Position, right, t.Position,
			th.GridColor, th.GridWidth)
	}
	for _, t := range av.xTicks {
		ctx.Line(t.Position, top, t.Position, bottom,
			th.GridColor, th.GridWidth)
	}

	ctx.Line(left, bottom, right, bottom, th.AxisColor, th.AxisWidth)
	ctx.Line(left, top, left, bottom, th.AxisColor, th.AxisWidth)

	tickLen, tickWidth, tickColor := resolvedTickMark(th)
	tickStyle := th.TickStyle
	fh := ctx.FontHeight(tickStyle)
	for _, t := range av.xTicks {
		ctx.Line(t.Position, bottom, t.Position, bottom+tickLen,
			tickColor, tickWidth)
		lw := ctx.TextWidth(t.Label, tickStyle)
		ctx.Text(t.Position-lw/2, bottom+tickLen+2, t.Label, tickStyle)
	}
	for _, t := range av.yTicks {
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

	alpha := uint8(cfg.Opacity * 255)

	// Cache plot area for cursor hit-testing in hover callback.
	av.lastPA = plotArea{plotRect{left, right, top, bottom}, xAxis, yAxis}

	// Hover highlight: find nearest series/point.
	// Stacked mode uses cumulative Y values so the hit-test matches the
	// drawn geometry; overlapping mode uses raw Y values.
	hovSI := -1
	var hovPx, hovPy float32
	if av.hovering && xAxis != nil {
		pa := av.lastPA
		if cfg.Stacked {
			si, _, px, py, snapOK := nearestStackedPoint(
				cfg.Series, pa, av.hoverPx, av.hoverPy, 20)
			if snapOK {
				hovSI, hovPx, hovPy = si, px, py
			}
		} else {
			si, _, px, py, snapOK := nearestXYPoint(
				cfg.Series, pa, av.hoverPx, av.hoverPy, 20)
			if snapOK {
				hovSI, hovPx, hovPy = si, px, py
			}
		}
	}

	progress := animProgress(av.win, cfg.ID)

	if cfg.Stacked {
		av.drawStacked(ctx, cfg, xAxis, yAxis, left, right, top, bottom, alpha, hovSI, progress)
	} else {
		av.drawOverlapping(ctx, cfg, xAxis, yAxis, left, right, top, bottom, alpha, hovSI, progress)
	}

	entries := make([]legendEntry, len(cfg.Series))
	for i, s := range cfg.Series {
		entries[i] = legendEntry{
			Name:  s.Name(),
			Color: seriesColor(s.Color(), i, th.Palette),
			Index: i,
		}
	}
	pr := plotRect{left, right, top, bottom}
	av.lastLB = drawLegend(ctx, entries, th, pr,
		cfg.LegendPosition, av.hidden)
	saveLegendBounds(av.win, cfg.ID, av.lastLB)

	// Enlarged point marker on hovered series (only if inside plot).
	if hovSI >= 0 && !av.hidden[hovSI] &&
		hovPx >= left && hovPx <= right &&
		hovPy >= top && hovPy <= bottom {
		hc := seriesColor(cfg.Series[hovSI].Color(), hovSI, th.Palette)
		ctx.FilledCircle(hovPx, hovPy, cfg.LineWidth*4, hc)
	}

	drawSelectionRectIf(ctx, zs, pr, th)

	// Crosshair and tooltip.
	if av.hovering && av.xAxis != nil {
		drawCrosshair(ctx, th, av.hoverPx, av.hoverPy, pr)
		pa := plotArea{pr, xAxis, yAxis}
		xName, yName := cfg.tooltipAxisNames()
		if cfg.Stacked {
			drawStackedXYTooltip(ctx, th, cfg.Series, pa,
				av.hoverPx, av.hoverPy, xName, yName)
		} else {
			drawXYTooltip(ctx, th, cfg.Series, pa,
				av.hoverPx, av.hoverPy, xName, yName)
		}
	}
}

func (av *areaView) drawOverlapping(
	ctx *render.Context, cfg *AreaCfg,
	xAxis, yAxis axis.Axis,
	left, right, top, bottom float32,
	alpha uint8, hovSI int, progress float32,
) {
	for i, s := range cfg.Series {
		n := s.Len()
		if n == 0 || av.hidden[i] {
			continue
		}
		color := seriesColor(s.Color(), i, cfg.Theme.Palette)
		fillAlpha := alpha
		if hovSI >= 0 && i != hovSI {
			color = dimColor(color, HoverDimAlpha)
			fillAlpha = HoverDimAlpha / 4
		}

		needed := n * 2
		if cap(av.ptsBuf) < needed {
			av.ptsBuf = make([]float32, 0, needed)
		}
		pts := av.ptsBuf[:0]
		for _, p := range s.Points[:n] {
			if !finite(p.X) || !finite(p.Y) {
				continue
			}
			px := xAxis.Transform(p.X, left, right)
			py := yAxis.Transform(p.Y, bottom, top)
			// Entry animation: lerp Y from baseline toward
			// actual value for smooth grow-from-zero effect.
			if progress < 1 {
				py = bottom + (py-bottom)*progress
			}
			pts = append(pts, px, py)
		}
		av.ptsBuf = pts

		// Area fill: clip each quad (line segment + baseline) to
		// the plot rect using Sutherland-Hodgman so the fill stays
		// correct when the line extends outside the visible area.
		if len(pts) >= 4 {
			fill := gui.RGBA(color.R, color.G, color.B, fillAlpha)
			var quad [8]float32
			for k := 0; k < len(pts)-2; k += 2 {
				// Clamp line Y to baseline so the quad stays
				// convex. Points below bottom contribute no
				// visible area and would create a reflex vertex.
				qy0 := min(pts[k+1], bottom)
				qy1 := min(pts[k+3], bottom)
				if qy0 == bottom && qy1 == bottom {
					continue // degenerate zero-height quad
				}
				quad[0] = pts[k]
				quad[1] = qy0
				quad[2] = pts[k+2]
				quad[3] = qy1
				quad[4] = pts[k+2]
				quad[5] = bottom
				quad[6] = pts[k]
				quad[7] = bottom
				var clippedQ []float32
				clippedQ, av.clipA, av.clipB = clipConvexToRect(
					quad[:], left, right, top, bottom,
					av.clipA, av.clipB)
				if clippedQ != nil {
					ctx.FilledPolygon(clippedQ, fill)
				}
			}
		}

		// Line stroke: full clip to plot rect.
		clipped := clipPolylineToRect(pts, left, right, top, bottom)
		ctx.Polyline(clipped, color, cfg.LineWidth)
	}
}

func (av *areaView) drawStacked(
	ctx *render.Context, cfg *AreaCfg,
	xAxis, yAxis axis.Axis,
	left, right, top, bottom float32,
	alpha uint8, hovSI int, progress float32,
) {
	// Find reference point count from first non-empty series.
	refLen := 0
	for _, s := range cfg.Series {
		if s.Len() > 0 {
			refLen = s.Len()
			break
		}
	}
	if refLen == 0 {
		return
	}

	cumY := make([]float64, refLen)

	// prevPts holds the pixel coords of the previous series' top edge.
	// Initialize to baseline Y using the first non-empty series' X positions.
	// cfg.Series[0] may be empty; using it directly would leave prev
	// zero-initialized (y=0 instead of y=bottom), corrupting the first fill.
	needed := refLen * 2
	if cap(av.prevPtsBuf) < needed {
		av.prevPtsBuf = make([]float32, needed)
	}
	prev := av.prevPtsBuf[:needed]
	for _, s := range cfg.Series {
		if s.Len() == 0 {
			continue
		}
		for j, p := range s.Points {
			if j >= refLen {
				break
			}
			prev[j*2] = xAxis.Transform(p.X, left, right)
			prev[j*2+1] = bottom
		}
		break
	}

	if cap(av.curPtsBuf) < needed {
		av.curPtsBuf = make([]float32, needed)
	}

	for i, s := range cfg.Series {
		if s.Len() == 0 || av.hidden[i] {
			continue
		}
		color := seriesColor(s.Color(), i, cfg.Theme.Palette)
		fillAlpha := alpha
		if hovSI >= 0 && i != hovSI {
			color = dimColor(color, HoverDimAlpha)
			fillAlpha = HoverDimAlpha / 4
		}
		fill := gui.RGBA(color.R, color.G, color.B, fillAlpha)

		n := min(s.Len(), refLen)
		cur := av.curPtsBuf[:n*2]
		for j := range n {
			p := s.Points[j]
			if !finite(p.X) || !finite(p.Y) {
				continue
			}
			cumY[j] += p.Y
			cur[j*2] = xAxis.Transform(p.X, left, right)
			py := yAxis.Transform(cumY[j], bottom, top)
			// Entry animation: lerp Y from baseline toward
			// actual value for smooth grow-from-zero effect.
			if progress < 1 {
				py = bottom + (py-bottom)*progress
			}
			cur[j*2+1] = py
		}

		// Fill quad between cur top edge and prev top edge (or
		// baseline). Clip to plot rect using Sutherland-Hodgman
		// so fill stays correct when lines extend outside.
		if len(cur) >= 4 {
			var quad [8]float32
			for k := 0; k < len(cur)-2; k += 2 {
				// Clamp Y values to baseline so quads stay
				// convex when zoomed beyond data range.
				cy0 := min(cur[k+1], bottom)
				cy1 := min(cur[k+3], bottom)
				py0 := min(prev[k+1], bottom)
				py1 := min(prev[k+3], bottom)
				if cy0 == py0 && cy1 == py1 {
					continue // degenerate zero-height band
				}
				quad[0] = cur[k]
				quad[1] = cy0
				quad[2] = cur[k+2]
				quad[3] = cy1
				quad[4] = prev[k+2]
				quad[5] = py1
				quad[6] = prev[k]
				quad[7] = py0
				var clippedQ []float32
				clippedQ, av.clipA, av.clipB = clipConvexToRect(
					quad[:], left, right, top, bottom,
					av.clipA, av.clipB)
				if clippedQ != nil {
					ctx.FilledPolygon(clippedQ, fill)
				}
			}
		}

		clipped := clipPolylineToRect(cur, left, right, top, bottom)
		ctx.Polyline(clipped, color, cfg.LineWidth)
		copy(prev[:n*2], cur)
	}
}
