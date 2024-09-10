package signal

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		payload []any
	}
	tests := []struct {
		name string
		args args
		want *Signal
	}{
		{
			name: "nil payload",
			args: args{
				payload: nil,
			},
			want: &Signal{payload: nil},
		},
		{
			name: "empty slice",
			args: args{
				payload: []any{},
			},
			want: &Signal{payload: []any{}},
		},
		{
			name: "single payload",
			args: args{
				payload: []any{123},
			},
			want: &Signal{payload: []any{123}},
		},
		{
			name: "multiple payloads",
			args: args{
				payload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
			},
			want: &Signal{payload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.payload...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignal_Len(t *testing.T) {
	type fields struct {
		payload []any
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "nil payload",
			fields: fields{
				payload: nil,
			},
			want: 0,
		},
		{
			name: "empty slice",
			fields: fields{
				payload: []any{},
			},
			want: 0,
		},
		{
			name: "single payload",
			fields: fields{
				payload: []any{123},
			},
			want: 1,
		},
		{
			name: "multiple payloads",
			fields: fields{
				payload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signal{
				payload: tt.fields.payload,
			}
			if got := s.Len(); got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignal_Merge(t *testing.T) {

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
				payload: nil,
			},
		},
		{
			name: "a is nil",
			sigA: New(),
			sigB: New(12, 13),
			want: &Signal{
				payload: []any{12, 13},
			},
		},
		{
			name: "b is nil",
			sigA: New(14, 15),
			sigB: New(),
			want: &Signal{
				payload: []any{14, 15},
			},
		},
		{
			name: "single payloads",
			sigA: New(16),
			sigB: New(map[string]string{"k": "v"}),
			want: &Signal{
				payload: []any{16, map[string]string{"k": "v"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sigA.Merge(tt.sigB); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignal_Payload(t *testing.T) {
	type fields struct {
		payload []any
	}
	tests := []struct {
		name   string
		fields fields
		want   []any
	}{
		{
			name: "nil payload",
			fields: fields{
				payload: nil,
			},
			want: nil,
		},
		{
			name: "empty slice",
			fields: fields{
				payload: []any{},
			},
			want: []any{},
		},
		{
			name: "single payload",
			fields: fields{
				payload: []any{123},
			},
			want: []any{123},
		},
		{
			name: "multiple payloads",
			fields: fields{
				payload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
			},
			want: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signal{
				payload: tt.fields.payload,
			}
			if got := s.Payload(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Payload() = %v, want %v", got, tt.want)
			}
		})
	}
}
