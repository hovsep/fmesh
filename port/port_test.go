package port

import (
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPort_HasSignals(t *testing.T) {
	portWithSignal := New("portWithSignal").WithSignals(signal.New(123))

	tests := []struct {
		name string
		port *Port
		want bool
	}{
		{
			name: "empty port",
			port: New("emptyPort"),
			want: false,
		},
		{
			name: "port has normal signals",
			port: portWithSignal,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.port.HasSignals())
		})
	}
}

func TestPort_Signals(t *testing.T) {
	portWithSignal := New("portWithSignal").WithSignals(signal.New(123))

	tests := []struct {
		name string
		port *Port
		want signal.Group
	}{
		{
			name: "no signals",
			port: New("noSignal"),
			want: signal.Group{},
		},
		{
			name: "with signal",
			port: portWithSignal,
			want: signal.NewGroup(123),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.port.Signals())
		})
	}
}

func TestPort_Clear(t *testing.T) {
	portWithSignal := New("portWithSignal").WithSignals(signal.New(111))

	tests := []struct {
		name   string
		before *Port
		after  *Port
	}{
		{
			name:   "happy path",
			before: portWithSignal,
			after:  &Port{name: "portWithSignal", pipes: Group{}, signals: signal.Group{}},
		},
		{
			name:   "cleaning empty port",
			before: New("emptyPort"),
			after:  &Port{name: "emptyPort", pipes: Group{}, signals: signal.Group{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before.Clear()
			assert.Equal(t, tt.after, tt.before)
		})
	}
}

func TestPort_PipeTo(t *testing.T) {
	p1, p2, p3, p4 := New("p1"), New("p2"), New("p3"), New("p4")

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
				name:    "p1",
				pipes:   Group{p2, p3},
				signals: signal.Group{},
			},
			args: args{
				toPorts: []*Port{p2, p3},
			},
		},
		{
			name:   "invalid ports are ignored",
			before: p4,
			after: &Port{
				name:    "p4",
				pipes:   Group{p2},
				signals: signal.Group{},
			},
			args: args{
				toPorts: []*Port{p2, nil},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before.PipeTo(tt.args.toPorts...)
			assert.Equal(t, tt.after, tt.before)
		})
	}
}

func TestPort_PutSignals(t *testing.T) {
	portWithSingleSignal := New("portWithSingleSignal").WithSignals(signal.New(11))

	portWithMultipleSignals := New("portWithMultipleSignals").WithSignals(signal.NewGroup(11, 12)...)

	portWithMultipleSignals2 := New("portWithMultipleSignals2").WithSignals(signal.NewGroup(55, 66)...)

	type args struct {
		signals []*signal.Signal
	}
	tests := []struct {
		name   string
		before *Port
		after  *Port
		args   args
	}{
		{
			name:   "single signal to empty port",
			before: New("emptyPort"),
			after: &Port{
				name:    "emptyPort",
				signals: signal.NewGroup(11),
				pipes:   Group{},
			},
			args: args{
				signals: signal.NewGroup(11),
			},
		},
		{
			name:   "multiple signals to empty port",
			before: New("p"),
			after: &Port{
				name:    "p",
				signals: signal.NewGroup(11, 12),
				pipes:   Group{},
			},
			args: args{
				signals: signal.NewGroup(11, 12),
			},
		},
		{
			name:   "single signal to port with single signal",
			before: portWithSingleSignal,
			after: &Port{
				name:    "portWithSingleSignal",
				signals: signal.NewGroup(11, 12),
				pipes:   Group{},
			},
			args: args{
				signals: signal.NewGroup(12),
			},
		},
		{
			name:   "single signals to port with multiple signals",
			before: portWithMultipleSignals,
			after: &Port{
				name:    "portWithMultipleSignals",
				signals: signal.NewGroup(11, 12, 13),
				pipes:   Group{},
			},
			args: args{
				signals: signal.NewGroup(13),
			},
		},
		{
			name:   "multiple signals to port with multiple signals",
			before: portWithMultipleSignals2,
			after: &Port{
				name:    "portWithMultipleSignals2",
				signals: signal.NewGroup(55, 66, 13, 14), //Notice LIFO order
				pipes:   Group{},
			},
			args: args{
				signals: signal.NewGroup(13, 14),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before.PutSignals(tt.args.signals...)
			assert.ElementsMatch(t, tt.after.Signals(), tt.before.Signals())
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
			port: New("p777"),
			want: "p777",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.port.Name())
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
			want: &Port{
				name:    "",
				pipes:   Group{},
				signals: signal.Group{},
			},
		},
		{
			name: "with name",
			args: args{
				name: "p1",
			},
			want: &Port{
				name:    "p1",
				pipes:   Group{},
				signals: signal.Group{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.args.name))
		})
	}
}
