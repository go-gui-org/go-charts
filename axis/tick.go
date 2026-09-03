package axis

import (
	"math"

	"github.com/go-gui-org/go-charts/internal/fmath"
)

// NiceNumber computes a "nice" number approximately equal to the
// given value. If round is true, it rounds; otherwise it takes
// the ceiling.
func NiceNumber(value float64, round bool) float64 {
	if value == 0 {
		return 0
	}
	sign := 1.0
	if value < 0 {
		sign = -1
		value = -value
	}
	exp := math.Floor(math.Log10(value))
	frac := value / math.Pow(10, exp)

	var nice float64
	if round {
		switch {
		case frac < 1.5:
			nice = 1
		case frac < 3:
			nice = 2
		case frac < 7:
			nice = 5
		default:
			nice = 10
		}
	} else {
		switch {
		case frac <= 1:
			nice = 1
		case frac <= 2:
			nice = 2
		case frac <= 5:
			nice = 5
		default:
			nice = 10
		}
	}
	return sign * nice * math.Pow(10, exp)
}

// GenerateNiceTicks generates evenly-spaced tick values for the
// given data range, targeting approximately maxTicks ticks.
// Non-finite or degenerate inputs produce a safe fallback.
func GenerateNiceTicks(dataMin, dataMax float64, maxTicks int) []float64 {
	if maxTicks < 2 {
		maxTicks = 2
	}

	// Guard: non-finite inputs.
	if !fmath.Finite(dataMin) || !fmath.Finite(dataMax) {
		if fmath.Finite(dataMin) {
			return []float64{dataMin}
		}
		if fmath.Finite(dataMax) {
			return []float64{dataMax}
		}
		return nil
	}

	rangeVal := NiceNumber(dataMax-dataMin, false)
	if !fmath.Finite(rangeVal) {
		return []float64{dataMin}
	}

	spacing := NiceNumber(rangeVal/float64(maxTicks-1), true)
	if spacing <= 0 || !fmath.Finite(spacing) {
		return []float64{dataMin}
	}

	niceMin := math.Floor(dataMin/spacing) * spacing
	niceMax := math.Ceil(dataMax/spacing) * spacing
	if !fmath.Finite(niceMin) || !fmath.Finite(niceMax) {
		return []float64{dataMin}
	}

	const maxTickCount = 500
	n := min(int(math.Round((niceMax-niceMin)/spacing))+1, maxTickCount)

	// Compute decimal precision from spacing to snap tick values.
	// E.g. spacing=0.2 → 1 decimal place, spacing=0.05 → 2.
	// Uses log10 instead of iterative multiplication to avoid
	// float64 representation issues (e.g. 0.3).
	prec := 0
	if spacing > 0 && spacing < 1 {
		prec = max(0, int(-math.Floor(math.Log10(spacing))))
	}
	factor := math.Pow(10, float64(prec))

	ticks := make([]float64, 0, n)
	for i := range n {
		v := niceMin + float64(i)*spacing
		v = math.Round(v*factor) / factor
		ticks = append(ticks, v)
	}
	return ticks
}

// maxExactTicks caps the number of ticks to avoid excessive
// allocations and draw work from user-supplied counts.
const maxExactTicks = 500

// GenerateExactTicks generates exactly count evenly-spaced tick values
// covering [dataMin, dataMax] with a round spacing.
//
// GenerateNiceTicks treats its count as a target: it picks the spacing
// closest to the ideal and returns however many ticks that spacing
// yields, so an axis whose data grows flips between three and four
// labels as the range crosses a rounding boundary. This picks the
// smallest round spacing whose rounded-out range fits in count-1
// intervals, then raises the top by whole steps until the count is
// met. The slack lands above the data, never below, so a zero floor
// stays at zero.
//
// Degenerate or non-finite input falls back to GenerateNiceTicks.
func GenerateExactTicks(dataMin, dataMax float64, count int) []float64 {
	if count < 2 {
		count = 2
	}
	if count > maxExactTicks {
		count = maxExactTicks
	}
	if !fmath.Finite(dataMin) || !fmath.Finite(dataMax) {
		return GenerateNiceTicks(dataMin, dataMax, count)
	}
	if dataMax < dataMin {
		dataMin, dataMax = dataMax, dataMin
	}

	// A flat or empty range has no spacing of its own. Size one off the
	// value itself so a constant series still gets a readable scale.
	span := dataMax - dataMin
	if span <= 0 {
		span = math.Abs(dataMax)
		if span == 0 {
			span = 1
		}
	}

	// Start at the smallest round step that could span the data in
	// count-1 intervals, then climb the 1-2-5 ladder until snapping the
	// ends outwards to multiples of the step no longer needs more ticks
	// than asked for.
	step := NiceNumber(span/float64(count-1), false)
	if step <= 0 || !fmath.Finite(step) {
		return GenerateNiceTicks(dataMin, dataMax, count)
	}
	var lo float64
	for range 64 {
		lo = math.Floor(dataMin/step) * step
		hi := math.Ceil(dataMax/step) * step
		if !fmath.Finite(lo) || !fmath.Finite(hi) {
			return GenerateNiceTicks(dataMin, dataMax, count)
		}
		if int(math.Round((hi-lo)/step))+1 <= count {
			break
		}
		step = nextNiceStep(step)
		if step <= 0 || !fmath.Finite(step) {
			return GenerateNiceTicks(dataMin, dataMax, count)
		}
	}

	// Snap values to the step's decimal precision, as GenerateNiceTicks
	// does: 3*0.1 is not 0.3 in binary floating point, and the label
	// would show it.
	prec := 0
	if step < 1 {
		prec = max(0, int(-math.Floor(math.Log10(step))))
	}
	factor := math.Pow(10, float64(prec))

	ticks := make([]float64, 0, count)
	for i := range count {
		v := lo + float64(i)*step
		ticks = append(ticks, math.Round(v*factor)/factor)
	}
	return ticks
}

// nextNiceStep is the next step up the 1-2-5 ladder: 1→2, 2→5, 5→10.
func nextNiceStep(step float64) float64 {
	exp := math.Floor(math.Log10(step))
	pow := math.Pow(10, exp)
	switch frac := math.Round(step / pow); {
	case frac < 2:
		return 2 * pow
	case frac < 5:
		return 5 * pow
	default:
		return 10 * pow
	}
}
