package port

import (
	"github.com/hovsep/fmesh/signal"
	"reflect"
	"testing"
)

func TestNewPipe(t *testing.T) {
	p1, p2 := NewPort(), NewPort()

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
	portWithSignal := NewPort()
	portWithSignal.PutSignal(signal.New(777))

	portWithMultipleSignals := NewPort()
	portWithMultipleSignals.PutSignal(signal.New(11, 12))

	emptyPort := NewPort()

	tests := []struct {
		name   string
		before *Pipe
		after  *Pipe
	}{
		{
			name:   "flush to empty port",
			before: NewPipe(portWithSignal, emptyPort),
			after:  NewPipe(portWithSignal, portWithSignal),
		},
		{
			name:   "flush to port with signal",
			before: NewPipe(portWithSignal, portWithMultipleSignals),
			after: NewPipe(portWithSignal, &Port{
				signal: signal.New(777, 11, 12),
			}),
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
