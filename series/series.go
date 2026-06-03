// Package series provides data series types for charts.
package series

import "github.com/go-gui-org/go-gui/gui"

// Series is the interface for all data series.
type Series interface {
	// Name returns the series display name.
	Name() string

	// Len returns the number of data points.
	Len() int

	// Color returns the series color.
	Color() gui.Color
}
