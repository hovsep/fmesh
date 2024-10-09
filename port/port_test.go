package port

import (
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
			name: "port has normal signals",
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

func TestPort_Signals(t *testing.T) {
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
			port: New("p").WithSignals(signal.New(123)),
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
			after:  New("p1").withPipes(p2, p3),
			args: args{
				toPorts: []*Port{p2, p3},
			},
		},
		{
			name:   "invalid ports are ignored",
			before: p4,
			after:  New("p4").withPipes(p2),
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
	type args struct {
		signals []*signal.Signal
	}
	tests := []struct {
		name         string
		port         *Port
		signalsAfter signal.Group
		args         args
	}{
		{
			name:         "single signal to empty port",
			port:         New("emptyPort"),
			signalsAfter: signal.NewGroup(11),
			args: args{
				signals: signal.NewGroup(11),
			},
		},
		{
			name:         "multiple signals to empty port",
			port:         New("p"),
			signalsAfter: signal.NewGroup(11, 12),
			args: args{
				signals: signal.NewGroup(11, 12),
			},
		},
		{
			name:         "single signal to port with single signal",
			port:         New("p").WithSignals(signal.New(11)),
			signalsAfter: signal.NewGroup(11, 12),
			args: args{
				signals: signal.NewGroup(12),
			},
		},
		{
			name:         "single signals to port with multiple signals",
			port:         New("p").WithSignals(signal.NewGroup(11, 12)...),
			signalsAfter: signal.NewGroup(11, 12, 13),
			args: args{
				signals: signal.NewGroup(13),
			},
		},
		{
			name:         "multiple signals to port with multiple signals",
			port:         New("p").WithSignals(signal.NewGroup(55, 66)...),
			signalsAfter: signal.NewGroup(55, 66, 13, 14), //Notice LIFO order
			args: args{
				signals: signal.NewGroup(13, 14),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.port.PutSignals(tt.args.signals...)
			assert.ElementsMatch(t, tt.signalsAfter, tt.port.Signals())
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
			port: New("p1").withPipes(New("p2")),
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
			srcPort: New("p").WithSignals(signal.NewGroup(1, 2, 3)...),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.True(t, srcPort.HasSignals())
				assert.Len(t, srcPort.Signals(), 3)
				assert.False(t, srcPort.HasPipes())
			},
		},
		{
			name:    "empty port with pipes is not flushed",
			srcPort: New("p").withPipes(New("p1"), New("p2")),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
			},
		},
		{
			name: "flush to empty ports",
			srcPort: New("p").WithSignals(signal.NewGroup(1, 2, 3)...).
				withPipes(
					New("p1"),
					New("p2")),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
				for _, destPort := range srcPort.pipes {
					assert.True(t, destPort.HasSignals())
					assert.Len(t, destPort.Signals(), 3)
					assert.Contains(t, destPort.Signals().AllPayloads(), 1)
					assert.Contains(t, destPort.Signals().AllPayloads(), 2)
					assert.Contains(t, destPort.Signals().AllPayloads(), 3)
				}
			},
		},
		{
			name: "flush to non empty ports",
			srcPort: New("p").WithSignals(signal.NewGroup(1, 2, 3)...).
				withPipes(
					New("p1").WithSignals(signal.NewGroup(4, 5, 6)...),
					New("p2").WithSignals(signal.NewGroup(7, 8, 9)...)),
			assertions: func(t *testing.T, srcPort *Port) {
				assert.False(t, srcPort.HasSignals())
				assert.True(t, srcPort.HasPipes())
				for _, destPort := range srcPort.pipes {
					assert.True(t, destPort.HasSignals())
					assert.Len(t, destPort.Signals(), 6)
					assert.Contains(t, destPort.Signals().AllPayloads(), 1)
					assert.Contains(t, destPort.Signals().AllPayloads(), 2)
					assert.Contains(t, destPort.Signals().AllPayloads(), 3)
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
