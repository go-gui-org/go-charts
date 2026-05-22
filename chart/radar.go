package chart

import (
	"fmt"
	"log/slog"
	"math"
	"slices"

	"github.com/mike-ward/go-charts/render"
	"github.com/mike-ward/go-charts/theme"
	"github.com/mike-ward/go-gui/gui"
)

// RadarAxis defines one axis of the radar chart.
type RadarAxis struct {
	Label string
	Min   float64 // default 0
	Max   float64 // default: auto-computed from series data
}

// RadarSeries defines one data polygon on the radar chart.
type RadarSeries struct {
	Name   string
	Values []float64 // one per axis, in axis order
	Color  gui.Color // optional; falls back to palette
}

// RadarCfg configures a radar/spider chart.
type RadarCfg struct {
	BaseCfg

	Axes   []RadarAxis
	Series []RadarSeries

	// ShowGrid draws concentric grid lines. Default true.
	ShowGrid bool
	// ShowArea fills data polygons with a translucent color.
	// Default true.
	ShowArea bool
	// FillOpacity controls the alpha of filled polygons.
	// Range [0,1]. Default 0.3.
	FillOpacity float32
	// ShowMarkers draws small circles at each data vertex.
	// Default true.
	ShowMarkers bool
	// LineWidth is the polygon outline width. Default 2.
	LineWidth float32
	// GridLevels is the number of concentric grid rings.
	// Default 5.
	GridLevels int
	// StartAngle is the direction of the first axis in
	// radians. Default -pi/2 (top).
	StartAngle float32
	// PolygonGrid draws polygon grid lines instead of circles.
	PolygonGrid bool
}

type radarView struct {
	cfg      RadarCfg
	hoverPx  float32
	hoverPy  float32
	hovering bool
	hidden   map[int]bool
	cx, cy   float32
	radius   float32
	lastLB   legendBounds
	win      *gui.Window
}

// Radar creates a radar/spider chart view.
func Radar(cfg RadarCfg) gui.View {
	cfg.applyRadarDefaults()
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableRadar(&cfg.BaseCfg, cfg.Axes, cfg.Series)
	}
	return &radarView{cfg: cfg}
}

// Draw renders the chart onto dc for headless export.
func (rv *radarView) Draw(dc *gui.DrawContext) { rv.draw(dc) }

func (rv *radarView) chartTheme() *theme.Theme { return rv.cfg.Theme }

func (rv *radarView) Content() []gui.View { return nil }

func (rv *radarView) GenerateLayout(w *gui.Window) gui.Layout {
	c := &rv.cfg
	hv := loadHover(w, c.ID,
		&rv.hovering, &rv.hoverPx, &rv.hoverPy)
	var hidV uint64
	rv.hidden, hidV = loadHiddenState(w, c.ID)
	rv.lastLB = loadLegendBounds(w, c.ID)
	rv.win = w
	animV := loadAnimVersion(w, c.ID)
	transV := loadTransitionVersion(w, c.ID)
	if c.Animate {
		startEntryAnimation(w, c.ID, c.AnimDuration)
	}
	width, height := resolveSize(c.Width, c.Height, w)
	return gui.DrawCanvas(gui.DrawCanvasCfg{
		ID:           c.ID,
		Sizing:       c.Sizing,
		Width:        width,
		Height:       height,
		Version:      c.Version + hv + hidV + animV + transV,
		Clip:         true,
		OnDraw:       rv.draw,
		OnClick:      rv.internalClick,
		OnHover:      rv.internalHover,
		OnMouseLeave: rv.internalMouseLeave,
	}).GenerateLayout(w)
}

func (rv *radarView) internalClick(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	mx := e.MouseX
	my := e.MouseY
	if idx := legendHitTest(rv.lastLB, mx, my); idx >= 0 {
		e.IsHandled = true
		l.Shape.Version = toggleHidden(w, rv.cfg.ID, idx)
		return
	}
	if rv.cfg.OnClick != nil {
		rv.cfg.OnClick(l, e, w)
	}
}

func (rv *radarView) internalHover(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	rv.hoverPx = e.MouseX - l.Shape.X
	rv.hoverPy = e.MouseY - l.Shape.Y
	rv.hovering = true
	saveHover(w, l, rv.cfg.ID, true, rv.hoverPx, rv.hoverPy)
	if legendHitTest(rv.lastLB, rv.hoverPx, rv.hoverPy) >= 0 {
		w.SetMouseCursorPointingHand()
	} else if rv.hoveredSeriesIndex(rv.hoverPx, rv.hoverPy) >= 0 {
		w.SetMouseCursorPointingHand()
	} else {
		w.SetMouseCursorArrow()
	}
	if rv.cfg.OnHover != nil {
		rv.cfg.OnHover(l, e, w)
	}
}

func (rv *radarView) internalMouseLeave(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	rv.hovering = false
	saveHover(w, l, rv.cfg.ID, false, 0, 0)
	w.SetMouseCursorArrow()
	if rv.cfg.OnMouseLeave != nil {
		rv.cfg.OnMouseLeave(l, e, w)
	}
}

// applyRadarDefaults sets sensible zero-value defaults. When
// the entire appearance block is zero-valued (no field explicitly
// set), enable ShowGrid, ShowArea, and ShowMarkers.
func (cfg *RadarCfg) applyRadarDefaults() {
	cfg.applyDefaults()

	// Detect "nothing configured" and enable all visual features.
	allZero := !cfg.ShowGrid && !cfg.ShowArea && !cfg.ShowMarkers &&
		cfg.FillOpacity == 0 && cfg.LineWidth == 0 &&
		cfg.GridLevels == 0 && cfg.StartAngle == 0
	if allZero {
		cfg.ShowGrid = true
		cfg.ShowArea = true
		cfg.ShowMarkers = true
	}

	if cfg.FillOpacity == 0 {
		cfg.FillOpacity = DefaultAreaOpacity
	}
	if cfg.LineWidth == 0 {
		cfg.LineWidth = DefaultLineWidth
	}
	if cfg.GridLevels == 0 {
		cfg.GridLevels = 5
	}
	if cfg.StartAngle == 0 {
		cfg.StartAngle = -math.Pi / 2
	}

	// Auto-compute axis Max from series data when both Min and
	// Max are zero.
	for i := range cfg.Axes {
		a := &cfg.Axes[i]
		if a.Min == 0 && a.Max == 0 {
			best := 0.0
			for _, s := range cfg.Series {
				if i < len(s.Values) && finite(s.Values[i]) &&
					s.Values[i] > best {
					best = s.Values[i]
				}
			}
			if best == 0 {
				best = 1
			}
			a.Max = best
		}
	}
}

// Validate checks RadarCfg for invalid settings.
func (cfg *RadarCfg) Validate() error {
	var errs []string
	if err := cfg.BaseCfg.Validate(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(cfg.Axes) < 3 {
		errs = append(errs, "need at least 3 axes")
	}
	nAxes := len(cfg.Axes)
	for i, s := range cfg.Series {
		if len(s.Values) != nAxes {
			errs = append(errs, fmt.Sprintf(
				"series %d has %d values, want %d",
				i, len(s.Values), nAxes))
		}
	}
	for i, a := range cfg.Axes {
		if a.Min >= a.Max {
			errs = append(errs, fmt.Sprintf(
				"axis %d: min (%g) >= max (%g)", i, a.Min, a.Max))
		}
	}
	if cfg.GridLevels <= 0 {
		errs = append(errs, "GridLevels must be > 0")
	}
	return buildError("radar", errs)
}

// radarAxisAngle returns the angle in radians for axis index i.
func radarAxisAngle(startAngle float32, i, nAxes int) float32 {
	return startAngle + float32(i)*2*math.Pi/float32(nAxes)
}

// radarNormalize returns the 0-1 fraction of value within
// [axisMin, axisMax], clamped.
func radarNormalize(value, axisMin, axisMax float64) float64 {
	if !finite(value) || !finite(axisMin) || !finite(axisMax) {
		return 0
	}
	if axisMax == axisMin {
		return 0
	}
	return max(0, min(1, (value-axisMin)/(axisMax-axisMin)))
}

// hoveredSeriesIndex returns the index of the series with a
// vertex nearest to (mx, my) within snap distance, or -1.
func (rv *radarView) hoveredSeriesIndex(mx, my float32) int {
	cfg := &rv.cfg
	if rv.radius == 0 || len(cfg.Series) == 0 {
		return -1
	}
	nAxes := len(cfg.Axes)
	snapDist := DefaultMarkerSize * 1.5
	snap2 := snapDist * snapDist

	bestIdx := -1
	bestDist := snap2

	for si, s := range cfg.Series {
		if rv.hidden[si] || len(s.Values) != nAxes {
			continue
		}
		for ai := range nAxes {
			angle := radarAxisAngle(cfg.StartAngle, ai, nAxes)
			frac := float32(radarNormalize(
				s.Values[ai], cfg.Axes[ai].Min, cfg.Axes[ai].Max))
			r := rv.radius * frac
			vx := rv.cx + r*float32(math.Cos(float64(angle)))
			vy := rv.cy + r*float32(math.Sin(float64(angle)))
			dx := mx - vx
			dy := my - vy
			d2 := dx*dx + dy*dy
			if d2 < bestDist {
				bestDist = d2
				bestIdx = si
			}
		}
	}
	return bestIdx
}

// hoveredAxisIndex returns the axis index nearest to (mx, my),
// based on the angle from the chart center.
func (rv *radarView) hoveredAxisIndex(mx, my float32) int {
	if rv.radius == 0 {
		return -1
	}
	nAxes := len(rv.cfg.Axes)
	angle := float32(math.Atan2(
		float64(my-rv.cy), float64(mx-rv.cx)))

	bestIdx := 0
	bestDiff := float32(math.Pi * 2)
	for i := range nAxes {
		a := radarAxisAngle(rv.cfg.StartAngle, i, nAxes)
		// Normalize angular difference to [0, pi] via mod 2pi.
		diff := float32(math.Remainder(float64(angle-a), 2*math.Pi))
		if diff < 0 {
			diff = -diff
		}
		if diff < bestDiff {
			bestDiff = diff
			bestIdx = i
		}
	}
	return bestIdx
}

func (rv *radarView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &rv.cfg
	th := cfg.Theme

	if len(cfg.Axes) < 3 || len(cfg.Series) == 0 {
		slog.Warn("insufficient radar data", "chart", cfg.ID)
		return
	}

	nAxes := len(cfg.Axes)

	// Plot bounds.
	left := th.PaddingLeft
	right := ctx.Width() - th.PaddingRight
	top := th.PaddingTop
	bottom := ctx.Height() - th.PaddingBottom

	names := make([]string, len(cfg.Series))
	for i, s := range cfg.Series {
		names[i] = s.Name
	}
	right -= legendRightReserve(ctx, th, cfg.LegendPosition, names)
	top += legendTopReserve(
		ctx, th, cfg.LegendPosition, names, left, right)
	bottom -= legendBottomReserve(
		ctx, th, cfg.LegendPosition, names, left, right)

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	drawTitle(ctx, cfg.Title, th)

	// Center and radius.
	plotW := right - left
	plotH := bottom - top
	radius := min(plotW, plotH) / 2 * 0.85
	cx := (left + right) / 2
	cy := (top + bottom) / 2

	rv.cx = cx
	rv.cy = cy
	rv.radius = radius

	gridColor := th.GridColor
	if !gridColor.IsSet() {
		gridColor = gui.RGBA(128, 128, 128, 60)
	}

	// Draw grid.
	if cfg.ShowGrid {
		for level := 1; level <= cfg.GridLevels; level++ {
			r := radius * float32(level) / float32(cfg.GridLevels)
			if cfg.PolygonGrid {
				pts := make([]float32, 0, (nAxes+1)*2)
				for i := range nAxes {
					a := radarAxisAngle(cfg.StartAngle, i, nAxes)
					pts = append(pts,
						cx+r*float32(math.Cos(float64(a))),
						cy+r*float32(math.Sin(float64(a))))
				}
				// Close polygon.
				pts = append(pts, pts[0], pts[1])
				ctx.Polyline(pts, gridColor, 1/ctx.Scale())
			} else {
				ctx.Circle(cx, cy, r, gridColor, 1/ctx.Scale())
			}
		}
	}

	// Draw spokes.
	for i := range nAxes {
		a := radarAxisAngle(cfg.StartAngle, i, nAxes)
		ex := cx + radius*float32(math.Cos(float64(a)))
		ey := cy + radius*float32(math.Sin(float64(a)))
		ctx.Line(cx, cy, ex, ey, gridColor, 1/ctx.Scale())
	}

	// Draw axis labels.
	rv.drawAxisLabels(ctx, th, cx, cy, radius)

	progress := animProgress(rv.win, rv.cfg.ID)

	// Determine hovered series.
	hovIdx := -1
	if rv.hovering {
		hovIdx = rv.hoveredSeriesIndex(rv.hoverPx, rv.hoverPy)
	}

	// Draw data polygons.
	for si, s := range cfg.Series {
		if rv.hidden[si] || len(s.Values) != nAxes {
			continue
		}
		color := seriesColor(s.Color, si, th.Palette)
		if hovIdx >= 0 && si != hovIdx {
			color = dimColor(color, HoverDimAlpha)
		}

		pts := make([]float32, 0, nAxes*2)
		for ai := range nAxes {
			a := radarAxisAngle(cfg.StartAngle, ai, nAxes)
			frac := float32(radarNormalize(
				s.Values[ai], cfg.Axes[ai].Min, cfg.Axes[ai].Max))
			r := radius * frac * progress
			pts = append(pts,
				cx+r*float32(math.Cos(float64(a))),
				cy+r*float32(math.Sin(float64(a))))
		}

		// Filled area.
		if cfg.ShowArea {
			alpha := uint8(255 * cfg.FillOpacity)
			if hovIdx >= 0 && si != hovIdx {
				alpha = HoverDimAlpha / 4
			}
			fillColor := gui.RGBA(color.R, color.G, color.B, alpha)
			ctx.FilledPolygon(pts, fillColor)
		}

		// Outline (close the polygon).
		closed := slices.Concat(pts, pts[:2])
		ctx.Polyline(closed, color, cfg.LineWidth)

		// Vertex markers.
		if cfg.ShowMarkers {
			for j := 0; j < len(pts); j += 2 {
				ctx.FilledCircle(
					pts[j], pts[j+1],
					DefaultMarkerSize/2, color)
			}
		}
	}

	// Legend.
	entries := make([]legendEntry, len(cfg.Series))
	for i, s := range cfg.Series {
		color := seriesColor(s.Color, i, th.Palette)
		entries[i] = legendEntry{
			Name: s.Name, Color: color, Index: i,
		}
	}
	rv.lastLB = drawLegend(ctx, entries, th,
		plotRect{left, right, top, bottom},
		cfg.LegendPosition, rv.hidden)
	saveLegendBounds(rv.win, cfg.ID, rv.lastLB)

	// Tooltip.
	if rv.hovering && hovIdx >= 0 {
		rv.tooltipRadar(ctx, th, hovIdx,
			plotRect{left, right, top, bottom})
	}
}

// drawAxisLabels renders axis labels around the perimeter.
func (rv *radarView) drawAxisLabels(
	ctx *render.Context, th *theme.Theme,
	cx, cy, radius float32,
) {
	cfg := &rv.cfg
	nAxes := len(cfg.Axes)
	style := th.TickStyle
	fh := ctx.FontHeight(style)
	labelR := radius + 12

	for i, axis := range cfg.Axes {
		if axis.Label == "" {
			continue
		}
		a := radarAxisAngle(cfg.StartAngle, i, nAxes)
		lx := cx + labelR*float32(math.Cos(float64(a)))
		ly := cy + labelR*float32(math.Sin(float64(a)))

		tw := ctx.TextWidth(axis.Label, style)

		// Adjust anchor based on angle quadrant.
		cos := float32(math.Cos(float64(a)))
		if cos < -0.1 {
			lx -= tw // left-align labels on the left side
		} else if cos > 0.1 {
			// right side: lx is already left edge
		} else {
			lx -= tw / 2 // center labels near top/bottom
		}
		ly -= fh / 2

		ctx.Text(lx, ly, axis.Label, style)
	}
}

// tooltipRadar draws a tooltip for the hovered series showing the
// value at the nearest axis.
func (rv *radarView) tooltipRadar(
	ctx *render.Context, th *theme.Theme, seriesIdx int,
	pr plotRect,
) {
	cfg := &rv.cfg
	s := cfg.Series[seriesIdx]

	ai := rv.hoveredAxisIndex(rv.hoverPx, rv.hoverPy)
	if ai < 0 || ai >= len(cfg.Axes) {
		return
	}

	axis := cfg.Axes[ai]
	var label string
	if s.Name != "" {
		label = fmt.Sprintf("%s\n%s: %g", s.Name, axis.Label, s.Values[ai])
	} else {
		label = fmt.Sprintf("%s: %g", axis.Label, s.Values[ai])
	}
	drawTooltip(ctx, rv.hoverPx, rv.hoverPy, label, th, pr)
}
