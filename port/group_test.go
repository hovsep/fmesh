package port

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustAll is a test helper that panics if All returns an error.
func (g *Group) mustAll() Ports {
	ports, err := g.All()
	if err != nil {
		panic(err)
	}
	return ports
}

func TestNewGroup(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name string
		args args
		want *Group
	}{
		{
			name: "empty group",
			args: args{
				names: nil,
			},
			want: &Group{
				chainableErr: nil,
				ports:        Ports{},
			},
		},
		{
			name: "non-empty group",
			args: args{
				names: []string{"p1", "p2"},
			},
			want: &Group{
				chainableErr: nil,
				ports: Ports{NewOutput("p1"),
					NewOutput("p2")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewGroup(tt.args.names...))
		})
	}
}

func TestNewIndexedGroup(t *testing.T) {
	type args struct {
		prefix     string
		startIndex int
		endIndex   int
	}
	tests := []struct {
		name string
		args args
		want *Group
	}{
		{
			name: "empty prefix is valid",
			args: args{
				prefix:     "",
				startIndex: 0,
				endIndex:   3,
			},
			want: NewGroup("0", "1", "2", "3"),
		},
		{
			name: "with prefix",
			args: args{
				prefix:     "in_",
				startIndex: 4,
				endIndex:   5,
			},
			want: NewGroup("in_4", "in_5"),
		},
		{
			name: "with invalid start index",
			args: args{
				prefix:     "",
				startIndex: 999,
				endIndex:   5,
			},
			want: NewGroup().WithChainableErr(ErrInvalidRangeForIndexedGroup),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewIndexedGroup(tt.args.prefix, tt.args.startIndex, tt.args.endIndex))
		})
	}
}

func TestGroup_With(t *testing.T) {
	type args struct {
		ports Ports
	}
	tests := []struct {
		name       string
		group      *Group
		args       args
		assertions func(t *testing.T, group *Group)
	}{
		{
			name:  "adding nothing to empty group",
			group: NewGroup(),
			args: args{
				ports: nil,
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Zero(t, group.Len())
			},
		},
		{
			name:  "adding to empty group",
			group: NewGroup(),
			args: args{
				ports: NewGroup("p1", "p2", "p3").mustAll(),
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 3, group.Len())
			},
		},
		{
			name:  "adding to non-empty group",
			group: NewIndexedGroup("p", 1, 3),
			args: args{
				ports: NewGroup("p4", "p5", "p6").mustAll(),
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 6, group.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupAfter := tt.group.Add(tt.args.ports...)
			if tt.assertions != nil {
				tt.assertions(t, groupAfter)
			}
		})
	}
}

func TestGroup_Without(t *testing.T) {
	t.Run("removes matching ports", func(t *testing.T) {
		group := NewGroup("a1", "a2", "b1")
		result := group.Without(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 1, result.Len())
	})

	t.Run("propagates error from source group", func(t *testing.T) {
		group := NewGroup("p1").WithChainableErr(assert.AnError)
		result := group.Without(func(p *Port) bool {
			return true
		})
		assert.True(t, result.HasChainableErr())
	})
}

func TestGroup_ForEach(t *testing.T) {
	t.Run("applies action to each port", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		count := 0
		group.ForEach(func(p *Port) error {
			count++
			return nil
		})
		assert.Equal(t, 3, count)
	})

	t.Run("stops on error and sets chainable error", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		result := group.ForEach(func(p *Port) error {
			return assert.AnError
		})
		assert.True(t, result.HasChainableErr())
	})

	t.Run("skips when group has error", func(t *testing.T) {
		group := NewGroup("p1").WithChainableErr(assert.AnError)
		visited := false
		group.ForEach(func(p *Port) error {
			visited = true
			return nil
		})
		assert.False(t, visited)
	})
}

func TestGroup_AllMatch(t *testing.T) {
	t.Run("returns true when all match", func(t *testing.T) {
		group := NewGroup("p1", "p2")
		result := group.AllMatch(func(p *Port) bool {
			return p.Name() != ""
		})
		assert.True(t, result)
	})

	t.Run("returns false when not all match", func(t *testing.T) {
		group := NewGroup("p1", "")
		result := group.AllMatch(func(p *Port) bool {
			return p.Name() != ""
		})
		assert.False(t, result)
	})

	t.Run("returns false when group has error", func(t *testing.T) {
		group := NewGroup("p1").WithChainableErr(assert.AnError)
		result := group.AllMatch(func(p *Port) bool {
			return true
		})
		assert.False(t, result)
	})
}

func TestGroup_AnyMatch(t *testing.T) {
	t.Run("returns true when at least one matches", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		result := group.AnyMatch(func(p *Port) bool {
			return p.Name() == "p2"
		})
		assert.True(t, result)
	})

	t.Run("returns false when none match", func(t *testing.T) {
		group := NewGroup("p1", "p2")
		result := group.AnyMatch(func(p *Port) bool {
			return p.Name() == "p3"
		})
		assert.False(t, result)
	})

	t.Run("returns false when group has error", func(t *testing.T) {
		group := NewGroup("p1").WithChainableErr(assert.AnError)
		result := group.AnyMatch(func(p *Port) bool {
			return true
		})
		assert.False(t, result)
	})
}

func TestGroup_CountMatch(t *testing.T) {
	t.Run("counts matching ports", func(t *testing.T) {
		group := NewGroup("a1", "a2", "b1")
		count := group.CountMatch(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 2, count)
	})

	t.Run("returns 0 for empty group", func(t *testing.T) {
		group := NewGroup()
		count := group.CountMatch(func(p *Port) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})

	t.Run("returns 0 when group has error", func(t *testing.T) {
		group := NewGroup("p1").WithChainableErr(assert.AnError)
		count := group.CountMatch(func(p *Port) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})
}

func TestGroup_Filter(t *testing.T) {
	t.Run("filters matching ports", func(t *testing.T) {
		group := NewGroup("a1", "a2", "b1")
		filtered := group.Filter(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 2, filtered.Len())
	})

	t.Run("propagates error from source group", func(t *testing.T) {
		group := NewGroup("p1").WithChainableErr(assert.AnError)
		filtered := group.Filter(func(p *Port) bool {
			return true
		})
		assert.True(t, filtered.HasChainableErr())
	})
}

func TestGroup_Map(t *testing.T) {
	t.Run("transforms ports", func(t *testing.T) {
		group := NewGroup("p1", "p2")
		mapped := group.Map(func(p *Port) *Port {
			return NewOutput("mapped_" + p.Name())
		})
		assert.Equal(t, 2, mapped.Len())
	})

	t.Run("filters out nil results", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		mapped := group.Map(func(p *Port) *Port {
			if p.Name() == "p2" {
				return nil
			}
			return p
		})
		assert.Equal(t, 2, mapped.Len())
	})

	t.Run("propagates error from source group", func(t *testing.T) {
		group := NewGroup("p1").WithChainableErr(assert.AnError)
		mapped := group.Map(func(p *Port) *Port {
			return p
		})
		assert.True(t, mapped.HasChainableErr())
	})
}

func TestGroup_Len(t *testing.T) {
	t.Run("returns count of ports", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		assert.Equal(t, 3, group.Len())
	})

	t.Run("returns 0 when group has error", func(t *testing.T) {
		group := NewGroup("p1").WithChainableErr(assert.AnError)
		assert.Equal(t, 0, group.Len())
	})
}

func TestGroup_First(t *testing.T) {
	t.Run("returns first port", func(t *testing.T) {
		group := NewGroup("p1", "p2")
		first := group.First()
		require.NotNil(t, first)
		assert.Equal(t, "p1", first.Name())
	})

	t.Run("returns nil for empty group", func(t *testing.T) {
		group := NewGroup()
		first := group.First()
		assert.Nil(t, first)
	})

	t.Run("returns nil when group has error", func(t *testing.T) {
		group := NewGroup("p1").WithChainableErr(assert.AnError)
		first := group.First()
		assert.Nil(t, first)
	})
}

func TestGroup_FirstDoesNotPoisonGroup(t *testing.T) {
	t.Run("First does not poison group when empty", func(t *testing.T) {
		group := NewGroup()

		// Query first on empty group
		result := group.First()

		// Result should be nil
		assert.Nil(t, result)

		// Group should NOT be poisoned
		assert.False(t, group.HasChainableErr())

		// Group should still be usable for adding
		group = group.Add(NewOutput("p1"))
		assert.Equal(t, 1, group.Len())
		assert.False(t, group.HasChainableErr())

		// Now First should work
		first := group.First()
		require.NotNil(t, first)
		assert.Equal(t, "p1", first.Name())
	})
}
