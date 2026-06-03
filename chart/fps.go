package chart

import (
	"time"

	"github.com/go-gui-org/go-gui/gui"
)

// fpsState tracks frame timing for adaptive rendering.
// Stored as a global singleton in StateMap.
type fpsState struct {
	LastDraw   time.Time
	AvgFrameMs float32 // exponential moving average
	Version    uint64
}

const (
	nsChartFPS  = "chart-fps"
	capChartFPS = 1
	fpsKey      = "__global__"
)

// chartFPSMap returns the persistent FPS state map.
func chartFPSMap(w *gui.Window) *gui.BoundedMap[string, fpsState] {
	return gui.StateMap[string, fpsState](w, nsChartFPS, capChartFPS)
}

// updateFPSTracker records a draw and updates the frame time
// moving average. Call at the start of each chart draw.
func updateFPSTracker(w *gui.Window) {
	if w == nil {
		return
	}
	sm := chartFPSMap(w)
	fs, _ := sm.Get(fpsKey)
	now := time.Now()
	if !fs.LastDraw.IsZero() {
		dt := float32(now.Sub(fs.LastDraw).Milliseconds())
		if fs.AvgFrameMs == 0 {
			fs.AvgFrameMs = dt
		} else {
			fs.AvgFrameMs = fpsEMAAlpha*dt +
				(1-fpsEMAAlpha)*fs.AvgFrameMs
		}
	}
	fs.LastDraw = now
	fs.Version++
	sm.Set(fpsKey, fs)
}

// shouldReduceDetail returns true if recent frame times exceed
// the FPS budget, indicating that draw methods should skip
// expensive operations (markers, grid lines, tooltips).
func shouldReduceDetail(w *gui.Window) bool {
	if w == nil {
		return false
	}
	sm := gui.StateMapRead[string, fpsState](w, nsChartFPS)
	if sm == nil {
		return false
	}
	fs, ok := sm.Get(fpsKey)
	if !ok {
		return false
	}
	return fs.AvgFrameMs > DefaultFPSBudgetMs
}

// avgFrameMs returns the current average frame time in
// milliseconds. Returns 0 if no data.
func avgFrameMs(w *gui.Window) float32 {
	if w == nil {
		return 0
	}
	sm := gui.StateMapRead[string, fpsState](w, nsChartFPS)
	if sm == nil {
		return 0
	}
	fs, ok := sm.Get(fpsKey)
	if !ok {
		return 0
	}
	return fs.AvgFrameMs
}
