package series

import "github.com/go-gui-org/go-gui/gui"

// TreeNode represents a node in a hierarchical dataset for treemap
// charts. Leaf nodes carry a Value directly; branch nodes derive
// their value from the sum of their children.
type TreeNode struct {
	Label     string
	Value     float64
	NodeColor gui.Color // optional; zero = use palette
	Children  []TreeNode
}

// Name implements Series.
func (n TreeNode) Name() string { return n.Label }

// Len implements Series. Returns total leaf count.
func (n TreeNode) Len() int {
	if len(n.Children) == 0 {
		return 1
	}
	total := 0
	for i := range n.Children {
		total += n.Children[i].Len()
	}
	return total
}

// Color implements Series.
func (n TreeNode) Color() gui.Color { return n.NodeColor }

// IsLeaf reports whether the node has no children.
func (n TreeNode) IsLeaf() bool { return len(n.Children) == 0 }

// TotalValue returns the node's value for layout purposes.
// Leaf nodes return Value directly. Branch nodes return the
// sum of their children's TotalValue.
func (n TreeNode) TotalValue() float64 {
	if len(n.Children) == 0 {
		return n.Value
	}
	total := 0.0
	for i := range n.Children {
		total += n.Children[i].TotalValue()
	}
	return total
}
