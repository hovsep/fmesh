package port

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPort_HasSignals(t *testing.T) {
	tests := []struct {
		name string
		port *Port
		want bool
	}{
		{
			name: "empty port",
			port: NewOutput("emptyPort"),
			want: false,
		},
		{
			name: "port has signals",
			port: NewOutput("p").PutSignals(signal.New(123)),
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
			port: NewOutput("noSignal"),
			assertions: func(t *testing.T, group *signal.Group) {
				assert.True(t, group.IsEmpty())
			},
		},
		{
			name: "with signal",
			port: NewOutput("p").PutSignals(signal.New(123)),
			assertions: func(t *testing.T, group *signal.Group) {
				assert.Equal(t, 1, group.Len())
				assert.Equal(t, 123, group.FirstPayloadOrNil())
			},
		},
		{
			name: "with chain error",
			port: NewOutput("p").WithChainableErr(errors.New("some error")),
			assertions: func(t *testing.T, group *signal.Group) {
				assert.ErrorContains(t, group.ChainableErr(), "some error")
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
		name   string
		before *Port
		after  *Port
	}{
		{
			name:   "happy path",
			before: NewOutput("p").PutSignals(signal.New(111)),
			after:  NewOutput("p"),
		},
		{
			name:   "cleaning empty port",
			before: NewOutput("emptyPort"),
			after:  NewOutput("emptyPort"),
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
	outputPorts := NewCollection()
	outputPorts = outputPorts.Add(
		NewOutput("out1"),
		NewOutput("out2"),
		NewOutput("out3"),
	)

	inputPorts := NewCollection()
	inputPorts = inputPorts.Add(
		NewInput("in1"),
		NewInput("in2"),
		NewInput("in3"),
	)

	type args struct {
		toPorts Ports
	}
	tests := []struct {
		name       string
		before     *Port
		assertions func(t *testing.T, portAfter *Port)
		args       args
	}{
		{
			name:   "happy path",
			before: outputPorts.ByName("out1"),
			args: args{
				toPorts: Ports{inputPorts.ByName("in2"), inputPorts.ByName("in3")},
			},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.False(t, portAfter.HasChainableErr())
				require.NoError(t, portAfter.ChainableErr())
				assert.Equal(t, 2, portAfter.Pipes().Len())
			},
		},
		{
			name:   "nil port is not allowed",
			before: outputPorts.ByName("out3"),
			args: args{
				toPorts: Ports{inputPorts.ByName("in2"), nil},
			},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.True(t, portAfter.HasChainableErr())
				assert.Error(t, portAfter.ChainableErr())
			},
		},
		{
			name:   "piping from input ports is not allowed",
			before: inputPorts.ByName("in1"),
			args: args{
				toPorts: Ports{
					inputPorts.ByName("in2"), outputPorts.ByName("out1"),
				},
			},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.True(t, portAfter.HasChainableErr())
				assert.Error(t, portAfter.ChainableErr())
			},
		},
		{
			name:   "piping to output ports is not allowed",
			before: outputPorts.ByName("out1"),
			args: args{
				toPorts: Ports{outputPorts.ByName("out2")},
			},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.True(t, portAfter.HasChainableErr())
				assert.Error(t, portAfter.ChainableErr())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portAfter := tt.before.PipeTo(tt.args.toPorts...)
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
			}
		})
	}
}

func TestPort_PutSignals(t *testing.T) {
	type args struct {
		signals signal.Signals
	}
	tests := []struct {
		name       string
		port       *Port
		args       args
		assertions func(t *testing.T, portAfter *Port)
	}{
		{
			name: "single signal to empty port",
			port: NewOutput("emptyPort"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11), portAfter.Signals())
			},
			args: func() args {
				signals, _ := signal.NewGroup(11).All()
				return args{signals: signals}
			}(),
		},
		{
			name: "multiple signals to empty port",
			port: NewOutput("p"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12), portAfter.Signals())
			},
			args: func() args {
				signals, _ := signal.NewGroup(11, 12).All()
				return args{signals: signals}
			}(),
		},
		{
			name: "single signal to port with single signal",
			port: NewOutput("p").PutSignals(signal.New(11)),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12), portAfter.Signals())
			},
			args: func() args {
				signals, _ := signal.NewGroup(12).All()
				return args{signals: signals}
			}(),
		},
		{
			name: "single signal to port with multiple signals",
			port: NewOutput("p").PutSignalGroups(signal.NewGroup(11, 12)),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12, 13), portAfter.Signals())
			},
			args: func() args {
				signals, _ := signal.NewGroup(13).All()
				return args{signals: signals}
			}(),
		},
		{
			name: "multiple signals to port with multiple signals",
			port: NewOutput("p").PutSignalGroups(signal.NewGroup(55, 66)),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(55, 66, 13, 14), portAfter.Signals())
			},
			args: func() args {
				signals, _ := signal.NewGroup(13, 14).All()
				return args{signals: signals}
			}(),
		},
		{
			name: "chain error propagated from signals",
			port: NewOutput("p"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Zero(t, portAfter.Signals().Len())
				assert.True(t, portAfter.Signals().HasChainableErr())
			},
			args: args{
				signals: signal.Signals{signal.New(111).WithChainableErr(errors.New("some error in signal"))},
			},
		},
		{
			name: "with chain error",
			port: NewOutput("p").WithChainableErr(errors.New("some error in port")),
			args: args{
				signals: signal.Signals{signal.New(123)},
			},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.True(t, portAfter.HasChainableErr())
				assert.Zero(t, portAfter.Signals().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portAfter := tt.port.PutSignals(tt.args.signals...)
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
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
			port: NewOutput("emptyPort"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11), portAfter.Signals())
			},
			payloads: []any{11},
		},
		{
			name: "multiple signals to empty port",
			port: NewOutput("p"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12), portAfter.Signals())
			},
			payloads: []any{11, 12},
		},
		{
			name: "single signal to port with single signal",
			port: NewOutput("p").PutSignals(signal.New(11)),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12), portAfter.Signals())
			},
			payloads: []any{12},
		},
		{
			name: "single signal to port with multiple signals",
			port: NewOutput("p").PutSignalGroups(signal.NewGroup(11, 12)),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12, 13), portAfter.Signals())
			},
			payloads: []any{13},
		},
		{
			name: "multiple signals to port with multiple signals",
			port: NewOutput("p").PutSignalGroups(signal.NewGroup(55, 66)),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(55, 66, 13, 14), portAfter.Signals())
			},
			payloads: []any{13, 14},
		},

		{
			name:     "with chain error",
			port:     NewOutput("p").WithChainableErr(errors.New("some error in port")),
			payloads: []any{123},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.True(t, portAfter.HasChainableErr())
				assert.Zero(t, portAfter.Signals().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portAfter := tt.port.PutPayloads(tt.payloads...)
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
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
		want *Port
	}{
		{
			name: "empty name is valid",
			args: args{
				name: "",
			},
			want: NewOutput(""),
		},
		{
			name: "with name",
			args: args{
				name: "p1",
			},
			want: NewOutput("p1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewOutput(tt.args.name))
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
			name: "default direction is out (zero-value)",
			port: NewOutput("p"),
			want: DirectionOut,
		},
		{
			name: "explicitly set to input",
			port: NewInput("p"),
			want: DirectionIn,
		},
		{
			name: "explicitly set to output",
			port: NewOutput("p"),
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
			port: NewOutput("p"),
			want: false,
		},
		{
			name: "with pipes",
			port: NewOutput("p1").PipeTo(NewInput("p2")),
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
		assertions func(t *testing.T, srcPort *Port)
	}{
		{
			name:    "port with signals and no pipes is not flushed",
			srcPort: NewOutput("p").PutSignalGroups(signal.NewGroup(1, 2, 3)),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.True(t, srcPort.HasSignals())
				assert.Equal(t, 3, srcPort.Signals().Len())
				assert.False(t, srcPort.HasPipes())
			},
		},
		{
			name: "empty port with pipes is not flushed",
			srcPort: NewOutput("p").PipeTo(
				NewInput("p1"),
				NewInput("p2"),
			),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
			},
		},
		{
			name: "flush to empty ports",
			srcPort: NewOutput("p").PutSignalGroups(signal.NewGroup(1, 2, 3)).
				PipeTo(
					NewInput("p1"),
					NewInput("p2")),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
				destPorts, err := srcPort.Pipes().All()
				require.NoError(t, err)
				for _, destPort := range destPorts {
					assert.True(t, destPort.HasSignals())
					assert.Equal(t, 3, destPort.Signals().Len())
					allPayloads, err := destPort.Signals().AllPayloads()
					require.NoError(t, err)
					assert.Contains(t, allPayloads, 1)
					assert.Contains(t, allPayloads, 2)
					assert.Contains(t, allPayloads, 3)
				}
			},
		},
		{
			name: "flush to non empty ports",
			srcPort: NewOutput("p").
				PutSignalGroups(signal.NewGroup(1, 2, 3)).
				PipeTo(
					NewInput("p1").PutSignalGroups(signal.NewGroup(4, 5, 6)),
					NewInput("p2").PutSignalGroups(signal.NewGroup(7, 8, 9))),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
				destPorts, err := srcPort.Pipes().All()
				require.NoError(t, err)
				for _, destPort := range destPorts {
					assert.True(t, destPort.HasSignals())
					assert.Equal(t, 6, destPort.Signals().Len())
					allPayloads, err := destPort.Signals().AllPayloads()
					require.NoError(t, err)
					assert.Contains(t, allPayloads, 1)
					assert.Contains(t, allPayloads, 2)
					assert.Contains(t, allPayloads, 3)
				}
			},
		},
		{
			name:    "GUARDRAIL: cannot flush input port",
			srcPort: NewInput("input1").PutSignalGroups(signal.NewGroup(1, 2, 3)),
			assertions: func(t *testing.T, srcPort *Port) {
				require.Error(t, srcPort.ChainableErr())
				assert.Contains(t, srcPort.ChainableErr().Error(), "cannot flush input port")
				assert.Contains(t, srcPort.ChainableErr().Error(), "input1")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portAfter := tt.srcPort.Flush()
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
			}
		})
	}
}

func TestPort_SetLabels(t *testing.T) {
	tests := []struct {
		name       string
		port       *Port
		labels     labels.Map
		assertions func(t *testing.T, port *Port)
	}{
		{
			name: "set labels on new port",
			port: NewOutput("p1"),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 2, port.Labels().Len())
				assert.True(t, port.labels.HasAll("l1", "l2"))
			},
		},
		{
			name: "set labels replaces existing labels",
			port: NewOutput("p1").AddLabels(labels.Map{"old": "value"}),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 2, port.Labels().Len())
				assert.True(t, port.labels.HasAll("l1", "l2"))
				assert.False(t, port.labels.Has("old"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portAfter := tt.port.SetLabels(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
			}
		})
	}
}

func TestPort_AddLabels(t *testing.T) {
	tests := []struct {
		name       string
		port       *Port
		labels     labels.Map
		assertions func(t *testing.T, port *Port)
	}{
		{
			name: "add labels to new port",
			port: NewOutput("p1"),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 2, port.Labels().Len())
				assert.True(t, port.labels.HasAll("l1", "l2"))
			},
		},
		{
			name: "add labels merges with existing",
			port: NewOutput("p1").AddLabels(labels.Map{"existing": "label"}),
			labels: labels.Map{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 3, port.Labels().Len())
				assert.True(t, port.labels.HasAll("existing", "l1", "l2"))
			},
		},
		{
			name: "add labels updates existing key",
			port: NewOutput("p1").AddLabels(labels.Map{"l1": "old"}),
			labels: labels.Map{
				"l1": "new",
			},
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 1, port.Labels().Len())
				assert.True(t, port.labels.ValueIs("l1", "new"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portAfter := tt.port.AddLabels(tt.labels)
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
			}
		})
	}
}

func TestPort_AddLabel(t *testing.T) {
	tests := []struct {
		name       string
		port       *Port
		labelName  string
		labelValue string
		assertions func(t *testing.T, port *Port)
	}{
		{
			name:       "add single label to new port",
			port:       NewOutput("p1"),
			labelName:  "direction",
			labelValue: "in",
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 1, port.Labels().Len())
				assert.True(t, port.labels.ValueIs("direction", "in"))
			},
		},
		{
			name:       "add label merges with existing",
			port:       NewOutput("p1").AddLabel("existing", "label"),
			labelName:  "direction",
			labelValue: "in",
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 2, port.Labels().Len())
				assert.True(t, port.labels.HasAll("existing", "direction"))
			},
		},
		{
			name:       "add label updates existing key",
			port:       NewOutput("p1").AddLabel("direction", "in"),
			labelName:  "direction",
			labelValue: "out",
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 1, port.Labels().Len())
				assert.True(t, port.labels.ValueIs("direction", "out"))
			},
		},
		{
			name:       "chainable",
			port:       NewOutput("p1"),
			labelName:  "l1",
			labelValue: "v1",
			assertions: func(t *testing.T, port *Port) {
				result := port.AddLabel("l2", "v2").AddLabel("l3", "v3")
				assert.Equal(t, 3, result.Labels().Len())
				assert.True(t, result.labels.HasAll("l1", "l2", "l3"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portAfter := tt.port.AddLabel(tt.labelName, tt.labelValue)
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
			}
		})
	}
}

func TestPort_ClearLabels(t *testing.T) {
	tests := []struct {
		name       string
		port       *Port
		assertions func(t *testing.T, port *Port)
	}{
		{
			name: "clear labels from port with labels",
			port: NewOutput("p1").AddLabels(labels.Map{"k1": "v1", "k2": "v2"}),
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 0, port.Labels().Len())
				assert.False(t, port.Labels().Has("k1"))
				assert.False(t, port.Labels().Has("k2"))
			},
		},
		{
			name: "clear labels from port without labels",
			port: NewOutput("p1"),
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 0, port.Labels().Len())
			},
		},
		{
			name: "chainable",
			port: NewOutput("p1").AddLabels(labels.Map{"k1": "v1"}),
			assertions: func(t *testing.T, port *Port) {
				result := port.ClearLabels().AddLabel("k2", "v2")
				assert.Equal(t, 1, result.Labels().Len())
				assert.False(t, result.Labels().Has("k1"))
				assert.True(t, result.Labels().ValueIs("k2", "v2"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portAfter := tt.port.ClearLabels()
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
			}
		})
	}
}

func TestPort_WithoutLabels(t *testing.T) {
	tests := []struct {
		name           string
		port           *Port
		labelsToRemove []string
		assertions     func(t *testing.T, port *Port)
	}{
		{
			name:           "remove single label",
			port:           NewOutput("p1").AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}),
			labelsToRemove: []string{"k1"},
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 2, port.Labels().Len())
				assert.False(t, port.Labels().Has("k1"))
				assert.True(t, port.Labels().Has("k2"))
				assert.True(t, port.Labels().Has("k3"))
			},
		},
		{
			name:           "remove multiple labels",
			port:           NewOutput("p1").AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}),
			labelsToRemove: []string{"k1", "k2"},
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 1, port.Labels().Len())
				assert.False(t, port.Labels().Has("k1"))
				assert.False(t, port.Labels().Has("k2"))
				assert.True(t, port.Labels().ValueIs("k3", "v3"))
			},
		},
		{
			name:           "remove non-existent label",
			port:           NewOutput("p1").AddLabels(labels.Map{"k1": "v1"}),
			labelsToRemove: []string{"k2"},
			assertions: func(t *testing.T, port *Port) {
				assert.Equal(t, 1, port.Labels().Len())
				assert.True(t, port.Labels().ValueIs("k1", "v1"))
			},
		},
		{
			name:           "chainable",
			port:           NewOutput("p1").AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}),
			labelsToRemove: []string{"k1"},
			assertions: func(t *testing.T, port *Port) {
				result := port.WithoutLabels("k2").AddLabel("k4", "v4")
				assert.Equal(t, 2, result.Labels().Len())
				assert.False(t, result.Labels().Has("k1"))
				assert.False(t, result.Labels().Has("k2"))
				assert.True(t, result.Labels().ValueIs("k3", "v3"))
				assert.True(t, result.Labels().ValueIs("k4", "v4"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portAfter := tt.port.WithoutLabels(tt.labelsToRemove...)
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
			}
		})
	}
}

func TestPort_Pipes(t *testing.T) {
	tests := []struct {
		name              string
		port              *Port
		want              *Group
		wantErrContaining string
	}{
		{
			name: "no pipes",
			port: NewOutput("p"),
			want: NewGroup(),
		},
		{
			name: "with pipes",
			port: NewOutput("p1").PipeTo(
				NewInput("p2"),
				NewInput("p3"),
			),
			want: NewGroup().Add(NewInput("p2"), NewInput("p3")),
		},
		{
			name:              "with chain error",
			port:              NewOutput("p").WithChainableErr(errors.New("some error")),
			want:              NewGroup().WithChainableErr(errors.New("some error")),
			wantErrContaining: "some error",
		},
		{
			name:              "GUARDRAIL: input port cannot have pipes",
			port:              NewInput("input1"),
			want:              NewGroup().WithChainableErr(errors.New("port 'input1' is an input port and cannot have outbound pipes")),
			wantErrContaining: "input port",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.port.Pipes()
			if tt.wantErrContaining != "" {
				require.Error(t, got.ChainableErr())
				assert.Contains(t, got.ChainableErr().Error(), tt.wantErrContaining)
			} else {
				assert.NoError(t, got.ChainableErr())
			}
		})
	}
}

func TestPort_SignalsAccess(t *testing.T) {
	t.Run("FirstSignalPayload", func(t *testing.T) {
		port := NewOutput("p").PutSignalGroups(signal.NewGroup(4, 7, 6, 5))
		payload, err := port.Signals().FirstPayload()
		require.NoError(t, err)
		assert.Equal(t, 4, payload)
	})

	t.Run("FirstSignalPayloadOrNil", func(t *testing.T) {
		port := NewOutput("p").PutSignals(signal.New(123).WithChainableErr(errors.New("some error")))
		assert.Nil(t, port.Signals().FirstPayloadOrNil())
	})

	t.Run("FirstSignalPayloadOrDefault", func(t *testing.T) {
		port := NewOutput("p").PutSignals(signal.New(123).WithChainableErr(errors.New("some error")))
		assert.Equal(t, 888, port.Signals().FirstPayloadOrDefault(888))
	})

	t.Run("All with error", func(t *testing.T) {
		port := NewOutput("p").PutSignals(signal.New(123).WithChainableErr(errors.New("some error")))
		_, err := port.Signals().All()
		assert.Error(t, err)
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
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(1, 2, 3)),
				destPort: NewOutput("p2"),
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
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(1, 2, 3)),
				destPort: NewOutput("p2").PutSignalGroups(signal.NewGroup(1, 2, 3, 4, 5, 6)),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.NoError(t, err)
				assert.Equal(t, 9, destPortAfter.Signals().Len())
				assert.Equal(t, 3, srcPortAfter.Signals().Len())
			},
		},
		{
			name: "src with chain error",
			args: args{
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(1, 2, 3)).WithChainableErr(errors.New("some error")),
				destPort: NewOutput("p2"),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.Error(t, err)
				assert.Equal(t, 0, destPortAfter.Signals().Len())
				assert.Equal(t, 0, srcPortAfter.Signals().Len())
			},
		},
		{
			name: "dest with chain error",
			args: args{
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(1, 2, 3)),
				destPort: NewOutput("p2").WithChainableErr(errors.New("some error")),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.Error(t, err)
				assert.Equal(t, 0, destPortAfter.Signals().Len())
				assert.Equal(t, 3, srcPortAfter.Signals().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ForwardSignals(tt.args.srcPort, tt.args.destPort)
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
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(1, 2, 3)),
				destPort: NewOutput("p2"),
				predicate: func(signal *signal.Signal) bool {
					return signal.PayloadOrDefault(0).(int) > 0
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
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(7, 8, 9, 10, 11, 12, 13)),
				destPort: NewOutput("p2").PutSignalGroups(signal.NewGroup(1, 2, 3, 4, 5, 6)),
				predicate: func(signal *signal.Signal) bool {
					return signal.PayloadOrDefault(0).(int) > 10
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
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(7, 8, 9, 10, 11, 12, 13)),
				destPort: NewOutput("p2").PutSignalGroups(signal.NewGroup(1, 2, 3, 4, 5, 6)),
				predicate: func(signal *signal.Signal) bool {
					return signal.PayloadOrDefault(0).(int) > 99
				},
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.NoError(t, err)
				assert.Equal(t, 6, destPortAfter.Signals().Len())
				assert.Equal(t, 7, srcPortAfter.Signals().Len())
			},
		},
		{
			name: "src with chain error",
			args: args{
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(1, 2, 3)).WithChainableErr(errors.New("some error")),
				destPort: NewOutput("p2"),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.Error(t, err)
				assert.Equal(t, 0, destPortAfter.Signals().Len())
				assert.Equal(t, 0, srcPortAfter.Signals().Len())
			},
		},
		{
			name: "dest with chain error",
			args: args{
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(1, 2, 3)),
				destPort: NewOutput("p2").WithChainableErr(errors.New("some error")),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.Error(t, err)
				assert.Equal(t, 0, destPortAfter.Signals().Len())
				assert.Equal(t, 3, srcPortAfter.Signals().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ForwardWithFilter(tt.args.srcPort, tt.args.destPort, tt.args.predicate)
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
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(1, 2, 3)),
				destPort: NewOutput("p2"),
				mapperFunc: func(signal *signal.Signal) *signal.Signal {
					return signal.SetLabels(labels.Map{
						"l1": "v1",
					})
				},
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.NoError(t, err)
				assert.Equal(t, 3, destPortAfter.Signals().Len())
				assert.Equal(t, 3, srcPortAfter.Signals().Len())
				assert.True(t, destPortAfter.Signals().AllMatch(func(signal *signal.Signal) bool {
					return signal.Labels().ValueIs("l1", "v1")
				}))
			},
		},
		{
			name: "src with chain error",
			args: args{
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(1, 2, 3)).WithChainableErr(errors.New("some error")),
				destPort: NewOutput("p2"),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.Error(t, err)
				assert.Equal(t, 0, destPortAfter.Signals().Len())
				assert.Equal(t, 0, srcPortAfter.Signals().Len())
			},
		},
		{
			name: "dest with chain error",
			args: args{
				srcPort:  NewOutput("p1").PutSignalGroups(signal.NewGroup(1, 2, 3)),
				destPort: NewOutput("p2").WithChainableErr(errors.New("some error")),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				require.Error(t, err)
				assert.Equal(t, 0, destPortAfter.Signals().Len())
				assert.Equal(t, 3, srcPortAfter.Signals().Len())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ForwardWithMap(tt.args.srcPort, tt.args.destPort, tt.args.mapperFunc)
			if tt.assertions != nil {
				tt.assertions(t, tt.args.srcPort, tt.args.destPort, err)
			}
		})
	}
}

func TestPort_Chainability(t *testing.T) {
	t.Run("SetLabels called twice replaces all labels", func(t *testing.T) {
		p := NewOutput("p1").
			SetLabels(labels.Map{"k1": "v1", "k2": "v2"}).
			SetLabels(labels.Map{"k3": "v3"})

		assert.Equal(t, 1, p.Labels().Len())
		assert.False(t, p.Labels().Has("k1"), "k1 should be replaced")
		assert.False(t, p.Labels().Has("k2"), "k2 should be replaced")
		assert.True(t, p.Labels().ValueIs("k3", "v3"))
	})

	t.Run("AddLabels called twice merges labels", func(t *testing.T) {
		p := NewOutput("p1").
			AddLabels(labels.Map{"k1": "v1", "k2": "v2"}).
			AddLabels(labels.Map{"k3": "v3", "k2": "v2-updated"})

		assert.Equal(t, 3, p.Labels().Len())
		assert.True(t, p.Labels().ValueIs("k1", "v1"))
		assert.True(t, p.Labels().ValueIs("k2", "v2-updated"), "should update existing key")
		assert.True(t, p.Labels().ValueIs("k3", "v3"))
	})

	t.Run("mixed Set and Add operations", func(t *testing.T) {
		p := NewOutput("p1").
			AddLabel("k1", "v1").
			AddLabels(labels.Map{"k2": "v2", "k3": "v3"}).
			SetLabels(labels.Map{"k4": "v4"}). // Wipes k1, k2, k3
			AddLabel("k5", "v5")               // Merges with k4

		assert.Equal(t, 2, p.Labels().Len())
		assert.False(t, p.Labels().Has("k1"), "wiped by SetLabels")
		assert.False(t, p.Labels().Has("k2"), "wiped by SetLabels")
		assert.False(t, p.Labels().Has("k3"), "wiped by SetLabels")
		assert.True(t, p.Labels().ValueIs("k4", "v4"))
		assert.True(t, p.Labels().ValueIs("k5", "v5"))
	})

	t.Run("WithDescription replaces previous value", func(t *testing.T) {
		p := NewOutput("p1").
			WithDescription("first").
			WithDescription("second")

		assert.Equal(t, "second", p.Description())
	})

	t.Run("PutSignals called twice adds signals", func(t *testing.T) {
		p := NewOutput("p1").
			PutSignals(signal.New(1), signal.New(2)).
			PutSignals(signal.New(3))

		assert.Equal(t, 3, p.Signals().Len())
	})

	t.Run("Clear removes all signals", func(t *testing.T) {
		p := NewOutput("p1").
			PutSignals(signal.New(1), signal.New(2)).
			PutSignals(signal.New(3)).
			Clear()

		assert.Equal(t, 0, p.Signals().Len())
		assert.False(t, p.HasSignals())
	})

	t.Run("ClearLabels removes all labels", func(t *testing.T) {
		p := NewOutput("p1").
			AddLabels(labels.Map{"k1": "v1", "k2": "v2"}).
			ClearLabels().
			AddLabel("k3", "v3")

		assert.Equal(t, 1, p.Labels().Len())
		assert.False(t, p.Labels().Has("k1"))
		assert.False(t, p.Labels().Has("k2"))
		assert.True(t, p.Labels().ValueIs("k3", "v3"))
	})

	t.Run("WithoutLabels removes specific labels", func(t *testing.T) {
		p := NewOutput("p1").
			AddLabels(labels.Map{"k1": "v1", "k2": "v2", "k3": "v3"}).
			WithoutLabels("k1", "k2").
			AddLabel("k4", "v4")

		assert.Equal(t, 2, p.Labels().Len())
		assert.False(t, p.Labels().Has("k1"))
		assert.False(t, p.Labels().Has("k2"))
		assert.True(t, p.Labels().ValueIs("k3", "v3"))
		assert.True(t, p.Labels().ValueIs("k4", "v4"))
	})
}
