package signal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignal_As(t *testing.T) {
	t.Parallel()
	t.Run("returns the payload as the asked-for type", func(t *testing.T) {
		got, err := New(3.5).As[float64]()
		require.NoError(t, err)
		assert.InDelta(t, 3.5, got, 1e-9)
	})

	t.Run("a payload of another type is an error, not a panic", func(t *testing.T) {
		// An unchecked assertion here takes down the mesh run when something
		// upstream changes what it publishes.
		got, err := New("not a number").As[float64]()

		require.Error(t, err)
		require.ErrorContains(t, err, "is string, not float64")
		assert.Zero(t, got)
	})

	t.Run("a nil signal is an error", func(t *testing.T) {
		var nilSignal *Signal
		_, err := nilSignal.As[float64]()
		require.ErrorContains(t, err, "signal is nil")
	})

	t.Run("a nil payload is reported as such", func(t *testing.T) {
		_, err := New(nil).As[float64]()
		require.Error(t, err)
	})

	t.Run("composite type arguments work", func(t *testing.T) {
		got, err := New([]int{1, 2}).As[[]int]()
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2}, got)
	})
}

func TestSignal_PayloadOrDefault(t *testing.T) {
	t.Parallel()
	var nilSignal *Signal
	assert.InDelta(t, 3.5, New(3.5).PayloadOrDefault(0.0), 1e-9)
	assert.InDelta(t, 7.0, New("wrong type").PayloadOrDefault(7.0), 1e-9,
		"the wrong type degrades to the default rather than failing the run")
	assert.InDelta(t, 7.0, nilSignal.PayloadOrDefault(7.0), 1e-9)

	assert.Equal(t, 7, New(7).PayloadOrDefault(0), "T is inferred from the default")
	assert.Equal(t, "x", New("x").PayloadOrDefault(""))
	assert.True(t, New(true).PayloadOrDefault(false))
}

func TestSignal_PayloadOrDefaultInfersFromTheDefault(t *testing.T) {
	t.Parallel()
	// An untyped 0 makes T int, so a float64 payload silently yields the
	// default. Float64OrDefault exists for exactly this.
	assert.Equal(t, 0, New(2.5).PayloadOrDefault(0))
	assert.InDelta(t, 2.5, New(2.5).Float64OrDefault(0), 1e-9)
	assert.InDelta(t, 9.0, New("x").Float64OrDefault(9), 1e-9)

	var nilSignal *Signal
	assert.InDelta(t, 9.0, nilSignal.Float64OrDefault(9), 1e-9)
}

func TestSignal_AsGroup(t *testing.T) {
	t.Parallel()
	inner := NewGroup(1, 2)

	got, err := New(inner).AsGroup()
	require.NoError(t, err)
	assert.Equal(t, 2, got.Len())

	_, err = New(1).AsGroup()
	require.Error(t, err)

	var nilSignal *Signal
	_, err = nilSignal.AsGroup()
	require.Error(t, err)
}

func TestSignal_AsNumber(t *testing.T) {
	t.Parallel()
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
		// values in numeric metadata. Telling the two apart is what this is for.
		{name: "a type tag is not a measurement", payload: "venous_blood", ok: false},
		{name: "nil payload", payload: nil, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := New(tt.payload).AsNumber()
			assert.Equal(t, tt.ok, ok)
			assert.InDelta(t, tt.want, got, 1e-9)
		})
	}

	t.Run("nil signal", func(t *testing.T) {
		var nilSignal *Signal
		_, ok := nilSignal.AsNumber()
		assert.False(t, ok)
	})
}

func TestSignal_AsAsMethodValue(t *testing.T) {
	t.Parallel()
	// A generic method must be instantiated to be used as a value; once it is,
	// it behaves like any method value or expression.
	sig := New(42)
	asInt := sig.As[int]
	got, err := asInt()
	require.NoError(t, err)
	assert.Equal(t, 42, got)

	asString := (*Signal).As[string]
	_, err = asString(sig)
	require.Error(t, err)
}
