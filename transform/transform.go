// Package transform provides pure data transforms for series.XY
// data: moving averages, regression, envelopes, downsampling, and
// binning. All functions accept series.XY and return series.XY.
package transform

import (
	"math"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-charts/series"
)

const (
	// maxWindow caps the sliding-window size for moving averages
	// and envelopes to prevent O(n*w) blowup.
	maxWindow = 10_000
	// maxDegree caps polynomial degree to prevent large matrix
	// allocations in normal equations (matrix is (d+1)²).
	maxDegree = 20
	// maxOutputPoints caps output slice length for functions that
	// accept a user-controlled point count.
	maxOutputPoints = 1_000_000
	// maxBins caps the number of bins to prevent unbounded
	// bucket-slice allocation.
	maxBins = 10_000
)

// finiteIndices returns a slice of indices into pts where both X
// and Y are finite.
func finiteIndices(pts []series.Point) []int {
	idx := make([]int, 0, len(pts))
	for i, p := range pts {
		if fmath.Finite(p.X) && fmath.Finite(p.Y) {
			idx = append(idx, i)
		}
	}
	return idx
}

// makeNaNPoints copies X values from src into a new slice with all
// Y values set to NaN.
func makeNaNPoints(src []series.Point) []series.Point {
	pts := make([]series.Point, len(src))
	for i, p := range src {
		pts[i] = series.Point{X: p.X, Y: math.NaN()}
	}
	return pts
}

// namedXY builds an XY series with a formatted name.
func namedXY(name string, pts []series.Point) series.XY {
	return series.NewXY(series.XYCfg{Name: name, Points: pts})
}
