package main

import (
	"math"
	"time"

	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoCandlestickBasic(w *gui.Window) gui.View {
	base := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	pts := []series.OHLC{
		{Open: 185.50, High: 188.20, Low: 184.10, Close: 187.30},
		{Open: 187.30, High: 190.50, Low: 186.00, Close: 189.80},
		{Open: 189.80, High: 191.00, Low: 186.50, Close: 187.10},
		{Open: 187.10, High: 189.40, Low: 185.20, Close: 188.60},
		{Open: 188.60, High: 193.20, Low: 187.90, Close: 192.50},
		{Open: 192.50, High: 194.80, Low: 190.10, Close: 191.30},
		{Open: 191.30, High: 192.00, Low: 187.50, Close: 188.00},
		{Open: 188.00, High: 190.30, Low: 185.80, Close: 189.70},
		{Open: 189.70, High: 195.10, Low: 188.90, Close: 194.20},
		{Open: 194.20, High: 196.40, Low: 192.80, Close: 193.50},
		{Open: 193.50, High: 197.20, Low: 192.10, Close: 196.80},
		{Open: 196.80, High: 198.50, Low: 194.30, Close: 195.10},
		{Open: 195.10, High: 196.00, Low: 191.50, Close: 192.40},
		{Open: 192.40, High: 194.70, Low: 190.20, Close: 193.90},
		{Open: 193.90, High: 199.30, Low: 193.10, Close: 198.60},
		{Open: 198.60, High: 200.10, Low: 196.40, Close: 197.20},
		{Open: 197.20, High: 198.80, Low: 193.70, Close: 194.50},
		{Open: 194.50, High: 196.20, Low: 192.00, Close: 195.80},
		{Open: 195.80, High: 201.50, Low: 195.30, Close: 200.40},
		{Open: 200.40, High: 203.20, Low: 199.10, Close: 202.70},
	}
	for i := range pts {
		pts[i].Time = base.AddDate(0, 0, i)
	}

	return demoWithCode(w, "candlestick-basic",
		chart.Candlestick(chart.CandlestickCfg{
			BaseCfg: chart.BaseCfg{
				ID:             "candlestick-basic",
				Title:          "AAPL — January 2024",
				Sizing:         gui.FillFixed,
				Height:         350,
				XTickRotation:  math.Pi / 4,
				LegendPosition: &posBottom,
			},
			XTimeFormat: "01/02",
			Series: []series.OHLCSeries{
				series.NewOHLC(series.OHLCCfg{
					Name:      "AAPL",
					ColorUp:   gui.Hex(0x26a69a),
					ColorDown: gui.Hex(0xef5350),
					Points:    pts,
				}),
			},
		}),
		`chart.Candlestick(chart.CandlestickCfg{
    BaseCfg: chart.BaseCfg{
        Title:         "AAPL — January 2024",
        XTickRotation: math.Pi / 4,
    },
    XTimeFormat: "01/02",
    Series: []series.OHLCSeries{
        series.NewOHLC(series.OHLCCfg{
            Name:      "AAPL",
            ColorUp:   gui.Hex(0x26a69a),
            ColorDown: gui.Hex(0xef5350),
            Points:    pts,
        }),
    },
})`)
}
