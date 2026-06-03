package series

import (
	"fmt"
	"math"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-gui/gui"
)

// Grid is a dense 2D matrix of values with row and column labels,
// used by heatmap charts. Values are stored row-major:
// values[row][col].
type Grid struct {
	name   string
	rows   []string
	cols   []string
	values [][]float64
}

// GridCfg configures a Grid series.
type GridCfg struct {
	Name   string
	Rows   []string
	Cols   []string
	Values [][]float64
}

// NewGrid creates a Grid series. Returns an error if Values
// dimensions do not match Rows/Cols lengths.
func NewGrid(cfg GridCfg) (Grid, error) {
	if len(cfg.Values) != len(cfg.Rows) {
		return Grid{}, fmt.Errorf(
			"series.NewGrid: len(Values)=%d != len(Rows)=%d",
			len(cfg.Values), len(cfg.Rows))
	}
	nc := len(cfg.Cols)
	for i, row := range cfg.Values {
		if len(row) != nc {
			return Grid{}, fmt.Errorf(
				"series.NewGrid: len(Values[%d])=%d != len(Cols)=%d",
				i, len(row), nc)
		}
	}
	return Grid{
		name:   cfg.Name,
		rows:   cfg.Rows,
		cols:   cfg.Cols,
		values: cfg.Values,
	}, nil
}

// GridFromFunc creates a Grid by calling fn(row, col) for each
// cell.
func GridFromFunc(
	name string, rows, cols []string,
	fn func(r, c int) float64,
) Grid {
	values := make([][]float64, len(rows))
	for r := range rows {
		values[r] = make([]float64, len(cols))
		for c := range cols {
			values[r][c] = fn(r, c)
		}
	}
	return Grid{
		name:   name,
		rows:   rows,
		cols:   cols,
		values: values,
	}
}

// Name implements Series.
func (g Grid) Name() string { return g.name }

// Len implements Series. Returns rows * cols.
func (g Grid) Len() int { return len(g.rows) * len(g.cols) }

// Color implements Series. Returns zero; heatmaps use a color
// scale instead of a series color.
func (g Grid) Color() gui.Color { return gui.Color{} }

// Rows returns the row labels.
func (g Grid) Rows() []string { return g.rows }

// Cols returns the column labels.
func (g Grid) Cols() []string { return g.cols }

// NumRows returns the number of rows.
func (g Grid) NumRows() int { return len(g.rows) }

// NumCols returns the number of columns.
func (g Grid) NumCols() int { return len(g.cols) }

// At returns the value at (row, col). Panics if out of range.
func (g Grid) At(row, col int) float64 { return g.values[row][col] }

// Bounds returns the min and max finite values across all cells.
// NaN and Inf values are skipped. If no finite values exist, both
// returned values are zero.
func (g Grid) Bounds() (vMin, vMax float64) {
	first := true
	for _, row := range g.values {
		for _, v := range row {
			if !fmath.Finite(v) {
				continue
			}
			if first {
				vMin, vMax = v, v
				first = false
				continue
			}
			vMin = min(vMin, v)
			vMax = max(vMax, v)
		}
	}
	return
}

// String implements fmt.Stringer.
func (g Grid) String() string {
	return fmt.Sprintf("Grid{%q, %dx%d}",
		g.name, len(g.rows), len(g.cols))
}

// IsNaN reports whether the value at (row, col) is NaN.
func (g Grid) IsNaN(row, col int) bool {
	return math.IsNaN(g.values[row][col])
}
