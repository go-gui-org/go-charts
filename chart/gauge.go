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

// GaugeValueAnchor selects the base point for the value text. The
// zero value keeps the historical placement, which drops the text
// below the centre so the needle and its hub do not cross it.
type GaugeValueAnchor uint8

const (
	// GaugeValueDefault drops the text below the centre by
	// DefaultGaugeValueDropRatio of the hole radius.
	GaugeValueDefault GaugeValueAnchor = iota

	// GaugeValueCentre puts the text on the centre. Correct when
	// no needle is drawn.
	GaugeValueCentre

	// GaugeValueAboveCentre lifts the text by the same fraction
	// the default drops it.
	GaugeValueAboveCentre

	// GaugeValueBelowArc puts the text under the whole dial, clear
	// of the arc and the hole.
	GaugeValueBelowArc
)

// maxGaugeValueOffset bounds ValueOffsetRatio, which keeps text with
// an absurd offset near the dial instead of far outside the clip box.
const maxGaugeValueOffset float32 = 10

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

	// GradientZones blends the Zones colors into one continuous
	// sweep along the arc instead of stepping at each threshold.
	// Each zone color is anchored at the middle of its own span, so
	// a zone still reads as its own color at its centre. Ignored
	// when Zones is empty or ArcGradient is set. The thresholds
	// still drive hit-testing, the tooltip and the legend; only the
	// fill changes.
	GradientZones bool

	// ArcGradient is an explicit ramp for callers not using Zones.
	// Stop positions are normalized over Min..Max. Takes precedence
	// over GradientZones. At most 256 stops; more are dropped.
	ArcGradient []gui.GradientStop

	// InnerRatio is the inner radius as a fraction of the outer
	// radius (0–1). Default: 0.7. Controls arc thickness.
	InnerRatio float32

	// RadiusRatio is the outer radius as a fraction of half the
	// smaller plot side (0–1]. Default: 0.85, which leaves room
	// for graduation labels inside the plot box. Raise it towards
	// 1 when the theme padding already reserves that room, or the
	// allowance is paid twice and the dial is needlessly small.
	RadiusRatio float32

	// ShowValue renders the current value as centered text.
	ShowValue bool

	// ValueAnchor selects where the value text sits vertically.
	// Default: GaugeValueDefault, dropped below the centre so the
	// needle does not cross it.
	ValueAnchor GaugeValueAnchor

	// ValueOffsetRatio shifts the value text from its anchor, as a
	// signed fraction of the hole radius (the outer radius when
	// there is no hole). Positive moves down. Zero adds nothing to
	// the anchor. Range: [-10, 10].
	ValueOffsetRatio float32

	// ValueLabel is an optional second line drawn under the value
	// in TickStyle, usually the unit. Empty draws nothing.
	ValueLabel string

	// ShowMinMax renders min/max labels at the arc endpoints.
	// Redundant when TickCount is set with ShowTickLabels: the
	// first and last graduation already carry those numbers.
	ShowMinMax bool

	// TickCount is how many intervals the arc is divided into by
	// major graduations, so it draws TickCount+1 marks including
	// both endpoints. Zero draws none.
	TickCount int

	// MinorTicks is how many unlabelled graduations fall between
	// two major ones. Zero draws none. Ignored when TickCount
	// is zero.
	MinorTicks int

	// ShowTickLabels renders the value of each major graduation
	// beyond it, formatted with ValueFormat.
	ShowTickLabels bool

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
	// stops is the arc ramp, nil when the gauge is not
	// gradient-filled. Built once: cfg never changes after
	// construction, so rebuilding it per frame would only repeat
	// the same allocation and sort on every redraw.
	stops []gui.GradientStop
	// Cached geometry for cursor hit-testing.
	cx, cy, outerR, innerR float32
	win                    *gui.Window
}

// newGaugeView builds the view and everything derived from cfg that
// the draw path would otherwise recompute each frame. cfg must have
// its defaults applied already.
func newGaugeView(cfg GaugeCfg) *gaugeView {
	return &gaugeView{
		cfg:   cfg,
		stops: gaugeGradientStops(&cfg, cfg.Theme),
	}
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
	return newGaugeView(cfg)
}

func (cfg *GaugeCfg) applyGaugeDefaults() {
	cfg.applyDefaults()
	if cfg.Min == 0 && cfg.Max == 0 {
		cfg.Max = 100
	}
	cfg.ArcAngle = cmp.Or(cfg.ArcAngle, DefaultGaugeArcAngle)
	cfg.InnerRatio = cmp.Or(cfg.InnerRatio, DefaultGaugeInnerRatio)
	cfg.RadiusRatio = cmp.Or(cfg.RadiusRatio, DefaultGaugeRadiusRatio)
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
	if cfg.RadiusRatio < 0 || cfg.RadiusRatio > 1 {
		errs = append(errs, "RadiusRatio out of [0, 1]")
	}
	if cfg.TickCount < 0 {
		errs = append(errs, "TickCount negative")
	}
	if cfg.MinorTicks < 0 {
		errs = append(errs, "MinorTicks negative")
	}
	off := float64(cfg.ValueOffsetRatio)
	switch {
	case math.IsNaN(off) || math.IsInf(off, 0):
		errs = append(errs, "ValueOffsetRatio not finite")
	case cfg.ValueOffsetRatio < -maxGaugeValueOffset ||
		cfg.ValueOffsetRatio > maxGaugeValueOffset:
		errs = append(errs, "ValueOffsetRatio out of [-10, 10]")
	}
	if cfg.ValueAnchor > GaugeValueBelowArc {
		errs = append(errs, "ValueAnchor unknown")
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
	if len(cfg.ArcGradient) > maxArcGradientStops {
		errs = append(errs, fmt.Sprintf("ArcGradient has %d stops > %d",
			len(cfg.ArcGradient), maxArcGradientStops))
	}
	for i, s := range cfg.ArcGradient {
		p := float64(s.Pos)
		switch {
		case math.IsNaN(p) || math.IsInf(p, 0):
			errs = append(errs, fmt.Sprintf("ArcGradient stop %d Pos not finite", i))
		case s.Pos < 0 || s.Pos > 1:
			errs = append(errs, fmt.Sprintf(
				"ArcGradient stop %d Pos %.4g out of [0, 1]", i, s.Pos))
		}
	}
	return buildError("gauge", errs)
}

// Draw renders the chart onto dc for headless export.
func (gv *gaugeView) Draw(dc *gui.DrawContext) { gv.draw(dc) }

func (gv *gaugeView) chartTheme() *theme.Theme { return gv.cfg.Theme }

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

// maxGaugeSteps caps total graduation marks to bound draw work.
const maxGaugeSteps = 500

// drawTicks draws the graduations around the outside of the arc: a
// major mark every arc/TickCount, MinorTicks unlabelled marks between
// each pair, and optionally the value of each major mark beyond it.
//
// The marks sit outside the arc rather than over it, so a zone colour
// is never broken up by a line drawn across it. That means they eat
// into the theme's padding: allow roughly the tick length plus a line
// of text on every side the arc reaches.
func (gv *gaugeView) drawTicks(ctx *render.Context,
	cx, cy, outerR, startAngle float32) {

	cfg := &gv.cfg
	th := cfg.Theme

	majorR := outerR + DefaultGaugeTickGapPx
	labelR := majorR + DefaultGaugeMajorTickPx + DefaultGaugeTickGapPx
	style := th.TickStyle
	fh := ctx.FontHeight(style)

	// Total marks, counting both endpoints and the minor marks in
	// between, so one loop walks the whole arc at a constant step.
	// Cap to avoid DoS from large user-supplied TickCount/MinorTicks
	// and to guard the multiplication against overflow.
	tickCount := min(cfg.TickCount, maxGaugeSteps)
	minorTicks := min(cfg.MinorTicks, maxGaugeSteps)
	steps := min(tickCount*(minorTicks+1), maxGaugeSteps)
	// Recompute effective minorTicks so major detection stays
	// consistent when steps was clamped.
	if tickCount > 0 {
		minorTicks = max(steps/tickCount-1, 0)
	}
	span := cfg.Max - cfg.Min

	for i := 0; i <= steps; i++ {
		frac := float32(i) / float32(steps)
		angle := float64(startAngle + frac*cfg.ArcAngle)
		cosA := float32(math.Cos(angle))
		sinA := float32(math.Sin(angle))

		major := i%(minorTicks+1) == 0
		length := DefaultGaugeMinorTickPx
		if major {
			length = DefaultGaugeMajorTickPx
		}

		ctx.Line(
			cx+majorR*cosA, cy+majorR*sinA,
			cx+(majorR+length)*cosA, cy+(majorR+length)*sinA,
			th.AxisColor, th.AxisWidth)

		if !major || !cfg.ShowTickLabels {
			continue
		}

		// The label is centred on the tick's own angle, which keeps
		// it clear of its neighbours as long as the caller does not
		// ask for more graduations than the dial has room for.
		val := cfg.Min + float64(frac)*span
		text := fmt.Sprintf(cfg.ValueFormat, val)
		tw := ctx.TextWidth(text, style)
		lx, ly := labelCentre(cx, cy, labelR, cosA, sinA, tw, fh)
		ctx.Text(lx-tw/2, ly-fh/2, text, style)
	}
}

// labelCentre places a label outside a circle of radius r so the text
// box clears it, and returns the centre point to draw around.
//
// Centring the box on the radius is not enough: at the left and right of
// the dial the label runs horizontally, so half its width reaches back
// over the arc, and a wide number touches the ring. Pushing the box out
// by its own half-extent along each axis — scaled by the direction
// cosine, so a label at the top pays with its height and one at the side
// pays with its width — puts the near edge on the radius instead of the
// centre.
func labelCentre(cx, cy, r, cosA, sinA, tw, fh float32) (x, y float32) {
	x = cx + r*cosA + tw/2*cosA
	y = cy + r*sinA + fh/2*sinA
	return x, y
}

func (gv *gaugeView) internalHover(ctx gui.EventCtx) {
	ctx.Consume()
	gv.hoverPx = ctx.Event.MouseX - ctx.Layout.Shape.X
	gv.hoverPy = ctx.Event.MouseY - ctx.Layout.Shape.Y
	gv.hovering = true
	saveHover(ctx.Window, ctx.Layout, gv.cfg.ID, true, gv.hoverPx, gv.hoverPy)
	if gv.outerR > 0 && gv.gaugeHitTest(gv.hoverPx, gv.hoverPy) {
		ctx.Window.SetMouseCursorPointingHand()
	} else {
		ctx.Window.SetMouseCursorArrow()
	}
	if gv.cfg.OnHover != nil {
		gv.cfg.OnHover(ctx)
	}
}

func (gv *gaugeView) internalMouseLeave(ctx gui.EventCtx) {
	ctx.Consume()
	gv.hovering = false
	saveHover(ctx.Window, ctx.Layout, gv.cfg.ID, false, 0, 0)
	ctx.Window.SetMouseCursorArrow()
	if gv.cfg.OnMouseLeave != nil {
		gv.cfg.OnMouseLeave(ctx)
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

// gaugeZoneColor returns the color of the zone the value falls in,
// falling back to the palette for an unset zone color and to the
// first palette entry when there are no zones at all.
func gaugeZoneColor(cfg *GaugeCfg, th *theme.Theme) gui.Color {
	for i, z := range cfg.Zones {
		if cfg.Value > z.Threshold {
			continue
		}
		if z.Color.IsSet() {
			return z.Color
		}
		return seriesColor(gui.Color{}, i, th.Palette)
	}
	return seriesColor(gui.Color{}, 0, th.Palette)
}

// drawTrack draws the unlit dial under the value arc: a grey ring,
// then either the zone tints as flat steps or the whole ramp as one
// sweep. stops is nil when the gauge is not gradient-filled.
func (gv *gaugeView) drawTrack(ctx *render.Context, cx, cy, outerR,
	startAngle float32, stops []gui.GradientStop) {

	cfg := &gv.cfg
	th := cfg.Theme

	if stops != nil {
		// No grey ring: every segment of the sweep is opaque and
		// spans the whole arc, so a ring beneath it would be
		// completely covered.
		drawGradientArc(ctx, cx, cy, outerR,
			startAngle, cfg.ArcAngle, stops, 0, 1,
			gaugeTrackAlpha, th.Background)
		return
	}

	trackColor := gui.RGBA(128, 128, 128, 50)
	ctx.FilledArc(cx, cy, outerR, outerR,
		startAngle, cfg.ArcAngle, trackColor)

	// Zone arcs on the background track.
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
	outerR := min(plotW, plotH) / 2 * cfg.RadiusRatio
	innerR := outerR * cfg.InnerRatio
	cx := (left + right) / 2
	cy := (top + bottom) / 2

	// Cache geometry for hit-testing in hover callback.
	gv.cx = cx
	gv.cy = cy
	gv.outerR = outerR
	gv.innerR = innerR

	startAngle := gaugeStartAngle(cfg.ArcAngle)

	// The ramp is used at two strengths: a tint for the background
	// track, full strength for the value arc.
	stops := gv.stops

	gv.drawTrack(ctx, cx, cy, outerR, startAngle, stops)

	// Value arc.
	progress := animProgress(gv.win, cfg.ID)
	valFrac := float32(gaugeValueFraction(cfg.Value, cfg.Min, cfg.Max))
	valSweep := valFrac * cfg.ArcAngle * progress

	// The lit fraction, which trails valFrac while the entry
	// animation runs. Using it rather than valFrac keeps every
	// segment on the same ramp position as the track beneath it.
	litFrac := valFrac * progress

	// The needle and hub take the color of the arc under the tip:
	// the ramp position when swept, the containing zone when not.
	var valColor gui.Color
	if stops != nil {
		valColor = gaugeSampleStops(stops, litFrac)
	} else {
		valColor = gaugeZoneColor(cfg, th)
	}
	switch {
	case valSweep <= 0:
		// Nothing lit yet.
	case stops != nil:
		// Only the swept part of the ramp, so the lit arc is the same
		// colors the track shows, at full strength.
		drawGradientArc(ctx, cx, cy, outerR,
			startAngle, valSweep, stops, 0, litFrac, 1, th.Background)
	default:
		ctx.FilledArc(cx, cy, outerR, outerR,
			startAngle, valSweep, valColor)
	}

	// Donut hole.
	if innerR > 0 {
		ctx.FilledCircle(cx, cy, innerR, th.Background)
	}

	// Graduations, drawn before the needle so the needle stays on
	// top of them.
	if cfg.TickCount > 0 {
		gv.drawTicks(ctx, cx, cy, outerR, startAngle)
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
		tw := ctx.TextWidth(minLabel, style)
		mx, my := labelCentre(cx, cy, labelR,
			float32(math.Cos(float64(startAngle))),
			float32(math.Sin(float64(startAngle))), tw, fh)
		ctx.Text(mx-tw/2, my-fh/2, minLabel, style)

		// Max label at end angle.
		maxLabel := fmt.Sprintf(cfg.ValueFormat, cfg.Max)
		endAngle := startAngle + cfg.ArcAngle
		tw = ctx.TextWidth(maxLabel, style)
		ex, ey := labelCentre(cx, cy, labelR,
			float32(math.Cos(float64(endAngle))),
			float32(math.Sin(float64(endAngle))), tw, fh)
		ctx.Text(ex-tw/2, ey-fh/2, maxLabel, style)
	}

	if cfg.ShowValue {
		gv.drawValueText(ctx, th, cx, cy, innerR, outerR)
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

// gaugeValueBaseY returns the vertical anchor point for the value
// text. radiusRef is the hole radius, or the outer radius when the
// gauge has no hole, so an anchor holds its place at any dial size.
func gaugeValueBaseY(anchor GaugeValueAnchor, cy, innerR, outerR float32) float32 {
	radiusRef := innerR
	if radiusRef == 0 {
		radiusRef = outerR
	}
	switch anchor {
	case GaugeValueCentre:
		return cy
	case GaugeValueAboveCentre:
		return cy - radiusRef*DefaultGaugeValueDropRatio
	case GaugeValueBelowArc:
		return cy + outerR + DefaultGaugeValueBelowArcGapPx
	default:
		// GaugeValueDefault, and any unknown value: the needle
		// pivots on the centre, so text placed there is drawn over
		// by the hub and struck through by the needle.
		return cy + radiusRef*DefaultGaugeValueDropRatio
	}
}

// gaugeValueOffsetPx converts ValueOffsetRatio into pixels. Validate
// only warns about a bad ratio, so the draw path must not trust it:
// a non-finite ratio contributes nothing and a wild one is clamped.
func gaugeValueOffsetPx(ratio, innerR, outerR float32) float32 {
	r := float64(ratio)
	if math.IsNaN(r) || math.IsInf(r, 0) {
		return 0
	}
	ratio = min(max(ratio, -maxGaugeValueOffset), maxGaugeValueOffset)
	radiusRef := innerR
	if radiusRef == 0 {
		radiusRef = outerR
	}
	return ratio * radiusRef
}

// drawValueText draws the value, and the unit line under it when
// ValueLabel is set. The two lines are treated as one block centred
// on the anchor, so adding a unit does not move the number off it.
func (gv *gaugeView) drawValueText(ctx *render.Context, th *theme.Theme,
	cx, cy, innerR, outerR float32) {
	cfg := &gv.cfg
	style := th.TitleStyle
	fh := ctx.FontHeight(style)
	valText := fmt.Sprintf(cfg.ValueFormat, cfg.Value)

	// Height of the whole block, so the anchor centres value plus
	// unit rather than the value alone.
	blockH := fh
	if cfg.ValueLabel != "" {
		blockH += DefaultGaugeValueLabelGapPx + ctx.FontHeight(th.TickStyle)
	}

	y := gaugeValueBaseY(cfg.ValueAnchor, cy, innerR, outerR) +
		gaugeValueOffsetPx(cfg.ValueOffsetRatio, innerR, outerR)
	top := y - blockH/2

	tw := ctx.TextWidth(valText, style)
	ctx.Text(cx-tw/2, top, valText, style)

	if cfg.ValueLabel != "" {
		lw := ctx.TextWidth(cfg.ValueLabel, th.TickStyle)
		ctx.Text(cx-lw/2, top+fh+DefaultGaugeValueLabelGapPx,
			cfg.ValueLabel, th.TickStyle)
	}
}
