package signal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAs(t *testing.T) {
	t.Run("returns the payload as the asked-for type", func(t *testing.T) {
		got, err := As[float64](New(3.5))
		require.NoError(t, err)
		assert.InDelta(t, 3.5, got, 1e-9)
	})

	t.Run("a payload of another type is an error, not a panic", func(t *testing.T) {
		// An unchecked assertion here takes down the mesh run when something
		// upstream changes what it publishes.
		got, err := As[float64](New("not a number"))

		require.Error(t, err)
		require.ErrorContains(t, err, "is string, not float64")
		assert.Zero(t, got)
	})

	t.Run("a nil signal is an error", func(t *testing.T) {
		_, err := As[float64](nil)
		require.ErrorContains(t, err, "signal is nil")
	})

	t.Run("a nil payload is reported as such", func(t *testing.T) {
		_, err := As[float64](New(nil))
		require.Error(t, err)
	})
}

func TestAsOrDefault(t *testing.T) {
	assert.InDelta(t, 3.5, AsOrDefault(New(3.5), 0.0), 1e-9)
	assert.InDelta(t, 7.0, AsOrDefault(New("wrong type"), 7.0), 1e-9,
		"the wrong type degrades to the default rather than failing the run")
	assert.InDelta(t, 7.0, AsOrDefault(nil, 7.0), 1e-9)

	assert.Equal(t, 7, AsOrDefault(New(7), 0), "T is inferred from the default")
	assert.Equal(t, "x", AsOrDefault(New("x"), ""))
	assert.True(t, AsOrDefault(New(true), false))
}

func TestAsOrDefaultInfersFromTheDefault(t *testing.T) {
	// An untyped 0 makes T int, so a float64 payload silently yields the
	// default. AsFloat64OrDefault exists for exactly this.
	assert.Equal(t, 0, AsOrDefault(New(2.5), 0))
	assert.InDelta(t, 2.5, AsFloat64OrDefault(New(2.5), 0), 1e-9)
}

func TestTypedShorthands(t *testing.T) {
	f, err := AsFloat64(New(1.5))
	require.NoError(t, err)
	assert.InDelta(t, 1.5, f, 1e-9)

	i, err := AsInt(New(4))
	require.NoError(t, err)
	assert.Equal(t, 4, i)

	s, err := AsString(New("hi"))
	require.NoError(t, err)
	assert.Equal(t, "hi", s)

	b, err := AsBool(New(true))
	require.NoError(t, err)
	assert.True(t, b)

	assert.InDelta(t, 9.0, AsFloat64OrDefault(New("x"), 9), 1e-9)
}

func TestAsGroup(t *testing.T) {
	inner := NewGroup(1, 2)

	got, err := AsGroup(New(inner))
	require.NoError(t, err)
	assert.Equal(t, 2, got.Len())

	_, err = AsGroup(New(1))
	require.Error(t, err)
}

func TestAsNumber(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		want    float64
		ok      bool
	}{
		{name: "float64", payload: 2.5, want: 2.5, ok: true},
		{name: "float32", payload: float32(2.5), want: 2.5, ok: true},
		{name: "int", payload: 3, want: 3, ok: true},
		{name: "int64", payload: int64(4), want: 4, ok: true},
		{name: "uint64", payload: uint64(5), want: 5, ok: true},
		{name: "true is one", payload: true, want: 1, ok: true},
		{name: "false is zero", payload: false, want: 0, ok: true},
		// A structured signal uses its payload as a type tag and keeps the real
		// values in scalars. Telling the two apart is what this is for.
		{name: "a type tag is not a measurement", payload: "venous_blood", ok: false},
		{name: "nil payload", payload: nil, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := AsNumber(New(tt.payload))
			assert.Equal(t, tt.ok, ok)
			assert.InDelta(t, tt.want, got, 1e-9)
		})
	}

	t.Run("nil signal", func(t *testing.T) {
		_, ok := AsNumber(nil)
		assert.False(t, ok)
	})
}
