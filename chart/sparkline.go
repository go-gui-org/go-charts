package chart

import (
	"cmp"
	"fmt"
	"log/slog"
	"math"

	"github.com/mike-ward/go-charts/axis"
	"github.com/mike-ward/go-charts/render"
	"github.com/mike-ward/go-charts/series"
	"github.com/mike-ward/go-charts/theme"
	"github.com/mike-ward/go-gui/gui"
)

// SparklineType selects the visual variant of a sparkline.
type SparklineType int

// Sparkline variant constants.
const (
	SparklineLine SparklineType = iota
	SparklineBar
	SparklineArea
)

// SparklineCfg configures a sparkline — a compact inline
// mini-chart with no axes, grid, or legend.
type SparklineCfg struct {
	BaseCfg

	// Values provides Y-only data; X is auto-indexed 0..N-1.
	// Ignored when Series is non-empty.
	Values []float64

	// Series provides full XY data. Takes precedence over
	// Values when non-empty.
	Series series.XY

	// Type selects the visual variant. Zero = SparklineLine.
	Type SparklineType

	// LineWidth for line/area variants. 0 = default (1.5).
	LineWidth float32

	// Color for line stroke or bar fill. Zero = palette[0].
	Color gui.Color

	// FillColor for area/bar fill. Zero = auto from Color
	// with reduced alpha.
	FillColor gui.Color

	// ShowReferenceLine draws a horizontal line at
	// ReferenceValue.
	ShowReferenceLine bool
	ReferenceValue    float64
	ReferenceColor    gui.Color

	// BandColoring fills above-reference with BandAboveColor
	// and below-reference with BandBelowColor.
	BandColoring   bool
	BandAboveColor gui.Color
	BandBelowColor gui.Color

	// Markers highlight specific data points.
	ShowMinMarker  bool
	ShowMaxMarker  bool
	ShowLastMarker bool
	MinColor       gui.Color
	MaxColor       gui.Color
	LastColor      gui.Color
	MarkerRadius   float32 // 0 = default (2.5)

	// ShowTooltip enables hover tooltip showing the value.
	ShowTooltip bool
	// ValueFormat is the fmt format for tooltip values.
	// Zero = "%g".
	ValueFormat string
}

type sparklineView struct {
	cfg         SparklineCfg
	resolved    series.XY
	resolvedBuf [1]series.XY // avoids alloc in tooltip path
	ptsBuf      []float32
	clipA       []float32
	clipB       []float32
	hoverPx     float32
	hoverPy     float32
	hovering    bool
	lastVersion uint64
	xAxis       axis.Axis
	yAxis       axis.Axis
	lastPA      plotArea
	win         *gui.Window
}

// Sparkline creates a sparkline view.
func Sparkline(cfg SparklineCfg) gui.View {
	applySparklineDefaults(&cfg)
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableSparkline(&cfg.BaseCfg, cfg.Values, cfg.Series)
	}
	sv := &sparklineView{cfg: cfg}
	sv.resolveData()
	return sv
}

func applySparklineDefaults(cfg *SparklineCfg) {
	cfg.applyDefaults()
	// Sparklines default to compact fixed-height.
	if cfg.Sizing == gui.FillFill {
		cfg.Sizing = gui.FillFixed
	}
	cfg.Height = cmp.Or(cfg.Height, DefaultSparklineHeight)
	cfg.LineWidth = cmp.Or(cfg.LineWidth, DefaultSparklineLineWidth)
	cfg.MarkerRadius = cmp.Or(cfg.MarkerRadius, DefaultSparklineMarkerRadius)
	cfg.ValueFormat = cmp.Or(cfg.ValueFormat, "%g")
	// Shallow-copy theme and zero padding for sparklines.
	th := *cfg.Theme
	th.PaddingTop = 0
	th.PaddingRight = 0
	th.PaddingBottom = 0
	th.PaddingLeft = 0
	cfg.Theme = &th
}

func (sv *sparklineView) resolveData() {
	if sv.cfg.Series.Len() > 0 {
		sv.resolved = sv.cfg.Series
		return
	}
	pts := make([]series.Point, len(sv.cfg.Values))
	for i, y := range sv.cfg.Values {
		pts[i] = series.Point{X: float64(i), Y: y}
	}
	sv.resolved = series.NewXY(series.XYCfg{Points: pts})
}

// Draw renders the sparkline onto dc for headless export.
func (sv *sparklineView) Draw(dc *gui.DrawContext) { sv.draw(dc) }

func (sv *sparklineView) chartTheme() *theme.Theme {
	return sv.cfg.Theme
}

func (sv *sparklineView) Content() []gui.View { return nil }

func (sv *sparklineView) GenerateLayout(
	w *gui.Window,
) gui.Layout {
	c := &sv.cfg
	sv.win = w
	hv := loadHover(w, c.ID,
		&sv.hovering, &sv.hoverPx, &sv.hoverPy)
	animV := loadAnimVersion(w, c.ID)
	transV := loadTransitionVersion(w, c.ID)
	if c.Animate {
		startEntryAnimation(w, c.ID, c.AnimDuration)
	}
	width, height := resolveSize(c.Width, c.Height, w)

	dcCfg := gui.DrawCanvasCfg{
		ID:      c.ID,
		Sizing:  c.Sizing,
		Width:   width,
		Height:  height,
		Version: c.Version + hv + animV + transV,
		Clip:    true,
		OnDraw:  sv.draw,
	}
	if c.ShowTooltip || c.OnHover != nil {
		dcCfg.OnHover = sv.internalHover
		dcCfg.OnMouseLeave = sv.internalMouseLeave
	}
	if c.OnClick != nil {
		dcCfg.OnClick = sv.internalClick
	}
	return gui.DrawCanvas(dcCfg).GenerateLayout(w)
}

func (sv *sparklineView) internalClick(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	if sv.cfg.OnClick != nil {
		sv.cfg.OnClick(l, e, w)
	}
}

func (sv *sparklineView) internalHover(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	sv.hoverPx = e.MouseX - l.Shape.X
	sv.hoverPy = e.MouseY - l.Shape.Y
	sv.hovering = true
	saveHover(w, l, sv.cfg.ID, true, sv.hoverPx, sv.hoverPy)
	if sv.cfg.OnHover != nil {
		sv.cfg.OnHover(l, e, w)
	}
}

func (sv *sparklineView) internalMouseLeave(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	sv.hovering = false
	saveHover(w, l, sv.cfg.ID, false, 0, 0)
	if sv.cfg.OnMouseLeave != nil {
		sv.cfg.OnMouseLeave(l, e, w)
	}
}

func (sv *sparklineView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &sv.cfg
	s := sv.resolved

	if s.Len() == 0 {
		return
	}

	// Full canvas as plot area with tiny margin for strokes.
	margin := cfg.LineWidth
	if cfg.Type == SparklineBar {
		margin = 0
	}
	left := margin
	right := ctx.Width() - margin
	top := margin
	bottom := ctx.Height() - margin

	if right <= left || bottom <= top {
		return
	}

	// Recompute axes only when version changes.
	if sv.xAxis == nil || cfg.Version != sv.lastVersion {
		if !sv.buildAxes() {
			return
		}
	}

	pr := plotRect{left, right, top, bottom}
	sv.lastPA = plotArea{pr, sv.xAxis, sv.yAxis}

	progress := animProgress(sv.win, sv.cfg.ID)
	color := seriesColor(cfg.Color, 0, cfg.Theme.Palette)

	switch cfg.Type {
	case SparklineBar:
		sv.drawBars(ctx, color, left, right, top, bottom, progress)
	case SparklineArea:
		sv.drawLineOrArea(ctx, color, left, right, top, bottom,
			true, progress)
	default:
		sv.drawLineOrArea(ctx, color, left, right, top, bottom,
			false, progress)
	}

	// Reference line.
	if cfg.ShowReferenceLine && finite(cfg.ReferenceValue) {
		ry := sv.yAxis.Transform(
			cfg.ReferenceValue, bottom, top)
		if ry >= top && ry <= bottom {
			rc := cfg.ReferenceColor
			if rc == (gui.Color{}) {
				rc = gui.RGBA(128, 128, 128, 128)
			}
			ctx.Line(left, ry, right, ry, rc, 1/ctx.Scale())
		}
	}

	// Markers.
	sv.drawMarkers(ctx, color, left, right, top, bottom)

	// Tooltip.
	if sv.hovering && cfg.ShowTooltip {
		sv.drawSparkTooltip(ctx, pr)
	}
}

func (sv *sparklineView) buildAxes() bool {
	s := sv.resolved
	minX, maxX, minY, maxY := s.Bounds()
	if !finite(minX) || !finite(maxX) ||
		!finite(minY) || !finite(maxY) {
		return false
	}
	if minX == maxX {
		minX -= 0.5
		maxX += 0.5
	}

	yRange := maxY - minY
	if yRange == 0 {
		yRange = 1
	}
	minY -= yRange * 0.05
	maxY += yRange * 0.05

	// Include reference value in Y range if shown.
	if sv.cfg.ShowReferenceLine || sv.cfg.BandColoring {
		ref := sv.cfg.ReferenceValue
		if finite(ref) {
			minY = min(minY, ref-yRange*0.05)
			maxY = max(maxY, ref+yRange*0.05)
		}
	}

	sv.xAxis = axis.NewLinear(
		axis.LinearCfg{AutoRange: true})
	sv.xAxis.SetRange(minX, maxX)

	sv.yAxis = axis.NewLinear(
		axis.LinearCfg{AutoRange: true})
	sv.yAxis.SetRange(minY, maxY)

	sv.lastVersion = sv.cfg.Version
	return true
}

func (sv *sparklineView) drawLineOrArea(
	ctx *render.Context, color gui.Color,
	left, right, top, bottom float32,
	fillArea bool, progress float32,
) {
	cfg := &sv.cfg
	s := sv.resolved

	// Build polyline points.
	n := s.Len()
	if progress < 1 {
		n = max(1, int(float32(n)*progress))
	}
	needed := n * 2
	if cap(sv.ptsBuf) < needed {
		sv.ptsBuf = make([]float32, 0, needed)
	}
	pts := sv.ptsBuf[:0]
	count := 0
	for _, p := range s.Points {
		if count >= n {
			break
		}
		if !finite(p.X) || !finite(p.Y) {
			continue
		}
		px := sv.xAxis.Transform(p.X, left, right)
		py := sv.yAxis.Transform(p.Y, bottom, top)
		pts = append(pts, px, py)
		count++
	}
	sv.ptsBuf = pts

	clipped := clipPolylineToRect(pts, left, right, top, bottom)

	// Band coloring fills above/below reference with different
	// colors.
	if cfg.BandColoring && len(pts) >= 4 &&
		finite(cfg.ReferenceValue) {
		refY := sv.yAxis.Transform(
			cfg.ReferenceValue, bottom, top)
		aboveC := cfg.BandAboveColor
		if aboveC == (gui.Color{}) {
			aboveC = gui.RGBA(0, 180, 0, 50)
		}
		belowC := cfg.BandBelowColor
		if belowC == (gui.Color{}) {
			belowC = gui.RGBA(220, 0, 0, 50)
		}
		sv.fillBands(ctx, pts, refY,
			left, right, top, bottom, aboveC, belowC)
	} else if fillArea && len(pts) >= 4 {
		// Standard area fill.
		fillC := cfg.FillColor
		if fillC == (gui.Color{}) {
			fillC = gui.RGBA(color.R, color.G, color.B, 50)
		}
		var quad [8]float32
		for k := 0; k < len(pts)-2; k += 2 {
			quad[0] = pts[k]
			quad[1] = pts[k+1]
			quad[2] = pts[k+2]
			quad[3] = pts[k+3]
			quad[4] = pts[k+2]
			quad[5] = bottom
			quad[6] = pts[k]
			quad[7] = bottom
			var clippedQ []float32
			clippedQ, sv.clipA, sv.clipB = clipConvexToRect(
				quad[:], left, right, top, bottom,
				sv.clipA, sv.clipB)
			if clippedQ != nil {
				ctx.FilledPolygon(clippedQ, fillC)
			}
		}
	}

	ctx.Polyline(clipped, color, cfg.LineWidth)
}

// fillBands fills the region between the polyline and refY,
// using aboveColor above the reference and belowColor below.
func (sv *sparklineView) fillBands(
	ctx *render.Context, pts []float32, refY float32,
	left, right, top, bottom float32,
	aboveColor, belowColor gui.Color,
) {
	var quad [8]float32
	for k := 0; k < len(pts)-2; k += 2 {
		py0 := pts[k+1]
		py1 := pts[k+3]
		// Above reference: fill from line up to refY (note:
		// screen Y is inverted, above = smaller Y).
		// Below reference: fill from refY down to line.
		// Split at refY for each segment.
		if py0 <= refY && py1 <= refY {
			// Entire segment above reference.
			quad[0] = pts[k]
			quad[1] = py0
			quad[2] = pts[k+2]
			quad[3] = py1
			quad[4] = pts[k+2]
			quad[5] = refY
			quad[6] = pts[k]
			quad[7] = refY
			sv.fillQuad(ctx, quad, left, right,
				top, bottom, aboveColor)
		} else if py0 >= refY && py1 >= refY {
			// Entire segment below reference.
			quad[0] = pts[k]
			quad[1] = refY
			quad[2] = pts[k+2]
			quad[3] = refY
			quad[4] = pts[k+2]
			quad[5] = py1
			quad[6] = pts[k]
			quad[7] = py0
			sv.fillQuad(ctx, quad, left, right,
				top, bottom, belowColor)
		} else {
			// Segment crosses reference — split at
			// intersection.
			dx := pts[k+2] - pts[k]
			dy := py1 - py0
			t := (refY - py0) / dy
			mx := pts[k] + t*dx

			// First half: pts[k] to mx.
			if py0 <= refY {
				quad[0] = pts[k]
				quad[1] = py0
				quad[2] = mx
				quad[3] = refY
				quad[4] = mx
				quad[5] = refY
				quad[6] = pts[k]
				quad[7] = refY
				sv.fillQuad(ctx, quad, left, right,
					top, bottom, aboveColor)
				quad[0] = mx
				quad[1] = refY
				quad[2] = pts[k+2]
				quad[3] = py1
				quad[4] = pts[k+2]
				quad[5] = refY
				quad[6] = mx
				quad[7] = refY
				sv.fillQuad(ctx, quad, left, right,
					top, bottom, belowColor)
			} else {
				quad[0] = pts[k]
				quad[1] = refY
				quad[2] = mx
				quad[3] = refY
				quad[4] = mx
				quad[5] = refY
				quad[6] = pts[k]
				quad[7] = py0
				sv.fillQuad(ctx, quad, left, right,
					top, bottom, belowColor)
				quad[0] = mx
				quad[1] = refY
				quad[2] = pts[k+2]
				quad[3] = py1
				quad[4] = pts[k+2]
				quad[5] = refY
				quad[6] = mx
				quad[7] = refY
				sv.fillQuad(ctx, quad, left, right,
					top, bottom, aboveColor)
			}
		}
	}
}

func (sv *sparklineView) fillQuad(
	ctx *render.Context, quad [8]float32,
	left, right, top, bottom float32,
	color gui.Color,
) {
	var clippedQ []float32
	clippedQ, sv.clipA, sv.clipB = clipConvexToRect(
		quad[:], left, right, top, bottom,
		sv.clipA, sv.clipB)
	if clippedQ != nil {
		ctx.FilledPolygon(clippedQ, color)
	}
}

func (sv *sparklineView) drawBars(
	ctx *render.Context, color gui.Color,
	left, right, top, bottom float32, progress float32,
) {
	cfg := &sv.cfg
	s := sv.resolved
	n := s.Len()
	if n == 0 {
		return
	}

	totalW := right - left
	gap := DefaultSparklineBarGap
	barW := (totalW - float32(n-1)*gap) / float32(n)
	if barW < 1 {
		gap = 0
		barW = totalW / float32(n)
	}

	// Reference Y for bar baseline.
	refVal := 0.0
	if (cfg.ShowReferenceLine || cfg.BandColoring) &&
		finite(cfg.ReferenceValue) {
		refVal = cfg.ReferenceValue
	}
	refPy := sv.yAxis.Transform(refVal, bottom, top)
	refPy = max(top, min(bottom, refPy))

	aboveC := color
	belowC := color
	if cfg.BandColoring {
		aboveC = cfg.BandAboveColor
		if aboveC == (gui.Color{}) {
			aboveC = gui.RGBA(0, 180, 0, 180)
		}
		belowC = cfg.BandBelowColor
		if belowC == (gui.Color{}) {
			belowC = gui.RGBA(220, 0, 0, 180)
		}
	}

	fillC := cfg.FillColor
	useFill := fillC != (gui.Color{}) && !cfg.BandColoring

	for i, p := range s.Points {
		if !finite(p.Y) {
			continue
		}
		bx := left + float32(i)*(barW+gap)
		py := sv.yAxis.Transform(p.Y, bottom, top)
		py = max(top, min(bottom, py))

		var by, bh float32
		if py < refPy {
			by = py
			bh = refPy - py
		} else {
			by = refPy
			bh = py - refPy
		}
		bh *= progress
		if py < refPy {
			by = refPy - bh
		}
		if bh < 0.5 {
			bh = 0.5
		}

		bc := aboveC
		if cfg.BandColoring && py > refPy {
			bc = belowC
		}
		if useFill {
			bc = fillC
		}
		ctx.FilledRect(bx, by, barW, bh, bc)
	}
}

func (sv *sparklineView) drawMarkers(
	ctx *render.Context, color gui.Color,
	left, right, top, bottom float32,
) {
	cfg := &sv.cfg
	if !cfg.ShowMinMarker && !cfg.ShowMaxMarker &&
		!cfg.ShowLastMarker {
		return
	}

	s := sv.resolved
	minIdx, maxIdx, lastIdx := -1, -1, -1
	minVal := math.MaxFloat64
	maxVal := -math.MaxFloat64

	for i, p := range s.Points {
		if !finite(p.Y) {
			continue
		}
		lastIdx = i
		if p.Y < minVal {
			minVal = p.Y
			minIdx = i
		}
		if p.Y > maxVal {
			maxVal = p.Y
			maxIdx = i
		}
	}

	r := cfg.MarkerRadius

	drawMark := func(idx int, c gui.Color) {
		if idx < 0 {
			return
		}
		p := s.Points[idx]
		px := sv.xAxis.Transform(p.X, left, right)
		py := sv.yAxis.Transform(p.Y, bottom, top)
		if px >= left && px <= right &&
			py >= top && py <= bottom {
			ctx.FilledCircle(px, py, r, c)
		}
	}

	if cfg.ShowMinMarker {
		mc := cfg.MinColor
		if mc == (gui.Color{}) {
			mc = gui.Hex(0xE15759)
		}
		drawMark(minIdx, mc)
	}
	if cfg.ShowMaxMarker {
		mc := cfg.MaxColor
		if mc == (gui.Color{}) {
			mc = gui.Hex(0x59A14F)
		}
		drawMark(maxIdx, mc)
	}
	if cfg.ShowLastMarker {
		mc := cfg.LastColor
		if mc == (gui.Color{}) {
			mc = color
		}
		drawMark(lastIdx, mc)
	}
}

func (sv *sparklineView) drawSparkTooltip(
	ctx *render.Context, pr plotRect,
) {
	pa := sv.lastPA
	// Reuse fixed-size array to avoid per-frame allocation.
	sv.resolvedBuf[0] = sv.resolved
	_, pi, _, _, ok := nearestXYPoint(
		sv.resolvedBuf[:], pa, sv.hoverPx, sv.hoverPy, 40)
	if !ok {
		return
	}
	p := sv.resolved.Points[pi]
	label := fmt.Sprintf(sv.cfg.ValueFormat, p.Y)
	px := pa.XAxis.Transform(p.X, pa.Left, pa.Right)
	py := pa.YAxis.Transform(p.Y, pa.Bottom, pa.Top)
	drawTooltip(ctx, px, py, label, sv.cfg.Theme, pr)
}
