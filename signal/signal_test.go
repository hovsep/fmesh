package signal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		payload any
	}
	tests := []struct {
		name string
		args args
		want *Signal
	}{
		{
			name: "nil payload is valid",
			args: args{
				payload: nil,
			},
			want: &Signal{
				payload: []any{nil},
			},
		},
		{
			name: "with payload",
			args: args{
				payload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
			},
			want: &Signal{payload: []any{
				[]any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.payload)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSignal_Payload(t *testing.T) {
	tests := []struct {
		name      string
		signal    *Signal
		want      any
		wantPanic bool
	}{
		{
			name:   "nil payload is valid",
			signal: New(nil),
			want:   nil,
		},
		{
			name:   "with payload",
			signal: New(123),
			want:   123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Errorf("The code did not panic")
				}

				if !tt.wantPanic && r != nil {
					t.Errorf("The code unexpectedly paniced")
				}
			}()
			got := tt.signal.Payload()

			assert.Equal(t, tt.want, got)
		})
	}
}
