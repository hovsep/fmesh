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
	for b.Loop() {
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
	for b.Loop() {
		_ = g.SumScalar("v")
		if _, err := g.AvgScalar("v"); err != nil {
			b.Fatal(err)
		}
	}
}

// groupBenchSizes is the size sweep for group scale benchmarks.
var groupBenchSizes = []int{10, 100, 1_000, 10_000}

// BenchmarkGroupBuild measures building a group by repeated With across sizes.
// With is O(n) per call (it allocates a fresh len+1 slice), so building this way is
// O(n²) overall — the sweep makes that visible and informs whether a bulk constructor
// would ever be worth adding.
func BenchmarkGroupBuild(b *testing.B) {
	for _, n := range groupBenchSizes {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				g := NewGroup()
				for i := range n {
					g = g.With(New(i))
				}
				_ = g
			}
		})
	}
}

// payloadOfSize returns a byte slice payload of the given size. Because signal CoW is
// a shallow copy, the payload's size should not affect per-op copy cost.
func payloadOfSize(size int) []byte {
	return make([]byte, size)
}

// BenchmarkGroupPayloadSize runs a CoW pipeline over a single signal while sweeping the
// payload size. ns/op should stay flat across sizes: CoW copies the interface header,
// not the pointed-to payload. A rise here means a deep copy crept in.
func BenchmarkGroupPayloadSize(b *testing.B) {
	for _, size := range []int{8, 1 << 10, 1 << 16, 1 << 20} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			s := New(payloadOfSize(size))
			b.ReportAllocs()
			for b.Loop() {
				_ = s.WithLabel("k", "v").
					WithScalar("n", 1).
					Map(func(sig *Signal) *Signal { return sig.WithLabel("m", "x") })
			}
		})
	}
}
