package axis

import (
	"fmt"
	"math"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-charts/scale"
)

// Log is a logarithmic axis.
type Log struct {
	title          string
	sc             *scale.Log
	tickFormat     TickFormat
	overrideDomain bool
}

// LogCfg configures a logarithmic axis.
type LogCfg struct {
	Title      string
	Min        float64
	Max        float64
	Base       float64
	TickFormat TickFormat
}

// NewLog creates a logarithmic axis. Base defaults to 10 via
// scale.NewLog when <= 0.
func NewLog(cfg LogCfg) *Log {
	return &Log{
		title:      cfg.Title,
		sc:         scale.NewLog(cfg.Min, cfg.Max, cfg.Base),
		tickFormat: cfg.TickFormat,
	}
}

// Label implements Axis.
func (a *Log) Label() string { return a.title }

// SetRange implements Axis.
func (a *Log) SetRange(min, max float64) { a.sc.SetDomain(min, max) }

// Domain implements Axis.
func (a *Log) Domain() (float64, float64) { return a.sc.Domain() }

// SetOverrideDomain implements Axis.
func (a *Log) SetOverrideDomain(v bool) { a.overrideDomain = v }

// Ticks implements Axis. Generates major ticks at powers of the
// base and minor ticks at integer multiples between powers.
func (a *Log) Ticks(pixelMin, pixelMax float32) []Tick {
	dMin, dMax := a.sc.Domain()
	base := a.sc.Base()
	if dMin <= 0 || dMax <= dMin || base <= 1 ||
		!fmath.Finite(dMin) || !fmath.Finite(dMax) {
		return nil
	}

	logBase := math.Log(base)
	if logBase == 0 || !fmath.Finite(logBase) {
		return nil
	}
	logMin := math.Floor(math.Log(dMin) / logBase)
	logMax := math.Ceil(math.Log(dMax) / logBase)
	if !fmath.Finite(logMin) || !fmath.Finite(logMax) {
		return nil
	}

	// Cap decades to prevent runaway.
	decades := logMax - logMin
	if decades > 100 {
		logMax = logMin + 100
		decades = 100
	}

	const maxTicks = 500
	// Cap intBase to avoid absurd minor tick counts from
	// fractional or very large bases.
	intBase := min(int(base), maxTicks)
	includeMinor := decades <= 10 && intBase >= 2

	// Estimate capacity.
	cap := int(decades) + 1
	if includeMinor {
		cap += int(decades) * (intBase - 2)
	}
	cap = min(cap, maxTicks)
	ticks := make([]Tick, 0, cap)

	for exp := int(logMin); exp <= int(logMax); exp++ {
		major := math.Pow(base, float64(exp))
		insideDomain := !a.overrideDomain ||
			(major >= dMin*(1-1e-9) && major <= dMax*(1+1e-9))

		if insideDomain {
			label := formatLogTickValue(major)
			if a.tickFormat != nil {
				label = a.tickFormat(major)
			}
			ticks = append(ticks, Tick{
				Value:    major,
				Label:    label,
				Position: a.Transform(major, pixelMin, pixelMax),
			})
			if len(ticks) >= maxTicks {
				return ticks
			}
		}

		// Minor ticks between this power and the next.
		// Skip when the major tick was outside the domain.
		if includeMinor && insideDomain && exp < int(logMax) {
			for m := 2; m < intBase; m++ {
				v := float64(m) * major
				if v < dMin || v > dMax {
					continue
				}
				ticks = append(ticks, Tick{
					Value:    v,
					Position: a.Transform(v, pixelMin, pixelMax),
					Minor:    true,
				})
				if len(ticks) >= maxTicks {
					return ticks
				}
			}
		}
	}
	return ticks
}

// formatLogTickValue formats a log axis major tick value using
// compact notation (e.g., "1K", "10M", "1e15").
func formatLogTickValue(v float64) string {
	switch {
	case v < 1e-9:
		return fmt.Sprintf("%.0e", v)
	case v < 1:
		return fmt.Sprintf("%g", v)
	case v < 1e3:
		return formatTickValue(v)
	case v < 1e6:
		return formatSI(v, 1e3, "K")
	case v < 1e9:
		return formatSI(v, 1e6, "M")
	case v < 1e12:
		return formatSI(v, 1e9, "B")
	case v < 1e15:
		return formatSI(v, 1e12, "T")
	default:
		return fmt.Sprintf("%.0e", v)
	}
}

func formatSI(v, divisor float64, suffix string) string {
	d := v / divisor
	if d == math.Trunc(d) {
		return fmt.Sprintf("%d%s", int64(d), suffix)
	}
	return fmt.Sprintf("%g%s", d, suffix)
}

// Transform implements Axis.
func (a *Log) Transform(value float64, pixelMin, pixelMax float32) float32 {
	return a.sc.Transform(value, pixelMin, pixelMax)
}

// Invert implements Axis.
func (a *Log) Invert(pixel, pixelMin, pixelMax float32) float64 {
	return a.sc.Invert(pixel, pixelMin, pixelMax)
}
