package chart

import (
	"fmt"

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
	mx, my float32,
) {
	si, pi, px, py, ok := nearestStackedPoint(serieses, pa, mx, my, 20)
	if !ok {
		return
	}
	s := serieses[si]
	p := s.Points[pi]
	var label string
	if s.Name() != "" {
		label = fmt.Sprintf(
			"%s\nX: %g\nY: %g", s.Name(), p.X, p.Y)
	} else {
		label = fmt.Sprintf("X: %g\nY: %g", p.X, p.Y)
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
	mx, my, snapPx float32,
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
		label = fmt.Sprintf(
			"%s\nX: %g\nY: %g\nSize: %g",
			s.Name(), p.X, p.Y, p.Z)
	} else {
		label = fmt.Sprintf(
			"X: %g\nY: %g\nSize: %g", p.X, p.Y, p.Z)
	}
	drawTooltip(ctx, px, py, label, th, pa.plotRect)
}

// drawXYTooltip draws a tooltip for the nearest XY data point.
// Shared by line, scatter, and area charts.
func drawXYTooltip(
	ctx *render.Context, th *theme.Theme,
	serieses []series.XY, pa plotArea,
	mx, my float32,
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
		label = fmt.Sprintf(
			"%s\nX: %g\nY: %g", s.Name(), p.X, p.Y)
	} else {
		label = fmt.Sprintf("X: %g\nY: %g", p.X, p.Y)
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
	mx, my float32,
) {
	si, pi, px, py, ok := nearestErrorXYPoint(
		serieses, pa, mx, my, 20)
	if !ok {
		return
	}
	s := serieses[si]
	p := s.Points[pi]
	label := formatErrorPointLabel(s.Name(), p)
	drawTooltip(ctx, px, py, label, th, pa.plotRect)
}

// formatErrorPointLabel builds a tooltip string for an
// ErrorPoint, including error bounds when non-zero.
func formatErrorPointLabel(name string, p series.ErrorPoint) string {
	noErr := series.ErrorBar{}
	var label string
	if name != "" {
		label = name + "\n"
	}
	label += fmt.Sprintf("X: %g\nY: %g", p.X, p.Y)
	if p.YErr != noErr {
		label += fmt.Sprintf("\nY err: +%g/-%g", p.YErr.High, p.YErr.Low)
	}
	if p.XErr != noErr {
		label += fmt.Sprintf("\nX err: +%g/-%g", p.XErr.High, p.XErr.Low)
	}
	return label
}
