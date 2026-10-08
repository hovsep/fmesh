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
			group: NewGroup().Add(New().AddActivationResults(component.NewActivationResult("c1", component.ActivationCodeUndefined))),
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
				cycles: []*Cycle{New().AddActivationResults(component.NewActivationResult("c1", component.ActivationCodeUndefined))},
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 1, group.Len())
			},
		},
		{
			name:  "adding to existing group",
			group: NewGroup().Add(New().AddActivationResults(component.NewActivationResult("c1", component.ActivationCodeOK))),
			args: args{
				cycles: []*Cycle{New().AddActivationResults(component.NewActivationResult("c1", component.ActivationCodeUndefined))},
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

func TestGroup_Filter(t *testing.T) {
	t.Parallel()
	c1 := New().SetNumber(1).AddActivationResults(component.NewActivationResult("c1", component.ActivationCodeOK))
	c2 := New().SetNumber(2).AddActivationResults(component.NewActivationResult("c2", component.ActivationCodeUndefined))
	c3 := New().SetNumber(3).AddActivationResults(component.NewActivationResult("c3", component.ActivationCodeOK))

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

	t.Run("keeps the len limit but starts with empty meta", func(t *testing.T) {
		group := NewGroup().SetLenLimit(5).Add(c1, c2, c3)
		group.Meta().Set("k", "v")

		filtered := group.Filter(func(c *Cycle) bool {
			return c.HasActivatedComponents()
		})

		assert.Equal(t, 5, filtered.lenLimit)
		assert.False(t, filtered.Meta().Has("k"))
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
