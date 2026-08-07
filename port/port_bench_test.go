package port

import (
	"context"
	"fmt"
	"testing"
)

// Flush is the hottest path in the framework: every activated component flushes
// every output port on every cycle. These guard the cost of the
// OnSignalsDelivered hook, whose context struct escapes to the heap whether or
// not anything reads it. A non-zero rise in allocs/op on the no-hook variants
// means the IsEmpty guard in Flush stopped working.
func BenchmarkPortFlush(b *testing.B) {
	for _, pipes := range []int{1, 4} {
		for _, hooked := range []bool{false, true} {
			name := fmt.Sprintf("pipes=%d/hook=%t", pipes, hooked)
			b.Run(name, func(b *testing.B) {
				ctx := context.Background()
				src := mustOutput("out")
				dests := make([]*Port, pipes)
				for i := range pipes {
					dests[i] = mustInput(fmt.Sprintf("in_%d", i))
					if err := src.PipeTo(dests[i]); err != nil {
						b.Fatal(err)
					}
				}
				if hooked {
					src.SetupHooks(func(hooks *Hooks) {
						hooks.OnSignalsDelivered(func(context.Context, *SignalsDeliveredContext) error {
							return nil
						})
					})
				}

				b.ReportAllocs()
				for b.Loop() {
					if err := src.PutPayloads(1, 2, 3); err != nil {
						b.Fatal(err)
					}
					if err := src.Flush(ctx); err != nil {
						b.Fatal(err)
					}
					// Nothing drains a bare port, and putSignals clones what the
					// destination already holds -- without this the benchmark
					// measures unbounded growth instead of one flush.
					for _, d := range dests {
						if err := d.Clear(ctx); err != nil {
							b.Fatal(err)
						}
					}
				}
			})
		}
	}
}
