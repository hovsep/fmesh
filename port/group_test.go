package port

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGroup(t *testing.T) {
	t.Run("empty group", func(t *testing.T) {
		assert.Equal(t, 0, NewGroup().Len())
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
			inputs := NewInputGroup(tt.args.names...)
			assert.Equal(t, tt.wantLen, inputs.Len())
			for _, p := range inputs.All() {
				assert.True(t, p.IsInput())
			}

			outputs := NewOutputGroup(tt.args.names...)
			assert.Equal(t, tt.wantLen, outputs.Len())
			for _, p := range outputs.All() {
				assert.True(t, p.IsOutput())
			}
		})
	}
}

func TestNewIndexedOutputGroup(t *testing.T) {
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
			g, err := NewIndexedOutputGroup(tt.args.prefix, tt.args.startIndex, tt.args.endIndex)
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
				ports: NewOutputGroup("p1", "p2", "p3").All(),
			},
			assertions: func(t *testing.T, group *Group) {
				assert.Equal(t, 3, group.Len())
			},
		},
		{
			name: "adding to non-empty group",
			group: func() *Group {
				g, err := NewIndexedOutputGroup("p", 1, 3)
				require.NoError(t, err)
				return g
			}(),
			args: args{
				ports: NewOutputGroup("p4", "p5", "p6").All(),
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

func TestGroup_ForEach(t *testing.T) {
	t.Run("applies action to each port", func(t *testing.T) {
		group := NewOutputGroup("p1", "p2", "p3")
		count := 0
		err := group.ForEach(func(p *Port) error {
			count++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("stops on error", func(t *testing.T) {
		group := NewOutputGroup("p1", "p2", "p3")
		err := group.ForEach(func(p *Port) error {
			return assert.AnError
		})
		assert.Error(t, err)
	})
}

func TestGroup_Filter(t *testing.T) {
	t.Run("filters matching ports", func(t *testing.T) {
		group := NewOutputGroup("a1", "a2", "b1")
		filtered := group.Filter(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 2, filtered.Len())
	})
}

func TestGroup_Len(t *testing.T) {
	t.Run("returns count of ports", func(t *testing.T) {
		group := NewOutputGroup("p1", "p2", "p3")
		assert.Equal(t, 3, group.Len())
	})

	t.Run("returns 0 for empty group", func(t *testing.T) {
		group := NewGroup()
		assert.Equal(t, 0, group.Len())
	})
}

func TestGroup_First(t *testing.T) {
	t.Run("returns first port", func(t *testing.T) {
		group := NewOutputGroup("p1", "p2")
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
		group := NewOutputGroup("p1", "special", "p2")
		got := group.Find(func(p *Port) bool {
			return strings.HasPrefix(p.Name(), "special")
		})
		require.NotNil(t, got)
		assert.Equal(t, "special", got.Name())
	})

	t.Run("returns nil when no port matches", func(t *testing.T) {
		group := NewOutputGroup("p1", "p2", "p3")
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

		first := group.First()
		require.NotNil(t, first)
		assert.Equal(t, "p1", first.Name())
	})
}
