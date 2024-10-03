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

func TestGroup_With(t *testing.T) {
	type args struct {
		signals []*Signal
	}
	tests := []struct {
		name  string
		group Group
		args  args
		want  Group
	}{
		{
			name:  "no addition to empty group",
			group: NewGroup(),
			args: args{
				signals: nil,
			},
			want: NewGroup(),
		},
		{
			name:  "no addition to group",
			group: NewGroup(1, 2, 3),
			args: args{
				signals: nil,
			},
			want: NewGroup(1, 2, 3),
		},
		{
			name:  "addition to empty group",
			group: NewGroup(),
			args: args{
				signals: NewGroup(3, 4, 5),
			},
			want: NewGroup(3, 4, 5),
		},
		{
			name:  "addition to group",
			group: NewGroup(1, 2, 3),
			args: args{
				signals: NewGroup(4, 5, 6),
			},
			want: NewGroup(1, 2, 3, 4, 5, 6),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.group.With(tt.args.signals...))
		})
	}
}

func TestGroup_WithPayloads(t *testing.T) {
	type args struct {
		payloads []any
	}
	tests := []struct {
		name  string
		group Group
		args  args
		want  Group
	}{
		{
			name:  "no addition to empty group",
			group: NewGroup(),
			args: args{
				payloads: nil,
			},
			want: NewGroup(),
		},
		{
			name:  "addition to empty group",
			group: NewGroup(),
			args: args{
				payloads: []any{1, 2, 3},
			},
			want: NewGroup(1, 2, 3),
		},
		{
			name:  "no addition to group",
			group: NewGroup(1, 2, 3),
			args: args{
				payloads: nil,
			},
			want: NewGroup(1, 2, 3),
		},
		{
			name:  "addition to group",
			group: NewGroup(1, 2, 3),
			args: args{
				payloads: []any{4, 5, 6},
			},
			want: NewGroup(1, 2, 3, 4, 5, 6),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.group.WithPayloads(tt.args.payloads...))
		})
	}
}
