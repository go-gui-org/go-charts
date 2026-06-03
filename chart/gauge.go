package chart

import (
	"cmp"
	"fmt"
	"log/slog"
	"math"

	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// GaugeZone defines a colored range on the gauge arc.
// Threshold is the upper bound of this zone (in data units).
type GaugeZone struct {
	Label     string
	Threshold float64
	Color     gui.Color
}

// GaugeCfg configures a gauge chart.
type GaugeCfg struct {
	BaseCfg

	// Value is the current reading displayed by the gauge.
	Value float64

	// Min and Max define the gauge range. Defaults: 0 and 100.
	Min, Max float64

	// ArcAngle is the total arc sweep in radians.
	// Default: 270° (3π/2). The gap is centered at the bottom.
	ArcAngle float32

	// Zones define colored ranges on the gauge background.
	// Each zone spans from the previous zone's threshold (or Min)
	// to its own Threshold. Thresholds must be ascending and
	// within [Min, Max].
	Zones []GaugeZone

	// InnerRatio is the inner radius as a fraction of the outer
	// radius (0–1). Default: 0.7. Controls arc thickness.
	InnerRatio float32

	// ShowValue renders the current value as centered text.
	ShowValue bool

	// ShowMinMax renders min/max labels at the arc endpoints.
	ShowMinMax bool

	// ShowPointer draws a triangular needle pointing at the
	// current value, with a hub circle at the center.
	ShowPointer bool

	// ValueFormat is the fmt format string for the value label.
	// Default: "%.0f".
	ValueFormat string
}

type gaugeView struct {
	cfg      GaugeCfg
	hoverPx  float32
	hoverPy  float32
	hovering bool
	// Cached geometry for cursor hit-testing.
	cx, cy, outerR, innerR float32
	win                    *gui.Window
}

// Gauge creates a gauge chart view.
func Gauge(cfg GaugeCfg) gui.View {
	cfg.applyGaugeDefaults()
	if err := cfg.Validate(); err != nil {
		slog.Warn("invalid config", "error", err)
	}
	if cfg.ShowDataTable {
		return dataTableGauge(&cfg.BaseCfg, cfg.Value, cfg.Min, cfg.Max)
	}
	return &gaugeView{cfg: cfg}
}

func (cfg *GaugeCfg) applyGaugeDefaults() {
	cfg.applyDefaults()
	if cfg.Min == 0 && cfg.Max == 0 {
		cfg.Max = 100
	}
	cfg.ArcAngle = cmp.Or(cfg.ArcAngle, DefaultGaugeArcAngle)
	cfg.InnerRatio = cmp.Or(cfg.InnerRatio, DefaultGaugeInnerRatio)
	cfg.ValueFormat = cmp.Or(cfg.ValueFormat, "%.0f")
}

// Validate checks GaugeCfg for invalid settings.
func (cfg *GaugeCfg) Validate() error {
	var errs []string
	if err := cfg.BaseCfg.Validate(); err != nil {
		errs = append(errs, err.Error())
	}
	if cfg.Min >= cfg.Max {
		errs = append(errs, "Min >= Max")
	}
	if cfg.ArcAngle <= 0 || cfg.ArcAngle > 2*math.Pi {
		errs = append(errs, "ArcAngle out of (0, 2π]")
	}
	if cfg.InnerRatio < 0 || cfg.InnerRatio >= 1 {
		errs = append(errs, "InnerRatio out of [0, 1)")
	}
	// Validate zone thresholds are ascending and within range.
	prev := cfg.Min
	for i, z := range cfg.Zones {
		if z.Threshold <= prev {
			errs = append(errs, fmt.Sprintf(
				"zone %d threshold %.4g <= previous %.4g", i, z.Threshold, prev))
		}
		if z.Threshold > cfg.Max {
			errs = append(errs, fmt.Sprintf(
				"zone %d threshold %.4g > Max %.4g", i, z.Threshold, cfg.Max))
		}
		prev = z.Threshold
	}
	return buildError("gauge", errs)
}

// Draw renders the chart onto dc for headless export.
func (gv *gaugeView) Draw(dc *gui.DrawContext) { gv.draw(dc) }

func (gv *gaugeView) chartTheme() *theme.Theme { return gv.cfg.Theme }

func (gv *gaugeView) Content() []gui.View { return nil }

func (gv *gaugeView) GenerateLayout(w *gui.Window) gui.Layout {
	c := &gv.cfg
	hv := loadHover(w, c.ID,
		&gv.hovering, &gv.hoverPx, &gv.hoverPy)
	gv.win = w
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
		Version:      c.Version + hv + av + tv,
		Clip:         true,
		OnDraw:       gv.draw,
		OnClick:      c.OnClick,
		OnHover:      gv.internalHover,
		OnMouseLeave: gv.internalMouseLeave,
	}).GenerateLayout(w)
}

func (gv *gaugeView) internalHover(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	gv.hoverPx = e.MouseX - l.Shape.X
	gv.hoverPy = e.MouseY - l.Shape.Y
	gv.hovering = true
	saveHover(w, l, gv.cfg.ID, true, gv.hoverPx, gv.hoverPy)
	if gv.outerR > 0 && gv.gaugeHitTest(gv.hoverPx, gv.hoverPy) {
		w.SetMouseCursorPointingHand()
	} else {
		w.SetMouseCursorArrow()
	}
	if gv.cfg.OnHover != nil {
		gv.cfg.OnHover(l, e, w)
	}
}

func (gv *gaugeView) internalMouseLeave(
	l *gui.Layout, e *gui.Event, w *gui.Window,
) {
	e.IsHandled = true
	gv.hovering = false
	saveHover(w, l, gv.cfg.ID, false, 0, 0)
	w.SetMouseCursorArrow()
	if gv.cfg.OnMouseLeave != nil {
		gv.cfg.OnMouseLeave(l, e, w)
	}
}

// gaugeStartAngle returns the angle where the gauge arc begins,
// placing the gap centered at the bottom.
func gaugeStartAngle(arcAngle float32) float32 {
	return math.Pi/2 + (2*math.Pi-arcAngle)/2
}

// gaugeValueFraction returns the clamped fraction of value
// within [lo, hi].
func gaugeValueFraction(value, lo, hi float64) float64 {
	if hi == lo ||
		math.IsNaN(value) || math.IsNaN(lo) || math.IsNaN(hi) ||
		math.IsInf(lo, 0) || math.IsInf(hi, 0) {
		return 0
	}
	f := (value - lo) / (hi - lo)
	return max(0, min(1, f))
}

// gaugeHitTest returns true if (mx, my) is within the gauge arc
// ring (between innerR and outerR, within the arc angle).
func (gv *gaugeView) gaugeHitTest(mx, my float32) bool {
	dx := mx - gv.cx
	dy := my - gv.cy
	r2 := dx*dx + dy*dy
	if r2 > gv.outerR*gv.outerR || r2 < gv.innerR*gv.innerR {
		return false
	}
	start := gaugeStartAngle(gv.cfg.ArcAngle)
	a := normAngle(float32(math.Atan2(float64(dy), float64(dx))), start)
	return a <= start+gv.cfg.ArcAngle
}

func (gv *gaugeView) draw(dc *gui.DrawContext) {
	ctx := render.NewContext(dc)
	cfg := &gv.cfg
	th := cfg.Theme

	left := th.PaddingLeft
	right := ctx.Width() - th.PaddingRight
	top := th.PaddingTop
	bottom := ctx.Height() - th.PaddingBottom

	names := make([]string, len(cfg.Zones))
	for i, z := range cfg.Zones {
		names[i] = z.Label
	}
	right -= legendRightReserve(ctx, th, cfg.LegendPosition, names)
	top += legendTopReserve(ctx, th, cfg.LegendPosition, names, left, right)
	bottom -= legendBottomReserve(ctx, th, cfg.LegendPosition, names, left, right)

	if right <= left || bottom <= top {
		slog.Warn("plot area too small", "chart", cfg.ID)
		return
	}

	drawTitle(ctx, cfg.Title, th)

	plotW := right - left
	plotH := bottom - top
	outerR := min(plotW, plotH) / 2 * 0.85
	innerR := outerR * cfg.InnerRatio
	cx := (left + right) / 2
	cy := (top + bottom) / 2

	// Cache geometry for hit-testing in hover callback.
	gv.cx = cx
	gv.cy = cy
	gv.outerR = outerR
	gv.innerR = innerR

	startAngle := gaugeStartAngle(cfg.ArcAngle)

	// Background track arc.
	trackColor := gui.RGBA(128, 128, 128, 50)
	ctx.FilledArc(cx, cy, outerR, outerR,
		startAngle, cfg.ArcAngle, trackColor)

	// Zone arcs on the background track.
	if len(cfg.Zones) > 0 {
		prevFrac := float32(0)
		for i, z := range cfg.Zones {
			frac := float32(gaugeValueFraction(z.Threshold, cfg.Min, cfg.Max))
			sweep := (frac - prevFrac) * cfg.ArcAngle
			color := z.Color
			if !color.IsSet() {
				color = seriesColor(gui.Color{}, i, th.Palette)
			}
			// Draw zone at reduced alpha as background.
			zoneColor := gui.RGBA(color.R, color.G, color.B, 60)
			ctx.FilledArc(cx, cy, outerR, outerR,
				startAngle+prevFrac*cfg.ArcAngle, sweep, zoneColor)
			prevFrac = frac
		}
	}

	// Value arc.
	progress := animProgress(gv.win, cfg.ID)
	valFrac := float32(gaugeValueFraction(cfg.Value, cfg.Min, cfg.Max))
	valSweep := valFrac * cfg.ArcAngle * progress
	valColor := seriesColor(gui.Color{}, 0, th.Palette)
	// Color the value arc based on which zone the value falls in.
	for i, z := range cfg.Zones {
		if cfg.Value <= z.Threshold {
			valColor = z.Color
			if !valColor.IsSet() {
				valColor = seriesColor(gui.Color{}, i, th.Palette)
			}
			break
		}
	}
	if valSweep > 0 {
		ctx.FilledArc(cx, cy, outerR, outerR,
			startAngle, valSweep, valColor)
	}

	// Donut hole.
	if innerR > 0 {
		ctx.FilledCircle(cx, cy, innerR, th.Background)
	}

	// Pointer needle.
	if cfg.ShowPointer && !math.IsNaN(float64(valSweep)) &&
		!math.IsInf(float64(valSweep), 0) && outerR > 2 {
		needleAngle := float64(startAngle + valSweep)
		tipR := outerR - 2
		hubR := innerR * DefaultGaugePointerHubRatio
		hw := DefaultGaugePointerWidthPx

		cosA := float32(math.Cos(needleAngle))
		sinA := float32(math.Sin(needleAngle))

		// Tip point and two base points offset perpendicular.
		var pts [6]float32
		pts[0] = cx + tipR*cosA
		pts[1] = cy + tipR*sinA
		pts[2] = cx - hw*sinA
		pts[3] = cy + hw*cosA
		pts[4] = cx + hw*sinA
		pts[5] = cy - hw*cosA
		ctx.FilledPolygon(pts[:], valColor)

		ctx.FilledCircle(cx, cy, hubR, valColor)
	}

	// Min/max labels at arc endpoints.
	if cfg.ShowMinMax {
		style := th.TickStyle
		fh := ctx.FontHeight(style)
		labelR := outerR + 8

		// Min label at start angle.
		minLabel := fmt.Sprintf(cfg.ValueFormat, cfg.Min)
		mx := cx + labelR*float32(math.Cos(float64(startAngle)))
		my := cy + labelR*float32(math.Sin(float64(startAngle)))
		tw := ctx.TextWidth(minLabel, style)
		ctx.Text(mx-tw/2, my-fh/2, minLabel, style)

		// Max label at end angle.
		maxLabel := fmt.Sprintf(cfg.ValueFormat, cfg.Max)
		endAngle := startAngle + cfg.ArcAngle
		ex := cx + labelR*float32(math.Cos(float64(endAngle)))
		ey := cy + labelR*float32(math.Sin(float64(endAngle)))
		tw = ctx.TextWidth(maxLabel, style)
		ctx.Text(ex-tw/2, ey-fh/2, maxLabel, style)
	}

	// Centered value text.
	if cfg.ShowValue {
		style := th.TitleStyle
		fh := ctx.FontHeight(style)
		valText := fmt.Sprintf(cfg.ValueFormat, cfg.Value)
		tw := ctx.TextWidth(valText, style)
		ctx.Text(cx-tw/2, cy-fh/2, valText, style)
	}

	// Legend for zones.
	if len(cfg.Zones) > 0 {
		entries := make([]legendEntry, len(cfg.Zones))
		for i, z := range cfg.Zones {
			color := z.Color
			if !color.IsSet() {
				color = seriesColor(gui.Color{}, i, th.Palette)
			}
			entries[i] = legendEntry{Name: z.Label, Color: color, Index: i}
		}
		drawLegend(ctx, entries, th,
			plotRect{left, right, top, bottom},
			cfg.LegendPosition, nil)
	}

	// Tooltip on hover.
	if gv.hovering && gv.gaugeHitTest(gv.hoverPx, gv.hoverPy) {
		zone := ""
		for _, z := range cfg.Zones {
			if cfg.Value <= z.Threshold {
				zone = z.Label
				break
			}
		}
		label := fmt.Sprintf(cfg.ValueFormat, cfg.Value)
		if zone != "" {
			label = zone + ": " + label
		}
		drawTooltip(ctx, gv.hoverPx, gv.hoverPy, label, th,
			plotRect{left, right, top, bottom})
	}
}
