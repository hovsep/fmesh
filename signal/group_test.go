package signal

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.NoError(t, err)
				assert.Empty(t, signals)
				assert.Zero(t, group.Len())
			},
		},
		{
			name: "with payloads",
			args: args{
				payloads: []any{1, nil, 3},
			},
			assertions: func(t *testing.T, group *Group) {
				signals, err := group.Signals()
				require.NoError(t, err)
				assert.Equal(t, 3, group.Len())
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
			group:           NewGroup(3, 4, 5).WithErr(errors.New("some error in chain")),
			want:            nil,
			wantErrorString: "some error in chain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.group.FirstPayload()
			if tt.wantErrorString != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErrorString)
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
			group:           NewGroup(1, 2, 3).WithErr(errors.New("some error in chain")),
			want:            nil,
			wantErrorString: "some error in chain",
		},
		{
			name:            "with error in signal",
			group:           NewGroup().withSignals(Signals{New(33).WithErr(errors.New("some error in signal"))}),
			want:            nil,
			wantErrorString: "some error in signal",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.group.AllPayloads()
			if tt.wantErrorString != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErrorString)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGroup_With(t *testing.T) {
	type args struct {
		signals Signals
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
			want: NewGroup().WithErr(errors.New("signal is invalid")),
		},
		{
			name:  "with error in signal",
			group: NewGroup(1, 2, 3).With(New(44).WithErr(errors.New("some error in signal"))),
			args: args{
				signals: Signals{New(456)},
			},
			want: NewGroup(1, 2, 3).
				With(New(44).WithErr(errors.New("some error in signal"))).
				WithErr(errors.New("some error in signal")), // error propagated from signal to group
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.With(tt.args.signals...)
			if tt.want.HasErr() {
				assert.Error(t, got.Err())
				assert.EqualError(t, got.Err(), tt.want.Err().Error())
			} else {
				assert.NoError(t, got.Err())
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
			want:  New(nil).WithErr(errors.New("group has no signals")),
		},
		{
			name:  "happy path",
			group: NewGroup(3, 5, 7),
			want:  New(3),
		},
		{
			name:  "with error in chain",
			group: NewGroup(1, 2, 3).WithErr(errors.New("some error in chain")),
			want:  New(nil).WithErr(errors.New("some error in chain")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.First()
			if tt.want.HasErr() {
				assert.True(t, got.HasErr())
				assert.Error(t, got.Err())
				assert.EqualError(t, got.Err(), tt.want.Err().Error())
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
		want            Signals
		wantErrorString string
	}{
		{
			name:            "empty group",
			group:           NewGroup(),
			want:            Signals{},
			wantErrorString: "",
		},
		{
			name:            "with signals",
			group:           NewGroup(1, nil, 3),
			want:            Signals{New(1), New(nil), New(3)},
			wantErrorString: "",
		},
		{
			name:            "with error in chain",
			group:           NewGroup(1, 2, 3).WithErr(errors.New("some error in chain")),
			want:            nil,
			wantErrorString: "some error in chain",
		},
		{
			name: "with labeled signals",
			group: NewGroup(1, nil, 3).WithSignalLabels(common.LabelsCollection{
				"flavor": "banana",
			}),
			want: Signals{
				New(1).WithLabels(common.LabelsCollection{
					"flavor": "banana",
				}),
				New(nil).WithLabels(common.LabelsCollection{
					"flavor": "banana",
				}),
				New(3).WithLabels(common.LabelsCollection{
					"flavor": "banana",
				})},
			wantErrorString: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.group.Signals()
			if tt.wantErrorString != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.wantErrorString)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGroup_SignalsOrDefault(t *testing.T) {
	type args struct {
		defaultSignals Signals
	}
	tests := []struct {
		name  string
		group *Group
		args  args
		want  Signals
	}{
		{
			name:  "empty group",
			group: NewGroup(),
			args: args{
				defaultSignals: nil,
			},
			want: Signals{}, // Empty group has empty slice of signals
		},
		{
			name:  "with signals",
			group: NewGroup(1, 2, 3),
			args: args{
				defaultSignals: Signals{New(4), New(5)}, // Default must be ignored
			},
			want: Signals{New(1), New(2), New(3)},
		},
		{
			name:  "with error in chain and nil default",
			group: NewGroup(1, 2, 3).WithErr(errors.New("some error in chain")),
			args: args{
				defaultSignals: nil,
			},
			want: nil,
		},
		{
			name:  "with error in chain and default",
			group: NewGroup(1, 2, 3).WithErr(errors.New("some error in chain")),
			args: args{
				defaultSignals: Signals{New(4), New(5)},
			},
			want: Signals{New(4), New(5)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.group.SignalsOrDefault(tt.args.defaultSignals))
		})
	}
}

func TestGroup_Filter(t *testing.T) {
	type args struct {
		filterFunc Filter
	}
	tests := []struct {
		name  string
		group *Group
		args  args
		want  *Group
	}{
		{
			name:  "empty group",
			group: NewGroup(),
			args: args{
				filterFunc: func(signal *Signal) bool {
					return true
				},
			},
			want: NewGroup(),
		},
		{
			name:  "nothing filtered out",
			group: NewGroup(1, 2, 3),
			args: args{
				filterFunc: func(signal *Signal) bool {
					return true
				},
			},
			want: NewGroup(1, 2, 3),
		},
		{
			name:  "some filtered out",
			group: NewGroup(1, 2, 3, 4),
			args: args{
				filterFunc: func(signal *Signal) bool {
					return signal.PayloadOrDefault(0).(int) <= 2
				},
			},
			want: NewGroup(1, 2),
		},
		{
			name:  "all dropped",
			group: NewGroup(1, 2, 3, 4),
			args: args{
				filterFunc: func(signal *Signal) bool {
					return signal.PayloadOrDefault(0).(int) > 10
				},
			},
			want: NewGroup(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.Filter(tt.args.filterFunc)
			if tt.want.HasErr() {
				assert.Error(t, got.Err())
				assert.EqualError(t, got.Err(), tt.want.Err().Error())
			} else {
				assert.NoError(t, got.Err())
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
