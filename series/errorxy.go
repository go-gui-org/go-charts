package series

import (
	"fmt"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-gui/gui"
)

// ErrorBar holds asymmetric error bounds as absolute distances
// from the data value. Zero value means no error bar.
type ErrorBar struct {
	Low, High float64
}

// Symmetric returns an ErrorBar with equal low and high bounds.
func Symmetric(v float64) ErrorBar {
	return ErrorBar{Low: v, High: v}
}

// ErrorPoint is a data point with optional X and Y error bars.
type ErrorPoint struct {
	X, Y float64
	YErr ErrorBar
	XErr ErrorBar
}

// ErrorXY is a series of data points with error bars.
type ErrorXY struct {
	name   string
	color  gui.Color
	Points []ErrorPoint
}

// ErrorXYCfg configures an ErrorXY series.
type ErrorXYCfg struct {
	Name   string
	Color  gui.Color
	Points []ErrorPoint
}

// NewErrorXY creates a new ErrorXY data series.
func NewErrorXY(cfg ErrorXYCfg) ErrorXY {
	return ErrorXY{
		name:   cfg.Name,
		color:  cfg.Color,
		Points: cfg.Points,
	}
}

// Name implements Series.
func (s ErrorXY) Name() string { return s.name }

// Len implements Series.
func (s ErrorXY) Len() int { return len(s.Points) }

// Color implements Series.
func (s ErrorXY) Color() gui.Color { return s.color }

// String implements fmt.Stringer.
func (s ErrorXY) String() string {
	return fmt.Sprintf("ErrorXY{%q, %d points}", s.name, len(s.Points))
}

// Bounds returns the min/max X and Y values including error
// extents. Non-finite points are skipped. If no finite points
// exist, all returned values are zero.
func (s ErrorXY) Bounds() (minX, maxX, minY, maxY float64) {
	i := 0
	for i < len(s.Points) {
		p := s.Points[i]
		if fmath.Finite(p.X) && fmath.Finite(p.Y) {
			break
		}
		i++
	}
	if i >= len(s.Points) {
		return
	}
	p := s.Points[i]
	minX, maxX = p.X-clampNeg(p.XErr.Low), p.X+clampNeg(p.XErr.High)
	minY, maxY = p.Y-clampNeg(p.YErr.Low), p.Y+clampNeg(p.YErr.High)
	for _, p := range s.Points[i+1:] {
		if !fmath.Finite(p.X) || !fmath.Finite(p.Y) {
			continue
		}
		minX = min(minX, p.X-clampNeg(p.XErr.Low))
		maxX = max(maxX, p.X+clampNeg(p.XErr.High))
		minY = min(minY, p.Y-clampNeg(p.YErr.Low))
		maxY = max(maxY, p.Y+clampNeg(p.YErr.High))
	}
	return
}

// clampNeg returns v if v is finite and >= 0, otherwise 0.
// Non-finite (NaN, +/-Inf) and negative error widths are
// treated as zero so bounds remain finite.
func clampNeg(v float64) float64 {
	if !fmath.Finite(v) || v < 0 {
		return 0
	}
	return v
}
