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

func TestGroup_Last(t *testing.T) {
	c1 := New().SetNumber(1)
	c2 := New().SetNumber(2)
	c3 := New().SetNumber(3)

	t.Run("get last from group", func(t *testing.T) {
		group := NewGroup().Add(c1, c2, c3)
		last := group.Last()
		assert.Equal(t, 3, last.Number())
	})

	t.Run("last from empty group returns nil", func(t *testing.T) {
		group := NewGroup()
		last := group.Last()
		assert.Nil(t, last)
	})
}

func TestGroup_First(t *testing.T) {
	c1 := New().SetNumber(1)
	c2 := New().SetNumber(2)

	t.Run("get first from group", func(t *testing.T) {
		group := NewGroup().Add(c1, c2)
		first := group.First()
		assert.Equal(t, 1, first.Number())
	})

	t.Run("first from empty group returns nil", func(t *testing.T) {
		group := NewGroup()
		first := group.First()
		assert.Nil(t, first)
	})
}

func TestGroup_All(t *testing.T) {
	c1 := New().SetNumber(1)
	c2 := New().SetNumber(2)

	t.Run("returns all cycles", func(t *testing.T) {
		group := NewGroup().Add(c1, c2)
		all := group.All()
		assert.Len(t, all, 2)
	})

	t.Run("returns empty slice for empty group", func(t *testing.T) {
		group := NewGroup()
		all := group.All()
		assert.Empty(t, all)
	})
}

func TestGroup_IsEmpty(t *testing.T) {
	t.Run("empty group", func(t *testing.T) {
		group := NewGroup()
		assert.True(t, group.IsEmpty())
	})

	t.Run("non-empty group", func(t *testing.T) {
		group := NewGroup().Add(New())
		assert.False(t, group.IsEmpty())
	})
}

func TestGroup_FirstDoesNotPoisonGroup(t *testing.T) {
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
