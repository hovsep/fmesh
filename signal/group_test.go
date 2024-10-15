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
	}{
		{
			name:            "empty group",
			group:           NewGroup(),
			want:            nil,
			wantErrorString: "group has no signals",
		},
		{
			name:  "first is nil",
			group: NewGroup(nil, 123),
			want:  nil,
		},
		{
			name:  "first is not nil",
			group: NewGroup([]string{"1", "2"}, 123),
			want:  []string{"1", "2"},
		},
		{
			name:            "with error in chain",
			group:           NewGroup(3, 4, 5).WithError(errors.New("some error in chain")),
			want:            nil,
			wantErrorString: "some error in chain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.group.FirstPayload()
			if tt.wantErrorString != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErrorString)
			} else {
				assert.Equal(t, tt.want, got)
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
		{
			name:            "with error in chain",
			group:           NewGroup(1, 2, 3).WithError(errors.New("some error in chain")),
			want:            nil,
			wantErrorString: "some error in chain",
		},
		{
			name:            "with error in signal",
			group:           NewGroup().withSignals([]*Signal{New(33).WithError(errors.New("some error in signal"))}),
			want:            nil,
			wantErrorString: "some error in signal",
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
			name: "with error in chain",
			group: NewGroup(1, 2, 3).
				With(New("valid before invalid")).
				With(nil).
				WithPayloads(4, 5, 6),
			args: args{
				signals: NewGroup(7, nil, 9).SignalsOrNil(),
			},
			want: NewGroup(1, 2, 3, "valid before invalid").WithError(errors.New("signal is nil")),
		},
		{
			name:  "with error in signal",
			group: NewGroup(1, 2, 3).With(New(44).WithError(errors.New("some error in signal"))),
			args: args{
				signals: []*Signal{New(456)},
			},
			want: NewGroup(1, 2, 3).
				With(New(44).
										WithError(errors.New("some error in signal"))).
				WithError(errors.New("some error in signal")), // error propagated from signal to group
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

func TestGroup_First(t *testing.T) {
	tests := []struct {
		name  string
		group *Group
		want  *Signal
	}{
		{
			name:  "empty group",
			group: NewGroup(),
			want:  New(nil).WithError(errors.New("group has no signals")),
		},
		{
			name:  "happy path",
			group: NewGroup(3, 5, 7),
			want:  New(3),
		},
		{
			name:  "with error in chain",
			group: NewGroup(1, 2, 3).WithError(errors.New("some error in chain")),
			want:  New(nil).WithError(errors.New("some error in chain")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.First()
			if tt.want.HasError() {
				assert.True(t, got.HasError())
				assert.Error(t, got.Error())
				assert.EqualError(t, got.Error(), tt.want.Error().Error())
			} else {
				assert.Equal(t, tt.want, tt.group.First())
			}
		})
	}
}

func TestGroup_Signals(t *testing.T) {
	tests := []struct {
		name            string
		group           *Group
		want            []*Signal
		wantErrorString string
	}{
		{
			name:            "empty group",
			group:           NewGroup(),
			want:            []*Signal{},
			wantErrorString: "",
		},
		{
			name:            "with signals",
			group:           NewGroup(1, nil, 3),
			want:            []*Signal{New(1), New(nil), New(3)},
			wantErrorString: "",
		},
		{
			name:            "with error in chain",
			group:           NewGroup(1, 2, 3).WithError(errors.New("some error in chain")),
			want:            nil,
			wantErrorString: "some error in chain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.group.Signals()
			if tt.wantErrorString != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErrorString)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGroup_SignalsOrDefault(t *testing.T) {
	type args struct {
		defaultSignals []*Signal
	}
	tests := []struct {
		name  string
		group *Group
		args  args
		want  []*Signal
	}{
		{
			name:  "empty group",
			group: NewGroup(),
			args: args{
				defaultSignals: nil,
			},
			want: []*Signal{}, // Empty group has empty slice of signals
		},
		{
			name:  "with signals",
			group: NewGroup(1, 2, 3),
			args: args{
				defaultSignals: []*Signal{New(4), New(5)}, //Default must be ignored
			},
			want: []*Signal{New(1), New(2), New(3)},
		},
		{
			name:  "with error in chain and nil default",
			group: NewGroup(1, 2, 3).WithError(errors.New("some error in chain")),
			args: args{
				defaultSignals: nil,
			},
			want: nil,
		},
		{
			name:  "with error in chain and default",
			group: NewGroup(1, 2, 3).WithError(errors.New("some error in chain")),
			args: args{
				defaultSignals: []*Signal{New(4), New(5)},
			},
			want: []*Signal{New(4), New(5)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.group.SignalsOrDefault(tt.args.defaultSignals))
		})
	}
}
