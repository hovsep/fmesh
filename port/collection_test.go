package port

import (
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestCollection_AllHaveSignal(t *testing.T) {
	oneEmptyPorts := NewPortsCollection().Add(NewPortGroup("p1", "p2", "p3")...)
	oneEmptyPorts.PutSignal(signal.New(123))
	oneEmptyPorts.ByName("p2").ClearSignal()

	allWithSignalPorts := NewPortsCollection().Add(NewPortGroup("out1", "out2", "out3")...)
	allWithSignalPorts.PutSignal(signal.New(77))

	allWithEmptySignalPorts := NewPortsCollection().Add(NewPortGroup("in1", "in2", "in3")...)
	allWithEmptySignalPorts.PutSignal(signal.New())

	tests := []struct {
		name  string
		ports Collection
		want  bool
	}{
		{
			name:  "all empty",
			ports: NewPortsCollection().Add(NewPortGroup("p1", "p2")...),
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
		{
			name:  "all set with empty signals",
			ports: allWithEmptySignalPorts,
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ports.AllHaveSignal(); got != tt.want {
				t.Errorf("AllHaveSignal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollection_AnyHasSignal(t *testing.T) {
	oneEmptyPorts := NewPortsCollection().Add(NewPortGroup("p1", "p2", "p3")...)
	oneEmptyPorts.PutSignal(signal.New(123))
	oneEmptyPorts.ByName("p2").ClearSignal()

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
			ports: NewPortsCollection().Add(NewPortGroup("p1", "p2", "p3")...),
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ports.AnyHasSignal(); got != tt.want {
				t.Errorf("AnyHasSignal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollection_ByName(t *testing.T) {
	portsWithSignals := NewPortsCollection().Add(NewPortGroup("p1", "p2")...)
	portsWithSignals.PutSignal(signal.New(12))

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
			ports: NewPortsCollection().Add(NewPortGroup("p1", "p2")...),
			args: args{
				name: "p1",
			},
			want: &Port{name: "p1", pipes: Group{}},
		},
		{
			name:  "port with signal found",
			ports: portsWithSignals,
			args: args{
				name: "p2",
			},
			want: &Port{
				name:   "p2",
				signal: signal.New(12),
				pipes:  Group{},
			},
		},
		{
			name:  "port not found",
			ports: NewPortsCollection().Add(NewPortGroup("p1", "p2")...),
			args: args{
				name: "p3",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ports.ByName(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ByName() = %v, want %v", got, tt.want)
			}
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
			ports: NewPortsCollection().Add(NewPortGroup("p1", "p2")...),
			args: args{
				names: []string{"p1"},
			},
			want: Collection{
				"p1": &Port{
					name:  "p1",
					pipes: Group{},
				},
			},
		},
		{
			name:  "multiple ports found",
			ports: NewPortsCollection().Add(NewPortGroup("p1", "p2")...),
			args: args{
				names: []string{"p1", "p2"},
			},
			want: Collection{
				"p1": &Port{name: "p1", pipes: Group{}},
				"p2": &Port{name: "p2", pipes: Group{}},
			},
		},
		{
			name:  "single port not found",
			ports: NewPortsCollection().Add(NewPortGroup("p1", "p2")...),
			args: args{
				names: []string{"p7"},
			},
			want: Collection{},
		},
		{
			name:  "some ports not found",
			ports: NewPortsCollection().Add(NewPortGroup("p1", "p2")...),
			args: args{
				names: []string{"p1", "p2", "p3"},
			},
			want: Collection{
				"p1": &Port{name: "p1", pipes: Group{}},
				"p2": &Port{name: "p2", pipes: Group{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ports.ByNames(tt.args.names...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ByNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollection_ClearSignal(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ports := NewPortsCollection().Add(NewPortGroup("p1", "p2", "p3")...)
		ports.PutSignal(signal.New(1, 2, 3))
		assert.True(t, ports.AllHaveSignal())
		ports.ClearSignal()
		assert.False(t, ports.AnyHasSignal())
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
			collection: NewPortsCollection(),
			args: args{
				ports: nil,
			},
			assertions: func(t *testing.T, collection Collection) {
				assert.Len(t, collection, 0)
			},
		},
		{
			name:       "adding to empty collection",
			collection: NewPortsCollection(),
			args: args{
				ports: NewPortGroup("p1", "p2"),
			},
			assertions: func(t *testing.T, collection Collection) {
				assert.Len(t, collection, 2)
				assert.Len(t, collection.ByNames("p1", "p2"), 2)
			},
		},
		{
			name:       "adding to existing collection",
			collection: NewPortsCollection().Add(NewPortGroup("p1", "p2")...),
			args: args{
				ports: NewPortGroup("p3", "p4"),
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
