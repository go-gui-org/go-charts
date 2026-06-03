package chart

import (
	"sync"
	"time"

	"github.com/go-gui-org/go-charts/internal/fmath"
	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

// maxRealTimePoints caps the ring buffer allocation to prevent
// OOM from unreasonable MaxLen values.
const maxRealTimePoints = 10_000_000

// RealTimeSeriesCfg configures a RealTimeSeries.
type RealTimeSeriesCfg struct {
	Name  string
	Color gui.Color
	// MaxLen is the rolling window size. Zero means unlimited.
	MaxLen int
}

// RealTimeSeries is a thread-safe XY series that supports
// streaming data append and rolling window eviction.
//
// When MaxLen is set, the buffer acts as a ring buffer: new
// points overwrite the oldest slot in O(1) without shifting.
type RealTimeSeries struct {
	mu      sync.RWMutex
	name    string
	color   gui.Color
	buf     []series.Point
	maxLen  int
	head    int // ring start index
	count   int // number of valid points
	version uint64
}

// NewRealTimeSeries creates a new real-time data series.
// Negative MaxLen is treated as zero (unlimited). MaxLen is
// capped at maxRealTimePoints to prevent OOM.
func NewRealTimeSeries(cfg RealTimeSeriesCfg) *RealTimeSeries {
	ml := min(max(cfg.MaxLen, 0), maxRealTimePoints)
	var buf []series.Point
	if ml > 0 {
		buf = make([]series.Point, ml)
	}
	return &RealTimeSeries{
		name:   cfg.Name,
		color:  cfg.Color,
		maxLen: ml,
		buf:    buf,
	}
}

// insertOne appends a single point to the ring buffer without
// locking. Caller must hold rts.mu.
func (rts *RealTimeSeries) insertOne(p series.Point) {
	if rts.maxLen > 0 {
		idx := (rts.head + rts.count) % rts.maxLen
		rts.buf[idx] = p
		if rts.count < rts.maxLen {
			rts.count++
		} else {
			rts.head = (rts.head + 1) % rts.maxLen
		}
	} else {
		rts.buf = append(rts.buf, p)
		rts.count++
	}
}

// Append adds a single data point. If MaxLen is set and the
// buffer is full, the oldest point is overwritten in O(1).
// Non-finite (NaN/Inf) points are silently dropped.
// Thread-safe.
func (rts *RealTimeSeries) Append(p series.Point) {
	if !fmath.Finite(p.X) || !fmath.Finite(p.Y) {
		return
	}
	rts.mu.Lock()
	rts.insertOne(p)
	rts.version++
	rts.mu.Unlock()
}

// AppendBatch adds multiple points atomically.
// Non-finite (NaN/Inf) points are silently dropped.
// Thread-safe.
func (rts *RealTimeSeries) AppendBatch(pts []series.Point) {
	if len(pts) == 0 {
		return
	}
	rts.mu.Lock()
	for _, p := range pts {
		if !fmath.Finite(p.X) || !fmath.Finite(p.Y) {
			continue
		}
		rts.insertOne(p)
	}
	rts.version++
	rts.mu.Unlock()
}

// Snapshot returns a copy of the current data as a series.XY.
// The returned value is safe to use without locking.
func (rts *RealTimeSeries) Snapshot() series.XY {
	rts.mu.RLock()
	pts := make([]series.Point, rts.count)
	if rts.maxLen > 0 {
		// Linearize ring buffer.
		first := rts.maxLen - rts.head
		first = min(first, rts.count)
		copy(pts, rts.buf[rts.head:rts.head+first])
		if first < rts.count {
			copy(pts[first:], rts.buf[:rts.count-first])
		}
	} else {
		copy(pts, rts.buf[:rts.count])
	}
	rts.mu.RUnlock()
	return series.NewXY(series.XYCfg{
		Name:   rts.name,
		Color:  rts.color,
		Points: pts,
	})
}

// Version returns the current version counter. Each Append or
// AppendBatch increments it. Use as BaseCfg.Version to trigger
// re-render.
func (rts *RealTimeSeries) Version() uint64 {
	rts.mu.RLock()
	v := rts.version
	rts.mu.RUnlock()
	return v
}

// Len returns the current number of points. Thread-safe.
func (rts *RealTimeSeries) Len() int {
	rts.mu.RLock()
	n := rts.count
	rts.mu.RUnlock()
	return n
}

// Clear removes all points and resets the version. Thread-safe.
func (rts *RealTimeSeries) Clear() {
	rts.mu.Lock()
	rts.head = 0
	rts.count = 0
	rts.version++
	rts.mu.Unlock()
}

// scrollState persists smooth-scroll animation state across
// frames via gui.StateMap. Keyed by chart ID.
type scrollState struct {
	TargetXMax  float64
	CurrentXMax float64
	Active      bool
	Initialized bool
	Version     uint64
}

const (
	nsChartScroll  = "chart-scroll"
	capChartScroll = 64

	// DefaultScrollDuration is the smooth scroll animation
	// duration.
	DefaultScrollDuration = 200 * time.Millisecond

	animScrollPrefix = "chart-scroll-"
)

// chartScrollMap returns the persistent scroll state map.
func chartScrollMap(w *gui.Window) *gui.BoundedMap[string, scrollState] {
	return gui.StateMap[string, scrollState](
		w, nsChartScroll, capChartScroll)
}

// loadScrollVersion returns the scroll version for inclusion
// in the DrawCanvasCfg.Version sum.
func loadScrollVersion(w *gui.Window, id string) uint64 {
	if w == nil || id == "" {
		return 0
	}
	sm := gui.StateMapRead[string, scrollState](w, nsChartScroll)
	if sm == nil {
		return 0
	}
	ss, ok := sm.Get(id)
	if !ok {
		return 0
	}
	return ss.Version
}

// scrollXMax returns the current animated X-axis maximum for
// smooth scrolling. Returns 0,false if no scroll state exists.
func scrollXMax(w *gui.Window, id string) (float64, bool) {
	if w == nil || id == "" {
		return 0, false
	}
	sm := gui.StateMapRead[string, scrollState](w, nsChartScroll)
	if sm == nil {
		return 0, false
	}
	ss, ok := sm.Get(id)
	if !ok {
		return 0, false
	}
	return ss.CurrentXMax, true
}

// updateAutoScroll checks if the data max has changed and
// starts a smooth scroll animation if needed.
func updateAutoScroll(
	w *gui.Window, id string, dataXMax float64, windowSize float64,
) {
	if w == nil || id == "" || windowSize <= 0 {
		return
	}
	if !fmath.Finite(dataXMax) || !fmath.Finite(windowSize) {
		return
	}
	sm := chartScrollMap(w)
	ss, _ := sm.Get(id)

	// First call: initialize without animating.
	if !ss.Initialized {
		ss.Initialized = true
		ss.TargetXMax = dataXMax
		ss.CurrentXMax = dataXMax
		ss.Version++
		sm.Set(id, ss)
		return
	}

	// No change.
	if dataXMax <= ss.TargetXMax {
		return
	}

	oldMax := ss.CurrentXMax
	ss.TargetXMax = dataXMax
	ss.Active = true
	ss.Version++
	sm.Set(id, ss)

	// Use a normalized 0→1 tween and lerp float64 values
	// manually to avoid float32 precision loss on large X.
	animID := animScrollPrefix + id
	capturedMax := dataXMax
	w.QueueCommand(func(w *gui.Window) {
		tw := &gui.TweenAnimation{
			AnimID:   animID,
			Duration: DefaultScrollDuration,
			Easing:   gui.EaseOutCubic,
			From:     0,
			To:       1,
			OnValue: func(t float32, w *gui.Window) {
				sm := chartScrollMap(w)
				ss, _ := sm.Get(id)
				ss.CurrentXMax = lerpFloat64(
					oldMax, capturedMax, float64(t))
				ss.Version++
				sm.Set(id, ss)
			},
			OnDone: func(w *gui.Window) {
				sm := chartScrollMap(w)
				ss, _ := sm.Get(id)
				ss.CurrentXMax = capturedMax
				ss.Active = false
				ss.Version++
				sm.Set(id, ss)
			},
		}
		w.AnimationAdd(tw)
	})
}
