// Package determinism pins down the ordering guarantees a mesh makes.
//
// These are not micro-tests of a sort function: they run whole meshes many
// times and assert that every run produced the same answer. Before port
// collections were traversed in name order, the fan-in case below produced four
// different results across 200 identical runs, which made any mesh that
// concatenates or reduces its inputs irreproducible — and untestable.
package determinism

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

const runs = 200

// collect runs build() the given number of times and returns the distinct results.
func collect(t *testing.T, build func(t *testing.T) string) map[string]int {
	t.Helper()
	seen := make(map[string]int)
	for range runs {
		seen[build(t)]++
	}
	return seen
}

func TestOrdering_AcrossPortsOfOneComponent(t *testing.T) {
	// A component reading everything that arrived, via Inputs().Signals().
	got := collect(t, func(t *testing.T) string {
		t.Helper()
		sink := testutil.MustComponent("sink",
			component.WithInputs("i1", "i2", "i3", "i4"),
			component.WithOutputs("out"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				var order strings.Builder
				_ = this.Inputs().Signals().ForEach(func(s *signal.Signal) error {
					fmt.Fprint(&order, s.PayloadOrNil())
					return nil
				})
				return this.OutputByName("out").PutSignals(signal.New(order.String()))
			}))

		fm, err := fmesh.New("fan-in")
		require.NoError(t, err)
		require.NoError(t, fm.AddComponents(sink))

		// Seeded in an order that does not match the port names, so a result
		// matching port-name order cannot be an accident of insertion order.
		require.NoError(t, sink.InputByName("i3").PutSignals(signal.New(3)))
		require.NoError(t, sink.InputByName("i1").PutSignals(signal.New(1)))
		require.NoError(t, sink.InputByName("i4").PutSignals(signal.New(4)))
		require.NoError(t, sink.InputByName("i2").PutSignals(signal.New(2)))

		_, err = fm.Run(context.Background())
		require.NoError(t, err)

		payload, err := sink.OutputByName("out").Signals().FirstPayload()
		require.NoError(t, err)
		return fmt.Sprint(payload)
	})

	require.Len(t, got, 1, "identical runs must produce one result, got %v", got)
	assert.Contains(t, got, "1234", "signals must arrive in port-name order")
}

func TestOrdering_FlushAcrossOutputPorts(t *testing.T) {
	// Two output ports of one component feeding one downstream port: the order
	// they are flushed in decides what the downstream component sees.
	got := collect(t, func(t *testing.T) string {
		t.Helper()
		source := testutil.MustComponent("source",
			component.WithInputs("in"),
			component.WithOutputs("a_out", "b_out", "c_out"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				for _, name := range []string{"c_out", "a_out", "b_out"} {
					if err := this.OutputByName(name).PutSignals(signal.New(name[:1])); err != nil {
						return err
					}
				}
				return nil
			}))
		sink := testutil.MustComponent("sink",
			component.WithInputs("in"),
			component.WithOutputs("out"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				var order strings.Builder
				_ = this.InputByName("in").Signals().ForEach(func(s *signal.Signal) error {
					fmt.Fprint(&order, s.PayloadOrNil())
					return nil
				})
				return this.OutputByName("out").PutSignals(signal.New(order.String()))
			}))

		fm, err := fmesh.New("fan-out-in")
		require.NoError(t, err)
		require.NoError(t, fm.AddComponents(source, sink))
		require.NoError(t, source.Outputs().PipeTo(sink.InputByName("in")))
		require.NoError(t, source.InputByName("in").PutSignals(signal.New("go")))

		_, err = fm.Run(context.Background())
		require.NoError(t, err)

		payload, err := sink.OutputByName("out").Signals().FirstPayload()
		require.NoError(t, err)
		return fmt.Sprint(payload)
	})

	require.Len(t, got, 1, "identical runs must produce one result, got %v", got)
	assert.Contains(t, got, "abc", "output ports must flush in port-name order")
}

func TestOrdering_MultipleUpstreamsIntoOnePort(t *testing.T) {
	// Three components feeding one input port. Drain order is component-name
	// order, which fmesh already guaranteed; this pins it down.
	got := collect(t, func(t *testing.T) string {
		t.Helper()
		mkSource := func(name, payload string) *component.Component {
			return testutil.MustComponent(name,
				component.WithInputs("in"),
				component.WithOutputs("out"),
				component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
					return this.OutputByName("out").PutSignals(signal.New(payload))
				}))
		}
		// Registered in an order that is not the name order.
		charlie, alpha, bravo := mkSource("charlie", "c"), mkSource("alpha", "a"), mkSource("bravo", "b")
		sink := testutil.MustComponent("zsink",
			component.WithInputs("in"),
			component.WithOutputs("out"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				var order strings.Builder
				_ = this.InputByName("in").Signals().ForEach(func(s *signal.Signal) error {
					fmt.Fprint(&order, s.PayloadOrNil())
					return nil
				})
				return this.OutputByName("out").PutSignals(signal.New(order.String()))
			}))

		fm, err := fmesh.New("fan-in-multi")
		require.NoError(t, err)
		require.NoError(t, fm.AddComponents(charlie, alpha, bravo, sink))
		for _, c := range []*component.Component{charlie, alpha, bravo} {
			require.NoError(t, c.OutputByName("out").PipeTo(sink.InputByName("in")))
			require.NoError(t, c.InputByName("in").PutSignals(signal.New("go")))
		}

		_, err = fm.Run(context.Background())
		require.NoError(t, err)

		payload, err := sink.OutputByName("out").Signals().FirstPayload()
		require.NoError(t, err)
		return fmt.Sprint(payload)
	})

	require.Len(t, got, 1, "identical runs must produce one result, got %v", got)
	assert.Contains(t, got, "abc", "upstreams must drain in component-name order")
}

func TestOrdering_WithinOnePortIsFIFO(t *testing.T) {
	p := testutil.MustComponent("c", component.WithInputs("in"))
	in := p.InputByName("in")
	for i := range 50 {
		require.NoError(t, in.PutSignals(signal.New(i)))
	}

	var got []any
	_ = in.Signals().ForEach(func(s *signal.Signal) error {
		got = append(got, s.PayloadOrNil())
		return nil
	})

	require.Len(t, got, 50)
	for i := range 50 {
		assert.Equal(t, i, got[i], "signals within one port keep insertion order")
	}
}

func TestOrdering_AllOrderedMatchesTraversal(t *testing.T) {
	c := testutil.MustComponent("c", component.WithInputs("zebra", "alpha", "middle"))

	var viaForEach []string
	require.NoError(t, c.Inputs().ForEach(func(p *port.Port) error {
		viaForEach = append(viaForEach, p.Name())
		return nil
	}))

	ordered := c.Inputs().AllOrdered()
	viaAllOrdered := make([]string, 0, len(ordered))
	for _, p := range ordered {
		viaAllOrdered = append(viaAllOrdered, p.Name())
	}

	assert.Equal(t, []string{"alpha", "middle", "zebra"}, viaAllOrdered)
	assert.Equal(t, viaAllOrdered, viaForEach, "ForEach must use the AllOrdered order")
}
