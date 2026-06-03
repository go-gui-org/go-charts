package series

import (
	"fmt"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-gui/gui"
)

// XYZPoint represents a single (X, Y, Z) data point where Z
// controls marker size in bubble charts.
type XYZPoint struct {
	X, Y, Z float64
}

// XYZ is a series of (X, Y, Z) data points for bubble charts.
type XYZ struct {
	name   string
	color  gui.Color
	Points []XYZPoint
}

// XYZCfg configures an XYZ series.
type XYZCfg struct {
	Name   string
	Color  gui.Color
	Points []XYZPoint
}

// NewXYZ creates a new XYZ data series.
func NewXYZ(cfg XYZCfg) XYZ {
	return XYZ{
		name:   cfg.Name,
		color:  cfg.Color,
		Points: cfg.Points,
	}
}

// XYZFromSlices creates an XYZ series from parallel X, Y, and Z
// slices. Returns an error if slice lengths differ.
func XYZFromSlices(name string, xVals, yVals, zVals []float64) (XYZ, error) {
	if len(xVals) != len(yVals) || len(xVals) != len(zVals) {
		return XYZ{}, fmt.Errorf(
			"series.XYZFromSlices: len(x)=%d, len(y)=%d, len(z)=%d",
			len(xVals), len(yVals), len(zVals))
	}
	pts := make([]XYZPoint, len(xVals))
	for i := range xVals {
		pts[i] = XYZPoint{X: xVals[i], Y: yVals[i], Z: zVals[i]}
	}
	return XYZ{name: name, Points: pts}, nil
}

// Name implements Series.
func (s XYZ) Name() string { return s.name }

// Len implements Series.
func (s XYZ) Len() int { return len(s.Points) }

// Color implements Series.
func (s XYZ) Color() gui.Color { return s.color }

// String implements fmt.Stringer.
func (s XYZ) String() string {
	return fmt.Sprintf("XYZ{%q, %d points}", s.name, len(s.Points))
}

// String implements fmt.Stringer.
func (p XYZPoint) String() string {
	return fmt.Sprintf("(%.4g, %.4g, %.4g)", p.X, p.Y, p.Z)
}

// Bounds returns the min/max X and Y values. Non-finite points
// (NaN, +/-Inf) are skipped. Z is excluded because it controls
// marker size, not position. If no finite points exist, all
// returned values are zero.
func (s XYZ) Bounds() (minX, maxX, minY, maxY float64) {
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

// ZBounds returns the min/max Z values across all finite points.
// If no finite Z exists, both values are zero.
func (s XYZ) ZBounds() (minZ, maxZ float64) {
	i := 0
	for i < len(s.Points) {
		if fmath.Finite(s.Points[i].Z) {
			break
		}
		i++
	}
	if i >= len(s.Points) {
		return
	}
	minZ, maxZ = s.Points[i].Z, s.Points[i].Z
	for _, p := range s.Points[i+1:] {
		if !fmath.Finite(p.Z) {
			continue
		}
		minZ = min(minZ, p.Z)
		maxZ = max(maxZ, p.Z)
	}
	return
}
