package series

import (
	"fmt"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-gui/gui"
)

// Point represents a single (X, Y) data point.
type Point struct {
	X, Y float64
}

// XY is a series of (X, Y) data points.
type XY struct {
	name   string
	color  gui.Color
	Points []Point
}

// XYCfg configures an XY series.
type XYCfg struct {
	Name   string
	Color  gui.Color
	Points []Point
}

// NewXY creates a new XY data series.
func NewXY(cfg XYCfg) XY {
	return XY{
		name:   cfg.Name,
		color:  cfg.Color,
		Points: cfg.Points,
	}
}

// XYFromSlices creates an XY series from parallel X and Y slices.
// Returns an error if len(xVals) != len(yVals).
func XYFromSlices(name string, xVals, yVals []float64) (XY, error) {
	if len(xVals) != len(yVals) {
		return XY{}, fmt.Errorf(
			"series.XYFromSlices: len(xVals)=%d != len(yVals)=%d",
			len(xVals), len(yVals))
	}
	pts := make([]Point, len(xVals))
	for i := range xVals {
		pts[i] = Point{X: xVals[i], Y: yVals[i]}
	}
	return XY{name: name, Points: pts}, nil
}

// XYFromYValues creates an XY series with auto-indexed X values
// (0, 1, 2, ...).
func XYFromYValues(name string, yVals []float64) XY {
	pts := make([]Point, len(yVals))
	for i, y := range yVals {
		pts[i] = Point{X: float64(i), Y: y}
	}
	return XY{name: name, Points: pts}
}

// Name implements Series.
func (s XY) Name() string { return s.name }

// Len implements Series.
func (s XY) Len() int { return len(s.Points) }

// Color implements Series.
func (s XY) Color() gui.Color { return s.color }

// String implements fmt.Stringer.
func (s XY) String() string {
	return fmt.Sprintf("XY{%q, %d points}", s.name, len(s.Points))
}

// String implements fmt.Stringer.
func (p Point) String() string {
	return fmt.Sprintf("(%.4g, %.4g)", p.X, p.Y)
}

// Bounds returns the min/max X and Y values. Non-finite points
// (NaN, +/-Inf) are skipped. If no finite points exist, all
// returned values are zero.
func (s XY) Bounds() (minX, maxX, minY, maxY float64) {
	// Find first finite point to seed min/max.
	i := 0
	for i < len(s.Points) {
		p := s.Points[i]
		if fmath.Finite(p.X) && fmath.Finite(p.Y) {
			break
		}
		i++
	}
	if i >= len(s.Points) {
		return // no finite points
	}
	minX, maxX = s.Points[i].X, s.Points[i].X
	minY, maxY = s.Points[i].Y, s.Points[i].Y
	for _, p := range s.Points[i+1:] {
		if !fmath.Finite(p.X) || !fmath.Finite(p.Y) {
			continue
		}
		minX = min(minX, p.X)
		maxX = max(maxX, p.X)
		minY = min(minY, p.Y)
		maxY = max(maxY, p.Y)
	}
	return
}
