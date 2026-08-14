package chart

import (
	"fmt"
	"log/slog"
	"math"
	"sort"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// HistogramCfg configures a histogram chart.
type HistogramCfg struct {
	BaseCfg
	InteractionCfg

	// Data holds raw values to bin and count.
	Data []float64

	// Bins is the number of bins. 0 = auto (Sturges rule).
	Bins int

	// BinEdges provides explicit bin boundaries (len = Bins+1).
	// When set, Bins is ignored.
	BinEdges []float64

	// Normalized reports frequency density instead of count.
	Normalized bool

	// Color overrides the default palette color for bars.
	Color gui.Color

	// Radius is the corner radius for bars (0 = square).
	Radius float32

	// YAxis overrides the auto-computed Y axis.
	YAxis axis.Axis

	// TickFormat formats bin-edge values for axis labels and
	// tooltips. nil = default "%.1f".
	TickFormat func(float64) string
}

type histogramView struct {
	cfg HistogramCfg
	xyBase
	lastVersion uint64
	binEdges    []float64
	binValues   []float64 // counts or densities
	xAxis       axis.Axis
	yAxis       axis.Axis
}

// Histogram creates a histogram chart view.
func Histogram(cfg HistogramCfg) gui.View {
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableHistogram(&cfg.BaseCfg, cfg.Data, cfg.Bins, cfg.BinEdges, cfg.Normalized)
	}
	hv := &histogramView{cfg: cfg}
	hv.base = &hv.cfg.BaseCfg
	hv.interaction = &hv.cfg.InteractionCfg
	hv.zoomX = true
	hv.zoomY = true
	// No nearestFn: histogram has no per-point cursor upgrade.
	return hv
}

// Draw renders the histogram onto dc for headless export.
func (hv *histogramView) Draw(dc *gui.DrawContext) { hv.draw(dc) }

func (hv *histogramView) chartTheme() *theme.Theme { return hv.cfg.Theme }

func (hv *histogramView) GenerateLayout(w *gui.Window) gui.Layout {
	return hv.generateLayout(w, hv.draw)
}

func (hv *histogramView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &hv.cfg
	th := cfg.Theme

	if len(cfg.Data) == 0 {
		slog.Warn("no data", "chart", cfg.ID)
		return
	}

	left := th.PaddingLeft
	right := ctx.Width() - th.PaddingRight
	top := th.PaddingTop
	bottom := ctx.Height() - th.PaddingBottom

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	drawTitle(ctx, cfg.Title, th)

	// Recompute bins only when version changes.
	if hv.yAxis == nil || cfg.Version != hv.lastVersion {
		edges, counts := calcBins(cfg.Data, cfg.Bins, cfg.BinEdges)
		if len(edges) < 2 {
			slog.Warn("insufficient bins", "chart", cfg.ID)
			return
		}
		hv.binEdges = edges
		numBins := len(counts)

		hv.binValues = make([]float64, numBins)
		if cfg.Normalized {
			total := 0
			for _, c := range counts {
				total += c
			}
			if total > 0 {
				for i, c := range counts {
					w := edges[i+1] - edges[i]
					if w > 0 {
						hv.binValues[i] = float64(c) / (float64(total) * w)
					}
				}
			}
		} else {
			for i, c := range counts {
				hv.binValues[i] = float64(c)
			}
		}

		// X axis: linear range over bin edges, no auto-expansion.
		hv.xAxis = axis.NewLinear(axis.LinearCfg{
			Min: edges[0],
			Max: edges[len(edges)-1],
		})

		// Y axis: [0, maxVal] with 5% top padding.
		maxVal := 0.0
		for _, v := range hv.binValues {
			maxVal = max(maxVal, v)
		}
		if maxVal == 0 {
			maxVal = 1
		}
		pad := maxVal * 0.05

		if cfg.YAxis != nil {
			hv.yAxis = cfg.YAxis
		} else {
			yLabel := "Frequency"
			if cfg.Normalized {
				yLabel = "Density"
			}
			hv.yAxis = axis.NewLinear(axis.LinearCfg{
				Title:     yLabel,
				AutoRange: true,
			})
			hv.yAxis.SetRange(0, maxVal+pad)
		}
		hv.lastVersion = cfg.Version
	}

	if len(hv.binEdges) < 2 {
		return
	}

	zs := loadAndApplyZoom(hv.win, hv.cfg.ID, hv.xAxis, hv.yAxis, true, true)

	left = resolveLeft(ctx, th, left, bottom, top, hv.yAxis)

	bottom = ctx.Height() - resolveBottom(ctx, th,
		maxTickLabelWidth(ctx, hv.xAxis.Ticks(left, right), th.TickStyle),
		cfg.XTickRotation, hv.xAxis.Label())

	// Cache plot bounds for hover hit-testing.
	hv.lastPA = plotArea{plotRect{left, right, top, bottom}, hv.xAxis, hv.yAxis}

	progress := animProgress(hv.win, hv.cfg.ID)

	pr := plotRect{left, right, top, bottom}
	hv.drawBars(ctx, th, pr, progress)

	drawSelectionRectIf(ctx, zs, pr, th)

	if hv.hovering {
		hv.tooltipHistogram(ctx, th, pr)
	}
}

func (hv *histogramView) drawBars(
	ctx *render.Context, th *theme.Theme, pr plotRect,
	progress float32,
) {
	left, right, top, bottom := pr.Left, pr.Right, pr.Top, pr.Bottom
	cfg := &hv.cfg
	yAxis := hv.yAxis
	xAxis := hv.xAxis
	binEdges := hv.binEdges
	binValues := hv.binValues
	numBins := len(binValues)

	// Y grid lines.
	yTicks := yAxis.Ticks(bottom, top)
	for _, t := range yTicks {
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
	for _, t := range yTicks {
		ctx.Line(left-tickLen, t.Position, left, t.Position,
			tickColor, tickWidth)
		tw := ctx.TextWidth(t.Label, tickStyle)
		ctx.Text(left-tickLen-tw-2, t.Position-fh/2, t.Label, tickStyle)
	}
	drawYAxisLabel(ctx, yAxis.Label(), th, top, bottom)

	// Annotations.
	drawAnnotations(ctx, &hv.cfg.Annotations, th, pr, xAxis, yAxis)

	// Resolve bar color.
	color := cfg.Color
	if !color.IsSet() {
		color = seriesColor(gui.Color{}, 0, th.Palette)
	}

	// Determine hovered bin index once.
	hovI := -1
	if hv.hovering {
		hovI = hv.hoveredBin(hv.hoverPx, left, right)
	}

	// Draw bars.
	for i, v := range binValues {
		if !finite(v) {
			continue
		}
		bx := xAxis.Transform(binEdges[i], left, right)
		bx2 := xAxis.Transform(binEdges[i+1], left, right)
		barWidth := bx2 - bx
		by := yAxis.Transform(v*float64(progress), bottom, top)
		bh := bottom - by
		if bh <= 0 {
			continue
		}
		barColor := color
		if hovI >= 0 && hovI != i {
			barColor = dimColor(color, HoverDimAlpha)
		}
		rx, ry, rw, rh, vis := clampRectToPlot(
			bx, by, barWidth, bh, left, right, top, bottom)
		if !vis {
			continue
		}
		if cfg.Radius > 0 {
			ctx.FilledRoundedRect(rx, ry, rw, rh, cfg.Radius, barColor)
			ctx.RoundedRect(rx, ry, rw, rh, cfg.Radius, th.AxisColor, th.AxisWidth)
		} else {
			ctx.FilledRect(rx, ry, rw, rh, barColor)
			ctx.Rect(rx, ry, rw, rh, th.AxisColor, th.AxisWidth)
		}
	}

	// X-axis tick marks and labels at bin edges, thinned to ≤10.
	stride := 1
	if numBins > 10 {
		stride = (numBins + 9) / 10
	}
	for i, edge := range binEdges {
		if i%stride != 0 && i != numBins {
			continue
		}
		px := xAxis.Transform(edge, left, right)
		ctx.Line(px, bottom, px, bottom+tickLen, tickColor, tickWidth)
		label := hv.formatEdge(edge)
		lw := ctx.TextWidth(label, tickStyle)
		ctx.Text(px-lw/2, bottom+tickLen+2, label, tickStyle)
	}
}

// hoveredBin returns the index of the bin under pixel mx,
// or -1 when outside the plot area.
func (hv *histogramView) hoveredBin(mx, left, right float32) int {
	if mx < left || mx > right || hv.xAxis == nil {
		return -1
	}
	xVal := hv.xAxis.Invert(mx, left, right)
	edges := hv.binEdges
	numBins := len(edges) - 1
	// First i where edges[i+1] > xVal.
	idx := sort.Search(numBins, func(i int) bool { return edges[i+1] > xVal })
	if idx >= numBins {
		idx = numBins - 1 // clamp: last edge closes the last bin
	}
	return idx
}

func (hv *histogramView) tooltipHistogram(
	ctx *render.Context, th *theme.Theme, pr plotRect,
) {
	if hv.yAxis == nil {
		return
	}
	left, right, top, bottom := pr.Left, pr.Right, pr.Top, pr.Bottom
	mx := hv.hoverPx
	my := hv.hoverPy
	edges := hv.binEdges
	values := hv.binValues

	i := hv.hoveredBin(mx, left, right)
	if i < 0 || i >= len(values) {
		return
	}

	v := values[i]
	by := hv.yAxis.Transform(v, bottom, top)
	if my < by || my > bottom {
		return
	}

	var label string
	lo := hv.formatEdge(edges[i])
	hi := hv.formatEdge(edges[i+1])
	if hv.cfg.Normalized {
		label = fmt.Sprintf("[%s, %s): density %.1f", lo, hi, v)
	} else {
		label = fmt.Sprintf("[%s, %s): count %.0f", lo, hi, v)
	}
	drawTooltip(ctx, mx, my, label, th,
		plotRect{left, right, top, bottom})
}

// formatEdge formats a bin-edge value using cfg.TickFormat,
// falling back to "%.1f".
func (hv *histogramView) formatEdge(v float64) string {
	if hv.cfg.TickFormat != nil {
		return hv.cfg.TickFormat(v)
	}
	return fmt.Sprintf("%.1f", v)
}

// calcBins computes bin edges and counts for data.
// numBins ≤ 0 triggers auto bin count (Sturges rule).
// If edges has 2+ entries it overrides numBins.
// NaN and Inf values in data are silently ignored.
func calcBins(data []float64, numBins int, edges []float64) ([]float64, []int) {
	clean := make([]float64, 0, len(data))
	for _, v := range data {
		if finite(v) {
			clean = append(clean, v)
		}
	}
	if len(clean) == 0 {
		return nil, nil
	}

	var binEdges []float64
	if len(edges) > 1 {
		binEdges = edges
	} else {
		n := len(clean)
		if numBins <= 0 {
			numBins = int(math.Ceil(math.Log2(float64(n)) + 1))
		}
		numBins = max(numBins, 1)

		lo, hi := clean[0], clean[0]
		for _, v := range clean[1:] {
			lo = min(lo, v)
			hi = max(hi, v)
		}

		// All identical values: single bin around the value.
		if lo == hi {
			delta := 1.0
			if lo != 0 {
				delta = math.Abs(lo) * 0.01
			}
			lo -= delta
			hi += delta
			numBins = 1
		}

		binEdges = make([]float64, numBins+1)
		w := (hi - lo) / float64(numBins)
		for i := range numBins + 1 {
			binEdges[i] = lo + float64(i)*w
		}
		binEdges[numBins] = hi // exact upper bound
	}

	numBins = len(binEdges) - 1
	counts := make([]int, numBins)
	for _, v := range clean {
		if v < binEdges[0] || v > binEdges[numBins] {
			continue
		}
		// Find first bin i where binEdges[i+1] > v.
		idx := sort.Search(numBins, func(i int) bool { return binEdges[i+1] > v })
		if idx >= numBins {
			idx = numBins - 1 // last bin: [lo, hi]
		}
		counts[idx]++
	}
	return binEdges, counts
}
