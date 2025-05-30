package signal

import (
	"errors"
	"github.com/hovsep/fmesh/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				payload:   []any{nil},
				Chainable: &common.Chainable{},
			},
		},
		{
			name: "with payload",
			args: args{
				payload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
			},
			want: &Signal{
				payload:   []any{[]any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil}},
				Chainable: &common.Chainable{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.args.payload))
		})
	}
}

func TestSignal_Payload(t *testing.T) {
	tests := []struct {
		name            string
		signal          *Signal
		want            any
		wantErrorString string
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
		{
			name:            "with error in chain",
			signal:          New(123).WithErr(errors.New("some error in chain")),
			want:            nil,
			wantErrorString: "some error in chain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.signal.Payload()
			if tt.wantErrorString != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErrorString)
			} else {
				assert.Equal(t, tt.want, got)
			}

		})
	}
}

func TestSignal_PayloadOrNil(t *testing.T) {
	tests := []struct {
		name   string
		signal *Signal
		want   any
	}{
		{
			name:   "payload returned",
			signal: New(123),
			want:   123,
		},
		{
			name:   "nil returned",
			signal: New(123).WithErr(errors.New("some error in chain")),
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.signal.PayloadOrNil())
		})
	}
}
