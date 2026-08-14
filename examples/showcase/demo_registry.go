//go:build !gallery

package main

import "github.com/go-gui-org/go-gui/gui"

// registerCharts is a no-op outside the gallery build: the chart
// registry is generator-only state, so the running app pays nothing
// per frame for it.
func registerCharts(id string, charts []gui.View) {}
