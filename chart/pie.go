package chart

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// PieSlice represents a single slice of a pie chart.
type PieSlice struct {
	Label string
	Value float64
	Color gui.Color
}

// PieCfg configures a pie or donut chart.
type PieCfg struct {
	BaseCfg

	// Data
	Slices []PieSlice

	// Appearance
	InnerRadius float32 // >0 makes it a donut chart
	StartAngle  float32 // in radians
	ShowLabels  bool
	ShowPercent bool
}

type pieView struct {
	cfg      PieCfg
	hoverPx  float32
	hoverPy  float32
	hovering bool
	hidden   map[int]bool // legend toggle state
	// Cached geometry for cursor hit-testing.
	cx, cy, outerR float32
	lastLB         legendBounds
	win            *gui.Window
}

// Pie creates a pie or donut chart view.
func Pie(cfg PieCfg) gui.View {
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableSlices(&cfg.BaseCfg, cfg.Slices)
	}
	return &pieView{cfg: cfg}
}

// Draw renders the chart onto dc for headless export.
func (pv *pieView) Draw(dc *gui.DrawContext) { pv.draw(dc) }

func (pv *pieView) chartTheme() *theme.Theme { return pv.cfg.Theme }

func (pv *pieView) Content() []gui.View { return nil }

func (pv *pieView) GenerateLayout(w *gui.Window) gui.Layout {
	c := &pv.cfg
	hv := loadHover(w, c.ID,
		&pv.hovering, &pv.hoverPx, &pv.hoverPy)
	var hidV uint64
	pv.hidden, hidV = loadHiddenState(w, c.ID)
	pv.lastLB = loadLegendBounds(w, c.ID)
	pv.win = w
	av := loadAnimVersion(w, c.ID)
	tv := loadTransitionVersion(w, c.ID)
	if c.Animate {
		startEntryAnimation(w, c.ID, c.AnimDuration)
	}
	width, height := resolveSize(c.Width, c.Height, w)
	return gui.DrawCanvas(gui.DrawCanvasCfg{
		ID:           c.ID,
		Sizing:       c.Sizing,
		Width:        width,
		Height:       height,
		Version:      c.Version + hv + hidV + av + tv,
		Clip:         true,
		OnDraw:       pv.draw,
		OnClick:      pv.internalClick,
		OnHover:      pv.internalHover,
		OnMouseLeave: pv.internalMouseLeave,
	}).GenerateLayout(w)
}

func (pv *pieView) internalClick(l *gui.Layout, e *gui.Event, w *gui.Window) {
	mx := e.MouseX
	my := e.MouseY
	if idx := legendHitTest(pv.lastLB, mx, my); idx >= 0 {
		e.IsHandled = true
		l.Shape.Version = toggleHidden(w, pv.cfg.ID, idx)
		return
	}
	if pv.cfg.OnClick != nil {
		pv.cfg.OnClick(l, e, w)
	}
}

func (pv *pieView) internalHover(l *gui.Layout, e *gui.Event, w *gui.Window) {
	e.IsHandled = true
	pv.hoverPx = e.MouseX - l.Shape.X
	pv.hoverPy = e.MouseY - l.Shape.Y
	pv.hovering = true
	saveHover(w, l, pv.cfg.ID, true, pv.hoverPx, pv.hoverPy)
	if legendHitTest(pv.lastLB, pv.hoverPx, pv.hoverPy) >= 0 {
		w.SetMouseCursorPointingHand()
	} else if pv.outerR > 0 &&
		pv.hoveredSliceIndex(pv.hoverPx, pv.hoverPy, pv.cx, pv.cy, pv.outerR) >= 0 {
		w.SetMouseCursorPointingHand()
	} else {
		w.SetMouseCursorArrow()
	}
	if pv.cfg.OnHover != nil {
		pv.cfg.OnHover(l, e, w)
	}
}

func (pv *pieView) internalMouseLeave(l *gui.Layout, e *gui.Event, w *gui.Window) {
	e.IsHandled = true
	pv.hovering = false
	saveHover(w, l, pv.cfg.ID, false, 0, 0)
	w.SetMouseCursorArrow()
	if pv.cfg.OnMouseLeave != nil {
		pv.cfg.OnMouseLeave(l, e, w)
	}
}

// normAngle returns a such that a >= ref, by adding whole multiples of 2π.
// Uses a single ceil call instead of a loop to avoid O(n) spinning on large
// StartAngle values.
func normAngle(a, ref float32) float32 {
	if diff := ref - a; diff > 0 {
		a += float32(math.Ceil(float64(diff/(2*math.Pi)))) * (2 * math.Pi)
	}
	return a
}

// hoveredSliceIndex returns the index of the slice under (mx, my),
// or -1 if none. Each slice is tested against its exploded center so
// that hit-testing matches the drawn geometry exactly. Hidden slices
// are excluded so angles match the rendered layout.
func (pv *pieView) hoveredSliceIndex(mx, my, cx, cy, outerR float32) int {
	cfg := &pv.cfg
	if len(cfg.Slices) == 0 {
		return -1
	}

	total := 0.0
	for i, s := range cfg.Slices {
		if s.Value > 0 && !pv.hidden[i] {
			total += s.Value
		}
	}
	if total == 0 {
		return -1
	}

	cumAngle := cfg.StartAngle
	for i, s := range cfg.Slices {
		if s.Value <= 0 || pv.hidden[i] {
			continue
		}
		sweep := float32(s.Value/total) * (2 * math.Pi)
		mid := cumAngle + sweep/2

		// Test against the exploded center so the hit region matches the
		// drawn slice position (explode offset applied in draw()).
		ocx := cx + HoverExplodeDist*float32(math.Cos(float64(mid)))
		ocy := cy + HoverExplodeDist*float32(math.Sin(float64(mid)))
		dx := mx - ocx
		dy := my - ocy
		r2 := dx*dx + dy*dy

		if r2 <= outerR*outerR &&
			(cfg.InnerRadius <= 0 || r2 >= cfg.InnerRadius*cfg.InnerRadius) {
			a := normAngle(float32(math.Atan2(float64(dy), float64(dx))), cumAngle)
			if a < cumAngle+sweep {
				return i
			}
		}
		cumAngle += sweep
	}
	return -1
}

// tooltipPie draws a tooltip for the slice under the cursor.
// Delegates hit-testing to hoveredSliceIndex so both use the same geometry.
func (pv *pieView) tooltipPie(
	ctx *render.Context, pr plotRect, th *theme.Theme,
) {
	cfg := &pv.cfg
	if len(cfg.Slices) == 0 {
		return
	}

	plotW := pr.Right - pr.Left
	plotH := pr.Bottom - pr.Top
	outerR := min(plotW, plotH) / 2 * 0.85
	cx := (pr.Left + pr.Right) / 2
	cy := (pr.Top + pr.Bottom) / 2

	idx := pv.hoveredSliceIndex(pv.hoverPx, pv.hoverPy, cx, cy, outerR)
	if idx < 0 {
		return
	}

	s := cfg.Slices[idx]
	total := 0.0
	for i, sl := range cfg.Slices {
		if sl.Value > 0 && !pv.hidden[i] {
			total += sl.Value
		}
	}
	if total == 0 {
		return
	}
	pct := s.Value / total * 100
	var label string
	if s.Label != "" {
		label = fmt.Sprintf("%s: %g (%.1f%%)", s.Label, s.Value, pct)
	} else {
		label = fmt.Sprintf("%g (%.1f%%)", s.Value, pct)
	}
	drawTooltip(ctx, pv.hoverPx, pv.hoverPy, label, th, pr)
}

func (pv *pieView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &pv.cfg
	th := cfg.Theme

	if len(cfg.Slices) == 0 {
		slog.Warn("no slice data", "chart", cfg.ID)
		return
	}

	left := th.PaddingLeft
	right := ctx.Width() - th.PaddingRight
	top := th.PaddingTop
	bottom := ctx.Height() - th.PaddingBottom

	names := make([]string, len(cfg.Slices))
	for i, s := range cfg.Slices {
		names[i] = s.Label
	}
	right -= legendRightReserve(ctx, th, cfg.LegendPosition, names)
	top += legendTopReserve(ctx, th, cfg.LegendPosition, names, left, right)
	bottom -= legendBottomReserve(ctx, th, cfg.LegendPosition, names, left, right)

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	drawTitle(ctx, cfg.Title, th)

	// Sum visible positive slice values.
	total := 0.0
	for i, s := range cfg.Slices {
		if s.Value > 0 && !pv.hidden[i] {
			total += s.Value
		}
	}
	if total == 0 {
		slog.Warn("all slice values zero or negative", "chart", cfg.ID)
		return
	}

	// Outer radius — leave 15% margin for labels.
	plotW := right - left
	plotH := bottom - top
	outerR := min(plotW, plotH) / 2 * 0.85
	cx := (left + right) / 2
	cy := (top + bottom) / 2

	// Cache geometry for cursor hit-testing in hover callback.
	pv.cx = cx
	pv.cy = cy
	pv.outerR = outerR

	// Hover highlight: find the slice under the cursor.
	hovIdx := -1
	if pv.hovering {
		hovIdx = pv.hoveredSliceIndex(pv.hoverPx, pv.hoverPy, cx, cy, outerR)
	}

	progress := animProgress(pv.win, cfg.ID)

	// Draw slices (skip hidden).
	angle := cfg.StartAngle
	for i, s := range cfg.Slices {
		if s.Value <= 0 || pv.hidden[i] {
			continue
		}
		sweep := float32(s.Value/total) * (2 * math.Pi) * progress
		color := s.Color
		if !color.IsSet() {
			color = seriesColor(gui.Color{}, i, th.Palette)
		}
		// Explode the hovered slice outward.
		ocx, ocy := cx, cy
		if hovIdx >= 0 && i == hovIdx {
			mid := angle + sweep/2
			ocx += HoverExplodeDist * float32(math.Cos(float64(mid)))
			ocy += HoverExplodeDist * float32(math.Sin(float64(mid)))
		}
		ctx.FilledArc(ocx, ocy, outerR, outerR, angle, sweep, color)
		angle += sweep
	}

	// Donut hole: overdraw center with background color.
	// Note: this only looks correct over a solid background.
	if cfg.InnerRadius > 0 {
		ctx.FilledCircle(cx, cy, cfg.InnerRadius, th.Background)
		// Draw a second hole at the exploded slice center so
		// the shifted geometry is also hollowed out.
		if hovIdx >= 0 {
			s := cfg.Slices[hovIdx]
			if s.Value > 0 {
				cumA := cfg.StartAngle
				for j, sl := range cfg.Slices[:hovIdx] {
					if sl.Value > 0 && !pv.hidden[j] {
						cumA += float32(sl.Value/total) * (2 * math.Pi)
					}
				}
				sweep := float32(s.Value/total) * (2 * math.Pi)
				mid := cumA + sweep/2
				ecx := cx + HoverExplodeDist*float32(math.Cos(float64(mid)))
				ecy := cy + HoverExplodeDist*float32(math.Sin(float64(mid)))
				ctx.FilledCircle(ecx, ecy, cfg.InnerRadius, th.Background)
			}
		}
	}

	pv.drawLabels(ctx, cfg, th, cx, cy, outerR, total)

	// Legend.
	entries := make([]legendEntry, len(cfg.Slices))
	for i, s := range cfg.Slices {
		color := s.Color
		if !color.IsSet() {
			color = seriesColor(gui.Color{}, i, th.Palette)
		}
		entries[i] = legendEntry{Name: s.Label, Color: color, Index: i}
	}
	pv.lastLB = drawLegend(ctx, entries, th,
		plotRect{left, right, top, bottom},
		cfg.LegendPosition, pv.hidden)
	saveLegendBounds(pv.win, cfg.ID, pv.lastLB)

	// Tooltip.
	if pv.hovering {
		pv.tooltipPie(ctx, plotRect{left, right, top, bottom}, th)
	}
}

// drawLabels renders percentage and name labels at each visible
// slice midpoint.
func (pv *pieView) drawLabels(
	ctx *render.Context, cfg *PieCfg, th *theme.Theme,
	cx, cy, outerR float32, total float64,
) {
	if !cfg.ShowLabels && !cfg.ShowPercent {
		return
	}
	style := th.TickStyle
	fh := ctx.FontHeight(style)
	labelR := outerR * 0.7
	if cfg.InnerRadius > 0 {
		labelR = (cfg.InnerRadius + outerR) / 2
	}

	angle := cfg.StartAngle
	for i, s := range cfg.Slices {
		if s.Value <= 0 || pv.hidden[i] {
			continue
		}
		sweep := float32(s.Value/total) * (2 * math.Pi)
		mid := angle + sweep/2
		lx := cx + labelR*float32(math.Cos(float64(mid)))
		ly := cy + labelR*float32(math.Sin(float64(mid)))

		text := ""
		if cfg.ShowLabels {
			text = s.Label
		}
		if cfg.ShowPercent {
			pct := fmt.Sprintf("%.1f%%", s.Value/total*100)
			if text != "" {
				text = text + " " + pct
			} else {
				text = pct
			}
		}
		if text == "" {
			angle += sweep
			continue
		}

		tw := ctx.TextWidth(text, style)
		ctx.Text(lx-tw/2, ly-fh/2, text, style)
		angle += sweep
	}
}
