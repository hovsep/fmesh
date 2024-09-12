package port

import (
	"github.com/hovsep/fmesh/signal"
	"reflect"
	"testing"
)

func TestNewPipe(t *testing.T) {
	p1, p2 := NewPort("p1"), NewPort("p2")

	type args struct {
		from *Port
		to   *Port
	}
	tests := []struct {
		name string
		args args
		want *Pipe
	}{
		{
			name: "happy path",
			args: args{
				from: p1,
				to:   p2,
			},
			want: &Pipe{
				From: p1,
				To:   p2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPipe(tt.args.from, tt.args.to); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipe_Flush(t *testing.T) {
	portWithSignal := NewPort("portWithSignal")
	portWithSignal.PutSignal(signal.New(777))

	portWithMultipleSignals := NewPort("portWithMultipleSignals")
	portWithMultipleSignals.PutSignal(signal.New(11, 12))

	emptyPort := NewPort("emptyPort")

	tests := []struct {
		name   string
		before *Pipe
		after  *Pipe
	}{
		{
			name:   "flush to empty port",
			before: NewPipe(portWithSignal, emptyPort),
			after: &Pipe{
				From: &Port{
					name:   "portWithSignal",
					signal: signal.New(777), //Flush does not clear source port
				},
				To: &Port{
					name:   "emptyPort",
					signal: signal.New(777),
				},
			},
		},
		{
			name:   "flush to port with signal",
			before: NewPipe(portWithSignal, portWithMultipleSignals),
			after: &Pipe{
				From: portWithSignal,
				To: &Port{
					name:   "portWithMultipleSignals",
					signal: signal.New(777, 11, 12),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before.Flush()
			if !reflect.DeepEqual(tt.before, tt.after) {
				t.Errorf("Flush() = %v, want %v", tt.before, tt.after)
			}
		})
	}
}
