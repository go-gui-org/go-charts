package main

import (
	"github.com/go-gui-org/go-gui/gui"
	"github.com/go-gui-org/go-gui/gui/backend"
)

const (
	scrollCatalog uint32 = iota + 1
	scrollDetail
)

const catalogWidth float32 = 250

func runShowcase() {
	gui.SetTheme(gui.ThemeDark.WithBorders(true))

	w := gui.NewWindow(gui.WindowCfg{
		State:  newShowcaseApp(),
		Title:  "Charts Showcase",
		Width:  800,
		Height: 768,
		OnInit: func(w *gui.Window) {
			w.UpdateView(mainView)
		},
	})

	backend.Run(w)
}

func mainView(w *gui.Window) gui.View {
	ww, wh := w.WindowSize()
	app := gui.State[ShowcaseApp](w)
	entries := filteredEntries(app)
	fixupSelection(app, entries, w)

	return gui.Row(gui.ContainerCfg{
		Width:   float32(ww),
		Height:  float32(wh),
		Sizing:  gui.FixedFixed,
		Padding: gui.NoPadding,
		Spacing: gui.NoSpacing,
		Content: []gui.View{catalogPanel(w, entries), detailPanel(w, entries)},
	})
}

func fixupSelection(app *ShowcaseApp, entries []DemoEntry, w *gui.Window) {
	switch {
	case len(entries) == 0:
		app.SelectedComponent = ""
		w.ScrollVerticalTo(scrollDetail, 0)
		w.ScrollHorizontalTo(scrollDetail, 0)
	case !hasEntry(entries, app.SelectedComponent):
		app.SelectedComponent = preferredComponentForGroup(entries)
		w.ScrollVerticalTo(scrollDetail, 0)
		w.ScrollHorizontalTo(scrollDetail, 0)
	}
}
