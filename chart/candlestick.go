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

// defaultCandleUp is the fallback color for up (bullish) candles.
var defaultCandleUp = gui.Hex(0x26a69a)

// defaultCandleDown is the fallback color for down (bearish) candles.
var defaultCandleDown = gui.Hex(0xef5350)

// CandlestickCfg configures a candlestick chart.
type CandlestickCfg struct {
	BaseCfg
	InteractionCfg

	// Data
	Series []series.OHLCSeries

	// Axes (optional; Y auto-created from High/Low bounds when nil)
	YAxis axis.Axis

	// CandleWidth is the body width in pixels. 0 = auto (slot width
	// multiplied by DefaultCandleWidthRatio).
	CandleWidth float32

	// XTimeFormat is the Go time layout used for X-axis tick labels.
	// Defaults to "01/02".
	XTimeFormat string
}

// Validate checks CandlestickCfg for invalid settings.
func (c *CandlestickCfg) Validate() error {
	if err := c.BaseCfg.Validate(); err != nil {
		return err
	}
	var errs []string
	if len(c.Series) == 0 {
		errs = append(errs, "no series data")
	}
	if c.CandleWidth < 0 {
		errs = append(errs, "negative CandleWidth")
	}
	return buildError("chart.Candlestick", errs)
}

type candlestickView struct {
	cfg CandlestickCfg
	xyBase
	lastVersion uint64
	yAxis       axis.Axis
	yTicks      []axis.Tick
	xAxis       *axis.Category
}

// Candlestick creates a candlestick chart view.
func Candlestick(cfg CandlestickCfg) gui.View {
	cfg.applyDefaults()
	cfg.XTimeFormat = cmp.Or(cfg.XTimeFormat, "01/02")
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableOHLC(&cfg.BaseCfg, cfg.Series, cfg.XTimeFormat)
	}
	cv := &candlestickView{cfg: cfg}
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
func (cv *candlestickView) Draw(dc *gui.DrawContext) { cv.draw(dc) }

func (cv *candlestickView) chartTheme() *theme.Theme { return cv.cfg.Theme }

func (cv *candlestickView) GenerateLayout(w *gui.Window) gui.Layout {
	return cv.generateLayout(w, cv.draw)
}

func (cv *candlestickView) draw(dc *gui.DrawContext) {
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
	top += legendTopReserve(ctx, th, cfg.LegendPosition, names, left, right)

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	drawTitle(ctx, cfg.Title, th)

	// Rebuild axes when data version changes.
	if cv.yAxis == nil || cfg.Version != cv.lastVersion {
		cv.buildAxes(cfg)
		cv.lastVersion = cfg.Version
	}

	zs := loadAndApplyZoom(cv.win, cv.cfg.ID, nil, cv.yAxis, false, true)

	left = resolveLeft(ctx, th, left, bottom, top, cv.yAxis)

	bottom = ctx.Height() - resolveBottom(ctx, th,
		maxTickLabelWidth(ctx, cv.xAxis.Ticks(left, right), th.TickStyle),
		cfg.XTickRotation, cv.xAxis.Label())
	bottom -= legendBottomReserve(ctx, th, cfg.LegendPosition, names, left, right)

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
		ctx.Line(left-tickLen, t.Position, left, t.Position, tickColor, tickWidth)
		tw := ctx.TextWidth(t.Label, tickStyle)
		ctx.Text(left-tickLen-tw-2, t.Position-fh/2, t.Label, tickStyle)
	}

	drawYAxisLabel(ctx, cv.yAxis.Label(), th, top, bottom)

	// Annotations.
	drawAnnotations(ctx, &cfg.Annotations, th,
		plotRect{left, right, top, bottom}, cv.xAxis, cv.yAxis)

	// X ticks and candles — only when at least one series is visible.
	firstSI := cv.firstVisibleSeries()
	if firstSI >= 0 {
		nPoints := len(cfg.Series[firstSI].Points)
		if nPoints > 0 {
			// X tick marks and labels.
			xts := tickStyle
			if cfg.XTickRotation != 0 {
				xts.RotationRadians = cfg.XTickRotation
			}
			xTicks := cv.xAxis.Ticks(left, right)
			for _, t := range xTicks {
				ctx.Line(t.Position, bottom, t.Position, bottom+tickLen, tickColor, tickWidth)
				if cfg.XTickRotation != 0 {
					ctx.Text(t.Position, bottom+tickLen+2, t.Label, xts)
				} else {
					lw := ctx.TextWidth(t.Label, xts)
					ctx.Text(t.Position-lw/2, bottom+tickLen+2, t.Label, xts)
				}
			}

			slotW := (right - left) / float32(nPoints)
			candleW := cfg.CandleWidth
			if candleW <= 0 {
				candleW = slotW * DefaultCandleWidthRatio
			}
			candleW = max(candleW, 1)

			cv.drawCandles(ctx, cfg, candleW,
				left, right, top, bottom)
		}
	}

	// Legend: two entries per series — up (index 2*i) and down (index 2*i+1).
	// Independent indices allow toggling up/down candles separately.
	entries := make([]legendEntry, 0, len(cfg.Series)*2)
	for i, s := range cfg.Series {
		name := s.Name()
		upLabel, downLabel := name+" ↑", name+" ↓"
		if name == "" {
			upLabel, downLabel = "↑", "↓"
		}
		entries = append(entries,
			legendEntry{Name: upLabel, Color: candleColor(s, true), Index: 2 * i},
			legendEntry{Name: downLabel, Color: candleColor(s, false), Index: 2*i + 1},
		)
	}
	pr := plotRect{left, right, top, bottom}
	cv.lastLB = drawLegend(ctx, entries, th, pr,
		cfg.LegendPosition, cv.hidden)
	saveLegendBounds(cv.win, cfg.ID, cv.lastLB)

	drawSelectionRectIf(ctx, zs, pr, th)

	if cv.hovering {
		drawCrosshair(ctx, th, cv.hoverPx, cv.hoverPy, pr)
		cv.tooltipCandlestick(ctx, pr, th)
	}
}

// buildAxes rebuilds yAxis and xAxis from current series data.
func (cv *candlestickView) buildAxes(cfg *CandlestickCfg) {
	// Y axis.
	if cfg.YAxis != nil {
		cv.yAxis = cfg.YAxis
	} else {
		minLow := math.MaxFloat64
		maxHigh := -math.MaxFloat64
		for si, s := range cfg.Series {
			if cv.hidden[2*si] && cv.hidden[2*si+1] {
				continue
			}
			for _, p := range s.Points {
				if !finite(p.Low) || !finite(p.High) {
					continue
				}
				minLow = min(minLow, p.Low)
				maxHigh = max(maxHigh, p.High)
			}
		}
		if minLow == math.MaxFloat64 {
			minLow = 0
			maxHigh = 1
		}
		if minLow == maxHigh {
			maxHigh = minLow + 1
		}
		span := maxHigh - minLow
		pad := span * 0.05
		cv.yAxis = axis.NewLinear(axis.LinearCfg{AutoRange: true})
		cv.yAxis.SetRange(minLow-pad, maxHigh+pad)
	}

	// X axis: labels from first visible series times.
	firstSI := cv.firstVisibleSeries()
	if firstSI < 0 {
		cv.xAxis = axis.NewCategory(axis.CategoryCfg{})
		return
	}
	pts := cfg.Series[firstSI].Points
	labels := make([]string, len(pts))
	for i, p := range pts {
		labels[i] = p.Time.Format(cfg.XTimeFormat)
	}
	cv.xAxis = axis.NewCategory(axis.CategoryCfg{Categories: labels})
}

// firstVisibleSeries returns the index of the first series that is not
// fully hidden (i.e. at least one of its up/down entries is visible),
// or -1 if all series are fully hidden.
func (cv *candlestickView) firstVisibleSeries() int {
	for i := range cv.cfg.Series {
		if !cv.hidden[2*i] || !cv.hidden[2*i+1] {
			return i
		}
	}
	return -1
}

// tooltipCandlestick draws an OHLC tooltip for the candle nearest the cursor.
func (cv *candlestickView) tooltipCandlestick(
	ctx *render.Context, pr plotRect, th *theme.Theme,
) {
	if cv.yAxis == nil {
		return
	}
	left, right, top, bottom := pr.Left, pr.Right, pr.Top, pr.Bottom
	cfg := &cv.cfg
	mx := cv.hoverPx
	my := cv.hoverPy
	if mx < left || mx > right || my < top || my > bottom {
		return
	}

	firstSI := cv.firstVisibleSeries()
	if firstSI < 0 {
		return
	}
	nCat := len(cfg.Series[firstSI].Points)
	if nCat == 0 {
		return
	}

	slotW := (right - left) / float32(nCat)
	slot := max(0, min(int((mx-left)/slotW), nCat-1))

	candleW := cfg.CandleWidth
	if candleW <= 0 {
		candleW = slotW * DefaultCandleWidthRatio
	}
	candleW = max(candleW, 1)

	// Find first visible series with this slot and hit-test against the candle.
	for si, s := range cfg.Series {
		if (cv.hidden[2*si] && cv.hidden[2*si+1]) || slot >= len(s.Points) {
			continue
		}
		p := s.Points[slot]
		if !finite(p.High) || !finite(p.Low) || !finite(p.Open) || !finite(p.Close) {
			continue
		}
		cx := cv.xAxis.Transform(float64(slot), left, right)

		// Horizontal hit: within the candle body width.
		if mx < cx-candleW/2 || mx > cx+candleW/2 {
			continue
		}
		// Vertical hit: within the full wick range (high to low).
		highPx := cv.yAxis.Transform(p.High, bottom, top)
		lowPx := cv.yAxis.Transform(p.Low, bottom, top)
		if my < highPx || my > lowPx {
			continue
		}
		label := fmt.Sprintf("%s\nO: %s\nH: %s\nL: %s\nC: %s",
			p.Time.Format(cfg.XTimeFormat),
			tooltipNum(p.Open), tooltipNum(p.High),
			tooltipNum(p.Low), tooltipNum(p.Close))
		drawTooltip(ctx, cx, highPx, label, th,
			plotRect{left, right, top, bottom})
		return
	}
}

// candleColor returns the up or down color for a candle, falling back to
// package defaults when the series color is unset.
func candleColor(s series.OHLCSeries, isUp bool) gui.Color {
	if isUp {
		if c := s.ColorUp(); c != (gui.Color{}) {
			return c
		}
		return defaultCandleUp
	}
	if c := s.ColorDown(); c != (gui.Color{}) {
		return c
	}
	return defaultCandleDown
}

// drawCandles renders all candle bodies and wicks, clamped to
// the plot area.
func (cv *candlestickView) drawCandles(
	ctx *render.Context, cfg *CandlestickCfg,
	candleW, left, right, top, bottom float32,
) {
	progress := animProgress(cv.win, cv.cfg.ID)
	for si, s := range cfg.Series {
		upHidden := cv.hidden[2*si]
		downHidden := cv.hidden[2*si+1]
		if upHidden && downHidden {
			continue
		}
		for i, p := range s.Points {
			if !finite(p.High) || !finite(p.Low) ||
				!finite(p.Open) || !finite(p.Close) {
				continue
			}
			isUp := p.Close >= p.Open
			if isUp && upHidden {
				continue
			}
			if !isUp && downHidden {
				continue
			}

			// Interpolate OHLC toward midpoint by (1-progress).
			mid := (p.Open + p.Close) / 2
			aOpen := mid + (p.Open-mid)*float64(progress)
			aClose := mid + (p.Close-mid)*float64(progress)
			aHigh := mid + (p.High-mid)*float64(progress)
			aLow := mid + (p.Low-mid)*float64(progress)

			cx := cv.xAxis.Transform(float64(i), left, right)
			highPx := cv.yAxis.Transform(aHigh, bottom, top)
			lowPx := cv.yAxis.Transform(aLow, bottom, top)
			openPx := cv.yAxis.Transform(aOpen, bottom, top)
			closePx := cv.yAxis.Transform(aClose, bottom, top)

			color := candleColor(s, isUp)

			// Wick: high to low (clamped to plot Y).
			if wy0, wy1, vis := clampVerticalLine(
				highPx, lowPx, top, bottom); vis {
				ctx.Line(cx, wy0, cx, wy1, color, 1.5)
			}

			// Body.
			bodyTop := min(openPx, closePx)
			bodyH := float32(math.Abs(float64(closePx - openPx)))
			bodyH = max(bodyH, 1)
			drawClampedBar(ctx, cx-candleW/2, bodyTop,
				candleW, bodyH, 0, color,
				left, right, top, bottom)
		}
	}
}
