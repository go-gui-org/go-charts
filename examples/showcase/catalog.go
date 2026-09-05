package main

import (
	"github.com/go-gui-org/go-gui/gui"

	"github.com/go-gui-org/go-charts/theme"
)

func catalogPanel(w *gui.Window, entries []DemoEntry) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ShowcaseApp](w)

	return gui.Column(gui.ContainerCfg{
		Width:   catalogWidth,
		Sizing:  gui.FixedFill,
		Color:   t.ColorPanel,
		Padding: gui.NewPadding(12, 12, 12, 12),
		Spacing: gui.SomeF(8),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Chart Catalog",
				TextStyle: t.B3,
			}),
			searchInput(app),
			groupPicker(app),
			themePicker(app),
			line(),
			gui.Column(gui.ContainerCfg{
				ID:            scrollCatalog,
				Scrollable:    true,
				Sizing:        gui.FillFill,
				Padding:       gui.NewPadding(0, t.ScrollbarStyle.Size+4, 0, 0),
				Spacing:       gui.SomeF(2),
				ScrollbarCfgY: &gui.ScrollbarCfg{GapEdge: gui.SomeF(3)},
				Content:       catalogRows(entries, app),
			}),
		},
	})
}

func searchInput(app *ShowcaseApp) gui.View {
	return gui.Input(gui.InputCfg{
		ID:          "showcase-nav-search",
		Sizing:      gui.FillFit,
		Text:        app.NavQuery,
		Placeholder: "Search charts...",
		OnTextChanged: func(text string, ctx gui.EventCtx) {
			gui.State[ShowcaseApp](ctx.Window).NavQuery = text
		},
	})
}

func groupPicker(app *ShowcaseApp) gui.View {
	items := make([]gui.View, len(demoGroups))
	for i, g := range demoGroups {
		items[i] = groupPickerItem(g.Label, g.Key, app)
	}
	return gui.Wrap(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.NoPadding,
		Spacing: gui.SomeF(3),
		Content: items,
	})
}

func groupPickerItem(label, key string, app *ShowcaseApp) gui.View {
	t := gui.CurrentTheme()
	selected := app.SelectedGroup == key
	color := t.ColorBackground
	if selected {
		color = t.ColorActive
	}

	return gui.Button(gui.ButtonCfg{
		ID:    "grp-" + key,
		Color: color,
		// Border only. Not ColorSet{Base: color}, and not Flat(color):
		// Base backs Hover/Click/Focus, so either would pin the three
		// interactive states and stop the button reacting. Color keeps
		// its old meaning — resting fill, states left to the theme.
		Colors:  gui.ColorSet{Border: color},
		Radius:  gui.SomeF(3),
		Padding: gui.NewPadding(3, 6, 3, 6),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      label,
				TextStyle: t.N5,
			}),
		},
		OnClick: func(ctx gui.EventCtx) {
			sa := gui.State[ShowcaseApp](ctx.Window)
			sa.SelectedGroup = key
			sa.NavQuery = ""
			entries := filteredEntries(sa)
			sa.SelectedComponent = preferredComponentForGroup(entries)
			ctx.Window.ScrollVerticalTo(scrollCatalog, 0)
			ctx.Window.ScrollVerticalTo(scrollDetail, 0)
			ctx.Window.ScrollHorizontalTo(scrollDetail, 0)
		},
	})
}

var themeChoices = []struct{ key, label string }{
	{themeDefault, "Default"},
	{themePastel, "Pastel"},
	{themeVivid, "Vivid"},
	{themeHighContrast, "Hi-Con"},
}

func themeLabel(key string) string {
	for _, tc := range themeChoices {
		if tc.key == key {
			return tc.label
		}
	}
	return themeChoices[0].label
}

func themePicker(app *ShowcaseApp) gui.View {
	t := gui.CurrentTheme()
	return gui.Button(gui.ButtonCfg{
		ID:    "theme-cycle",
		Color: t.ColorBackground,
		// Border matches the fill so the button reads as flat at rest;
		// hover and click still come from the theme. See groupPickerItem
		// for why this is not Flat().
		Colors:  gui.ColorSet{Border: t.ColorBackground},
		Radius:  gui.SomeF(3),
		Padding: gui.NewPadding(3, 6, 3, 6),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Theme: " + themeLabel(app.SelectedTheme),
				TextStyle: t.N5,
			}),
		},
		OnClick: func(ctx gui.EventCtx) {
			sa := gui.State[ShowcaseApp](ctx.Window)
			next := 0
			for i, tc := range themeChoices {
				if tc.key == sa.SelectedTheme {
					next = (i + 1) % len(themeChoices)
					break
				}
			}
			sa.SelectedTheme = themeChoices[next].key
			applyChartTheme(sa.SelectedTheme)
			ctx.Window.ClearDrawCanvasCache()
		},
	})
}

func applyChartTheme(key string) {
	theme.SetDefault(nil)
	switch key {
	case themePastel:
		t := theme.Default()
		t.Palette = theme.Pastel()
		theme.SetDefault(t)
	case themeVivid:
		t := theme.Default()
		t.Palette = theme.Vivid()
		theme.SetDefault(t)
	case themeHighContrast:
		theme.SetDefault(theme.HighContrastTheme())
	default:
		theme.SetDefault(theme.Default())
	}
}

func catalogRows(entries []DemoEntry, app *ShowcaseApp) []gui.View {
	t := gui.CurrentTheme()

	if len(entries) == 0 {
		return []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "No matching charts",
				TextStyle: t.N4,
			}),
		}
	}

	rows := make([]gui.View, 0, len(entries)+len(demoGroups)*2)
	for _, group := range demoGroups {
		if group.Key == groupAll {
			continue
		}
		hasGroup := false
		for _, entry := range entries {
			if entry.Group != group.Key {
				continue
			}
			if !hasGroup {
				hasGroup = true
				if len(rows) > 0 {
					rows = append(rows, gui.Row(gui.ContainerCfg{
						Height:  6,
						Sizing:  gui.FillFixed,
						Padding: gui.NoPadding,
					}))
				}
				rows = append(rows, gui.Text(gui.TextCfg{
					Text:      group.Label,
					TextStyle: t.B5,
				}))
			}
			rows = append(rows, catalogRow(entry, app, t))
		}
	}
	return rows
}

func catalogRow(entry DemoEntry, app *ShowcaseApp, t gui.Theme) gui.View {
	selected := app.SelectedComponent == entry.ID
	color := gui.ColorTransparent
	if selected {
		color = t.ColorActive
	}

	return gui.Button(gui.ButtonCfg{
		ID:     "cat-" + entry.ID,
		Sizing: gui.FillFit,
		// Every interactive state is named here, so Base can carry the
		// resting color and the Color shorthand is redundant: Base only
		// backs states that are unset, and none are. Borders stay
		// transparent in both states — the row is delimited by its fill.
		Colors: gui.ColorSet{
			Base:        color,
			Hover:       t.MenubarStyle.ColorSelect,
			Click:       t.ColorActive,
			Focus:       color,
			Border:      gui.ColorTransparent,
			BorderFocus: gui.ColorTransparent,
		},
		Radius:  gui.SomeF(4),
		Padding: gui.NewPadding(3, 6, 3, 6),
		HAlign:  gui.Some(gui.HAlignLeft),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      entry.Label,
				TextStyle: t.N4,
			}),
		},
		OnClick: func(ctx gui.EventCtx) {
			sa := gui.State[ShowcaseApp](ctx.Window)
			sa.SelectedComponent = entry.ID
			ctx.Window.ScrollVerticalTo(scrollDetail, 0)
			ctx.Window.ScrollHorizontalTo(scrollDetail, 0)
		},
	})
}
