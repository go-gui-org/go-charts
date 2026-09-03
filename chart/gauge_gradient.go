package chart

import (
	"cmp"
	"math"
	"slices"

	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/theme"
	"github.com/go-gui-org/go-gui/gui"
)

// maxArcGradientStops bounds ArcGradient so a hostile or generated
// config cannot make stop sorting and sampling arbitrarily expensive.
// 256 stops already exceed any fidelity a dial can show.
const maxArcGradientStops = 256

const (
	// gaugeGradientSegmentPx is how many pixels of outer-edge arc one
	// flat segment covers. Three pixels puts the color step below
	// what the eye resolves against an antialiased edge.
	gaugeGradientSegmentPx float32 = 3

	// gaugeMinGradientSegments keeps a tiny dial from collapsing to
	// one or two visible bands.
	gaugeMinGradientSegments = 8

	// gaugeMaxGradientSegments caps draw calls. Over a 270° arc this
	// is 1.5° a segment, where banding is already invisible, so a
	// large dial gains nothing from more.
	gaugeMaxGradientSegments = 180

	// gaugeGradientOverlapPx is how far each segment runs past its
	// own end, in pixels of outer-edge arc. It hides the antialiased
	// radial seam between neighbours. Measured at the outer edge, so
	// the inner edge gets less; 1.5 covers both.
	gaugeGradientOverlapPx float32 = 1.5

	// gaugeTrackAlpha is the strength of the background sweep. It
	// matches the alpha the stepped zone arcs already use, so the
	// two fills read at the same weight.
	gaugeTrackAlpha float32 = 60.0 / 255
)

// gaugeGradientStops builds the ramp the arc is filled with, or
// returns nil when the gauge is not gradient-filled.
//
// ArcGradient wins over GradientZones: an explicit ramp is a more
// specific statement of intent than "blend what the zones already
// say". Validate only warns about bad input, so everything here is
// re-checked — non-finite positions are dropped, positions clamped,
// the count truncated, and the result sorted.
func gaugeGradientStops(cfg *GaugeCfg, th *theme.Theme) []gui.GradientStop {
	if len(cfg.ArcGradient) > 0 {
		return sanitizeStops(cfg.ArcGradient)
	}
	if !cfg.GradientZones || len(cfg.Zones) < 2 {
		// One zone has nothing to blend towards, and no zones has
		// nothing to blend at all.
		return nil
	}

	// Anchor each zone color at the middle of its own span, so the
	// zone still reads as its own color at its centre and the blend
	// happens across the thresholds. Anchoring at the threshold
	// instead would leave Min..firstThreshold unramped and put the
	// last zone's color only at Max.
	stops := make([]gui.GradientStop, 0, len(cfg.Zones))
	prevFrac := float32(0)
	for i, z := range cfg.Zones {
		frac := float32(gaugeValueFraction(z.Threshold, cfg.Min, cfg.Max))
		color := z.Color
		if !color.IsSet() {
			color = seriesColor(gui.Color{}, i, th.Palette)
		}
		stops = append(stops, gui.GradientStop{
			Color: color,
			Pos:   (prevFrac + frac) / 2,
		})
		prevFrac = frac
	}

	// Thresholds out of order are only a warning, so the midpoints
	// are sorted through the same path an explicit ramp takes.
	stops = sanitizeStops(stops)
	if len(stops) < 2 {
		return nil
	}

	// Pin the ends, after sorting so the first and last really are
	// the extremes. Without this the arc is flat from 0 to the
	// first midpoint and from the last midpoint to 1.
	stops[0].Pos = 0
	stops[len(stops)-1].Pos = 1
	return stops
}

// sanitizeStops copies stops with non-finite positions dropped,
// positions clamped to [0, 1], the count truncated, and the result
// sorted ascending. Returns nil when nothing usable survives.
func sanitizeStops(in []gui.GradientStop) []gui.GradientStop {
	if len(in) > maxArcGradientStops {
		in = in[:maxArcGradientStops]
	}
	out := make([]gui.GradientStop, 0, len(in))
	for _, s := range in {
		p := float64(s.Pos)
		if math.IsNaN(p) || math.IsInf(p, 0) {
			continue
		}
		out = append(out, gui.GradientStop{
			Color: s.Color,
			Pos:   min(max(s.Pos, 0), 1),
		})
	}
	if len(out) == 0 {
		return nil
	}
	slices.SortStableFunc(out, func(a, b gui.GradientStop) int {
		return cmp.Compare(a.Pos, b.Pos)
	})
	return out
}

// gaugeSampleStops returns the ramp color at pos, interpolating in
// Oklab so a red-to-green sweep does not pass through mud. Stops are
// assumed sanitized: finite, sorted, and non-empty. pos is clamped.
func gaugeSampleStops(stops []gui.GradientStop, pos float32) gui.Color {
	if len(stops) == 0 {
		return gui.Color{}
	}
	pos = min(max(pos, 0), 1)
	if pos <= stops[0].Pos {
		return stops[0].Color
	}
	for i := 1; i < len(stops); i++ {
		right := stops[i]
		if pos > right.Pos {
			continue
		}
		left := stops[i-1]
		span := right.Pos - left.Pos
		if span <= 0 {
			// Coincident stops are a hard step by definition.
			return right.Color
		}
		t := float64((pos - left.Pos) / span)
		return theme.LerpOklab(left.Color, right.Color, t)
	}
	return stops[len(stops)-1].Color
}

// isFinite32 reports whether v is a real number. Geometry reaching
// the gradient path comes from config the caller controls and
// Validate only warns about, so it is checked rather than trusted.
func isFinite32(v float32) bool {
	f := float64(v)
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

// gaugeSegments returns how many flat arcs to draw across sweep. The
// count follows the arc length in pixels, not a fixed number, so a
// small dial does not pay for 180 draw calls.
func gaugeSegments(outerR, sweep float32) int {
	// The product, not the operands: two finite values can still
	// multiply to an infinity in float32.
	arcLen := outerR * sweep
	if !isFinite32(arcLen) {
		return gaugeMinGradientSegments
	}
	n := int(math.Round(math.Abs(float64(arcLen)) / float64(gaugeGradientSegmentPx)))
	return min(max(n, gaugeMinGradientSegments), gaugeMaxGradientSegments)
}

// flattenOnto composites c over bg at the given strength and returns
// an opaque color. The segments must be opaque: see drawGradientArc.
// alpha multiplies c's own alpha, so a caller's translucent stop
// still reads as translucent against the chart background.
func flattenOnto(bg, c gui.Color, alpha float32) gui.Color {
	t := float64(c.A) / 255 * float64(alpha)
	out := theme.Lerp(bg, c, t)
	return gui.RGBA(out.R, out.G, out.B, 255)
}

// drawGradientArc fills sweep radians starting at start with the ramp,
// as a run of flat arcs each sampled at its own midpoint angle. The
// span of the ramp used is fracLo..fracHi, which lets the value arc
// draw the first part of the same ramp the background track shows in
// full. alpha is the strength of the fill against bg, so the track can
// be a tint of the same colors the value arc uses at full strength.
//
// Every segment is composited against bg to an opaque color rather
// than drawn translucent. FilledArc antialiases its radial edges, so
// neighbouring segments cannot be joined cleanly while translucent:
// abutting them leaves a hairline gap where both edges fade out, and
// overlapping them paints the shared strip twice, which shows as a
// darker stripe. Opaque segments make overdraw idempotent, so they
// can overlap and the seam disappears. The overlap is measured at the
// outer edge, where the segments are widest; nearer the hole the same
// angle covers fewer pixels, which is why it is more than one.
func drawGradientArc(ctx *render.Context, cx, cy, r, start, sweep float32,
	stops []gui.GradientStop, fracLo, fracHi, alpha float32, bg gui.Color) {

	if len(stops) == 0 || r <= 0 || sweep == 0 {
		return
	}
	// NaN passes every comparison above, so both operands are
	// checked explicitly. A non-finite radius would otherwise reach
	// the overlap division and poison every segment angle.
	if !isFinite32(sweep) || !isFinite32(r) {
		return
	}

	n := gaugeSegments(r, sweep)
	step := sweep / float32(n)
	overlap := min(gaugeGradientOverlapPx/r, step)

	for i := range n {
		segStart := start + float32(i)*step
		segSweep := step
		if i < n-1 {
			segSweep += overlap
		}
		// Sample at the segment's own midpoint, not its edge, so the
		// run of flat fills averages to the true ramp.
		t := (float32(i) + 0.5) / float32(n)
		c := gaugeSampleStops(stops, fracLo+(fracHi-fracLo)*t)
		ctx.FilledArc(cx, cy, r, r, segStart, segSweep,
			flattenOnto(bg, c, alpha))
	}
}
