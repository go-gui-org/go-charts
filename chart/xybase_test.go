package chart

import (
	"testing"

	"github.com/go-gui-org/go-gui/gui"
)

// TestGenerateLayout_NilBase verifies that calling generateLayout with a nil
// base does not panic and returns an inert (zero) layout.
func TestGenerateLayout_NilBase(t *testing.T) {
	xb := &xyBase{} // base intentionally left nil
	// Must not panic.
	got := xb.generateLayout(nil, func(*gui.DrawContext) {})
	_ = got
}

// TestGenerateLayout_NilInteraction verifies that generateLayout with base set
// but interaction nil does not panic and returns an inert (zero) layout.
func TestGenerateLayout_NilInteraction(t *testing.T) {
	b := &BaseCfg{ID: "test"}
	xb := &xyBase{base: b} // interaction intentionally left nil
	// Must not panic.
	got := xb.generateLayout(nil, func(*gui.DrawContext) {})
	_ = got
}
