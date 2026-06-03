package axis

import (
	"fmt"

	"github.com/go-gui-org/go-charts/scale"
)

// Linear is a linear numeric axis with auto-tick generation.
type Linear struct {
	title          string
	sc             *scale.Linear
	autoRange      bool
	overrideDomain bool
	tickFormat     TickFormat
}

// LinearCfg configures a linear axis.
type LinearCfg struct {
	Title      string
	Min        float64
	Max        float64
	AutoRange  bool
	TickFormat TickFormat
}

// NewLinear creates a linear axis.
func NewLinear(cfg LinearCfg) *Linear {
	return &Linear{
		title:      cfg.Title,
		sc:         scale.NewLinear(cfg.Min, cfg.Max),
		autoRange:  cfg.AutoRange,
		tickFormat: cfg.TickFormat,
	}
}

// SetRange updates the axis data range.
func (a *Linear) SetRange(min, max float64) {
	a.sc.SetDomain(min, max)
}

// Domain returns the current data range.
func (a *Linear) Domain() (float64, float64) {
	return a.sc.Domain()
}

// SetOverrideDomain controls whether Ticks() skips auto-range
// expansion. When true, the domain set by SetRange is preserved
// exactly — useful for zoomed views.
func (a *Linear) SetOverrideDomain(v bool) { a.overrideDomain = v }

// Label implements Axis.
func (a *Linear) Label() string { return a.title }

// Ticks implements Axis.
func (a *Linear) Ticks(pixelMin, pixelMax float32) []Tick {
	dMin, dMax := a.sc.Domain()
	values := GenerateNiceTicks(dMin, dMax, 8)

	// Expand domain to match nice tick range so gridlines and
	// data points use the same coordinate space.
	if a.autoRange && !a.overrideDomain && len(values) >= 2 {
		a.sc.SetDomain(values[0], values[len(values)-1])
	}

	ticks := make([]Tick, 0, len(values))
	for _, v := range values {
		// When the domain is overridden (zoomed), skip ticks
		// outside the domain so off-screen labels aren't drawn.
		if a.overrideDomain && (v < dMin-1e-9 || v > dMax+1e-9) {
			continue
		}
		label := formatTickValue(v)
		if a.tickFormat != nil {
			label = a.tickFormat(v)
		}
		ticks = append(ticks, Tick{
			Value:    v,
			Label:    label,
			Position: a.Transform(v, pixelMin, pixelMax),
		})
	}
	return ticks
}

func formatTickValue(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%g", v)
}

// Transform implements Axis.
func (a *Linear) Transform(value float64, pixelMin, pixelMax float32) float32 {
	return a.sc.Transform(value, pixelMin, pixelMax)
}

// Invert implements Axis.
func (a *Linear) Invert(pixel, pixelMin, pixelMax float32) float64 {
	return a.sc.Invert(pixel, pixelMin, pixelMax)
}
