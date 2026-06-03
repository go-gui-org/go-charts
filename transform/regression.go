package transform

import (
	"fmt"
	"math"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-charts/series"
)

// LinearRegression returns a 2-point series representing the
// least-squares best-fit line through s. Returns empty XY if fewer
// than 2 finite points exist.
func LinearRegression(s series.XY) series.XY {
	slope, intercept, ok := linearFit(s.Points)
	if !ok {
		return series.XY{}
	}
	minX, maxX := finiteBoundsX(s.Points)
	pts := []series.Point{
		{X: minX, Y: slope*minX + intercept},
		{X: maxX, Y: slope*maxX + intercept},
	}
	return namedXY(fmt.Sprintf("%s Linear Fit", s.Name()), pts)
}

// LinearTrend returns a series with the same X values as s but Y
// replaced by the regression line prediction. Returns empty XY if
// fewer than 2 finite points exist.
func LinearTrend(s series.XY) series.XY {
	slope, intercept, ok := linearFit(s.Points)
	if !ok {
		return series.XY{}
	}
	pts := make([]series.Point, len(s.Points))
	for i, p := range s.Points {
		pts[i] = series.Point{X: p.X, Y: slope*p.X + intercept}
	}
	return namedXY(fmt.Sprintf("%s Linear Trend", s.Name()), pts)
}

// PolynomialRegression fits a polynomial of the given degree to s
// and evaluates it at nPoints evenly-spaced X values. degree must
// be >= 1 and < number of finite points. nPoints <= 0 defaults to
// len(s.Points). Returns empty XY on invalid input.
func PolynomialRegression(s series.XY, degree, nPoints int) series.XY {
	idx := finiteIndices(s.Points)
	n := len(idx)
	if degree < 1 || degree > maxDegree || n <= degree {
		return series.XY{}
	}
	if nPoints <= 0 {
		nPoints = len(s.Points)
	}
	if nPoints > maxOutputPoints {
		nPoints = maxOutputPoints
	}

	// Extract finite points and center/scale X for stability.
	xs := make([]float64, n)
	ys := make([]float64, n)
	sumX := 0.0
	for i, k := range idx {
		xs[i] = s.Points[k].X
		ys[i] = s.Points[k].Y
		sumX += xs[i]
	}
	meanX := sumX / float64(n)
	scaleX := 0.0
	for _, x := range xs {
		d := math.Abs(x - meanX)
		if d > scaleX {
			scaleX = d
		}
	}
	if scaleX == 0 {
		scaleX = 1
	}
	for i := range xs {
		xs[i] = (xs[i] - meanX) / scaleX
	}

	// Precompute x^k for k = 0..2*degree to avoid math.Pow in
	// the normal equations loop.
	m := degree + 1
	maxPow := 2 * degree
	xpows := make([][]float64, n)
	for i, x := range xs {
		pw := make([]float64, maxPow+1)
		pw[0] = 1
		for k := 1; k <= maxPow; k++ {
			pw[k] = pw[k-1] * x
		}
		xpows[i] = pw
	}

	// Build normal equations: (degree+1) x (degree+2) augmented.
	aug := make([][]float64, m)
	for i := range aug {
		aug[i] = make([]float64, m+1)
	}
	for i := range m {
		for j := range m {
			for k := range n {
				aug[i][j] += xpows[k][i+j]
			}
		}
		for k := range n {
			aug[i][m] += ys[k] * xpows[k][i]
		}
	}

	// Gaussian elimination with partial pivoting.
	coeffs := solveGauss(aug, m)
	if coeffs == nil {
		return series.XY{}
	}

	// Evaluate at nPoints evenly spaced in original X domain.
	minX, maxX := finiteBoundsX(s.Points)
	step := 0.0
	if nPoints > 1 {
		step = (maxX - minX) / float64(nPoints-1)
	}
	pts := make([]series.Point, nPoints)
	for i := range nPoints {
		x := minX + float64(i)*step
		xn := (x - meanX) / scaleX
		y := coeffs[0]
		xnPow := xn
		for j := 1; j < len(coeffs); j++ {
			y += coeffs[j] * xnPow
			xnPow *= xn
		}
		pts[i] = series.Point{X: x, Y: y}
	}
	return namedXY(
		fmt.Sprintf("%s Poly(%d)", s.Name(), degree), pts)
}

// linearFit computes the least-squares slope and intercept.
func linearFit(pts []series.Point) (slope, intercept float64, ok bool) {
	var sumX, sumY, sumXX, sumXY float64
	var n float64
	for _, p := range pts {
		if !fmath.Finite(p.X) || !fmath.Finite(p.Y) {
			continue
		}
		sumX += p.X
		sumY += p.Y
		sumXX += p.X * p.X
		sumXY += p.X * p.Y
		n++
	}
	if n < 2 {
		return 0, 0, false
	}
	denom := n*sumXX - sumX*sumX
	if denom == 0 {
		return 0, 0, false
	}
	slope = (n*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / n
	return slope, intercept, true
}

// finiteBoundsX returns min/max X among finite points.
func finiteBoundsX(pts []series.Point) (minX, maxX float64) {
	first := true
	for _, p := range pts {
		if !fmath.Finite(p.X) {
			continue
		}
		if first {
			minX, maxX = p.X, p.X
			first = false
			continue
		}
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
	}
	return
}

// solveGauss solves an m x (m+1) augmented matrix via Gaussian
// elimination with partial pivoting. Returns nil if singular.
func solveGauss(aug [][]float64, m int) []float64 {
	for col := range m {
		// Partial pivot.
		maxRow := col
		maxVal := math.Abs(aug[col][col])
		for row := col + 1; row < m; row++ {
			v := math.Abs(aug[row][col])
			if v > maxVal {
				maxVal = v
				maxRow = row
			}
		}
		if maxVal < 1e-15 {
			return nil
		}
		aug[col], aug[maxRow] = aug[maxRow], aug[col]

		// Eliminate.
		for row := col + 1; row < m; row++ {
			factor := aug[row][col] / aug[col][col]
			for k := col; k <= m; k++ {
				aug[row][k] -= factor * aug[col][k]
			}
		}
	}

	// Back-substitute.
	coeffs := make([]float64, m)
	for i := m - 1; i >= 0; i-- {
		coeffs[i] = aug[i][m]
		for j := i + 1; j < m; j++ {
			coeffs[i] -= aug[i][j] * coeffs[j]
		}
		coeffs[i] /= aug[i][i]
	}
	return coeffs
}
