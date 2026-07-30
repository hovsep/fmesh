package port

import (
	"fmt"
	"testing"
)

// Traversal happens for every component on every cycle (ClearInputs,
// FlushOutputs), and it is ordered so that runs are reproducible. These guard
// the cost of that ordering: it must stay allocation-free. A non-zero allocs/op
// here means a traversal started materializing the port list — see
// Collection.each.
func benchmarkCollection(b *testing.B, portCount int) *Collection {
	b.Helper()
	c := NewCollection()
	for i := range portCount {
		p, err := NewInput(fmt.Sprintf("port_%03d", i))
		if err != nil {
			b.Fatal(err)
		}
		if err := c.Add(p); err != nil {
			b.Fatal(err)
		}
	}
	return c
}

func BenchmarkCollectionForEach(b *testing.B) {
	for _, portCount := range []int{2, 8, 32} {
		b.Run(fmt.Sprintf("ports=%d", portCount), func(b *testing.B) {
			c := benchmarkCollection(b, portCount)
			b.ReportAllocs()
			for b.Loop() {
				_ = c.ForEach(func(*Port) error { return nil })
			}
		})
	}
}

func BenchmarkCollectionAnyHasSignals(b *testing.B) {
	c := benchmarkCollection(b, 8)
	b.ReportAllocs()
	for b.Loop() {
		_ = c.AnyHasSignals()
	}
}
