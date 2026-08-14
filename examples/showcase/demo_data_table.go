package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoDataTable(w *gui.Window) gui.View {
	lineData := styleSeries()

	barData := []series.Category{
		series.NewCategory(series.CategoryCfg{
			Name: "Q1",
			Values: []series.CategoryValue{
				{Label: "North", Value: 45},
				{Label: "South", Value: 32},
				{Label: "East", Value: 58},
				{Label: "West", Value: 41},
			},
		}),
		series.NewCategory(series.CategoryCfg{
			Name: "Q2",
			Values: []series.CategoryValue{
				{Label: "North", Value: 52},
				{Label: "South", Value: 38},
				{Label: "East", Value: 49},
				{Label: "West", Value: 55},
			},
		}),
	}

	pieSlices := []chart.PieSlice{
		{Label: "Desktop", Value: 58},
		{Label: "Mobile", Value: 32},
		{Label: "Tablet", Value: 10},
	}

	lineChart := chart.Line(chart.LineCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "dt-line-chart",
			Title:          "Product Sales",
			Sizing:         gui.FillFixed,
			Height:         200,
			LegendPosition: &posBottom,
		},
		ShowMarkers: true,
		Series:      lineData,
	})
	barChart := chart.Bar(chart.BarCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "dt-bar-chart",
			Title:          "Regional Sales",
			Sizing:         gui.FillFixed,
			Height:         200,
			LegendPosition: &posBottom,
		},
		Series: barData,
	})
	pieChart := chart.Pie(chart.PieCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "dt-pie-chart",
			Title:          "Device Share",
			Sizing:         gui.FillFixed,
			Height:         200,
			LegendPosition: &posBottom,
		},
		ShowPercent: true,
		Slices:      pieSlices,
	})

	// The ShowDataTable variants resolve to dataTableXY, which is not
	// a Drawer: they get no export buttons and are not exportable.
	content := []gui.View{
		lineChart,
		chart.Line(chart.LineCfg{
			BaseCfg: chart.BaseCfg{
				ID:            "dt-line-table",
				Title:         "Product Sales",
				Sizing:        gui.FillFixed,
				Height:        200,
				ShowDataTable: true,
			},
			Series: lineData,
		}),
		barChart,
		chart.Bar(chart.BarCfg{
			BaseCfg: chart.BaseCfg{
				ID:            "dt-bar-table",
				Title:         "Regional Sales",
				Sizing:        gui.FillFixed,
				Height:        200,
				ShowDataTable: true,
			},
			Series: barData,
		}),
		pieChart,
		chart.Pie(chart.PieCfg{
			BaseCfg: chart.BaseCfg{
				ID:            "dt-pie-table",
				Title:         "Device Share",
				Sizing:        gui.FillFixed,
				Height:        200,
				ShowDataTable: true,
			},
			Slices: pieSlices,
		}),
	}

	return demoWithCodeCharts(w, "style-data-table", content,
		[]gui.View{lineChart, barChart, pieChart}, `// Set ShowDataTable: true to render as a table.
chart.Line(chart.LineCfg{
    BaseCfg: chart.BaseCfg{
        Title:         "Product Sales",
        ShowDataTable: true,
    },
    Series: data,
})`)
}
