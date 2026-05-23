package signal

import (
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.With(tt.args.signals...)
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
}

func TestGroup_Last(t *testing.T) {
	t.Run("empty group returns nil", func(t *testing.T) {
		assert.Nil(t, NewGroup().Last())
	})

	t.Run("single element", func(t *testing.T) {
		group := NewGroup(42)
		got := group.Last()
		require.NotNil(t, got)
		payload, err := got.Payload()
		require.NoError(t, err)
		assert.Equal(t, 42, payload)
	})

	t.Run("returns last element", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		got := group.Last()
		require.NotNil(t, got)
		payload, err := got.Payload()
		require.NoError(t, err)
		assert.Equal(t, 3, payload)
	})
}

func TestGroup_Join(t *testing.T) {
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
}

func TestGroup_Contains(t *testing.T) {
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
	t.Run("found", func(t *testing.T) {
		g := NewGroup(1, 2, 3)
		found, err := g.ContainsPayload(2)
		require.NoError(t, err)
		assert.True(t, found)
	})

	t.Run("not found", func(t *testing.T) {
		g := NewGroup(1, 2, 3)
		found, err := g.ContainsPayload(99)
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("nil payload found", func(t *testing.T) {
		g := NewGroup(nil, 1)
		found, err := g.ContainsPayload(nil)
		require.NoError(t, err)
		assert.True(t, found)
	})

	t.Run("empty group", func(t *testing.T) {
		found, err := NewGroup().ContainsPayload(1)
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("non-comparable payload returns error", func(t *testing.T) {
		_, err := NewGroup(1).ContainsPayload([]int{1, 2})
		assert.Error(t, err)
	})
}

func TestGroup_ContainsPayloadFunc(t *testing.T) {
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

func TestGroup_Find(t *testing.T) {
	t.Run("returns first matching signal", func(t *testing.T) {
		group := NewGroup(1, 2, 3, 4)
		got := group.Find(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0
		})
		require.NotNil(t, got)
		payload, err := got.Payload()
		require.NoError(t, err)
		assert.Equal(t, 2, payload)
	})

	t.Run("returns nil when no signal matches", func(t *testing.T) {
		group := NewGroup(1, 3, 5)
		got := group.Find(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0
		})
		assert.Nil(t, got)
	})

	t.Run("returns nil for empty group", func(t *testing.T) {
		group := NewGroup()
		got := group.Find(func(s *Signal) bool { return true })
		assert.Nil(t, got)
	})
}

func TestGroup_All(t *testing.T) {
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
			name: "with labeled signals",
			group: NewGroup(1, nil, 3).Map(func(s *Signal) *Signal {
				return s.WithOnlyLabels(labels.Map{"flavor": "banana"})
			}),
			want: Signals{
				New(1).WithOnlyLabels(labels.Map{"flavor": "banana"}),
				New(nil).WithOnlyLabels(labels.Map{"flavor": "banana"}),
				New(3).WithOnlyLabels(labels.Map{"flavor": "banana"}),
			},
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
			assert.Equal(t, tt.want, tt.group.Filter(tt.args.predicate))
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.Map(tt.args.mapperFunc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroup_MapIf(t *testing.T) {
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
					payload, _ := s.Payload()
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
					payload, _ := s.Payload()
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

func TestGroup_Every(t *testing.T) {
	t.Run("returns true when all match", func(t *testing.T) {
		group := NewGroup(2, 4, 6)
		result := group.Every(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0
		})
		assert.True(t, result)
	})

	t.Run("returns false when not all match", func(t *testing.T) {
		group := NewGroup(2, 3, 4)
		result := group.Every(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0
		})
		assert.False(t, result)
	})

	t.Run("returns true for empty group (vacuous truth)", func(t *testing.T) {
		group := NewGroup()
		result := group.Every(func(s *Signal) bool {
			return true
		})
		assert.True(t, result)
	})
}

func TestGroup_Any(t *testing.T) {
	t.Run("returns true when at least one matches", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		result := group.Any(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int) == 2
		})
		assert.True(t, result)
	})

	t.Run("returns false when none match", func(t *testing.T) {
		group := NewGroup(1, 3, 5)
		result := group.Any(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0
		})
		assert.False(t, result)
	})

	t.Run("returns false for empty group", func(t *testing.T) {
		group := NewGroup()
		result := group.Any(func(s *Signal) bool {
			return true
		})
		assert.False(t, result)
	})
}

func TestGroup_Count(t *testing.T) {
	t.Run("counts matching signals", func(t *testing.T) {
		group := NewGroup(1, 2, 3, 4, 5)
		count := group.Count(func(s *Signal) bool {
			payload, _ := s.Payload()
			return payload.(int)%2 == 0
		})
		assert.Equal(t, 2, count)
	})

	t.Run("returns 0 for empty group", func(t *testing.T) {
		group := NewGroup()
		count := group.Count(func(s *Signal) bool {
			return true
		})
		assert.Equal(t, 0, count)
	})
}

func TestGroup_ForEach(t *testing.T) {
	t.Run("applies action to each signal", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		count := 0
		_, err := group.ForEach(func(s *Signal) error {
			count++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("stops on error", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		_, err := group.ForEach(func(s *Signal) error {
			return assert.AnError
		})
		assert.Error(t, err)
	})
}

func TestGroup_ForEachIf(t *testing.T) {
	t.Run("applies action only to matching signals", func(t *testing.T) {
		group := NewGroup(1, 2, 3, 4)
		count := 0
		_, err := group.ForEachIf(
			func(s *Signal) bool {
				payload, _ := s.Payload()
				return payload.(int)%2 == 0
			},
			func(s *Signal) error {
				count++
				return nil
			},
		)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("applies action to all when predicate always true", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		count := 0
		_, err := group.ForEachIf(
			func(s *Signal) bool { return true },
			func(s *Signal) error { count++; return nil },
		)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("applies action to none when predicate always false", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		count := 0
		_, err := group.ForEachIf(
			func(s *Signal) bool { return false },
			func(s *Signal) error { count++; return nil },
		)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("stops on error", func(t *testing.T) {
		group := NewGroup(2, 4, 6)
		_, err := group.ForEachIf(
			func(s *Signal) bool { return true },
			func(s *Signal) error { return assert.AnError },
		)
		assert.Error(t, err)
	})
}

func TestGroup_Len(t *testing.T) {
	t.Run("returns count of signals", func(t *testing.T) {
		group := NewGroup(1, 2, 3)
		assert.Equal(t, 3, group.Len())
	})

	t.Run("returns 0 for empty group", func(t *testing.T) {
		assert.Equal(t, 0, NewGroup().Len())
	})
}

func TestGroup_Reduce(t *testing.T) {
	t.Run("accumulates signals", func(t *testing.T) {
		g := NewGroup(1, 2, 3)
		result := g.Reduce(New(0), func(acc, s *Signal) *Signal {
			accVal, _ := acc.Payload()
			sVal, _ := s.Payload()
			return New(accVal.(int) + sVal.(int))
		})
		require.NotNil(t, result)
		payload, err := result.Payload()
		require.NoError(t, err)
		assert.Equal(t, 6, payload)
	})

	t.Run("returns initial for empty group", func(t *testing.T) {
		initial := New(99)
		result := NewGroup().Reduce(initial, func(acc, s *Signal) *Signal { return s })
		assert.Equal(t, initial, result)
	})
}

func TestGroup_ReducePayloads(t *testing.T) {
	t.Run("sums integers", func(t *testing.T) {
		g := NewGroup(1, 2, 3, 4)
		result := g.ReducePayloads(0, func(acc, payload any) any {
			return acc.(int) + payload.(int)
		})
		assert.Equal(t, 10, result)
	})

	t.Run("concatenates strings", func(t *testing.T) {
		g := NewGroup("a", "b", "c")
		result := g.ReducePayloads("", func(acc, payload any) any {
			return acc.(string) + payload.(string)
		})
		assert.Equal(t, "abc", result)
	})

	t.Run("returns initial for empty group", func(t *testing.T) {
		result := NewGroup().ReducePayloads(42, func(acc, payload any) any { return payload })
		assert.Equal(t, 42, result)
	})
}

// TestGroup_NilPayloadInvariant verifies that nil is a valid payload in a group
// and survives group operations unchanged.
func TestGroup_NilPayloadInvariant(t *testing.T) {
	t.Run("First returns nil-payload signal", func(t *testing.T) {
		got, err := NewGroup(nil, 1).First().Payload()
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Last returns nil-payload signal", func(t *testing.T) {
		got, err := NewGroup(1, nil).Last().Payload()
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Filter preserves nil-payload signals", func(t *testing.T) {
		filtered := NewGroup(nil, 1, nil).Filter(func(s *Signal) bool {
			return s.PayloadOrDefault("x") == nil
		})
		assert.Equal(t, 2, filtered.Len())
		got, err := filtered.First().Payload()
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Map preserves nil-payload signals", func(t *testing.T) {
		got, err := NewGroup(nil).Map(func(s *Signal) *Signal {
			return s.WithLabel("touched", "yes")
		}).First().Payload()
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Join preserves nil-payload signals", func(t *testing.T) {
		joined := NewGroup(nil).Join(NewGroup(nil))
		assert.Equal(t, 2, joined.Len())

		first, err := joined.First().Payload()
		require.NoError(t, err)
		assert.Nil(t, first)

		last, err := joined.Last().Payload()
		require.NoError(t, err)
		assert.Nil(t, last)
	})

	t.Run("AllPayloads includes nil entries", func(t *testing.T) {
		payloads, err := NewGroup(1, nil, 2).AllPayloads()
		require.NoError(t, err)
		assert.Equal(t, []any{1, nil, 2}, payloads)
	})

	t.Run("ContainsPayload finds nil", func(t *testing.T) {
		found1, err1 := NewGroup(nil, 1).ContainsPayload(nil)
		require.NoError(t, err1)
		assert.True(t, found1)

		found2, err2 := NewGroup(1, 2).ContainsPayload(nil)
		require.NoError(t, err2)
		assert.False(t, found2)
	})
}
