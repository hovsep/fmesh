package signal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGroup(t *testing.T) {
	type args struct {
		payloads []any
	}
	tests := []struct {
		name string
		args args
		want Group
	}{
		{
			name: "no payloads",
			args: args{
				payloads: nil,
			},
			want: Group{},
		},
		{
			name: "with payloads",
			args: args{
				payloads: []any{1, nil, 3},
			},
			want: Group{
				&Signal{
					payload: []any{1},
				},
				&Signal{
					payload: []any{nil},
				},
				&Signal{
					payload: []any{3},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewGroup(tt.args.payloads...))
		})
	}
}

func TestGroup_FirstPayload(t *testing.T) {
	tests := []struct {
		name      string
		group     Group
		want      any
		wantPanic bool
	}{
		{
			name:      "empty group",
			group:     NewGroup(),
			want:      nil,
			wantPanic: true,
		},
		{
			name:      "first is nil",
			group:     NewGroup(nil, 123),
			want:      nil,
			wantPanic: false,
		},
		{
			name:      "first is not nil",
			group:     NewGroup([]string{"1", "2"}, 123),
			want:      []string{"1", "2"},
			wantPanic: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					tt.group.FirstPayload()
				})
			} else {
				assert.Equal(t, tt.want, tt.group.FirstPayload())
			}
		})
	}
}

func TestGroup_AllPayloads(t *testing.T) {
	tests := []struct {
		name  string
		group Group
		want  []any
	}{
		{
			name:  "empty group",
			group: NewGroup(),
			want:  []any{},
		},
		{
			name:  "with payloads",
			group: NewGroup(1, nil, 3, []int{4, 5, 6}, map[byte]byte{7: 8}),
			want:  []any{1, nil, 3, []int{4, 5, 6}, map[byte]byte{7: 8}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.group.AllPayloads())
		})
	}
}
