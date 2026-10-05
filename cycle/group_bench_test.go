package cycle

import (
	"strconv"
	"testing"
)

// BenchmarkGroupAdd_AtLenLimit guards eviction cost: adding to a full group
// evicts one cycle per Add, and ns/op must stay flat as the limit grows. A
// curve that rises with the limit means eviction went back to shifting the
// whole window.
func BenchmarkGroupAdd_AtLenLimit(b *testing.B) {
	for _, limit := range []int{10, 1_000, 100_000} {
		b.Run(strconv.Itoa(limit), func(b *testing.B) {
			group := NewGroup().SetLenLimit(limit)
			for i := range limit {
				group.Add(New().SetNumber(i))
			}
			c := New()
			b.ReportAllocs()
			for b.Loop() {
				group.Add(c)
			}
		})
	}
}
