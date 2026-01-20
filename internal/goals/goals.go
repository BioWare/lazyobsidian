// Package goals handles goal tracking and hierarchy.
package goals

import (
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// Tree represents a hierarchical goal structure.
type Tree struct {
	roots []*types.Goal
}

// NewTree creates a new goal tree.
func NewTree() *Tree {
	return &Tree{
		roots: make([]*types.Goal, 0),
	}
}

// AddRoot adds a root-level goal.
func (t *Tree) AddRoot(goal *types.Goal) {
	t.roots = append(t.roots, goal)
}

// Roots returns all root-level goals.
func (t *Tree) Roots() []*types.Goal {
	return t.roots
}

// CalculateProgress recursively calculates progress for a goal.
func CalculateProgress(goal *types.Goal) float64 {
	if len(goal.Children) == 0 {
		return goal.Progress
	}

	var total float64
	for _, child := range goal.Children {
		total += CalculateProgress(&child)
	}
	return total / float64(len(goal.Children))
}

// AggregatePomodoros calculates total pomodoros including children.
func AggregatePomodoros(goal *types.Goal) int {
	total := goal.OwnPomodoros
	for _, child := range goal.Children {
		total += AggregatePomodoros(&child)
	}
	return total
}
