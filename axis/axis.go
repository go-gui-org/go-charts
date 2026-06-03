// Package axis provides axis types for chart positioning and labeling.
package axis

import (
	"fmt"

	"github.com/go-gui-org/go-gui/gui"
)

// Axis defines the interface for chart axes.
type Axis interface {
	// Label returns the axis title.
	Label() string

	// Ticks returns the tick positions and labels for the given
	// pixel range and data domain.
	Ticks(pixelMin, pixelMax float32) []Tick

	// Transform converts a data value to a pixel position.
	Transform(value float64, pixelMin, pixelMax float32) float32

	// Invert converts a pixel position back to a data value.
	Invert(pixel, pixelMin, pixelMax float32) float64

	// SetRange updates the axis data range.
	SetRange(min, max float64)

	// Domain returns the current data range.
	Domain() (float64, float64)

	// SetOverrideDomain controls whether Ticks() skips
	// auto-range expansion. When true, the domain set by
	// SetRange is preserved exactly.
	SetOverrideDomain(v bool)
}

// Position indicates where an axis is drawn.
type Position uint8

// Position constants.
const (
	Bottom Position = iota
	Top
	Left
	Right
)

// Tick represents a single tick mark on an axis.
type Tick struct {
	Value    float64
	Label    string
	Position float32
	Minor    bool
}

// GridLine describes a grid line associated with a tick.
type GridLine struct {
	Position float32
	Color    gui.Color
	Width    float32
}

// String implements fmt.Stringer.
func (t Tick) String() string {
	return fmt.Sprintf("Tick{%q @ %.1f}", t.Label, t.Position)
}

// TickFormat converts a numeric axis value to its display string.
// When nil, the axis uses its default formatting.
type TickFormat func(float64) string
