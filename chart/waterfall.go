package chart

import (
	"cmp"
	"fmt"
	"log/slog"
	"math"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// WaterfallValue describes one bar in a waterfall chart.
type WaterfallValue struct {
	Label string
	Value float64
	// IsTotal marks a summary bar spanning from 0 to the running
	// total. If Value is nonzero, it overrides the running total
	// (e.g. opening balance).
	IsTotal bool
}

// WaterfallCfg configures a waterfall chart.
type WaterfallCfg struct {
	BaseCfg
	InteractionCfg

	// Values holds one entry per bar.
	Values []WaterfallValue

	// YAxis overrides the auto-computed Y axis.
	YAxis axis.Axis

	// BarWidth is the body width in pixels. 0 = auto
	// (slot width * DefaultWaterfallWidthRatio).
	BarWidth float32

	// Radius is the corner radius for bars. 0 = square.
	Radius float32

	// UpColor is the color for positive-delta bars.
	// Zero value uses default green.
	UpColor gui.Color

	// DownColor is the color for negative-delta bars.
	// Zero value uses default red.
	DownColor gui.Color

	// TotalColor is the color for total bars.
	// Zero value uses default blue.
	TotalColor gui.Color

	// ShowConnectors controls connector line rendering
	// between bars. nil or true = draw; false = hide.
	ShowConnectors *bool
}

// Bar kind constants for legend indexing.
const (
	waterfallUp    = 0
	waterfallDown  = 1
	waterfallTotal = 2
)

// Default waterfall colors (Tableau 10 palette selections).
var (
	defaultUpColor    = gui.Hex(0x59a14f) // green
	defaultDownColor  = gui.Hex(0xe15759) // red
	defaultTotalColor = gui.Hex(0x4e79a7) // blue

	waterfallLegendNames = []string{"Increase", "Decrease", "Total"}
)

// waterfallBar caches computed geometry for one bar.
type waterfallBar struct {
	Bottom       float64 // data-space bottom
	Top          float64 // data-space top
	RunningTotal float64
	Kind         int // waterfallUp, waterfallDown, waterfallTotal
}

type waterfallView struct {
	cfg WaterfallCfg
	xyBase
	lastVersion uint64
	bars        []waterfallBar
	yAxis       axis.Axis
	yTicks      []axis.Tick
	xAxis       *axis.Category
}

// Waterfall creates a waterfall chart view.
func Waterfall(cfg WaterfallCfg) gui.View {
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableWaterfall(&cfg.BaseCfg, cfg.Values)
	}
	wv := &waterfallView{cfg: cfg}
	wv.base = &wv.cfg.BaseCfg
	wv.interaction = &wv.cfg.InteractionCfg
	wv.zoomX = false
	wv.zoomY = true
	wv.nearestFn = func(px, py float32) bool {
		return px >= wv.lastPA.Left && px <= wv.lastPA.Right &&
			py >= wv.lastPA.Top && py <= wv.lastPA.Bottom
	}
	return wv
}

// Draw renders the chart onto dc for headless export.
func (wv *waterfallView) Draw(dc *gui.DrawContext) { wv.draw(dc) }

func (wv *waterfallView) chartTheme() *theme.Theme { return wv.cfg.Theme }

func (wv *waterfallView) GenerateLayout(w *gui.Window) gui.Layout {
	return wv.generateLayout(w, wv.draw)
}

func (wv *waterfallView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &wv.cfg
	th := cfg.Theme

	if len(cfg.Values) == 0 {
		slog.Warn("no data", "chart", cfg.ID)
		return
	}

	left := th.PaddingLeft
	right := ctx.Width() - th.PaddingRight
	top := th.PaddingTop
	bottom := ctx.Height() - th.PaddingBottom

	right -= legendRightReserve(ctx, th, cfg.LegendPosition, waterfallLegendNames)
	top += legendTopReserve(ctx, th, cfg.LegendPosition,
		waterfallLegendNames, left, right)

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	drawTitle(ctx, cfg.Title, th)

	// Rebuild bars and axes when version changes.
	if wv.yAxis == nil || cfg.Version != wv.lastVersion {
		wv.buildBars(cfg)
		wv.lastVersion = cfg.Version
	}

	zs := loadAndApplyZoom(wv.win, wv.cfg.ID, nil, wv.yAxis, false, true)

	left = resolveLeft(ctx, th, left, bottom, top, wv.yAxis)

	bottom = ctx.Height() - resolveBottom(ctx, th,
		maxTickLabelWidth(ctx, wv.xAxis.Ticks(left, right), th.TickStyle),
		cfg.XTickRotation, wv.xAxis.Label())
	bottom -= legendBottomReserve(ctx, th, cfg.LegendPosition,
		waterfallLegendNames, left, right)

	wv.lastPA = plotArea{plotRect{left, right, top, bottom}, nil, wv.yAxis}

	wv.yTicks = wv.yAxis.Ticks(bottom, top)

	// Grid lines.
	for _, t := range wv.yTicks {
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
	for _, t := range wv.yTicks {
		ctx.Line(left-tickLen, t.Position, left, t.Position,
			tickColor, tickWidth)
		tw := ctx.TextWidth(t.Label, tickStyle)
		ctx.Text(left-tickLen-tw-2, t.Position-fh/2, t.Label, tickStyle)
	}
	drawYAxisLabel(ctx, wv.yAxis.Label(), th, top, bottom)

	// Annotations.
	drawAnnotations(ctx, &cfg.Annotations, th,
		plotRect{left, right, top, bottom}, wv.xAxis, wv.yAxis)

	progress := animProgress(wv.win, wv.cfg.ID)

	// Determine hovered bar index.
	hovI := -1
	if wv.hovering {
		hovI = wv.hoveredBar(wv.hoverPx, left, right)
	}

	nBars := len(cfg.Values)
	slotW := (right - left) / float32(nBars)
	barW := cfg.BarWidth
	if barW <= 0 {
		barW = slotW * DefaultWaterfallWidthRatio
	}
	barW = max(barW, 2)

	upColor := wv.resolveColor(waterfallUp)
	downColor := wv.resolveColor(waterfallDown)
	totalColor := wv.resolveColor(waterfallTotal)

	showConn := cfg.ShowConnectors == nil || *cfg.ShowConnectors

	// Draw bars and connectors.
	for i, bb := range wv.bars {
		if wv.hidden[bb.Kind] {
			continue
		}

		cx := wv.xAxis.Transform(float64(i), left, right)
		p := float64(progress)
		mid := (bb.Top + bb.Bottom) / 2
		topPx := wv.yAxis.Transform(mid+(bb.Top-mid)*p, bottom, top)
		botPx := wv.yAxis.Transform(mid+(bb.Bottom-mid)*p, bottom, top)

		barTop := min(topPx, botPx)
		bh := float32(math.Abs(float64(topPx - botPx)))
		bh = max(bh, 1)

		color := upColor
		switch bb.Kind {
		case waterfallDown:
			color = downColor
		case waterfallTotal:
			color = totalColor
		}
		if hovI >= 0 && hovI != i {
			color = dimColor(color, HoverDimAlpha)
		}

		drawClampedBar(ctx, cx-barW/2, barTop, barW, bh,
			cfg.Radius, color, left, right, top, bottom)

		// Connector line from previous bar (skip if outside plot Y).
		if showConn && i > 0 {
			prevBB := wv.bars[i-1]
			if !wv.hidden[prevBB.Kind] {
				connY := wv.yAxis.Transform(
					prevBB.RunningTotal, bottom, top)
				if connY >= top && connY <= bottom {
					prevCx := wv.xAxis.Transform(
						float64(i-1), left, right)
					connColor := th.GridColor
					if hovI >= 0 {
						connColor = dimColor(connColor, HoverDimAlpha)
					}
					ctx.DashedLine(prevCx+barW/2, connY,
						cx-barW/2, connY,
						connColor, 1/ctx.Scale(), 4, 3)
				}
			}
		}
	}

	// X tick marks and labels.
	xts := tickStyle
	if cfg.XTickRotation != 0 {
		xts.RotationRadians = cfg.XTickRotation
	}
	xTicks := wv.xAxis.Ticks(left, right)
	for _, t := range xTicks {
		ctx.Line(t.Position, bottom, t.Position, bottom+tickLen,
			tickColor, tickWidth)
		if cfg.XTickRotation != 0 {
			ctx.Text(t.Position, bottom+tickLen+2, t.Label, xts)
		} else {
			lw := ctx.TextWidth(t.Label, xts)
			ctx.Text(t.Position-lw/2, bottom+tickLen+2, t.Label, xts)
		}
	}

	// Legend.
	entries := []legendEntry{
		{Name: "Increase", Color: upColor, Index: waterfallUp},
		{Name: "Decrease", Color: downColor, Index: waterfallDown},
		{Name: "Total", Color: totalColor, Index: waterfallTotal},
	}
	pr := plotRect{left, right, top, bottom}
	wv.lastLB = drawLegend(ctx, entries, th, pr,
		cfg.LegendPosition, wv.hidden)
	saveLegendBounds(wv.win, cfg.ID, wv.lastLB)

	drawSelectionRectIf(ctx, zs, pr, th)

	if wv.hovering {
		drawCrosshair(ctx, th, wv.hoverPx, wv.hoverPy, pr)
		wv.tooltipWaterfall(ctx, pr, th)
	}
}

// buildBars computes running totals and bar geometry.
func (wv *waterfallView) buildBars(cfg *WaterfallCfg) {
	n := len(cfg.Values)
	wv.bars = make([]waterfallBar, n)

	runningTotal := 0.0
	globalMin := 0.0
	globalMax := 0.0

	for i, v := range cfg.Values {
		var bb waterfallBar
		if v.IsTotal {
			// Nonzero Value on a total bar overrides the
			// running total (e.g. opening balance).
			if v.Value != 0 {
				runningTotal = v.Value
			}
			bb.Bottom = 0
			bb.Top = runningTotal
			bb.Kind = waterfallTotal
		} else if v.Value >= 0 {
			bb.Bottom = runningTotal
			runningTotal += v.Value
			bb.Top = runningTotal
			bb.Kind = waterfallUp
		} else {
			bb.Top = runningTotal
			runningTotal += v.Value
			bb.Bottom = runningTotal
			bb.Kind = waterfallDown
		}
		bb.RunningTotal = runningTotal
		wv.bars[i] = bb

		globalMin = min(globalMin, bb.Bottom, bb.Top)
		globalMax = max(globalMax, bb.Bottom, bb.Top)
	}

	// Y axis.
	if cfg.YAxis != nil {
		wv.yAxis = cfg.YAxis
	} else {
		span := globalMax - globalMin
		if span == 0 {
			span = 1
		}
		pad := span * 0.05
		wv.yAxis = axis.NewLinear(axis.LinearCfg{AutoRange: true})
		wv.yAxis.SetRange(min(0, globalMin-pad), max(0, globalMax+pad))
	}

	// X axis: category labels.
	labels := make([]string, n)
	for i, v := range cfg.Values {
		labels[i] = v.Label
	}
	wv.xAxis = axis.NewCategory(axis.CategoryCfg{Categories: labels})
}

// resolveColor returns the color for a bar kind, using config
// overrides or defaults.
func (wv *waterfallView) resolveColor(kind int) gui.Color {
	cfg := &wv.cfg
	switch kind {
	case waterfallUp:
		return cmp.Or(cfg.UpColor, defaultUpColor)
	case waterfallDown:
		return cmp.Or(cfg.DownColor, defaultDownColor)
	default:
		return cmp.Or(cfg.TotalColor, defaultTotalColor)
	}
}

// hoveredBar returns the index of the bar under pixel mx,
// or -1 when outside the plot area.
func (wv *waterfallView) hoveredBar(mx, left, right float32) int {
	if mx < left || mx > right {
		return -1
	}
	nBars := len(wv.cfg.Values)
	if nBars == 0 {
		return -1
	}
	slotW := (right - left) / float32(nBars)
	slot := int((mx - left) / slotW)
	return max(0, min(slot, nBars-1))
}

// tooltipWaterfall draws a tooltip when the cursor is over a bar.
func (wv *waterfallView) tooltipWaterfall(
	ctx *render.Context, pr plotRect, th *theme.Theme,
) {
	if wv.yAxis == nil {
		return
	}
	left, right, top, bottom := pr.Left, pr.Right, pr.Top, pr.Bottom
	cfg := &wv.cfg
	mx := wv.hoverPx
	my := wv.hoverPy
	if mx < left || mx > right || my < top || my > bottom {
		return
	}

	i := wv.hoveredBar(mx, left, right)
	if i < 0 || i >= len(wv.bars) {
		return
	}
	bb := wv.bars[i]
	if wv.hidden[bb.Kind] {
		return
	}

	cx := wv.xAxis.Transform(float64(i), left, right)

	// Resolve bar width (must match draw loop).
	nBars := len(cfg.Values)
	slotW := (right - left) / float32(nBars)
	barW := cfg.BarWidth
	if barW <= 0 {
		barW = slotW * DefaultWaterfallWidthRatio
	}
	barW = max(barW, 2)

	// Horizontal: cursor must be within the bar body.
	if mx < cx-barW/2 || mx > cx+barW/2 {
		return
	}

	// Vertical: cursor must be within the bar bounds.
	topPx := wv.yAxis.Transform(bb.Top, bottom, top)
	botPx := wv.yAxis.Transform(bb.Bottom, bottom, top)
	barTopPx := min(topPx, botPx)
	barBotPx := max(topPx, botPx)
	if my < barTopPx || my > barBotPx {
		return
	}

	v := cfg.Values[i]
	var text string
	if v.IsTotal {
		text = fmt.Sprintf("%s\nTotal: %.4g", v.Label, bb.RunningTotal)
	} else {
		sign := "+"
		if v.Value < 0 {
			sign = ""
		}
		text = fmt.Sprintf("%s\nDelta: %s%.4g\nTotal: %.4g",
			v.Label, sign, v.Value, bb.RunningTotal)
	}

	drawTooltip(ctx, cx, barTopPx, text, th,
		plotRect{left, right, top, bottom})
}
