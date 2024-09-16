package port

import (
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestPort_HasSignal(t *testing.T) {
	portWithSignal := NewPort("portWithSignal")
	portWithSignal.PutSignal(signal.New(123))

	portWithEmptySignal := NewPort("portWithEmptySignal")
	portWithEmptySignal.PutSignal(signal.New())

	tests := []struct {
		name string
		port *Port
		want bool
	}{
		{
			name: "empty port",
			port: NewPort("emptyPort"),
			want: false,
		},
		{
			name: "port has normal signal",
			port: portWithSignal,
			want: true,
		},
		{
			name: "port has empty signal",
			port: portWithEmptySignal,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.port.HasSignal(); got != tt.want {
				t.Errorf("HasSignal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPort_Signal(t *testing.T) {
	portWithSignal := NewPort("portWithSignal")
	portWithSignal.PutSignal(signal.New(123))

	portWithEmptySignal := NewPort("portWithEmptySignal")
	portWithEmptySignal.PutSignal(signal.New())

	tests := []struct {
		name string
		port *Port
		want *signal.Signal
	}{
		{
			name: "no signal",
			port: NewPort("noSignal"),
			want: nil,
		},
		{
			name: "with signal",
			port: portWithSignal,
			want: signal.New(123),
		},
		{
			name: "with empty signal",
			port: portWithEmptySignal,
			want: signal.New(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.port.Signal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Signal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPort_ClearSignal(t *testing.T) {
	portWithSignal := NewPort("portWithSignal")
	portWithSignal.PutSignal(signal.New(111))

	tests := []struct {
		name   string
		before *Port
		after  *Port
	}{
		{
			name:   "happy path",
			before: portWithSignal,
			after:  &Port{name: "portWithSignal", pipes: Collection{}},
		},
		{
			name:   "cleaning empty port",
			before: NewPort("emptyPort"),
			after:  &Port{name: "emptyPort", pipes: Collection{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before.ClearSignal()
			if !reflect.DeepEqual(tt.before, tt.after) {
				t.Errorf("ClearSignal() = %v, want %v", tt.before, tt.after)
			}
		})
	}
}

func TestPort_PipeTo(t *testing.T) {
	p1, p2, p3, p4 := NewPort("p1"), NewPort("p2"), NewPort("p3"), NewPort("p4")

	type args struct {
		toPorts []*Port
	}
	tests := []struct {
		name   string
		before *Port
		after  *Port
		args   args
	}{
		{
			name:   "happy path",
			before: p1,
			after: &Port{
				name:  "p1",
				pipes: NewPortsCollection().Add(p2, p3),
			},
			args: args{
				toPorts: []*Port{p2, p3},
			},
		},
		{
			name:   "invalid ports are ignored",
			before: p4,
			after: &Port{
				name:  "p4",
				pipes: NewPortsCollection().Add(p2),
			},
			args: args{
				toPorts: []*Port{p2, nil},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before.PipeTo(tt.args.toPorts...)
			if !reflect.DeepEqual(tt.before, tt.after) {
				t.Errorf("PipeTo() = %v, want %v", tt.before, tt.after)
			}
		})
	}
}

func TestPort_PutSignal(t *testing.T) {
	portWithSingleSignal := NewPort("portWithSingleSignal")
	portWithSingleSignal.PutSignal(signal.New(11))

	portWithMultipleSignals := NewPort("portWithMultipleSignals")
	portWithMultipleSignals.PutSignal(signal.New(11, 12))

	portWithMultipleSignals2 := NewPort("portWithMultipleSignals2")
	portWithMultipleSignals2.PutSignal(signal.New(55, 66))

	type args struct {
		sig *signal.Signal
	}
	tests := []struct {
		name   string
		before *Port
		after  *Port
		args   args
	}{
		{
			name:   "single signal to empty port",
			before: NewPort("emptyPort"),
			after: &Port{
				name:   "emptyPort",
				signal: signal.New(11),
				pipes:  Collection{},
			},
			args: args{
				sig: signal.New(11),
			},
		},
		{
			name:   "multiple signals to empty port",
			before: NewPort("p"),
			after: &Port{
				name:   "p",
				signal: signal.New(11, 12),
				pipes:  Collection{},
			},
			args: args{
				sig: signal.New(11, 12),
			},
		},
		{
			name:   "single signal to port with single signal",
			before: portWithSingleSignal,
			after: &Port{
				name:   "portWithSingleSignal",
				signal: signal.New(12, 11), //Notice LIFO order
				pipes:  Collection{},
			},
			args: args{
				sig: signal.New(12),
			},
		},
		{
			name:   "single signal to port with multiple signals",
			before: portWithMultipleSignals,
			after: &Port{
				name:   "portWithMultipleSignals",
				signal: signal.New(13, 11, 12), //Notice LIFO order
				pipes:  Collection{},
			},
			args: args{
				sig: signal.New(13),
			},
		},
		{
			name:   "multiple signals to port with multiple signals",
			before: portWithMultipleSignals2,
			after: &Port{
				name:   "portWithMultipleSignals2",
				signal: signal.New(13, 14, 55, 66), //Notice LIFO order
				pipes:  Collection{},
			},
			args: args{
				sig: signal.New(13, 14),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before.PutSignal(tt.args.sig)
			if !reflect.DeepEqual(tt.before, tt.after) {
				t.Errorf("ClearSignal() = %v, want %v", tt.before, tt.after)
			}
		})
	}
}

func TestPort_Name(t *testing.T) {
	tests := []struct {
		name string
		port *Port
		want string
	}{
		{
			name: "happy path",
			port: NewPort("p777"),
			want: "p777",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.port.Name(), "Name()")
		})
	}
}

func TestNewPort(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *Port
	}{
		{
			name: "empty name is valid",
			args: args{
				name: "",
			},
			want: &Port{name: "", pipes: Collection{}},
		},
		{
			name: "with name",
			args: args{
				name: "p1",
			},
			want: &Port{name: "p1", pipes: Collection{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewPort(tt.args.name), "NewPort(%v)", tt.args.name)
		})
	}
}

func TestPort_Flush(t *testing.T) {
	portWithSignal1 := NewPort("portWithSignal1")
	portWithSignal1.PutSignal(signal.New(777))

	portWithSignal2 := NewPort("portWithSignal2")
	portWithSignal2.PutSignal(signal.New(888))

	portWithMultipleSignals := NewPort("portWithMultipleSignals")
	portWithMultipleSignals.PutSignal(signal.New(11, 12))

	emptyPort := NewPort("emptyPort")

	tests := []struct {
		name       string
		source     *Port
		dest       *Port
		assertions func(t *testing.T, source *Port, dest *Port)
	}{
		{
			name:   "port with no signal",
			source: NewPort("empty_src"),
			dest:   NewPort("empty_dest"),
			assertions: func(t *testing.T, source *Port, dest *Port) {
				assert.False(t, source.HasSignal())
				assert.False(t, dest.HasSignal())
			},
		},
		{
			name:   "flush to empty port",
			source: portWithSignal1,
			dest:   emptyPort,
			assertions: func(t *testing.T, source *Port, dest *Port) {
				//Source port is clear
				assert.False(t, source.HasSignal())

				//Signal transferred to destination port
				assert.True(t, dest.HasSignal())
				assert.Equal(t, dest.Signal().Payload().(int), 777)
			},
		},
		{
			name:   "flush to port with signal",
			source: portWithSignal2,
			dest:   portWithMultipleSignals,
			assertions: func(t *testing.T, source *Port, dest *Port) {
				//Source port is clear
				assert.False(t, source.HasSignal())

				//Destination port now has 1 more signal
				assert.True(t, dest.HasSignal())
				assert.Equal(t, 3, dest.Signal().Len())
				assert.Contains(t, dest.Signal().Payloads(), 888)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.source.PipeTo(tt.dest)
			tt.source.Flush()
			if tt.assertions != nil {
				tt.assertions(t, tt.source, tt.dest)
			}
		})
	}
}
