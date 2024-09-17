package port

import (
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollection_AllHaveSignal(t *testing.T) {
	oneEmptyPorts := NewCollection().Add(NewGroup("p1", "p2", "p3")...)
	oneEmptyPorts.PutSignals(signal.New(123))
	oneEmptyPorts.ByName("p2").ClearSignals()

	allWithSignalPorts := NewCollection().Add(NewGroup("out1", "out2", "out3")...)
	allWithSignalPorts.PutSignals(signal.New(77))

	tests := []struct {
		name  string
		ports Collection
		want  bool
	}{
		{
			name:  "all empty",
			ports: NewCollection().Add(NewGroup("p1", "p2")...),
			want:  false,
		},
		{
			name:  "one empty",
			ports: oneEmptyPorts,
			want:  false,
		},
		{
			name:  "all set",
			ports: allWithSignalPorts,
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
	oneEmptyPorts := NewCollection().Add(NewGroup("p1", "p2", "p3")...)
	oneEmptyPorts.PutSignals(signal.New(123))
	oneEmptyPorts.ByName("p2").ClearSignals()

	tests := []struct {
		name  string
		ports Collection
		want  bool
	}{
		{
			name:  "one empty",
			ports: oneEmptyPorts,
			want:  true,
		},
		{
			name:  "all empty",
			ports: NewCollection().Add(NewGroup("p1", "p2", "p3")...),
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
	portsWithSignals := NewCollection().Add(NewGroup("p1", "p2")...)
	portsWithSignals.PutSignals(signal.New(12))

	type args struct {
		name string
	}
	tests := []struct {
		name  string
		ports Collection
		args  args
		want  *Port
	}{
		{
			name:  "empty port found",
			ports: NewCollection().Add(NewGroup("p1", "p2")...),
			args: args{
				name: "p1",
			},
			want: &Port{name: "p1", pipes: Group{}, signals: signal.Group{}},
		},
		{
			name:  "port with signals found",
			ports: portsWithSignals,
			args: args{
				name: "p2",
			},
			want: &Port{
				name:    "p2",
				signals: signal.NewGroup(12),
				pipes:   Group{},
			},
		},
		{
			name:  "port not found",
			ports: NewCollection().Add(NewGroup("p1", "p2")...),
			args: args{
				name: "p3",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ports.ByName(tt.args.name))
		})
	}
}

func TestCollection_ByNames(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name  string
		ports Collection
		args  args
		want  Collection
	}{
		{
			name:  "single port found",
			ports: NewCollection().Add(NewGroup("p1", "p2")...),
			args: args{
				names: []string{"p1"},
			},
			want: Collection{
				"p1": &Port{
					name:    "p1",
					pipes:   Group{},
					signals: signal.Group{},
				},
			},
		},
		{
			name:  "multiple ports found",
			ports: NewCollection().Add(NewGroup("p1", "p2")...),
			args: args{
				names: []string{"p1", "p2"},
			},
			want: Collection{
				"p1": &Port{name: "p1", pipes: Group{}, signals: signal.Group{}},
				"p2": &Port{name: "p2", pipes: Group{}, signals: signal.Group{}},
			},
		},
		{
			name:  "single port not found",
			ports: NewCollection().Add(NewGroup("p1", "p2")...),
			args: args{
				names: []string{"p7"},
			},
			want: Collection{},
		},
		{
			name:  "some ports not found",
			ports: NewCollection().Add(NewGroup("p1", "p2")...),
			args: args{
				names: []string{"p1", "p2", "p3"},
			},
			want: Collection{
				"p1": &Port{name: "p1", pipes: Group{}, signals: signal.Group{}},
				"p2": &Port{name: "p2", pipes: Group{}, signals: signal.Group{}},
			},
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
		ports := NewCollection().Add(NewGroup("p1", "p2", "p3")...)
		ports.PutSignals(signal.NewGroup(1, 2, 3)...)
		assert.True(t, ports.AllHaveSignals())
		ports.ClearSignals()
		assert.False(t, ports.AnyHasSignals())
	})
}

func TestCollection_Add(t *testing.T) {
	type args struct {
		ports []*Port
	}
	tests := []struct {
		name       string
		collection Collection
		args       args
		assertions func(t *testing.T, collection Collection)
	}{
		{
			name:       "adding nothing to empty collection",
			collection: NewCollection(),
			args: args{
				ports: nil,
			},
			assertions: func(t *testing.T, collection Collection) {
				assert.Len(t, collection, 0)
			},
		},
		{
			name:       "adding to empty collection",
			collection: NewCollection(),
			args: args{
				ports: NewGroup("p1", "p2"),
			},
			assertions: func(t *testing.T, collection Collection) {
				assert.Len(t, collection, 2)
				assert.Len(t, collection.ByNames("p1", "p2"), 2)
			},
		},
		{
			name:       "adding to existing collection",
			collection: NewCollection().Add(NewGroup("p1", "p2")...),
			args: args{
				ports: NewGroup("p3", "p4"),
			},
			assertions: func(t *testing.T, collection Collection) {
				assert.Len(t, collection, 4)
				assert.Len(t, collection.ByNames("p1", "p2", "p3", "p4"), 4)
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
