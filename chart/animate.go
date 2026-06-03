package chart

import (
	"time"

	"github.com/go-gui-org/go-charts/axis"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

// animState persists entry/transition animation progress across
// frames via gui.StateMap. Keyed by chart ID.
type animState struct {
	Progress float32 // 0→1
	Started  bool    // animation queued/running
	Done     bool    // animation completed
	Version  uint64
}

// transitionState persists data-transition animation progress
// across frames via gui.StateMap. Keyed by chart ID.
type transitionState struct {
	Progress   float32 // 0→1
	Active     bool
	LastCfgVer uint64 // last BaseCfg.Version seen
	Version    uint64
}

// transitionDataState caches old data-space values for
// interpolation during data transitions. Keyed by chart ID.
type transitionDataState struct {
	// OldYValues stores old Y values per series (flat slices).
	OldYValues [][]float64
	// OldBounds stores the axis domain at the start of the
	// transition so scales can animate alongside data.
	OldBounds [4]float64 // minX, maxX, minY, maxY
	HasBounds bool
}

const (
	nsChartAnim            = "chart-anim"
	capChartAnim           = 64
	nsChartTransition      = "chart-transition"
	capChartTransition     = 64
	nsChartTransitionData  = "chart-tdata"
	capChartTransitionData = 64

	// DefaultAnimDuration is the entry animation duration.
	DefaultAnimDuration = 500 * time.Millisecond

	// DefaultTransitionDuration is the data transition duration.
	DefaultTransitionDuration = 300 * time.Millisecond

	animEntryPrefix      = "chart-entry-"
	animTransitionPrefix = "chart-trans-"
)

// chartAnimMap returns the persistent animation state map.
func chartAnimMap(w *gui.Window) *gui.BoundedMap[string, animState] {
	return gui.StateMap[string, animState](w, nsChartAnim, capChartAnim)
}

// chartTransitionMap returns the persistent transition state map.
func chartTransitionMap(w *gui.Window) *gui.BoundedMap[string, transitionState] {
	return gui.StateMap[string, transitionState](
		w, nsChartTransition, capChartTransition)
}

// loadAnimVersion returns the animation version for inclusion in
// the DrawCanvasCfg.Version sum.
func loadAnimVersion(w *gui.Window, id string) uint64 {
	if w == nil || id == "" {
		return 0
	}
	sm := gui.StateMapRead[string, animState](w, nsChartAnim)
	if sm == nil {
		return 0
	}
	as, ok := sm.Get(id)
	if !ok {
		return 0
	}
	return as.Version
}

// animProgress returns the current entry animation progress for
// the given chart. Returns 1.0 when no animation is active
// (charts render fully by default).
func animProgress(w *gui.Window, id string) float32 {
	if w == nil || id == "" {
		return 1
	}
	sm := gui.StateMapRead[string, animState](w, nsChartAnim)
	if sm == nil {
		return 1
	}
	as, ok := sm.Get(id)
	if !ok {
		return 1
	}
	if as.Done {
		return 1
	}
	return as.Progress
}

// startEntryAnimation begins a 0→1 tween for the chart entry
// animation. Idempotent: does nothing if already started.
//
// Called from GenerateLayout which runs under w.mu, so the
// actual AnimationAdd is deferred via QueueCommand to avoid
// deadlock with w.mu.
func startEntryAnimation(w *gui.Window, id string, dur time.Duration) {
	if w == nil || id == "" {
		return
	}
	sm := chartAnimMap(w)
	as, _ := sm.Get(id)
	if as.Done || as.Started {
		return
	}
	// Mark started and set progress=0 so the first draw
	// renders the chart at zero progress (not fully visible).
	as.Started = true
	as.Progress = 0
	as.Version++
	sm.Set(id, as)

	if dur <= 0 {
		dur = DefaultAnimDuration
	}
	animID := animEntryPrefix + id
	// Defer AnimationAdd: GenerateLayout holds w.mu and
	// AnimationAdd also acquires w.mu. QueueCommand runs
	// at frame start outside w.mu.
	w.QueueCommand(func(w *gui.Window) {
		tw := &gui.TweenAnimation{
			AnimID:   animID,
			Duration: dur,
			Easing:   gui.EaseOutCubic,
			From:     0,
			To:       1,
			OnValue: func(v float32, w *gui.Window) {
				sm := chartAnimMap(w)
				as, _ := sm.Get(id)
				as.Progress = v
				as.Version++
				sm.Set(id, as)
			},
			OnDone: func(w *gui.Window) {
				sm := chartAnimMap(w)
				as, _ := sm.Get(id)
				as.Progress = 1
				as.Done = true
				as.Version++
				sm.Set(id, as)
			},
		}
		w.AnimationAdd(tw)
	})
}

// ResetEntryAnimation clears the done/started flags so the entry
// animation can replay. Called from event handlers (outside w.mu).
func ResetEntryAnimation(w *gui.Window, id string) {
	if w == nil || id == "" {
		return
	}
	w.AnimationRemove(animEntryPrefix + id)
	sm := chartAnimMap(w)
	sm.Set(id, animState{})
}

// loadTransitionVersion returns the transition version for
// inclusion in the DrawCanvasCfg.Version sum.
func loadTransitionVersion(w *gui.Window, id string) uint64 {
	if w == nil || id == "" {
		return 0
	}
	sm := gui.StateMapRead[string, transitionState](w, nsChartTransition)
	if sm == nil {
		return 0
	}
	ts, ok := sm.Get(id)
	if !ok {
		return 0
	}
	return ts.Version
}

// transitionProgress returns the current transition progress.
// Returns 1.0 when no transition is active.
func transitionProgress(w *gui.Window, id string) float32 {
	if w == nil || id == "" {
		return 1
	}
	sm := gui.StateMapRead[string, transitionState](w, nsChartTransition)
	if sm == nil {
		return 1
	}
	ts, ok := sm.Get(id)
	if !ok || !ts.Active {
		return 1
	}
	return ts.Progress
}

// transitionActive reports whether a data transition is in
// progress for the given chart.
func transitionActive(w *gui.Window, id string) bool {
	if w == nil || id == "" {
		return false
	}
	sm := gui.StateMapRead[string, transitionState](w, nsChartTransition)
	if sm == nil {
		return false
	}
	ts, ok := sm.Get(id)
	return ok && ts.Active
}

// chartTransitionDataMap returns the transition data cache map.
func chartTransitionDataMap(
	w *gui.Window,
) *gui.BoundedMap[string, transitionDataState] {
	return gui.StateMap[string, transitionDataState](
		w, nsChartTransitionData, capChartTransitionData)
}

// saveTransitionData stores old Y values for interpolation.
func saveTransitionData(
	w *gui.Window, id string, old [][]float64,
) {
	if w == nil || id == "" {
		return
	}
	sm := chartTransitionDataMap(w)
	td, _ := sm.Get(id)
	td.OldYValues = old
	sm.Set(id, td)
}

// saveTransitionBounds stores old axis domain for animated
// scale transitions.
func saveTransitionBounds(
	w *gui.Window, id string,
	minX, maxX, minY, maxY float64,
) {
	if w == nil || id == "" {
		return
	}
	sm := chartTransitionDataMap(w)
	td, _ := sm.Get(id)
	td.OldBounds = [4]float64{minX, maxX, minY, maxY}
	td.HasBounds = true
	sm.Set(id, td)
}

// loadTransitionData returns cached old Y values.
func loadTransitionData(
	w *gui.Window, id string,
) ([][]float64, bool) {
	if w == nil || id == "" {
		return nil, false
	}
	sm := gui.StateMapRead[string, transitionDataState](
		w, nsChartTransitionData)
	if sm == nil {
		return nil, false
	}
	td, ok := sm.Get(id)
	if !ok || len(td.OldYValues) == 0 {
		return nil, false
	}
	return td.OldYValues, true
}

// loadTransitionBounds returns cached old axis bounds.
func loadTransitionBounds(
	w *gui.Window, id string,
) (minX, maxX, minY, maxY float64, ok bool) {
	if w == nil || id == "" {
		return
	}
	sm := gui.StateMapRead[string, transitionDataState](
		w, nsChartTransitionData)
	if sm == nil {
		return
	}
	td, found := sm.Get(id)
	if !found || !td.HasBounds {
		return
	}
	return td.OldBounds[0], td.OldBounds[1],
		td.OldBounds[2], td.OldBounds[3], true
}

// lerpAxisRange interpolates the axis domain between old and
// new bounds during a transition. Skips if any input is
// non-finite or if the interpolated range is degenerate.
func lerpAxisRange(
	a axis.Axis, tp float32,
	oldMin, oldMax, newMin, newMax float64,
) {
	if !finite(oldMin) || !finite(oldMax) ||
		!finite(newMin) || !finite(newMax) {
		return
	}
	lo := lerpFloat64(oldMin, newMin, float64(tp))
	hi := lerpFloat64(oldMax, newMax, float64(tp))
	if !finite(lo) || !finite(hi) || lo >= hi {
		return
	}
	a.SetRange(lo, hi)
}

// lerpFloat64 linearly interpolates between a and b.
func lerpFloat64(a, b, t float64) float64 {
	return a + (b-a)*t
}

// applyTransitionAndProgress applies transition interpolation and
// entry animation scaling to a value. Returns the animated value.
func applyTransitionAndProgress(
	v float64, si, ci int,
	tp float32, oldVals [][]float64,
	progress float32,
) float64 {
	if tp < 1 && si < len(oldVals) && ci < len(oldVals[si]) {
		v = lerpFloat64(oldVals[si][ci], v, float64(tp))
	}
	return v * float64(progress)
}

// snapshotYValues captures current Y values from XY series for
// transition interpolation.
func snapshotYValues(ss []series.XY) [][]float64 {
	out := make([][]float64, len(ss))
	for i, s := range ss {
		ys := make([]float64, s.Len())
		for j, p := range s.Points {
			ys[j] = p.Y
		}
		out[i] = ys
	}
	return out
}

// seriesBoundsXY returns the combined min/max X and Y bounds
// across all XY series.
func seriesBoundsXY(ss []series.XY) (
	minX, maxX, minY, maxY float64,
) {
	first := true
	for _, s := range ss {
		if s.Len() == 0 {
			continue
		}
		sx0, sx1, sy0, sy1 := s.Bounds()
		if first {
			minX, maxX, minY, maxY = sx0, sx1, sy0, sy1
			first = false
			continue
		}
		minX = min(minX, sx0)
		maxX = max(maxX, sx1)
		minY = min(minY, sy0)
		maxY = max(maxY, sy1)
	}
	return
}

// snapshotCategoryValues captures current values from Category
// series for transition interpolation.
func snapshotCategoryValues(ss []series.Category) [][]float64 {
	out := make([][]float64, len(ss))
	for i, s := range ss {
		vs := make([]float64, len(s.Values))
		for j, v := range s.Values {
			vs[j] = v.Value
		}
		out[i] = vs
	}
	return out
}

// startTransition begins a 0→1 tween for data transition
// animation. Idempotent: does nothing if already active.
//
// Called from draw which runs under w.mu, so AnimationAdd is
// deferred via QueueCommand.
func startTransition(w *gui.Window, id string, dur time.Duration) {
	if w == nil || id == "" {
		return
	}
	sm := chartTransitionMap(w)
	ts, _ := sm.Get(id)
	if ts.Active {
		return
	}
	if dur <= 0 {
		dur = DefaultTransitionDuration
	}
	// Mark active immediately so first draw sees progress=0.
	ts.Active = true
	ts.Progress = 0
	ts.Version++
	sm.Set(id, ts)

	animID := animTransitionPrefix + id
	w.QueueCommand(func(w *gui.Window) {
		tw := &gui.TweenAnimation{
			AnimID:   animID,
			Duration: dur,
			Easing:   gui.EaseOutCubic,
			From:     0,
			To:       1,
			OnValue: func(v float32, w *gui.Window) {
				sm := chartTransitionMap(w)
				ts, _ := sm.Get(id)
				ts.Progress = v
				ts.Version++
				sm.Set(id, ts)
			},
			OnDone: func(w *gui.Window) {
				sm := chartTransitionMap(w)
				ts, _ := sm.Get(id)
				ts.Progress = 1
				ts.Active = false
				ts.Version++
				sm.Set(id, ts)
			},
		}
		w.AnimationAdd(tw)
	})
}
