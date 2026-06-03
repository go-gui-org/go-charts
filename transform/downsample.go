package transform

import (
	"fmt"
	"math"
	"slices"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-charts/series"
)

// LTTB returns a downsampled series of exactly threshold points
// using the Largest-Triangle-Three-Buckets algorithm (Steinarsson
// 2013). If threshold >= len(s.Points) or threshold < 3, returns a
// copy. Returns empty XY if input is empty.
func LTTB(s series.XY, threshold int) series.XY {
	// Filter non-finite points to prevent NaN cascading through
	// bucket area calculations.
	clean := make([]series.Point, 0, len(s.Points))
	for _, p := range s.Points {
		if fmath.Finite(p.X) && fmath.Finite(p.Y) {
			clean = append(clean, p)
		}
	}
	n := len(clean)
	if n == 0 {
		return series.XY{}
	}
	if threshold >= n || threshold < 3 {
		return namedXY(
			fmt.Sprintf("%s (downsampled)", s.Name()), slices.Clone(clean))
	}

	out := make([]series.Point, 0, threshold)
	out = append(out, clean[0]) // always keep first

	bucketSize := float64(n-2) / float64(threshold-2)

	prevSelected := 0

	for i := range threshold - 2 {
		// Current bucket range.
		bStart := int(float64(i)*bucketSize) + 1
		bEnd := min(int(float64(i+1)*bucketSize)+1, n-1)

		// Next bucket average (for triangle area calculation).
		nStart := int(float64(i+1)*bucketSize) + 1
		nEnd := min(int(float64(i+2)*bucketSize)+1, n-1)
		avgX, avgY := 0.0, 0.0
		nCount := nEnd - nStart
		if nCount <= 0 {
			nCount = 1
			nStart = n - 1
			nEnd = n
		}
		for j := nStart; j < nEnd; j++ {
			avgX += clean[j].X
			avgY += clean[j].Y
		}
		avgX /= float64(nCount)
		avgY /= float64(nCount)

		// Select point in current bucket with largest triangle area.
		bestArea := -1.0
		bestIdx := bStart
		px := clean[prevSelected].X
		py := clean[prevSelected].Y

		for j := bStart; j < bEnd; j++ {
			area := math.Abs(
				(px-avgX)*(clean[j].Y-py) -
					(px-clean[j].X)*(avgY-py))
			if area > bestArea {
				bestArea = area
				bestIdx = j
			}
		}
		out = append(out, clean[bestIdx])
		prevSelected = bestIdx
	}

	out = append(out, clean[n-1]) // always keep last
	return namedXY(
		fmt.Sprintf("%s (downsampled)", s.Name()), out)
}
