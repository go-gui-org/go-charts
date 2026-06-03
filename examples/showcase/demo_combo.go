package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoComboBasic(w *gui.Window) gui.View {
	return demoWithCode(w, "combo-basic", chart.Combo(chart.ComboCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "combo-basic",
			Title:          "Revenue & Trend",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Series: []chart.ComboSeries{
			{
				Category: series.NewCategory(series.CategoryCfg{
					Name: "Revenue",
					Values: []series.CategoryValue{
						{Label: "Jan", Value: 420},
						{Label: "Feb", Value: 380},
						{Label: "Mar", Value: 510},
						{Label: "Apr", Value: 470},
						{Label: "May", Value: 560},
						{Label: "Jun", Value: 620},
					},
				}),
				Type: chart.ComboBar,
			},
			{
				Category: series.NewCategory(series.CategoryCfg{
					Name: "Trend",
					Values: []series.CategoryValue{
						{Label: "Jan", Value: 400},
						{Label: "Feb", Value: 420},
						{Label: "Mar", Value: 460},
						{Label: "Apr", Value: 490},
						{Label: "May", Value: 530},
						{Label: "Jun", Value: 580},
					},
				}),
				Type: chart.ComboLine,
			},
		},
		ShowMarkers: true,
	}), `chart.Combo(chart.ComboCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Revenue & Trend",
    },
    Series: []chart.ComboSeries{
        {
            Category: series.NewCategory(series.CategoryCfg{
                Name:   "Revenue",
                Values: []series.CategoryValue{
                    {Label: "Jan", Value: 420},
                    {Label: "Feb", Value: 380},
                    ...
                },
            }),
            Type: chart.ComboBar,
        },
        {
            Category: series.NewCategory(series.CategoryCfg{
                Name:   "Trend",
                Values: []series.CategoryValue{
                    {Label: "Jan", Value: 400},
                    ...
                },
            }),
            Type: chart.ComboLine,
        },
    },
    ShowMarkers: true,
})`)
}

func demoComboMulti(w *gui.Window) gui.View {
	return demoWithCode(w, "combo-multi", chart.Combo(chart.ComboCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "combo-multi",
			Title:          "Sales & Growth",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Series: []chart.ComboSeries{
			{
				Category: series.NewCategory(series.CategoryCfg{
					Name: "Online Sales",
					Values: []series.CategoryValue{
						{Label: "Q1", Value: 240},
						{Label: "Q2", Value: 310},
						{Label: "Q3", Value: 280},
						{Label: "Q4", Value: 390},
					},
				}),
				Type: chart.ComboBar,
			},
			{
				Category: series.NewCategory(series.CategoryCfg{
					Name: "In-Store Sales",
					Values: []series.CategoryValue{
						{Label: "Q1", Value: 180},
						{Label: "Q2", Value: 200},
						{Label: "Q3", Value: 220},
						{Label: "Q4", Value: 210},
					},
				}),
				Type: chart.ComboBar,
			},
			{
				Category: series.NewCategory(series.CategoryCfg{
					Name: "Online Growth",
					Values: []series.CategoryValue{
						{Label: "Q1", Value: 220},
						{Label: "Q2", Value: 270},
						{Label: "Q3", Value: 260},
						{Label: "Q4", Value: 350},
					},
				}),
				Type: chart.ComboLine,
			},
			{
				Category: series.NewCategory(series.CategoryCfg{
					Name: "In-Store Growth",
					Values: []series.CategoryValue{
						{Label: "Q1", Value: 170},
						{Label: "Q2", Value: 190},
						{Label: "Q3", Value: 210},
						{Label: "Q4", Value: 200},
					},
				}),
				Type: chart.ComboLine,
			},
		},
		ShowMarkers: true,
		Radius:      3,
	}), `chart.Combo(chart.ComboCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Sales & Growth",
    },
    Series: []chart.ComboSeries{
        {Category: series.NewCategory(series.CategoryCfg{
            Name: "Online Sales",
            Values: []series.CategoryValue{{Label: "Q1", Value: 240}, ...},
        }), Type: chart.ComboBar},
        {Category: series.NewCategory(series.CategoryCfg{
            Name: "In-Store Sales",
            Values: []series.CategoryValue{{Label: "Q1", Value: 180}, ...},
        }), Type: chart.ComboBar},
        {Category: series.NewCategory(series.CategoryCfg{
            Name: "Online Growth",
            Values: []series.CategoryValue{{Label: "Q1", Value: 220}, ...},
        }), Type: chart.ComboLine},
        {Category: series.NewCategory(series.CategoryCfg{
            Name: "In-Store Growth",
            Values: []series.CategoryValue{{Label: "Q1", Value: 170}, ...},
        }), Type: chart.ComboLine},
    },
    ShowMarkers: true,
    Radius:      3,
})`)
}
