package signal

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGroup(t *testing.T) {
	type args struct {
		payloads []any
	}
	tests := []struct {
		name       string
		args       args
		assertions func(t *testing.T, group *Group)
	}{
		{
			name: "no payloads",
			args: args{
				payloads: nil,
			},
			assertions: func(t *testing.T, group *Group) {
				signals, err := group.Signals()
				assert.NoError(t, err)
				assert.Len(t, signals, 0)
			},
		},
		{
			name: "with payloads",
			args: args{
				payloads: []any{1, nil, 3},
			},
			assertions: func(t *testing.T, group *Group) {
				signals, err := group.Signals()
				assert.NoError(t, err)
				assert.Len(t, signals, 3)
				assert.Contains(t, signals, New(1))
				assert.Contains(t, signals, New(nil))
				assert.Contains(t, signals, New(3))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := NewGroup(tt.args.payloads...)
			if tt.assertions != nil {
				tt.assertions(t, group)
			}
		})
	}
}

func TestGroup_FirstPayload(t *testing.T) {
	tests := []struct {
		name            string
		group           *Group
		want            any
		wantErrorString string
		wantPanic       bool
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
					_, _ = tt.group.FirstPayload()
				})
			} else {
				got, err := tt.group.FirstPayload()
				if tt.wantErrorString != "" {
					assert.Error(t, err)
					assert.EqualError(t, err, tt.wantErrorString)
				} else {
					assert.Equal(t, tt.want, got)
				}
			}
		})
	}
}

func TestGroup_AllPayloads(t *testing.T) {
	tests := []struct {
		name            string
		group           *Group
		want            []any
		wantErrorString string
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
			got, err := tt.group.AllPayloads()
			if tt.wantErrorString != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErrorString)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGroup_With(t *testing.T) {
	type args struct {
		signals []*Signal
	}
	tests := []struct {
		name  string
		group *Group
		args  args
		want  *Group
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
				signals: NewGroup(3, 4, 5).SignalsOrNil(),
			},
			want: NewGroup(3, 4, 5),
		},
		{
			name:  "addition to group",
			group: NewGroup(1, 2, 3),
			args: args{
				signals: NewGroup(4, 5, 6).SignalsOrNil(),
			},
			want: NewGroup(1, 2, 3, 4, 5, 6),
		},
		{
			name: "error handling",
			group: NewGroup(1, 2, 3).
				With(New("valid before invalid")).
				With(nil).
				WithPayloads(4, 5, 6),
			args: args{
				signals: NewGroup(7, nil, 9).SignalsOrNil(),
			},
			want: NewGroup(1, 2, 3, "valid before invalid").WithError(errors.New("signal is nil")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.With(tt.args.signals...)
			if tt.want.HasError() {
				assert.Error(t, got.Error())
				assert.EqualError(t, got.Error(), tt.want.Error().Error())
			} else {
				assert.NoError(t, got.Error())
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_WithPayloads(t *testing.T) {
	type args struct {
		payloads []any
	}
	tests := []struct {
		name  string
		group *Group
		args  args
		want  *Group
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
