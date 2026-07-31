package port

import (
	"context"
	"testing"

	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Local helpers: this in-package test cannot import internal/testutil
// (testutil imports port).

// mustInput is a test helper that panics if NewInput returns an error.
func mustInput(name string) *Port {
	p, err := NewInput(name)
	if err != nil {
		panic(err)
	}
	return p
}

// mustOutput is a test helper that panics if NewOutput returns an error.
func mustOutput(name string) *Port {
	p, err := NewOutput(name)
	if err != nil {
		panic(err)
	}
	return p
}

// The *Labeled/*Scalared helpers exist because metadata mutation returns the
// store, not the port — table entries need a *Port.
func mustInputLabeled(name string, labels map[string]string) *Port {
	p := mustInput(name)
	p.Labels().SetMany(labels)
	return p
}

func mustOutputLabeled(name string, labels map[string]string) *Port {
	p := mustOutput(name)
	p.Labels().SetMany(labels)
	return p
}

func mustInputScalared(name string, scalars map[string]float64) *Port {
	p := mustInput(name)
	p.Scalars().SetMany(scalars)
	return p
}

func mustOutputScalared(name string, scalars map[string]float64) *Port {
	p := mustOutput(name)
	p.Scalars().SetMany(scalars)
	return p
}

func TestPort_HasSignals(t *testing.T) {
	tests := []struct {
		name string
		port *Port
		want bool
	}{
		{
			name: "empty port",
			port: mustOutput("emptyPort"),
			want: false,
		},
		{
			name: "port has signals",
			port: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PutSignals(signal.New(123)))
				return p
			}(),
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
	tests := []struct {
		name       string
		port       *Port
		assertions func(t *testing.T, group *signal.Group)
	}{
		{
			name: "empty signals",
			port: mustOutput("noSignal"),
			assertions: func(t *testing.T, group *signal.Group) {
				assert.True(t, group.IsEmpty())
			},
		},
		{
			name: "with signal",
			port: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PutSignals(signal.New(123)))
				return p
			}(),
			assertions: func(t *testing.T, group *signal.Group) {
				assert.Equal(t, 1, group.Len())
				assert.Equal(t, 123, group.FirstPayloadOrNil())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.port.Signals()
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestPort_Clear(t *testing.T) {
	tests := []struct {
		name       string
		before     *Port
		assertions func(t *testing.T, p *Port)
	}{
		{
			name: "happy path",
			before: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PutSignals(signal.New(111)))
				return p
			}(),
			assertions: func(t *testing.T, p *Port) {
				assert.False(t, p.HasSignals())
			},
		},
		{
			name:   "cleaning empty port",
			before: mustOutput("emptyPort"),
			assertions: func(t *testing.T, p *Port) {
				assert.False(t, p.HasSignals())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.before.Clear(context.Background())
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.before)
			}
		})
	}
}

func TestPort_PipeTo(t *testing.T) {
	out1 := mustOutput("out1")
	out2 := mustOutput("out2")
	out3 := mustOutput("out3")
	in1 := mustInput("in1")
	in2 := mustInput("in2")
	in3 := mustInput("in3")

	tests := []struct {
		name       string
		before     *Port
		toPorts    []*Port
		wantErr    bool
		assertions func(t *testing.T, portAfter *Port)
	}{
		{
			name:    "happy path",
			before:  out1,
			toPorts: []*Port{in2, in3},
			wantErr: false,
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, 2, portAfter.Pipes().Len())
			},
		},
		{
			name:    "nil port is not allowed",
			before:  out3,
			toPorts: []*Port{in2, nil},
			wantErr: true,
		},
		{
			name:    "piping from input ports is not allowed",
			before:  in1,
			toPorts: []*Port{in2, out2},
			wantErr: true,
		},
		{
			name:    "piping to output ports is not allowed",
			before:  out2,
			toPorts: []*Port{out3},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.before.PipeTo(tt.toPorts...)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if tt.assertions != nil {
				tt.assertions(t, tt.before)
			}
		})
	}
}

func TestPort_PutSignals(t *testing.T) {
	type args struct {
		signals []*signal.Signal
	}
	tests := []struct {
		name       string
		port       *Port
		args       args
		assertions func(t *testing.T, portAfter *Port)
	}{
		{
			name: "single signal to empty port",
			port: mustOutput("emptyPort"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11), portAfter.Signals())
			},
			args: func() args {
				signals := signal.NewGroup(11).All()
				return args{signals: signals}
			}(),
		},
		{
			name: "multiple signals to empty port",
			port: mustOutput("p"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12), portAfter.Signals())
			},
			args: func() args {
				signals := signal.NewGroup(11, 12).All()
				return args{signals: signals}
			}(),
		},
		{
			name: "single signal to port with single signal",
			port: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PutSignals(signal.New(11)))
				return p
			}(),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12), portAfter.Signals())
			},
			args: func() args {
				signals := signal.NewGroup(12).All()
				return args{signals: signals}
			}(),
		},
		{
			name: "single signal to port with multiple signals",
			port: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PutSignalGroups(signal.NewGroup(11, 12)))
				return p
			}(),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12, 13), portAfter.Signals())
			},
			args: func() args {
				signals := signal.NewGroup(13).All()
				return args{signals: signals}
			}(),
		},
		{
			name: "multiple signals to port with multiple signals",
			port: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PutSignalGroups(signal.NewGroup(55, 66)))
				return p
			}(),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(55, 66, 13, 14), portAfter.Signals())
			},
			args: func() args {
				signals := signal.NewGroup(13, 14).All()
				return args{signals: signals}
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.port.PutSignals(tt.args.signals...)
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.port)
			}
		})
	}
}

func TestPort_PutPayloads(t *testing.T) {
	tests := []struct {
		name       string
		port       *Port
		payloads   []any
		assertions func(t *testing.T, portAfter *Port)
	}{
		{
			name: "single payload to empty port",
			port: mustOutput("emptyPort"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11), portAfter.Signals())
			},
			payloads: []any{11},
		},
		{
			name: "multiple signals to empty port",
			port: mustOutput("p"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12), portAfter.Signals())
			},
			payloads: []any{11, 12},
		},
		{
			name: "single signal to port with single signal",
			port: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PutSignals(signal.New(11)))
				return p
			}(),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12), portAfter.Signals())
			},
			payloads: []any{12},
		},
		{
			name: "single signal to port with multiple signals",
			port: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PutSignalGroups(signal.NewGroup(11, 12)))
				return p
			}(),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12, 13), portAfter.Signals())
			},
			payloads: []any{13},
		},
		{
			name: "multiple signals to port with multiple signals",
			port: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PutSignalGroups(signal.NewGroup(55, 66)))
				return p
			}(),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(55, 66, 13, 14), portAfter.Signals())
			},
			payloads: []any{13, 14},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.port.PutPayloads(tt.payloads...)
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, tt.port)
			}
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
	}{
		{
			name: "empty name is valid",
			args: args{name: ""},
		},
		{
			name: "with name",
			args: args{name: "p1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewOutput(tt.args.name)
			require.NoError(t, err)
			assert.Equal(t, tt.args.name, p.Name())
		})
	}
}

func TestPort_Direction(t *testing.T) {
	tests := []struct {
		name string
		port *Port
		want Direction
	}{
		{
			name: "explicitly set to input",
			port: mustInput("p"),
			want: DirectionIn,
		},
		{
			name: "explicitly set to output",
			port: mustOutput("p"),
			want: DirectionOut,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.port.Direction())

			// Also verify IsInput() and IsOutput() consistency
			if tt.want == DirectionIn {
				assert.True(t, tt.port.IsInput())
				assert.False(t, tt.port.IsOutput())
			} else {
				assert.False(t, tt.port.IsInput())
				assert.True(t, tt.port.IsOutput())
			}
		})
	}
}

func TestPort_HasPipes(t *testing.T) {
	tests := []struct {
		name string
		port *Port
		want bool
	}{
		{
			name: "no pipes",
			port: mustOutput("p"),
			want: false,
		},
		{
			name: "with pipes",
			port: func() *Port {
				p := mustOutput("p1")
				require.NoError(t, p.PipeTo(mustInput("p2")))
				return p
			}(),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.port.HasPipes())
		})
	}
}

func TestPort_Flush(t *testing.T) {
	tests := []struct {
		name       string
		srcPort    *Port
		wantErr    bool
		assertions func(t *testing.T, srcPort *Port)
	}{
		{
			name: "port with signals and no pipes is not flushed",
			srcPort: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3)))
				return p
			}(),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.True(t, srcPort.HasSignals())
				assert.Equal(t, 3, srcPort.Signals().Len())
				assert.False(t, srcPort.HasPipes())
			},
		},
		{
			name: "empty port with pipes is not flushed",
			srcPort: func() *Port {
				p := mustOutput("p")
				require.NoError(t, p.PipeTo(mustInput("p1"), mustInput("p2")))
				return p
			}(),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
			},
		},
		{
			name: "flush to empty ports",
			srcPort: func() *Port {
				dst1 := mustInput("p1")
				dst2 := mustInput("p2")
				p := mustOutput("p")
				require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3)))
				require.NoError(t, p.PipeTo(dst1, dst2))
				return p
			}(),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
				destPorts := srcPort.Pipes().All()
				for _, destPort := range destPorts {
					assert.True(t, destPort.HasSignals())
					assert.Equal(t, 3, destPort.Signals().Len())
					allPayloads := destPort.Signals().AllPayloads()
					assert.Contains(t, allPayloads, 1)
					assert.Contains(t, allPayloads, 2)
					assert.Contains(t, allPayloads, 3)
				}
			},
		},
		{
			name: "flush to non empty ports",
			srcPort: func() *Port {
				dst1 := mustInput("p1")
				dst2 := mustInput("p2")
				require.NoError(t, dst1.PutSignalGroups(signal.NewGroup(4, 5, 6)))
				require.NoError(t, dst2.PutSignalGroups(signal.NewGroup(7, 8, 9)))
				p := mustOutput("p")
				require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3)))
				require.NoError(t, p.PipeTo(dst1, dst2))
				return p
			}(),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
				destPorts := srcPort.Pipes().All()
				for _, destPort := range destPorts {
					assert.True(t, destPort.HasSignals())
					assert.Equal(t, 6, destPort.Signals().Len())
					allPayloads := destPort.Signals().AllPayloads()
					assert.Contains(t, allPayloads, 1)
					assert.Contains(t, allPayloads, 2)
					assert.Contains(t, allPayloads, 3)
				}
			},
		},
		{
			name: "GUARDRAIL: cannot flush input port",
			srcPort: func() *Port {
				p := mustInput("input1")
				require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3)))
				return p
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.srcPort.Flush(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "cannot flush input port")
			} else {
				require.NoError(t, err)
			}
			if tt.assertions != nil {
				tt.assertions(t, tt.srcPort)
			}
		})
	}
}

// Ports hold metadata in a live store; the store's own behavior is covered in the
// meta package, so this only checks the wiring.
func TestPort_Metadata(t *testing.T) {
	t.Run("Labels returns the live store", func(t *testing.T) {
		p := mustInput("p1")
		p.Labels().Set("role", "primary").SetMany(map[string]string{"tier": "edge"})

		assert.Equal(t, 2, p.Labels().Len())
		assert.True(t, p.Labels().ValueIs("role", "primary"))

		p.Labels().Remove("role")
		assert.False(t, p.Labels().Has("role"))

		p.Labels().Clear()
		assert.Zero(t, p.Labels().Len())
	})

	t.Run("Scalars returns the live store", func(t *testing.T) {
		p := mustOutput("p1")
		p.Scalars().Set("rate", 100)

		assert.True(t, p.Scalars().ValueIs("rate", 100))

		p.Scalars().Clear()
		assert.Zero(t, p.Scalars().Len())
	})

	t.Run("constructor options seed both stores", func(t *testing.T) {
		p, err := NewInput("p1", WithLabel("role", "primary"), WithScalar("rate", 5))
		require.NoError(t, err)

		assert.True(t, p.Labels().ValueIs("role", "primary"))
		assert.True(t, p.Scalars().ValueIs("rate", 5))
	})

	t.Run("stores are per port", func(t *testing.T) {
		p1, p2 := mustInput("p1"), mustInput("p2")
		p1.Labels().Set("only", "p1")

		assert.False(t, p2.Labels().Has("only"))
	})
}

func TestPort_Pipes(t *testing.T) {
	tests := []struct {
		name              string
		port              *Port
		wantLen           int
		wantErrContaining string
	}{
		{
			name:    "no pipes",
			port:    mustOutput("p"),
			wantLen: 0,
		},
		{
			name: "with pipes",
			port: func() *Port {
				p := mustOutput("p1")
				require.NoError(t, p.PipeTo(mustInput("p2"), mustInput("p3")))
				return p
			}(),
			wantLen: 2,
		},
		{
			name: "input port has no pipes",
			port: mustInput("input1"),
			// Input ports simply have an empty pipes group
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.port.Pipes()
			assert.Equal(t, tt.wantLen, got.Len())
		})
	}
}

func TestPort_SignalsAccess(t *testing.T) {
	t.Run("FirstSignalPayload", func(t *testing.T) {
		p := mustOutput("p")
		require.NoError(t, p.PutSignalGroups(signal.NewGroup(4, 7, 6, 5)))
		payload, err := p.Signals().FirstPayload()
		require.NoError(t, err)
		assert.Equal(t, 4, payload)
	})
}

func TestPort_ForwardSignals(t *testing.T) {
	type args struct {
		srcPort  *Port
		destPort *Port
	}
	tests := []struct {
		name       string
		args       args
		assertions func(t *testing.T, srcPortAfter, destPortAfter *Port, err error)
	}{
		{
			name: "happy path",
			args: args{
				srcPort: func() *Port {
					p := mustOutput("p1")
					require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3)))
					return p
				}(),
				destPort: mustOutput("p2"),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.NoError(t, err)
				assert.Equal(t, 3, destPortAfter.Signals().Len())
				assert.Equal(t, 3, srcPortAfter.Signals().Len())
			},
		},
		{
			name: "signals are added to dest port",
			args: args{
				srcPort: func() *Port {
					p := mustOutput("p1")
					require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3)))
					return p
				}(),
				destPort: func() *Port {
					p := mustOutput("p2")
					require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3, 4, 5, 6)))
					return p
				}(),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.NoError(t, err)
				assert.Equal(t, 9, destPortAfter.Signals().Len())
				assert.Equal(t, 3, srcPortAfter.Signals().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ForwardSignals(context.Background(), tt.args.srcPort, tt.args.destPort)
			if tt.assertions != nil {
				tt.assertions(t, tt.args.srcPort, tt.args.destPort, err)
			}
		})
	}
}

func TestPort_ForwardWithFilter(t *testing.T) {
	type args struct {
		srcPort   *Port
		destPort  *Port
		predicate signal.Predicate
	}
	tests := []struct {
		name       string
		args       args
		assertions func(t *testing.T, srcPortAfter, destPortAfter *Port, err error)
	}{
		{
			name: "all kept",
			args: args{
				srcPort: func() *Port {
					p := mustOutput("p1")
					require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3)))
					return p
				}(),
				destPort: mustOutput("p2"),
				predicate: func(signal *signal.Signal) bool {
					return signal.Payload().(int) > 0
				},
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.NoError(t, err)
				assert.Equal(t, 3, destPortAfter.Signals().Len())
				assert.Equal(t, 3, srcPortAfter.Signals().Len())
			},
		},
		{
			name: "some dropped",
			args: args{
				srcPort: func() *Port {
					p := mustOutput("p1")
					require.NoError(t, p.PutSignalGroups(signal.NewGroup(7, 8, 9, 10, 11, 12, 13)))
					return p
				}(),
				destPort: func() *Port {
					p := mustOutput("p2")
					require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3, 4, 5, 6)))
					return p
				}(),
				predicate: func(signal *signal.Signal) bool {
					return signal.Payload().(int) > 10
				},
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.NoError(t, err)
				assert.Equal(t, 9, destPortAfter.Signals().Len())
				assert.Equal(t, 7, srcPortAfter.Signals().Len())
			},
		},
		{
			name: "all dropped",
			args: args{
				srcPort: func() *Port {
					p := mustOutput("p1")
					require.NoError(t, p.PutSignalGroups(signal.NewGroup(7, 8, 9, 10, 11, 12, 13)))
					return p
				}(),
				destPort: func() *Port {
					p := mustOutput("p2")
					require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3, 4, 5, 6)))
					return p
				}(),
				predicate: func(signal *signal.Signal) bool {
					return signal.Payload().(int) > 99
				},
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.NoError(t, err)
				assert.Equal(t, 6, destPortAfter.Signals().Len())
				assert.Equal(t, 7, srcPortAfter.Signals().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ForwardWithFilter(context.Background(), tt.args.srcPort, tt.args.destPort, tt.args.predicate)
			if tt.assertions != nil {
				tt.assertions(t, tt.args.srcPort, tt.args.destPort, err)
			}
		})
	}
}

func TestPort_ForwardWithMap(t *testing.T) {
	type args struct {
		srcPort    *Port
		destPort   *Port
		mapperFunc signal.Mapper
	}
	tests := []struct {
		name       string
		args       args
		assertions func(t *testing.T, srcPortAfter, destPortAfter *Port, err error)
	}{
		{
			name: "happy path",
			args: args{
				srcPort: func() *Port {
					p := mustOutput("p1")
					require.NoError(t, p.PutSignalGroups(signal.NewGroup(1, 2, 3)))
					return p
				}(),
				destPort: mustOutput("p2"),
				mapperFunc: func(sig *signal.Signal) *signal.Signal {
					return sig.WithOnlyLabels(map[string]string{
						"l1": "v1",
					})
				},
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.NoError(t, err)
				assert.Equal(t, 3, destPortAfter.Signals().Len())
				assert.Equal(t, 3, srcPortAfter.Signals().Len())
				assert.True(t, destPortAfter.Signals().Every(signal.LabelEquals("l1", "v1")))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ForwardWithMap(context.Background(), tt.args.srcPort, tt.args.destPort, tt.args.mapperFunc)
			if tt.assertions != nil {
				tt.assertions(t, tt.args.srcPort, tt.args.destPort, err)
			}
		})
	}
}

func TestPort_Chainability(t *testing.T) {
	t.Run("SetLabels called twice replaces all labels", func(t *testing.T) {
		p := mustOutput("p1")
		p.Labels().Clear().SetMany(map[string]string{"k1": "v1", "k2": "v2"})
		p.Labels().Clear().SetMany(map[string]string{"k3": "v3"})

		assert.Equal(t, 1, p.Labels().Len())
		assert.False(t, p.Labels().Has("k1"), "k1 should be replaced")
		assert.False(t, p.Labels().Has("k2"), "k2 should be replaced")
		assert.True(t, p.Labels().ValueIs("k3", "v3"))
	})

	t.Run("AddLabels called twice merges labels", func(t *testing.T) {
		p := mustOutput("p1")
		p.Labels().SetMany(map[string]string{"k1": "v1", "k2": "v2"})
		p.Labels().SetMany(map[string]string{"k3": "v3", "k2": "v2-updated"})

		assert.Equal(t, 3, p.Labels().Len())
		assert.True(t, p.Labels().ValueIs("k1", "v1"))
		assert.True(t, p.Labels().ValueIs("k2", "v2-updated"), "should update existing key")
		assert.True(t, p.Labels().ValueIs("k3", "v3"))
	})

	t.Run("mixed Set and Add operations", func(t *testing.T) {
		p := mustOutput("p1")
		p.Labels().
			Set("k1", "v1").
			SetMany(map[string]string{"k2": "v2", "k3": "v3"}).
			Clear().SetMany(map[string]string{"k4": "v4"}). // Wipes k1, k2, k3
			Set("k5", "v5")                                 // Merges with k4

		assert.Equal(t, 2, p.Labels().Len())
		assert.False(t, p.Labels().Has("k1"), "wiped by Clear")
		assert.False(t, p.Labels().Has("k2"), "wiped by Clear")
		assert.False(t, p.Labels().Has("k3"), "wiped by Clear")
		assert.True(t, p.Labels().ValueIs("k4", "v4"))
		assert.True(t, p.Labels().ValueIs("k5", "v5"))
	})

	t.Run("WithDescription replaces previous value", func(t *testing.T) {
		p, err := NewOutput("p1", WithDescription("first"), WithDescription("second"))
		require.NoError(t, err)
		assert.Equal(t, "second", p.Description())
	})

	t.Run("PutSignals called twice adds signals", func(t *testing.T) {
		p := mustOutput("p1")
		require.NoError(t, p.PutSignals(signal.New(1), signal.New(2)))
		require.NoError(t, p.PutSignals(signal.New(3)))
		assert.Equal(t, 3, p.Signals().Len())
	})

	t.Run("Clear removes all signals", func(t *testing.T) {
		p := mustOutput("p1")
		require.NoError(t, p.PutSignals(signal.New(1), signal.New(2)))
		require.NoError(t, p.PutSignals(signal.New(3)))
		require.NoError(t, p.Clear(context.Background()))
		assert.Equal(t, 0, p.Signals().Len())
		assert.False(t, p.HasSignals())
	})

	t.Run("ClearLabels removes all labels", func(t *testing.T) {
		p := mustOutput("p1")
		p.Labels().
			SetMany(map[string]string{"k1": "v1", "k2": "v2"}).
			Clear().
			Set("k3", "v3")

		assert.Equal(t, 1, p.Labels().Len())
		assert.False(t, p.Labels().Has("k1"))
		assert.False(t, p.Labels().Has("k2"))
		assert.True(t, p.Labels().ValueIs("k3", "v3"))
	})

	t.Run("RemoveLabels removes specific labels", func(t *testing.T) {
		p := mustOutput("p1")
		p.Labels().
			SetMany(map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}).
			Remove("k1", "k2").
			Set("k4", "v4")

		assert.Equal(t, 2, p.Labels().Len())
		assert.False(t, p.Labels().Has("k1"))
		assert.False(t, p.Labels().Has("k2"))
		assert.True(t, p.Labels().ValueIs("k3", "v3"))
		assert.True(t, p.Labels().ValueIs("k4", "v4"))
	})
}
