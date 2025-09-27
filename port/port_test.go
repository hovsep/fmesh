package port

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/common"
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
			port: New("emptyPort"),
			want: false,
		},
		{
			name: "port has normal buffer",
			port: New("p").WithSignals(signal.New(123)),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.port.HasSignals())
		})
	}
}

func TestPort_Buffer(t *testing.T) {
	tests := []struct {
		name string
		port *Port
		want *signal.Group
	}{
		{
			name: "empty buffer",
			port: New("noSignal"),
			want: signal.NewGroup(),
		},
		{
			name: "with signal",
			port: New("p").WithSignals(signal.New(123)),
			want: signal.NewGroup(123),
		},
		{
			name: "with chain error",
			port: New("p").WithErr(errors.New("some error")),
			want: signal.NewGroup().WithErr(errors.New("some error")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.port.Buffer())
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
			before: New("p").WithSignals(signal.New(111)),
			after:  New("p"),
		},
		{
			name:   "cleaning empty port",
			before: New("emptyPort"),
			after:  New("emptyPort"),
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
	outputPorts := NewCollection().
		WithDefaultLabels(
			common.LabelsCollection{
				DirectionLabel: DirectionOut,
			}).With(
		NewIndexedGroup("out", 1, 3).PortsOrNil()...,
	)
	inputPorts := NewCollection().
		WithDefaultLabels(
			common.LabelsCollection{
				DirectionLabel: DirectionIn,
			}).With(
		NewIndexedGroup("in", 1, 3).PortsOrNil()...,
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
				assert.False(t, portAfter.HasErr())
				require.NoError(t, portAfter.Err())
				assert.Equal(t, 2, portAfter.Pipes().Len())
			},
		},
		{
			name:   "port must have direction label",
			before: New("out_without_dir"),
			args: args{
				toPorts: Ports{inputPorts.ByName("in1")},
			},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Empty(t, portAfter.Name())
				assert.True(t, portAfter.HasErr())
				assert.Error(t, portAfter.Err())
			},
		},
		{
			name:   "nil port is not allowed",
			before: outputPorts.ByName("out3"),
			args: args{
				toPorts: Ports{inputPorts.ByName("in2"), nil},
			},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Empty(t, portAfter.Name())
				assert.True(t, portAfter.HasErr())
				assert.Error(t, portAfter.Err())
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
				assert.Empty(t, portAfter.Name())
				assert.True(t, portAfter.HasErr())
				assert.Error(t, portAfter.Err())
			},
		},
		{
			name:   "piping to output ports is not allowed",
			before: outputPorts.ByName("out1"),
			args: args{
				toPorts: Ports{outputPorts.ByName("out2")},
			},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Empty(t, portAfter.Name())
				assert.True(t, portAfter.HasErr())
				assert.Error(t, portAfter.Err())
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
			port: New("emptyPort"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11), portAfter.Buffer())
			},
			args: args{
				signals: signal.NewGroup(11).SignalsOrNil(),
			},
		},
		{
			name: "multiple buffer to empty port",
			port: New("p"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12), portAfter.Buffer())
			},
			args: args{
				signals: signal.NewGroup(11, 12).SignalsOrNil(),
			},
		},
		{
			name: "single signal to port with single signal",
			port: New("p").WithSignals(signal.New(11)),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12), portAfter.Buffer())
			},
			args: args{
				signals: signal.NewGroup(12).SignalsOrNil(),
			},
		},
		{
			name: "single buffer to port with multiple buffer",
			port: New("p").WithSignalGroups(signal.NewGroup(11, 12)),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(11, 12, 13), portAfter.Buffer())
			},
			args: args{
				signals: signal.NewGroup(13).SignalsOrNil(),
			},
		},
		{
			name: "multiple buffer to port with multiple buffer",
			port: New("p").WithSignalGroups(signal.NewGroup(55, 66)),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Equal(t, signal.NewGroup(55, 66, 13, 14), portAfter.Buffer())
			},
			args: args{
				signals: signal.NewGroup(13, 14).SignalsOrNil(),
			},
		},
		{
			name: "chain error propagated from buffer",
			port: New("p"),
			assertions: func(t *testing.T, portAfter *Port) {
				assert.Zero(t, portAfter.Buffer().Len())
				assert.True(t, portAfter.Buffer().HasErr())
			},
			args: args{
				signals: signal.Signals{signal.New(111).WithErr(errors.New("some error in signal"))},
			},
		},
		{
			name: "with chain error",
			port: New("p").WithErr(errors.New("some error in port")),
			args: args{
				signals: signal.Signals{signal.New(123)},
			},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.True(t, portAfter.HasErr())
				assert.Zero(t, portAfter.Buffer().Len())
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
			want: New(""),
		},
		{
			name: "with name",
			args: args{
				name: "p1",
			},
			want: New("p1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.args.name))
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
			port: New("p"),
			want: false,
		},
		{
			name: "with pipes",
			port: New("p1").WithLabels(common.LabelsCollection{
				DirectionLabel: DirectionOut,
			}).PipeTo(New("p2").WithLabels(common.LabelsCollection{
				DirectionLabel: DirectionIn,
			})),
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
			name:    "port with buffer and no pipes is not flushed",
			srcPort: New("p").WithSignalGroups(signal.NewGroup(1, 2, 3)),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.True(t, srcPort.HasSignals())
				assert.Equal(t, 3, srcPort.Buffer().Len())
				assert.False(t, srcPort.HasPipes())
			},
		},
		{
			name: "empty port with pipes is not flushed",
			srcPort: New("p").
				WithLabels(
					common.LabelsCollection{
						DirectionLabel: DirectionOut,
					}).PipeTo(
				New("p1").
					WithLabels(
						common.LabelsCollection{
							DirectionLabel: DirectionIn,
						}), New("p2").
					WithLabels(
						common.LabelsCollection{
							DirectionLabel: DirectionIn,
						})),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
			},
		},
		{
			name: "flush to empty ports",
			srcPort: New("p").WithLabels(common.LabelsCollection{
				DirectionLabel: DirectionOut,
			}).WithSignalGroups(signal.NewGroup(1, 2, 3)).
				PipeTo(
					New("p1").WithLabels(common.LabelsCollection{
						DirectionLabel: DirectionIn,
					}),
					New("p2").WithLabels(common.LabelsCollection{
						DirectionLabel: DirectionIn,
					})),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
				for _, destPort := range srcPort.Pipes().PortsOrNil() {
					assert.True(t, destPort.HasSignals())
					assert.Equal(t, 3, destPort.Buffer().Len())
					allPayloads, err := destPort.AllSignalsPayloads()
					require.NoError(t, err)
					assert.Contains(t, allPayloads, 1)
					assert.Contains(t, allPayloads, 2)
					assert.Contains(t, allPayloads, 3)
				}
			},
		},
		{
			name: "flush to non empty ports",
			srcPort: New("p").WithLabels(common.LabelsCollection{
				DirectionLabel: DirectionOut,
			}).
				WithSignalGroups(signal.NewGroup(1, 2, 3)).
				PipeTo(
					New("p1").WithLabels(common.LabelsCollection{
						DirectionLabel: DirectionIn,
					}).WithSignalGroups(signal.NewGroup(4, 5, 6)),
					New("p2").WithLabels(common.LabelsCollection{
						DirectionLabel: DirectionIn,
					}).WithSignalGroups(signal.NewGroup(7, 8, 9))),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
				for _, destPort := range srcPort.Pipes().PortsOrNil() {
					assert.True(t, destPort.HasSignals())
					assert.Equal(t, 6, destPort.Buffer().Len())
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
			portAfter := tt.srcPort.Flush()
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
			}
		})
	}
}

func TestPort_WithLabels(t *testing.T) {
	type args struct {
		labels common.LabelsCollection
	}
	tests := []struct {
		name       string
		port       *Port
		args       args
		assertions func(t *testing.T, port *Port)
	}{
		{
			name: "happy path",
			port: New("p1"),
			args: args{
				labels: common.LabelsCollection{
					"l1": "v1",
					"l2": "v2",
				},
			},
			assertions: func(t *testing.T, port *Port) {
				assert.Len(t, port.Labels(), 2)
				assert.True(t, port.HasAllLabels("l1", "l2"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portAfter := tt.port.WithLabels(tt.args.labels)
			if tt.assertions != nil {
				tt.assertions(t, portAfter)
			}
		})
	}
}

func TestPort_Pipes(t *testing.T) {
	tests := []struct {
		name string
		port *Port
		want *Group
	}{
		{
			name: "no pipes",
			port: New("p"),
			want: NewGroup(),
		},
		{
			name: "with pipes",
			port: New("p1").
				WithLabels(common.LabelsCollection{
					DirectionLabel: DirectionOut,
				}).PipeTo(
				New("p2").
					WithLabels(common.LabelsCollection{
						DirectionLabel: DirectionIn,
					}), New("p3").
					WithLabels(common.LabelsCollection{
						DirectionLabel: DirectionIn,
					})),
			want: NewGroup("p2", "p3").WithPortLabels(common.LabelsCollection{
				DirectionLabel: DirectionIn,
			}),
		},
		{
			name: "with chain error",
			port: New("p").WithErr(errors.New("some error")),
			want: NewGroup().WithErr(errors.New("some error")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.port.Pipes())
		})
	}
}

func TestPort_ShortcutGetters(t *testing.T) {
	t.Run("FirstSignalPayload", func(t *testing.T) {
		port := New("p").WithSignalGroups(signal.NewGroup(4, 7, 6, 5))
		payload, err := port.FirstSignalPayload()
		require.NoError(t, err)
		assert.Equal(t, 4, payload)
	})

	t.Run("FirstSignalPayloadOrNil", func(t *testing.T) {
		port := New("p").WithSignals(signal.New(123).WithErr(errors.New("some error")))
		assert.Nil(t, port.FirstSignalPayloadOrNil())
	})

	t.Run("FirstSignalPayloadOrDefault", func(t *testing.T) {
		port := New("p").WithSignals(signal.New(123).WithErr(errors.New("some error")))
		assert.Equal(t, 888, port.FirstSignalPayloadOrDefault(888))
	})

	t.Run("AllSignalsOrNil", func(t *testing.T) {
		port := New("p").WithSignals(signal.New(123).WithErr(errors.New("some error")))
		assert.Nil(t, port.AllSignalsOrNil())
	})

	t.Run("AllSignalsOrDefault", func(t *testing.T) {
		port := New("p").WithSignals(signal.New(123).WithErr(errors.New("some error")))
		assert.Equal(t, signal.NewGroup(999).SignalsOrNil(), port.AllSignalsOrDefault(signal.NewGroup(999).SignalsOrNil()))
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
				srcPort:  New("p1").WithSignalGroups(signal.NewGroup(1, 2, 3)),
				destPort: New("p2"),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 3, destPortAfter.Buffer().Len())
				assert.Equal(t, 3, srcPortAfter.Buffer().Len())
			},
		},
		{
			name: "signals are added to dest port",
			args: args{
				srcPort:  New("p1").WithSignalGroups(signal.NewGroup(1, 2, 3)),
				destPort: New("p2").WithSignalGroups(signal.NewGroup(1, 2, 3, 4, 5, 6)),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 9, destPortAfter.Buffer().Len())
				assert.Equal(t, 3, srcPortAfter.Buffer().Len())
			},
		},
		{
			name: "src with chain error",
			args: args{
				srcPort:  New("p1").WithSignalGroups(signal.NewGroup(1, 2, 3)).WithErr(errors.New("some error")),
				destPort: New("p2"),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				assert.Error(t, err)
				assert.Equal(t, 0, destPortAfter.Buffer().Len())
				assert.Equal(t, 0, srcPortAfter.Buffer().Len())
			},
		},
		{
			name: "dest with chain error",
			args: args{
				srcPort:  New("p1").WithSignalGroups(signal.NewGroup(1, 2, 3)),
				destPort: New("p2").WithErr(errors.New("some error")),
			},
			assertions: func(t *testing.T, srcPortAfter, destPortAfter *Port, err error) {
				assert.Error(t, err)
				assert.Equal(t, 0, destPortAfter.Buffer().Len())
				assert.Equal(t, 3, srcPortAfter.Buffer().Len())
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
