package port

import (
	"errors"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"testing"
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
			port: New("p").WithChainError(errors.New("some error")),
			want: signal.NewGroup().WithChainError(errors.New("some error")),
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
	p1, p2, p3, p4 := New("p1"), New("p2"), New("p3"), New("p4")

	type args struct {
		toPorts Ports
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
			after:  New("p1").PipeTo(p2, p3),
			args: args{
				toPorts: Ports{p2, p3},
			},
		},
		{
			name:   "invalid ports are ignored",
			before: p4,
			after:  New("p4").PipeTo(p2),
			args: args{
				toPorts: Ports{p2, nil},
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
				assert.True(t, portAfter.Buffer().HasChainError())
			},
			args: args{
				signals: signal.Signals{signal.New(111).WithChainError(errors.New("some error in signal"))},
			},
		},
		{
			name: "with chain error",
			port: New("p").WithChainError(errors.New("some error in port")),
			args: args{
				signals: signal.Signals{signal.New(123)},
			},
			assertions: func(t *testing.T, portAfter *Port) {
				assert.True(t, portAfter.HasChainError())
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
			port: New("p1").PipeTo(New("p2")),
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
				assert.Equal(t, srcPort.Buffer().Len(), 3)
				assert.False(t, srcPort.HasPipes())
			},
		},
		{
			name:    "empty port with pipes is not flushed",
			srcPort: New("p").PipeTo(New("p1"), New("p2")),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
			},
		},
		{
			name: "flush to empty ports",
			srcPort: New("p").WithSignalGroups(signal.NewGroup(1, 2, 3)).
				PipeTo(
					New("p1"),
					New("p2")),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
				for _, destPort := range srcPort.Pipes().PortsOrNil() {
					assert.True(t, destPort.HasSignals())
					assert.Equal(t, destPort.Buffer().Len(), 3)
					allPayloads, err := destPort.AllSignalsPayloads()
					assert.NoError(t, err)
					assert.Contains(t, allPayloads, 1)
					assert.Contains(t, allPayloads, 2)
					assert.Contains(t, allPayloads, 3)
				}
			},
		},
		{
			name: "flush to non empty ports",
			srcPort: New("p").WithSignalGroups(signal.NewGroup(1, 2, 3)).
				PipeTo(
					New("p1").WithSignalGroups(signal.NewGroup(4, 5, 6)),
					New("p2").WithSignalGroups(signal.NewGroup(7, 8, 9))),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
				for _, destPort := range srcPort.Pipes().PortsOrNil() {
					assert.True(t, destPort.HasSignals())
					assert.Equal(t, destPort.Buffer().Len(), 6)
					allPayloads, err := destPort.AllSignalsPayloads()
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
			tt.srcPort.Flush()
			if tt.assertions != nil {
				tt.assertions(t, tt.srcPort)
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
			port: New("p1").PipeTo(New("p2"), New("p3")),
			want: NewGroup("p2", "p3"),
		},
		{
			name: "with chain error",
			port: New("p").WithChainError(errors.New("some error")),
			want: NewGroup().WithChainError(errors.New("some error")),
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
		assert.NoError(t, err)
		assert.Equal(t, 4, payload)
	})

	t.Run("FirstSignalPayloadOrNil", func(t *testing.T) {
		port := New("p").WithSignals(signal.New(123).WithChainError(errors.New("some error")))
		assert.Nil(t, port.FirstSignalPayloadOrNil())
	})

	t.Run("FirstSignalPayloadOrDefault", func(t *testing.T) {
		port := New("p").WithSignals(signal.New(123).WithChainError(errors.New("some error")))
		assert.Equal(t, 888, port.FirstSignalPayloadOrDefault(888))
	})

	t.Run("AllSignalsOrNil", func(t *testing.T) {
		port := New("p").WithSignals(signal.New(123).WithChainError(errors.New("some error")))
		assert.Nil(t, port.AllSignalsOrNil())
	})

	t.Run("AllSignalsOrDefault", func(t *testing.T) {
		port := New("p").WithSignals(signal.New(123).WithChainError(errors.New("some error")))
		assert.Equal(t, signal.NewGroup(999).SignalsOrNil(), port.AllSignalsOrDefault(signal.NewGroup(999).SignalsOrNil()))
	})
}
