package port

import (
	"testing"

	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustAll is a test helper that panics if All returns an error.
func (c *Collection) mustAll() map[string]*Port {
	ports, err := c.All()
	if err != nil {
		panic(err)
	}
	return ports
}

func TestCollection_AllHaveSignals(t *testing.T) {
	oneEmptyPorts := NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...)
	require.NoError(t, oneEmptyPorts.PutSignals(signal.New(123)))
	require.NoError(t, oneEmptyPorts.ByName("p2").Clear())

	tests := []struct {
		name  string
		ports *Collection
		want  bool
	}{
		{
			name:  "all empty",
			ports: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			want:  false,
		},
		{
			name:  "one empty",
			ports: oneEmptyPorts,
			want:  false,
		},
		{
			name: "all set",
			ports: func() *Collection {
				c := NewCollection().Add(NewGroup("out1", "out2", "out3").mustAll()...)
				require.NoError(t, c.PutSignals(signal.New(77)))
				return c
			}(),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ports.AllHaveSignals())
		})
	}
}

func TestCollection_AnyHasSignals(t *testing.T) {
	oneEmptyPorts := NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...)
	require.NoError(t, oneEmptyPorts.PutSignals(signal.New(123)))
	require.NoError(t, oneEmptyPorts.ByName("p2").Clear())

	tests := []struct {
		name  string
		ports *Collection
		want  bool
	}{
		{
			name:  "one empty",
			ports: oneEmptyPorts,
			want:  true,
		},
		{
			name:  "all empty",
			ports: NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...),
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ports.AnyHasSignals())
		})
	}
}

func TestCollection_ByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		collection *Collection
		args       args
		wantName   string
		wantNil    bool
	}{
		{
			name:       "empty port found",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args:       args{name: "p1"},
			wantName:   "p1",
		},
		{
			name: "port with signals found",
			collection: func() *Collection {
				c := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)
				require.NoError(t, c.PutSignals(signal.New(12)))
				return c
			}(),
			args:     args{name: "p2"},
			wantName: "p2",
		},
		{
			name:       "port not found returns nil",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args:       args{name: "p3"},
			wantNil:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.collection.ByName(tt.args.name)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.wantName, got.Name())
			}
		})
	}
}

func TestCollection_ByNames(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name       string
		collection *Collection
		args       args
		wantLen    int
	}{
		{
			name:       "single port found",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args:       args{names: []string{"p1"}},
			wantLen:    1,
		},
		{
			name:       "multiple ports found",
			collection: NewCollection().Add(NewGroup("p1", "p2", "p3", "p4").mustAll()...),
			args:       args{names: []string{"p1", "p2"}},
			wantLen:    2,
		},
		{
			name:       "single port not found",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args:       args{names: []string{"p7"}},
			wantLen:    0,
		},
		{
			name:       "some ports not found",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args:       args{names: []string{"p1", "p2", "p3"}},
			wantLen:    2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.collection.ByNames(tt.args.names...)
			assert.Equal(t, tt.wantLen, result.Len())
		})
	}
}

func TestCollection_ForEachClear(t *testing.T) {
	t.Run("clear all ports signals using ForEach", func(t *testing.T) {
		ports := NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...)
		require.NoError(t, ports.PutSignals(signal.New(1), signal.New(2), signal.New(3)))
		assert.True(t, ports.AllHaveSignals())
		err := ports.ForEach(func(p *Port) error {
			return p.Clear()
		})
		require.NoError(t, err)
		assert.False(t, ports.AnyHasSignals())
	})
}

func TestCollection_With(t *testing.T) {
	type args struct {
		ports []*Port
	}
	tests := []struct {
		name       string
		collection *Collection
		args       args
		assertions func(t *testing.T, collection *Collection)
	}{
		{
			name:       "adding nothing to empty collection",
			collection: NewCollection(),
			args: args{
				ports: nil,
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name:       "adding to empty collection",
			collection: NewCollection(),
			args: args{
				ports: NewGroup("p1", "p2").mustAll(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 2, collection.Len())
				assert.Equal(t, 2, collection.ByNames("p1", "p2").Len())
			},
		},
		{
			name:       "adding to non-empty collection",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args: args{
				ports: NewGroup("p3", "p4").mustAll(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 4, collection.Len())
				assert.Equal(t, 4, collection.ByNames("p1", "p2", "p3", "p4").Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.collection = tt.collection.Add(tt.args.ports...)
			if tt.assertions != nil {
				tt.assertions(t, tt.collection)
			}
		})
	}
}

func TestCollection_Flush(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		assertions func(t *testing.T, collection *Collection)
	}{
		{
			name:       "empty collection",
			collection: NewCollection(),
			assertions: func(t *testing.T, collection *Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name: "all ports in collection are flushed",
			collection: func() *Collection {
				dst1 := mustInput("dst1")
				dst2 := mustInput("dst2")
				src := mustOutput("src")
				require.NoError(t, src.PutSignalGroups(signal.NewGroup(1, 2, 3)))
				require.NoError(t, src.PipeTo(dst1, dst2))
				return NewCollection().Add(src)
			}(),
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 1, collection.Len())
				assert.False(t, collection.ByName("src").HasSignals())
				for _, destPort := range collection.ByName("src").Pipes().mustAll() {
					assert.Equal(t, 3, destPort.Signals().Len())
					allPayloads, err := destPort.Signals().AllPayloads()
					require.NoError(t, err)
					assert.Contains(t, allPayloads, 1)
					assert.Contains(t, allPayloads, 2)
					assert.Contains(t, allPayloads, 3)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collection.Flush()
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.collection)
			}
		})
	}
}

func TestCollection_PipeTo(t *testing.T) {
	type args struct {
		destPorts []*Port
	}
	tests := []struct {
		name       string
		collection *Collection
		args       args
		assertions func(t *testing.T, collection *Collection)
	}{
		{
			name:       "empty collection",
			collection: NewCollection(),
			args: args{
				destPorts: func() []*Port {
					g, err := NewIndexedGroup("dest_", 1, 3)
					require.NoError(t, err)
					return g.mustAll()
				}(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name: "add pipes to each port in collection",
			collection: NewCollection().Add(
				mustOutput("p_1"),
				mustOutput("p_2"),
				mustOutput("p_3"),
			),
			args: args{
				destPorts: []*Port{
					mustInput("dest_1"),
					mustInput("dest_2"),
					mustInput("dest_3"),
					mustInput("dest_4"),
					mustInput("dest_5"),
				},
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 3, collection.Len())
				for _, p := range collection.mustAll() {
					assert.True(t, p.HasPipes())
					assert.Equal(t, 5, p.Pipes().Len())
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collection.PipeTo(tt.args.destPorts...)
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.collection)
			}
		})
	}
}

func TestCollection_WithIndexed(t *testing.T) {
	type args struct {
		prefix     string
		startIndex int
		endIndex   int
	}
	tests := []struct {
		name       string
		collection *Collection
		args       args
		assertions func(t *testing.T, collection *Collection)
	}{
		{
			name:       "adding to empty collection",
			collection: NewCollection(),
			args: args{
				prefix:     "p",
				startIndex: 1,
				endIndex:   3,
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 3, collection.Len())
			},
		},
		{
			name:       "adding to non-empty collection",
			collection: NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...),
			args: args{
				prefix:     "p",
				startIndex: 4,
				endIndex:   5,
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 5, collection.Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collection.AddIndexed(tt.args.prefix, tt.args.startIndex, tt.args.endIndex)
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.collection)
			}
		})
	}
}

func TestCollection_Signals(t *testing.T) {
	tests := []struct {
		name       string
		collection *Collection
		want       *signal.Group
	}{
		{
			name:       "empty collection",
			collection: NewCollection(),
			want:       signal.NewGroup(),
		},
		{
			name: "non-empty collection",
			collection: func() *Collection {
				c := NewCollection().AddIndexed("p", 1, 3)
				// AddIndexed now returns error; we ignore it since it mutates in place
				_ = c
				c2 := NewCollection()
				require.NoError(t, c2.AddIndexed("p", 1, 3))
				require.NoError(t, c2.PutSignals(signal.New(1), signal.New(2), signal.New(3)))
				require.NoError(t, c2.PutSignals(signal.New("test")))
				return c2
			}(),
			want: signal.NewGroup(1, 2, 3, "test", 1, 2, 3, "test", 1, 2, 3, "test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.Signals())
		})
	}
}

func TestCollection_Any(t *testing.T) {
	t.Run("returns port from non-empty collection", func(t *testing.T) {
		collection := NewCollection().Add(mustOutput("p1"))
		result := collection.Any()
		require.NotNil(t, result)
		assert.Equal(t, "p1", result.Name())
	})

	t.Run("returns nil from empty collection", func(t *testing.T) {
		collection := NewCollection()
		result := collection.Any()
		assert.Nil(t, result)
	})
}

func TestCollection_FindAny(t *testing.T) {
	t.Run("finds matching port", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2", "target").mustAll()...)
		result := collection.FindAny(func(p *Port) bool {
			return p.Name() == "target"
		})
		require.NotNil(t, result)
		assert.Equal(t, "target", result.Name())
	})

	t.Run("returns nil when no match", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)
		result := collection.FindAny(func(p *Port) bool {
			return p.Name() == "p3"
		})
		assert.Nil(t, result)
	})
}

func TestCollection_CountMatch(t *testing.T) {
	t.Run("counts matching ports", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("a1", "a2", "b1").mustAll()...)
		count := collection.Count(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 2, count)
	})

	t.Run("returns 0 for empty collection", func(t *testing.T) {
		collection := NewCollection()
		count := collection.Count(func(p *Port) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})
}

func TestCollection_AllMatch(t *testing.T) {
	t.Run("returns true when all match", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)
		result := collection.Every(func(p *Port) bool {
			return p.Name() != ""
		})
		assert.True(t, result)
	})

	t.Run("returns false when not all match", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "").mustAll()...)
		result := collection.Every(func(p *Port) bool {
			return p.Name() != ""
		})
		assert.False(t, result)
	})
}

func TestCollection_AnyMatch(t *testing.T) {
	t.Run("returns true when at least one matches", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "target").mustAll()...)
		result := collection.AnyMatch(func(p *Port) bool {
			return p.Name() == "target"
		})
		assert.True(t, result)
	})

	t.Run("returns false when none match", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)
		result := collection.AnyMatch(func(p *Port) bool {
			return p.Name() == "nonexistent"
		})
		assert.False(t, result)
	})
}

func TestCollection_Filter(t *testing.T) {
	t.Run("filters matching ports", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("a1", "a2", "b1").mustAll()...)
		filtered := collection.Filter(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 2, filtered.Len())
	})
}

func TestCollection_Map(t *testing.T) {
	t.Run("transforms ports", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)
		mapped := collection.Map(func(p *Port) *Port {
			return mustOutput("mapped_" + p.Name())
		})
		assert.Equal(t, 2, mapped.Len())
		assert.NotNil(t, mapped.ByName("mapped_p1"))
	})

	t.Run("filters out nil results", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...)
		mapped := collection.Map(func(p *Port) *Port {
			if p.Name() == "p2" {
				return nil
			}
			return p
		})
		assert.Equal(t, 2, mapped.Len())
	})
}

func TestCollection_Len(t *testing.T) {
	t.Run("returns count of ports", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...)
		assert.Equal(t, 3, collection.Len())
	})

	t.Run("returns 0 for empty collection", func(t *testing.T) {
		collection := NewCollection()
		assert.Equal(t, 0, collection.Len())
	})
}

func TestCollection_IterationOperationsDoNotPoisonCollection(t *testing.T) {
	t.Run("PutSignals does not fail on valid collection", func(t *testing.T) {
		collection := NewCollection().Add(mustOutput("p1"), mustOutput("p2"))
		err := collection.PutSignals(signal.New(42))
		require.NoError(t, err)
		assert.Equal(t, 2, collection.Len())
	})

	t.Run("Flush does not fail on valid collection", func(t *testing.T) {
		collection := NewCollection().Add(mustOutput("p1"), mustOutput("p2"))
		err := collection.Flush()
		require.NoError(t, err)
		assert.Equal(t, 2, collection.Len())
	})

	t.Run("PipeTo does not fail on valid collection", func(t *testing.T) {
		dest := mustInput("dest")
		collection := NewCollection().Add(mustOutput("p1"), mustOutput("p2"))
		err := collection.PipeTo(dest)
		require.NoError(t, err)
		assert.Equal(t, 2, collection.Len())
	})
}

func TestCollection_LeafMethodsDoNotPoisonCollection(t *testing.T) {
	t.Run("ByName returns nil on not found", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)

		result := collection.ByName("nonexistent")
		assert.Nil(t, result)

		// Collection should still be usable
		assert.Equal(t, 2, collection.Len())
		p1 := collection.ByName("p1")
		require.NotNil(t, p1)
		assert.Equal(t, "p1", p1.Name())
	})

	t.Run("Any returns nil when empty", func(t *testing.T) {
		collection := NewCollection()

		result := collection.Any()
		assert.Nil(t, result)

		// Collection should still be usable for adding
		collection.Add(mustOutput("p1"))
		assert.Equal(t, 1, collection.Len())
	})

	t.Run("FindAny returns nil when no match", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)

		result := collection.FindAny(func(p *Port) bool {
			return p.Name() == "nonexistent"
		})
		assert.Nil(t, result)

		// Collection should still be usable
		assert.Equal(t, 2, collection.Len())
		found := collection.FindAny(func(p *Port) bool {
			return p.Name() == "p1"
		})
		require.NotNil(t, found)
		assert.Equal(t, "p1", found.Name())
	})
}
