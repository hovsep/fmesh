package port

import (
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollection_AllHaveSignal(t *testing.T) {
	oneEmptyPorts := NewCollection().With(NewGroup("p1", "p2", "p3").PortsOrNil()...).PutSignals(signal.New(123))
	oneEmptyPorts.ByName("p2").Clear()

	tests := []struct {
		name  string
		ports *Collection
		want  bool
	}{
		{
			name:  "all empty",
			ports: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			want:  false,
		},
		{
			name:  "one empty",
			ports: oneEmptyPorts,
			want:  false,
		},
		{
			name:  "all set",
			ports: NewCollection().With(NewGroup("out1", "out2", "out3").PortsOrNil()...).PutSignals(signal.New(77)),
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ports.AllHaveSignals())
		})
	}
}

func TestCollection_AnyHasSignal(t *testing.T) {
	oneEmptyPorts := NewCollection().With(NewGroup("p1", "p2", "p3").PortsOrNil()...).PutSignals(signal.New(123))
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
			ports: NewCollection().With(NewGroup("p1", "p2", "p3").PortsOrNil()...),
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
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				name: "p1",
			},
			want: New("p1"),
		},
		{
			name:       "port with buffer found",
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...).PutSignals(signal.New(12)),
			args: args{
				name: "p2",
			},
			want: New("p2").WithSignals(signal.New(12)),
		},
		{
			name:       "port not found",
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				name: "p3",
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
		name  string
		ports *Collection
		args  args
		want  *Collection
	}{
		{
			name:  "single port found",
			ports: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				names: []string{"p1"},
			},
			want: NewCollection().With(New("p1")),
		},
		{
			name:  "multiple ports found",
			ports: NewCollection().With(NewGroup("p1", "p2", "p3", "p4").PortsOrNil()...),
			args: args{
				names: []string{"p1", "p2"},
			},
			want: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
		},
		{
			name:  "single port not found",
			ports: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				names: []string{"p7"},
			},
			want: NewCollection(),
		},
		{
			name:  "some ports not found",
			ports: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				names: []string{"p1", "p2", "p3"},
			},
			want: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ports.ByNames(tt.args.names...))
		})
	}
}

func TestCollection_ClearSignal(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ports := NewCollection().With(NewGroup("p1", "p2", "p3").PortsOrNil()...).PutSignals(signal.New(1), signal.New(2), signal.New(3))
		assert.True(t, ports.AllHaveSignals())
		ports.Clear()
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
				ports: NewGroup("p1", "p2").PortsOrNil(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, collection.Len(), 2)
				assert.Equal(t, collection.ByNames("p1", "p2").Len(), 2)
			},
		},
		{
			name:       "adding to non-empty collection",
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				ports: NewGroup("p3", "p4").PortsOrNil(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, collection.Len(), 4)
				assert.Equal(t, collection.ByNames("p1", "p2", "p3", "p4").Len(), 4)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.collection = tt.collection.With(tt.args.ports...)
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
			collection: NewCollection().With(
				New("src").
					WithSignalGroups(signal.NewGroup(1, 2, 3)).
					PipeTo(New("dst1"), New("dst2")),
			),
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, collection.Len(), 1)
				assert.False(t, collection.ByName("src").HasSignals())
				for _, destPort := range collection.ByName("src").Pipes().PortsOrNil() {
					assert.Equal(t, destPort.Buffer().Len(), 3)
					allPayloads, err := destPort.Buffer().AllPayloads()
					assert.NoError(t, err)
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
				destPorts: NewIndexedGroup("dest_", 1, 3).PortsOrNil(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name:       "add pipes to each port in collection",
			collection: NewCollection().With(NewIndexedGroup("p", 1, 3).PortsOrNil()...),
			args: args{
				destPorts: NewIndexedGroup("dest", 1, 5).PortsOrNil(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, collection.Len(), 3)
				for _, p := range collection.PortsOrNil() {
					assert.True(t, p.HasPipes())
					assert.Equal(t, p.Pipes().Len(), 5)
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
				assert.Equal(t, collection.Len(), 3)
			},
		},
		{
			name:       "adding to non-empty collection",
			collection: NewCollection().With(NewGroup("p1", "p2", "p3").PortsOrNil()...),
			args: args{
				prefix:     "p",
				startIndex: 4,
				endIndex:   5,
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, collection.Len(), 5)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectionAfter := tt.collection.WithIndexed(tt.args.prefix, tt.args.startIndex, tt.args.endIndex)
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
				WithIndexed("p", 1, 3).
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
