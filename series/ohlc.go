package series

import (
	"fmt"
	"time"

	"github.com/go-gui-org/go-gui/gui"
)

// OHLC represents a single open/high/low/close candlestick.
type OHLC struct {
	Time                   time.Time
	Open, High, Low, Close float64
	Volume                 float64
}

// OHLCSeries is a series of OHLC data for candlestick charts.
type OHLCSeries struct {
	name      string
	colorUp   gui.Color
	colorDown gui.Color
	Points    []OHLC
}

// OHLCCfg configures an OHLC series.
type OHLCCfg struct {
	Name      string
	ColorUp   gui.Color
	ColorDown gui.Color
	Points    []OHLC
}

// NewOHLC creates a new OHLC data series.
func NewOHLC(cfg OHLCCfg) OHLCSeries {
	return OHLCSeries{
		name:      cfg.Name,
		colorUp:   cfg.ColorUp,
		colorDown: cfg.ColorDown,
		Points:    cfg.Points,
	}
}

// Name implements Series.
func (s OHLCSeries) Name() string { return s.name }

// Len implements Series.
func (s OHLCSeries) Len() int { return len(s.Points) }

// Color implements Series. Returns ColorUp; OHLC renderers
// should use ColorUp/ColorDown directly.
func (s OHLCSeries) Color() gui.Color { return s.colorUp }

// ColorUp returns the color for upward (close >= open) candles.
func (s OHLCSeries) ColorUp() gui.Color { return s.colorUp }

// ColorDown returns the color for downward (close < open) candles.
func (s OHLCSeries) ColorDown() gui.Color { return s.colorDown }

// String implements fmt.Stringer.
func (s OHLCSeries) String() string {
	return fmt.Sprintf("OHLC{%q, %d candles}",
		s.name, len(s.Points))
}
