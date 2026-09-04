package chart

import (
	"math"
	"sync"
	"testing"

	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func TestRealTimeSeriesAppend(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{Name: "test"})
	rts.Append(series.Point{X: 1, Y: 10})
	rts.Append(series.Point{X: 2, Y: 20})
	if rts.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", rts.Len())
	}
	snap := rts.Snapshot()
	if snap.Len() != 2 {
		t.Fatalf("Snapshot Len() = %d, want 2", snap.Len())
	}
	if snap.Points[0].Y != 10 || snap.Points[1].Y != 20 {
		t.Errorf("unexpected values: %v", snap.Points)
	}
}

func TestRealTimeSeriesRollingWindow(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{
		Name:   "test",
		MaxLen: 3,
	})
	for i := range 5 {
		rts.Append(series.Point{X: float64(i), Y: float64(i * 10)})
	}
	if rts.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", rts.Len())
	}
	snap := rts.Snapshot()
	// Should have points 2,3,4 (oldest evicted).
	if snap.Points[0].X != 2 || snap.Points[2].X != 4 {
		t.Errorf("rolling window wrong: %v", snap.Points)
	}
}

func TestRealTimeSeriesAppendBatch(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{
		Name:   "test",
		MaxLen: 4,
	})
	pts := []series.Point{
		{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2},
		{X: 3, Y: 3}, {X: 4, Y: 4}, {X: 5, Y: 5},
	}
	rts.AppendBatch(pts)
	if rts.Len() != 4 {
		t.Fatalf("Len() = %d, want 4", rts.Len())
	}
	snap := rts.Snapshot()
	if snap.Points[0].X != 2 {
		t.Errorf("first point X = %v, want 2", snap.Points[0].X)
	}
}

func TestRealTimeSeriesVersion(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{Name: "test"})
	v0 := rts.Version()
	rts.Append(series.Point{X: 1, Y: 1})
	v1 := rts.Version()
	if v1 <= v0 {
		t.Errorf("version did not increase: %d -> %d", v0, v1)
	}
}

func TestRealTimeSeriesClear(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{Name: "test"})
	rts.Append(series.Point{X: 1, Y: 1})
	rts.Clear()
	if rts.Len() != 0 {
		t.Fatalf("Len() after Clear = %d, want 0", rts.Len())
	}
}

func TestRealTimeSeriesSnapshotIsolation(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{Name: "test"})
	rts.Append(series.Point{X: 1, Y: 1})
	snap := rts.Snapshot()
	rts.Append(series.Point{X: 2, Y: 2})
	// Snapshot should not see the new point.
	if snap.Len() != 1 {
		t.Errorf("snapshot Len = %d after new append, want 1",
			snap.Len())
	}
}

func TestRealTimeSeriesConcurrency(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{
		Name:   "test",
		MaxLen: 100,
	})
	var wg sync.WaitGroup
	for g := range 10 {
		wg.Go(func() {
			for i := range 100 {
				rts.Append(series.Point{
					X: float64(g*100 + i),
					Y: float64(i),
				})
			}
		})
	}
	// Concurrent reads.
	for range 5 {
		wg.Go(func() {
			for range 50 {
				_ = rts.Snapshot()
				_ = rts.Version()
				_ = rts.Len()
			}
		})
	}
	wg.Wait()
	if rts.Len() != 100 {
		t.Errorf("final Len = %d, want 100", rts.Len())
	}
}

func TestRealTimeSeriesAppendBatchEmpty(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{Name: "test"})
	v := rts.Version()
	rts.AppendBatch(nil)
	if rts.Version() != v {
		t.Error("empty AppendBatch should not bump version")
	}
}

func TestLoadScrollVersionNilWindow(t *testing.T) {
	if got := loadScrollVersion(nil, "test"); got != 0 {
		t.Errorf("loadScrollVersion(nil) = %v, want 0", got)
	}
}

func TestScrollXMaxNilWindow(t *testing.T) {
	_, ok := scrollXMax(nil, "test")
	if ok {
		t.Error("scrollXMax(nil) returned ok=true")
	}
}

func TestUpdateAutoScrollNilWindow(t *testing.T) {
	// Should not panic.
	updateAutoScroll(nil, "test", 100, 50)
}

func TestNewRealTimeSeriesNegativeMaxLen(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{
		Name: "neg", MaxLen: -5,
	})
	// Negative MaxLen → unlimited mode.
	rts.Append(series.Point{X: 1, Y: 1})
	if rts.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", rts.Len())
	}
}

func TestNewRealTimeSeriesHugeMaxLen(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{
		Name: "huge", MaxLen: 1 << 40,
	})
	// Should be capped, not OOM.
	rts.Append(series.Point{X: 1, Y: 1})
	if rts.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", rts.Len())
	}
}

func TestAppendDropsNaN(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{Name: "nan"})
	rts.Append(series.Point{X: 1, Y: 1})
	rts.Append(series.Point{X: math.NaN(), Y: 2})
	rts.Append(series.Point{X: 3, Y: math.Inf(1)})
	rts.Append(series.Point{X: 4, Y: math.Inf(-1)})
	if rts.Len() != 1 {
		t.Errorf("Len() = %d, want 1 (NaN/Inf dropped)", rts.Len())
	}
}

func TestAppendBatchDropsNaN(t *testing.T) {
	rts := NewRealTimeSeries(RealTimeSeriesCfg{
		Name: "nan", MaxLen: 10,
	})
	pts := []series.Point{
		{X: 1, Y: 1},
		{X: math.NaN(), Y: 2},
		{X: 3, Y: math.Inf(1)},
		{X: 4, Y: 4},
	}
	rts.AppendBatch(pts)
	if rts.Len() != 2 {
		t.Errorf("Len() = %d, want 2 (NaN/Inf dropped)", rts.Len())
	}
	snap := rts.Snapshot()
	if snap.Points[0].X != 1 || snap.Points[1].X != 4 {
		t.Errorf("unexpected points: %v", snap.Points)
	}
}

// TestUpdateAutoScrollSnapsBackOnAReset covers a chart whose series is
// cleared and refilled from zero, which is what a second run of a live
// dashboard does. The window has to follow the data back to the start;
// leaving it on the old run's right edge draws the new curve off the
// left of the plot.
func TestUpdateAutoScrollSnapsBackOnAReset(t *testing.T) {
	w := &gui.Window{}
	const id = "live"

	// A run that reaches 18 seconds.
	updateAutoScroll(w, id, 0, 30)
	updateAutoScroll(w, id, 18, 30)
	sm := chartScrollMap(w)
	ss, _ := sm.Get(id)
	if ss.TargetXMax != 18 {
		t.Fatalf("TargetXMax after the first run = %v, want 18", ss.TargetXMax)
	}

	// The series is cleared: bounds collapse to zero.
	updateAutoScroll(w, id, 0, 30)
	ss, _ = sm.Get(id)
	if ss.TargetXMax != 0 || ss.CurrentXMax != 0 {
		t.Fatalf("after a reset the window is at %v/%v, want 0/0",
			ss.CurrentXMax, ss.TargetXMax)
	}
	if ss.Active {
		t.Error("a reset should snap, not animate")
	}

	// The second run scrolls again from the start.
	updateAutoScroll(w, id, 0.5, 30)
	ss, _ = sm.Get(id)
	if ss.TargetXMax != 0.5 {
		t.Errorf("TargetXMax in the second run = %v, want 0.5", ss.TargetXMax)
	}
}
