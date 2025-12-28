package signal

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/labels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustAll is a test helper that panics if All returns an error.
func (g *Group) mustAll() Signals {
	signals, err := g.All()
	if err != nil {
		panic(err)
	}
	return signals
}

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
				signals, err := group.All()
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
				signals, err := group.All()
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
			group:           NewGroup(3, 4, 5).WithChainableErr(errors.New("some error in chain")),
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
			group:           NewGroup(1, 2, 3).WithChainableErr(errors.New("some error in chain")),
			want:            nil,
			wantErrorString: "some error in chain",
		},
		{
			name:            "with error in signal",
			group:           NewGroup().withSignals(Signals{New(33).WithChainableErr(errors.New("some error in signal"))}),
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
				signals: NewGroup(3, 4, 5).mustAll(),
			},
			want: NewGroup(3, 4, 5),
		},
		{
			name:  "addition to group",
			group: NewGroup(1, 2, 3),
			args: args{
				signals: NewGroup(4, 5, 6).mustAll(),
			},
			want: NewGroup(1, 2, 3, 4, 5, 6),
		},
		{
			name: "with error in chain",
			group: NewGroup(1, 2, 3).
				Add(New("valid before invalid")).
				Add(nil).
				AddFromPayloads(4, 5, 6),
			args: args{
				signals: NewGroup(7, nil, 9).mustAll(),
			},
			want: NewGroup().WithChainableErr(errors.New("signal is invalid")),
		},
		{
			name:  "with error in signal",
			group: NewGroup(1, 2, 3).Add(New(44).WithChainableErr(errors.New("some error in signal"))),
			args: args{
				signals: Signals{New(456)},
			},
			want: NewGroup(1, 2, 3).
				Add(New(44).WithChainableErr(errors.New("some error in signal"))).
				WithChainableErr(errors.New("some error in signal")), // error propagated from signal to group
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.Add(tt.args.signals...)
			if tt.want.HasChainableErr() {
				assert.Error(t, got.ChainableErr())
				assert.EqualError(t, got.ChainableErr(), tt.want.ChainableErr().Error())
			} else {
				assert.NoError(t, got.ChainableErr())
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
			assert.Equal(t, tt.want, tt.group.AddFromPayloads(tt.args.payloads...))
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
			want:  New(nil).WithChainableErr(errors.New("group has no signals")),
		},
		{
			name:  "happy path",
			group: NewGroup(3, 5, 7),
			want:  New(3),
		},
		{
			name:  "with error in chain",
			group: NewGroup(1, 2, 3).WithChainableErr(errors.New("some error in chain")),
			want:  New(nil).WithChainableErr(errors.New("some error in chain")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.First()
			if tt.want.HasChainableErr() {
				assert.True(t, got.HasChainableErr())
				assert.Error(t, got.ChainableErr())
				assert.EqualError(t, got.ChainableErr(), tt.want.ChainableErr().Error())
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
			group:           NewGroup(1, 2, 3).WithChainableErr(errors.New("some error in chain")),
			want:            nil,
			wantErrorString: "some error in chain",
		},
		{
			name: "with labeled signals",
			group: NewGroup(1, nil, 3).ForEach(func(s *Signal) error {
				return s.SetLabels(labels.Map{"flavor": "banana"}).ChainableErr()
			}),
			want: Signals{
				New(1).SetLabels(labels.Map{
					"flavor": "banana",
				}),
				New(nil).SetLabels(labels.Map{
					"flavor": "banana",
				}),
				New(3).SetLabels(labels.Map{
					"flavor": "banana",
				})},
			wantErrorString: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.group.All()
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
			group: NewGroup(1, 2, 3).WithChainableErr(errors.New("some error in chain")),
			args: args{
				defaultSignals: nil,
			},
			want: nil,
		},
		{
			name:  "with error in chain and default",
			group: NewGroup(1, 2, 3).WithChainableErr(errors.New("some error in chain")),
			args: args{
				defaultSignals: Signals{New(4), New(5)},
			},
			want: Signals{New(4), New(5)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.group.All()
			if err != nil {
				assert.Equal(t, tt.want, tt.args.defaultSignals)
			} else {
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestGroup_Filter(t *testing.T) {
	type args struct {
		predicate Predicate
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
				predicate: func(signal *Signal) bool {
					return true
				},
			},
			want: NewGroup(),
		},
		{
			name:  "nothing filtered out",
			group: NewGroup(1, 2, 3),
			args: args{
				predicate: func(signal *Signal) bool {
					return true
				},
			},
			want: NewGroup(1, 2, 3),
		},
		{
			name:  "some filtered out",
			group: NewGroup(1, 2, 3, 4),
			args: args{
				predicate: func(signal *Signal) bool {
					return signal.PayloadOrDefault(0).(int) <= 2
				},
			},
			want: NewGroup(1, 2),
		},
		{
			name:  "all dropped",
			group: NewGroup(1, 2, 3, 4),
			args: args{
				predicate: func(signal *Signal) bool {
					return signal.PayloadOrDefault(0).(int) > 10
				},
			},
			want: NewGroup(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.Filter(tt.args.predicate)
			if tt.want.HasChainableErr() {
				assert.Error(t, got.ChainableErr())
				assert.EqualError(t, got.ChainableErr(), tt.want.ChainableErr().Error())
			} else {
				assert.NoError(t, got.ChainableErr())
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_Map(t *testing.T) {
	type args struct {
		mapperFunc Mapper
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
				mapperFunc: func(signal *Signal) *Signal {
					return signal
				},
			},
			want: NewGroup(),
		},
		{
			name:  "happy path",
			group: NewGroup(1, 2, 3),
			args: args{
				mapperFunc: func(signal *Signal) *Signal {
					return signal.MapPayload(func(payload any) any {
						return payload.(int) * 7
					})
				},
			},
			want: NewGroup(7, 14, 21),
		},
		{
			name:  "signal with error",
			group: NewGroup(1, 2, 3).Add(New(4).WithChainableErr(errors.New("some error in chain"))),
			args: args{
				mapperFunc: func(signal *Signal) *Signal {
					return signal.MapPayload(func(payload any) any {
						return payload.(int) * 8
					})
				},
			},
			want: NewGroup().WithChainableErr(errors.New("some error in chain")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.Map(tt.args.mapperFunc)
			if tt.want.HasChainableErr() {
				assert.Error(t, got.ChainableErr())
				assert.EqualError(t, got.ChainableErr(), tt.want.ChainableErr().Error())
			} else {
				assert.NoError(t, got.ChainableErr())
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_MapPayloads(t *testing.T) {
	type args struct {
		mapperFunc PayloadMapper
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
				mapperFunc: func(payload any) any {
					return nil
				},
			},
			want: NewGroup(),
		},
		{
			name:  "happy path",
			group: NewGroup(1, 2, 3),
			args: args{
				mapperFunc: func(payload any) any {
					return payload.(int) * 7
				},
			},
			want: NewGroup(7, 14, 21),
		},
		{
			name:  "signal with error",
			group: NewGroup(1, 2, 3).Add(New(4).WithChainableErr(errors.New("some error in chain"))),
			args: args{
				mapperFunc: func(payload any) any {
					return payload.(int) * 7
				},
			},
			want: NewGroup().WithChainableErr(errors.New("some error in chain")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.MapPayloads(tt.args.mapperFunc)
			if tt.want.HasChainableErr() {
				assert.Error(t, got.ChainableErr())
				assert.EqualError(t, got.ChainableErr(), tt.want.ChainableErr().Error())
			} else {
				assert.NoError(t, got.ChainableErr())
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_FirstDoesNotPoisonGroup(t *testing.T) {
	t.Run("First does not poison group when empty", func(t *testing.T) {
		group := NewGroup()

		// Query first on empty group
		result := group.First()

		// Result should have error
		assert.True(t, result.HasChainableErr())
		require.ErrorIs(t, result.ChainableErr(), ErrNoSignalsInGroup)

		// But group should NOT be poisoned
		assert.False(t, group.HasChainableErr())

		// Group should still be usable for adding
		group = group.Add(New(42))
		assert.Equal(t, 1, group.Len())
		assert.False(t, group.HasChainableErr())

		// Now First should work
		first := group.First()
		assert.False(t, first.HasChainableErr())
		payload, err := first.Payload()
		require.NoError(t, err)
		assert.Equal(t, 42, payload)
	})
}
