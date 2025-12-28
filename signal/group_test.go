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
	t.Run("empty group returns nil", func(t *testing.T) {
		group := NewGroup()
		got := group.First()
		assert.Nil(t, got)
	})

	t.Run("happy path", func(t *testing.T) {
		group := NewGroup(3, 5, 7)
		got := group.First()
		require.NotNil(t, got)
		payload, err := got.Payload()
		require.NoError(t, err)
		assert.Equal(t, 3, payload)
	})

	t.Run("with error in chain returns nil", func(t *testing.T) {
		group := NewGroup(1, 2, 3).WithChainableErr(errors.New("some error in chain"))
		got := group.First()
		assert.Nil(t, got)
	})
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

func TestGroup_Without(t *testing.T) {
	t.Run("removes matching signals", func(t *testing.T) {
		group := NewGroup(1, 2, 3, 4, 5)
		result := group.Without(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0 // remove even numbers
		})
		assert.Equal(t, 3, result.Len()) // 1, 3, 5 remain
	})

	t.Run("returns same group when has error", func(t *testing.T) {
		group := NewGroup(1, 2, 3).WithChainableErr(assert.AnError)
		result := group.Without(func(s *Signal) bool {
			return true
		})
		assert.True(t, result.HasChainableErr())
	})
}

func TestGroup_AllMatch(t *testing.T) {
	t.Run("returns true when all match", func(t *testing.T) {
		group := NewGroup(2, 4, 6)
		result := group.AllMatch(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0
		})
		assert.True(t, result)
	})

	t.Run("returns false when not all match", func(t *testing.T) {
		group := NewGroup(2, 3, 4)
		result := group.AllMatch(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0
		})
		assert.False(t, result)
	})

	t.Run("returns false for empty group", func(t *testing.T) {
		group := NewGroup()
		result := group.AllMatch(func(s *Signal) bool {
			return true
		})
		assert.False(t, result)
	})

	t.Run("returns false when group has error", func(t *testing.T) {
		group := NewGroup(1, 2).WithChainableErr(assert.AnError)
		result := group.AllMatch(func(s *Signal) bool {
			return true
		})
		assert.False(t, result)
	})
}

func TestGroup_AnyMatch(t *testing.T) {
	t.Run("returns true when at least one matches", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		result := group.AnyMatch(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int) == 2
		})
		assert.True(t, result)
	})

	t.Run("returns false when none match", func(t *testing.T) {
		group := NewGroup(1, 3, 5)
		result := group.AnyMatch(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0
		})
		assert.False(t, result)
	})

	t.Run("returns false for empty group", func(t *testing.T) {
		group := NewGroup()
		result := group.AnyMatch(func(s *Signal) bool {
			return true
		})
		assert.False(t, result)
	})

	t.Run("returns false when group has error", func(t *testing.T) {
		group := NewGroup(1, 2).WithChainableErr(assert.AnError)
		result := group.AnyMatch(func(s *Signal) bool {
			return true
		})
		assert.False(t, result)
	})
}

func TestGroup_CountMatch(t *testing.T) {
	t.Run("counts matching signals", func(t *testing.T) {
		group := NewGroup(1, 2, 3, 4, 5)
		count := group.CountMatch(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0
		})
		assert.Equal(t, 2, count) // 2 and 4
	})

	t.Run("returns 0 for empty group", func(t *testing.T) {
		group := NewGroup()
		count := group.CountMatch(func(s *Signal) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})

	t.Run("returns 0 when group has error", func(t *testing.T) {
		group := NewGroup(1, 2, 3).WithChainableErr(assert.AnError)
		count := group.CountMatch(func(s *Signal) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})
}

func TestGroup_ForEach(t *testing.T) {
	t.Run("applies action to each signal", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		count := 0
		group.ForEach(func(s *Signal) error {
			count++
			return nil
		})
		assert.Equal(t, 3, count)
	})

	t.Run("stops on error and sets chainable error", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		result := group.ForEach(func(s *Signal) error {
			return assert.AnError
		})
		assert.True(t, result.HasChainableErr())
	})

	t.Run("skips when group has error", func(t *testing.T) {
		group := NewGroup(1, 2, 3).WithChainableErr(assert.AnError)
		visited := false
		group.ForEach(func(s *Signal) error {
			visited = true
			return nil
		})
		assert.False(t, visited)
	})
}

func TestGroup_Len(t *testing.T) {
	t.Run("returns count of signals", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		assert.Equal(t, 3, group.Len())
	})

	t.Run("returns 0 when group has error", func(t *testing.T) {
		group := NewGroup(1, 2, 3).WithChainableErr(assert.AnError)
		assert.Equal(t, 0, group.Len())
	})
}

func TestGroup_FirstDoesNotPoisonGroup(t *testing.T) {
	t.Run("First does not poison group when empty", func(t *testing.T) {
		group := NewGroup()

		// Query first on empty group
		result := group.First()

		// Result should be nil
		assert.Nil(t, result)

		// Group should NOT be poisoned
		assert.False(t, group.HasChainableErr())

		// Group should still be usable for adding
		group = group.Add(New(42))
		assert.Equal(t, 1, group.Len())
		assert.False(t, group.HasChainableErr())

		// Now First should work
		first := group.First()
		require.NotNil(t, first)
		payload, err := first.Payload()
		require.NoError(t, err)
		assert.Equal(t, 42, payload)
	})
}
