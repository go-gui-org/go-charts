package chart

import (
	"log/slog"
	"math"

	"github.com/go-gui-org/go-charts/axis"
)

// autoLinearAxis returns a configured Axis for the given data range.
// padFrac adds fractional padding to both sides when creating an auto-ranged
// axis; use 0 for no padding. If cfgAxis is non-nil it is returned with its
// range updated when hasBounds is true; when hasBounds is false the axis is
// returned unchanged (the caller-supplied axis owns its own domain in that
// case). Returns (nil, false) when no config axis was provided and bounds are
// unusable.
//
// Callers must check non-finite bounds before calling.
func autoLinearAxis(
	cfgAxis axis.Axis,
	minV, maxV, padFrac float64,
	id string,
) (axis.Axis, bool) {
	hasBounds := minV <= maxV
	if cfgAxis != nil {
		if hasBounds {
			cfgAxis.SetRange(minV, maxV)
		}
		return cfgAxis, true
	}
	if !hasBounds {
		slog.Warn("all series empty", "chart", id)
		return nil, false
	}
	r := maxV - minV
	if r == 0 || math.IsInf(r, 0) {
		r = 1
	}
	ax := axis.NewLinear(axis.LinearCfg{AutoRange: true})
	ax.SetRange(minV-r*padFrac, maxV+r*padFrac)
	return ax, true
}
