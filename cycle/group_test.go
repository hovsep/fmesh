package cycle

import (
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGroup(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		group := NewGroup()
		assert.NotNil(t, group)
	})
}

func TestGroup_With(t *testing.T) {
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
			group: NewGroup().With(New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))),
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
				cycles: []*Cycle{New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))},
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 1, group.Len())
			},
		},
		{
			name:  "adding to existing group",
			group: NewGroup().With(New().WithActivationResults(component.NewActivationResult("c1").SetActivated(true))),
			args: args{
				cycles: []*Cycle{New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))},
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 2, group.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupAfter := tt.group.With(tt.args.cycles...)
			if tt.assertions != nil {
				tt.assertions(t, groupAfter)
			}
		})
	}
}

func TestGroup_Without(t *testing.T) {
	c1 := New().WithNumber(1)
	c2 := New().WithNumber(2)
	c3 := New().WithNumber(3)

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
			group: NewGroup().With(c1, c2, c3),
			predicate: func(c *Cycle) bool {
				return c.Number() == 2
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 2, group.Len())
			},
		},
		{
			name:  "remove odd numbered cycles",
			group: NewGroup().With(c1, c2, c3),
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

func TestGroup_ForEach(t *testing.T) {
	c1 := New().WithNumber(1)
	c2 := New().WithNumber(2)
	c3 := New().WithNumber(3)

	t.Run("applies action to all cycles", func(t *testing.T) {
		group := NewGroup().With(c1, c2, c3)
		count := 0
		group.ForEach(func(c *Cycle) {
			count++
		})
		assert.Equal(t, 3, count)
	})

	t.Run("empty group", func(t *testing.T) {
		group := NewGroup()
		count := 0
		group.ForEach(func(c *Cycle) {
			count++
		})
		assert.Equal(t, 0, count)
	})
}

func TestGroup_Last(t *testing.T) {
	c1 := New().WithNumber(1)
	c2 := New().WithNumber(2)
	c3 := New().WithNumber(3)

	t.Run("get last from group", func(t *testing.T) {
		group := NewGroup().With(c1, c2, c3)
		last := group.Last()
		assert.Equal(t, 3, last.Number())
	})

	t.Run("last from empty group returns error", func(t *testing.T) {
		group := NewGroup()
		last := group.Last()
		assert.True(t, last.HasChainableErr())
	})
}

func TestGroup_First(t *testing.T) {
	c1 := New().WithNumber(1)
	c2 := New().WithNumber(2)

	t.Run("get first from group", func(t *testing.T) {
		group := NewGroup().With(c1, c2)
		first := group.First()
		assert.Equal(t, 1, first.Number())
	})

	t.Run("first from empty group returns error", func(t *testing.T) {
		group := NewGroup()
		first := group.First()
		assert.True(t, first.HasChainableErr())
	})
}

func TestGroup_All(t *testing.T) {
	c1 := New().WithNumber(1)
	c2 := New().WithNumber(2)

	t.Run("returns all cycles", func(t *testing.T) {
		group := NewGroup().With(c1, c2)
		all, err := group.All()
		require.NoError(t, err)
		assert.Len(t, all, 2)
	})

	t.Run("returns empty slice for empty group", func(t *testing.T) {
		group := NewGroup()
		all, err := group.All()
		require.NoError(t, err)
		assert.Empty(t, all)
	})
}

func TestGroup_AllMatch(t *testing.T) {
	c1 := New().WithActivationResults(component.NewActivationResult("c1").SetActivated(true))
	c2 := New().WithActivationResults(component.NewActivationResult("c2").SetActivated(true))
	c3 := New().WithActivationResults(component.NewActivationResult("c3").SetActivated(false))

	t.Run("all match", func(t *testing.T) {
		group := NewGroup().With(c1, c2)
		result := group.AllMatch(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})
		assert.True(t, result)
	})

	t.Run("not all match", func(t *testing.T) {
		group := NewGroup().With(c1, c3)
		result := group.AllMatch(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})
		assert.False(t, result)
	})

	t.Run("empty group returns true", func(t *testing.T) {
		group := NewGroup()
		result := group.AllMatch(func(c *Cycle) bool {
			return false
		})
		assert.True(t, result)
	})
}

func TestGroup_AnyMatch(t *testing.T) {
	c1 := New().WithActivationResults(component.NewActivationResult("c1").SetActivated(true))
	c2 := New().WithActivationResults(component.NewActivationResult("c2").SetActivated(false))

	t.Run("at least one matches", func(t *testing.T) {
		group := NewGroup().With(c1, c2)
		result := group.AnyMatch(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})
		assert.True(t, result)
	})

	t.Run("none match", func(t *testing.T) {
		group := NewGroup().With(c2)
		result := group.AnyMatch(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})
		assert.False(t, result)
	})

	t.Run("empty group returns false", func(t *testing.T) {
		group := NewGroup()
		result := group.AnyMatch(func(c *Cycle) bool {
			return true
		})
		assert.False(t, result)
	})
}

func TestGroup_CountMatch(t *testing.T) {
	c1 := New().WithNumber(1).WithActivationResults(component.NewActivationResult("c1").SetActivated(true))
	c2 := New().WithNumber(2).WithActivationResults(component.NewActivationResult("c2").SetActivated(false))
	c3 := New().WithNumber(3).WithActivationResults(component.NewActivationResult("c3").SetActivated(true))

	t.Run("counts matching cycles", func(t *testing.T) {
		group := NewGroup().With(c1, c2, c3)
		count := group.CountMatch(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})
		assert.Equal(t, 2, count)
	})

	t.Run("no matches", func(t *testing.T) {
		group := NewGroup().With(c2)
		count := group.CountMatch(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})
		assert.Equal(t, 0, count)
	})
}

func TestGroup_Filter(t *testing.T) {
	c1 := New().WithNumber(1).WithActivationResults(component.NewActivationResult("c1").SetActivated(true))
	c2 := New().WithNumber(2).WithActivationResults(component.NewActivationResult("c2").SetActivated(false))
	c3 := New().WithNumber(3).WithActivationResults(component.NewActivationResult("c3").SetActivated(true))

	t.Run("filters matching cycles", func(t *testing.T) {
		group := NewGroup().With(c1, c2, c3)
		filtered := group.Filter(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})
		assert.Equal(t, 2, filtered.Len())
	})

	t.Run("no matches returns empty group", func(t *testing.T) {
		group := NewGroup().With(c2)
		filtered := group.Filter(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})
		assert.Equal(t, 0, filtered.Len())
	})
}

func TestGroup_Map(t *testing.T) {
	c1 := New().WithNumber(1)
	c2 := New().WithNumber(2)

	t.Run("transforms all cycles", func(t *testing.T) {
		group := NewGroup().With(c1, c2)
		mapped := group.Map(func(c *Cycle) *Cycle {
			return c.WithNumber(c.Number() * 10)
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

func TestGroup_IsEmpty(t *testing.T) {
	t.Run("empty group", func(t *testing.T) {
		group := NewGroup()
		assert.True(t, group.IsEmpty())
	})

	t.Run("non-empty group", func(t *testing.T) {
		group := NewGroup().With(New())
		assert.False(t, group.IsEmpty())
	})
}

func TestGroup_ChainableErr(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		group := NewGroup().WithChainableErr(assert.AnError)
		assert.True(t, group.HasChainableErr())
		assert.Equal(t, assert.AnError, group.ChainableErr())
	})

	t.Run("without error", func(t *testing.T) {
		group := NewGroup()
		assert.False(t, group.HasChainableErr())
		assert.NoError(t, group.ChainableErr())
	})
}
