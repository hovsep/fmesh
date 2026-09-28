package cycle

import (
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGroup(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		group := NewGroup()
		assert.NotNil(t, group)
	})
}

func TestGroup_Add(t *testing.T) {
	t.Parallel()
	type args struct {
		cycles []*Cycle
	}
	tests := []struct {
		name       string
		group      *Group
		args       args
		assertions func(t *testing.T, group *Group)
	}{
		{
			name:  "no addition to empty group",
			group: NewGroup(),
			args: args{
				cycles: nil,
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Zero(t, group.Len())
			},
		},
		{
			name:  "adding nothing to existing group",
			group: NewGroup().Add(New().AddActivationResults(component.NewActivationResult("c1").SetActivated(false))),
			args: args{
				cycles: nil,
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 1, group.Len())
			},
		},
		{
			name:  "adding to empty group",
			group: NewGroup(),
			args: args{
				cycles: []*Cycle{New().AddActivationResults(component.NewActivationResult("c1").SetActivated(false))},
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 1, group.Len())
			},
		},
		{
			name:  "adding to existing group",
			group: NewGroup().Add(New().AddActivationResults(component.NewActivationResult("c1").SetActivated(true))),
			args: args{
				cycles: []*Cycle{New().AddActivationResults(component.NewActivationResult("c1").SetActivated(false))},
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 2, group.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupAfter := tt.group.Add(tt.args.cycles...)
			if tt.assertions != nil {
				tt.assertions(t, groupAfter)
			}
		})
	}
}

func TestGroup_RemoveOldest(t *testing.T) {
	t.Parallel()
	newFourCycles := func() *Group {
		return NewGroup().Add(New().SetNumber(1), New().SetNumber(2), New().SetNumber(3), New().SetNumber(4))
	}

	tests := []struct {
		name       string
		group      *Group
		count      int
		assertions func(t *testing.T, group *Group)
	}{
		{
			name:  "remove some",
			group: newFourCycles(),
			count: 2,
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 2, group.Len())
				assert.Equal(t, 3, group.First().Number())
				assert.Equal(t, 4, group.Last().Number())
			},
		},
		{
			name:  "remove zero",
			group: newFourCycles(),
			count: 0,
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 4, group.Len())
				assert.Equal(t, 1, group.First().Number())
			},
		},
		{
			name:  "remove negative is a no-op",
			group: newFourCycles(),
			count: -1,
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 4, group.Len())
			},
		},
		{
			name:  "remove all",
			group: newFourCycles(),
			count: 4,
			assertions: func(t *testing.T, group *Group) {
				assert.Zero(t, group.Len())
				assert.Nil(t, group.Last())
			},
		},
		{
			name:  "count greater than length is clamped",
			group: newFourCycles(),
			count: 100,
			assertions: func(t *testing.T, group *Group) {
				assert.Zero(t, group.Len())
			},
		},
		{
			name:  "empty group",
			group: NewGroup(),
			count: 2,
			assertions: func(t *testing.T, group *Group) {
				assert.Zero(t, group.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.group.RemoveOldest(tt.count)
			assert.Same(t, tt.group, result, "RemoveOldest mutates and returns the receiver")
			if tt.assertions != nil {
				tt.assertions(t, result)
			}
		})
	}
}

func TestGroup_SetLenLimit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		group      *Group
		assertions func(t *testing.T, group *Group)
	}{
		{
			name:  "no limit by default",
			group: NewGroup().Add(New().SetNumber(1), New().SetNumber(2), New().SetNumber(3)),
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 3, group.Len())
			},
		},
		{
			name:  "add beyond the limit evicts the oldest",
			group: NewGroup().SetLenLimit(2).Add(New().SetNumber(1), New().SetNumber(2), New().SetNumber(3)),
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 2, group.Len())
				assert.Equal(t, 2, group.First().Number())
				assert.Equal(t, 3, group.Last().Number())
			},
		},
		{
			name: "limit of 1 keeps only the most recent cycle",
			group: NewGroup().SetLenLimit(1).
				Add(New().SetNumber(1)).
				Add(New().SetNumber(2)).
				Add(New().SetNumber(3)),
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 1, group.Len())
				assert.Equal(t, 3, group.Last().Number())
			},
		},
		{
			name:  "setting a limit on an oversized group evicts immediately",
			group: NewGroup().Add(New().SetNumber(1), New().SetNumber(2), New().SetNumber(3)).SetLenLimit(2),
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 2, group.Len())
				assert.Equal(t, 2, group.First().Number())
			},
		},
		{
			name:  "zero limit means unlimited",
			group: NewGroup().SetLenLimit(0).Add(New().SetNumber(1), New().SetNumber(2), New().SetNumber(3)),
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 3, group.Len())
			},
		},
		{
			name:  "negative limit means unlimited",
			group: NewGroup().SetLenLimit(-5).Add(New().SetNumber(1), New().SetNumber(2), New().SetNumber(3)),
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 3, group.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertions(t, tt.group)
		})
	}
}

func TestGroup_Without(t *testing.T) {
	t.Parallel()
	c1 := New().SetNumber(1)
	c2 := New().SetNumber(2)
	c3 := New().SetNumber(3)

	tests := []struct {
		name       string
		group      *Group
		predicate  Predicate
		assertions func(t *testing.T, group *Group)
	}{
		{
			name:  "remove from empty group",
			group: NewGroup(),
			predicate: func(c *Cycle) bool {
				return c.Number() == 1
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Zero(t, group.Len())
			},
		},
		{
			name:  "remove existing cycle by number",
			group: NewGroup().Add(c1, c2, c3),
			predicate: func(c *Cycle) bool {
				return c.Number() == 2
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 2, group.Len())
			},
		},
		{
			name:  "remove odd numbered cycles",
			group: NewGroup().Add(c1, c2, c3),
			predicate: func(c *Cycle) bool {
				return c.Number()%2 == 1
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 1, group.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.group.Without(tt.predicate)
			tt.assertions(t, result)
		})
	}
}

func TestGroup_Filter(t *testing.T) {
	t.Parallel()
	c1 := New().SetNumber(1).AddActivationResults(component.NewActivationResult("c1").SetActivated(true))
	c2 := New().SetNumber(2).AddActivationResults(component.NewActivationResult("c2").SetActivated(false))
	c3 := New().SetNumber(3).AddActivationResults(component.NewActivationResult("c3").SetActivated(true))

	t.Run("filters matching cycles", func(t *testing.T) {
		group := NewGroup().Add(c1, c2, c3)
		filtered := group.Filter(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})
		assert.Equal(t, 2, filtered.Len())
	})

	t.Run("no matches returns empty group", func(t *testing.T) {
		group := NewGroup().Add(c2)
		filtered := group.Filter(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})
		assert.Equal(t, 0, filtered.Len())
	})

	t.Run("preserves group meta and len limit", func(t *testing.T) {
		group := NewGroup().SetLenLimit(5).Add(c1, c2, c3)
		group.Meta().Set("k", "v")

		filtered := group.Filter(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})

		assert.Equal(t, 5, filtered.lenLimit)
		assert.True(t, filtered.Meta().ValueIs("k", "v"))
		// The derived stores are copies, not shared with the source group.
		filtered.Meta().Set("k2", "v2")
		assert.False(t, group.Meta().Has("k2"))
	})
}

func TestGroup_MapIf(t *testing.T) {
	t.Parallel()
	t.Run("maps only matching cycles", func(t *testing.T) {
		group := NewGroup().Add(New().SetNumber(1), New().SetNumber(2), New().SetNumber(3), New().SetNumber(4))
		mapped := group.MapIf(
			func(c *Cycle) bool { return c.Number()%2 == 0 },
			func(c *Cycle) *Cycle { return c.SetNumber(c.Number() * 100) },
		)
		assert.Equal(t, 4, mapped.Len())
		assert.Equal(t, 1, mapped.First().Number()) // odd unchanged
		assert.NotNil(t, mapped.Find(func(c *Cycle) bool { return c.Number() == 200 }))
		assert.NotNil(t, mapped.Find(func(c *Cycle) bool { return c.Number() == 400 }))
	})

	t.Run("predicate matches none - all cycles kept as-is", func(t *testing.T) {
		group := NewGroup().Add(New().SetNumber(1), New().SetNumber(2), New().SetNumber(3))
		mapped := group.MapIf(
			func(c *Cycle) bool { return false },
			func(c *Cycle) *Cycle { return c.SetNumber(-1) },
		)
		assert.Equal(t, 3, mapped.Len())
		assert.Equal(t, 1, mapped.First().Number())
	})

	t.Run("predicate matches all - all cycles mapped", func(t *testing.T) {
		group := NewGroup().Add(New().SetNumber(1), New().SetNumber(2))
		mapped := group.MapIf(
			func(c *Cycle) bool { return true },
			func(c *Cycle) *Cycle { return c.SetNumber(c.Number() * 10) },
		)
		assert.Equal(t, 2, mapped.Len())
		assert.Equal(t, 10, mapped.First().Number())
	})

	t.Run("nil mapper result drops the cycle", func(t *testing.T) {
		group := NewGroup().Add(New().SetNumber(1), New().SetNumber(2), New().SetNumber(3))
		mapped := group.MapIf(
			func(c *Cycle) bool { return c.Number() == 2 },
			func(c *Cycle) *Cycle { return nil },
		)
		assert.Equal(t, 2, mapped.Len()) // c2 dropped, c1 and c3 kept
	})
}

func TestGroup_Map(t *testing.T) {
	t.Parallel()
	c1 := New().SetNumber(1)
	c2 := New().SetNumber(2)

	t.Run("transforms all cycles", func(t *testing.T) {
		group := NewGroup().Add(c1, c2)
		mapped := group.Map(func(c *Cycle) *Cycle {
			return c.SetNumber(c.Number() * 10)
		})
		assert.Equal(t, 2, mapped.Len())
		first := mapped.First()
		assert.Equal(t, 10, first.Number())
	})

	t.Run("empty group", func(t *testing.T) {
		group := NewGroup()
		mapped := group.Map(func(c *Cycle) *Cycle {
			return c
		})
		assert.Equal(t, 0, mapped.Len())
	})
}

func TestGroup_FirstDoesNotPoisonGroup(t *testing.T) {
	t.Parallel()
	t.Run("First does not poison group when empty", func(t *testing.T) {
		group := NewGroup()

		result := group.First()
		assert.Nil(t, result)

		// Group should still be usable for adding
		group = group.Add(New().SetNumber(42))
		assert.Equal(t, 1, group.Len())

		first := group.First()
		require.NotNil(t, first)
		assert.Equal(t, 42, first.Number())
	})
}

// TestGroup_PromotedReadSurface smoke-checks the read methods promoted from
// internal/collection.Slice; that package's suite is their source of truth.
func TestGroup_PromotedReadSurface(t *testing.T) {
	t.Parallel()
	t.Run("promoted methods work through the Group facade", func(t *testing.T) {
		group := NewGroup().Add(New().SetNumber(1), New().SetNumber(2))
		assert.Equal(t, 2, group.Len())
		assert.Equal(t, 1, group.First().Number())
		assert.Equal(t, 2, group.Last().Number())
		assert.True(t, NewGroup().IsEmpty())
	})
}
