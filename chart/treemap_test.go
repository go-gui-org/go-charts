package chart

import (
	"math"
	"testing"

	"github.com/go-gui-org/go-charts/series"
	"github.com/go-gui-org/go-gui/gui"
)

func TestTreeNodeTotalValue(t *testing.T) {
	tests := []struct {
		name string
		node series.TreeNode
		want float64
	}{
		{"leaf", series.TreeNode{Value: 42}, 42},
		{"branch sums children",
			series.TreeNode{Children: []series.TreeNode{
				{Value: 10},
				{Value: 20},
				{Value: 30},
			}}, 60},
		{"nested",
			series.TreeNode{Children: []series.TreeNode{
				{Children: []series.TreeNode{
					{Value: 5}, {Value: 3},
				}},
				{Value: 12},
			}}, 20},
		{"empty branch", series.TreeNode{Children: []series.TreeNode{}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.node.TotalValue()
			if got != tt.want {
				t.Errorf("TotalValue() = %g, want %g", got, tt.want)
			}
		})
	}
}

func TestTreeNodeIsLeaf(t *testing.T) {
	leaf := series.TreeNode{Value: 1}
	if !leaf.IsLeaf() {
		t.Error("expected leaf")
	}
	branch := series.TreeNode{Children: []series.TreeNode{{Value: 1}}}
	if branch.IsLeaf() {
		t.Error("expected non-leaf")
	}
}

func TestWorstAspectRatio(t *testing.T) {
	tests := []struct {
		name      string
		areas     []float64
		total     float64
		shortSide float32
		wantMax   float64
	}{
		{"single square", []float64{100}, 100, 10, 1.0},
		{"single rect", []float64{200}, 200, 10, 2.0},
		{"empty", nil, 0, 10, math.MaxFloat64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := worstAspectRatio(tt.areas, tt.total, tt.shortSide)
			if math.Abs(got-tt.wantMax) > 0.01 {
				t.Errorf("worstAspectRatio = %g, want %g",
					got, tt.wantMax)
			}
		})
	}
}

func TestSquarifyLayout(t *testing.T) {
	data := []series.TreeNode{
		{Label: "A", Value: 60},
		{Label: "B", Value: 30},
		{Label: "C", Value: 10},
	}
	tv := &treemapView{
		cfg: TreemapCfg{
			Data:     data,
			MaxDepth: 2,
			CellGap:  0,
		},
	}

	roots := make([]nodeArea, len(data))
	for i := range data {
		roots[i] = nodeArea{
			node:     &tv.cfg.Data[i],
			area:     data[i].TotalValue(),
			groupIdx: i,
		}
	}

	tv.squarify(roots, 0, 0, 100, 100, 0)

	// All cells should tile the 100x100 rect.
	if len(tv.cells) != 3 {
		t.Fatalf("expected 3 cells, got %d", len(tv.cells))
	}

	totalArea := float64(0)
	for _, c := range tv.cells {
		if c.W <= 0 || c.H <= 0 {
			t.Errorf("cell %q has non-positive dims: %gx%g",
				c.Node.Label, c.W, c.H)
		}
		totalArea += float64(c.W) * float64(c.H)
	}

	if math.Abs(totalArea-10000) > 1 {
		t.Errorf("total cell area = %g, want 10000", totalArea)
	}

	// No overlaps: check pairwise.
	for i := 0; i < len(tv.cells); i++ {
		for j := i + 1; j < len(tv.cells); j++ {
			a, b := tv.cells[i], tv.cells[j]
			if a.X < b.X+b.W && a.X+a.W > b.X &&
				a.Y < b.Y+b.H && a.Y+a.H > b.Y {
				overlapW := min(a.X+a.W, b.X+b.W) - max(a.X, b.X)
				overlapH := min(a.Y+a.H, b.Y+b.H) - max(a.Y, b.Y)
				if overlapW > 0.5 && overlapH > 0.5 {
					t.Errorf("cells %d and %d overlap: "+
						"(%g,%g,%g,%g) vs (%g,%g,%g,%g)",
						i, j,
						a.X, a.Y, a.W, a.H,
						b.X, b.Y, b.W, b.H)
				}
			}
		}
	}
}

func TestSquarifyNested(t *testing.T) {
	data := []series.TreeNode{
		{Label: "Group", Children: []series.TreeNode{
			{Label: "A", Value: 40},
			{Label: "B", Value: 20},
		}},
	}
	tv := &treemapView{
		cfg: TreemapCfg{
			Data:     data,
			MaxDepth: 3,
			CellGap:  0,
		},
	}

	roots := []nodeArea{{
		node:     &tv.cfg.Data[0],
		area:     60,
		groupIdx: 0,
	}}

	tv.squarify(roots, 0, 0, 100, 60, 0)

	// Should have 2 leaf cells (A and B).
	if len(tv.cells) != 2 {
		t.Fatalf("expected 2 leaf cells, got %d", len(tv.cells))
	}
}

func TestTreemapHitTest(t *testing.T) {
	tv := &treemapView{
		cells: []treemapCell{
			{X: 0, Y: 0, W: 50, H: 100,
				Node: &series.TreeNode{Label: "A"}},
			{X: 50, Y: 0, W: 50, H: 60,
				Node: &series.TreeNode{Label: "B"}},
			{X: 50, Y: 60, W: 50, H: 40,
				Node: &series.TreeNode{Label: "C"}},
		},
	}

	idx, ok := tv.hitTest(25, 50)
	if !ok || tv.cells[idx].Node.Label != "A" {
		t.Errorf("hitTest(25,50) = (%d,%v), want cell A", idx, ok)
	}

	idx, ok = tv.hitTest(75, 30)
	if !ok || tv.cells[idx].Node.Label != "B" {
		t.Errorf("hitTest(75,30) = (%d,%v), want cell B", idx, ok)
	}

	idx, ok = tv.hitTest(75, 80)
	if !ok || tv.cells[idx].Node.Label != "C" {
		t.Errorf("hitTest(75,80) = (%d,%v), want cell C", idx, ok)
	}

	_, ok = tv.hitTest(200, 200)
	if ok {
		t.Error("hitTest outside should return false")
	}
}

func TestTreemapNaNNodeSkipped(t *testing.T) {
	data := []series.TreeNode{
		{Label: "Good", Value: 60},
		{Label: "NaN", Value: math.NaN()},
		{Label: "Inf", Value: math.Inf(1)},
	}
	tv := &treemapView{
		cfg: TreemapCfg{
			Data:     data,
			MaxDepth: 2,
			CellGap:  0,
		},
	}

	roots := make([]nodeArea, 0, len(data))
	for i := range data {
		v := data[i].TotalValue()
		if !finite(v) || v <= 0 {
			continue
		}
		roots = append(roots, nodeArea{
			node:     &tv.cfg.Data[i],
			area:     v,
			groupIdx: i,
		})
	}

	// Only the "Good" node should survive filtering.
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].node.Label != "Good" {
		t.Errorf("expected Good node, got %s", roots[0].node.Label)
	}

	tv.squarify(roots, 0, 0, 100, 100, 0)
	if len(tv.cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(tv.cells))
	}
}

func TestTreemapValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     TreemapCfg
		wantErr bool
	}{
		{"empty data",
			TreemapCfg{}, true},
		{"negative CellGap",
			TreemapCfg{
				Data:    []series.TreeNode{{Value: 1}},
				CellGap: -1,
			}, true},
		{"negative MaxDepth",
			TreemapCfg{
				Data:     []series.TreeNode{{Value: 1}},
				MaxDepth: -1,
			}, true},
		{"negative HeaderHeight",
			TreemapCfg{
				Data:         []series.TreeNode{{Value: 1}},
				HeaderHeight: -1,
			}, true},
		{"negative leaf value",
			TreemapCfg{
				Data: []series.TreeNode{
					{Children: []series.TreeNode{{Value: -5}}},
				},
			}, true},
		{"valid",
			TreemapCfg{
				Data: []series.TreeNode{{Value: 10}},
			}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}
}

func TestTreemapColor(t *testing.T) {
	palette := []gui.Color{gui.Hex(0xFF0000)}

	// Depth 0 = base color.
	c0 := treemapColor(0, 0, palette, gui.Color{})
	if c0 != gui.Hex(0xFF0000) {
		t.Errorf("depth 0: got %v, want red", c0)
	}

	// Depth 1 = lightened.
	c1 := treemapColor(0, 1, palette, gui.Color{})
	if c1.R <= c0.R || c1.G <= c0.G {
		// Lightened red should have higher G (and possibly B).
		// At minimum, it should differ from base.
		if c1 == c0 {
			t.Errorf("depth 1 should differ from depth 0")
		}
	}

	// Explicit node color overrides palette.
	custom := gui.Hex(0x00FF00)
	cc := treemapColor(0, 0, palette, custom)
	if cc != custom {
		t.Errorf("custom color: got %v, want %v", cc, custom)
	}
}
