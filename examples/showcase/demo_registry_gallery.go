//go:build gallery

package main

import "github.com/go-gui-org/go-gui/gui"

// demoCharts records, per demo ID, the chart.Drawer views the demo
// built via demoWithCode/demoWithCodeCharts. Demos re-run every frame,
// so the registry is overwritten per frame; renderEntry reads it right
// after the demo call returns.
var demoCharts = map[string][]gui.View{}

// registerCharts stores the charts a demo built, keyed by demo id.
//
// go-gui's View interface no longer allows walking a View tree, so the
// gallery generator cannot rediscover charts by inspection; the demos
// hand them over explicitly and this registry is where the gallery
// reads them back. The app build (demo_registry.go) is a no-op, so the
// running app pays nothing per frame for this gallery-only state.
func registerCharts(id string, charts []gui.View) {
	demoCharts[id] = charts
}
