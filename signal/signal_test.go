package signal

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		payloads []any
	}
	tests := []struct {
		name string
		args args
		want *Signal
	}{
		{
			name: "nil payloads",
			args: args{
				payloads: nil,
			},
			want: &Signal{payloads: nil},
		},
		{
			name: "empty slice",
			args: args{
				payloads: []any{},
			},
			want: &Signal{payloads: []any{}},
		},
		{
			name: "single payloads",
			args: args{
				payloads: []any{123},
			},
			want: &Signal{payloads: []any{123}},
		},
		{
			name: "multiple payloads",
			args: args{
				payloads: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
			},
			want: &Signal{payloads: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.payloads...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignal_Len(t *testing.T) {
	tests := []struct {
		name   string
		signal *Signal
		want   int
	}{
		{
			name:   "no args",
			signal: New(),
			want:   0,
		},
		{
			name:   "nil payload is valid",
			signal: New(nil),
			want:   1,
		},
		{
			name:   "single payload",
			signal: New(123),
			want:   1,
		},
		{
			name:   "multiple payloads",
			signal: New(123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil),
			want:   6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.signal.Len(); got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignal_Combine(t *testing.T) {
	tests := []struct {
		name string
		sigA *Signal
		sigB *Signal
		want *Signal
	}{
		{
			name: "two nils",
			sigA: New(),
			sigB: New(),
			want: &Signal{
				payloads: nil,
			},
		},
		{
			name: "a is nil",
			sigA: New(),
			sigB: New(12, 13),
			want: &Signal{
				payloads: []any{12, 13},
			},
		},
		{
			name: "b is nil",
			sigA: New(14, 15),
			sigB: New(),
			want: &Signal{
				payloads: []any{14, 15},
			},
		},
		{
			name: "single payloads",
			sigA: New(16),
			sigB: New(map[string]string{"k": "v"}),
			want: &Signal{
				payloads: []any{16, map[string]string{"k": "v"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sigA.Combine(tt.sigB); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Combine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignal_Payloads(t *testing.T) {
	tests := []struct {
		name   string
		signal *Signal
		want   []any
	}{
		{
			name:   "no arg",
			signal: New(),
			want:   nil,
		},
		{
			name:   "nil payload",
			signal: New(nil),
			want:   []any{nil},
		},

		{
			name:   "single payload",
			signal: New(123),
			want:   []any{123},
		},
		{
			name:   "multiple payloads",
			signal: New(123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil),
			want:   []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.signal.Payloads(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Payloads() = %v, want %v", got, tt.want)
			}
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
			name:      "no arg",
			signal:    New(),
			wantPanic: true,
		},
		{
			name:   "nil payload",
			signal: New(nil),
			want:   nil,
		},
		{
			name:   "single payload",
			signal: New(123),
			want:   123,
		},
		{
			name:      "multiple payloads",
			signal:    New(1, 2),
			wantPanic: true,
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
			if got := tt.signal.Payload(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Payloads() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignal_HasPayload(t *testing.T) {
	tests := []struct {
		name   string
		signal *Signal
		want   bool
	}{
		{
			name:   "has payload",
			signal: New(123),
			want:   true,
		},
		{
			name:   "has no payload",
			signal: New(),
			want:   false,
		},
		{
			name:   "nil payload is valid",
			signal: New(nil),
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.signal.HasPayload(); got != tt.want {
				t.Errorf("HasPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}
