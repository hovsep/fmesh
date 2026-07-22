package signal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// FuzzGroupOps guards CoW and length invariants on Group: With/Filter/Map/Join return
// a new group and never mutate the receiver, nil signals in With are skipped, and the
// resulting lengths follow the documented rules.
func FuzzGroupOps(f *testing.F) {
	f.Add(0, 0)
	f.Add(3, 2)
	f.Add(50, 10)

	f.Fuzz(func(t *testing.T, nRaw, mRaw int) {
		// Bound sizes so fuzzing stays fast and can't OOM.
		n := boundSize(nRaw)
		m := boundSize(mRaw)

		g := NewGroup()
		for i := range n {
			g = g.With(New(i))
		}
		lenBefore := g.Len()
		assert.Equal(t, n, lenBefore)

		// With appends and leaves the receiver untouched.
		withOne := g.With(New(-1))
		assert.Equal(t, lenBefore, g.Len(), "With mutated receiver")
		assert.Equal(t, lenBefore+1, withOne.Len())

		// nil signals are skipped by With.
		withNil := g.With(nil, nil)
		assert.Equal(t, lenBefore, withNil.Len(), "With should skip nil signals")

		// Filter never grows the group and never mutates the receiver.
		filtered := g.Filter(func(s *Signal) bool { return s.PayloadOrDefault(-1).(int)%2 == 0 })
		assert.Equal(t, lenBefore, g.Len(), "Filter mutated receiver")
		assert.LessOrEqual(t, filtered.Len(), lenBefore)

		// Map preserves length and leaves the receiver untouched.
		mapped := g.Map(func(s *Signal) *Signal { return s.WithLabel("m", "1") })
		assert.Equal(t, lenBefore, g.Len(), "Map mutated receiver")
		assert.Equal(t, lenBefore, mapped.Len())

		// Join concatenates lengths without mutating either operand.
		other := NewGroup()
		for i := range m {
			other = other.With(New(i))
		}
		joined := g.Join(other)
		assert.Equal(t, lenBefore, g.Len(), "Join mutated receiver")
		assert.Equal(t, m, other.Len(), "Join mutated argument")
		assert.Equal(t, lenBefore+m, joined.Len())
	})
}

// boundSize clamps a fuzzed count into [0, 1000].
func boundSize(v int) int {
	if v < 0 {
		v = -v
	}
	if v < 0 || v > 1000 { // v < 0 catches math.MinInt after negation
		return 1000
	}
	return v
}
