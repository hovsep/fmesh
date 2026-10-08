package port

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGroup(t *testing.T) {
	t.Run("empty group", func(t *testing.T) {
		assert.Equal(t, 0, newGroup().Len())
	})
}

func TestNewDirectedGroups(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name: "empty group",
			args: args{
				names: nil,
			},
			wantLen: 0,
		},
		{
			name: "non-empty group",
			args: args{
				names: []string{"p1", "p2"},
			},
			wantLen: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs := newInputGroup(tt.args.names...)
			assert.Equal(t, tt.wantLen, inputs.Len())
			assert.True(t, inputs.Every(func(p *Port) bool { return p.IsInput() }))

			outputs := newOutputGroup(tt.args.names...)
			assert.Equal(t, tt.wantLen, outputs.Len())
			assert.True(t, outputs.Every(func(p *Port) bool { return p.IsOutput() }))
		})
	}
}

func TestGroup_With(t *testing.T) {
	type args struct {
		ports []*Port
	}
	tests := []struct {
		name       string
		group      *Group
		args       args
		assertions func(t *testing.T, group *Group)
	}{
		{
			name:  "adding nothing to empty group",
			group: newGroup(),
			args: args{
				ports: nil,
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Zero(t, group.Len())
			},
		},
		{
			name:  "adding to empty group",
			group: newGroup(),
			args: args{
				ports: slices.Collect(newOutputGroup("p1", "p2", "p3").All()),
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 3, group.Len())
			},
		},
		{
			name:  "adding to non-empty group",
			group: newOutputGroup("p1", "p2", "p3"),
			args: args{
				ports: slices.Collect(newOutputGroup("p4", "p5", "p6").All()),
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 6, group.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.group.add(tt.args.ports...)
			if tt.assertions != nil {
				tt.assertions(t, tt.group)
			}
		})
	}
}

func TestGroup_All(t *testing.T) {
	t.Run("yields every port in insertion order", func(t *testing.T) {
		group := newOutputGroup("p2", "p1", "p3")
		var names []string
		for p := range group.All() {
			names = append(names, p.Name())
		}
		assert.Equal(t, []string{"p2", "p1", "p3"}, names)
	})
}

func TestGroup_AllMatch(t *testing.T) {
	t.Run("returns true when all match", func(t *testing.T) {
		group := newOutputGroup("p1", "p2")
		result := group.Every(func(p *Port) bool {
			return p.Name() != ""
		})
		assert.True(t, result)
	})

	t.Run("returns false when not all match", func(t *testing.T) {
		group := newOutputGroup("p1", "")
		result := group.Every(func(p *Port) bool {
			return p.Name() != ""
		})
		assert.False(t, result)
	})
}

func TestGroup_Any(t *testing.T) {
	t.Run("returns true when at least one matches", func(t *testing.T) {
		group := newOutputGroup("p1", "p2", "p3")
		result := group.Any(func(p *Port) bool {
			return p.Name() == "p2"
		})
		assert.True(t, result)
	})

	t.Run("returns false when none match", func(t *testing.T) {
		group := newOutputGroup("p1", "p2")
		result := group.Any(func(p *Port) bool {
			return p.Name() == "p3"
		})
		assert.False(t, result)
	})
}

func TestGroup_CountMatch(t *testing.T) {
	t.Run("counts matching ports", func(t *testing.T) {
		group := newOutputGroup("a1", "a2", "b1")
		count := group.Count(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 2, count)
	})

	t.Run("returns 0 for empty group", func(t *testing.T) {
		group := newGroup()
		count := group.Count(func(p *Port) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})
}

func TestGroup_Filter(t *testing.T) {
	t.Run("filters matching ports", func(t *testing.T) {
		group := newOutputGroup("a1", "a2", "b1")
		filtered := group.Filter(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 2, filtered.Len())
	})
}

func TestGroup_Len(t *testing.T) {
	t.Run("returns count of ports", func(t *testing.T) {
		group := newOutputGroup("p1", "p2", "p3")
		assert.Equal(t, 3, group.Len())
	})

	t.Run("returns 0 for empty group", func(t *testing.T) {
		group := newGroup()
		assert.Equal(t, 0, group.Len())
	})
}

func TestGroup_First(t *testing.T) {
	t.Run("returns first port", func(t *testing.T) {
		group := newOutputGroup("p1", "p2")
		first := group.First()
		require.NotNil(t, first)
		assert.Equal(t, "p1", first.Name())
	})

	t.Run("returns nil for empty group", func(t *testing.T) {
		group := newGroup()
		first := group.First()
		assert.Nil(t, first)
	})
}

func TestGroup_Find(t *testing.T) {
	t.Run("returns first matching port", func(t *testing.T) {
		group := newOutputGroup("p1", "special", "p2")
		got := group.Find(func(p *Port) bool {
			return strings.HasPrefix(p.Name(), "special")
		})
		require.NotNil(t, got)
		assert.Equal(t, "special", got.Name())
	})

	t.Run("returns nil when no port matches", func(t *testing.T) {
		group := newOutputGroup("p1", "p2", "p3")
		got := group.Find(func(p *Port) bool {
			return strings.HasPrefix(p.Name(), "x")
		})
		assert.Nil(t, got)
	})

	t.Run("returns nil for empty group", func(t *testing.T) {
		group := newGroup()
		got := group.Find(func(p *Port) bool { return true })
		assert.Nil(t, got)
	})
}

func TestGroup_FirstDoesNotPoisonGroup(t *testing.T) {
	t.Run("First does not break group when empty", func(t *testing.T) {
		group := newGroup()

		// Query first on empty group
		result := group.First()

		// Result should be nil
		assert.Nil(t, result)

		// Group should still be usable for adding
		group.add(mustOutput("p1"))
		assert.Equal(t, 1, group.Len())

		first := group.First()
		require.NotNil(t, first)
		assert.Equal(t, "p1", first.Name())
	})
}
