package signal

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FuzzSignalCoW guards the copy-on-write invariant: every mutating-style method must
// return a new *Signal and leave the receiver untouched (payload and metadata).
// nil is a valid payload, so it is exercised via the nilPayload arm.
func FuzzSignalCoW(f *testing.F) {
	f.Add("payload", false, "str", "value", "num", 1.5)
	f.Add("", true, "", "", "", 0.0)
	f.Add("x", false, "k", "", "s", -3.0)
	f.Add("x", false, "k", "v", "k", math.NaN())

	f.Fuzz(func(t *testing.T,
		payloadStr string, nilPayload bool,
		strKey, strVal, numKey string, numVal float64,
	) {
		var payload any = payloadStr
		if nilPayload {
			payload = nil
		}
		s := New(payload)

		// Snapshot the receiver's observable state.
		metaBefore := s.Meta().All()
		payloadBefore := s.Payload()

		assertUnchanged := func(what string) {
			assert.Equal(t, metaBefore, s.Meta().All(), "%s mutated receiver metadata", what)
			assert.Equal(t, payloadBefore, s.Payload(), "%s mutated receiver payload", what)
		}

		// WithMeta (string): returns new, receiver unchanged, round-trips the value.
		withStr := s.WithMeta(strKey, strVal)
		assertUnchanged("WithMeta string")
		assert.Equal(t, strVal, withStr.Meta().ValueOrDefault(strKey, "\x00sentinel"))

		// WithMeta (float): same guarantees. Bits are compared because NaN != NaN,
		// so ValueIs cannot round-trip it.
		withNum := s.WithMeta(numKey, numVal)
		assertUnchanged("WithMeta float")
		got, err := withNum.Meta().Value[float64](numKey)
		require.NoError(t, err)
		assert.Equal(t, math.Float64bits(numVal), math.Float64bits(got))

		// WithoutMeta removes from the copy but not from its source.
		without := withStr.WithoutMeta(strKey)
		assert.True(t, withStr.Meta().Has(strKey), "WithoutMeta mutated its source")
		assert.False(t, without.Meta().Has(strKey))

		// Map / MapPayload must not touch the receiver's payload.
		mapped := s.Map(func(sig *Signal) *Signal { return sig.WithMeta("mapped", "true") })
		assertUnchanged("Map")
		assert.Equal(t, "true", mapped.Meta().ValueOrDefault("mapped", ""))

		remapped := s.MapPayload(func(any) any { return "remapped" })
		assertUnchanged("MapPayload")
		assert.Equal(t, "remapped", remapped.Payload())
	})
}
