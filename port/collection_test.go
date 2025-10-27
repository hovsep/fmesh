package port

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name: "empty port found",
			collection: NewCollection().WithDefaultLabels(labels.Map{
				DirectionLabel: DirectionOut,
			}).With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				name: "p1",
			},
			want: New("p1").WithLabels(labels.Map{
				DirectionLabel: DirectionOut,
			}),
		},
		{
			name: "port with buffer found",
			collection: NewCollection().WithDefaultLabels(labels.Map{
				DirectionLabel: DirectionOut,
			}).With(NewGroup("p1", "p2").PortsOrNil()...).PutSignals(signal.New(12)),
			args: args{
				name: "p2",
			},
			want: New("p2").WithLabels(labels.Map{
				DirectionLabel: DirectionOut,
			}).WithSignals(signal.New(12)),
		},
		{
			name:       "port not found",
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				name: "p3",
			},
			want: New("").WithChainableErr(fmt.Errorf("%w, port name: %s", ErrPortNotFoundInCollection, "p3")),
		},
		{
			name:       "with chain error",
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...).WithChainableErr(errors.New("some error")),
			args: args{
				name: "p1",
			},
			want: New("").WithChainableErr(errors.New("some error")),
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
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				names: []string{"p1"},
			},
			want: NewCollection().With(New("p1")),
		},
		{
			name:       "multiple ports found",
			collection: NewCollection().With(NewGroup("p1", "p2", "p3", "p4").PortsOrNil()...),
			args: args{
				names: []string{"p1", "p2"},
			},
			want: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
		},
		{
			name:       "single port not found",
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				names: []string{"p7"},
			},
			want: NewCollection(),
		},
		{
			name:       "some ports not found",
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				names: []string{"p1", "p2", "p3"},
			},
			want: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
		},
		{
			name:       "with chain error",
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...).WithChainableErr(errors.New("some error")),
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
				ports: NewGroup("p1", "p2").PortsOrNil(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 2, collection.Len())
				assert.Equal(t, 2, collection.ByNames("p1", "p2").Len())
			},
		},
		{
			name:       "adding to non-empty collection",
			collection: NewCollection().With(NewGroup("p1", "p2").PortsOrNil()...),
			args: args{
				ports: NewGroup("p3", "p4").PortsOrNil(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 4, collection.Len())
				assert.Equal(t, 4, collection.ByNames("p1", "p2", "p3", "p4").Len())
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
					WithLabels(labels.Map{
						DirectionLabel: DirectionOut,
					}).
					WithSignalGroups(signal.NewGroup(1, 2, 3)).
					PipeTo(New("dst1").
						WithLabels(labels.Map{
							DirectionLabel: DirectionIn,
						}), New("dst2").
						WithLabels(labels.Map{
							DirectionLabel: DirectionIn,
						})),
			),
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 1, collection.Len())
				assert.False(t, collection.ByName("src").HasSignals())
				for _, destPort := range collection.ByName("src").Pipes().PortsOrNil() {
					assert.Equal(t, 3, destPort.Buffer().Len())
					allPayloads, err := destPort.AllSignalsPayloads()
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
				destPorts: NewIndexedGroup("dest_", 1, 3).PortsOrNil(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Zero(t, collection.Len())
			},
		},
		{
			name: "add pipes to each port in collection",
			collection: NewCollection().With(NewIndexedGroup("p", 1, 3).WithPortLabels(labels.Map{
				DirectionLabel: DirectionOut,
			}).PortsOrNil()...),
			args: args{
				destPorts: NewIndexedGroup("dest", 1, 5).
					WithPortLabels(labels.Map{
						DirectionLabel: DirectionIn,
					}).
					PortsOrNil(),
			},
			assertions: func(t *testing.T, collection *Collection) {
				assert.Equal(t, 3, collection.Len())
				for _, p := range collection.AllOrNil() {
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
			collection: NewCollection().With(NewGroup("p1", "p2", "p3").PortsOrNil()...),
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
