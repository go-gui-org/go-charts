package transform

import (
	"fmt"
	"math"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-charts/series"
)

// BollingerBands returns upper, middle (SMA), and lower band series.
// window is the SMA period; mult is the standard-deviation multiplier
// (typically 2.0). Returns empty series on invalid input.
func BollingerBands(
	s series.XY, window int, mult float64,
) (upper, middle, lower series.XY) {
	if window < 1 || window > maxWindow || mult < 0 ||
		len(s.Points) == 0 {
		return
	}
	mid := SMA(s, window)
	uPts := makeNaNPoints(s.Points)
	lPts := makeNaNPoints(s.Points)

	for i := range s.Points {
		if math.IsNaN(mid.Points[i].Y) {
			continue
		}
		// SMA is defined at i, so all window values are finite.
		mean := mid.Points[i].Y
		sumSq := 0.0
		for j := i - window + 1; j <= i; j++ {
			d := s.Points[j].Y - mean
			sumSq += d * d
		}
		sd := math.Sqrt(sumSq / float64(window))
		uPts[i].Y = mean + mult*sd
		lPts[i].Y = mean - mult*sd
	}

	name := s.Name()
	upper = namedXY(fmt.Sprintf("%s BB Upper", name), uPts)
	middle = namedXY(fmt.Sprintf("%s BB Middle", name), mid.Points)
	lower = namedXY(fmt.Sprintf("%s BB Lower", name), lPts)
	return
}

// MinMaxEnvelope returns rolling maximum and minimum envelopes over
// a sliding window of size n. Positions with fewer than n preceding
// finite values produce NaN. Returns empty series on invalid input.
func MinMaxEnvelope(
	s series.XY, n int,
) (maxEnv, minEnv series.XY) {
	if n < 1 || n > maxWindow || len(s.Points) == 0 {
		return
	}
	maxPts := makeNaNPoints(s.Points)
	minPts := makeNaNPoints(s.Points)

	for i := range s.Points {
		if i < n-1 {
			continue
		}
		hi := math.Inf(-1)
		lo := math.Inf(+1)
		allFinite := true
		for j := i - n + 1; j <= i; j++ {
			y := s.Points[j].Y
			if !fmath.Finite(y) {
				allFinite = false
				break
			}
			if y > hi {
				hi = y
			}
			if y < lo {
				lo = y
			}
		}
		if allFinite {
			maxPts[i].Y = hi
			minPts[i].Y = lo
		}
	}

	name := s.Name()
	maxEnv = namedXY(fmt.Sprintf("%s Max(%d)", name, n), maxPts)
	minEnv = namedXY(fmt.Sprintf("%s Min(%d)", name, n), minPts)
	return
}
