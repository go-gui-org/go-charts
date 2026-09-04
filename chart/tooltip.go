package chart

import (
	"fmt"
	"math"
	"strconv"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/render"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-charts/theme"
)

// plotRect holds the pixel bounds of the chart's data region.
type plotRect struct {
	Left, Right, Top, Bottom float32
}

// plotArea describes the pixel bounds and axes of the chart's data
// region. Passed to tooltip helpers to avoid long parameter lists.
type plotArea struct {
	plotRect
	XAxis, YAxis axis.Axis
}

// nearestXYPoint finds the series/point index and pixel position
// of the data point closest to (mx, my) within snapPx pixels.
// Returns ok=false when no point is within the threshold.
func nearestXYPoint(
	serieses []series.XY, pa plotArea,
	mx, my, snapPx float32,
) (si, pi int, px, py float32, ok bool) {
	best := snapPx * snapPx
	for i, s := range serieses {
		for j, p := range s.Points {
			if !finite(p.X) || !finite(p.Y) {
				continue
			}
			ppx := pa.XAxis.Transform(p.X, pa.Left, pa.Right)
			ppy := pa.YAxis.Transform(p.Y, pa.Bottom, pa.Top)
			dx := ppx - mx
			dy := ppy - my
			d2 := dx*dx + dy*dy
			if d2 < best {
				best = d2
				si, pi, px, py = i, j, ppx, ppy
				ok = true
			}
		}
	}
	return
}

// nearestStackedPoint finds the series/point index and stacked pixel position
// of the data point closest to (mx, my) in a stacked area chart.
// Cumulative Y values are used so the comparison matches the drawn geometry.
func nearestStackedPoint(
	serieses []series.XY, pa plotArea,
	mx, my, snapPx float32,
) (si, pi int, px, py float32, ok bool) {
	if len(serieses) == 0 {
		return
	}
	refLen := 0
	for _, s := range serieses {
		if s.Len() > 0 {
			refLen = s.Len()
			break
		}
	}
	if refLen == 0 {
		return
	}

	cumY := make([]float64, refLen)
	best := snapPx * snapPx

	for i, s := range serieses {
		if s.Len() == 0 {
			continue
		}
		n := min(s.Len(), refLen)
		for j := range n {
			p := s.Points[j]
			if !finite(p.X) || !finite(p.Y) {
				continue
			}
			cumY[j] += p.Y
			ppx := pa.XAxis.Transform(p.X, pa.Left, pa.Right)
			ppy := pa.YAxis.Transform(cumY[j], pa.Bottom, pa.Top)
			dx := ppx - mx
			dy := ppy - my
			d2 := dx*dx + dy*dy
			if d2 < best {
				best = d2
				si, pi, px, py = i, j, ppx, ppy
				ok = true
			}
		}
	}
	return
}

// drawStackedXYTooltip draws a tooltip for the nearest point in a stacked
// area chart. Uses cumulative Y positions for hit-testing so the result
// matches the drawn geometry.
func drawStackedXYTooltip(
	ctx *render.Context, th *theme.Theme,
	serieses []series.XY, pa plotArea,
	mx, my float32, xName, yName string,
) {
	si, pi, px, py, ok := nearestStackedPoint(serieses, pa, mx, my, 20)
	if !ok {
		return
	}
	s := serieses[si]
	p := s.Points[pi]
	var label string
	if s.Name() != "" {
		label = fmt.Sprintf("%s\n%s: %s\n%s: %s", s.Name(),
			xName, tooltipNum(p.X), yName, tooltipNum(p.Y))
	} else {
		label = fmt.Sprintf("%s: %s\n%s: %s",
			xName, tooltipNum(p.X), yName, tooltipNum(p.Y))
	}
	drawTooltip(ctx, px, py, label, th, pa.plotRect)
}

// nearestXYZPoint finds the series/point index and pixel position
// of the XYZ data point closest to (mx, my) within snapPx pixels.
// Returns ok=false when no point is within the threshold.
func nearestXYZPoint(
	serieses []series.XYZ, pa plotArea,
	mx, my, snapPx float32,
) (si, pi int, px, py float32, ok bool) {
	best := snapPx * snapPx
	for i, s := range serieses {
		for j, p := range s.Points {
			if !finite(p.X) || !finite(p.Y) {
				continue
			}
			ppx := pa.XAxis.Transform(p.X, pa.Left, pa.Right)
			ppy := pa.YAxis.Transform(p.Y, pa.Bottom, pa.Top)
			dx := ppx - mx
			dy := ppy - my
			d2 := dx*dx + dy*dy
			if d2 < best {
				best = d2
				si, pi, px, py = i, j, ppx, ppy
				ok = true
			}
		}
	}
	return
}

// drawXYZTooltip draws a tooltip for the nearest XYZ data point
// showing X, Y, and Size values. snapPx is the maximum pixel
// distance from a point center to trigger the tooltip.
func drawXYZTooltip(
	ctx *render.Context, th *theme.Theme,
	serieses []series.XYZ, pa plotArea,
	mx, my, snapPx float32, xName, yName string,
) {
	si, pi, px, py, ok := nearestXYZPoint(
		serieses, pa, mx, my, snapPx)
	if !ok {
		return
	}
	s := serieses[si]
	p := s.Points[pi]
	var label string
	if s.Name() != "" {
		label = fmt.Sprintf("%s\n%s: %s\n%s: %s\nSize: %s",
			s.Name(), xName, tooltipNum(p.X),
			yName, tooltipNum(p.Y), tooltipNum(p.Z))
	} else {
		label = fmt.Sprintf("%s: %s\n%s: %s\nSize: %s",
			xName, tooltipNum(p.X),
			yName, tooltipNum(p.Y), tooltipNum(p.Z))
	}
	drawTooltip(ctx, px, py, label, th, pa.plotRect)
}

// drawXYTooltip draws a tooltip for the nearest XY data point.
// Shared by line, scatter, and area charts.
func drawXYTooltip(
	ctx *render.Context, th *theme.Theme,
	serieses []series.XY, pa plotArea,
	mx, my float32, xName, yName string,
) {
	si, pi, px, py, ok := nearestXYPoint(
		serieses, pa, mx, my, 20)
	if !ok {
		return
	}
	s := serieses[si]
	p := s.Points[pi]
	var label string
	if s.Name() != "" {
		label = fmt.Sprintf("%s\n%s: %s\n%s: %s", s.Name(),
			xName, tooltipNum(p.X), yName, tooltipNum(p.Y))
	} else {
		label = fmt.Sprintf("%s: %s\n%s: %s",
			xName, tooltipNum(p.X), yName, tooltipNum(p.Y))
	}
	drawTooltip(ctx, px, py, label, th, pa.plotRect)
}

// nearestErrorXYPoint finds the series/point index and pixel
// position of the ErrorXY data point closest to (mx, my)
// within snapPx pixels.
func nearestErrorXYPoint(
	serieses []series.ErrorXY, pa plotArea,
	mx, my, snapPx float32,
) (si, pi int, px, py float32, ok bool) {
	best := snapPx * snapPx
	for i, s := range serieses {
		for j, p := range s.Points {
			if !finite(p.X) || !finite(p.Y) {
				continue
			}
			ppx := pa.XAxis.Transform(p.X, pa.Left, pa.Right)
			ppy := pa.YAxis.Transform(p.Y, pa.Bottom, pa.Top)
			dx := ppx - mx
			dy := ppy - my
			d2 := dx*dx + dy*dy
			if d2 < best {
				best = d2
				si, pi, px, py = i, j, ppx, ppy
				ok = true
			}
		}
	}
	return
}

// drawErrorXYTooltip draws a tooltip for the nearest ErrorXY
// data point showing X, Y values and error bounds.
func drawErrorXYTooltip(
	ctx *render.Context, th *theme.Theme,
	serieses []series.ErrorXY, pa plotArea,
	mx, my float32, xName, yName string,
) {
	si, pi, px, py, ok := nearestErrorXYPoint(
		serieses, pa, mx, my, 20)
	if !ok {
		return
	}
	s := serieses[si]
	p := s.Points[pi]
	label := formatErrorPointLabel(s.Name(), p, xName, yName)
	drawTooltip(ctx, px, py, label, th, pa.plotRect)
}

// formatErrorPointLabel builds a tooltip string for an
// ErrorPoint, including error bounds when non-zero.
func formatErrorPointLabel(
	name string, p series.ErrorPoint, xName, yName string,
) string {
	noErr := series.ErrorBar{}
	var label string
	if name != "" {
		label = name + "\n"
	}
	label += fmt.Sprintf("%s: %s\n%s: %s",
		xName, tooltipNum(p.X), yName, tooltipNum(p.Y))
	if p.YErr != noErr {
		label += fmt.Sprintf("\n%s err: +%s/-%s", yName,
			tooltipNum(p.YErr.High), tooltipNum(p.YErr.Low))
	}
	if p.XErr != noErr {
		label += fmt.Sprintf("\n%s err: +%s/-%s", xName,
			tooltipNum(p.XErr.High), tooltipNum(p.XErr.Low))
	}
	return label
}

// tooltipNum formats a data value for a tooltip.
//
// Tooltips read as a caption on a hovered point, not as a data dump,
// so one decimal is enough: %g prints a float's full precision, and a
// reading like 24.533333333333335 buries the two digits the reader
// wants. Whole numbers keep no decimal at all, and values too small
// for one decimal fall back to %g so they do not collapse to "0.0".
func tooltipNum(v float64) string {
	switch {
	case math.IsNaN(v) || math.IsInf(v, 0):
		return fmt.Sprintf("%g", v)
	// Outside this band a fixed-point form is worse than %g: below
	// it one decimal rounds the value away to "0.0", above it the
	// digits run past the width of any tooltip.
	case math.Abs(v) >= 1e15 || (v != 0 && math.Abs(v) < 0.05):
		return fmt.Sprintf("%g", v)
	case v == math.Trunc(v):
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return strconv.FormatFloat(v, 'f', 1, 64)
	}
}
