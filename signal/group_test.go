package signal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGroup(t *testing.T) {
	t.Parallel()
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
				signals := group.All()
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
				signals := group.All()
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
	t.Parallel()
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
	t.Parallel()
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
			assert.Equal(t, tt.want, tt.group.AllPayloads())
		})
	}
}

func TestGroup_With(t *testing.T) {
	t.Parallel()
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
				signals: NewGroup(3, 4, 5).All(),
			},
			want: NewGroup(3, 4, 5),
		},
		{
			name:  "addition to group",
			group: NewGroup(1, 2, 3),
			args: args{
				signals: NewGroup(4, 5, 6).All(),
			},
			want: NewGroup(1, 2, 3, 4, 5, 6),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.With(tt.args.signals...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_WithPayloads(t *testing.T) {
	t.Parallel()
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

func TestGroup_Join(t *testing.T) {
	t.Parallel()
	t.Run("join two non-empty groups", func(t *testing.T) {
		a := NewGroup(1, 2)
		b := NewGroup(3, 4)
		got := a.Join(b)
		assert.Equal(t, 4, got.Len())
		assert.Equal(t, NewGroup(1, 2, 3, 4), got)
	})

	t.Run("join with empty group", func(t *testing.T) {
		a := NewGroup(1, 2)
		got := a.Join(NewGroup())
		assert.Equal(t, NewGroup(1, 2), got)
	})

	t.Run("join empty with non-empty", func(t *testing.T) {
		got := NewGroup().Join(NewGroup(5, 6))
		assert.Equal(t, NewGroup(5, 6), got)
	})

	t.Run("receiver unchanged after join", func(t *testing.T) {
		a := NewGroup(1, 2)
		_ = a.Join(NewGroup(3, 4))
		assert.Equal(t, 2, a.Len())
	})

	t.Run("join with nil group is treated as empty", func(t *testing.T) {
		a := NewGroup(1, 2)
		got := a.Join(nil)
		assert.Equal(t, NewGroup(1, 2), got)
	})
}

func TestGroup_Contains(t *testing.T) {
	t.Parallel()
	t.Run("found by pointer identity", func(t *testing.T) {
		s := New(42)
		g := NewGroup().With(s)
		assert.True(t, g.Contains(s))
	})

	t.Run("not found — different pointer same value", func(t *testing.T) {
		g := NewGroup(42)
		assert.False(t, g.Contains(New(42)))
	})

	t.Run("empty group", func(t *testing.T) {
		assert.False(t, NewGroup().Contains(New(1)))
	})
}

func TestGroup_ContainsPayload(t *testing.T) {
	t.Parallel()
	t.Run("found", func(t *testing.T) {
		assert.True(t, NewGroup(1, 2, 3).ContainsPayload(2))
	})

	t.Run("not found", func(t *testing.T) {
		assert.False(t, NewGroup(1, 2, 3).ContainsPayload(99))
	})

	t.Run("nil payload found", func(t *testing.T) {
		assert.True(t, NewGroup(nil, 1).ContainsPayload[any](nil))
	})

	t.Run("empty group", func(t *testing.T) {
		assert.False(t, NewGroup().ContainsPayload(1))
	})

	t.Run("a different numeric type is a different payload", func(t *testing.T) {
		// Interface equality, not conversion: int64(2) is not the int 2.
		assert.False(t, NewGroup(1, 2, 3).ContainsPayload(int64(2)))
	})
}

func TestGroup_ContainsPayloadFunc(t *testing.T) {
	t.Parallel()
	t.Run("found with custom comparator", func(t *testing.T) {
		g := NewGroup([]int{1, 2}, []int{3, 4})
		found := g.ContainsPayloadFunc(func(p any) bool {
			s, ok := p.([]int)
			return ok && len(s) == 2 && s[0] == 3
		})
		assert.True(t, found)
	})

	t.Run("not found", func(t *testing.T) {
		g := NewGroup(1, 2, 3)
		assert.False(t, g.ContainsPayloadFunc(func(p any) bool {
			v, ok := p.(int)
			return ok && v > 100
		}))
	})

	t.Run("empty group", func(t *testing.T) {
		assert.False(t, NewGroup().ContainsPayloadFunc(func(any) bool { return true }))
	})
}

func TestGroup_Filter(t *testing.T) {
	t.Parallel()
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
					return signal.Payload().(int) <= 2
				},
			},
			want: NewGroup(1, 2),
		},
		{
			name:  "all dropped",
			group: NewGroup(1, 2, 3, 4),
			args: args{
				predicate: func(signal *Signal) bool {
					return signal.Payload().(int) > 10
				},
			},
			want: NewGroup(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.group.Filter(tt.args.predicate))
		})
	}
}

func TestGroup_Map(t *testing.T) {
	t.Parallel()
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.Map(tt.args.mapperFunc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_MapIf(t *testing.T) {
	t.Parallel()
	type args struct {
		predicate  Predicate
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
				predicate: func(s *Signal) bool { return true },
				mapperFunc: func(s *Signal) *Signal {
					return s.MapPayload(func(p any) any { return p.(int) * 2 })
				},
			},
			want: NewGroup(),
		},
		{
			name:  "predicate matches all - all mapped",
			group: NewGroup(1, 2, 3),
			args: args{
				predicate: func(s *Signal) bool { return true },
				mapperFunc: func(s *Signal) *Signal {
					return s.MapPayload(func(p any) any { return p.(int) * 10 })
				},
			},
			want: NewGroup(10, 20, 30),
		},
		{
			name:  "predicate matches none - nothing mapped",
			group: NewGroup(1, 2, 3),
			args: args{
				predicate:  func(s *Signal) bool { return false },
				mapperFunc: func(s *Signal) *Signal { return s.MapPayload(func(p any) any { return -1 }) },
			},
			want: NewGroup(1, 2, 3),
		},
		{
			name:  "predicate matches some - only matching signals mapped",
			group: NewGroup(1, 2, 3, 4),
			args: args{
				predicate: func(s *Signal) bool {
					payload := s.Payload()
					return payload.(int)%2 == 0
				},
				mapperFunc: func(s *Signal) *Signal {
					return s.MapPayload(func(p any) any { return p.(int) * 100 })
				},
			},
			want: NewGroup(1, 200, 3, 400),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.MapIf(tt.args.predicate, tt.args.mapperFunc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_MapPayloadsIf(t *testing.T) {
	t.Parallel()
	type args struct {
		predicate  Predicate
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
				predicate:  func(s *Signal) bool { return true },
				mapperFunc: func(p any) any { return p.(int) * 2 },
			},
			want: NewGroup(),
		},
		{
			name:  "predicate matches all - all payloads mapped",
			group: NewGroup(1, 2, 3),
			args: args{
				predicate:  func(s *Signal) bool { return true },
				mapperFunc: func(p any) any { return p.(int) * 10 },
			},
			want: NewGroup(10, 20, 30),
		},
		{
			name:  "predicate matches none - no payloads changed",
			group: NewGroup(1, 2, 3),
			args: args{
				predicate:  func(s *Signal) bool { return false },
				mapperFunc: func(p any) any { return -1 },
			},
			want: NewGroup(1, 2, 3),
		},
		{
			name:  "predicate matches some - only matching payloads mapped",
			group: NewGroup(1, 2, 3, 4),
			args: args{
				predicate: func(s *Signal) bool {
					payload := s.Payload()
					return payload.(int)%2 == 0
				},
				mapperFunc: func(p any) any { return p.(int) * 100 },
			},
			want: NewGroup(1, 200, 3, 400),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.MapPayloadsIf(tt.args.predicate, tt.args.mapperFunc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_MapPayloads(t *testing.T) {
	t.Parallel()
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.MapPayloads(tt.args.mapperFunc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_Reduce(t *testing.T) {
	t.Parallel()
	t.Run("accumulates signals", func(t *testing.T) {
		g := NewGroup(1, 2, 3)
		result := g.Reduce(New(0), func(acc, s *Signal) *Signal {
			accVal := acc.Payload()
			sVal := s.Payload()
			return New(accVal.(int) + sVal.(int))
		})
		require.NotNil(t, result)
		payload := result.Payload()
		assert.Equal(t, 6, payload)
	})

	t.Run("returns initial for empty group", func(t *testing.T) {
		initial := New(99)
		result := NewGroup().Reduce(initial, func(acc, s *Signal) *Signal { return s })
		assert.Equal(t, initial, result)
	})
}

func TestGroup_ReducePayloads(t *testing.T) {
	t.Parallel()
	t.Run("sums integers", func(t *testing.T) {
		g := NewGroup(1, 2, 3, 4)
		result := g.ReducePayloads(0, func(acc int, payload any) int {
			return acc + payload.(int)
		})
		assert.Equal(t, 10, result)
	})

	t.Run("concatenates strings", func(t *testing.T) {
		g := NewGroup("a", "b", "c")
		result := g.ReducePayloads("", func(acc string, payload any) string {
			return acc + payload.(string)
		})
		assert.Equal(t, "abc", result)
	})

	t.Run("returns initial for empty group", func(t *testing.T) {
		result := NewGroup().ReducePayloads(42, func(_ int, payload any) int { return payload.(int) })
		assert.Equal(t, 42, result)
	})

	t.Run("the accumulator type need not be a payload type", func(t *testing.T) {
		// Typed accumulation is the point: no cast on acc, and A is whatever the
		// caller wants to build.
		g := NewGroup(1, "two", 3.0)
		kinds := g.ReducePayloads([]string{}, func(acc []string, payload any) []string {
			return append(acc, fmt.Sprintf("%T", payload))
		})
		assert.Equal(t, []string{"int", "string", "float64"}, kinds)
	})
}

// TestGroup_NilPayloadInvariant verifies that nil is a valid payload in a group
// and survives group operations unchanged.
func TestGroup_NilPayloadInvariant(t *testing.T) {
	t.Parallel()
	t.Run("First returns nil-payload signal", func(t *testing.T) {
		got := NewGroup(nil, 1).First().Payload()
		assert.Nil(t, got)
	})

	t.Run("Last returns nil-payload signal", func(t *testing.T) {
		got := NewGroup(1, nil).Last().Payload()
		assert.Nil(t, got)
	})

	t.Run("Filter preserves nil-payload signals", func(t *testing.T) {
		filtered := NewGroup(nil, 1, nil).Filter(func(s *Signal) bool {
			return s.Payload() == nil
		})
		assert.Equal(t, 2, filtered.Len())
		got := filtered.First().Payload()
		assert.Nil(t, got)
	})

	t.Run("Map preserves nil-payload signals", func(t *testing.T) {
		got := NewGroup(nil).Map(func(s *Signal) *Signal {
			return s.WithMeta("touched", "yes")
		}).First().Payload()
		assert.Nil(t, got)
	})

	t.Run("Join preserves nil-payload signals", func(t *testing.T) {
		joined := NewGroup(nil).Join(NewGroup(nil))
		assert.Equal(t, 2, joined.Len())

		first := joined.First().Payload()
		assert.Nil(t, first)

		last := joined.Last().Payload()
		assert.Nil(t, last)
	})

	t.Run("AllPayloads includes nil entries", func(t *testing.T) {
		payloads := NewGroup(1, nil, 2).AllPayloads()
		assert.Equal(t, []any{1, nil, 2}, payloads)
	})

	t.Run("ContainsPayload finds nil", func(t *testing.T) {
		assert.True(t, NewGroup(nil, 1).ContainsPayload[any](nil))
		assert.False(t, NewGroup(1, 2).ContainsPayload[any](nil))
	})
}

func TestGroup_MapDropsNilResults(t *testing.T) {
	t.Parallel()
	dropOdd := func(s *Signal) *Signal {
		if s.Payload().(int)%2 != 0 {
			return nil
		}
		return s
	}

	t.Run("Map drops nil mapper results", func(t *testing.T) {
		g := NewGroup(1, 2, 3, 4).Map(dropOdd)
		assert.Equal(t, 2, g.Len())
		require.NoError(t, g.ForEach(func(s *Signal) error {
			assert.NotNil(t, s.Payload())
			return nil
		}))
	})

	t.Run("MapIf drops nil mapper results and keeps non-matching signals", func(t *testing.T) {
		isOdd := func(s *Signal) bool {
			return s.Payload().(int)%2 != 0
		}
		g := NewGroup(1, 2, 3, 4).MapIf(isOdd, func(*Signal) *Signal {
			return nil
		})
		assert.Equal(t, 2, g.Len())

		payloads := g.AllPayloads()
		assert.Equal(t, []any{2, 4}, payloads)
	})
}

// TestGroup_PromotedReadSurface smoke-checks the read methods promoted from
// internal/collection.Slice; that package's suite is their source of truth.
func TestGroup_PromotedReadSurface(t *testing.T) {
	t.Parallel()
	t.Run("promoted methods work through the Group facade", func(t *testing.T) {
		g := NewGroup(1, 2, 3)
		assert.Equal(t, 3, g.Len())
		assert.Equal(t, 1, g.First().Payload())
		assert.Equal(t, 3, g.Last().Payload())
		assert.Nil(t, NewGroup().First(), "empty group yields nil, not a zero Signal")
	})
}

func TestGroup_FirstAs(t *testing.T) {
	t.Parallel()
	t.Run("returns the first payload as T", func(t *testing.T) {
		got, err := NewGroup(7, "later").FirstAs[int]()
		require.NoError(t, err)
		assert.Equal(t, 7, got)
	})

	t.Run("an empty group is ErrNoSignalsInGroup", func(t *testing.T) {
		_, err := NewGroup().FirstAs[int]()
		require.ErrorIs(t, err, ErrNoSignalsInGroup)
	})

	t.Run("a wrong type is an error, not a panic", func(t *testing.T) {
		_, err := NewGroup("seven").FirstAs[int]()
		require.ErrorContains(t, err, "is string, not int")
	})

	t.Run("a nil payload is a wrong type", func(t *testing.T) {
		_, err := NewGroup(nil).FirstAs[int]()
		require.Error(t, err)
	})
}

func TestGroup_FirstPayloadOrDefault(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 7, NewGroup(7, 8).FirstPayloadOrDefault(0))
	assert.Equal(t, 9, NewGroup().FirstPayloadOrDefault(9), "empty group yields the default")
	assert.Equal(t, 9, NewGroup("seven").FirstPayloadOrDefault(9), "wrong type yields the default")
	assert.Equal(t, 9, NewGroup(nil).FirstPayloadOrDefault(9), "nil payload yields the default")
}
