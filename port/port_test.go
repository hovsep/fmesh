package port

import (
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewPorts(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name string
		args args
		want Ports
	}{
		{
			name: "no names",
			args: args{
				names: nil,
			},
			want: Ports{},
		},
		{
			name: "happy path",
			args: args{
				names: []string{"i1", "i2"},
			},
			want: Ports{
				"i1": {name: "i1"},
				"i2": {name: "i2"},
			},
		},
		{
			name: "duplicate names are ignored",
			args: args{
				names: []string{"i1", "i2", "i1"},
			},
			want: Ports{
				"i1": {name: "i1"},
				"i2": {name: "i2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPorts(tt.args.names...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

func TestPort_Pipes(t *testing.T) {
	destPort1, destPort2, destPort3 := NewPort("destPort1"), NewPort("destPort2"), NewPort("destPort3")
	portWithOnePipe := NewPort("portWithOnePipe")
	portWithOnePipe.PipeTo(destPort1)

	portWithMultiplePipes := NewPort("portWithMultiplePipes")
	portWithMultiplePipes.PipeTo(destPort2, destPort3)

	tests := []struct {
		name string
		port *Port
		want Pipes
	}{
		{
			name: "no pipes",
			port: NewPort("noPipes"),
			want: nil,
		},
		{
			name: "one pipe",
			port: portWithOnePipe,
			want: Pipes{
				{
					From: portWithOnePipe,
					To:   destPort1,
				},
			},
		},
		{
			name: "multiple pipes",
			port: portWithMultiplePipes,
			want: Pipes{
				{
					From: portWithMultiplePipes,
					To:   destPort2,
				},
				{
					From: portWithMultiplePipes,
					To:   destPort3,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.port.Pipes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pipes() = %v, want %v", got, tt.want)
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

func TestPorts_AllHaveSignal(t *testing.T) {
	oneEmptyPorts := NewPorts("p1", "p2", "p3")
	oneEmptyPorts.PutSignal(signal.New(123))
	oneEmptyPorts.ByName("p2").ClearSignal()

	allWithSignalPorts := NewPorts("out1", "out2", "out3")
	allWithSignalPorts.PutSignal(signal.New(77))

	allWithEmptySignalPorts := NewPorts("in1", "in2", "in3")
	allWithEmptySignalPorts.PutSignal(signal.New())

	tests := []struct {
		name  string
		ports Ports
		want  bool
	}{
		{
			name:  "all empty",
			ports: NewPorts("p1", "p2"),
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

func TestPorts_AnyHasSignal(t *testing.T) {
	oneEmptyPorts := NewPorts("p1", "p2", "p3")
	oneEmptyPorts.PutSignal(signal.New(123))
	oneEmptyPorts.ByName("p2").ClearSignal()

	tests := []struct {
		name  string
		ports Ports
		want  bool
	}{
		{
			name:  "one empty",
			ports: oneEmptyPorts,
			want:  true,
		},
		{
			name:  "all empty",
			ports: NewPorts("p1", "p2", "p3"),
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

func TestPorts_ByName(t *testing.T) {
	portsWithSignals := NewPorts("p1", "p2")
	portsWithSignals.PutSignal(signal.New(12))

	type args struct {
		name string
	}
	tests := []struct {
		name  string
		ports Ports
		args  args
		want  *Port
	}{
		{
			name:  "empty port found",
			ports: NewPorts("p1", "p2"),
			args: args{
				name: "p1",
			},
			want: &Port{name: "p1"},
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
			},
		},
		{
			name:  "port not found",
			ports: NewPorts("p1", "p2"),
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

func TestPorts_ByNames(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name  string
		ports Ports
		args  args
		want  Ports
	}{
		{
			name:  "single port found",
			ports: NewPorts("p1", "p2"),
			args: args{
				names: []string{"p1"},
			},
			want: Ports{
				"p1": &Port{name: "p1"},
			},
		},
		{
			name:  "multiple ports found",
			ports: NewPorts("p1", "p2"),
			args: args{
				names: []string{"p1", "p2"},
			},
			want: Ports{
				"p1": &Port{name: "p1"},
				"p2": &Port{name: "p2"},
			},
		},
		{
			name:  "single port not found",
			ports: NewPorts("p1", "p2"),
			args: args{
				names: []string{"p7"},
			},
			want: Ports{},
		},
		{
			name:  "some ports not found",
			ports: NewPorts("p1", "p2"),
			args: args{
				names: []string{"p1", "p2", "p3"},
			},
			want: Ports{
				"p1": &Port{name: "p1"},
				"p2": &Port{name: "p2"},
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
			after:  &Port{name: "portWithSignal"},
		},
		{
			name:   "cleaning empty port",
			before: NewPort("emptyPort"),
			after:  &Port{name: "emptyPort"},
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
				name: "p1",
				pipes: Pipes{
					{
						From: p1,
						To:   p2,
					},
					{
						From: p1,
						To:   p3,
					},
				},
			},
			args: args{
				toPorts: []*Port{p2, p3},
			},
		},
		{
			name:   "invalid ports are ignored",
			before: p4,
			after: &Port{
				name: "p4",
				pipes: Pipes{
					{
						From: p4,
						To:   p2,
					},
				},
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

func TestPorts_ClearSignal(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ports := NewPorts("p1", "p2", "p3")
		ports.PutSignal(signal.New(1, 2, 3))
		assert.True(t, ports.AllHaveSignal())
		ports.ClearSignal()
		assert.False(t, ports.AnyHasSignal())
	})
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
