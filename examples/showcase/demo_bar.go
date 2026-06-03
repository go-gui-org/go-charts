package main

import (
	"github.com/go-gui-org/go-charts/chart"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func demoBarBasic(w *gui.Window) gui.View {
	return demoWithCode(w, "bar-basic", chart.Bar(chart.BarCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "bar-basic",
			Title:          "Sales by Region",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Name:  "Q1",
				Color: gui.Hex(0x4E79A7),
				Values: []series.CategoryValue{
					{Label: "North", Value: 45},
					{Label: "South", Value: 32},
					{Label: "East", Value: 58},
					{Label: "West", Value: 41},
				},
			}),
			series.NewCategory(series.CategoryCfg{
				Name:  "Q2",
				Color: gui.Hex(0xF28E2B),
				Values: []series.CategoryValue{
					{Label: "North", Value: 52},
					{Label: "South", Value: 38},
					{Label: "East", Value: 49},
					{Label: "West", Value: 55},
				},
			}),
		},
	}), `chart.Bar(chart.BarCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Sales by Region",
    },
    Series: []series.Category{
        series.NewCategory(series.CategoryCfg{
            Name:  "Q1",
            Color: gui.Hex(0x4E79A7),
            Values: []series.CategoryValue{
                {Label: "North", Value: 45},
                {Label: "South", Value: 32},
                {Label: "East", Value: 58},
                {Label: "West", Value: 41},
            },
        }),
        series.NewCategory(series.CategoryCfg{
            Name:  "Q2",
            Color: gui.Hex(0xF28E2B),
            Values: []series.CategoryValue{ ... },
        }),
    },
})`)
}

func demoBarSingle(w *gui.Window) gui.View {
	return demoWithCode(w, "bar-single", chart.Bar(chart.BarCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "bar-single",
			Title:          "Monthly Rainfall (mm)",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Name:  "2025",
				Color: gui.Hex(0x76B7B2),
				Values: []series.CategoryValue{
					{Label: "Jan", Value: 78},
					{Label: "Feb", Value: 63},
					{Label: "Mar", Value: 85},
					{Label: "Apr", Value: 92},
					{Label: "May", Value: 110},
					{Label: "Jun", Value: 72},
				},
			}),
		},
	}), `chart.Bar(chart.BarCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Monthly Rainfall (mm)",
    },
    Series: []series.Category{
        series.NewCategory(series.CategoryCfg{
            Name:  "2025",
            Color: gui.Hex(0x76B7B2),
            Values: []series.CategoryValue{
                {Label: "Jan", Value: 78},
                {Label: "Feb", Value: 63},
                {Label: "Mar", Value: 85},
                {Label: "Apr", Value: 92},
                {Label: "May", Value: 110},
                {Label: "Jun", Value: 72},
            },
        }),
    },
})`)
}

func demoBarWide(w *gui.Window) gui.View {
	return demoWithCode(w, "bar-wide", chart.Bar(chart.BarCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "bar-wide",
			Title:          "Department Headcount",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		BarWidth: 40,
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Name:  "Employees",
				Color: gui.Hex(0xB07AA1),
				Values: []series.CategoryValue{
					{Label: "Eng", Value: 120},
					{Label: "Sales", Value: 85},
					{Label: "Mktg", Value: 42},
					{Label: "Ops", Value: 67},
					{Label: "HR", Value: 28},
				},
			}),
		},
	}), `chart.Bar(chart.BarCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Department Headcount",
    },
    BarWidth: 40,
    Series: []series.Category{
        series.NewCategory(series.CategoryCfg{
            Name:  "Employees",
            Color: gui.Hex(0xB07AA1),
            Values: []series.CategoryValue{
                {Label: "Eng", Value: 120},
                {Label: "Sales", Value: 85},
                {Label: "Mktg", Value: 42},
                {Label: "Ops", Value: 67},
                {Label: "HR", Value: 28},
            },
        }),
    },
})`)
}

func demoBarHorizontal(w *gui.Window) gui.View {
	return demoWithCode(w, "bar-horizontal", chart.Bar(chart.BarCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "bar-horizontal",
			Title:          "Survey Results",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Horizontal: true,
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Name:  "Agree",
				Color: gui.Hex(0x4E79A7),
				Values: []series.CategoryValue{
					{Label: "Q1", Value: 72},
					{Label: "Q2", Value: 58},
					{Label: "Q3", Value: 65},
					{Label: "Q4", Value: 81},
				},
			}),
			series.NewCategory(series.CategoryCfg{
				Name:  "Disagree",
				Color: gui.Hex(0xE15759),
				Values: []series.CategoryValue{
					{Label: "Q1", Value: 28},
					{Label: "Q2", Value: 42},
					{Label: "Q3", Value: 35},
					{Label: "Q4", Value: 19},
				},
			}),
		},
	}), `chart.Bar(chart.BarCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Survey Results",
    },
    Horizontal: true,
    Series: []series.Category{
        series.NewCategory(series.CategoryCfg{
            Name:  "Agree",
            Color: gui.Hex(0x4E79A7),
            Values: []series.CategoryValue{
                {Label: "Q1", Value: 72},
                {Label: "Q2", Value: 58},
                {Label: "Q3", Value: 65},
                {Label: "Q4", Value: 81},
            },
        }),
        series.NewCategory(series.CategoryCfg{
            Name:  "Disagree",
            Color: gui.Hex(0xE15759),
            Values: []series.CategoryValue{ ... },
        }),
    },
})`)
}

func demoBarStacked(w *gui.Window) gui.View {
	return demoWithCode(w, "bar-stacked", chart.Bar(chart.BarCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "bar-stacked",
			Title:          "Traffic by Channel",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Stacked: true,
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Name:  "Organic",
				Color: gui.Hex(0x4E79A7),
				Values: []series.CategoryValue{
					{Label: "Jan", Value: 420},
					{Label: "Feb", Value: 390},
					{Label: "Mar", Value: 450},
					{Label: "Apr", Value: 510},
					{Label: "May", Value: 480},
				},
			}),
			series.NewCategory(series.CategoryCfg{
				Name:  "Paid",
				Color: gui.Hex(0xF28E2B),
				Values: []series.CategoryValue{
					{Label: "Jan", Value: 180},
					{Label: "Feb", Value: 210},
					{Label: "Mar", Value: 195},
					{Label: "Apr", Value: 230},
					{Label: "May", Value: 260},
				},
			}),
			series.NewCategory(series.CategoryCfg{
				Name:  "Referral",
				Color: gui.Hex(0x59A14F),
				Values: []series.CategoryValue{
					{Label: "Jan", Value: 95},
					{Label: "Feb", Value: 88},
					{Label: "Mar", Value: 110},
					{Label: "Apr", Value: 102},
					{Label: "May", Value: 120},
				},
			}),
		},
	}), `chart.Bar(chart.BarCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Traffic by Channel",
    },
    Stacked: true,
    Series: []series.Category{
        series.NewCategory(series.CategoryCfg{
            Name:  "Organic",
            Color: gui.Hex(0x4E79A7),
            Values: []series.CategoryValue{
                {Label: "Jan", Value: 420}, ...
            },
        }),
        series.NewCategory(series.CategoryCfg{
            Name:  "Paid",
            Color: gui.Hex(0xF28E2B),
            Values: []series.CategoryValue{ ... },
        }),
        series.NewCategory(series.CategoryCfg{
            Name:  "Referral",
            Color: gui.Hex(0x59A14F),
            Values: []series.CategoryValue{ ... },
        }),
    },
})`)
}

func demoBarRounded(w *gui.Window) gui.View {
	return demoWithCode(w, "bar-rounded", chart.Bar(chart.BarCfg{
		BaseCfg: chart.BaseCfg{
			ID:             "bar-rounded",
			Title:          "Product Revenue",
			Sizing:         gui.FillFixed,
			Height:         350,
			LegendPosition: &posBottom,
		},
		Radius: 4,
		Series: []series.Category{
			series.NewCategory(series.CategoryCfg{
				Name:  "Online",
				Color: gui.Hex(0x59A14F),
				Values: []series.CategoryValue{
					{Label: "Widgets", Value: 340},
					{Label: "Gadgets", Value: 280},
					{Label: "Gizmos", Value: 195},
					{Label: "Doohickeys", Value: 150},
				},
			}),
			series.NewCategory(series.CategoryCfg{
				Name:  "Retail",
				Color: gui.Hex(0xEDC948),
				Values: []series.CategoryValue{
					{Label: "Widgets", Value: 210},
					{Label: "Gadgets", Value: 320},
					{Label: "Gizmos", Value: 175},
					{Label: "Doohickeys", Value: 230},
				},
			}),
		},
	}), `chart.Bar(chart.BarCfg{
    BaseCfg: chart.BaseCfg{
        Title: "Product Revenue",
    },
    Radius: 4,
    Series: []series.Category{
        series.NewCategory(series.CategoryCfg{
            Name:  "Online",
            Color: gui.Hex(0x59A14F),
            Values: []series.CategoryValue{
                {Label: "Widgets", Value: 340},
                {Label: "Gadgets", Value: 280},
                {Label: "Gizmos", Value: 195},
                {Label: "Doohickeys", Value: 150},
            },
        }),
        series.NewCategory(series.CategoryCfg{
            Name:  "Retail",
            Color: gui.Hex(0xEDC948),
            Values: []series.CategoryValue{ ... },
        }),
    },
})`)
}
