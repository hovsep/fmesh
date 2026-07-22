package signal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// FuzzSignalCoW guards the copy-on-write invariant: every mutating-style method must
// return a new *Signal and leave the receiver untouched (payload, labels, scalars).
// nil is a valid payload, so it is exercised via the nilPayload arm.
func FuzzSignalCoW(f *testing.F) {
	f.Add("payload", false, "label", "value", "scalar", 1.5)
	f.Add("", true, "", "", "", 0.0)
	f.Add("x", false, "k", "", "s", -3.0)

	f.Fuzz(func(t *testing.T,
		payloadStr string, nilPayload bool,
		labelKey, labelVal, scalarName string, scalarVal float64,
	) {
		var payload any = payloadStr
		if nilPayload {
			payload = nil
		}
		s := New(payload)

		// Snapshot the receiver's observable state.
		labelsBefore := s.Labels().All()
		scalarsBefore := s.Scalars().All()
		payloadBefore := s.PayloadOrNil()

		assertUnchanged := func(what string) {
			assert.Equal(t, labelsBefore, s.Labels().All(), "%s mutated receiver labels", what)
			assert.Equal(t, scalarsBefore, s.Scalars().All(), "%s mutated receiver scalars", what)
			assert.Equal(t, payloadBefore, s.PayloadOrNil(), "%s mutated receiver payload", what)
		}

		// WithLabel: returns new, receiver unchanged, round-trips the value.
		withLabel := s.WithLabel(labelKey, labelVal)
		assertUnchanged("WithLabel")
		assert.Equal(t, labelVal, withLabel.Labels().ValueOrDefault(labelKey, "\x00sentinel"))

		// WithScalar: same guarantees.
		withScalar := s.WithScalar(scalarName, scalarVal)
		assertUnchanged("WithScalar")
		assert.True(t, withScalar.Scalars().ValueIs(scalarName, scalarVal))

		// WithoutLabels removes from the copy but not from its source.
		without := withLabel.WithoutLabels(labelKey)
		assert.True(t, withLabel.Labels().Has(labelKey), "WithoutLabels mutated its source")
		assert.False(t, without.Labels().Has(labelKey))

		// Map / MapPayload must not touch the receiver's payload.
		mapped := s.Map(func(sig *Signal) *Signal { return sig.WithLabel("mapped", "true") })
		assertUnchanged("Map")
		assert.Equal(t, "true", mapped.Labels().ValueOrDefault("mapped", ""))

		remapped := s.MapPayload(func(any) any { return "remapped" })
		assertUnchanged("MapPayload")
		assert.Equal(t, "remapped", remapped.PayloadOrNil())
	})
}
