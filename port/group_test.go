package port

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustAll is a test helper that panics if All returns an error.
func (g *Group) mustAll() []*Port {
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
			g := NewGroup(tt.args.names...)
			assert.Equal(t, tt.wantLen, g.Len())
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
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name: "empty prefix is valid",
			args: args{
				prefix:     "",
				startIndex: 0,
				endIndex:   3,
			},
			wantLen: 4,
		},
		{
			name: "with prefix",
			args: args{
				prefix:     "in_",
				startIndex: 4,
				endIndex:   5,
			},
			wantLen: 2,
		},
		{
			name: "with invalid start index",
			args: args{
				prefix:     "",
				startIndex: 999,
				endIndex:   5,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewIndexedGroup(tt.args.prefix, tt.args.startIndex, tt.args.endIndex)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, g)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantLen, g.Len())
			}
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
			name: "adding to non-empty group",
			group: func() *Group {
				g, err := NewIndexedGroup("p", 1, 3)
				require.NoError(t, err)
				return g
			}(),
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
			tt.group.add(tt.args.ports...)
			if tt.assertions != nil {
				tt.assertions(t, tt.group)
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
}

func TestGroup_ForEach(t *testing.T) {
	t.Run("applies action to each port", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		count := 0
		err := group.ForEach(func(p *Port) error {
			count++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("stops on error", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		err := group.ForEach(func(p *Port) error {
			return assert.AnError
		})
		assert.Error(t, err)
	})
}

func TestGroup_ForEachIf(t *testing.T) {
	t.Run("applies action only to matching ports", func(t *testing.T) {
		group := NewGroup("p1", "p2", "special1", "special2")
		count := 0
		err := group.ForEachIf(
			func(p *Port) bool { return strings.HasPrefix(p.Name(), "special") },
			func(p *Port) error { count++; return nil },
		)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("applies action to all when predicate always true", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		count := 0
		err := group.ForEachIf(
			func(p *Port) bool { return true },
			func(p *Port) error { count++; return nil },
		)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("applies action to none when predicate always false", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		count := 0
		err := group.ForEachIf(
			func(p *Port) bool { return false },
			func(p *Port) error { count++; return nil },
		)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("stops on error", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		err := group.ForEachIf(
			func(p *Port) bool { return true },
			func(p *Port) error { return assert.AnError },
		)
		assert.Error(t, err)
	})
}

func TestGroup_AllMatch(t *testing.T) {
	t.Run("returns true when all match", func(t *testing.T) {
		group := NewGroup("p1", "p2")
		result := group.Every(func(p *Port) bool {
			return p.Name() != ""
		})
		assert.True(t, result)
	})

	t.Run("returns false when not all match", func(t *testing.T) {
		group := NewGroup("p1", "")
		result := group.Every(func(p *Port) bool {
			return p.Name() != ""
		})
		assert.False(t, result)
	})
}

func TestGroup_AnyMatch(t *testing.T) {
	t.Run("returns true when at least one matches", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		result := group.Any(func(p *Port) bool {
			return p.Name() == "p2"
		})
		assert.True(t, result)
	})

	t.Run("returns false when none match", func(t *testing.T) {
		group := NewGroup("p1", "p2")
		result := group.Any(func(p *Port) bool {
			return p.Name() == "p3"
		})
		assert.False(t, result)
	})
}

func TestGroup_CountMatch(t *testing.T) {
	t.Run("counts matching ports", func(t *testing.T) {
		group := NewGroup("a1", "a2", "b1")
		count := group.Count(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 2, count)
	})

	t.Run("returns 0 for empty group", func(t *testing.T) {
		group := NewGroup()
		count := group.Count(func(p *Port) bool {
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
}

func TestGroup_MapIf(t *testing.T) {
	t.Run("maps only matching ports", func(t *testing.T) {
		group := NewGroup("p1", "special", "p2")
		mapped := group.MapIf(
			func(p *Port) bool { return strings.HasPrefix(p.Name(), "special") },
			func(p *Port) *Port { return mustOutput("mapped_" + p.Name()) },
		)
		assert.Equal(t, 3, mapped.Len())
		assert.Equal(t, "mapped_special", mapped.Find(func(p *Port) bool {
			return strings.HasPrefix(p.Name(), "mapped_")
		}).Name())
	})

	t.Run("predicate matches none - all ports kept as-is", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		mapped := group.MapIf(
			func(p *Port) bool { return false },
			func(p *Port) *Port { return mustOutput("x") },
		)
		assert.Equal(t, 3, mapped.Len())
	})

	t.Run("predicate matches all - all ports mapped", func(t *testing.T) {
		group := NewGroup("p1", "p2")
		mapped := group.MapIf(
			func(p *Port) bool { return true },
			func(p *Port) *Port { return mustOutput("mapped_" + p.Name()) },
		)
		assert.Equal(t, 2, mapped.Len())
	})

	t.Run("nil mapper result drops the port", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		mapped := group.MapIf(
			func(p *Port) bool { return p.Name() == "p2" },
			func(p *Port) *Port { return nil },
		)
		assert.Equal(t, 2, mapped.Len()) // p2 dropped, p1 and p3 kept
	})
}

func TestGroup_Map(t *testing.T) {
	t.Run("transforms ports", func(t *testing.T) {
		group := NewGroup("p1", "p2")
		mapped := group.Map(func(p *Port) *Port {
			return mustOutput("mapped_" + p.Name())
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
}

func TestGroup_Len(t *testing.T) {
	t.Run("returns count of ports", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		assert.Equal(t, 3, group.Len())
	})

	t.Run("returns 0 for empty group", func(t *testing.T) {
		group := NewGroup()
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
}

func TestGroup_Find(t *testing.T) {
	t.Run("returns first matching port", func(t *testing.T) {
		group := NewGroup("p1", "special", "p2")
		got := group.Find(func(p *Port) bool {
			return strings.HasPrefix(p.Name(), "special")
		})
		require.NotNil(t, got)
		assert.Equal(t, "special", got.Name())
	})

	t.Run("returns nil when no port matches", func(t *testing.T) {
		group := NewGroup("p1", "p2", "p3")
		got := group.Find(func(p *Port) bool {
			return strings.HasPrefix(p.Name(), "x")
		})
		assert.Nil(t, got)
	})

	t.Run("returns nil for empty group", func(t *testing.T) {
		group := NewGroup()
		got := group.Find(func(p *Port) bool { return true })
		assert.Nil(t, got)
	})
}

func TestGroup_FirstDoesNotPoisonGroup(t *testing.T) {
	t.Run("First does not break group when empty", func(t *testing.T) {
		group := NewGroup()

		// Query first on empty group
		result := group.First()

		// Result should be nil
		assert.Nil(t, result)

		// Group should still be usable for adding
		group.add(mustOutput("p1"))
		assert.Equal(t, 1, group.Len())

		// Now First should work
		first := group.First()
		require.NotNil(t, first)
		assert.Equal(t, "p1", first.Name())
	})
}
