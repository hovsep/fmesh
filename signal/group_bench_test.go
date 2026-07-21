package signal

import (
	"strconv"
	"testing"
)

// benchGroup returns a group of n signals with int payloads and a "v" scalar.
func benchGroup(n int) *Group {
	g := NewGroup()
	for i := range n {
		g = g.With(New(i).WithScalar("v", float64(i)).WithLabel("idx", strconv.Itoa(i)))
	}
	return g
}

// BenchmarkGroupCoWOps measures the copy-on-write pipeline: With + Map + Filter
// over a 100-signal group.
func BenchmarkGroupCoWOps(b *testing.B) {
	g := benchGroup(100)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = g.With(New(-1)).
			Map(func(s *Signal) *Signal {
				return s.WithLabel("mapped", "true")
			}).
			Filter(func(s *Signal) bool {
				return s.PayloadOrDefault(0).(int)%2 == 0
			})
	}
}

// BenchmarkGroupScalarAggregation measures cross-signal scalar aggregation:
// SumScalar + AvgScalar over a 100-signal group.
func BenchmarkGroupScalarAggregation(b *testing.B) {
	g := benchGroup(100)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = g.SumScalar("v")
		if _, err := g.AvgScalar("v"); err != nil {
			b.Fatal(err)
		}
	}
}
