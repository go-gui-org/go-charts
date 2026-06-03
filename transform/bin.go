package transform

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"sort"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-charts/series"
)

// AggFunc specifies how to aggregate Y values within a bin.
type AggFunc int

// Aggregation functions for binned data.
const (
	AggSum   AggFunc = iota // Sum of Y values.
	AggMean                 // Mean of Y values.
	AggCount                // Number of points.
	AggMin                  // Minimum Y value.
	AggMax                  // Maximum Y value.
)

// BinCfg configures data binning.
type BinCfg struct {
	// Bins is the number of bins. 0 = auto (Sturges' rule).
	Bins int
	// Edges provides explicit bin boundaries. Overrides Bins if
	// non-empty. Must be sorted ascending with at least 2 values.
	Edges []float64
	// Aggregation is the function applied to Y values in each bin.
	// Default: AggSum.
	Aggregation AggFunc
}

// Bin groups XY data into bins along the X axis and applies the
// configured aggregation to Y values. Returns one point per bin
// with X at the bin center. Returns empty XY on invalid input.
func Bin(s series.XY, cfg BinCfg) series.XY {
	// Collect finite points sorted by X.
	finite := make([]series.Point, 0, len(s.Points))
	for _, p := range s.Points {
		if fmath.Finite(p.X) && fmath.Finite(p.Y) {
			finite = append(finite, p)
		}
	}
	if len(finite) == 0 {
		return series.XY{}
	}
	slices.SortFunc(finite, func(a, b series.Point) int {
		return cmp.Compare(a.X, b.X)
	})

	edges := cfg.Edges
	if len(edges) < 2 {
		edges = autoEdges(finite, cfg.Bins)
	}
	nBins := len(edges) - 1
	if nBins < 1 {
		return series.XY{}
	}
	if nBins > maxBins {
		nBins = maxBins
		edges = edges[:nBins+1]
	}

	// Assign points to bins. Convention: bin i covers
	// [edges[i], edges[i+1]) except the last bin which includes
	// edges[nBins].
	buckets := make([][]float64, nBins)
	for _, p := range finite {
		// First edge index strictly greater than p.X, minus 1.
		idx := sort.Search(len(edges), func(i int) bool {
			return edges[i] > p.X
		}) - 1
		idx = max(idx, 0)
		idx = min(idx, nBins-1)
		buckets[idx] = append(buckets[idx], p.Y)
	}

	pts := make([]series.Point, nBins)
	for i := range nBins {
		center := (edges[i] + edges[i+1]) / 2
		pts[i] = series.Point{X: center, Y: aggregate(buckets[i], cfg.Aggregation)}
	}
	return namedXY(fmt.Sprintf("%s Binned", s.Name()), pts)
}

// autoEdges computes bin edges using Sturges' rule.
func autoEdges(sorted []series.Point, nBins int) []float64 {
	n := len(sorted)
	if nBins <= 0 {
		nBins = int(math.Ceil(math.Log2(float64(n)) + 1))
	}
	nBins = max(nBins, 1)
	nBins = min(nBins, maxBins)
	lo := sorted[0].X
	hi := sorted[n-1].X
	if lo == hi {
		hi = lo + 1
	}
	step := (hi - lo) / float64(nBins)
	edges := make([]float64, nBins+1)
	for i := range nBins + 1 {
		edges[i] = lo + float64(i)*step
	}
	return edges
}

func aggregate(ys []float64, agg AggFunc) float64 {
	if len(ys) == 0 {
		return 0
	}
	switch agg {
	case AggMean:
		s := 0.0
		for _, y := range ys {
			s += y
		}
		return s / float64(len(ys))
	case AggCount:
		return float64(len(ys))
	case AggMin:
		return slices.Min(ys)
	case AggMax:
		return slices.Max(ys)
	default: // AggSum
		s := 0.0
		for _, y := range ys {
			s += y
		}
		return s
	}
}
