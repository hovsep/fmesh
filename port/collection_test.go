package port

import (
	"context"
	"testing"

	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustNewCollection is a test helper that creates a collection from ports, panicking on error.
func mustNewCollection(ports ...*Port) *Collection {
	c := NewCollection()
	if err := c.Add(ports...); err != nil {
		panic(err)
	}
	return c
}

func TestCollection_AllHaveSignals(t *testing.T) {
	oneEmptyPorts := mustNewCollection(NewOutputGroup("p1", "p2", "p3").All()...)
	require.NoError(t, oneEmptyPorts.PutSignalsOnEach(signal.New(123)))
	require.NoError(t, oneEmptyPorts.ByName("p2").Clear(context.Background()))

	tests := []struct {
		name  string
		ports *Collection
		want  bool
	}{
		{
			name:  "all empty",
			ports: mustNewCollection(NewOutputGroup("p1", "p2").All()...),
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
				c := mustNewCollection(NewOutputGroup("out1", "out2", "out3").All()...)
				require.NoError(t, c.PutSignalsOnEach(signal.New(77)))
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
	oneEmptyPorts := mustNewCollection(NewOutputGroup("p1", "p2", "p3").All()...)
	require.NoError(t, oneEmptyPorts.PutSignalsOnEach(signal.New(123)))
	require.NoError(t, oneEmptyPorts.ByName("p2").Clear(context.Background()))

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
			ports: mustNewCollection(NewOutputGroup("p1", "p2", "p3").All()...),
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
			collection: mustNewCollection(NewOutputGroup("p1", "p2").All()...),
			args:       args{name: "p1"},
			wantName:   "p1",
		},
		{
			name: "port with signals found",
			collection: func() *Collection {
				c := mustNewCollection(NewOutputGroup("p1", "p2").All()...)
				require.NoError(t, c.PutSignalsOnEach(signal.New(12)))
				return c
			}(),
			args:     args{name: "p2"},
			wantName: "p2",
		},
		{
			name:       "port not found returns nil",
			collection: mustNewCollection(NewOutputGroup("p1", "p2").All()...),
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
			collection: mustNewCollection(NewOutputGroup("p1", "p2").All()...),
			args:       args{names: []string{"p1"}},
			wantLen:    1,
		},
		{
			name:       "multiple ports found",
			collection: mustNewCollection(NewOutputGroup("p1", "p2", "p3", "p4").All()...),
			args:       args{names: []string{"p1", "p2"}},
			wantLen:    2,
		},
		{
			name:       "single port not found",
			collection: mustNewCollection(NewOutputGroup("p1", "p2").All()...),
			args:       args{names: []string{"p7"}},
			wantLen:    0,
		},
		{
			name:       "some ports not found",
			collection: mustNewCollection(NewOutputGroup("p1", "p2").All()...),
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

func TestCollection_Add(t *testing.T) {
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
				ports: NewOutputGroup("p1", "p2").All(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 2, collection.Len())
				assert.Equal(t, 2, collection.ByNames("p1", "p2").Len())
			},
		},
		{
			name:       "adding to non-empty collection",
			collection: mustNewCollection(NewOutputGroup("p1", "p2").All()...),
			args: args{
				ports: NewOutputGroup("p3", "p4").All(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 4, collection.Len())
				assert.Equal(t, 4, collection.ByNames("p1", "p2", "p3", "p4").Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, tt.collection.Add(tt.args.ports...))
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
				return mustNewCollection(src)
			}(),
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 1, collection.Len())
				assert.False(t, collection.ByName("src").HasSignals())
				for _, destPort := range collection.ByName("src").Pipes().All() {
					assert.Equal(t, 3, destPort.Signals().Len())
					allPayloads := destPort.Signals().AllPayloads()
					assert.Contains(t, allPayloads, 1)
					assert.Contains(t, allPayloads, 2)
					assert.Contains(t, allPayloads, 3)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collection.Flush(context.Background())
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.collection)
			}
		})
	}
}

func TestCollection_PipeEachTo(t *testing.T) {
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
				destPorts: NewOutputGroup("dest_1", "dest_2", "dest_3").All(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name: "add pipes to each port in collection",
			collection: mustNewCollection(
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
				for _, p := range collection.All() {
					assert.True(t, p.HasPipes())
					assert.Equal(t, 5, p.Pipes().Len())
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collection.PipeEachTo(tt.args.destPorts...)
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
				c := mustNewCollection(NewOutputGroup("p1", "p2", "p3").All()...)
				require.NoError(t, c.PutSignalsOnEach(signal.New(1), signal.New(2), signal.New(3)))
				require.NoError(t, c.PutSignalsOnEach(signal.New("test")))
				return c
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

func TestCollection_Filter(t *testing.T) {
	t.Run("filters matching ports", func(t *testing.T) {
		collection := mustNewCollection(NewOutputGroup("a1", "a2", "b1").All()...)
		filtered := collection.Filter(func(p *Port) bool {
			return p.Name()[0] == 'a'
		})
		assert.Equal(t, 2, filtered.Len())
	})
}

func TestCollection_Map(t *testing.T) {
	t.Run("transforms ports", func(t *testing.T) {
		collection := mustNewCollection(NewOutputGroup("p1", "p2").All()...)
		mapped, err := collection.Map(func(p *Port) *Port {
			return mustOutput("mapped_" + p.Name())
		})
		require.NoError(t, err)
		assert.Equal(t, 2, mapped.Len())
		assert.NotNil(t, mapped.ByName("mapped_p1"))
	})

	t.Run("filters out nil results", func(t *testing.T) {
		collection := mustNewCollection(NewOutputGroup("p1", "p2", "p3").All()...)
		mapped, err := collection.Map(func(p *Port) *Port {
			if p.Name() == "p2" {
				return nil
			}
			return p
		})
		require.NoError(t, err)
		assert.Equal(t, 2, mapped.Len())
	})
}

func TestCollection_IterationOperationsDoNotPoisonCollection(t *testing.T) {
	t.Run("PutSignals does not fail on valid collection", func(t *testing.T) {
		collection := mustNewCollection(mustOutput("p1"), mustOutput("p2"))
		err := collection.PutSignalsOnEach(signal.New(42))
		require.NoError(t, err)
		assert.Equal(t, 2, collection.Len())
	})

	t.Run("Flush does not fail on valid collection", func(t *testing.T) {
		collection := mustNewCollection(mustOutput("p1"), mustOutput("p2"))
		err := collection.Flush(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 2, collection.Len())
	})

	t.Run("PipeTo does not fail on valid collection", func(t *testing.T) {
		dest := mustInput("dest")
		collection := mustNewCollection(mustOutput("p1"), mustOutput("p2"))
		err := collection.PipeEachTo(dest)
		require.NoError(t, err)
		assert.Equal(t, 2, collection.Len())
	})
}

func TestCollection_LeafMethodsDoNotPoisonCollection(t *testing.T) {
	t.Run("ByName returns nil on not found", func(t *testing.T) {
		collection := mustNewCollection(NewOutputGroup("p1", "p2").All()...)

		result := collection.ByName("nonexistent")
		assert.Nil(t, result)

		// Collection should still be usable
		assert.Equal(t, 2, collection.Len())
		p1 := collection.ByName("p1")
		require.NotNil(t, p1)
		assert.Equal(t, "p1", p1.Name())
	})
}

// TestCollection_PromotedReadSurface smoke-checks the read methods promoted
// from internal/collection.Keyed; that package's suite is their source of truth.
func TestCollection_PromotedReadSurface(t *testing.T) {
	t.Run("promoted methods work through the Collection facade", func(t *testing.T) {
		collection := mustNewCollection(NewOutputGroup("p2", "p1").All()...)
		assert.Equal(t, 2, collection.Len())
		assert.False(t, collection.IsEmpty())
		assert.Equal(t, "p1", collection.AllOrdered()[0].Name(), "traversal is name-ordered")
		assert.True(t, collection.AnyMatch(func(p *Port) bool { return p.Name() == "p2" }))
	})
}
