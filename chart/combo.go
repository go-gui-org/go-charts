package chart

import (
	"cmp"
	"fmt"
	"log/slog"
	"math"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// ComboSeriesType distinguishes bar and line renderers within a
// combo chart.
type ComboSeriesType int

const (
	// ComboBar renders the series as vertical bars.
	ComboBar ComboSeriesType = iota
	// ComboLine renders the series as a polyline.
	ComboLine
)

// ComboSeries pairs a category data series with its render type.
type ComboSeries struct {
	series.Category
	Type ComboSeriesType
}

// ComboCfg configures a combo chart that overlays bar and line
// series on shared category axes.
type ComboCfg struct {
	BaseCfg
	InteractionCfg

	// Series holds one or more bar/line series. All series must
	// have the same number of category values.
	Series []ComboSeries

	// YAxis overrides the auto-computed Y axis.
	YAxis axis.Axis

	// BarWidth is the body width in pixels. 0 = auto
	// (evenly divided across bar series within each group).
	BarWidth float32

	// BarGap is the gap between adjacent bars in pixels.
	BarGap float32

	// Radius is the corner radius for bars. 0 = square.
	Radius float32

	// LineWidth is the polyline stroke width. 0 = DefaultLineWidth.
	LineWidth float32

	// ShowMarkers draws filled circles at each line data point.
	ShowMarkers bool
}

type comboView struct {
	cfg ComboCfg
	xyBase
	lastVersion uint64
	xAxis       *axis.Category
	yAxis       axis.Axis
	yTicks      []axis.Tick
	ptsBuf      []float32
}

// Combo creates a combo chart view.
func Combo(cfg ComboCfg) gui.View {
	cfg.applyDefaults()
	cfg.BarGap = cmp.Or(cfg.BarGap, DefaultBarGap)
	cfg.LineWidth = cmp.Or(cfg.LineWidth, DefaultLineWidth)
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableCombo(&cfg.BaseCfg, cfg.Series)
	}
	cv := &comboView{cfg: cfg}
	cv.base = &cv.cfg.BaseCfg
	cv.interaction = &cv.cfg.InteractionCfg
	cv.zoomX = false
	cv.zoomY = true
	cv.nearestFn = func(px, py float32) bool {
		return px >= cv.lastPA.Left && px <= cv.lastPA.Right &&
			py >= cv.lastPA.Top && py <= cv.lastPA.Bottom
	}
	return cv
}

// Draw renders the chart onto dc for headless export.
func (cv *comboView) Draw(dc *gui.DrawContext) { cv.draw(dc) }

func (cv *comboView) chartTheme() *theme.Theme { return cv.cfg.Theme }

func (cv *comboView) GenerateLayout(w *gui.Window) gui.Layout {
	return cv.generateLayout(w, cv.draw)
}

// draw is the main rendering callback.
func (cv *comboView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &cv.cfg
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
	top += legendTopReserve(ctx, th, cfg.LegendPosition,
		names, left, right)

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	drawTitle(ctx, cfg.Title, th)

	// Category labels from first series.
	labels := cfg.Series[0].Values
	nCategories := len(labels)
	if nCategories == 0 {
		slog.Warn("no category data", "chart", cfg.ID)
		return
	}

	// Resolve bottom from category label widths.
	maxTickW := float32(0)
	for _, v := range labels {
		maxTickW = max(maxTickW, ctx.TextWidth(v.Label, th.TickStyle))
	}
	bottom = ctx.Height() - resolveBottom(
		ctx, th, maxTickW, cfg.XTickRotation, "")
	bottom -= legendBottomReserve(ctx, th, cfg.LegendPosition,
		names, left, right)

	// Recompute Y axis when version changes.
	if cv.yAxis == nil || cfg.Version != cv.lastVersion {
		cv.updateYAxis(cfg, nCategories)
		cv.lastVersion = cfg.Version
	}

	zs := loadAndApplyZoom(cv.win, cv.cfg.ID, nil, cv.yAxis, false, true)

	left = resolveLeft(ctx, th, left, bottom, top, cv.yAxis)

	cv.lastPA = plotArea{plotRect{left, right, top, bottom}, nil, cv.yAxis}

	cv.yTicks = cv.yAxis.Ticks(bottom, top)

	// Grid lines.
	for _, t := range cv.yTicks {
		ctx.Line(left, t.Position, right, t.Position,
			th.GridColor, th.GridWidth)
	}

	// Axis lines.
	ctx.Line(left, bottom, right, bottom, th.AxisColor, th.AxisWidth)
	ctx.Line(left, top, left, bottom, th.AxisColor, th.AxisWidth)

	// Y tick marks and labels.
	tickLen, tickWidth, tickColor := resolvedTickMark(th)
	tickStyle := th.TickStyle
	fh := ctx.FontHeight(tickStyle)
	for _, t := range cv.yTicks {
		ctx.Line(left-tickLen, t.Position, left, t.Position,
			tickColor, tickWidth)
		tw := ctx.TextWidth(t.Label, tickStyle)
		ctx.Text(left-tickLen-tw-2, t.Position-fh/2, t.Label, tickStyle)
	}
	drawYAxisLabel(ctx, cv.yAxis.Label(), th, top, bottom)

	// Annotations.
	drawAnnotations(ctx, &cfg.Annotations, th,
		plotRect{left, right, top, bottom}, cv.xAxis, cv.yAxis)

	// Determine hovered element.
	hovCI, hovSI, hovOK := -1, -1, false
	if cv.hovering {
		hovCI, hovSI, hovOK = cv.hoveredElement(
			cv.hoverPx, cv.hoverPy, left, right, top, bottom)
	}

	chartW := right - left
	groupWidth := chartW / float32(nCategories)

	// Count visible bar series for width calculation.
	nBarSeries := 0
	for si, s := range cfg.Series {
		if s.Type == ComboBar && !cv.hidden[si] {
			nBarSeries++
		}
	}

	progress := animProgress(cv.win, cv.cfg.ID)

	// Draw bars underneath, then lines on top.
	cv.drawBars(ctx, cfg, th, nCategories, nBarSeries, groupWidth,
		hovCI, hovSI, hovOK, left, top, bottom, progress)
	cv.drawLines(ctx, cfg, th, nCategories, groupWidth,
		hovSI, hovOK, left, top, bottom, progress)

	// --- X tick marks and labels ---
	xts := tickStyle
	if cfg.XTickRotation != 0 {
		xts.RotationRadians = cfg.XTickRotation
	}
	for ci := range nCategories {
		cx := left + float32(ci)*groupWidth + groupWidth/2
		ctx.Line(cx, bottom, cx, bottom+tickLen, tickColor, tickWidth)
		if cfg.XTickRotation != 0 {
			ctx.Text(cx, bottom+tickLen+2, labels[ci].Label, xts)
		} else {
			lw := ctx.TextWidth(labels[ci].Label, xts)
			ctx.Text(cx-lw/2, bottom+tickLen+2, labels[ci].Label, xts)
		}
	}

	// --- Legend ---
	entries := make([]legendEntry, len(cfg.Series))
	for i, s := range cfg.Series {
		entries[i] = legendEntry{
			Name:  s.Name(),
			Color: seriesColor(s.Color(), i, th.Palette),
			Index: i,
		}
	}
	pr := plotRect{left, right, top, bottom}
	cv.lastLB = drawLegend(ctx, entries, th, pr,
		cfg.LegendPosition, cv.hidden)
	saveLegendBounds(cv.win, cfg.ID, cv.lastLB)

	drawSelectionRectIf(ctx, zs, pr, th)

	// --- Crosshair and tooltip ---
	if cv.hovering {
		drawCrosshair(ctx, th, cv.hoverPx, cv.hoverPy, pr)
		cv.tooltipCombo(ctx, pr, th)
	}
}

// drawBars renders all ComboBar series as grouped vertical bars.
func (cv *comboView) drawBars(
	ctx *render.Context, cfg *ComboCfg, th *theme.Theme,
	nCategories, nBarSeries int, groupWidth float32,
	hovCI, hovSI int, hovOK bool,
	left, top, bottom float32, progress float32,
) {
	if nBarSeries == 0 {
		return
	}

	barGap := cfg.BarGap
	barWidth := cfg.BarWidth
	if barWidth == 0 {
		usable := groupWidth - barGap*2
		barWidth = (usable - barGap*float32(nBarSeries-1)) /
			float32(nBarSeries)
		barWidth = max(barWidth, 2)
	}

	baseline := cv.yAxis.Transform(0, bottom, top)

	for ci := range nCategories {
		groupX := left + float32(ci)*groupWidth
		barStart := groupX + (groupWidth-
			float32(nBarSeries)*barWidth-
			float32(nBarSeries-1)*barGap)/2

		barIdx := 0
		for si, s := range cfg.Series {
			if s.Type != ComboBar || cv.hidden[si] {
				continue
			}
			if ci >= len(s.Values) {
				continue
			}
			v := s.Values[ci].Value
			if !finite(v) {
				barIdx++
				continue
			}
			av := v * float64(progress)
			color := seriesColor(s.Color(), si, th.Palette)
			if hovOK && (ci != hovCI || si != hovSI) {
				color = dimColor(color, HoverDimAlpha)
			}

			bx := barStart + float32(barIdx)*(barWidth+barGap)
			by := cv.yAxis.Transform(av, bottom, top)
			barTop := min(by, baseline)
			bh := float32(math.Abs(float64(by - baseline)))

			drawClampedBar(ctx, bx, barTop, barWidth, bh,
				cfg.Radius, color,
				left, cv.lastPA.Right, top, bottom)
			barIdx++
		}
	}
}

// drawLines renders all ComboLine series as polylines.
func (cv *comboView) drawLines(
	ctx *render.Context, cfg *ComboCfg, th *theme.Theme,
	nCategories int, groupWidth float32,
	hovSI int, hovOK bool,
	left, top, bottom float32, progress float32,
) {
	for si, s := range cfg.Series {
		if s.Type != ComboLine || cv.hidden[si] {
			continue
		}
		color := seriesColor(s.Color(), si, th.Palette)
		if hovOK && si != hovSI {
			color = dimColor(color, HoverDimAlpha)
		}

		n := min(len(s.Values), nCategories)
		if progress < 1 {
			n = max(1, int(float32(n)*progress))
		}
		needed := n * 2
		if cap(cv.ptsBuf) < needed {
			cv.ptsBuf = make([]float32, 0, needed)
		}
		pts := cv.ptsBuf[:0]

		for ci := range n {
			v := s.Values[ci].Value
			if !finite(v) {
				continue
			}
			cx := left + float32(ci)*groupWidth + groupWidth/2
			cy := cv.yAxis.Transform(v, bottom, top)
			pts = append(pts, cx, cy)
		}

		if len(pts) >= 4 {
			ctx.Polyline(pts, color, cfg.LineWidth)
		}

		if cfg.ShowMarkers {
			for j := 0; j+1 < len(pts); j += 2 {
				ctx.FilledCircle(pts[j], pts[j+1],
					cfg.LineWidth*2, color)
			}
		}
		cv.ptsBuf = pts
	}
}

// updateYAxis computes the Y axis from all series data.
func (cv *comboView) updateYAxis(cfg *ComboCfg, nCategories int) {
	if cfg.YAxis != nil {
		cv.yAxis = cfg.YAxis
		return
	}
	minVal := 0.0
	maxVal := 0.0
	for _, s := range cfg.Series {
		for _, v := range s.Values {
			if !finite(v.Value) {
				continue
			}
			minVal = min(minVal, v.Value)
			maxVal = max(maxVal, v.Value)
		}
	}
	if minVal == 0 && maxVal == 0 {
		maxVal = 1
	}
	rangeVal := maxVal - minVal
	if rangeVal == 0 {
		rangeVal = 1
	}
	pad := rangeVal * 0.05
	lo := minVal
	if lo < 0 {
		lo -= pad
	}
	hi := maxVal
	if hi > 0 {
		hi += pad
	}
	cv.yAxis = axis.NewLinear(axis.LinearCfg{AutoRange: true})
	cv.yAxis.SetRange(min(0, lo), max(0, hi))

	// X axis: category labels.
	catLabels := make([]string, nCategories)
	for i, v := range cfg.Series[0].Values {
		catLabels[i] = v.Label
	}
	cv.xAxis = axis.NewCategory(axis.CategoryCfg{
		Categories: catLabels,
	})
}

// hoveredElement returns the (categoryIdx, seriesIdx) of the
// element under (mx, my). Bars are checked first, then line
// points. Returns ok=false when nothing is within range.
func (cv *comboView) hoveredElement(
	mx, my, left, right, top, bottom float32,
) (ci, si int, ok bool) {
	if mx < left || mx > right || my < top || my > bottom {
		return -1, -1, false
	}
	cfg := &cv.cfg
	if len(cfg.Series) == 0 || cv.yAxis == nil {
		return -1, -1, false
	}
	nCategories := len(cfg.Series[0].Values)
	chartW := right - left
	groupWidth := chartW / float32(nCategories)

	// Count visible bar series.
	nBarSeries := 0
	for ssi, s := range cfg.Series {
		if s.Type == ComboBar && !cv.hidden[ssi] {
			nBarSeries++
		}
	}

	barGap := cfg.BarGap

	// Check bars.
	if nBarSeries > 0 {
		barWidth := cfg.BarWidth
		if barWidth == 0 {
			usable := groupWidth - barGap*2
			barWidth = (usable - barGap*float32(nBarSeries-1)) /
				float32(nBarSeries)
			barWidth = max(barWidth, 2)
		}

		baseline := cv.yAxis.Transform(0, bottom, top)

		for cci := range nCategories {
			groupX := left + float32(cci)*groupWidth
			barStart := groupX + (groupWidth-
				float32(nBarSeries)*barWidth-
				float32(nBarSeries-1)*barGap)/2

			barIdx := 0
			for ssi, s := range cfg.Series {
				if s.Type != ComboBar || cv.hidden[ssi] {
					continue
				}
				if cci >= len(s.Values) {
					barIdx++
					continue
				}
				v := s.Values[cci].Value
				if !finite(v) {
					barIdx++
					continue
				}

				bx := barStart + float32(barIdx)*(barWidth+barGap)
				by := cv.yAxis.Transform(v, bottom, top)
				barTop := min(by, baseline)
				bh := float32(math.Abs(float64(by - baseline)))

				if mx >= bx && mx <= bx+barWidth &&
					my >= barTop && my <= barTop+bh {
					return cci, ssi, true
				}
				barIdx++
			}
		}
	}

	// Check line points (nearest within snap distance).
	const snapPx float32 = 20
	best := snapPx * snapPx
	foundCI, foundSI := -1, -1
	for ssi, s := range cfg.Series {
		if s.Type != ComboLine || cv.hidden[ssi] {
			continue
		}
		n := min(len(s.Values), nCategories)
		for cci := range n {
			v := s.Values[cci].Value
			if !finite(v) {
				continue
			}
			px := left + float32(cci)*groupWidth + groupWidth/2
			py := cv.yAxis.Transform(v, bottom, top)
			dx := px - mx
			dy := py - my
			d2 := dx*dx + dy*dy
			if d2 < best {
				best = d2
				foundCI = cci
				foundSI = ssi
			}
		}
	}
	if foundCI >= 0 {
		return foundCI, foundSI, true
	}
	return -1, -1, false
}

// tooltipCombo draws a tooltip for the hovered element.
func (cv *comboView) tooltipCombo(
	ctx *render.Context, pr plotRect, th *theme.Theme,
) {
	if cv.yAxis == nil {
		return
	}
	mx := cv.hoverPx
	my := cv.hoverPy
	if mx < pr.Left || mx > pr.Right ||
		my < pr.Top || my > pr.Bottom {
		return
	}

	ci, si, ok := cv.hoveredElement(
		mx, my, pr.Left, pr.Right, pr.Top, pr.Bottom)
	if !ok {
		return
	}

	cfg := &cv.cfg
	s := cfg.Series[si]
	if ci >= len(s.Values) {
		return
	}
	v := s.Values[ci]

	nCategories := len(cfg.Series[0].Values)
	groupWidth := (pr.Right - pr.Left) / float32(nCategories)
	cx := pr.Left + float32(ci)*groupWidth + groupWidth/2
	py := cv.yAxis.Transform(v.Value, pr.Bottom, pr.Top)

	var label string
	if s.Name() != "" {
		label = fmt.Sprintf("%s / %s: %s",
			s.Name(), v.Label, tooltipNum(v.Value))
	} else {
		label = fmt.Sprintf("%s: %s", v.Label, tooltipNum(v.Value))
	}
	drawTooltip(ctx, cx, py, label, th, pr)
}
