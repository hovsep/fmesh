package port

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustAll is a test helper that panics if All returns an error.
func (c *Collection) mustAll() Map {
	ports, err := c.All()
	if err != nil {
		panic(err)
	}
	return ports
}

func TestCollection_AllHaveSignals(t *testing.T) {
	oneEmptyPorts := NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...).PutSignals(signal.New(123))
	oneEmptyPorts.ByName("p2").Clear()

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
			name:  "all set",
			ports: NewCollection().Add(NewGroup("out1", "out2", "out3").mustAll()...).PutSignals(signal.New(77)),
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ports.AllHaveSignals())
		})
	}
}

func TestCollection_AnyHasSignals(t *testing.T) {
	oneEmptyPorts := NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...).PutSignals(signal.New(123))
	oneEmptyPorts.ByName("p2").Clear()

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
		want       *Port
	}{
		{
			name:       "empty port found",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args: args{
				name: "p1",
			},
			want: NewOutput("p1"),
		},
		{
			name:       "port with signals found",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...).PutSignals(signal.New(12)),
			args: args{
				name: "p2",
			},
			want: NewOutput("p2").PutSignals(signal.New(12)),
		},
		{
			name:       "port not found returns nil",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args: args{
				name: "p3",
			},
			want: nil,
		},
		{
			name:       "with chain error returns nil",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...).WithChainableErr(errors.New("some error")),
			args: args{
				name: "p1",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.collection.ByName(tt.args.name)
			assert.Equal(t, tt.want, got)
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
		want       *Collection
	}{
		{
			name:       "single port found",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args: args{
				names: []string{"p1"},
			},
			want: NewCollection().Add(NewOutput("p1")),
		},
		{
			name:       "multiple ports found",
			collection: NewCollection().Add(NewGroup("p1", "p2", "p3", "p4").mustAll()...),
			args: args{
				names: []string{"p1", "p2"},
			},
			want: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
		},
		{
			name:       "single port not found",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args: args{
				names: []string{"p7"},
			},
			want: NewCollection(),
		},
		{
			name:       "some ports not found",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
			args: args{
				names: []string{"p1", "p2", "p3"},
			},
			want: NewCollection().Add(NewGroup("p1", "p2").mustAll()...),
		},
		{
			name:       "with chain error",
			collection: NewCollection().Add(NewGroup("p1", "p2").mustAll()...).WithChainableErr(errors.New("some error")),
			args: args{
				names: []string{"p1", "p2"},
			},
			want: NewCollection().WithChainableErr(errors.New("some error")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.collection.ByNames(tt.args.names...))
		})
	}
}

func TestCollection_ForEachClear(t *testing.T) {
	t.Run("clear all ports signals using ForEach", func(t *testing.T) {
		ports := NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...).PutSignals(signal.New(1), signal.New(2), signal.New(3))
		assert.True(t, ports.AllHaveSignals())
		ports.ForEach(func(p *Port) error {
			return p.Clear().ChainableErr()
		})
		assert.False(t, ports.AnyHasSignals())
	})
}

func TestCollection_With(t *testing.T) {
	type args struct {
		ports Ports
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
			collection: NewCollection().Add(
				NewOutput("src").
					PutSignalGroups(signal.NewGroup(1, 2, 3)).
					PipeTo(
						NewInput("dst1"),
						NewInput("dst2"),
					),
			),
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
			tt.collection.Flush()
			if tt.assertions != nil {
				tt.assertions(t, tt.collection)
			}
		})
	}
}

func TestCollection_PipeTo(t *testing.T) {
	type args struct {
		destPorts Ports
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
				destPorts: NewIndexedGroup("dest_", 1, 3).mustAll(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name: "add pipes to each port in collection",
			collection: NewCollection().Add(
				NewOutput("p_1"),
				NewOutput("p_2"),
				NewOutput("p_3"),
			),
			args: args{
				destPorts: Ports{
					NewInput("dest_1"),
					NewInput("dest_2"),
					NewInput("dest_3"),
					NewInput("dest_4"),
					NewInput("dest_5"),
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
			tt.collection.PipeTo(tt.args.destPorts...)
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
			collectionAfter := tt.collection.AddIndexed(tt.args.prefix, tt.args.startIndex, tt.args.endIndex)
			if tt.assertions != nil {
				tt.assertions(t, collectionAfter)
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
			collection: NewCollection().
				AddIndexed("p", 1, 3).
				PutSignals(signal.New(1), signal.New(2), signal.New(3)).
				PutSignals(signal.New("test")),
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
		collection := NewCollection().Add(NewOutput("p1"))
		result := collection.Any()
		require.NotNil(t, result)
		assert.Equal(t, "p1", result.Name())
	})

	t.Run("returns nil from empty collection", func(t *testing.T) {
		collection := NewCollection()
		result := collection.Any()
		assert.Nil(t, result)
	})

	t.Run("returns nil when collection has error", func(t *testing.T) {
		collection := NewCollection().WithChainableErr(assert.AnError)
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

	t.Run("returns nil when collection has error", func(t *testing.T) {
		collection := NewCollection().Add(NewOutput("p1")).WithChainableErr(assert.AnError)
		result := collection.FindAny(func(p *Port) bool {
			return true
		})
		assert.Nil(t, result)
	})
}

func TestCollection_CountMatch(t *testing.T) {
	t.Run("counts matching ports", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("a1", "a2", "b1").mustAll()...)
		count := collection.CountMatch(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 2, count)
	})

	t.Run("returns 0 for empty collection", func(t *testing.T) {
		collection := NewCollection()
		count := collection.CountMatch(func(p *Port) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})

	t.Run("returns 0 when collection has error", func(t *testing.T) {
		collection := NewCollection().Add(NewOutput("p1")).WithChainableErr(assert.AnError)
		count := collection.CountMatch(func(p *Port) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})
}

func TestCollection_AllMatch(t *testing.T) {
	t.Run("returns true when all match", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)
		result := collection.AllMatch(func(p *Port) bool {
			return p.Name() != ""
		})
		assert.True(t, result)
	})

	t.Run("returns false when not all match", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "").mustAll()...)
		result := collection.AllMatch(func(p *Port) bool {
			return p.Name() != ""
		})
		assert.False(t, result)
	})

	t.Run("returns false when collection has error", func(t *testing.T) {
		collection := NewCollection().WithChainableErr(assert.AnError)
		result := collection.AllMatch(func(p *Port) bool {
			return true
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

	t.Run("returns false when collection has error", func(t *testing.T) {
		collection := NewCollection().WithChainableErr(assert.AnError)
		result := collection.AnyMatch(func(p *Port) bool {
			return true
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

	t.Run("propagates error from source collection", func(t *testing.T) {
		collection := NewCollection().WithChainableErr(assert.AnError)
		filtered := collection.Filter(func(p *Port) bool {
			return true
		})
		assert.True(t, filtered.HasChainableErr())
	})
}

func TestCollection_Map(t *testing.T) {
	t.Run("transforms ports", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)
		mapped := collection.Map(func(p *Port) *Port {
			return NewOutput("mapped_" + p.Name())
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

	t.Run("propagates error from source collection", func(t *testing.T) {
		collection := NewCollection().WithChainableErr(assert.AnError)
		mapped := collection.Map(func(p *Port) *Port {
			return p
		})
		assert.True(t, mapped.HasChainableErr())
	})
}

func TestCollection_Len(t *testing.T) {
	t.Run("returns count of ports", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2", "p3").mustAll()...)
		assert.Equal(t, 3, collection.Len())
	})

	t.Run("returns 0 when collection has error", func(t *testing.T) {
		collection := NewCollection().Add(NewOutput("p1")).WithChainableErr(assert.AnError)
		assert.Equal(t, 0, collection.Len())
	})
}

func TestCollection_LeafMethodsDoNotPoisonCollection(t *testing.T) {
	t.Run("ByName does not poison collection on not found", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)

		// Query for non-existent port
		result := collection.ByName("nonexistent")

		// Result should be nil
		assert.Nil(t, result)

		// Collection should NOT be poisoned
		assert.False(t, collection.HasChainableErr())
		assert.Equal(t, 2, collection.Len())

		// Collection should still be usable
		p1 := collection.ByName("p1")
		require.NotNil(t, p1)
		assert.Equal(t, "p1", p1.Name())
	})

	t.Run("Any does not poison collection when empty", func(t *testing.T) {
		collection := NewCollection()

		// Query any on empty collection
		result := collection.Any()

		// Result should be nil
		assert.Nil(t, result)

		// Collection should NOT be poisoned
		assert.False(t, collection.HasChainableErr())

		// Collection should still be usable for adding
		collection.Add(NewOutput("p1"))
		assert.Equal(t, 1, collection.Len())
	})

	t.Run("FindAny does not poison collection when no match", func(t *testing.T) {
		collection := NewCollection().Add(NewGroup("p1", "p2").mustAll()...)

		// Query with predicate that matches nothing
		result := collection.FindAny(func(p *Port) bool {
			return p.Name() == "nonexistent"
		})

		// Result should be nil
		assert.Nil(t, result)

		// Collection should NOT be poisoned
		assert.False(t, collection.HasChainableErr())
		assert.Equal(t, 2, collection.Len())

		// Subsequent FindAny should work
		found := collection.FindAny(func(p *Port) bool {
			return p.Name() == "p1"
		})
		require.NotNil(t, found)
		assert.Equal(t, "p1", found.Name())
	})
}
