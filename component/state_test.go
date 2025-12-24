package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponent_WithInitialState(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		wantState State
	}{
		{
			name:      "no initial state",
			component: New("c1"),
			wantState: NewState(),
		},
		{
			name: "with initial state",
			component: New("c1").WithInitialState(func(state State) {
				state.Set("battery", 100.00)
				state.Set("speed", 200)
				state.Set("data", []byte{1, 2, 3})
				state.Set("secret", "LEON")
			}),
			wantState: State{
				"battery": 100.00,
				"speed":   200,
				"data":    []byte{1, 2, 3},
				"secret":  "LEON",
			},
		},
		{
			name:      "with nil state initializer",
			component: New("c1").WithInitialState(nil),
			wantState: NewState(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantState, tt.component.State())
		})
	}
}

func TestComponent_ResetState(t *testing.T) {
	tests := []struct {
		name       string
		initState  func(State)
		assertions func(t *testing.T, c *Component)
	}{
		{
			name: "empty state remains empty",
			initState: func(s State) {
				// do nothing
			},
			assertions: func(t *testing.T, c *Component) {
				c.ResetState()
				assert.Empty(t, c.State(), "state should remain empty after reset")
			},
		},
		{
			name: "state with keys is cleared",
			initState: func(s State) {
				s.Set("a", 1)
				s.Set("b", "text")
				s.Set("c", []int{1, 2, 3})
			},
			assertions: func(t *testing.T, c *Component) {
				c.ResetState()
				assert.Empty(t, c.State(), "state should be empty after reset")
			},
		},
		{
			name: "state can be reused after reset",
			initState: func(s State) {
				s.Set("key", "value")
			},
			assertions: func(t *testing.T, c *Component) {
				c.ResetState()
				assert.Empty(t, c.State(), "state should be empty after reset")
				c.State().Set("new", 123)
				assert.Equal(t, 123, c.State().Get("new"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New("comp").WithInitialState(tt.initState)
			if tt.assertions != nil {
				tt.assertions(t, c)
			}
		})
	}
}

func TestState_Has(t *testing.T) {
	tests := []struct {
		name      string
		initState func(State)
		key       string
		want      bool
	}{
		{
			name:      "key absent",
			initState: func(s State) {},
			key:       "name",
			want:      false,
		},
		{
			name: "key present",
			initState: func(s State) {
				s.Set("name", "Leon")
			},
			key:  "name",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState()
			tt.initState(s)
			assert.Equal(t, tt.want, s.Has(tt.key))
		})
	}
}

func TestState_Get(t *testing.T) {
	tests := []struct {
		name      string
		initState func(State)
		key       string
		want      any
	}{
		{
			name:      "key absent",
			initState: func(s State) {},
			key:       "name",
			want:      nil,
		},
		{
			name: "key present",
			initState: func(s State) {
				s.Set("name", "Leon")
			},
			key:  "name",
			want: "Leon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState()
			tt.initState(s)
			assert.Equal(t, tt.want, s.Get(tt.key))
		})
	}
}

func TestState_GetOrDefault(t *testing.T) {
	tests := []struct {
		name      string
		initState func(State)
		key       string
		defVal    any
		want      any
	}{
		{
			name:      "key absent",
			initState: func(s State) {},
			key:       "name",
			defVal:    "Vita",
			want:      "Vita",
		},
		{
			name: "key present",
			initState: func(s State) {
				s.Set("name", "Leon")
			},
			key:    "name",
			defVal: "Vita",
			want:   "Leon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState()
			tt.initState(s)
			assert.Equal(t, tt.want, s.GetOrDefault(tt.key, tt.defVal))
		})
	}
}

func TestState_Delete(t *testing.T) {
	tests := []struct {
		name       string
		initState  func(State)
		key        string
		wantExists bool
	}{
		{
			name:       "delete non-existent key",
			initState:  func(s State) {},
			key:        "missing",
			wantExists: false,
		},
		{
			name: "delete existing key",
			initState: func(s State) {
				s.Set("name", "Leon")
				s.Set("fruit", "banana")
			},
			key:        "name",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState()
			tt.initState(s)
			s.Delete(tt.key)
			assert.Equal(t, tt.wantExists, s.Has(tt.key))
		})
	}
}

func TestState_GetTyped(t *testing.T) {
	tests := []struct {
		name      string
		initState func(State)
		key       string
		want      any
		typ       any // type parameter for GetTyped
		wantPanic bool
	}{
		{
			name:      "existing int key",
			initState: func(s State) { s.Set("num", 42) },
			key:       "num",
			want:      42,
			typ:       int(0),
		},
		{
			name:      "existing string key",
			initState: func(s State) { s.Set("text", "hello") },
			key:       "text",
			want:      "hello",
			typ:       "",
		},
		{
			name:      "missing key",
			initState: func(s State) {},
			key:       "missing",
			wantPanic: true,
			typ:       int(0),
		},
		{
			name:      "wrong type",
			initState: func(s State) { s.Set("num", 42) },
			key:       "num",
			wantPanic: true,
			typ:       "", // attempt to get string from int key
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState()
			tt.initState(s)

			fn := func() {
				switch tt.typ.(type) {
				case int:
					_ = GetTyped[int](s, tt.key)
				case string:
					_ = GetTyped[string](s, tt.key)
				default:
					panic("unsupported type in test")
				}
			}

			if tt.wantPanic {
				assert.Panics(t, fn)
			} else {
				fn() // no panic expected
				switch tt.typ.(type) {
				case int:
					assert.Equal(t, tt.want, GetTyped[int](s, tt.key))
				case string:
					assert.Equal(t, tt.want, GetTyped[string](s, tt.key))
				}
			}
		})
	}
}

func TestState_SetIfAbsent(t *testing.T) {
	tests := []struct {
		name      string
		initState func(State)
		key       string
		value     any
		wantSet   bool
		wantVal   any
	}{
		{
			name:      "sets missing key",
			initState: func(s State) {},
			key:       "k",
			value:     100,
			wantSet:   true,
			wantVal:   100,
		},
		{
			name:      "does not overwrite existing",
			initState: func(s State) { s.Set("k", 50) },
			key:       "k",
			value:     100,
			wantSet:   false,
			wantVal:   50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState()
			tt.initState(s)
			got := s.SetIfAbsent(tt.key, tt.value)
			assert.Equal(t, tt.wantSet, got)
			assert.Equal(t, tt.wantVal, s.Get(tt.key))
		})
	}
}

func TestState_Upsert(t *testing.T) {
	tests := []struct {
		name      string
		initState func(State)
		key       string
		fn        func(any) any
		want      any
	}{
		{
			name:      "upsert missing key",
			initState: func(s State) {},
			key:       "counter",
			fn: func(old any) any {
				if old == nil {
					return 1
				}
				return old.(int) + 1
			},
			want: 1,
		},
		{
			name:      "upsert existing key",
			initState: func(s State) { s.Set("counter", 5) },
			key:       "counter",
			fn: func(old any) any {
				if old == nil {
					return 1
				}
				return old.(int) + 1
			},
			want: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState()
			tt.initState(s)
			s.Upsert(tt.key, tt.fn)
			assert.Equal(t, tt.want, s.Get(tt.key))
		})
	}
}

func TestState_Update(t *testing.T) {
	tests := []struct {
		name      string
		initState func(State)
		key       string
		fn        func(any) any
		want      any
		wantBool  bool
	}{
		{
			name:      "update existing key",
			initState: func(s State) { s.Set("val", 10) },
			key:       "val",
			fn:        func(old any) any { return old.(int) + 5 },
			want:      15,
			wantBool:  true,
		},
		{
			name:      "update missing key",
			initState: func(s State) {},
			key:       "missing",
			fn:        func(old any) any { return 1 },
			want:      nil,
			wantBool:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState()
			tt.initState(s)
			got := s.Update(tt.key, tt.fn)
			assert.Equal(t, tt.wantBool, got)
			if got {
				assert.Equal(t, tt.want, s.Get(tt.key))
			} else {
				assert.Nil(t, s.Get(tt.key))
			}
		})
	}
}
