package transform

import (
	"fmt"
	"math"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-charts/series"
)

// SMA returns a simple moving average of s over a sliding window of
// size n. Positions with fewer than n preceding finite values produce
// NaN. Returns empty XY if n < 1 or input is empty.
func SMA(s series.XY, n int) series.XY {
	if n < 1 || n > maxWindow || len(s.Points) == 0 {
		return series.XY{}
	}
	pts := makeNaNPoints(s.Points)
	sum := 0.0
	count := 0
	for i, p := range s.Points {
		if fmath.Finite(p.Y) {
			sum += p.Y
			count++
		}
		if i >= n {
			old := s.Points[i-n]
			if fmath.Finite(old.Y) {
				sum -= old.Y
				count--
			}
		}
		if count == n {
			pts[i].Y = sum / float64(n)
		}
	}
	return namedXY(fmt.Sprintf("%s SMA(%d)", s.Name(), n), pts)
}

// EMA returns an exponential moving average with smoothing factor
// alpha = 2/(n+1). The first finite value seeds the EMA. Returns
// empty XY if n < 1 or input is empty.
func EMA(s series.XY, n int) series.XY {
	if n < 1 || n > maxWindow || len(s.Points) == 0 {
		return series.XY{}
	}
	pts := makeNaNPoints(s.Points)
	alpha := 2.0 / float64(n+1)
	ema := math.NaN()
	for i, p := range s.Points {
		if !fmath.Finite(p.Y) {
			continue
		}
		if math.IsNaN(ema) {
			ema = p.Y
		} else {
			ema = alpha*p.Y + (1-alpha)*ema
		}
		pts[i].Y = ema
	}
	return namedXY(fmt.Sprintf("%s EMA(%d)", s.Name(), n), pts)
}

// WMA returns a weighted moving average where recent values have
// linearly increasing weight (1, 2, ..., n). Positions with fewer
// than n preceding finite values produce NaN. Returns empty XY if
// n < 1 or input is empty.
func WMA(s series.XY, n int) series.XY {
	if n < 1 || n > maxWindow || len(s.Points) == 0 {
		return series.XY{}
	}
	pts := makeNaNPoints(s.Points)
	denom := float64(n) * float64(n+1) / 2
	for i := range s.Points {
		if i < n-1 {
			continue
		}
		wsum := 0.0
		allFinite := true
		for j := range n {
			y := s.Points[i-n+1+j].Y
			if !fmath.Finite(y) {
				allFinite = false
				break
			}
			wsum += float64(j+1) * y
		}
		if allFinite {
			pts[i].Y = wsum / denom
		}
	}
	return namedXY(fmt.Sprintf("%s WMA(%d)", s.Name(), n), pts)
}
