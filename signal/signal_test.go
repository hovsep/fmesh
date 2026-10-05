package signal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	type args struct {
		payload any
	}
	tests := []struct {
		name string
		args args
		want *Signal
	}{
		{
			name: "nil payload is valid",
			args: args{
				payload: nil,
			},
			want: &Signal{
				payload: nil,
			},
		},
		{
			name: "with payload",
			args: args{
				payload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
			},
			want: &Signal{
				payload: []any{123, "hello", []int{1, 2, 3}, map[string]int{"key": 42}, []byte{}, nil},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.args.payload))
		})
	}
}

func TestSignal_Payload(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		signal *Signal
		want   any
	}{
		{
			// Payload no longer reports this: a signal without one can only come
			// from building the zero value instead of calling New, and it reads
			// as nil like any other absent value.
			name:   "zero-value signal reads as nil",
			signal: &Signal{},
			want:   nil,
		},
		{
			name:   "nil payload is valid",
			signal: New(nil),
			want:   nil,
		},
		{
			name:   "with payload",
			signal: New(123),
			want:   123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.signal.Payload())
		})
	}
}

func TestSignal_Map(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		signal     *Signal
		mapperFunc Mapper
		want       *Signal
	}{
		{
			name:   "happy path",
			signal: New(1),
			mapperFunc: func(signal *Signal) *Signal {
				return signal.WithMeta("l1", "v1")
			},
			want: New(1).WithMeta("l1", "v1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.signal.Map(tt.mapperFunc))
		})
	}
}

func TestSignal_MapPayload(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		signal     *Signal
		mapperFunc PayloadMapper
		want       *Signal
	}{
		{
			name:   "happy path",
			signal: New(1).WithMeta("foo", "bar"),
			mapperFunc: func(payload any) any {
				return payload.(int) * 2
			},
			want: New(2).WithMeta("foo", "bar"),
		},
		{
			name:   "payload nil",
			signal: New(nil).WithMeta("x", "y"),
			mapperFunc: func(payload any) any {
				if payload == nil {
					return "default"
				}
				return payload
			},
			want: New("default").WithMeta("x", "y"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.signal.MapPayload(tt.mapperFunc)
			assert.Equal(t, tt.want.Payload(), got.Payload())
			assert.Equal(t, tt.want.Meta(), got.Meta())
		})
	}
}

func TestSignal_WithMetaMany(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		signal     *Signal
		entries    map[string]string
		assertions func(t *testing.T, signal *Signal)
	}{
		{
			name:   "add entries to new signal",
			signal: New(123),
			entries: map[string]string{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 2, signal.Meta().Len())
				assert.True(t, signal.meta.Has("l1", "l2"))
			},
		},
		{
			name:   "add entries merges with existing",
			signal: New(123).WithMetaMany(map[string]string{"existing": "entry"}),
			entries: map[string]string{
				"l1": "v1",
				"l2": "v2",
			},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 3, signal.Meta().Len())
				assert.True(t, signal.meta.Has("existing", "l1", "l2"))
			},
		},
		{
			name:   "add entries updates existing key",
			signal: New(123).WithMetaMany(map[string]string{"l1": "old"}),
			entries: map[string]string{
				"l1": "new",
			},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Meta().Len())
				assert.True(t, signal.meta.ValueIs("l1", "new"))
			},
		},
		{
			name:    "nil map is a no-op",
			signal:  New(123).WithMeta("l1", "v1"),
			entries: nil,
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Meta().Len())
				assert.True(t, signal.meta.ValueIs("l1", "v1"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signalAfter := tt.signal.WithMetaMany(tt.entries)
			if tt.assertions != nil {
				tt.assertions(t, signalAfter)
			}
		})
	}
}

func TestSignal_WithMeta(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		signal     *Signal
		key        string
		value      string
		assertions func(t *testing.T, signal *Signal)
	}{
		{
			name:   "add single entry to new signal",
			signal: New(123),
			key:    "priority",
			value:  "high",
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Meta().Len())
				assert.True(t, signal.meta.ValueIs("priority", "high"))
			},
		},
		{
			name:   "add entry merges with existing",
			signal: New(123).WithMeta("existing", "entry"),
			key:    "priority",
			value:  "high",
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 2, signal.Meta().Len())
				assert.True(t, signal.meta.Has("existing", "priority"))
			},
		},
		{
			name:   "add entry updates existing key",
			signal: New(123).WithMeta("priority", "low"),
			key:    "priority",
			value:  "high",
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Meta().Len())
				assert.True(t, signal.meta.ValueIs("priority", "high"))
			},
		},
		{
			name:   "chainable",
			signal: New(123),
			key:    "l1",
			value:  "v1",
			assertions: func(t *testing.T, signal *Signal) {
				result := signal.WithMeta("l2", "v2").WithMeta("l3", "v3")
				assert.Equal(t, 3, result.Meta().Len())
				assert.True(t, result.meta.Has("l1", "l2", "l3"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signalAfter := tt.signal.WithMeta(tt.key, tt.value)
			if tt.assertions != nil {
				tt.assertions(t, signalAfter)
			}
		})
	}
}

func TestSignal_WithoutMeta(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		signal       *Signal
		keysToRemove []string
		assertions   func(t *testing.T, signal *Signal)
	}{
		{
			name:         "remove single entry",
			signal:       New(123).WithMetaMany(map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
			keysToRemove: []string{"k1"},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 2, signal.Meta().Len())
				assert.False(t, signal.Meta().Has("k1"))
				assert.True(t, signal.Meta().Has("k2"))
				assert.True(t, signal.Meta().Has("k3"))
			},
		},
		{
			name:         "remove multiple entries",
			signal:       New(123).WithMetaMany(map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
			keysToRemove: []string{"k1", "k2"},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Meta().Len())
				assert.False(t, signal.Meta().Has("k1"))
				assert.False(t, signal.Meta().Has("k2"))
				assert.True(t, signal.Meta().ValueIs("k3", "v3"))
			},
		},
		{
			name:         "remove non-existent entry",
			signal:       New(123).WithMetaMany(map[string]string{"k1": "v1"}),
			keysToRemove: []string{"k2"},
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Meta().Len())
				assert.True(t, signal.Meta().ValueIs("k1", "v1"))
			},
		},
		{
			name:         "chainable",
			signal:       New(123).WithMetaMany(map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}),
			keysToRemove: []string{"k1"},
			assertions: func(t *testing.T, signal *Signal) {
				result := signal.WithoutMeta("k2").WithMeta("k4", "v4")
				assert.Equal(t, 2, result.Meta().Len())
				assert.False(t, result.Meta().Has("k1"))
				assert.False(t, result.Meta().Has("k2"))
				assert.True(t, result.Meta().ValueIs("k3", "v3"))
				assert.True(t, result.Meta().ValueIs("k4", "v4"))
			},
		},
		{
			name:         "no keys is a no-op",
			signal:       New(123).WithMetaMany(map[string]string{"k1": "v1"}),
			keysToRemove: nil,
			assertions: func(t *testing.T, signal *Signal) {
				assert.Equal(t, 1, signal.Meta().Len())
				assert.True(t, signal.Meta().ValueIs("k1", "v1"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signalAfter := tt.signal.WithoutMeta(tt.keysToRemove...)
			if tt.assertions != nil {
				tt.assertions(t, signalAfter)
			}
		})
	}
}

func TestSignal_Chainability(t *testing.T) {
	t.Parallel()
	t.Run("WithMetaMany called twice merges entries", func(t *testing.T) {
		s := New(123).
			WithMetaMany(map[string]string{"k1": "v1", "k2": "v2"}).
			WithMetaMany(map[string]string{"k3": "v3", "k2": "v2-updated"})

		assert.Equal(t, 3, s.Meta().Len())
		assert.True(t, s.Meta().ValueIs("k1", "v1"))
		assert.True(t, s.Meta().ValueIs("k2", "v2-updated"), "should update existing key")
		assert.True(t, s.Meta().ValueIs("k3", "v3"))
	})

	t.Run("WithoutMeta removes specific entries", func(t *testing.T) {
		s := New(123).
			WithMetaMany(map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}).
			WithoutMeta("k1", "k2").
			WithMeta("k4", "v4")

		assert.Equal(t, 2, s.Meta().Len())
		assert.False(t, s.Meta().Has("k1"))
		assert.False(t, s.Meta().Has("k2"))
		assert.True(t, s.Meta().ValueIs("k3", "v3"))
		assert.True(t, s.Meta().ValueIs("k4", "v4"))
	})
}

// TestSignal_NilPayloadInvariant verifies that nil is a valid payload and survives
// all mutation operations (copy-on-write metadata changes) unchanged.
func TestSignal_NilPayloadInvariant(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		signal *Signal
	}{
		{
			name:   "New(nil)",
			signal: New(nil),
		},
		{
			name:   "after WithMeta",
			signal: New(nil).WithMeta("k", "v"),
		},
		{
			name:   "after WithMetaMany",
			signal: New(nil).WithMetaMany(map[string]string{"k": "v"}),
		},
		{
			name:   "after WithoutMeta",
			signal: New(nil).WithMeta("k", "v").WithoutMeta("k"),
		},
		{
			name:   "after MapPayload identity",
			signal: New(nil).MapPayload(func(p any) any { return p }),
		},
		{
			name:   "after Map",
			signal: New(nil).Map(func(s *Signal) *Signal { return s.WithMeta("x", "y") }),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := tt.signal.Payload()
			assert.Nil(t, payload)
		})
	}
}

func TestSignal_ZeroValueReadsAsNil(t *testing.T) {
	t.Parallel()
	// A signal without a payload can only be built by skipping New. Payload does
	// not report that as an error — it is a construction bug, not a runtime
	// condition — so it reads as nil, and the typed accessors say what is wrong.
	var s Signal

	assert.Nil(t, s.Payload())

	_, err := s.As[string]()
	require.Error(t, err, "As must still report that the payload is not a string")
	assert.Equal(t, "fallback", s.PayloadOrDefault("fallback"))
}
