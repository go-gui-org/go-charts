package axis

import (
	"fmt"
	"math"
	"testing"
)

func TestNiceNumber(t *testing.T) {
	tests := []struct {
		value float64
		round bool
		want  float64
	}{
		{0.7, true, 0.5},
		{3.5, true, 5},
		{7.5, true, 10},
		{12, true, 10},
		{0, true, 0},
	}
	for _, tt := range tests {
		got := NiceNumber(tt.value, tt.round)
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("NiceNumber(%v, %v) = %v, want %v",
				tt.value, tt.round, got, tt.want)
		}
	}
}

func TestGenerateNiceTicks(t *testing.T) {
	ticks := GenerateNiceTicks(0, 100, 5)
	if len(ticks) == 0 {
		t.Fatal("expected ticks, got none")
	}
	if ticks[0] > 0 {
		t.Errorf("first tick %v should be <= 0", ticks[0])
	}
	last := ticks[len(ticks)-1]
	if last < 100 {
		t.Errorf("last tick %v should be >= 100", last)
	}
}

func TestLinearTickFormat(t *testing.T) {
	a := NewLinear(LinearCfg{
		Min: 0, Max: 100,
		TickFormat: func(v float64) string {
			return fmt.Sprintf("$%.0f", v)
		},
	})
	ticks := a.Ticks(0, 500)
	if len(ticks) == 0 {
		t.Fatal("expected ticks")
	}
	for _, tk := range ticks {
		if tk.Label[0] != '$' {
			t.Errorf("tick label %q missing $ prefix", tk.Label)
		}
	}
}

func TestLinearTickFormatNil(t *testing.T) {
	a := NewLinear(LinearCfg{Min: 0, Max: 10})
	ticks := a.Ticks(0, 100)
	if len(ticks) == 0 {
		t.Fatal("expected ticks")
	}
	// Default format should not be empty.
	for _, tk := range ticks {
		if tk.Label == "" {
			t.Errorf("tick label should not be empty")
		}
	}
}

func TestOverrideDomainPreventsTickExpansion(t *testing.T) {
	t.Parallel()
	a := NewLinear(LinearCfg{AutoRange: true})
	a.SetRange(1.3, 4.7) // not "nice" boundaries
	a.SetOverrideDomain(true)
	_ = a.Ticks(0, 400) // should not expand domain

	dMin, dMax := a.Domain()
	if dMin != 1.3 || dMax != 4.7 {
		t.Errorf("domain changed to [%g, %g], want [1.3, 4.7]",
			dMin, dMax)
	}
}

func TestOverrideDomainFalseAllowsExpansion(t *testing.T) {
	t.Parallel()
	a := NewLinear(LinearCfg{AutoRange: true})
	a.SetRange(1.3, 4.7)
	// overrideDomain defaults to false — Ticks should expand.
	_ = a.Ticks(0, 400)

	dMin, dMax := a.Domain()
	// Nice ticks for [1.3, 4.7] will expand to something like [1, 5].
	if dMin == 1.3 && dMax == 4.7 {
		t.Error("domain was not expanded by Ticks with AutoRange")
	}
	if dMin > 1.3 || dMax < 4.7 {
		t.Errorf("domain [%g, %g] does not contain [1.3, 4.7]",
			dMin, dMax)
	}
}

func TestGenerateNiceTicksPrecision(t *testing.T) {
	t.Parallel()
	// Verify ticks are cleanly snapped for various spacings,
	// including those with float64 representation issues.
	tests := []struct {
		min, max float64
		ticks    int
	}{
		{0, 1, 10},   // spacing ~ 0.1
		{0, 2, 10},   // spacing ~ 0.2
		{0, 3, 10},   // spacing ~ 0.3
		{0, 0.5, 10}, // spacing ~ 0.05
		{0, 50, 10},  // spacing ~ 5 (integer)
	}
	for _, tt := range tests {
		ticks := GenerateNiceTicks(tt.min, tt.max, tt.ticks)
		if len(ticks) < 2 {
			t.Errorf("[%g,%g]: got %d ticks, want >= 2",
				tt.min, tt.max, len(ticks))
			continue
		}
		spacing := ticks[1] - ticks[0]
		for i := 1; i < len(ticks); i++ {
			gap := ticks[i] - ticks[i-1]
			if math.Abs(gap-spacing) > 1e-9 {
				t.Errorf("[%g,%g]: uneven gap at tick %d: %g vs %g",
					tt.min, tt.max, i, gap, spacing)
			}
		}
	}
}

func TestDomainMethod(t *testing.T) {
	t.Parallel()
	a := NewLinear(LinearCfg{Min: 10, Max: 20})
	lo, hi := a.Domain()
	if lo != 10 || hi != 20 {
		t.Errorf("Domain() = [%g, %g], want [10, 20]", lo, hi)
	}
}

func TestGenerateExactTicksAlwaysReturnsCount(t *testing.T) {
	// A live throughput axis: the maximum climbs through every
	// magnitude, and the label count must not move with it.
	for _, max := range []float64{1, 7, 12, 45, 99, 100, 137, 480, 501, 999, 1024, 8600} {
		for _, count := range []int{3, 4, 5, 8} {
			ticks := GenerateExactTicks(0, max, count)
			if len(ticks) != count {
				t.Fatalf("max %v count %d: got %d ticks %v", max, count, len(ticks), ticks)
			}
			if ticks[0] != 0 {
				t.Errorf("max %v count %d: floor moved to %v", max, count, ticks[0])
			}
			if ticks[count-1] < max {
				t.Errorf("max %v count %d: top %v is below the data", max, count, ticks[count-1])
			}
		}
	}
}

func TestGenerateExactTicksHandlesDegenerateRanges(t *testing.T) {
	cases := []struct{ min, max float64 }{
		{0, 0},    // no data yet
		{50, 50},  // a flat series
		{-20, 20}, // straddling zero
		{100, 0},  // reversed
	}
	for _, c := range cases {
		ticks := GenerateExactTicks(c.min, c.max, 4)
		if len(ticks) != 4 {
			t.Errorf("[%v, %v]: got %d ticks %v", c.min, c.max, len(ticks), ticks)
		}
	}
}

func TestGenerateExactTicksNonFiniteFallback(t *testing.T) {
	for _, c := range []struct{ min, max float64 }{
		{math.NaN(), 10},
		{0, math.NaN()},
		{math.NaN(), math.NaN()},
		{math.Inf(1), 10},
		{0, math.Inf(1)},
	} {
		ticks := GenerateExactTicks(c.min, c.max, 4)
		// Fallback to GenerateNiceTicks must not panic and must return
		// a finite slice (possibly nil or single-value).
		for _, v := range ticks {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Errorf("[%v,%v]: non-finite tick %v", c.min, c.max, v)
			}
		}
	}
}

func TestGenerateExactTicksCapsLargeCount(t *testing.T) {
	ticks := GenerateExactTicks(0, 100, 100000)
	if len(ticks) != maxExactTicks {
		t.Errorf("large count: got %d, want %d", len(ticks), maxExactTicks)
	}
	// Also via Linear path: TickCount huge must not allocate unbounded.
	a := NewLinear(LinearCfg{Min: 0, Max: 100, TickCount: 100000})
	got := a.Ticks(0, 200)
	if len(got) != maxExactTicks {
		t.Errorf("Linear huge TickCount: got %d, want %d", len(got), maxExactTicks)
	}
}

func TestGenerateExactTicksClampsSmallCount(t *testing.T) {
	for _, c := range []int{0, 1, -5} {
		ticks := GenerateExactTicks(0, 10, c)
		if len(ticks) != 2 {
			t.Errorf("count %d: got %d, want 2", c, len(ticks))
		}
	}
}

func TestLinearTickCountExact(t *testing.T) {
	a := NewLinear(LinearCfg{Min: 0, Max: 100, TickCount: 5})
	ticks := a.Ticks(0, 200)
	if len(ticks) != 5 {
		t.Fatalf("TickCount=5: got %d ticks %v", len(ticks), ticks)
	}
	if ticks[0].Value != 0 {
		t.Errorf("floor moved to %v, want 0", ticks[0].Value)
	}
	if ticks[len(ticks)-1].Value < 100 {
		t.Errorf("top %v below data 100", ticks[len(ticks)-1].Value)
	}
}
