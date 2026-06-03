package chart

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/axis"
)

// TestAutoLinearAxis_NilCfgHasBounds verifies a new auto-ranged axis is
// created and SetRange is applied when no config axis is provided.
func TestAutoLinearAxis_NilCfgHasBounds(t *testing.T) {
	ax, ok := autoLinearAxis(nil, 10, 50, 0, "test")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if ax == nil {
		t.Fatal("expected non-nil axis")
	}
	lo, hi := ax.Domain()
	if lo != 10 || hi != 50 {
		t.Errorf("expected domain [10,50], got [%.4g, %.4g]", lo, hi)
	}
}

// TestAutoLinearAxis_NilCfgNoBounds verifies (nil, false) when no config
// axis is provided and there are no bounds.
func TestAutoLinearAxis_NilCfgNoBounds(t *testing.T) {
	ax, ok := autoLinearAxis(nil, math.MaxFloat64, -math.MaxFloat64, 0, "test")
	if ok {
		t.Error("expected ok=false for empty bounds")
	}
	if ax != nil {
		t.Error("expected nil axis")
	}
}

// TestAutoLinearAxis_CfgAxisHasBounds verifies a provided config axis has its
// range updated when bounds are valid.
func TestAutoLinearAxis_CfgAxisHasBounds(t *testing.T) {
	cfg := axis.NewLinear(axis.LinearCfg{})
	cfg.SetRange(0, 1) // initial range, should be replaced

	ax, ok := autoLinearAxis(cfg, 20, 80, 0, "test")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if ax != cfg {
		t.Error("expected the same cfg axis to be returned")
	}
	lo, hi := ax.Domain()
	if lo != 20 || hi != 80 {
		t.Errorf("expected domain [20,80], got [%.4g, %.4g]", lo, hi)
	}
}

// TestAutoLinearAxis_CfgAxisNoBounds verifies a provided config axis is
// returned unchanged (ok=true) when there are no data bounds.
func TestAutoLinearAxis_CfgAxisNoBounds(t *testing.T) {
	cfg := axis.NewLinear(axis.LinearCfg{})
	cfg.SetRange(5, 95)

	ax, ok := autoLinearAxis(cfg, math.MaxFloat64, -math.MaxFloat64, 0.05, "test")
	if !ok {
		t.Fatal("expected ok=true: cfg axis owns its domain")
	}
	if ax != cfg {
		t.Error("expected the same cfg axis to be returned")
	}
	// Domain must not have been overwritten.
	lo, hi := ax.Domain()
	if lo != 5 || hi != 95 {
		t.Errorf("domain changed unexpectedly: got [%.4g, %.4g]", lo, hi)
	}
}

// TestAutoLinearAxis_SingleValue verifies that a degenerate range (minV==maxV)
// does not cause a division by zero or zero-width axis.
func TestAutoLinearAxis_SingleValue(t *testing.T) {
	ax, ok := autoLinearAxis(nil, 42, 42, 0.05, "test")
	if !ok {
		t.Fatal("expected ok=true")
	}
	lo, hi := ax.Domain()
	if lo >= 42 || hi <= 42 {
		t.Errorf("expected domain to bracket 42, got [%.4g, %.4g]", lo, hi)
	}
}

// TestAutoLinearAxis_OverflowRange verifies that an extreme-but-finite range
// (r overflows to Inf) does not produce infinite axis bounds.
func TestAutoLinearAxis_OverflowRange(t *testing.T) {
	ax, ok := autoLinearAxis(nil, -1e308, 1e308, 0.05, "test")
	if !ok {
		t.Fatal("expected ok=true")
	}
	lo, hi := ax.Domain()
	if math.IsInf(lo, 0) || math.IsNaN(lo) {
		t.Errorf("domain low is non-finite: %g", lo)
	}
	if math.IsInf(hi, 0) || math.IsNaN(hi) {
		t.Errorf("domain high is non-finite: %g", hi)
	}
}

// TestAutoLinearAxis_PaddingZero verifies padFrac=0 sets the axis domain
// exactly to [minV, maxV] with no outset.
func TestAutoLinearAxis_PaddingZero(t *testing.T) {
	ax, ok := autoLinearAxis(nil, 10, 90, 0, "test")
	if !ok {
		t.Fatal("expected ok=true")
	}
	lo, hi := ax.Domain()
	if lo != 10 || hi != 90 {
		t.Errorf("expected exact domain [10,90], got [%.4g, %.4g]", lo, hi)
	}
}

// TestAutoLinearAxis_PaddingFive verifies padFrac=0.05 expands the auto
// domain symmetrically beyond [minV, maxV].
func TestAutoLinearAxis_PaddingFive(t *testing.T) {
	ax, ok := autoLinearAxis(nil, 0, 100, 0.05, "test")
	if !ok {
		t.Fatal("expected ok=true")
	}
	lo, hi := ax.Domain()
	if lo >= 0 {
		t.Errorf("expected lo < 0 with 5%% padding, got %.4g", lo)
	}
	if hi <= 100 {
		t.Errorf("expected hi > 100 with 5%% padding, got %.4g", hi)
	}
}
